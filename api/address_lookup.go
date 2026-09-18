package api

import (
	"context"

	"github.com/getAlby/hub/chain"
)

func (api *api) addressHasTransactions(ctx context.Context, address string) (bool, error) {
	lookup, err := chain.NewAddressLookup(ctx, api.cfg)
	if err != nil {
		return false, err
	}
	defer lookup.Close()

	return lookup.AddressHasTransactions(ctx, address)
}
