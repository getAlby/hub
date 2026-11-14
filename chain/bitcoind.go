package chain

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// bitcoindScanDepth is how many blocks back from the tip GetTransactionHex
// searches for a confirmed transaction. Bitcoin Core keeps at least 288 blocks
// even on pruned nodes, and a swap's timeout is well within that window.
const bitcoindScanDepth = 288

// bitcoind RPC error code for a transaction or block that was not found
const bitcoindRpcInvalidAddressOrKey = -5

// ErrAddressLookupUnsupported is returned by chain sources without an address
// index (bitcoind) when asked for an address's transaction history.
var ErrAddressLookupUnsupported = errors.New("address lookups are not supported with the bitcoind chain source")

// bitcoindLookup looks up transactions from the user's own bitcoind over
// JSON-RPC. It does not need -txindex: unconfirmed transactions come from the
// mempool and confirmed ones from a scan of recent blocks.
type bitcoindLookup struct {
	url      string
	user     string
	password string
	client   http.Client
}

type bitcoindRpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *bitcoindRpcError) Error() string {
	return fmt.Sprintf("bitcoind rpc error %d: %s", e.Code, e.Message)
}

func newBitcoindLookup(host, port, user, password string) (*bitcoindLookup, error) {
	if port == "" {
		return nil, errors.New("LDK_BITCOIND_RPC_PORT is required")
	}
	url := host + ":" + port
	if !strings.Contains(host, "://") {
		url = "http://" + url
	}
	return &bitcoindLookup{
		url:      url,
		user:     user,
		password: password,
		client: http.Client{
			Timeout: requestTimeout,
		},
	}, nil
}

func (b *bitcoindLookup) AddressHasTransactions(ctx context.Context, address string) (bool, error) {
	return false, ErrAddressLookupUnsupported
}

func (b *bitcoindLookup) GetTransactionHex(ctx context.Context, txId string) (string, error) {
	// without -txindex this only finds mempool transactions
	var txHex string
	err := b.call(ctx, "getrawtransaction", []any{txId, false}, &txHex)
	if err == nil {
		return txHex, nil
	}
	if !isBitcoindNotFound(err) {
		return "", fmt.Errorf("failed to get transaction: %w", err)
	}

	var tipHeight int64
	if err := b.call(ctx, "getblockcount", []any{}, &tipHeight); err != nil {
		return "", fmt.Errorf("failed to get block count: %w", err)
	}

	for height := tipHeight; height >= 0 && height > tipHeight-bitcoindScanDepth; height-- {
		var blockHash string
		if err := b.call(ctx, "getblockhash", []any{height}, &blockHash); err != nil {
			return "", fmt.Errorf("failed to get block hash at height %d: %w", height, err)
		}
		err := b.call(ctx, "getrawtransaction", []any{txId, false, blockHash}, &txHex)
		if err == nil {
			return txHex, nil
		}
		if !isBitcoindNotFound(err) {
			// e.g. the block has been pruned
			return "", fmt.Errorf("failed to get transaction from block %d: %w", height, err)
		}
	}

	return "", fmt.Errorf("transaction %s not found in the mempool or the last %d blocks", txId, bitcoindScanDepth)
}

func (b *bitcoindLookup) call(ctx context.Context, method string, params []any, result any) error {
	reqBody, err := json.Marshal(map[string]any{
		"jsonrpc": "1.0",
		"id":      "albyhub",
		"method":  method,
		"params":  params,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.url, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(b.user, b.password)

	res, err := b.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// bitcoind returns RPC errors with a non-200 status and a JSON body
	var rpcResponse struct {
		Result json.RawMessage   `json:"result"`
		Error  *bitcoindRpcError `json:"error"`
	}
	if err := json.Unmarshal(body, &rpcResponse); err != nil {
		return fmt.Errorf("unexpected bitcoind response (status %d): %s", res.StatusCode, string(body))
	}
	if rpcResponse.Error != nil {
		return rpcResponse.Error
	}

	return json.Unmarshal(rpcResponse.Result, result)
}

func isBitcoindNotFound(err error) bool {
	var rpcErr *bitcoindRpcError
	return errors.As(err, &rpcErr) && rpcErr.Code == bitcoindRpcInvalidAddressOrKey
}

func (b *bitcoindLookup) Close() {
	b.client.CloseIdleConnections()
}

var _ AddressLookup = (*bitcoindLookup)(nil)
