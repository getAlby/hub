package alby

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/logger"
)

type oauthRetryConfig struct {
	config.Config
	values       map[string]string
	failedKey    string
	failCount    int
	failedWrites int
}

func (cfg *oauthRetryConfig) Get(key, _ string) (string, error) {
	return cfg.values[key], nil
}

func (cfg *oauthRetryConfig) SetUpdate(key, value, _ string) error {
	if key == cfg.failedKey && cfg.failedWrites < cfg.failCount {
		cfg.failedWrites++
		return errors.New("transient database error")
	}
	cfg.values[key] = value
	return nil
}

func TestFetchUserTokenRetriesPersistence(t *testing.T) {
	logger.Init("4")
	for _, failedKey := range []string{accessTokenExpiryKey, accessTokenKey, refreshTokenKey} {
		t.Run(failedKey, func(t *testing.T) {
			values := map[string]string{
				accessTokenKey:       "old-access-token",
				accessTokenExpiryKey: strconv.FormatInt(time.Now().Add(-time.Hour).Unix(), 10),
				refreshTokenKey:      "old-refresh-token",
			}
			cfg := &oauthRetryConfig{values: values, failedKey: failedKey, failCount: tokenSaveAttempts - 1}
			var refreshes atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				refreshes.Add(1)
				if err := r.ParseForm(); err != nil || r.FormValue("refresh_token") != "old-refresh-token" {
					http.Error(w, "invalid refresh token", http.StatusBadRequest)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"access_token":"new-access-token","refresh_token":"new-refresh-token","token_type":"bearer","expires_in":3600}`))
			}))
			defer server.Close()
			svc := &albyOAuthService{cfg: cfg, oauthConf: &oauth2.Config{
				ClientID: "test-client", Endpoint: oauth2.Endpoint{TokenURL: server.URL},
			}}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			const workers = 8
			errs := make(chan error, workers)
			for range workers {
				go func() {
					token, err := svc.fetchUserToken(ctx)
					if err == nil && (token == nil || token.RefreshToken != "new-refresh-token") {
						err = errors.New("refreshed token was not reused")
					}
					errs <- err
				}()
			}
			for range workers {
				select {
				case err := <-errs:
					require.NoError(t, err)
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				}
			}
			require.EqualValues(t, 1, refreshes.Load())
			require.Equal(t, tokenSaveAttempts-1, cfg.failedWrites)
			require.Equal(t, "new-access-token", values[accessTokenKey])
			require.Equal(t, "new-refresh-token", values[refreshTokenKey])
		})
	}
}

func TestSaveTokenStopsAfterBoundedRetries(t *testing.T) {
	logger.Init("4")
	cfg := &oauthRetryConfig{failedKey: accessTokenExpiryKey, failCount: tokenSaveAttempts}
	svc := &albyOAuthService{cfg: cfg}
	svc.saveToken(&oauth2.Token{Expiry: time.Now().Add(time.Hour)})
	require.Equal(t, tokenSaveAttempts, cfg.failedWrites)
}
