package http

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/tests/db"
	"github.com/getAlby/hub/tests/mocks"
)

func TestSetupRoutes_RequireSetupCode(t *testing.T) {
	routes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/setup"},
		{http.MethodPost, "/api/setup/verify-code"},
		{http.MethodPost, "/api/restore"},
		{http.MethodGet, "/api/alby/callback"},
	}
	setupCodes := map[string]string{
		"missing": "",
		"wrong":   "wrong-code",
	}

	for _, route := range routes {
		for name, setupCode := range setupCodes {
			t.Run(route.path+" "+name, func(t *testing.T) {
				mockConfig := mocks.NewMockConfig(t)
				mockConfig.On("GetEnv").Return(&config.AppConfig{SetupCode: "setup-code"})
				mockConfig.On("SetupCompleted").Return(false, nil)
				e := createSetupCodeTestEcho(t, mockConfig)

				req := httptest.NewRequest(route.method, route.path, bytes.NewBufferString(`{"unlockPassword":"123"}`))
				req.Header.Set("Content-Type", "application/json")
				if setupCode != "" {
					req.Header.Set(setupCodeHeader, setupCode)
				}
				rec := httptest.NewRecorder()
				e.ServeHTTP(rec, req)

				assert.Equal(t, http.StatusUnauthorized, rec.Code)
			})
		}
	}
}

func TestSetupRoutes_RateLimited(t *testing.T) {
	mockConfig := mocks.NewMockConfig(t)
	mockConfig.On("GetEnv").Return(&config.AppConfig{SetupCode: "setup-code"})
	mockConfig.On("SetupCompleted").Return(false, nil)
	e := createSetupCodeTestEcho(t, mockConfig)

	statusCodes := []int{}
	for range 4 {
		req := httptest.NewRequest(http.MethodPost, "/api/setup", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		statusCodes = append(statusCodes, rec.Code)
	}

	assert.Equal(t, []int{
		http.StatusUnauthorized,
		http.StatusUnauthorized,
		http.StatusUnauthorized,
		http.StatusTooManyRequests,
	}, statusCodes)
}

func TestVerifySetupCode_CorrectCode(t *testing.T) {
	mockConfig := mocks.NewMockConfig(t)
	mockConfig.On("GetEnv").Return(&config.AppConfig{SetupCode: "setup-code"})
	mockConfig.On("SetupCompleted").Return(false, nil)
	e := createSetupCodeTestEcho(t, mockConfig)

	req := httptest.NewRequest(http.MethodPost, "/api/setup/verify-code", nil)
	req.Header.Set(setupCodeHeader, "setup-code")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestRequireSetupCode(t *testing.T) {
	testCases := []struct {
		name            string
		configuredCode  string
		setupCompleted  bool
		providedCode    string
		expectedAllowed bool
	}{
		{name: "correct code", configuredCode: "setup-code", providedCode: "setup-code", expectedAllowed: true},
		{name: "missing code", configuredCode: "setup-code", providedCode: "", expectedAllowed: false},
		{name: "wrong code", configuredCode: "setup-code", providedCode: "setup-cod", expectedAllowed: false},
		{name: "no code configured", configuredCode: "", providedCode: "", expectedAllowed: true},
		{name: "setup completed", configuredCode: "setup-code", setupCompleted: true, providedCode: "", expectedAllowed: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockConfig := mocks.NewMockConfig(t)
			mockConfig.On("GetEnv").Return(&config.AppConfig{SetupCode: tc.configuredCode})
			if tc.configuredCode != "" {
				mockConfig.On("SetupCompleted").Return(tc.setupCompleted, nil)
			}
			httpSvc := &HttpService{cfg: mockConfig}

			handlerCalled := false
			handler := httpSvc.requireSetupCode(func(c echo.Context) error {
				handlerCalled = true
				return c.NoContent(http.StatusNoContent)
			})

			req := httptest.NewRequest(http.MethodPost, "/api/setup", nil)
			if tc.providedCode != "" {
				req.Header.Set(setupCodeHeader, tc.providedCode)
			}
			rec := httptest.NewRecorder()
			require.NoError(t, handler(echo.New().NewContext(req, rec)))

			assert.Equal(t, tc.expectedAllowed, handlerCalled)
			if tc.expectedAllowed {
				assert.Equal(t, http.StatusNoContent, rec.Code)
			} else {
				assert.Equal(t, http.StatusUnauthorized, rec.Code)
			}
		})
	}
}

func createSetupCodeTestEcho(t *testing.T, mockConfig *mocks.MockConfig) *echo.Echo {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	gormDb, err := db.NewDB(t)
	require.NoError(t, err)
	t.Cleanup(func() { db.CloseDB(gormDb) })

	mockSvc := mocks.NewMockService(t)
	mockSvc.On("GetDB").Return(gormDb)
	mockSvc.On("GetConfig").Return(mockConfig)
	mockSvc.On("GetKeys").Return(mocks.NewMockKeys(t))
	mockSvc.On("GetAlbySvc").Return(mocks.NewMockAlbyService(t))
	mockSvc.On("GetAlbyOAuthSvc").Return(mocks.NewMockAlbyOAuthService(t))

	e := echo.New()
	httpSvc := NewHttpService(mockSvc, events.NewEventPublisher())
	httpSvc.RegisterSharedRoutes(e)
	return e
}
