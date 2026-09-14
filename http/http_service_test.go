package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/getAlby/hub/api"
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/tests/db"
	"github.com/getAlby/hub/tests/mocks"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUnlock_IncorrectPassword(t *testing.T) {
	e := echo.New()
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	mockSvc := mocks.NewMockService(t)
	gormDb, err := db.NewDB(t)
	require.NoError(t, err)
	defer db.CloseDB(gormDb)

	mockEventPublisher := events.NewEventPublisher()

	mockConfig := mocks.NewMockConfig(t)
	mockConfig.On("GetEnv").Return(&config.AppConfig{})
	mockConfig.On("CheckUnlockPassword", "123").Return(false)

	mockSvc.On("GetDB").Return(gormDb)
	mockSvc.On("GetConfig").Return(mockConfig)
	mockSvc.On("GetKeys").Return(mocks.NewMockKeys(t))
	mockSvc.On("GetAlbySvc").Return(mocks.NewMockAlbyService(t))
	mockSvc.On("GetAlbyOAuthSvc").Return(mocks.NewMockAlbyOAuthService(t))

	httpSvc := NewHttpService(mockSvc, mockEventPublisher)
	httpSvc.RegisterSharedRoutes(e)

	requestBody := api.UnlockRequest{UnlockPassword: "123", Permission: "full"}
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/unlock", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json") // Set Content-Type header
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	mockConfig.AssertNotCalled(t, "GetJWTSecret")
}

// TestStart_IncorrectPassword verifies that the start endpoint rejects an
// invalid unlock password before issuing a JWT or launching the node start
// (regression test for #1560: previously the token was created unconditionally
// and the password was only checked inside the async node start).
func TestStart_IncorrectPassword(t *testing.T) {
	e := echo.New()
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	mockSvc := mocks.NewMockService(t)
	gormDb, err := db.NewDB(t)
	require.NoError(t, err)
	defer db.CloseDB(gormDb)

	mockEventPublisher := events.NewEventPublisher()

	mockConfig := mocks.NewMockConfig(t)
	mockConfig.On("GetEnv").Return(&config.AppConfig{})
	mockConfig.On("CheckUnlockPassword", "123").Return(false)

	mockSvc.On("GetDB").Return(gormDb)
	mockSvc.On("GetConfig").Return(mockConfig)
	mockSvc.On("GetKeys").Return(mocks.NewMockKeys(t))
	mockSvc.On("GetAlbySvc").Return(mocks.NewMockAlbyService(t))
	mockSvc.On("GetAlbyOAuthSvc").Return(mocks.NewMockAlbyOAuthService(t))

	httpSvc := NewHttpService(mockSvc, mockEventPublisher)
	httpSvc.RegisterSharedRoutes(e)

	requestBody := api.StartRequest{UnlockPassword: "123", Session: true}
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/start", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	assert.Equal(t, "Invalid password", response.Message)

	// black-box: nothing JWT-shaped may reach the client anywhere in the response
	assert.Empty(t, findJWTs(rec), "no token must be issued for an invalid password")

	// white-box: the calls that TestStart_CorrectPassword proves are required to
	// issue a token must not happen here. Keep both tests in sync.
	mockConfig.AssertNotCalled(t, "LoadJWTSecret", mock.Anything)
	mockConfig.AssertNotCalled(t, "GetJWTSecret")
	mockSvc.AssertNotCalled(t, "StartApp", mock.Anything)
}

// TestStart_CorrectPassword is the counterpart of TestStart_IncorrectPassword:
// it proves that starting a browser session requires LoadJWTSecret,
// GetJWTSecret and StartApp, so the AssertNotCalled checks in the negative test
// cannot silently become vacuous if the implementation changes. It also proves
// the issued HttpOnly cookie grants full access without exposing the JWT body.
func TestStart_CorrectPassword(t *testing.T) {
	e := echo.New()
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	mockSvc := mocks.NewMockService(t)
	gormDb, err := db.NewDB(t)
	require.NoError(t, err)
	defer db.CloseDB(gormDb)

	mockEventPublisher := events.NewEventPublisher()

	mockConfig := mocks.NewMockConfig(t)
	mockConfig.On("GetEnv").Return(&config.AppConfig{})
	mockConfig.On("CheckUnlockPassword", "123").Return(true)
	mockConfig.On("LoadJWTSecret", "123").Return(nil)
	mockConfig.On("GetJWTSecret").Return("dummy secret", nil)

	// the node start runs asynchronously; wait for it so the mock expectation
	// is observed before the test ends
	started := make(chan struct{})
	mockSvc.On("StartApp", "123").Run(func(args mock.Arguments) { close(started) }).Return(nil)

	mockSvc.On("GetDB").Return(gormDb)
	mockSvc.On("GetConfig").Return(mockConfig)
	mockSvc.On("GetKeys").Return(mocks.NewMockKeys(t))
	mockSvc.On("GetAlbySvc").Return(mocks.NewMockAlbyService(t))
	mockSvc.On("GetAlbyOAuthSvc").Return(mocks.NewMockAlbyOAuthService(t))

	httpSvc := NewHttpService(mockSvc, mockEventPublisher)
	httpSvc.RegisterSharedRoutes(e)

	requestBody := api.StartRequest{UnlockPassword: "123", Session: true}
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/start", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(echo.HeaderXForwardedProto, "https")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Body.String(), "browser session JWT must not be exposed in the response body")

	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("StartApp was not called")
	}

	mockConfig.AssertCalled(t, "LoadJWTSecret", "123")
	mockConfig.AssertCalled(t, "GetJWTSecret")

	var sessionCookie *http.Cookie
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == sessionCookieName {
			sessionCookie = cookie
			break
		}
	}
	require.NotNil(t, sessionCookie)
	assert.Equal(t, []string{sessionCookie.Value}, findJWTs(rec))
	assert.True(t, sessionCookie.HttpOnly)
	assert.True(t, sessionCookie.Secure)
	assert.Equal(t, http.SameSiteStrictMode, sessionCookie.SameSite)
	assert.Empty(t, sessionCookie.Path)
	assert.Greater(t, sessionCookie.MaxAge, 0)

	// The cookie grants full access while Authorization remains available for
	// HTTP Basic Authentication at the reverse proxy.
	req2 := httptest.NewRequest(http.MethodGet, "/api/apps", nil)
	req2.AddCookie(sessionCookie)
	req2.Header.Set("Authorization", "Basic dXNlcjpwYXNzd29yZA==")
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code)
	assert.Equal(t, "private, no-store", rec2.Header().Get(echo.HeaderCacheControl))
	assert.ElementsMatch(t, []string{echo.HeaderCookie, echo.HeaderAuthorization}, rec2.Header().Values(echo.HeaderVary))

	// Keep the standard Authorization header working for external API clients.
	req3 := httptest.NewRequest(http.MethodGet, "/api/apps", nil)
	req3.Header.Set("Authorization", "Bearer "+sessionCookie.Value)
	rec3 := httptest.NewRecorder()
	e.ServeHTTP(rec3, req3)
	assert.Equal(t, http.StatusOK, rec3.Code)
}

type infoAPIStub struct {
	api.API
}

func (infoAPIStub) GetInfo(context.Context) (*api.InfoResponse, error) {
	return &api.InfoResponse{}, nil
}

func TestInfo_RecognizesJWTAuthentication(t *testing.T) {
	const jwtSecret = "dummy secret"
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(jwtSecret))
	require.NoError(t, err)

	tests := []struct {
		name         string
		cookies      []*http.Cookie
		headers      http.Header
		wantUnlocked bool
	}{
		{
			name: "session cookie alongside HTTP Basic Auth",
			cookies: []*http.Cookie{
				{Name: sessionCookieName, Value: token},
			},
			headers: http.Header{
				"Authorization": {"Basic dXNlcjpwYXNzd29yZA=="},
			},
			wantUnlocked: true,
		},
		{
			name: "valid duplicate session cookie",
			cookies: []*http.Cookie{
				{Name: sessionCookieName, Value: "invalid"},
				{Name: sessionCookieName, Value: token},
			},
			wantUnlocked: true,
		},
		{
			name: "case-insensitive Bearer scheme",
			headers: http.Header{
				"Authorization": {"bearer " + token},
			},
			wantUnlocked: true,
		},
		{
			name: "standard Authorization header fallback",
			headers: http.Header{
				"Authorization": {"Bearer " + token},
			},
			wantUnlocked: true,
		},
		{
			name: "invalid session cookie falls back to Authorization",
			cookies: []*http.Cookie{
				{Name: sessionCookieName, Value: "invalid"},
			},
			headers: http.Header{
				"Authorization": {"Bearer " + token},
			},
			wantUnlocked: true,
		},
		{
			name: "HTTP Basic Auth alone does not unlock the Hub",
			headers: http.Header{
				"Authorization": {"Basic dXNlcjpwYXNzd29yZA=="},
			},
			wantUnlocked: false,
		},
		{
			name: "malformed Bearer header suppresses session cookie",
			cookies: []*http.Cookie{
				{Name: sessionCookieName, Value: token},
			},
			headers: http.Header{
				"Authorization": {"Bearer"},
			},
			wantUnlocked: false,
		},
		{
			name: "invalid Bearer token suppresses session cookie",
			cookies: []*http.Cookie{
				{Name: sessionCookieName, Value: token},
			},
			headers: http.Header{
				"Authorization": {"Bearer invalid"},
			},
			wantUnlocked: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConfig := mocks.NewMockConfig(t)
			mockConfig.On("GetJWTSecret").Maybe().Return(jwtSecret, nil)
			httpSvc := &HttpService{api: infoAPIStub{}, cfg: mockConfig}

			e := echo.New()
			e.GET("/api/info", httpSvc.infoHandler, httpSvc.jwtMiddleware(true))
			req := httptest.NewRequest(http.MethodGet, "/api/info", nil)
			for _, cookie := range tt.cookies {
				req.AddCookie(cookie)
			}
			for name, values := range tt.headers {
				for _, value := range values {
					req.Header.Add(name, value)
				}
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			require.Equal(t, http.StatusOK, rec.Code)
			var response api.InfoResponse
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
			assert.Equal(t, tt.wantUnlocked, response.Unlocked)
		})
	}
}

func signTestJWT(t *testing.T, secret string, permission string) string {
	t.Helper()

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtCustomClaims{
		Permission: permission,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}).SignedString([]byte(secret))
	require.NoError(t, err)
	return token
}

func TestBearerTokenTakesPrecedenceOverSessionCookie(t *testing.T) {
	const jwtSecret = "dummy secret"
	fullSessionToken := signTestJWT(t, jwtSecret, "full")
	readonlyAPIToken := signTestJWT(t, jwtSecret, "readonly")

	mockConfig := mocks.NewMockConfig(t)
	mockConfig.On("GetJWTSecret").Return(jwtSecret, nil)
	httpSvc := &HttpService{cfg: mockConfig}

	e := echo.New()
	e.GET("/full", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	}, httpSvc.jwtMiddleware(false), httpSvc.requireSameOriginForSession, httpSvc.requireFullAccess)

	tests := []struct {
		name           string
		bearerToken    string
		expectedStatus int
	}{
		{
			name:           "valid readonly token does not inherit full cookie permissions",
			bearerToken:    readonlyAPIToken,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "invalid token does not fall back to full cookie",
			bearerToken:    "invalid",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/full", nil)
			req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: fullSessionToken})
			req.Header.Set(echo.HeaderAuthorization, bearerPrefix+tt.bearerToken)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestCookieSessionRequiresSameOriginForUnsafeRequests(t *testing.T) {
	const jwtSecret = "dummy secret"
	sessionToken := signTestJWT(t, jwtSecret, "full")

	mockConfig := mocks.NewMockConfig(t)
	mockConfig.On("GetJWTSecret").Return(jwtSecret, nil)
	httpSvc := &HttpService{cfg: mockConfig}

	e := echo.New()
	e.POST("/api/action", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	}, httpSvc.jwtMiddleware(false), httpSvc.requireSameOriginForSession)

	tests := []struct {
		name           string
		origin         string
		withCookie     bool
		authorization  string
		expectedStatus int
	}{
		{
			name:           "same-origin cookie request behind proxy",
			origin:         "https://hub.example.com",
			withCookie:     true,
			authorization:  "Basic dXNlcjpwYXNzd29yZA==",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "cross-origin cookie request",
			origin:         "https://attacker.example.com",
			withCookie:     true,
			authorization:  "Basic dXNlcjpwYXNzd29yZA==",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "cookie request without origin",
			withCookie:     true,
			authorization:  "Basic dXNlcjpwYXNzd29yZA==",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "explicit Bearer API token does not require origin",
			withCookie:     true,
			authorization:  bearerPrefix + sessionToken,
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/action", nil)
			req.Host = "127.0.0.1:8080"
			req.Header.Set(forwardedHostHeader, "hub.example.com")
			req.Header.Set(echo.HeaderXForwardedProto, "https")
			if tt.origin != "" {
				req.Header.Set(echo.HeaderOrigin, tt.origin)
			}
			if tt.authorization != "" {
				req.Header.Set(echo.HeaderAuthorization, tt.authorization)
			}
			if tt.withCookie {
				req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionToken})
			}

			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestSessionCookieAllowsPlainHTTP(t *testing.T) {
	httpSvc := &HttpService{}
	e := echo.New()
	e.GET("/set-cookie", func(c echo.Context) error {
		httpSvc.setSessionCookie(c, "token")
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/set-cookie", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Len(t, rec.Result().Cookies(), 1)
	assert.False(t, rec.Result().Cookies()[0].Secure)
}

func TestSessionCookieDefaultsToExternalAPIPath(t *testing.T) {
	httpSvc := &HttpService{}
	e := echo.New()
	e.GET("/hub/api/start", func(c echo.Context) error {
		httpSvc.setSessionCookie(c, "token")
		return c.NoContent(http.StatusNoContent)
	})
	e.POST("/hub/api/logout", func(c echo.Context) error {
		httpSvc.clearSessionCookie(c)
		return c.NoContent(http.StatusNoContent)
	})

	server := httptest.NewServer(e)
	defer server.Close()

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	client := &http.Client{Jar: jar}

	response, err := client.Get(server.URL + "/hub/api/start")
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())

	apiURL, err := url.Parse(server.URL + "/hub/api/apps")
	require.NoError(t, err)
	otherURL, err := url.Parse(server.URL + "/unrelated-service")
	require.NoError(t, err)
	assert.Len(t, jar.Cookies(apiURL), 1)
	assert.Empty(t, jar.Cookies(otherURL))

	request, err := http.NewRequest(http.MethodPost, server.URL+"/hub/api/logout", nil)
	require.NoError(t, err)
	response, err = client.Do(request)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	assert.Empty(t, jar.Cookies(apiURL))
}

func TestLogoutExpiresSessionCookie(t *testing.T) {
	httpSvc := &HttpService{}

	e := echo.New()
	e.POST("/api/logout", httpSvc.logoutHandler, httpSvc.requireSameOrigin)
	req := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
	req.Host = "hub.example.com"
	req.Header.Set(echo.HeaderOrigin, "https://hub.example.com")
	req.Header.Set(echo.HeaderXForwardedProto, "https")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	require.Len(t, rec.Result().Cookies(), 1)
	cookie := rec.Result().Cookies()[0]
	assert.Equal(t, sessionCookieName, cookie.Name)
	assert.Empty(t, cookie.Value)
	assert.Equal(t, -1, cookie.MaxAge)
	assert.Empty(t, cookie.Path)
	assert.True(t, cookie.Secure)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
}

func TestLogoutRejectsCrossOrigin(t *testing.T) {
	httpSvc := &HttpService{}
	e := echo.New()
	e.POST("/api/logout", httpSvc.logoutHandler, httpSvc.requireSameOrigin)

	req := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
	req.Host = "hub.example.com"
	req.Header.Set(echo.HeaderOrigin, "https://attacker.example.com")
	req.Header.Set(echo.HeaderXForwardedProto, "https")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Empty(t, rec.Result().Cookies())
}

var jwtPattern = regexp.MustCompile(`eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+`)

// findJWTs returns every JWT-shaped string in the response body and headers
// (including Set-Cookie), independent of JSON field names.
func findJWTs(rec *httptest.ResponseRecorder) []string {
	var haystack bytes.Buffer
	haystack.Write(rec.Body.Bytes())
	for key, values := range rec.Header() {
		for _, v := range values {
			haystack.WriteString("\n" + key + ": " + v)
		}
	}
	return jwtPattern.FindAllString(haystack.String(), -1)
}

func TestUnlock_UnknownPermission(t *testing.T) {
	e := echo.New()
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	mockSvc := mocks.NewMockService(t)
	gormDb, err := db.NewDB(t)
	require.NoError(t, err)
	defer db.CloseDB(gormDb)

	mockEventPublisher := events.NewEventPublisher()

	mockConfig := mocks.NewMockConfig(t)
	mockConfig.On("GetEnv").Return(&config.AppConfig{})
	mockConfig.On("CheckUnlockPassword", "123").Return(true)

	mockSvc.On("GetDB").Return(gormDb)
	mockSvc.On("GetConfig").Return(mockConfig)
	mockSvc.On("GetKeys").Return(mocks.NewMockKeys(t))
	mockSvc.On("GetAlbySvc").Return(mocks.NewMockAlbyService(t))
	mockSvc.On("GetAlbyOAuthSvc").Return(mocks.NewMockAlbyOAuthService(t))

	httpSvc := NewHttpService(mockSvc, mockEventPublisher)
	httpSvc.RegisterSharedRoutes(e)

	requestBody := api.UnlockRequest{UnlockPassword: "123", Permission: "unknown"}
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/unlock", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json") // Set Content-Type header
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockConfig.AssertNotCalled(t, "GetJWTSecret")
}

// TestUnlock_RateLimited verifies that repeated requests to an unlock-password
// endpoint are throttled with HTTP 429 once the limit is exceeded.
func TestUnlock_RateLimited(t *testing.T) {
	e := echo.New()
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	mockSvc := mocks.NewMockService(t)
	gormDb, err := db.NewDB(t)
	require.NoError(t, err)
	defer db.CloseDB(gormDb)

	mockEventPublisher := events.NewEventPublisher()

	mockConfig := mocks.NewMockConfig(t)
	mockConfig.On("GetEnv").Return(&config.AppConfig{})
	mockConfig.On("CheckUnlockPassword", "wrong").Return(false)

	mockSvc.On("GetDB").Return(gormDb)
	mockSvc.On("GetConfig").Return(mockConfig)
	mockSvc.On("GetKeys").Return(mocks.NewMockKeys(t))
	mockSvc.On("GetAlbySvc").Return(mocks.NewMockAlbyService(t))
	mockSvc.On("GetAlbyOAuthSvc").Return(mocks.NewMockAlbyOAuthService(t))

	httpSvc := NewHttpService(mockSvc, mockEventPublisher)
	httpSvc.RegisterSharedRoutes(e)

	jsonBody, _ := json.Marshal(api.UnlockRequest{UnlockPassword: "wrong", Permission: "full"})
	send := func() int {
		req := httptest.NewRequest(http.MethodPost, "/api/unlock", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec.Code
	}

	// the burst of 2 is served (wrong password, so unauthorized)
	assert.Equal(t, http.StatusUnauthorized, send())
	assert.Equal(t, http.StatusUnauthorized, send())
	// the next request exceeds the limit and is rejected with 429
	assert.Equal(t, http.StatusTooManyRequests, send())
}

// TestUnlock_RateLimitNotBypassedBySpoofedIP verifies that the unlock rate
// limiter is global rather than per-IP: varying the X-Forwarded-For header per
// request does not grant each request a fresh bucket.
func TestUnlock_RateLimitNotBypassedBySpoofedIP(t *testing.T) {
	e := echo.New()
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	mockSvc := mocks.NewMockService(t)
	gormDb, err := db.NewDB(t)
	require.NoError(t, err)
	defer db.CloseDB(gormDb)

	mockEventPublisher := events.NewEventPublisher()

	mockConfig := mocks.NewMockConfig(t)
	mockConfig.On("GetEnv").Return(&config.AppConfig{})
	mockConfig.On("CheckUnlockPassword", "wrong").Return(false)

	mockSvc.On("GetDB").Return(gormDb)
	mockSvc.On("GetConfig").Return(mockConfig)
	mockSvc.On("GetKeys").Return(mocks.NewMockKeys(t))
	mockSvc.On("GetAlbySvc").Return(mocks.NewMockAlbyService(t))
	mockSvc.On("GetAlbyOAuthSvc").Return(mocks.NewMockAlbyOAuthService(t))

	httpSvc := NewHttpService(mockSvc, mockEventPublisher)
	httpSvc.RegisterSharedRoutes(e)

	jsonBody, _ := json.Marshal(api.UnlockRequest{UnlockPassword: "wrong", Permission: "full"})

	send := func(forwardedFor string) int {
		req := httptest.NewRequest(http.MethodPost, "/api/unlock", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", forwardedFor)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec.Code
	}

	rateLimited := 0
	for i := 0; i < 12; i++ {
		// each request presents a distinct client address
		if send("10.0.0."+strconv.Itoa(i)) == http.StatusTooManyRequests {
			rateLimited++
		}
	}

	assert.Positive(t, rateLimited, "spoofing X-Forwarded-For must not grant a fresh rate-limit bucket")
}

func TestGetApps_NoToken(t *testing.T) {
	e := echo.New()
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	mockSvc := mocks.NewMockService(t)
	gormDb, err := db.NewDB(t)
	require.NoError(t, err)
	defer db.CloseDB(gormDb)

	mockEventPublisher := mocks.NewMockEventPublisher(t)

	mockConfig := mocks.NewMockConfig(t)
	mockConfig.On("GetEnv").Return(&config.AppConfig{})

	mockSvc.On("GetDB").Return(gormDb)
	mockSvc.On("GetConfig").Return(mockConfig)
	mockSvc.On("GetKeys").Return(mocks.NewMockKeys(t))
	mockSvc.On("GetAlbySvc").Return(mocks.NewMockAlbyService(t))
	mockSvc.On("GetAlbyOAuthSvc").Return(mocks.NewMockAlbyOAuthService(t))

	httpSvc := NewHttpService(mockSvc, mockEventPublisher)
	httpSvc.RegisterSharedRoutes(e)

	req := httptest.NewRequest(http.MethodGet, "/api/apps", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestUnlock_NodeNotStarted(t *testing.T) {
	e := echo.New()
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	mockSvc := mocks.NewMockService(t)
	gormDb, err := db.NewDB(t)
	require.NoError(t, err)
	defer db.CloseDB(gormDb)

	mockEventPublisher := events.NewEventPublisher()

	mockConfig := mocks.NewMockConfig(t)
	mockConfig.On("GetEnv").Return(&config.AppConfig{})
	mockConfig.On("CheckUnlockPassword", "123").Return(true)

	mockSvc.On("GetDB").Return(gormDb)
	mockSvc.On("GetConfig").Return(mockConfig)
	mockSvc.On("GetKeys").Return(mocks.NewMockKeys(t))
	mockSvc.On("GetAlbySvc").Return(mocks.NewMockAlbyService(t))
	mockSvc.On("GetAlbyOAuthSvc").Return(mocks.NewMockAlbyOAuthService(t))
	mockSvc.On("GetLNClient").Return(nil)

	httpSvc := NewHttpService(mockSvc, mockEventPublisher)
	httpSvc.RegisterSharedRoutes(e)

	requestBody := api.UnlockRequest{UnlockPassword: "123", Permission: "full"}
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/unlock", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	body, err := io.ReadAll(rec.Body)
	require.NoError(t, err)

	var response ErrorResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	assert.Equal(t, "Node is not running, start it before unlocking.", response.Message)
	mockConfig.AssertNotCalled(t, "GetJWTSecret")
}

func TestGetApps_ReadonlyPermission(t *testing.T) {
	e := echo.New()
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	mockSvc := mocks.NewMockService(t)
	gormDb, err := db.NewDB(t)
	require.NoError(t, err)
	defer db.CloseDB(gormDb)

	mockEventPublisher := events.NewEventPublisher()

	mockConfig := mocks.NewMockConfig(t)
	mockConfig.On("GetEnv").Return(&config.AppConfig{})
	mockConfig.On("CheckUnlockPassword", "123").Return(true)
	mockConfig.On("GetJWTSecret").Return("dummy secret", nil)

	mockSvc.On("GetDB").Return(gormDb)
	mockSvc.On("GetConfig").Return(mockConfig)
	mockSvc.On("GetKeys").Return(mocks.NewMockKeys(t))
	mockSvc.On("GetAlbySvc").Return(mocks.NewMockAlbyService(t))
	mockSvc.On("GetAlbyOAuthSvc").Return(mocks.NewMockAlbyOAuthService(t))
	lnClient := mocks.NewMockLNClient(t)
	lnClient.On("GetNodeStatus", mock.Anything).Return(&lnclient.NodeStatus{}, nil)
	mockSvc.On("GetLNClient").Return(lnClient)

	httpSvc := NewHttpService(mockSvc, mockEventPublisher)
	httpSvc.RegisterSharedRoutes(e)

	requestBody := api.UnlockRequest{UnlockPassword: "123", Permission: "readonly"}
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/unlock", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json") // Set Content-Type header
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Result().Cookies(), "API token minting must not replace the browser session")

	body, err := io.ReadAll(rec.Body)
	require.NoError(t, err)

	type authTokenResponse struct {
		Token string `json:"token"`
	}

	var unlockAuthTokenResponse authTokenResponse
	err = json.Unmarshal(body, &unlockAuthTokenResponse)
	require.NoError(t, err)
	assert.NotEmpty(t, unlockAuthTokenResponse.Token)

	req2 := httptest.NewRequest(http.MethodGet, "/api/apps", nil)
	req2.Header.Set("Authorization", "Bearer "+unlockAuthTokenResponse.Token)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestGetApps_FullPermission(t *testing.T) {
	e := echo.New()
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	mockSvc := mocks.NewMockService(t)
	gormDb, err := db.NewDB(t)
	require.NoError(t, err)
	defer db.CloseDB(gormDb)

	mockEventPublisher := events.NewEventPublisher()

	mockConfig := mocks.NewMockConfig(t)
	mockConfig.On("GetEnv").Return(&config.AppConfig{})
	mockConfig.On("CheckUnlockPassword", "123").Return(true)
	mockConfig.On("GetJWTSecret").Return("dummy secret", nil)

	mockSvc.On("GetDB").Return(gormDb)
	mockSvc.On("GetConfig").Return(mockConfig)
	mockSvc.On("GetKeys").Return(mocks.NewMockKeys(t))
	mockSvc.On("GetAlbySvc").Return(mocks.NewMockAlbyService(t))
	mockSvc.On("GetAlbyOAuthSvc").Return(mocks.NewMockAlbyOAuthService(t))
	lnClient := mocks.NewMockLNClient(t)
	lnClient.On("GetNodeStatus", mock.Anything).Return(&lnclient.NodeStatus{}, nil)
	mockSvc.On("GetLNClient").Return(lnClient)

	httpSvc := NewHttpService(mockSvc, mockEventPublisher)
	httpSvc.RegisterSharedRoutes(e)

	requestBody := api.UnlockRequest{UnlockPassword: "123", Permission: "full", Session: true}
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/unlock", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json") // Set Content-Type header
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Body.String(), "browser session JWT must not be exposed in the response body")

	var sessionCookie *http.Cookie
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == sessionCookieName {
			sessionCookie = cookie
			break
		}
	}
	require.NotNil(t, sessionCookie)

	req2 := httptest.NewRequest(http.MethodGet, "/api/apps", nil)
	req2.AddCookie(sessionCookie)
	req2.Header.Set("Authorization", "Basic dXNlcjpwYXNzd29yZA==")
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestCreateApp_NoToken(t *testing.T) {
	e := echo.New()
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	mockSvc := mocks.NewMockService(t)
	gormDb, err := db.NewDB(t)
	require.NoError(t, err)
	defer db.CloseDB(gormDb)

	mockEventPublisher := mocks.NewMockEventPublisher(t)

	mockConfig := mocks.NewMockConfig(t)
	mockConfig.On("GetEnv").Return(&config.AppConfig{})

	mockSvc.On("GetDB").Return(gormDb)
	mockSvc.On("GetConfig").Return(mockConfig)
	mockSvc.On("GetKeys").Return(mocks.NewMockKeys(t))
	mockSvc.On("GetAlbySvc").Return(mocks.NewMockAlbyService(t))
	mockSvc.On("GetAlbyOAuthSvc").Return(mocks.NewMockAlbyOAuthService(t))

	httpSvc := NewHttpService(mockSvc, mockEventPublisher)
	httpSvc.RegisterSharedRoutes(e)

	requestBody := api.CreateAppRequest{Name: "Test app", Scopes: []string{constants.PAY_INVOICE_SCOPE}}
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/apps", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json") // Set Content-Type header
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCreateApp_FullPermission(t *testing.T) {
	e := echo.New()
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	mockSvc := mocks.NewMockService(t)
	gormDb, err := db.NewDB(t)
	require.NoError(t, err)
	defer db.CloseDB(gormDb)

	mockEventPublisher := events.NewEventPublisher()

	mockConfig := mocks.NewMockConfig(t)
	mockConfig.On("GetEnv").Return(&config.AppConfig{})
	mockConfig.On("CheckUnlockPassword", "123").Return(true)
	mockConfig.On("GetJWTSecret").Return("dummy secret", nil)
	mockConfig.On("GetRelayUrls").Return([]string{})

	mockKeys := mocks.NewMockKeys(t)
	mockKeys.On("GetAppWalletKey", uint(1)).Return("", nil)

	mockAlbyOAuthService := mocks.NewMockAlbyOAuthService(t)
	mockAlbyOAuthService.On("GetLightningAddress").Return("", nil)

	mockSvc.On("GetDB").Return(gormDb)
	mockSvc.On("GetConfig").Return(mockConfig)
	mockSvc.On("GetKeys").Return(mockKeys)
	mockSvc.On("GetAlbySvc").Return(mocks.NewMockAlbyService(t))
	mockSvc.On("GetAlbyOAuthSvc").Return(mockAlbyOAuthService)
	lnClient := mocks.NewMockLNClient(t)
	lnClient.On("GetNodeStatus", mock.Anything).Return(&lnclient.NodeStatus{}, nil)
	mockSvc.On("GetLNClient").Return(lnClient)

	httpSvc := NewHttpService(mockSvc, mockEventPublisher)
	httpSvc.RegisterSharedRoutes(e)

	requestBody := api.UnlockRequest{UnlockPassword: "123", Permission: "full"}
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/unlock", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json") // Set Content-Type header
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	body, err := io.ReadAll(rec.Body)
	require.NoError(t, err)

	type authTokenResponse struct {
		Token string `json:"token"`
	}

	var unlockAuthTokenResponse authTokenResponse
	err = json.Unmarshal(body, &unlockAuthTokenResponse)
	require.NoError(t, err)
	assert.NotEmpty(t, unlockAuthTokenResponse.Token)

	requestBody2 := api.CreateAppRequest{Name: "Test app", Scopes: []string{constants.PAY_INVOICE_SCOPE}}
	jsonBody2, _ := json.Marshal(requestBody2)
	req2 := httptest.NewRequest(http.MethodPost, "/api/apps", bytes.NewBuffer(jsonBody2))
	req2.Header.Set("Authorization", "Bearer "+unlockAuthTokenResponse.Token)
	req2.Header.Set("Content-Type", "application/json") // Set Content-Type header

	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestCreateApp_ReadonlyPermission(t *testing.T) {
	e := echo.New()
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	mockSvc := mocks.NewMockService(t)
	gormDb, err := db.NewDB(t)
	require.NoError(t, err)
	defer db.CloseDB(gormDb)

	mockEventPublisher := events.NewEventPublisher()

	mockConfig := mocks.NewMockConfig(t)
	mockConfig.On("GetEnv").Return(&config.AppConfig{})
	mockConfig.On("CheckUnlockPassword", "123").Return(true)
	mockConfig.On("GetJWTSecret").Return("dummy secret", nil)

	mockKeys := mocks.NewMockKeys(t)

	mockAlbyOAuthService := mocks.NewMockAlbyOAuthService(t)

	mockSvc.On("GetDB").Return(gormDb)
	mockSvc.On("GetConfig").Return(mockConfig)
	mockSvc.On("GetKeys").Return(mockKeys)
	mockSvc.On("GetAlbySvc").Return(mocks.NewMockAlbyService(t))
	mockSvc.On("GetAlbyOAuthSvc").Return(mockAlbyOAuthService)
	lnClient := mocks.NewMockLNClient(t)
	lnClient.On("GetNodeStatus", mock.Anything).Return(&lnclient.NodeStatus{}, nil)
	mockSvc.On("GetLNClient").Return(lnClient)

	httpSvc := NewHttpService(mockSvc, mockEventPublisher)
	httpSvc.RegisterSharedRoutes(e)

	requestBody := api.UnlockRequest{UnlockPassword: "123", Permission: "readonly"}
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/unlock", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json") // Set Content-Type header
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	body, err := io.ReadAll(rec.Body)
	require.NoError(t, err)

	type authTokenResponse struct {
		Token string `json:"token"`
	}

	var unlockAuthTokenResponse authTokenResponse
	err = json.Unmarshal(body, &unlockAuthTokenResponse)
	require.NoError(t, err)
	assert.NotEmpty(t, unlockAuthTokenResponse.Token)

	requestBody2 := api.CreateAppRequest{Name: "Test app", Scopes: []string{constants.PAY_INVOICE_SCOPE}}
	jsonBody2, _ := json.Marshal(requestBody2)
	req2 := httptest.NewRequest(http.MethodPost, "/api/apps", bytes.NewBuffer(jsonBody2))
	req2.Header.Set("Authorization", "Bearer "+unlockAuthTokenResponse.Token)
	req2.Header.Set("Content-Type", "application/json") // Set Content-Type header

	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusForbidden, rec2.Code)
}

func TestGetLogOutput_ReadonlyPermission(t *testing.T) {
	e := echo.New()
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	mockSvc := mocks.NewMockService(t)
	gormDb, err := db.NewDB(t)
	require.NoError(t, err)
	defer db.CloseDB(gormDb)

	mockEventPublisher := events.NewEventPublisher()

	mockConfig := mocks.NewMockConfig(t)
	mockConfig.On("GetEnv").Return(&config.AppConfig{})
	mockConfig.On("CheckUnlockPassword", "123").Return(true)
	mockConfig.On("GetJWTSecret").Return("dummy secret", nil)

	mockSvc.On("GetDB").Return(gormDb)
	mockSvc.On("GetConfig").Return(mockConfig)
	mockSvc.On("GetKeys").Return(mocks.NewMockKeys(t))
	mockSvc.On("GetAlbySvc").Return(mocks.NewMockAlbyService(t))
	mockSvc.On("GetAlbyOAuthSvc").Return(mocks.NewMockAlbyOAuthService(t))
	lnClient := mocks.NewMockLNClient(t)
	lnClient.On("GetNodeStatus", mock.Anything).Return(&lnclient.NodeStatus{}, nil)
	mockSvc.On("GetLNClient").Return(lnClient)

	httpSvc := NewHttpService(mockSvc, mockEventPublisher)
	httpSvc.RegisterSharedRoutes(e)

	requestBody := api.UnlockRequest{UnlockPassword: "123", Permission: "readonly"}
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/unlock", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json") // Set Content-Type header
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	body, err := io.ReadAll(rec.Body)
	require.NoError(t, err)

	type authTokenResponse struct {
		Token string `json:"token"`
	}

	var unlockAuthTokenResponse authTokenResponse
	err = json.Unmarshal(body, &unlockAuthTokenResponse)
	require.NoError(t, err)
	assert.NotEmpty(t, unlockAuthTokenResponse.Token)

	req2 := httptest.NewRequest(http.MethodGet, "/api/log/app", nil)
	req2.Header.Set("Authorization", "Bearer "+unlockAuthTokenResponse.Token)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusForbidden, rec2.Code)
}

func TestGetLogOutput_FullPermission(t *testing.T) {
	e := echo.New()
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	mockSvc := mocks.NewMockService(t)
	gormDb, err := db.NewDB(t)
	require.NoError(t, err)
	defer db.CloseDB(gormDb)

	mockEventPublisher := events.NewEventPublisher()

	mockConfig := mocks.NewMockConfig(t)
	mockConfig.On("GetEnv").Return(&config.AppConfig{})
	mockConfig.On("CheckUnlockPassword", "123").Return(true)
	mockConfig.On("GetJWTSecret").Return("dummy secret", nil)

	mockSvc.On("GetDB").Return(gormDb)
	mockSvc.On("GetConfig").Return(mockConfig)
	mockSvc.On("GetKeys").Return(mocks.NewMockKeys(t))
	mockSvc.On("GetAlbySvc").Return(mocks.NewMockAlbyService(t))
	mockSvc.On("GetAlbyOAuthSvc").Return(mocks.NewMockAlbyOAuthService(t))
	lnClient := mocks.NewMockLNClient(t)
	lnClient.On("GetNodeStatus", mock.Anything).Return(&lnclient.NodeStatus{}, nil)
	mockSvc.On("GetLNClient").Return(lnClient)

	httpSvc := NewHttpService(mockSvc, mockEventPublisher)
	httpSvc.RegisterSharedRoutes(e)

	requestBody := api.UnlockRequest{UnlockPassword: "123", Permission: "full"}
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/unlock", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json") // Set Content-Type header
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	body, err := io.ReadAll(rec.Body)
	require.NoError(t, err)

	type authTokenResponse struct {
		Token string `json:"token"`
	}

	var unlockAuthTokenResponse authTokenResponse
	err = json.Unmarshal(body, &unlockAuthTokenResponse)
	require.NoError(t, err)
	assert.NotEmpty(t, unlockAuthTokenResponse.Token)

	req2 := httptest.NewRequest(http.MethodGet, "/api/log/app", nil)
	req2.Header.Set("Authorization", "Bearer "+unlockAuthTokenResponse.Token)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusOK, rec2.Code)

	var logResponse api.GetLogOutputResponse
	err = json.Unmarshal(rec2.Body.Bytes(), &logResponse)
	require.NoError(t, err)
}
