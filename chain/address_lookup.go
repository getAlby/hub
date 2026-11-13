package chain

import (
	"context"

	"github.com/getAlby/hub/config"
)

// AddressLookup queries onchain data from the configured chain source.
type AddressLookup interface {
	// AddressHasTransactions returns true if the address has any confirmed or
	// unconfirmed transaction history. Returns ErrAddressLookupUnsupported if
	// the chain source has no address index.
	AddressHasTransactions(ctx context.Context, address string) (bool, error)
	// GetTransactionHex returns the raw hex of a confirmed or unconfirmed transaction.
	GetTransactionHex(ctx context.Context, txId string) (string, error)
	Close()
}

// NewAddressLookup returns an AddressLookup backed by the same chain source LDK
// uses: bitcoind if LDK_BITCOIND_RPC_HOST is set, esplora if
// LDK_ESPLORA_SERVER is set, otherwise electrum. Lookups are never sent to a
// server other than the one LDK uses.
func NewAddressLookup(ctx context.Context, cfg config.Config) (AddressLookup, error) {
	env := cfg.GetEnv()
	if env.LDKBitcoindRpcHost != "" {
		return newBitcoindLookup(env.LDKBitcoindRpcHost, env.LDKBitcoindRpcPort, env.LDKBitcoindRpcUser, env.LDKBitcoindRpcPassword)
	}
	if env.LDKEsploraServer != "" {
		return newEsploraLookup(env.LDKEsploraServer), nil
	}
	return newElectrumLookup(ctx, env.LDKElectrumServer, cfg.GetNetwork())
}
