package chain

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// startMockElectrumServer answers server.version and blockchain.scripthash.get_history
// with an empty history until the test finishes.
func startMockElectrumServer(t *testing.T) string {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { listener.Close() })

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				scanner := bufio.NewScanner(conn)
				for scanner.Scan() {
					var req struct {
						ID     uint64 `json:"id"`
						Method string `json:"method"`
					}
					if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
						return
					}
					var result any = []any{}
					if req.Method == "server.version" {
						result = []string{"mock 1.0", "1.4"}
					}
					resp, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
					if _, err := conn.Write(append(resp, '\n')); err != nil {
						return
					}
				}
				// the client closing the connection ends the scan; nothing to report
				_ = scanner.Err()
			}()
		}
	}()

	return "tcp://" + listener.Addr().String()
}

func TestElectrumLookupCloseReleasesGoroutines(t *testing.T) {
	serverURL := startMockElectrumServer(t)
	ctx := context.Background()

	lookupAndClose := func() {
		lookup, err := newElectrumLookup(ctx, serverURL, "regtest")
		require.NoError(t, err)
		hasTransactions, err := lookup.AddressHasTransactions(ctx, "bcrt1qw508d6qejxtdg4y5r3zarvary0c5xw7kygt080")
		require.NoError(t, err)
		assert.False(t, hasTransactions)
		lookup.Close()
	}

	// warm up so the mock server's accept loop and runtime goroutines are running
	lookupAndClose()
	baseline := waitForGoroutines(0)

	const iterations = 5
	for range iterations {
		lookupAndClose()
	}

	// waitForGoroutines gives server-side connection goroutines time to exit
	assert.LessOrEqual(t, waitForGoroutines(baseline), baseline, "goroutines leaked after %d lookup/close cycles", iterations)
}

// waitForGoroutines polls until the goroutine count drops to target or a deadline passes,
// then returns the current count.
func waitForGoroutines(target int) int {
	deadline := time.Now().Add(2 * time.Second)
	for {
		n := runtime.NumGoroutine()
		if n <= target || time.Now().After(deadline) {
			return n
		}
		time.Sleep(20 * time.Millisecond)
	}
}
