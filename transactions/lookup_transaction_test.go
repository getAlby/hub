package transactions

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/tests"
)

func TestLookupTransaction_IncomingPayment(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	incomingTransaction, err := transactionsService.LookupTransaction(ctx, tests.MockLNClientTransaction.PaymentHash, nil, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), incomingTransaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_PENDING, incomingTransaction.State)
	assert.Equal(t, tests.MockLNClientTransaction.Preimage, *incomingTransaction.Preimage)
	assert.Zero(t, incomingTransaction.FeeReserveMsat)
}

func TestLookupTransaction_OutgoingPayment(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	outgoingTransaction, err := transactionsService.LookupTransaction(ctx, tests.MockLNClientTransaction.PaymentHash, nil, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), outgoingTransaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_PENDING, outgoingTransaction.State)
	assert.Equal(t, tests.MockLNClientTransaction.Preimage, *outgoingTransaction.Preimage)
	assert.Zero(t, outgoingTransaction.FeeReserveMsat)
}

func TestLookupTransaction_OutgoingPaymentFailedThenSettled(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	// the failed attempt has no settled_at, and Postgres sorts NULLs first in descending order
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_FAILED,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
	})
	mockPreimage := tests.MockLNClientTransaction.Preimage
	settledAt := time.Now()
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		SettledAt:      &settledAt,
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	outgoingTransaction, err := transactionsService.LookupTransaction(ctx, tests.MockLNClientTransaction.PaymentHash, nil, svc.LNClient, nil)
	require.NoError(t, err)
	require.Equal(t, constants.TRANSACTION_STATE_SETTLED, outgoingTransaction.State)
	assert.Equal(t, tests.MockLNClientTransaction.Preimage, *outgoingTransaction.Preimage)
	assert.Equal(t, settledAt.Unix(), outgoingTransaction.SettledAt.Unix())
}
