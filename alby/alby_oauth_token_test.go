package alby

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/logger"
)

type oauthTestConfig struct {
	values            map[string]string
	remainingFailures int
	failKeys          map[string]int
}

func newOAuthTestConfig() *oauthTestConfig {
	return &oauthTestConfig{values: map[string]string{}}
}

func (c *oauthTestConfig) Get(key string, _ string) (string, error) {
	return c.values[key], nil
}

func (c *oauthTestConfig) SetUpdate(key string, value string, _ string) error {
	if c.remainingFailures > 0 {
		c.remainingFailures--
		return errors.New("transient db error")
	}
	if c.failKeys != nil {
		if remaining, ok := c.failKeys[key]; ok && remaining > 0 {
			c.failKeys[key] = remaining - 1
			return errors.New("transient db error")
		}
	}
	c.values[key] = value
	return nil
}

func (c *oauthTestConfig) SetIgnore(string, string, string) error { return nil }
func (c *oauthTestConfig) LoadJWTSecret(string) error             { return nil }
func (c *oauthTestConfig) GetJWTSecret() (string, error)          { return "", nil }
func (c *oauthTestConfig) GetRelayUrls() []string                 { return nil }
func (c *oauthTestConfig) GetNetwork() string                     { return "" }
func (c *oauthTestConfig) GetMempoolUrl() string                  { return "" }
func (c *oauthTestConfig) GetEnv() *config.AppConfig {
	return &config.AppConfig{AlbyClientId: "test-client-id", AlbyClientSecret: "test-client-secret"}
}
func (c *oauthTestConfig) CheckUnlockPassword(string) bool           { return false }
func (c *oauthTestConfig) IsUnlockPasswordCheckSet() (bool, error)   { return false, nil }
func (c *oauthTestConfig) ChangeUnlockPassword(string, string) error { return nil }
func (c *oauthTestConfig) SetAutoUnlockPassword(string) error        { return nil }
func (c *oauthTestConfig) SaveUnlockPasswordCheck(string) error      { return nil }
func (c *oauthTestConfig) SetupCompleted() (bool, error)             { return true, nil }
func (c *oauthTestConfig) GetCurrency() string                       { return "" }
func (c *oauthTestConfig) SetCurrency(string) error                  { return nil }
func (c *oauthTestConfig) GetBitcoinDisplayFormat() string           { return "" }
func (c *oauthTestConfig) SetBitcoinDisplayFormat(string) error      { return nil }

func newOAuthTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	r1Consumed := false
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth/token" {
			http.NotFound(w, r)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if r.FormValue("grant_type") != "refresh_token" {
			http.Error(w, "unsupported grant", http.StatusBadRequest)
			return
		}

		refreshToken := r.FormValue("refresh_token")
		switch refreshToken {
		case "refresh-token-r1":
			if r1Consumed {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant"})
				return
			}
			r1Consumed = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token":  "access-token-a2",
				"token_type":    "bearer",
				"expires_in":    3600,
				"refresh_token": "refresh-token-r2",
			})
		case "refresh-token-r2":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token":  "access-token-a3",
				"token_type":    "bearer",
				"expires_in":    3600,
				"refresh_token": "refresh-token-r2",
			})
		default:
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant"})
		}
	}))
}

func seedExpiredOAuthTokens(t *testing.T, cfg config.Config) {
	t.Helper()
	expired := time.Now().Add(-2 * time.Hour).Unix()
	require.NoError(t, cfg.SetUpdate(accessTokenKey, "access-token-a1", ""))
	require.NoError(t, cfg.SetUpdate(accessTokenExpiryKey, strconv.FormatInt(expired, 10), ""))
	require.NoError(t, cfg.SetUpdate(refreshTokenKey, "refresh-token-r1", ""))
}

func newOAuthTestService(t *testing.T, cfg config.Config, tokenURL string) *albyOAuthService {
	t.Helper()
	svc := NewAlbyOAuthService(nil, cfg, nil, events.NewEventPublisher())
	svc.oauthConf = &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Endpoint: oauth2.Endpoint{
			TokenURL:  tokenURL,
			AuthStyle: oauth2.AuthStyleInHeader,
		},
	}
	return svc
}

func initOAuthTokenTest(t *testing.T) {
	t.Helper()
	logger.Init("4")
}

func TestFetchUserTokenSurvivesTransientRefreshTokenPersistenceFailure(t *testing.T) {
	initOAuthTokenTest(t)

	server := newOAuthTestServer(t)
	defer server.Close()

	cfg := newOAuthTestConfig()
	seedExpiredOAuthTokens(t, cfg)
	cfg.failKeys = map[string]int{refreshTokenKey: 10}

	oauthSvc := newOAuthTestService(t, cfg, server.URL+"/oauth/token")
	ctx := context.Background()

	firstToken, err := oauthSvc.fetchUserToken(ctx)
	require.NoError(t, err)
	require.Equal(t, "access-token-a2", firstToken.AccessToken)
	require.Equal(t, "refresh-token-r2", firstToken.RefreshToken)

	storedRefreshToken, err := cfg.Get(refreshTokenKey, "")
	require.NoError(t, err)
	assert.Equal(t, "refresh-token-r1", storedRefreshToken, "stale refresh token must remain in db after failed persistence")

	cfg.failKeys = nil
	secondToken, err := oauthSvc.fetchUserToken(ctx)
	require.NoError(t, err)
	require.Equal(t, "access-token-a2", secondToken.AccessToken)
	require.Equal(t, "refresh-token-r2", secondToken.RefreshToken)
	storedRefreshToken, err = cfg.Get(refreshTokenKey, "")
	require.NoError(t, err)
	require.Equal(t, "refresh-token-r2", storedRefreshToken)
}

func TestFetchUserTokenRetriesRefreshTokenPersistence(t *testing.T) {
	initOAuthTokenTest(t)

	server := newOAuthTestServer(t)
	defer server.Close()

	cfg := newOAuthTestConfig()
	seedExpiredOAuthTokens(t, cfg)
	cfg.failKeys = map[string]int{refreshTokenKey: 1}

	oauthSvc := newOAuthTestService(t, cfg, server.URL+"/oauth/token")
	ctx := context.Background()

	_, err := oauthSvc.fetchUserToken(ctx)
	require.NoError(t, err)

	storedRefreshToken, err := cfg.Get(refreshTokenKey, "")
	require.NoError(t, err)
	assert.Equal(t, "refresh-token-r2", storedRefreshToken)
}

func TestFetchUserTokenConcurrentRefreshUsesSingleRotatedToken(t *testing.T) {
	initOAuthTokenTest(t)

	server := newOAuthTestServer(t)
	defer server.Close()

	cfg := newOAuthTestConfig()
	seedExpiredOAuthTokens(t, cfg)

	oauthSvc := newOAuthTestService(t, cfg, server.URL+"/oauth/token")
	ctx := context.Background()

	const workers = 8
	results := make(chan *oauth2.Token, workers)
	errs := make(chan error, workers)

	for i := 0; i < workers; i++ {
		go func() {
			token, err := oauthSvc.fetchUserToken(ctx)
			if err != nil {
				errs <- err
				return
			}
			results <- token
		}()
	}

	for i := 0; i < workers; i++ {
		select {
		case err := <-errs:
			require.NoError(t, err)
		case token := <-results:
			require.Equal(t, "access-token-a2", token.AccessToken)
			require.Equal(t, "refresh-token-r2", token.RefreshToken)
		}
	}
}

func newNonRotatingRefreshOAuthServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "access-token-a2",
			"token_type":   "bearer",
			"expires_in":   3600,
		})
	}))
}

func TestFetchUserTokenPrefersNewerDatabaseTokenOverStaleMemory(t *testing.T) {
	initOAuthTokenTest(t)

	server := newOAuthTestServer(t)
	defer server.Close()

	cfg := newOAuthTestConfig()
	seedExpiredOAuthTokens(t, cfg)
	cfg.failKeys = map[string]int{refreshTokenKey: 10}

	oauthSvc := newOAuthTestService(t, cfg, server.URL+"/oauth/token")
	ctx := context.Background()

	_, err := oauthSvc.fetchUserToken(ctx)
	require.NoError(t, err)

	cfg.failKeys = nil
	require.NoError(t, cfg.SetUpdate(refreshTokenKey, "refresh-token-r3", ""))
	require.NoError(t, cfg.SetUpdate(accessTokenKey, "access-token-a3", ""))
	require.NoError(t, cfg.SetUpdate(accessTokenExpiryKey, strconv.FormatInt(time.Now().Add(2*time.Hour).Unix(), 10), ""))

	token, err := oauthSvc.fetchUserToken(ctx)
	require.NoError(t, err)
	require.Equal(t, "refresh-token-r3", token.RefreshToken)
	require.Equal(t, "access-token-a3", token.AccessToken)
}

func TestFetchUserTokenPrefersNewerDatabaseExpiryWithSameRefreshToken(t *testing.T) {
	initOAuthTokenTest(t)

	cfg := newOAuthTestConfig()
	future := time.Now().Add(2 * time.Hour)
	past := time.Now().Add(-2 * time.Hour)
	require.NoError(t, cfg.SetUpdate(accessTokenKey, "access-token-a-db", ""))
	require.NoError(t, cfg.SetUpdate(accessTokenExpiryKey, strconv.FormatInt(future.Unix(), 10), ""))
	require.NoError(t, cfg.SetUpdate(refreshTokenKey, "refresh-token-r2", ""))

	oauthSvc := newOAuthTestService(t, cfg, "http://localhost/unused")
	oauthSvc.setLatestToken(&oauth2.Token{
		AccessToken:  "access-token-a-mem",
		RefreshToken: "refresh-token-r2",
		Expiry:       past,
	})

	token, err := oauthSvc.fetchUserToken(context.Background())
	require.NoError(t, err)
	require.Equal(t, "access-token-a-db", token.AccessToken)
}

func TestFetchUserTokenPrefersNewerMemoryExpiryWithSameRefreshToken(t *testing.T) {
	initOAuthTokenTest(t)

	cfg := newOAuthTestConfig()
	past := time.Now().Add(-2 * time.Hour)
	future := time.Now().Add(2 * time.Hour)
	require.NoError(t, cfg.SetUpdate(accessTokenKey, "access-token-a-db", ""))
	require.NoError(t, cfg.SetUpdate(accessTokenExpiryKey, strconv.FormatInt(past.Unix(), 10), ""))
	require.NoError(t, cfg.SetUpdate(refreshTokenKey, "refresh-token-r2", ""))

	oauthSvc := newOAuthTestService(t, cfg, "http://localhost/unused")
	oauthSvc.setLatestToken(&oauth2.Token{
		AccessToken:  "access-token-a-mem",
		RefreshToken: "refresh-token-r2",
		Expiry:       future,
	})

	token, err := oauthSvc.fetchUserToken(context.Background())
	require.NoError(t, err)
	require.Equal(t, "access-token-a-mem", token.AccessToken)
}

func TestFetchUserTokenNonRotatingRefreshPersistsSameRefreshToken(t *testing.T) {
	initOAuthTokenTest(t)

	server := newNonRotatingRefreshOAuthServer(t)
	defer server.Close()

	cfg := newOAuthTestConfig()
	seedExpiredOAuthTokens(t, cfg)

	oauthSvc := newOAuthTestService(t, cfg, server.URL)
	ctx := context.Background()

	token, err := oauthSvc.fetchUserToken(ctx)
	require.NoError(t, err)
	require.Equal(t, "access-token-a2", token.AccessToken)
	require.Equal(t, "refresh-token-r1", token.RefreshToken)

	storedRefreshToken, err := cfg.Get(refreshTokenKey, "")
	require.NoError(t, err)
	require.Equal(t, "refresh-token-r1", storedRefreshToken)
}

func TestRemoveOAuthAccessTokenClearsLatestToken(t *testing.T) {
	initOAuthTokenTest(t)

	cfg := newOAuthTestConfig()
	future := time.Now().Add(2 * time.Hour)
	require.NoError(t, cfg.SetUpdate(accessTokenKey, "access-token-a1", ""))
	require.NoError(t, cfg.SetUpdate(accessTokenExpiryKey, strconv.FormatInt(future.Unix(), 10), ""))
	require.NoError(t, cfg.SetUpdate(refreshTokenKey, "refresh-token-r1", ""))

	oauthSvc := newOAuthTestService(t, cfg, "http://localhost/unused")
	oauthSvc.setLatestToken(&oauth2.Token{
		AccessToken:  "access-token-a1",
		RefreshToken: "refresh-token-r1",
		Expiry:       future,
	})

	require.NoError(t, oauthSvc.RemoveOAuthAccessToken())
	require.Nil(t, oauthSvc.latestToken)

	token, err := oauthSvc.fetchUserToken(context.Background())
	require.NoError(t, err)
	require.Nil(t, token)
}

func TestLockAndClearTokenStateClearsLatestToken(t *testing.T) {
	initOAuthTokenTest(t)

	cfg := newOAuthTestConfig()
	oauthSvc := newOAuthTestService(t, cfg, "http://localhost/unused")
	oauthSvc.setLatestToken(&oauth2.Token{
		AccessToken:  "access-token-a1",
		RefreshToken: "refresh-token-r1",
		Expiry:       time.Now().Add(2 * time.Hour),
	})
	oauthSvc.pendingRefreshFrom = "refresh-token-r0"

	oauthSvc.lockAndClearTokenState()
	require.Nil(t, oauthSvc.latestToken)
	require.Equal(t, "", oauthSvc.pendingRefreshFrom)
}

func TestFetchUserTokenPersistenceEventuallyConverges(t *testing.T) {
	initOAuthTokenTest(t)

	server := newOAuthTestServer(t)
	defer server.Close()

	cfg := newOAuthTestConfig()
	seedExpiredOAuthTokens(t, cfg)
	cfg.failKeys = map[string]int{refreshTokenKey: 1}

	oauthSvc := newOAuthTestService(t, cfg, server.URL+"/oauth/token")
	ctx := context.Background()

	_, err := oauthSvc.fetchUserToken(ctx)
	require.NoError(t, err)
	require.Equal(t, "", oauthSvc.pendingRefreshFrom)

	token, err := oauthSvc.fetchUserToken(ctx)
	require.NoError(t, err)
	require.Equal(t, "access-token-a2", token.AccessToken)
	require.Equal(t, "refresh-token-r2", token.RefreshToken)
	require.Equal(t, "", oauthSvc.pendingRefreshFrom)
	storedRefreshToken, err := cfg.Get(refreshTokenKey, "")
	require.NoError(t, err)
	require.Equal(t, "refresh-token-r2", storedRefreshToken)
}
