package chain

import (
	"context"

	"github.com/getAlby/hub/config"
)

// AddressLookup queries onchain data from the configured chain source.
type AddressLookup interface {
	// AddressHasTransactions returns true if the address has any confirmed or
	// unconfirmed transaction history. Returns an error if the chain
	// source has no address index.
	AddressHasTransactions(ctx context.Context, address string) (bool, error)
	// GetTransactionHex returns the raw hex of a confirmed or unconfirmed transaction.
	GetTransactionHex(ctx context.Context, txId string) (string, error)
	Close()
}

// NewAddressLookup returns an AddressLookup backed by the configured chain
// source: bitcoind if BITCOIND_RPC_HOST is set, esplora if ESPLORA_SERVER is
// set, otherwise electrum. Lookups are never sent to any other server.
func NewAddressLookup(ctx context.Context, cfg config.Config) (AddressLookup, error) {
	env := cfg.GetEnv()
	if env.BitcoindRpcHost != "" {
		return newBitcoindLookup(env.BitcoindRpcHost, env.BitcoindRpcPort, env.BitcoindRpcUser, env.BitcoindRpcPassword)
	}
	if env.EsploraServer != "" {
		return newEsploraLookup(env.EsploraServer), nil
	}
	return newElectrumLookup(ctx, env.ElectrumServer, cfg.GetNetwork())
}
