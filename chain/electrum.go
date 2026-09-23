package chain

import (
	"bufio"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
)

const requestTimeout = 10 * time.Second

// electrum protocol version negotiated with the server. 1.4 is the minimum that
// supports blockchain.scripthash.get_history and is supported by electrs and ElectrumX.
const (
	electrumClientName      = "albyhub"
	electrumProtocolVersion = "1.4"
)

// electrumLookup is a minimal synchronous electrum JSON-RPC client. Requests are written
// and their responses read on the caller's goroutine, so closing it leaves nothing running.
type electrumLookup struct {
	conn      net.Conn
	reader    *bufio.Reader
	netParams *chaincfg.Params
	nextID    uint64
	mu        sync.Mutex
}

type electrumRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      uint64 `json:"id"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
}

type electrumResponse struct {
	ID     *uint64         `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  json.RawMessage `json:"error"`
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

	var conn net.Conn
	switch {
	case strings.HasPrefix(serverURL, "ssl://"):
		dialer := tls.Dialer{Config: &tls.Config{}}
		conn, err = dialer.DialContext(ctx, "tcp", strings.TrimPrefix(serverURL, "ssl://"))
	case strings.HasPrefix(serverURL, "tcp://"):
		var dialer net.Dialer
		conn, err = dialer.DialContext(ctx, "tcp", strings.TrimPrefix(serverURL, "tcp://"))
	default:
		var dialer net.Dialer
		conn, err = dialer.DialContext(ctx, "tcp", serverURL)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to electrum server: %w", err)
	}

	e := &electrumLookup{
		conn:      conn,
		reader:    bufio.NewReader(conn),
		netParams: netParams,
	}

	// the protocol version must be negotiated before any other request
	if err := e.request(ctx, "server.version", []any{electrumClientName, electrumProtocolVersion}, nil); err != nil {
		e.Close()
		return nil, fmt.Errorf("failed to negotiate electrum server version: %w", err)
	}

	return e, nil
}

func (e *electrumLookup) AddressHasTransactions(ctx context.Context, address string) (bool, error) {
	scriptHash, err := addressToScriptHash(address, e.netParams)
	if err != nil {
		return false, err
	}

	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	var history []json.RawMessage
	if err := e.request(ctx, "blockchain.scripthash.get_history", []any{scriptHash}, &history); err != nil {
		return false, fmt.Errorf("failed to get address history: %w", err)
	}

	return len(history) > 0, nil
}

func (e *electrumLookup) Close() {
	_ = e.conn.Close()
}

// request sends a single request and waits for the response with the matching id.
// Any failure leaves the connection in an unknown state, so it is closed.
func (e *electrumLookup) request(ctx context.Context, method string, params []any, result any) (err error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	defer func() {
		if err != nil {
			_ = e.conn.Close()
		}
	}()

	// unblock the read/write below if ctx is cancelled before its deadline
	stop := context.AfterFunc(ctx, func() {
		_ = e.conn.SetDeadline(time.Now())
	})
	defer stop()
	if deadline, ok := ctx.Deadline(); ok {
		if err := e.conn.SetDeadline(deadline); err != nil {
			return err
		}
	}

	e.nextID++
	id := e.nextID
	body, err := json.Marshal(electrumRequest{JSONRPC: "2.0", ID: id, Method: method, Params: params})
	if err != nil {
		return err
	}
	if _, err := e.conn.Write(append(body, '\n')); err != nil {
		return err
	}

	for {
		line, err := e.reader.ReadBytes('\n')
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}

		var resp electrumResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			return fmt.Errorf("invalid electrum response: %w", err)
		}
		// skip notifications and responses to other requests
		if resp.ID == nil || *resp.ID != id {
			continue
		}
		if len(resp.Error) > 0 && string(resp.Error) != "null" {
			return fmt.Errorf("electrum error: %s", resp.Error)
		}
		if result == nil {
			return nil
		}
		return json.Unmarshal(resp.Result, result)
	}
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
