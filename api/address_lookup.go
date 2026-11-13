package api

import (
	"context"
	"fmt"
	"strconv"

	"github.com/getAlby/hub/chain"
	"github.com/getAlby/hub/config"
)

func (api *api) addressHasTransactions(ctx context.Context, address string) (bool, error) {
	lookup, err := chain.NewAddressLookup(ctx, api.cfg)
	if err != nil {
		return false, err
	}
	defer lookup.Close()

	return lookup.AddressHasTransactions(ctx, address)
}

// checkIncomingOnchainPaymentSinceAddressIssued returns true if the wallet
// received an on-chain payment since the current receive address was issued.
func (api *api) checkIncomingOnchainPaymentSinceAddressIssued(ctx context.Context) (bool, error) {
	lnClient := api.svc.GetLNClient()
	if lnClient == nil {
		return false, ErrLNClientNotStarted
	}

	issuedAtStr, err := api.cfg.Get(config.OnchainAddressIssuedAtKey, "")
	if err != nil {
		return false, err
	}
	// addresses issued before the issue time was saved have issuedAt 0, so
	// they are rotated if the wallet has ever received an on-chain payment
	var issuedAt uint64
	if issuedAtStr != "" {
		issuedAt, err = strconv.ParseUint(issuedAtStr, 10, 64)
		if err != nil {
			return false, fmt.Errorf("failed to parse onchain address issue time: %w", err)
		}
	}

	transactions, err := lnClient.ListOnchainTransactions(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to list onchain transactions: %w", err)
	}

	for _, transaction := range transactions {
		if transaction.Type == "incoming" && transaction.CreatedAt >= issuedAt {
			return true, nil
		}
	}
	return false, nil
}
