package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
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
	// the same detector used by the negative test must recognise a real token
	assert.Equal(t, []string{tokenResponse.Token}, findJWTs(rec))

	mockConfig.AssertCalled(t, "LoadJWTSecret", "123")
	mockConfig.AssertCalled(t, "GetJWTSecret")

	// the token grants full access
	req2 := httptest.NewRequest(http.MethodGet, "/api/apps", nil)
	req2.Header.Set("Authorization", "Bearer "+tokenResponse.Token)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code)
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
