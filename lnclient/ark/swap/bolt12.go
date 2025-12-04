package swap

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/btcsuite/btcd/btcutil/bech32"
	"github.com/ccoveille/go-safecast"
	"github.com/lightningnetwork/lnd/input"
	"github.com/lightningnetwork/lnd/tlv"
)

type Offer struct {
	ID             string
	Amount         uint64
	AmountInSats   uint64
	Description    []byte
	DescriptionStr string
	AbsoluteExpiry uint64
	QuantityMax    uint64
}

type Invoice struct {
	Amount         uint64
	AmountInSats   uint64
	PaymentHash    []byte
	PaymentHash160 []byte
}

// BOLT12 TLV types
const (
	OFFER_CHAINS          uint64 = 2
	OFFER_METADATA        uint64 = 4
	OFFER_CURRENCY        uint64 = 6
	OFFER_AMOUNT          uint64 = 8
	OFFER_DESCRIPTION     uint64 = 10
	OFFER_FEATURES        uint64 = 12
	OFFER_ABSOLUTE_EXPIRY uint64 = 14
	OFFER_PATHS           uint64 = 16
	OFFER_ISSUER          uint64 = 18
	OFFER_QUANTITY_MAX    uint64 = 20
	OFFER_ISSUER_ID       uint64 = 22

	INVOICE_PAYMENT_HASH uint64 = 168
	INVOICE_AMOUNT       uint64 = 170
)

func DecodeBolt12Invoice(invoice string) (*Invoice, error) {
	payload, err := bech32ToWords(invoice, "lni")
	if err != nil {
		return nil, fmt.Errorf("bech32ToWords: %w", err)
	}

	invoiceData := new(Invoice)

	sizeFunc := func() uint64 {
		return tlv.SizeTUint64(invoiceData.Amount)
	}

	tlvStream := tlv.MustNewStream(
		tlv.MakePrimitiveRecord(tlv.Type(INVOICE_PAYMENT_HASH), &invoiceData.PaymentHash),
		tlv.MakeDynamicRecord(tlv.Type(INVOICE_AMOUNT), &invoiceData.Amount, sizeFunc, tlv.ETUint64, tlv.DTUint64),
	)

	err = tlvStream.Decode(bytes.NewReader(payload))

	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	if len(invoiceData.PaymentHash) != 32 {
		return nil, errors.New("invoice payment hash must be 32 bytes")
	}

	if invoiceData.Amount == 0 {
		return nil, errors.New("invoice amount is required")
	}

	invoiceData.AmountInSats, err = safecast.ToUint64(invoiceData.Amount / 1000)
	if err != nil {
		return nil, fmt.Errorf("invalid amount in sats: %w", err)
	}

	invoiceData.PaymentHash160 = input.Ripemd160H(invoiceData.PaymentHash)

	return invoiceData, nil
}

func DecodeBolt12Offer(offer string) (*Offer, error) {
	payload, err := bech32ToWords(offer, "lno")
	if err != nil {
		return nil, fmt.Errorf("bech32ToWords: %w", err)
	}

	offerData := new(Offer)

	sizeFunc := func() uint64 {
		return tlv.SizeTUint64(offerData.Amount)
	}

	tlvStream := tlv.MustNewStream(
		tlv.MakeDynamicRecord(tlv.Type(OFFER_AMOUNT), &offerData.Amount, sizeFunc, tlv.ETUint64, tlv.DTUint64),
		tlv.MakePrimitiveRecord(tlv.Type(OFFER_DESCRIPTION), &offerData.Description),
		tlv.MakeDynamicRecord(tlv.Type(OFFER_ABSOLUTE_EXPIRY), &offerData.AbsoluteExpiry, sizeFunc, tlv.ETUint64, tlv.DTUint64),
		tlv.MakePrimitiveRecord(tlv.Type(OFFER_QUANTITY_MAX), &offerData.QuantityMax),
	)

	err = tlvStream.Decode(bytes.NewReader(payload))

	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	if len(offerData.Description) == 0 {
		return nil, errors.New("offer description is required")
	}

	if offerData.Amount == 0 {
		return nil, errors.New("offer amount is required")
	}

	offerData.AmountInSats, err = safecast.ToUint64(offerData.Amount / 1000)
	if err != nil {
		return nil, fmt.Errorf("invalid amount in sats: %w", err)
	}

	offerData.DescriptionStr = string(offerData.Description)

	offerHash := sha256.Sum256(payload)

	offerData.ID = hex.EncodeToString(offerHash[:])

	return offerData, nil
}

func cleanBech32String(bech32String string) string {
	var cleaned strings.Builder
	skip := false
	for _, r := range bech32String {
		if r == '+' {
			skip = true
			continue
		}
		if skip && unicode.IsSpace(r) {
			continue
		}
		skip = false
		cleaned.WriteRune(r)
	}
	return cleaned.String()
}

func bech32ToWords(bech32String string, hrp string) (payload []byte, err error) {
	cleanString := strings.ToLower(cleanBech32String(bech32String))

	// Split into prefix and data by first '1'
	parts := strings.SplitN(cleanString, "1", 2)
	if len(parts) != 2 {
		return nil, errors.New("invalid BOLT12 format: missing separator")
	}
	prefix, data := parts[0], parts[1]

	const charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

	words := make([]byte, 0, len(data))
	for _, r := range data {
		index := strings.IndexRune(charset, r)
		if index == -1 {
			return nil, errors.New("invalid character in data part")
		}
		words = append(words, byte(index))
	}

	if prefix != hrp {
		return nil, fmt.Errorf("invalid prefix: expected %s, got %s", hrp, prefix)
	}

	payload, err = bech32.ConvertBits(words, 5, 8, false)
	if err != nil {
		return nil, fmt.Errorf("ConvertBits: %w", err)
	}

	return payload, nil
}

func IsBolt12Offer(offer string) bool {
	return strings.HasPrefix(strings.ToLower(offer), "lno")
}

func SatsFromBolt12Offer(offer string) int {
	decodedOffer, err := DecodeBolt12Offer(offer)
	if err != nil {
		return 0
	}
	return int(decodedOffer.AmountInSats)
}

func IsValidBolt12Offer(offer string) bool {
	return SatsFromBolt12Offer(offer) > 0
}

func IsBolt12Invoice(invoice string) bool {
	return strings.HasPrefix(strings.ToLower(invoice), "lni")
}
