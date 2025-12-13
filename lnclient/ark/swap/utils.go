package swap

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/lightningnetwork/lnd/input"
	decodepay "github.com/nbd-wtf/ln-decodepay"
)

func DecodeInvoice(invoice string) (uint64, []byte, error) {
	bolt11, err := decodepay.Decodepay(invoice)
	if err != nil {
		return 0, nil, err
	}

	amount := uint64(bolt11.MSatoshi / 1000)
	preimageHash, err := hex.DecodeString(bolt11.PaymentHash)
	if err != nil {
		return 0, nil, err
	}

	return amount, input.Ripemd160H(preimageHash), nil
}

func parsePubkey(pubkey string) (*secp256k1.PublicKey, error) {
	if len(pubkey) <= 0 {
		return nil, nil
	}

	dec, err := hex.DecodeString(pubkey)
	if err != nil {
		return nil, fmt.Errorf("invalid pubkey: %s", err)
	}

	pk, err := secp256k1.ParsePubKey(dec)
	if err != nil {
		return nil, fmt.Errorf("invalid pubkey: %s", err)
	}

	return pk, nil
}

func Retry(
	ctx context.Context, interval time.Duration, fn func(ctx context.Context) (bool, error),
) error {
	for {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return fmt.Errorf("timed out")
			}
			return ctx.Err()
		default:
			done, err := fn(ctx)
			if err != nil {
				return err
			}
			if done {
				return nil
			}
			<-time.After(interval)
		}
	}
}
