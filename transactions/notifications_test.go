package transactions

import (
	"context"
	"testing"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/tests"
	"github.com/stretchr/testify/assert"
)

func TestNotifications_ReceivedKnownPayment(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
	})

	transactionsService := NewTransactionsService(svc.DB)

	transactionsService.ConsumeEvent(ctx, &events.Event{
		Event:      "nwc_payment_received",
		Properties: tests.MockLNClientTransaction,
	}, map[string]interface{}{})

	incomingTransaction, err := transactionsService.LookupTransaction(ctx, tests.MockLNClientTransaction.PaymentHash, nil, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), incomingTransaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, incomingTransaction.State)
	assert.Equal(t, tests.MockLNClientTransaction.Preimage, *incomingTransaction.Preimage)
	assert.Zero(t, incomingTransaction.FeeReserveMsat)

	transactions := []db.Transaction{}
	result := svc.DB.Find(&transactions)
	assert.Equal(t, int64(1), result.RowsAffected)
}

func TestNotifications_ReceivedUnknownPayment(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	transactionsService := NewTransactionsService(svc.DB)

	transactionsService.ConsumeEvent(ctx, &events.Event{
		Event:      "nwc_payment_received",
		Properties: tests.MockLNClientTransaction,
	}, map[string]interface{}{})

	transactionType := constants.TRANSACTION_TYPE_INCOMING
	incomingTransaction, err := transactionsService.LookupTransaction(ctx, tests.MockLNClientTransaction.PaymentHash, &transactionType, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(tests.MockLNClientTransaction.Amount), incomingTransaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, incomingTransaction.State)
	assert.Equal(t, tests.MockLNClientTransaction.Preimage, *incomingTransaction.Preimage)
	assert.Zero(t, incomingTransaction.FeeReserveMsat)

	transactions := []db.Transaction{}
	result := svc.DB.Find(&transactions)
	assert.Equal(t, int64(1), result.RowsAffected)
}

func TestNotifications_SentKnownPayment(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		AmountMsat:     123000,
		FeeReserveMsat: uint64(10000),
	})

	transactionsService := NewTransactionsService(svc.DB)

	transactionsService.ConsumeEvent(ctx, &events.Event{
		Event:      "nwc_payment_sent",
		Properties: tests.MockLNClientTransaction,
	}, map[string]interface{}{})

	transactionType := constants.TRANSACTION_TYPE_OUTGOING
	outgoingTransaction, err := transactionsService.LookupTransaction(ctx, tests.MockLNClientTransaction.PaymentHash, &transactionType, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), outgoingTransaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, outgoingTransaction.State)
	assert.Equal(t, tests.MockLNClientTransaction.Preimage, *outgoingTransaction.Preimage)
	assert.Zero(t, outgoingTransaction.FeeReserveMsat)

	transactions := []db.Transaction{}
	result := svc.DB.Find(&transactions)
	assert.Equal(t, int64(1), result.RowsAffected)
}

func TestNotifications_SentUnknownPayment(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	transactionsService := NewTransactionsService(svc.DB)

	transactions := []db.Transaction{}
	result := svc.DB.Find(&transactions)
	assert.Equal(t, int64(0), result.RowsAffected)

	transactionsService.ConsumeEvent(ctx, &events.Event{
		Event:      "nwc_payment_sent",
		Properties: tests.MockLNClientTransaction,
	}, map[string]interface{}{})

	transactionType := constants.TRANSACTION_TYPE_OUTGOING
	outgoingTransaction, err := transactionsService.LookupTransaction(ctx, tests.MockLNClientTransaction.PaymentHash, &transactionType, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(tests.MockLNClientTransaction.Amount), outgoingTransaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, outgoingTransaction.State)
	assert.Equal(t, tests.MockLNClientTransaction.Preimage, *outgoingTransaction.Preimage)
	assert.Zero(t, outgoingTransaction.FeeReserveMsat)

	transactions = []db.Transaction{}
	result = svc.DB.Find(&transactions)
	assert.Equal(t, int64(1), result.RowsAffected)
}

func TestNotifications_FailedKnownPayment(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		AmountMsat:     123000,
		FeeReserveMsat: uint64(10000),
	})

	transactionsService := NewTransactionsService(svc.DB)

	transactionsService.ConsumeEvent(ctx, &events.Event{
		Event:      "nwc_payment_failed_async",
		Properties: tests.MockLNClientTransaction,
	}, map[string]interface{}{})

	transactionType := constants.TRANSACTION_TYPE_OUTGOING
	outgoingTransaction, err := transactionsService.LookupTransaction(ctx, tests.MockLNClientTransaction.PaymentHash, &transactionType, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, constants.TRANSACTION_STATE_FAILED, outgoingTransaction.State)
	assert.Nil(t, outgoingTransaction.Preimage)
	assert.Zero(t, outgoingTransaction.FeeReserveMsat)

	transactions := []db.Transaction{}
	result := svc.DB.Find(&transactions)
	assert.Equal(t, int64(1), result.RowsAffected)
}
