package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/transactions"
)

func TestToNip47TransactionKeepsAcceptedStateAfterInvoiceExpiry(t *testing.T) {
	expiresAt := time.Now().Add(-time.Hour)
	settleDeadline := uint32(840_000)

	result := ToNip47Transaction(&transactions.Transaction{
		State:          constants.TRANSACTION_STATE_ACCEPTED,
		ExpiresAt:      &expiresAt,
		SettleDeadline: &settleDeadline,
	})

	require.Equal(t, "accepted", result.State)
	require.Equal(t, &settleDeadline, result.SettleDeadline)
}

func TestToNip47TransactionExpiresPendingState(t *testing.T) {
	expiresAt := time.Now().Add(-time.Hour)

	result := ToNip47Transaction(&transactions.Transaction{
		State:     constants.TRANSACTION_STATE_PENDING,
		ExpiresAt: &expiresAt,
	})

	require.Equal(t, "expired", result.State)
}

func TestToNip47TransactionAllowsSettledIncomingWithoutPreimage(t *testing.T) {
	settledAt := time.Now()

	result := ToNip47Transaction(&transactions.Transaction{
		State:     constants.TRANSACTION_STATE_SETTLED,
		Type:      constants.TRANSACTION_TYPE_INCOMING,
		SettledAt: &settledAt,
	})

	require.Equal(t, "settled", result.State)
	require.Equal(t, settledAt.Unix(), *result.SettledAt)
	require.Empty(t, result.Preimage)
}
