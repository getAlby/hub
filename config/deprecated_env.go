package config

import (
	"github.com/getAlby/hub/logger"
	"github.com/sirupsen/logrus"
)

// applyDeprecatedEnv keeps the old LDK_ names of renamed variables working: a
// deprecated variable that is set is used instead of its replacement.
func (c *AppConfig) applyDeprecatedEnv() {
	for _, variable := range []struct {
		deprecatedName  string
		deprecatedValue string
		name            string
		value           *string
	}{
		{"LDK_NETWORK", c.LDKNetwork, "NETWORK", &c.Network},
		{"LDK_ESPLORA_SERVER", c.LDKEsploraServer, "ESPLORA_SERVER", &c.EsploraServer},
		{"LDK_ELECTRUM_SERVER", c.LDKElectrumServer, "ELECTRUM_SERVER", &c.ElectrumServer},
		{"LDK_BITCOIND_RPC_HOST", c.LDKBitcoindRpcHost, "BITCOIND_RPC_HOST", &c.BitcoindRpcHost},
		{"LDK_BITCOIND_RPC_PORT", c.LDKBitcoindRpcPort, "BITCOIND_RPC_PORT", &c.BitcoindRpcPort},
		{"LDK_BITCOIND_RPC_USER", c.LDKBitcoindRpcUser, "BITCOIND_RPC_USER", &c.BitcoindRpcUser},
		{"LDK_BITCOIND_RPC_PASSWORD", c.LDKBitcoindRpcPassword, "BITCOIND_RPC_PASSWORD", &c.BitcoindRpcPassword},
	} {
		if variable.deprecatedValue == "" {
			continue
		}
		logger.Logger.WithFields(logrus.Fields{
			"deprecated":  variable.deprecatedName,
			"replacement": variable.name,
		}).Warn("Deprecated environment variable is set and takes priority, please rename it")
		*variable.value = variable.deprecatedValue
	}
}
