package swaps

import (
	"testing"

	"github.com/BoltzExchange/boltz-client/v2/pkg/boltz"
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/db"
	"github.com/stretchr/testify/assert"
)

// testConfig overrides only the config methods used by swapProviderAPIURL
// (tests/mocks cannot be used here as it imports this package)
type testConfig struct {
	config.Config
	env     *config.AppConfig
	network string
}

func (c *testConfig) GetEnv() *config.AppConfig { return c.env }
func (c *testConfig) GetNetwork() string        { return c.network }

func TestSwapProviderAPIURL(t *testing.T) {
	tests := []struct {
		name     string
		boltzApi string
		network  string
		provider string
		expected string
	}{
		{"boltz on mainnet", "", "bitcoin", "boltz", "https://api.boltz.exchange"},
		{"satsrouting on mainnet", "", "bitcoin", "satsrouting", "https://api.satsrouting.exchange"},
		{"provider is normalized", "", "bitcoin", " SatsRouting ", "https://api.satsrouting.exchange"},
		{"empty provider on mainnet", "", "bitcoin", "", ""},
		{"unknown provider on mainnet", "", "bitcoin", "unknown", ""},
		{"no override off mainnet", "", "regtest", "boltz", ""},
		{"override off mainnet", "http://localhost:9001", "regtest", "", "http://localhost:9001"},
		{"override takes precedence on mainnet", "http://localhost:9001", "bitcoin", "satsrouting", "http://localhost:9001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &testConfig{
				env:     &config.AppConfig{BoltzApi: tt.boltzApi},
				network: tt.network,
			}
			assert.Equal(t, tt.expected, swapProviderAPIURL(tt.provider, cfg))
		})
	}
}

func TestCheckSwapApiUrl(t *testing.T) {
	svc := &swapsService{boltzApi: &boltz.Api{URL: "https://api.satsrouting.exchange"}}

	assert.NoError(t, svc.checkSwapApiUrl(&db.Swap{ApiUrl: "https://api.satsrouting.exchange"}))

	err := svc.checkSwapApiUrl(&db.Swap{ApiUrl: "https://api.boltz.exchange"})
	assert.EqualError(t, err, "swap was created with swap API https://api.boltz.exchange but the current swap API is https://api.satsrouting.exchange, set BOLTZ_API=https://api.boltz.exchange to process it")
}
