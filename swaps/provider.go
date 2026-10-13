package swaps

import (
	"strings"

	"github.com/getAlby/hub/config"
)

const (
	swapProviderBoltz       = "boltz"
	swapProviderSatsRouting = "satsrouting"
)

// swapProviderAPIURL returns the swap API URL to use, or an empty string if
// swaps are unavailable. An explicit BOLTZ_API always takes precedence;
// otherwise the provider from Alby info is mapped to its mainnet API.
func swapProviderAPIURL(provider string, cfg config.Config) string {
	if boltzApi := cfg.GetEnv().BoltzApi; boltzApi != "" {
		return boltzApi
	}

	if cfg.GetNetwork() != "bitcoin" {
		return ""
	}

	switch strings.ToLower(strings.TrimSpace(provider)) {
	case swapProviderBoltz:
		return "https://api.boltz.exchange"
	case swapProviderSatsRouting:
		return "https://api.satsrouting.exchange"
	default:
		return ""
	}
}
