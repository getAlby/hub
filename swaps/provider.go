package swaps

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/getAlby/hub/alby"
	"github.com/getAlby/hub/config"
)

const (
	swapProviderBoltz       = "boltz"
	swapProviderSatsRouting = "satsrouting"
)

var (
	ErrSwapProviderChanged     = errors.New("the swap provider has changed, please restart Alby Hub")
	ErrSwapProviderUnavailable = errors.New("swaps are currently unavailable")
)

// checkSwapProvider ensures the swap provider selected by Alby has not changed
// since the swaps service started. It is skipped when BOLTZ_API is set.
func (svc *swapsService) checkSwapProvider() error {
	if svc.cfg.GetEnv().BoltzApi != "" {
		return nil
	}

	apiURL, err := fetchSwapProviderAPIURL(svc.ctx, svc.albySvc, svc.cfg)
	if err != nil {
		return err
	}
	if apiURL == "" {
		return ErrSwapProviderUnavailable
	}
	if apiURL != svc.boltzApi.URL {
		return ErrSwapProviderChanged
	}
	return nil
}

// fetchSwapProviderAPIURL returns the swap API URL for the swap provider
// currently selected by Alby, or an empty string if swaps are unavailable.
// Swap providers are only selected by Alby on mainnet.
func fetchSwapProviderAPIURL(ctx context.Context, albySvc alby.AlbyService, cfg config.Config) (string, error) {
	if cfg.GetNetwork() != "bitcoin" {
		return "", nil
	}

	albyInfo, err := albySvc.GetInfo(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to fetch swap provider from Alby info: %w", err)
	}
	return mapSwapProviderToAPIURL(albyInfo.Hub.SwapProvider), nil
}

// mapSwapProviderToAPIURL returns the mainnet swap API URL for a swap
// provider, or an empty string if the provider is unknown
func mapSwapProviderToAPIURL(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case swapProviderBoltz:
		return "https://api.boltz.exchange"
	case swapProviderSatsRouting:
		return "https://api.satsrouting.exchange"
	default:
		return ""
	}
}
