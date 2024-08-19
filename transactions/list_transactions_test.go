package transactions

import (
	"context"
	"testing"
	"time"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/tests"
	"github.com/stretchr/testify/assert"
)

func TestListTransactions(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

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

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	incomingTransactions, err := transactionsService.ListTransactions(ctx, 0, 0, 0, 0, false, nil, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(incomingTransactions))
	assert.Equal(t, uint64(123000), incomingTransactions[0].AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, incomingTransactions[0].State)
	assert.Equal(t, tests.MockLNClientTransaction.Preimage, *incomingTransactions[0].Preimage)
	assert.Zero(t, incomingTransactions[0].FeeReserveMsat)
}

func TestListTransactions_Unsettled(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

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

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	incomingTransactions, err := transactionsService.ListTransactions(ctx, 0, 0, 0, 0, true, nil, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(incomingTransactions))
}

func TestListTransactions_Limit(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "first",
		CreatedAt:      time.Now().Add(1 * time.Minute),
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

	incomingTransactions, err := transactionsService.ListTransactions(ctx, 0, 0, 1, 0, false, nil, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(incomingTransactions))
	assert.Equal(t, "first", incomingTransactions[0].Description)
}

func TestListTransactions_Offset(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "first",
		CreatedAt:      time.Now().Add(1 * time.Minute),
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

	incomingTransactions, err := transactionsService.ListTransactions(ctx, 0, 0, 1, 1, false, nil, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(incomingTransactions))
	assert.Equal(t, "second", incomingTransactions[0].Description)
}

func TestListTransactions_FromUntil(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "first",
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
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	incomingTransactions, err := transactionsService.ListTransactions(ctx, uint64(time.Now().Add(4*time.Minute).Unix()), uint64(time.Now().Add(6*time.Minute).Unix()), 0, 0, false, nil, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(incomingTransactions))
	assert.Equal(t, "second", incomingTransactions[0].Description)
}
