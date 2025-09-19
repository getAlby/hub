package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/getAlby/hub/api"
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/tests/db"
	"github.com/getAlby/hub/tests/mocks"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, http.StatusBadRequest, rec.Code)
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
	mockConfig.On("GetJWTSecret").Return("dummy secret")

	mockSvc.On("GetDB").Return(gormDb)
	mockSvc.On("GetConfig").Return(mockConfig)
	mockSvc.On("GetKeys").Return(mocks.NewMockKeys(t))
	mockSvc.On("GetAlbySvc").Return(mocks.NewMockAlbyService(t))
	mockSvc.On("GetAlbyOAuthSvc").Return(mocks.NewMockAlbyOAuthService(t))

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
	mockConfig.On("GetJWTSecret").Return("dummy secret")

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

	assert.Equal(t, http.StatusBadRequest, rec.Code)
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
	mockConfig.On("GetJWTSecret").Return("dummy secret")
	mockConfig.On("GetRelayUrl").Return("")

	mockKeys := mocks.NewMockKeys(t)
	mockKeys.On("GetAppWalletKey", uint(1)).Return("", nil)

	mockAlbyOAuthService := mocks.NewMockAlbyOAuthService(t)
	mockAlbyOAuthService.On("GetLightningAddress").Return("", nil)

	mockSvc.On("GetDB").Return(gormDb)
	mockSvc.On("GetConfig").Return(mockConfig)
	mockSvc.On("GetKeys").Return(mockKeys)
	mockSvc.On("GetAlbySvc").Return(mocks.NewMockAlbyService(t))
	mockSvc.On("GetAlbyOAuthSvc").Return(mockAlbyOAuthService)

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
	mockConfig.On("GetJWTSecret").Return("dummy secret")

	mockKeys := mocks.NewMockKeys(t)

	mockAlbyOAuthService := mocks.NewMockAlbyOAuthService(t)

	mockSvc.On("GetDB").Return(gormDb)
	mockSvc.On("GetConfig").Return(mockConfig)
	mockSvc.On("GetKeys").Return(mockKeys)
	mockSvc.On("GetAlbySvc").Return(mocks.NewMockAlbyService(t))
	mockSvc.On("GetAlbyOAuthSvc").Return(mockAlbyOAuthService)

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
