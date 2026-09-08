package ldk

import (
	"errors"
	"testing"
	"time"

	"github.com/getAlby/ldk-node-go/ldk_node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/tests"
)

func TestGetVssNodeIdentifier(t *testing.T) {
	mnemonic := "thought turkey ask pottery head say catalog desk pledge elbow naive mimic"
	expectedVssNodeIdentifier := "751636"

	svc, err := tests.CreateTestServiceWithMnemonic(t, mnemonic, "123")
	require.NoError(t, err)
	defer svc.Remove()

	vssNodeIdentifier, err := GetVssNodeIdentifier(svc.Keys)
	require.NoError(t, err)

	assert.Equal(t, expectedVssNodeIdentifier, vssNodeIdentifier)
}
func TestGetVssNodeIdentifier2(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	expectedVssNodeIdentifier := "770256"

	svc, err := tests.CreateTestServiceWithMnemonic(t, mnemonic, "123")
	require.NoError(t, err)
	defer svc.Remove()

	vssNodeIdentifier, err := GetVssNodeIdentifier(svc.Keys)
	require.NoError(t, err)

	assert.Equal(t, expectedVssNodeIdentifier, vssNodeIdentifier)
}

func makeLsps2OpeningFeeParams(minFeeMsat uint64, proportional uint32, minPaymentSizeMsat uint64, maxPaymentSizeMsat uint64) ldk_node.Lsps2OpeningFeeParams {
	return ldk_node.Lsps2OpeningFeeParams{
		MinFeeMsat:           minFeeMsat,
		Proportional:         proportional,
		ValidUntil:           "2035-01-01T00:00:00Z",
		MinLifetime:          4032,
		MaxClientToSelfDelay: 2016,
		MinPaymentSizeMsat:   minPaymentSizeMsat,
		MaxPaymentSizeMsat:   maxPaymentSizeMsat,
		Promise:              "promise",
	}
}

func TestComputeLsps2MaxTotalOpeningFeeMsat(t *testing.T) {
	t.Run("proportional fee above minimum fee", func(t *testing.T) {
		menu := []ldk_node.Lsps2OpeningFeeParams{
			// 0.5% of 10M msat = 50k msat > 10k msat minimum
			makeLsps2OpeningFeeParams(10_000, 5_000, 1_000_000, 100_000_000),
		}
		assert.Equal(t, uint64(50_000), computeLsps2MaxTotalOpeningFeeMsat(10_000_000, menu))
	})

	t.Run("minimum fee above proportional fee", func(t *testing.T) {
		menu := []ldk_node.Lsps2OpeningFeeParams{
			// 0.5% of 1M msat = 5k msat < 10k msat minimum
			makeLsps2OpeningFeeParams(10_000, 5_000, 1_000_000, 100_000_000),
		}
		assert.Equal(t, uint64(10_000), computeLsps2MaxTotalOpeningFeeMsat(1_000_000, menu))
	})

	t.Run("highest fee across menu entries", func(t *testing.T) {
		menu := []ldk_node.Lsps2OpeningFeeParams{
			makeLsps2OpeningFeeParams(10_000, 5_000, 1_000_000, 100_000_000),
			makeLsps2OpeningFeeParams(10_000, 20_000, 1_000_000, 100_000_000),
		}
		assert.Equal(t, uint64(200_000), computeLsps2MaxTotalOpeningFeeMsat(10_000_000, menu))
	})

	t.Run("entries not covering the payment size are skipped", func(t *testing.T) {
		menu := []ldk_node.Lsps2OpeningFeeParams{
			makeLsps2OpeningFeeParams(10_000, 5_000, 1_000_000, 100_000_000),
			// covers larger payments only, would otherwise win with 2%
			makeLsps2OpeningFeeParams(10_000, 20_000, 20_000_000, 100_000_000),
		}
		assert.Equal(t, uint64(50_000), computeLsps2MaxTotalOpeningFeeMsat(10_000_000, menu))
	})

	t.Run("menu fee above ceiling is clamped to ceiling", func(t *testing.T) {
		menu := []ldk_node.Lsps2OpeningFeeParams{
			// 20% of 100M msat = 20M msat, above the 10% / 10M msat ceiling
			makeLsps2OpeningFeeParams(5_000_000, 200_000, 1_000_000, 1_000_000_000),
		}
		assert.Equal(t, uint64(10_000_000), computeLsps2MaxTotalOpeningFeeMsat(100_000_000, menu))
	})

	t.Run("base ceiling applies when no entry covers the payment size", func(t *testing.T) {
		menu := []ldk_node.Lsps2OpeningFeeParams{
			makeLsps2OpeningFeeParams(10_000, 5_000, 1_000_000, 100_000_000),
		}
		// 10% of 200M msat = 20M msat
		assert.Equal(t, uint64(20_000_000), computeLsps2MaxTotalOpeningFeeMsat(200_000_000, menu))
	})

	t.Run("base ceiling applies on empty menu", func(t *testing.T) {
		assert.Equal(t, uint64(5_000_000), computeLsps2MaxTotalOpeningFeeMsat(10_000_000, nil))
	})
}

func TestComputeLsps2MinPaymentSizeMsat(t *testing.T) {
	t.Run("minimum fee below ceiling base", func(t *testing.T) {
		// 1000 sat minimum fee: smallest usable payment nets 1 sat above the fee
		params := makeLsps2OpeningFeeParams(1_000_000, 10_000, 1_000, 100_000_000_000)
		minPaymentSizeMsat, ok := computeLsps2MinPaymentSizeMsat(params)
		require.True(t, ok)
		assert.Equal(t, uint64(1_001_000), minPaymentSizeMsat)
	})

	t.Run("minimum fee above ceiling base", func(t *testing.T) {
		// 8000 sat minimum fee exceeds the 5000 sat ceiling base, so the
		// smallest payment is where the fee equals 10% of the payment
		params := makeLsps2OpeningFeeParams(8_000_000, 10_000, 1_000, 100_000_000_000)
		minPaymentSizeMsat, ok := computeLsps2MinPaymentSizeMsat(params)
		require.True(t, ok)
		assert.Equal(t, uint64(80_000_000), minPaymentSizeMsat)
	})

	t.Run("proportional fee above ceiling percentage never fits", func(t *testing.T) {
		// 30% proportional fee with a minimum fee above the ceiling base can
		// never satisfy the 10% ceiling
		params := makeLsps2OpeningFeeParams(6_000_000, 300_000, 1_000, 1_000_000_000)
		_, ok := computeLsps2MinPaymentSizeMsat(params)
		assert.False(t, ok)
	})
}

func TestFetchLsps2OpeningFeeParamsCachesSuccessfulResponse(t *testing.T) {
	ls := &LDKService{
		lsps2Pubkey:  "pubkey",
		lsps2Address: "127.0.0.1:9735",
	}
	menu := []ldk_node.Lsps2OpeningFeeParams{
		makeLsps2OpeningFeeParams(10_000, 5_000, 1_000_000, 100_000_000),
	}
	calls := 0
	request := func() ([]ldk_node.Lsps2OpeningFeeParams, error) {
		calls++
		return menu, nil
	}

	ls.fetchLsps2OpeningFeeParamsWith(time.Hour, request)
	ls.fetchLsps2OpeningFeeParamsWith(time.Hour, request)

	assert.Equal(t, 1, calls)
	require.NotNil(t, ls.lsps2MinPaymentSizeMsat)
	require.NotNil(t, ls.lsps2MaxPaymentSizeMsat)
	assert.Equal(t, uint64(1_000_000), *ls.lsps2MinPaymentSizeMsat)
	assert.Equal(t, uint64(100_000_000), *ls.lsps2MaxPaymentSizeMsat)
	assert.Equal(t, menu, ls.lsps2OpeningFeeParamsMenu)
}

func TestFetchLsps2OpeningFeeParamsBacksOffAfterFailure(t *testing.T) {
	oldMin := uint64(2_000_000)
	oldMax := uint64(50_000_000)
	oldMenu := []ldk_node.Lsps2OpeningFeeParams{
		makeLsps2OpeningFeeParams(20_000, 5_000, oldMin, oldMax),
	}
	ls := &LDKService{
		lsps2Pubkey:               "pubkey",
		lsps2Address:              "127.0.0.1:9735",
		lsps2InfoFetchedAt:        time.Now().Add(-2 * time.Hour),
		lsps2MinPaymentSizeMsat:   &oldMin,
		lsps2MaxPaymentSizeMsat:   &oldMax,
		lsps2OpeningFeeParamsMenu: oldMenu,
	}
	calls := 0
	request := func() ([]ldk_node.Lsps2OpeningFeeParams, error) {
		calls++
		return nil, errors.New("LSP unavailable")
	}

	ls.fetchLsps2OpeningFeeParamsWith(time.Hour, request)
	ls.fetchLsps2OpeningFeeParamsWith(time.Hour, request)

	assert.Equal(t, 1, calls)
	assert.Equal(t, oldMin, *ls.lsps2MinPaymentSizeMsat)
	assert.Equal(t, oldMax, *ls.lsps2MaxPaymentSizeMsat)
	assert.Equal(t, oldMenu, ls.lsps2OpeningFeeParamsMenu)

	ls.lsps2InfoMu.Lock()
	ls.lsps2InfoLastAttemptedAt = time.Now().Add(-lsps2InfoRetryBackoff - time.Second)
	ls.lsps2InfoMu.Unlock()
	ls.fetchLsps2OpeningFeeParamsWith(time.Hour, request)

	assert.Equal(t, 2, calls)
}

func TestFetchLsps2OpeningFeeParamsDoesNotBlockConcurrentCallers(t *testing.T) {
	ls := &LDKService{
		lsps2Pubkey:  "pubkey",
		lsps2Address: "127.0.0.1:9735",
	}
	fetchStarted := make(chan struct{})
	releaseFetch := make(chan struct{})
	firstDone := make(chan struct{})
	go func() {
		defer close(firstDone)
		ls.fetchLsps2OpeningFeeParamsWith(time.Hour, func() ([]ldk_node.Lsps2OpeningFeeParams, error) {
			close(fetchStarted)
			<-releaseFetch
			return nil, errors.New("LSP unavailable")
		})
	}()
	<-fetchStarted

	secondDone := make(chan struct{})
	secondFetchCalled := false
	go func() {
		defer close(secondDone)
		ls.fetchLsps2OpeningFeeParamsWith(time.Hour, func() ([]ldk_node.Lsps2OpeningFeeParams, error) {
			secondFetchCalled = true
			return nil, nil
		})
	}()

	select {
	case <-secondDone:
	case <-time.After(time.Second):
		t.Fatal("concurrent LSPS2 caller blocked behind the network request")
	}
	assert.False(t, secondFetchCalled)

	close(releaseFetch)
	<-firstDone
}

func TestSanitizeChainEndpointForBitcoind(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		rpcPort  string
		expected string
	}{
		{
			name:     "adds configured port to host",
			endpoint: "127.0.0.1",
			rpcPort:  "8332",
			expected: "127.0.0.1:8332",
		},
		{
			name:     "preserves endpoint port",
			endpoint: "127.0.0.1:18443",
			rpcPort:  "8332",
			expected: "127.0.0.1:18443",
		},
		{
			name:     "formats ipv6 host",
			endpoint: "[2001:db8::1]",
			rpcPort:  "8332",
			expected: "[2001:db8::1]:8332",
		},
		{
			name:     "strips credentials from url-shaped input",
			endpoint: "user:pass@[2001:db8::1]:18443",
			rpcPort:  "8332",
			expected: "[2001:db8::1]:18443",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, sanitizeChainEndpoint(tt.endpoint, tt.rpcPort))
		})
	}
}

func TestSanitizeChainEndpointForURLs(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		expected string
	}{
		{
			name:     "keeps bare electrum endpoint unchanged",
			endpoint: "electrum.example.com:50002",
			expected: "electrum.example.com:50002",
		},
		{
			name:     "strips url credentials",
			endpoint: "ssl://user:pass@electrum.example.com:50002",
			expected: "ssl://electrum.example.com:50002",
		},
		{
			name:     "strips esplora credentials",
			endpoint: "https://user:pass@esplora.example.com/api",
			expected: "https://esplora.example.com/api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, sanitizeChainEndpoint(tt.endpoint, ""))
		})
	}
}
