package transactions

import (
	"context"
	"testing"
	"time"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListTransactions_Paid(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	incomingTransactions, err := transactionsService.ListTransactions(ctx, 0, 0, 0, 0, false, false, nil, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(incomingTransactions))
	assert.Equal(t, uint64(123000), incomingTransactions[0].AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, incomingTransactions[0].State)
	assert.Equal(t, tests.MockLNClientTransaction.Preimage, *incomingTransactions[0].Preimage)
	assert.Zero(t, incomingTransactions[0].FeeReserveMsat)
}

func TestListTransactions_UnpaidIncoming(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now(),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_FAILED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now().Add(-1 * time.Second),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now().Add(-2 * time.Second),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_FAILED,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now().Add(-2 * time.Second),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now().Add(-2 * time.Second),
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	incomingTransactions, err := transactionsService.ListTransactions(ctx, 0, 0, 0, 0, false, true, nil, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(incomingTransactions))
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, incomingTransactions[0].State)
	assert.Equal(t, constants.TRANSACTION_STATE_FAILED, incomingTransactions[1].State)
	assert.Equal(t, constants.TRANSACTION_STATE_PENDING, incomingTransactions[2].State)
	for _, transaction := range incomingTransactions {
		assert.Equal(t, constants.TRANSACTION_TYPE_INCOMING, transaction.Type)
	}
}

func TestListTransactions_UnpaidOutgoing(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now(),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_FAILED,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now().Add(-1 * time.Second),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now().Add(-2 * time.Second),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_FAILED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now().Add(-2 * time.Second),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now().Add(-2 * time.Second),
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	outgoingTransactions, err := transactionsService.ListTransactions(ctx, 0, 0, 0, 0, true, false, nil, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(outgoingTransactions))
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, outgoingTransactions[0].State)
	assert.Equal(t, constants.TRANSACTION_STATE_FAILED, outgoingTransactions[1].State)
	assert.Equal(t, constants.TRANSACTION_STATE_PENDING, outgoingTransactions[2].State)
	for _, transaction := range outgoingTransactions {
		assert.Equal(t, constants.TRANSACTION_TYPE_OUTGOING, transaction.Type)
	}
}

func TestListTransactions_Unpaid(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now(),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_FAILED,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now().Add(-1 * time.Second),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now().Add(-2 * time.Second),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_FAILED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now().Add(-2 * time.Second),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now().Add(-2 * time.Second),
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	outgoingTransactions, err := transactionsService.ListTransactions(ctx, 0, 0, 0, 0, true, true, nil, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(outgoingTransactions))
}

func TestListTransactions_Limit(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "first",
		UpdatedAt:      time.Now().Add(1 * time.Minute),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "second",
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	incomingTransactions, err := transactionsService.ListTransactions(ctx, 0, 0, 1, 0, false, false, nil, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(incomingTransactions))
	assert.Equal(t, "first", incomingTransactions[0].Description)
}

func TestListTransactions_Offset(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "first",
		UpdatedAt:      time.Now().Add(3 * time.Minute),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "second",
		UpdatedAt:      time.Now().Add(2 * time.Minute),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "third",
		UpdatedAt:      time.Now().Add(1 * time.Minute),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "fourth",
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	incomingTransactions, err := transactionsService.ListTransactions(ctx, 0, 0, 1, 2, false, false, nil, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(incomingTransactions))
	assert.Equal(t, "third", incomingTransactions[0].Description)
}

func TestListTransactions_FromUntil(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "first",
		UpdatedAt:      time.Now().Add(10 * time.Minute),
		CreatedAt:      time.Now().Add(10 * time.Minute),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "second",
		UpdatedAt:      time.Now().Add(5 * time.Minute),
		CreatedAt:      time.Now().Add(5 * time.Minute),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "third",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	incomingTransactions, err := transactionsService.ListTransactions(ctx, uint64(time.Now().Add(4*time.Minute).Unix()), uint64(time.Now().Add(6*time.Minute).Unix()), 0, 0, false, false, nil, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(incomingTransactions))
	assert.Equal(t, "second", incomingTransactions[0].Description)
}

func TestListTransactions_FromUntilUnpaidOutgoing(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "first",
		UpdatedAt:      time.Now().Add(10 * time.Minute),
		CreatedAt:      time.Now().Add(10 * time.Minute),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "second",
		UpdatedAt:      time.Now().Add(5 * time.Minute),
		CreatedAt:      time.Now().Add(5 * time.Minute),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "second",
		UpdatedAt:      time.Now().Add(5 * time.Minute),
		CreatedAt:      time.Now().Add(5 * time.Minute),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "third",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	incomingTransactions, err := transactionsService.ListTransactions(ctx, uint64(time.Now().Add(4*time.Minute).Unix()), uint64(time.Now().Add(6*time.Minute).Unix()), 0, 0, true, false, nil, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(incomingTransactions))
	assert.Equal(t, "second", incomingTransactions[0].Description)
	assert.Equal(t, constants.TRANSACTION_TYPE_OUTGOING, incomingTransactions[0].Type)
}

func TestListTransactions_FromUntilUnpaidIncoming(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "first",
		UpdatedAt:      time.Now().Add(10 * time.Minute),
		CreatedAt:      time.Now().Add(10 * time.Minute),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "second",
		UpdatedAt:      time.Now().Add(5 * time.Minute),
		CreatedAt:      time.Now().Add(5 * time.Minute),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "second",
		UpdatedAt:      time.Now().Add(5 * time.Minute),
		CreatedAt:      time.Now().Add(5 * time.Minute),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "third",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	incomingTransactions, err := transactionsService.ListTransactions(ctx, uint64(time.Now().Add(4*time.Minute).Unix()), uint64(time.Now().Add(6*time.Minute).Unix()), 0, 0, false, true, nil, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(incomingTransactions))
	assert.Equal(t, "second", incomingTransactions[0].Description)
	assert.Equal(t, constants.TRANSACTION_TYPE_INCOMING, incomingTransactions[0].Type)
}
