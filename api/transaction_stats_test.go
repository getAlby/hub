package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/tests"
)

func TestGetTransactionStats(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	// Two settled outgoing payments — these count towards the stats.
	svc.DB.Create(&db.Transaction{
		State:       constants.TRANSACTION_STATE_SETTLED,
		Type:        constants.TRANSACTION_TYPE_OUTGOING,
		PaymentHash: "hash1",
		AmountMsat:  1_000_000,
		FeeMsat:     3000,
	})
	svc.DB.Create(&db.Transaction{
		State:       constants.TRANSACTION_STATE_SETTLED,
		Type:        constants.TRANSACTION_TYPE_OUTGOING,
		PaymentHash: "hash2",
		AmountMsat:  500_000,
		FeeMsat:     2000,
	})
	// Excluded: pending outgoing.
	svc.DB.Create(&db.Transaction{
		State:       constants.TRANSACTION_STATE_PENDING,
		Type:        constants.TRANSACTION_TYPE_OUTGOING,
		PaymentHash: "hash3",
		AmountMsat:  999_000,
		FeeMsat:     9000,
	})
	// Excluded: settled incoming.
	svc.DB.Create(&db.Transaction{
		State:       constants.TRANSACTION_STATE_SETTLED,
		Type:        constants.TRANSACTION_TYPE_INCOMING,
		PaymentHash: "hash4",
		AmountMsat:  777_000,
	})
	// Excluded: self-payment (never traverses the network).
	svc.DB.Create(&db.Transaction{
		State:       constants.TRANSACTION_STATE_SETTLED,
		Type:        constants.TRANSACTION_TYPE_OUTGOING,
		PaymentHash: "hash5",
		AmountMsat:  200_000,
		FeeMsat:     0,
		SelfPayment: true,
	})

	theAPI := &api{db: svc.DB}

	stats, err := theAPI.GetTransactionStats()
	require.NoError(t, err)
	require.NotNil(t, stats)

	assert.Equal(t, uint64(1_500_000), stats.TotalVolumeMsat)
	assert.Equal(t, uint64(1500), stats.TotalVolumeSat)
	assert.Equal(t, uint64(5000), stats.TotalFeesPaidMsat)
	assert.Equal(t, uint64(5), stats.TotalFeesPaidSat)
	assert.Equal(t, uint64(2), stats.NumPayments)
}

func TestGetTransactionStats_Empty(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	theAPI := &api{db: svc.DB}

	stats, err := theAPI.GetTransactionStats()
	require.NoError(t, err)
	require.NotNil(t, stats)

	assert.Equal(t, uint64(0), stats.TotalVolumeMsat)
	assert.Equal(t, uint64(0), stats.TotalFeesPaidMsat)
	assert.Equal(t, uint64(0), stats.NumPayments)
}
