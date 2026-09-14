package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/getAlby/hub/api"
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/tests/db"
	"github.com/getAlby/hub/tests/mocks"
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

	requestBody := api.StartRequest{UnlockPassword: "123"}
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
// it proves that issuing a token from the start endpoint requires
// LoadJWTSecret, GetJWTSecret and StartApp, so the AssertNotCalled checks in
// the negative test cannot silently become vacuous if the implementation
// changes. It also proves the issued token grants full access.
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

	requestBody := api.StartRequest{UnlockPassword: "123"}
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/start", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("StartApp was not called")
	}

	var tokenResponse authTokenResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &tokenResponse))
	require.NotEmpty(t, tokenResponse.Token)
	// the same detector used by the negative test must recognise a real token.
	// It is both returned in the body and set as the session cookie.
	assert.Equal(t, []string{tokenResponse.Token, tokenResponse.Token}, findJWTs(rec))

	mockConfig.AssertCalled(t, "LoadJWTSecret", "123")
	mockConfig.AssertCalled(t, "GetJWTSecret")

	// the session cookie grants full access, even while the Authorization
	// header carries HTTP Basic Authentication credentials for a reverse proxy
	req2 := httptest.NewRequest(http.MethodGet, "/api/apps", nil)
	req2.AddCookie(sessionCookie(t, rec))
	req2.Header.Set("Authorization", "Basic dXNlcjpwYXNzd29yZA==")
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code)
	assert.Equal(t, "private, no-store", rec2.Header().Get(echo.HeaderCacheControl))

	// the standard Authorization header keeps working for external API clients
	req3 := httptest.NewRequest(http.MethodGet, "/api/apps", nil)
	req3.Header.Set("Authorization", "Bearer "+tokenResponse.Token)
	rec3 := httptest.NewRecorder()
	e.ServeHTTP(rec3, req3)
	assert.Equal(t, http.StatusOK, rec3.Code)
}

// sessionCookie returns the session cookie set on a response, failing the test
// if there is none.
func sessionCookie(t *testing.T, rec *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()

	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == sessionCookieName {
			return cookie
		}
	}

	t.Fatalf("no %q cookie was set", sessionCookieName)
	return nil
}

// hasSessionCookie reports whether a response sets a non-empty session cookie.
func hasSessionCookie(rec *httptest.ResponseRecorder) bool {
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == sessionCookieName && cookie.Value != "" {
			return true
		}
	}
	return false
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

	req2 := httptest.NewRequest(http.MethodGet, "/api/apps", nil)
	req2.Header.Set("Authorization", "Bearer "+unlockAuthTokenResponse.Token)
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

// registerSessionTestRoutes wires up the routes with the mocks needed to log in
// and then serve authenticated API requests. Expectations are optional so that
// tests exercising only part of the flow do not fail on unmet calls.
func registerSessionTestRoutes(t *testing.T, e *echo.Echo) {
	t.Helper()

	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))

	gormDb, err := db.NewDB(t)
	require.NoError(t, err)
	t.Cleanup(func() { db.CloseDB(gormDb) })

	mockConfig := mocks.NewMockConfig(t)
	mockConfig.On("GetEnv").Return(&config.AppConfig{}).Maybe()
	mockConfig.On("CheckUnlockPassword", "123").Return(true).Maybe()
	mockConfig.On("GetJWTSecret").Return("dummy secret", nil).Maybe()

	lnClient := mocks.NewMockLNClient(t)
	lnClient.On("GetNodeStatus", mock.Anything).Return(&lnclient.NodeStatus{}, nil).Maybe()

	mockSvc := mocks.NewMockService(t)
	mockSvc.On("GetDB").Return(gormDb)
	mockSvc.On("GetConfig").Return(mockConfig)
	mockSvc.On("GetKeys").Return(mocks.NewMockKeys(t))
	mockSvc.On("GetAlbySvc").Return(mocks.NewMockAlbyService(t))
	mockSvc.On("GetAlbyOAuthSvc").Return(mocks.NewMockAlbyOAuthService(t))
	mockSvc.On("GetLNClient").Return(lnClient).Maybe()

	httpSvc := NewHttpService(mockSvc, events.NewEventPublisher())
	httpSvc.RegisterSharedRoutes(e)
}

// login posts to the unlock endpoint and returns the response.
func login(t *testing.T, e *echo.Echo, unlockRequest api.UnlockRequest, requestHeaders http.Header) *httptest.ResponseRecorder {
	t.Helper()

	jsonBody, err := json.Marshal(unlockRequest)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/unlock", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	for name, values := range requestHeaders {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	return rec
}

// TestUnlock_SessionCoexistsWithBasicAuth is the reason the session lives in a
// cookie: a browser sends a single Authorization header, so a Hub bearer token
// set from JavaScript would replace the HTTP Basic Authentication credentials a
// reverse proxy asked the user for.
func TestUnlock_SessionCoexistsWithBasicAuth(t *testing.T) {
	e := echo.New()
	registerSessionTestRoutes(t, e)

	rec := login(t, e, api.UnlockRequest{UnlockPassword: "123", Permission: "full"}, nil)

	cookie := sessionCookie(t, rec)
	assert.True(t, cookie.HttpOnly, "the session cookie must not be readable from JavaScript")
	assert.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	assert.Equal(t, "/", cookie.Path)
	assert.False(t, cookie.Secure, "a plain HTTP request must not get a cookie the browser would drop")

	req := httptest.NewRequest(http.MethodGet, "/api/apps", nil)
	req.AddCookie(cookie)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNzd29yZA==")
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req)

	assert.Equal(t, http.StatusOK, rec2.Code)
}

// TestUnlock_SecureCookieBehindHTTPSProxy covers the Secure flag being decided
// by the forwarded scheme, since the flag would make browsers drop the cookie
// on the plain HTTP setups Alby Hub also supports.
func TestUnlock_SecureCookieBehindHTTPSProxy(t *testing.T) {
	e := echo.New()
	registerSessionTestRoutes(t, e)

	rec := login(t, e, api.UnlockRequest{UnlockPassword: "123", Permission: "full"},
		http.Header{"X-Forwarded-Proto": {"https"}})

	assert.True(t, sessionCookie(t, rec).Secure)
}

// TestUnlock_CreateApiTokenDoesNotStartSession covers the unlock endpoint's
// second job: minting a token for external API clients. Starting a session
// there would let creating a readonly token downgrade the caller's own session.
func TestUnlock_CreateApiTokenDoesNotStartSession(t *testing.T) {
	e := echo.New()
	registerSessionTestRoutes(t, e)

	rec := login(t, e, api.UnlockRequest{
		UnlockPassword: "123",
		Permission:     "readonly",
		CreateApiToken: true,
	}, nil)

	assert.False(t, hasSessionCookie(rec), "minting an API token must not start a session")

	var tokenResponse authTokenResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &tokenResponse))
	require.NotEmpty(t, tokenResponse.Token)

	// the token is still usable by an external client through Authorization
	req := httptest.NewRequest(http.MethodGet, "/api/apps", nil)
	req.Header.Set("Authorization", "Bearer "+tokenResponse.Token)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req)

	assert.Equal(t, http.StatusOK, rec2.Code)
}

func TestLogout_ClearsSessionCookie(t *testing.T) {
	e := echo.New()
	registerSessionTestRoutes(t, e)

	rec := login(t, e, api.UnlockRequest{UnlockPassword: "123", Permission: "full"}, nil)
	cookie := sessionCookie(t, rec)

	req := httptest.NewRequest(http.MethodGet, "/logout", nil)
	req.AddCookie(cookie)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req)

	require.Equal(t, http.StatusFound, rec2.Code)
	clearedCookie := sessionCookie(t, rec2)
	assert.Empty(t, clearedCookie.Value)
	assert.Less(t, clearedCookie.MaxAge, 0, "the cookie must be expired, not just emptied")
	// the attributes have to match the ones used when setting the cookie,
	// otherwise the browser keeps the original
	assert.Equal(t, cookie.Path, clearedCookie.Path)
	assert.Equal(t, cookie.SameSite, clearedCookie.SameSite)
	assert.Equal(t, cookie.Secure, clearedCookie.Secure)
}

type infoAPIStub struct {
	api.API
}

func (infoAPIStub) GetInfo(context.Context) (*api.InfoResponse, error) {
	return &api.InfoResponse{}, nil
}

// TestInfo_RecognizesSessions covers how the frontend learns whether it is
// logged in. The info endpoint is unauthenticated, so it inspects the request
// itself and has to accept the same credentials as the JWT middleware.
func TestInfo_RecognizesSessions(t *testing.T) {
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
			name:         "session cookie",
			cookies:      []*http.Cookie{{Name: sessionCookieName, Value: token}},
			wantUnlocked: true,
		},
		{
			name:         "session cookie alongside HTTP Basic Authentication",
			cookies:      []*http.Cookie{{Name: sessionCookieName, Value: token}},
			headers:      http.Header{"Authorization": {"Basic dXNlcjpwYXNzd29yZA=="}},
			wantUnlocked: true,
		},
		{
			name:         "Authorization header for external API clients",
			headers:      http.Header{"Authorization": {bearerPrefix + token}},
			wantUnlocked: true,
		},
		{
			name:         "expired session cookie falls back to the Authorization header",
			cookies:      []*http.Cookie{{Name: sessionCookieName, Value: "invalid"}},
			headers:      http.Header{"Authorization": {bearerPrefix + token}},
			wantUnlocked: true,
		},
		{
			name:         "HTTP Basic Authentication alone does not unlock the hub",
			headers:      http.Header{"Authorization": {"Basic dXNlcjpwYXNzd29yZA=="}},
			wantUnlocked: false,
		},
		{
			name:         "malformed bearer token",
			headers:      http.Header{"Authorization": {"Bearer"}},
			wantUnlocked: false,
		},
		{
			name:         "no credentials",
			wantUnlocked: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger.Init(strconv.Itoa(int(logrus.DebugLevel)))

			mockConfig := mocks.NewMockConfig(t)
			mockConfig.On("GetJWTSecret").Return(jwtSecret, nil).Maybe()
			httpSvc := &HttpService{api: infoAPIStub{}, cfg: mockConfig}

			e := echo.New()
			e.GET("/api/info", httpSvc.infoHandler)

			req := httptest.NewRequest(http.MethodGet, "/api/info", nil)
			for name, values := range tt.headers {
				for _, value := range values {
					req.Header.Add(name, value)
				}
			}
			for _, cookie := range tt.cookies {
				req.AddCookie(cookie)
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
