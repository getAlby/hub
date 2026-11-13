package chain

import (
	"context"
	"errors"

	"github.com/getAlby/hub/config"
)

// AddressLookup queries onchain data from the configured chain source.
type AddressLookup interface {
	// AddressHasTransactions returns true if the address has any confirmed or
	// unconfirmed transaction history.
	AddressHasTransactions(ctx context.Context, address string) (bool, error)
	// GetTransactionHex returns the raw hex of a confirmed or unconfirmed transaction.
	GetTransactionHex(ctx context.Context, txId string) (string, error)
	Close()
}

// ErrBitcoindLookupUnsupported is returned when the chain source is bitcoind,
// which has no address index and cannot look up confirmed transactions.
var ErrBitcoindLookupUnsupported = errors.New("address and transaction lookups are not supported with the bitcoind chain source: use an esplora or electrum server as the chain source instead")

// NewAddressLookup returns an AddressLookup backed by the same chain source LDK
// uses: esplora if LDK_ESPLORA_SERVER is set, otherwise electrum.
// It never falls back to another server when LDK uses bitcoind, so that a
// bitcoind user's lookups are not sent to a server they did not choose.
func NewAddressLookup(ctx context.Context, cfg config.Config) (AddressLookup, error) {
	env := cfg.GetEnv()
	if env.LDKBitcoindRpcHost != "" {
		return nil, ErrBitcoindLookupUnsupported
	}
	if env.LDKEsploraServer != "" {
		return newEsploraLookup(env.LDKEsploraServer), nil
	}
	return newElectrumLookup(ctx, env.LDKElectrumServer, cfg.GetNetwork())
}
