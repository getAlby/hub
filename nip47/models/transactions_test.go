package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/transactions"
)

func TestBolt12HistoryPreservesRecoveryFields(t *testing.T) {
	for _, state := range []string{constants.TRANSACTION_STATE_PENDING, constants.TRANSACTION_STATE_FAILED, constants.TRANSACTION_STATE_SETTLED} {
		result := ToNip47Transaction(&transactions.Transaction{
			ID: 42, Type: "outgoing", State: state, PaymentRequest: "lno1offer",
			PaymentHash: "learned-hash", FailureReason: "invoice request expired",
			Metadata: []byte(`{"bitblik_attempt":"attempt-id"}`),
		})
		encoded, err := json.Marshal(result)
		require.NoError(t, err)
		var fields map[string]interface{}
		require.NoError(t, json.Unmarshal(encoded, &fields))
		require.Equal(t, "42", fields["transaction_id"])
		require.Equal(t, "bolt12", fields["instruction_type"])
		require.Equal(t, strings.ToLower(state), fields["state"])
		require.Equal(t, "invoice request expired", fields["failure_reason"])
		require.Equal(t, "attempt-id", fields["metadata"].(map[string]interface{})["bitblik_attempt"])
	}
}

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
