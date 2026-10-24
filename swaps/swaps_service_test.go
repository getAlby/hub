package swaps

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/BoltzExchange/boltz-client/v2/pkg/boltz"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/getAlby/hub/config"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/lightningnetwork/lnd/zpay32"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeTestInvoice(t *testing.T, paymentHash [32]byte, amountMsat uint64) string {
	t.Helper()

	privKey, err := btcec.NewPrivateKey()
	require.NoError(t, err)

	invoice, err := zpay32.NewInvoice(
		&chaincfg.MainNetParams,
		paymentHash,
		time.Now(),
		zpay32.Amount(lnwire.MilliSatoshi(amountMsat)),
		zpay32.Description("test swap invoice"),
	)
	require.NoError(t, err)

	encoded, err := invoice.Encode(zpay32.MessageSigner{
		SignCompact: func(msg []byte) ([]byte, error) {
			return ecdsa.SignCompact(privKey, chainhash.HashB(msg), true), nil
		},
	})
	require.NoError(t, err)

	return encoded
}

func makeTestPaymentHash(t *testing.T) ([32]byte, string) {
	t.Helper()

	preimage := make([]byte, 32)
	_, err := rand.Read(preimage)
	require.NoError(t, err)
	paymentHash := sha256.Sum256(preimage)
	return paymentHash, hex.EncodeToString(paymentHash[:])
}

func TestVerifySwapOutInvoice(t *testing.T) {
	paymentHash, paymentHashHex := makeTestPaymentHash(t)

	t.Run("accepts invoice with matching payment hash and amount", func(t *testing.T) {
		invoice := makeTestInvoice(t, paymentHash, 100_000_000)

		sendAmountSat, err := verifySwapOutInvoice(invoice, paymentHashHex, 100_000)
		require.NoError(t, err)
		assert.Equal(t, uint64(100_000), sendAmountSat)
	})

	t.Run("rejects invoice with different payment hash", func(t *testing.T) {
		otherPaymentHash, _ := makeTestPaymentHash(t)
		invoice := makeTestInvoice(t, otherPaymentHash, 100_000_000)

		_, err := verifySwapOutInvoice(invoice, paymentHashHex, 100_000)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not match swap payment hash")
	})

	t.Run("rejects invoice exceeding maximum amount", func(t *testing.T) {
		invoice := makeTestInvoice(t, paymentHash, 100_001_000)

		_, err := verifySwapOutInvoice(invoice, paymentHashHex, 100_000)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds maximum expected amount")
	})

	t.Run("rejects invoice without an amount", func(t *testing.T) {
		invoice := makeTestInvoice(t, paymentHash, 0)

		_, err := verifySwapOutInvoice(invoice, paymentHashHex, 100_000)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not have an amount")
	})

	t.Run("rejects unparseable invoice", func(t *testing.T) {
		_, err := verifySwapOutInvoice("lnbc1notaninvoice", paymentHashHex, 100_000)
		require.Error(t, err)
	})
}

func TestCalculateMaxSwapOutSendAmountSat(t *testing.T) {
	// 100_000 requested + 300 claim fee + 500 lockup fee = 100_800,
	// marked up by 0.5% boltz + 1% alby fee on the invoice amount:
	// ceil(100_800 / 0.985) = 102_336, plus 10 sat tolerance
	assert.Equal(t, uint64(102_346), calculateMaxSwapOutSendAmountSat(100_000, 0.5, 500, 300))

	// with a 0% boltz fee only the alby fee percentage applies:
	// ceil(100_800 / 0.99) = 101_819, plus 10 sat tolerance
	assert.Equal(t, uint64(101_829), calculateMaxSwapOutSendAmountSat(100_000, 0, 500, 300))

	// invalid fee rates of 100% or more are never accepted
	assert.Equal(t, uint64(0), calculateMaxSwapOutSendAmountSat(100_000, 100, 500, 300))
}

func TestValidateAddress(t *testing.T) {
	tests := []struct {
		name        string
		network     string
		address     string
		expectedErr string
	}{
		{"mainnet address on mainnet", "bitcoin", "bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4", ""},
		{"testnet address on mainnet", "bitcoin", "tb1qw508d6qejxtdg4y5r3zarvary0c5xw7kxpjzsx", "not a mainnet address"},
		{"node public key", "bitcoin", "0279be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798", "not a mainnet address"},
		{"BIP21 URI", "bitcoin", "bitcoin:bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4?amount=0.001", "invalid bitcoin address"},
		{"signet address on signet", "signet", "tb1qw508d6qejxtdg4y5r3zarvary0c5xw7kxpjzsx", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &swapsService{cfg: &testConfig{network: tt.network}}
			err := svc.ValidateAddress(tt.address)
			if tt.expectedErr == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.expectedErr)
			}
		})
	}
}

func TestSwapOutRejectsInvalidDestination(t *testing.T) {
	// BOLTZ_API skips the swap provider check, and without validation the
	// swap would fail later at the unreachable swap API instead
	svc := &swapsService{
		cfg:      &testConfig{env: &config.AppConfig{BoltzApi: "http://127.0.0.1:1"}, network: "bitcoin"},
		boltzApi: &boltz.Api{URL: "http://127.0.0.1:1"},
	}

	_, err := svc.SwapOut(100_000, "bitcoin:bc1qw508d6qejxtdg4y5r3zarvary0c5xw7kv8f3t4?amount=0.001", false, false)
	assert.ErrorContains(t, err, "invalid bitcoin address")
}
