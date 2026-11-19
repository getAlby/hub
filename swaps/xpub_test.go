package swaps

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testXpub = "xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EGMcet8"

// xpubTestConfig has no saved xpub start index
type xpubTestConfig struct {
	testConfig
}

func (c *xpubTestConfig) Get(key string, encryptionKey string) (string, error) {
	return "", nil
}

func newXpubTestService(env *config.AppConfig) *swapsService {
	return &swapsService{
		ctx: context.Background(),
		cfg: &xpubTestConfig{testConfig{env: env, network: "bitcoin"}},
	}
}

func TestValidateXpub(t *testing.T) {
	logger.Init("4")

	esplora := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("[]"))
	}))
	defer esplora.Close()

	t.Run("xpub with an unused address", func(t *testing.T) {
		svc := newXpubTestService(&config.AppConfig{EsploraServer: esplora.URL})
		require.NoError(t, svc.ValidateXpub(testXpub))
	})

	t.Run("invalid xpub", func(t *testing.T) {
		svc := newXpubTestService(&config.AppConfig{EsploraServer: esplora.URL})
		assert.ErrorContains(t, svc.ValidateXpub("not-an-xpub"), "invalid xpub")
	})

	// bitcoind has no address index, so it cannot find the next unused address
	t.Run("chain source without address lookups", func(t *testing.T) {
		svc := newXpubTestService(&config.AppConfig{BitcoindRpcHost: "127.0.0.1", BitcoindRpcPort: "1"})
		assert.ErrorContains(t, svc.ValidateXpub(testXpub), "address lookups are not supported")
	})
}
