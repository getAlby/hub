package tests

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/lightningnetwork/lnd/zpay32"
)

type Invoice struct {
	Decoded     *zpay32.Invoice
	Encoded     string
	PaymentHash string
}

func NewDummyInvoice(amountMsat uint64, descr string, tm time.Time) (*Invoice, error) {
	random := make([]byte, 32)
	_, err := rand.Read(random)
	if err != nil {
		return nil, fmt.Errorf("failed to read random bytes: %w", err)
	}

	phash := sha256.Sum256(random)

	privKey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	inv, err := zpay32.NewInvoice(&chaincfg.TestNet3Params, phash, tm,
		zpay32.Amount(lnwire.MilliSatoshi(amountMsat)),
		zpay32.Description(descr),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	invoiceStr, err := inv.Encode(zpay32.MessageSigner{
		SignCompact: func(msg []byte) ([]byte, error) {
			return ecdsa.SignCompact(privKey, msg, true), nil
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encode invoice: %w", err)
	}

	return &Invoice{
		Decoded:     inv,
		Encoded:     invoiceStr,
		PaymentHash: hex.EncodeToString(phash[:]),
	}, nil
}
