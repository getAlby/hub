package swaps

import (
	"strings"

	"github.com/getAlby/hub/config"
)

const (
	swapProviderBoltz       = "boltz"
	swapProviderSatsRouting = "satsrouting"
)

func swapProviderAPIURL(provider string, cfg config.Config) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case swapProviderBoltz:
		return cfg.GetEnv().BoltzApi
	case swapProviderSatsRouting:
		return cfg.GetEnv().SatsRoutingApi
	default:
		return ""
	}
}
