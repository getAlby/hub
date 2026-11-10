package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/alby"
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/tests/mocks"
)

const (
	testOAuthState    = "0000000000000000000000000000000000000000000000000000000000000001"
	testOAuthVerifier = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAE"
)

func TestAlbyAuthUrl_SetsStateCookie(t *testing.T) {
	e, mockAlbyOAuthSvc := setupAlbyHttpService(t)
	mockAlbyOAuthSvc.On("GetAuthUrl", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(func(state string, verifier string) string {
		return "https://getalby.com/oauth?state=" + state + "&verifier=" + verifier
	})

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/alby/auth-url", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	stateCookie := findOAuthCookie(t, rec, albyOAuthStateCookieName)
	verifierCookie := findOAuthCookie(t, rec, albyOAuthVerifierCookieName)
	assert.True(t, isValidOAuthState(stateCookie.Value))
	assert.True(t, isValidOAuthVerifier(verifierCookie.Value))
	for _, cookie := range []*http.Cookie{stateCookie, verifierCookie} {
		assert.True(t, cookie.HttpOnly)
		assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	}

	var response alby.AlbyAuthUrlResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	assert.Equal(t, "https://getalby.com/oauth?state="+stateCookie.Value+"&verifier="+verifierCookie.Value, response.Url)
}

func TestAlbyAuthUrl_ReusesExistingState(t *testing.T) {
	e, mockAlbyOAuthSvc := setupAlbyHttpService(t)
	mockAlbyOAuthSvc.On("GetAuthUrl", testOAuthState, testOAuthVerifier).Return("https://getalby.com/oauth?state=" + testOAuthState)

	req := httptest.NewRequest(http.MethodGet, "/api/alby/auth-url", nil)
	addOAuthCookies(req, testOAuthState, testOAuthVerifier)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, testOAuthState, findOAuthCookie(t, rec, albyOAuthStateCookieName).Value)
	assert.Equal(t, testOAuthVerifier, findOAuthCookie(t, rec, albyOAuthVerifierCookieName).Value)
}

func TestAlbyAuthUrl_RegeneratesStateWithoutVerifier(t *testing.T) {
	e, mockAlbyOAuthSvc := setupAlbyHttpService(t)
	mockAlbyOAuthSvc.On("GetAuthUrl", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return("https://getalby.com/oauth")

	req := httptest.NewRequest(http.MethodGet, "/api/alby/auth-url", nil)
	req.AddCookie(&http.Cookie{Name: albyOAuthStateCookieName, Value: testOAuthState})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.NotEqual(t, testOAuthState, findOAuthCookie(t, rec, albyOAuthStateCookieName).Value)
	assert.True(t, isValidOAuthVerifier(findOAuthCookie(t, rec, albyOAuthVerifierCookieName).Value))
}

func TestAlbyCallback_RejectsInvalidState(t *testing.T) {
	testCases := []struct {
		name     string
		query    string
		state    string
		verifier string
	}{
		{name: "no state", query: "code=abc", state: testOAuthState, verifier: testOAuthVerifier},
		{name: "no cookies", query: "code=abc&state=" + testOAuthState},
		{name: "mismatched state", query: "code=abc&state=" + strings.Repeat("f", 64), state: testOAuthState, verifier: testOAuthVerifier},
		{name: "placeholder state", query: "code=abc&state=unused", state: "unused", verifier: testOAuthVerifier},
		{name: "no verifier", query: "code=abc&state=" + testOAuthState, state: testOAuthState},
		{name: "invalid verifier", query: "code=abc&state=" + testOAuthState, state: testOAuthState, verifier: "short"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e, mockAlbyOAuthSvc := setupAlbyHttpService(t)

			req := httptest.NewRequest(http.MethodGet, "/api/alby/callback?"+tc.query, nil)
			addOAuthCookies(req, tc.state, tc.verifier)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
			mockAlbyOAuthSvc.AssertNotCalled(t, "CallbackHandler", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func TestAlbyCallback_ValidState(t *testing.T) {
	e, mockAlbyOAuthSvc := setupAlbyHttpService(t)
	mockAlbyOAuthSvc.On("CallbackHandler", mock.Anything, "abc", testOAuthVerifier).Return(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/alby/callback?code=abc&state="+testOAuthState, nil)
	addOAuthCookies(req, testOAuthState, testOAuthVerifier)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, -1, findOAuthCookie(t, rec, albyOAuthStateCookieName).MaxAge)
	assert.Equal(t, -1, findOAuthCookie(t, rec, albyOAuthVerifierCookieName).MaxAge)
}

func TestAlbyCallback_KeepsStateWhenExchangeFails(t *testing.T) {
	e, mockAlbyOAuthSvc := setupAlbyHttpService(t)
	mockAlbyOAuthSvc.On("CallbackHandler", mock.Anything, "abc", testOAuthVerifier).Return(assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/api/alby/callback?code=abc&state="+testOAuthState, nil)
	addOAuthCookies(req, testOAuthState, testOAuthVerifier)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Empty(t, rec.Result().Cookies())
}

func setupAlbyHttpService(t *testing.T) (*echo.Echo, *mocks.MockAlbyOAuthService) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	e := echo.New()
	mockAlbyOAuthSvc := mocks.NewMockAlbyOAuthService(t)
	appConfig := &config.AppConfig{AlbyClientId: "J2PbXS1yOf"}
	albyHttpSvc := NewAlbyHttpService(mocks.NewMockService(t), mocks.NewMockAlbyService(t), mockAlbyOAuthSvc, appConfig)
	apiGroup := e.Group("/api")
	albyHttpSvc.RegisterSharedRoutes(apiGroup, apiGroup, e, nil)
	return e, mockAlbyOAuthSvc
}

func addOAuthCookies(req *http.Request, state string, verifier string) {
	if state != "" {
		req.AddCookie(&http.Cookie{Name: albyOAuthStateCookieName, Value: state})
	}
	if verifier != "" {
		req.AddCookie(&http.Cookie{Name: albyOAuthVerifierCookieName, Value: verifier})
	}
}

func findOAuthCookie(t *testing.T, rec *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	require.FailNow(t, "OAuth cookie not set", name)
	return nil
}
