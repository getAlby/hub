package chain

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeBitcoind serves the RPC methods used by bitcoindLookup. It has no
// -txindex: getrawtransaction without a block hash only sees the mempool.
type fakeBitcoind struct {
	mempool      map[string]string   // txid -> hex
	blocks       []map[string]string // height -> txid -> hex
	prunedBelow  int                 // blocks below this height are pruned
	calledMethod map[string]int
}

func (f *fakeBitcoind) start(t *testing.T) *bitcoindLookup {
	f.calledMethod = map[string]int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, password, ok := r.BasicAuth()
		if !ok || user != "user" || password != "pass" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var req struct {
			Method string `json:"method"`
			Params []any  `json:"params"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		f.calledMethod[req.Method]++

		result, rpcErr := f.handle(req.Method, req.Params)
		if rpcErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{"result": nil, "error": rpcErr})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"result": result, "error": nil})
	}))
	t.Cleanup(server.Close)

	_, port, _ := strings.Cut(server.URL, "://127.0.0.1:")
	lookup, err := newBitcoindLookup("127.0.0.1", port, "user", "pass")
	require.NoError(t, err)
	t.Cleanup(lookup.Close)
	return lookup
}

func (f *fakeBitcoind) handle(method string, params []any) (any, *bitcoindRpcError) {
	notFound := &bitcoindRpcError{Code: -5, Message: "No such mempool or blockchain transaction"}
	switch method {
	case "getblockcount":
		return len(f.blocks) - 1, nil
	case "getblockhash":
		return fmt.Sprintf("hash%d", int(params[0].(float64))), nil
	case "getrawtransaction":
		txId := params[0].(string)
		if len(params) < 3 {
			if txHex, ok := f.mempool[txId]; ok {
				return txHex, nil
			}
			return nil, notFound
		}
		var height int
		_, _ = fmt.Sscanf(params[2].(string), "hash%d", &height)
		if height < f.prunedBelow {
			return nil, &bitcoindRpcError{Code: -1, Message: "Block not available (pruned data)"}
		}
		if txHex, ok := f.blocks[height][txId]; ok {
			return txHex, nil
		}
		return nil, notFound
	}
	return nil, &bitcoindRpcError{Code: -32601, Message: "Method not found"}
}

func newFakeChain(height int) []map[string]string {
	blocks := make([]map[string]string, height+1)
	for i := range blocks {
		blocks[i] = map[string]string{}
	}
	return blocks
}

func TestBitcoindGetTransactionHex_Mempool(t *testing.T) {
	fake := &fakeBitcoind{mempool: map[string]string{"mempooltx": "02000000aa"}, blocks: newFakeChain(1000)}
	lookup := fake.start(t)

	txHex, err := lookup.GetTransactionHex(context.Background(), "mempooltx")

	require.NoError(t, err)
	assert.Equal(t, "02000000aa", txHex)
	assert.Equal(t, 0, fake.calledMethod["getblockhash"])
}

func TestBitcoindGetTransactionHex_RecentBlock(t *testing.T) {
	blocks := newFakeChain(1000)
	blocks[995]["confirmedtx"] = "02000000bb"
	fake := &fakeBitcoind{blocks: blocks}
	lookup := fake.start(t)

	txHex, err := lookup.GetTransactionHex(context.Background(), "confirmedtx")

	require.NoError(t, err)
	assert.Equal(t, "02000000bb", txHex)
	// scans back from the tip and stops at the block containing the tx
	assert.Equal(t, 6, fake.calledMethod["getblockhash"])
}

func TestBitcoindGetTransactionHex_NotFound(t *testing.T) {
	blocks := newFakeChain(1000)
	// older than the scan window
	blocks[1000-bitcoindScanDepth]["oldtx"] = "02000000cc"
	fake := &fakeBitcoind{blocks: blocks}
	lookup := fake.start(t)

	_, err := lookup.GetTransactionHex(context.Background(), "oldtx")

	assert.ErrorContains(t, err, "not found in the mempool or the last 288 blocks")
	assert.Equal(t, bitcoindScanDepth, fake.calledMethod["getblockhash"])
}

func TestBitcoindGetTransactionHex_PrunedBlock(t *testing.T) {
	fake := &fakeBitcoind{blocks: newFakeChain(1000), prunedBelow: 990}
	lookup := fake.start(t)

	_, err := lookup.GetTransactionHex(context.Background(), "unknowntx")

	assert.ErrorContains(t, err, "Block not available")
}

func TestBitcoindGetTransactionHex_ShortChain(t *testing.T) {
	blocks := newFakeChain(10)
	blocks[0]["genesistx"] = "02000000dd"
	fake := &fakeBitcoind{blocks: blocks}
	lookup := fake.start(t)

	txHex, err := lookup.GetTransactionHex(context.Background(), "genesistx")

	require.NoError(t, err)
	assert.Equal(t, "02000000dd", txHex)
}

func TestBitcoindAddressHasTransactions_Unsupported(t *testing.T) {
	fake := &fakeBitcoind{blocks: newFakeChain(1)}
	lookup := fake.start(t)

	_, err := lookup.AddressHasTransactions(context.Background(), "bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4")

	require.ErrorIs(t, err, ErrAddressLookupUnsupported)
	assert.Empty(t, fake.calledMethod)
}

func TestBitcoindRequiresPort(t *testing.T) {
	_, err := newBitcoindLookup("127.0.0.1", "", "user", "pass")
	assert.EqualError(t, err, "LDK_BITCOIND_RPC_PORT is required")
}
