package swaps

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
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
