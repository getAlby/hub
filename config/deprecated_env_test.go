package config

import (
	"testing"

	"github.com/getAlby/hub/logger"
	"github.com/stretchr/testify/assert"
)

func TestApplyDeprecatedEnv(t *testing.T) {
	logger.Init("4")

	t.Run("deprecated variable is used instead of its replacement", func(t *testing.T) {
		env := &AppConfig{
			Network:                "bitcoin",
			ElectrumServer:         "ssl://default.example.com:50002",
			LDKNetwork:             "signet",
			LDKEsploraServer:       "https://esplora.example.com/api",
			LDKElectrumServer:      "ssl://electrum.example.com:50002",
			LDKBitcoindRpcHost:     "127.0.0.1",
			LDKBitcoindRpcPort:     "8332",
			LDKBitcoindRpcUser:     "user",
			LDKBitcoindRpcPassword: "password",
		}

		env.applyDeprecatedEnv()

		assert.Equal(t, "signet", env.Network)
		assert.Equal(t, "https://esplora.example.com/api", env.EsploraServer)
		assert.Equal(t, "ssl://electrum.example.com:50002", env.ElectrumServer)
		assert.Equal(t, "127.0.0.1", env.BitcoindRpcHost)
		assert.Equal(t, "8332", env.BitcoindRpcPort)
		assert.Equal(t, "user", env.BitcoindRpcUser)
		assert.Equal(t, "password", env.BitcoindRpcPassword)
	})

	t.Run("new variables are kept when no deprecated variable is set", func(t *testing.T) {
		env := &AppConfig{
			Network:        "regtest",
			EsploraServer:  "https://esplora.example.com/api",
			ElectrumServer: "ssl://electrum.example.com:50002",
		}

		env.applyDeprecatedEnv()

		assert.Equal(t, "regtest", env.Network)
		assert.Equal(t, "https://esplora.example.com/api", env.EsploraServer)
		assert.Equal(t, "ssl://electrum.example.com:50002", env.ElectrumServer)
		assert.Empty(t, env.BitcoindRpcHost)
	})
}
