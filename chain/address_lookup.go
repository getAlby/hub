package chain

import (
	"context"

	"github.com/getAlby/hub/config"
)

// AddressLookup checks onchain address usage against the configured chain source.
type AddressLookup interface {
	// AddressHasTransactions returns true if the address has any confirmed or
	// unconfirmed transaction history.
	AddressHasTransactions(ctx context.Context, address string) (bool, error)
	Close()
}

// NewAddressLookup returns an AddressLookup backed by the same chain source LDK
// uses: esplora if LDK_ESPLORA_SERVER is set, otherwise electrum.
// Bitcoind has no address index, so bitcoind users also fall back to electrum.
func NewAddressLookup(ctx context.Context, cfg config.Config) (AddressLookup, error) {
	env := cfg.GetEnv()
	if env.LDKEsploraServer != "" {
		return newEsploraLookup(env.LDKEsploraServer), nil
	}
	return newElectrumLookup(ctx, env.LDKElectrumServer, cfg.GetNetwork())
}
