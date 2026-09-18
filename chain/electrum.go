package chain

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/BoltzExchange/go-electrum/electrum"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

const requestTimeout = 10 * time.Second

type electrumLookup struct {
	client    *electrum.Client
	netParams *chaincfg.Params
}

// newElectrumLookup connects to the given electrum server. The server URL uses the same
// format as LDK_ELECTRUM_SERVER: "ssl://host:port", "tcp://host:port" or "host:port" (tcp).
func newElectrumLookup(ctx context.Context, serverURL string, network string) (*electrumLookup, error) {
	if serverURL == "" {
		return nil, errors.New("no electrum server configured")
	}

	netParams, err := chainParams(network)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	var client *electrum.Client
	switch {
	case strings.HasPrefix(serverURL, "ssl://"):
		client, err = electrum.NewClientSSL(ctx, strings.TrimPrefix(serverURL, "ssl://"), &tls.Config{})
	case strings.HasPrefix(serverURL, "tcp://"):
		client, err = electrum.NewClientTCP(ctx, strings.TrimPrefix(serverURL, "tcp://"))
	default:
		client, err = electrum.NewClientTCP(ctx, serverURL)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to electrum server: %w", err)
	}

	// the protocol version must be negotiated before any other request
	if _, _, err := client.ServerVersion(ctx); err != nil {
		client.Shutdown()
		return nil, fmt.Errorf("failed to negotiate electrum server version: %w", err)
	}

	return &electrumLookup{
		client:    client,
		netParams: netParams,
	}, nil
}

func (e *electrumLookup) AddressHasTransactions(ctx context.Context, address string) (bool, error) {
	scriptHash, err := addressToScriptHash(address, e.netParams)
	if err != nil {
		return false, err
	}

	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	history, err := e.client.GetHistory(ctx, scriptHash)
	if err != nil {
		return false, fmt.Errorf("failed to get address history: %w", err)
	}

	return len(history) > 0, nil
}

func (e *electrumLookup) Close() {
	e.client.Shutdown()
}

var _ AddressLookup = (*electrumLookup)(nil)

// chainParams maps a network name as returned by config.GetNetwork() to btcd chain params.
func chainParams(network string) (*chaincfg.Params, error) {
	switch network {
	case "bitcoin", "mainnet":
		return &chaincfg.MainNetParams, nil
	case "testnet":
		return &chaincfg.TestNet3Params, nil
	case "regtest":
		return &chaincfg.RegressionNetParams, nil
	case "signet":
		return &chaincfg.SigNetParams, nil
	default:
		return nil, fmt.Errorf("unsupported network: %s", network)
	}
}

// addressToScriptHash converts an address to an electrum script hash:
// sha256 of the output script, reversed and hex encoded.
// https://electrumx.readthedocs.io/en/latest/protocol-basics.html#script-hashes
func addressToScriptHash(address string, netParams *chaincfg.Params) (string, error) {
	decoded, err := btcutil.DecodeAddress(address, netParams)
	if err != nil {
		return "", fmt.Errorf("invalid address: %w", err)
	}
	script, err := txscript.PayToAddrScript(decoded)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(script)
	for i, j := 0, len(hash)-1; i < j; i, j = i+1, j-1 {
		hash[i], hash[j] = hash[j], hash[i]
	}

	return hex.EncodeToString(hash[:]), nil
}
