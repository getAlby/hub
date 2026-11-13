package chain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/config"
)

// testConfig overrides only the config methods used to create a lookup
type testConfig struct {
	config.Config
	env *config.AppConfig
}

func (c *testConfig) GetEnv() *config.AppConfig { return c.env }
func (c *testConfig) GetNetwork() string        { return "bitcoin" }

// lookups for a bitcoind user must not be sent to another server
func TestNewAddressLookup_BitcoindDoesNotFallBack(t *testing.T) {
	cfg := &testConfig{env: &config.AppConfig{
		LDKBitcoindRpcHost: "127.0.0.1",
		LDKEsploraServer:   "https://esplora.example.com/api",
		LDKElectrumServer:  "ssl://electrum.example.com:50002",
	}}

	lookup, err := NewAddressLookup(context.Background(), cfg)

	require.EqualError(t, err, "address and transaction lookups are not supported with the bitcoind chain source: use an esplora or electrum server as the chain source instead")
	assert.Nil(t, lookup)
}

func TestNewAddressLookup_Esplora(t *testing.T) {
	cfg := &testConfig{env: &config.AppConfig{
		LDKEsploraServer:  "https://esplora.example.com/api",
		LDKElectrumServer: "ssl://electrum.example.com:50002",
	}}

	lookup, err := NewAddressLookup(context.Background(), cfg)

	require.NoError(t, err)
	defer lookup.Close()
	assert.IsType(t, &esploraLookup{}, lookup)
}
