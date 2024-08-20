package transactions

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/tests"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestSendPaymentSync_NoApp(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockLNClientTransaction.Invoice, svc.LNClient, nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Zero(t, transaction.FeeReserveMsat)
	assert.Equal(t, "123preimage", *transaction.Preimage)
}

func TestSendPaymentSync_Duplicate(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	svc.DB.Create(&db.Transaction{
		State:       constants.TRANSACTION_STATE_SETTLED,
		Type:        constants.TRANSACTION_TYPE_OUTGOING,
		PaymentHash: tests.MockLNClientTransaction.PaymentHash,
		AmountMsat:  123000,
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockLNClientTransaction.Invoice, svc.LNClient, nil, nil)

	assert.Error(t, err)
	assert.Equal(t, "this invoice has already been paid", err.Error())
	assert.Nil(t, transaction)
}

func TestMarkSettled_Sent(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	dbTransaction := db.Transaction{
		State:       constants.TRANSACTION_STATE_PENDING,
		Type:        constants.TRANSACTION_TYPE_OUTGOING,
		PaymentHash: tests.MockLNClientTransaction.PaymentHash,
		AmountMsat:  123000,
	}
	svc.DB.Create(&dbTransaction)

	mockEventConsumer := tests.NewMockEventConsumer()
	svc.EventPublisher.RegisterSubscriber(mockEventConsumer)
	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	err = svc.DB.Transaction(func(tx *gorm.DB) error {
		return transactionsService.markTransactionSettled(tx, &dbTransaction, "test", 0, false)
	})

	assert.Nil(t, err)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, dbTransaction.State)
	assert.Equal(t, 1, len(mockEventConsumer.GetConsumeEvents()))
	assert.Equal(t, "nwc_payment_sent", mockEventConsumer.GetConsumeEvents()[0].Event)
	settledTransaction := mockEventConsumer.GetConsumeEvents()[0].Properties.(*db.Transaction)
	assert.Equal(t, &dbTransaction, settledTransaction)
}

func TestMarkSettled_Received(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	dbTransaction := db.Transaction{
		State:       constants.TRANSACTION_STATE_PENDING,
		Type:        constants.TRANSACTION_TYPE_INCOMING,
		PaymentHash: tests.MockLNClientTransaction.PaymentHash,
		AmountMsat:  123000,
	}
	svc.DB.Create(&dbTransaction)

	mockEventConsumer := tests.NewMockEventConsumer()
	svc.EventPublisher.RegisterSubscriber(mockEventConsumer)
	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	err = svc.DB.Transaction(func(tx *gorm.DB) error {
		return transactionsService.markTransactionSettled(tx, &dbTransaction, "test", 0, false)
	})

	assert.Nil(t, err)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, dbTransaction.State)
	assert.Equal(t, 1, len(mockEventConsumer.GetConsumeEvents()))
	assert.Equal(t, "nwc_payment_received", mockEventConsumer.GetConsumeEvents()[0].Event)
	settledTransaction := mockEventConsumer.GetConsumeEvents()[0].Properties.(*db.Transaction)
	assert.Equal(t, &dbTransaction, settledTransaction)
}

func TestDoNotMarkSettledTwice(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	settledAt := time.Now().Add(time.Duration(-1) * time.Minute)
	dbTransaction := db.Transaction{
		State:       constants.TRANSACTION_STATE_SETTLED,
		Type:        constants.TRANSACTION_TYPE_OUTGOING,
		PaymentHash: tests.MockLNClientTransaction.PaymentHash,
		AmountMsat:  123000,
		SettledAt:   &settledAt,
	}
	svc.DB.Create(&dbTransaction)

	mockEventConsumer := tests.NewMockEventConsumer()
	svc.EventPublisher.RegisterSubscriber(mockEventConsumer)
	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	err = svc.DB.Transaction(func(tx *gorm.DB) error {
		return transactionsService.markTransactionSettled(tx, &dbTransaction, "test", 0, false)
	})

	assert.Nil(t, err)
	assert.Equal(t, settledAt, *dbTransaction.SettledAt)
	assert.Zero(t, len(mockEventConsumer.GetConsumeEvents()))
}

func TestMarkFailed(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	dbTransaction := db.Transaction{
		State:       constants.TRANSACTION_STATE_PENDING,
		Type:        constants.TRANSACTION_TYPE_OUTGOING,
		PaymentHash: tests.MockLNClientTransaction.PaymentHash,
		AmountMsat:  123000,
	}
	svc.DB.Create(&dbTransaction)

	mockEventConsumer := tests.NewMockEventConsumer()
	svc.EventPublisher.RegisterSubscriber(mockEventConsumer)
	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	err = svc.DB.Transaction(func(tx *gorm.DB) error {
		return transactionsService.markPaymentFailed(tx, &dbTransaction, "some routing error")
	})

	assert.Nil(t, err)
	assert.Equal(t, constants.TRANSACTION_STATE_FAILED, dbTransaction.State)
	assert.Equal(t, 1, len(mockEventConsumer.GetConsumeEvents()))
	assert.Equal(t, "nwc_payment_failed", mockEventConsumer.GetConsumeEvents()[0].Event)
	settledTransaction := mockEventConsumer.GetConsumeEvents()[0].Properties.(*db.Transaction)
	assert.Equal(t, &dbTransaction, settledTransaction)
	assert.Equal(t, "some routing error", settledTransaction.FailureReason)
}

func TestDoNotMarkFailedTwice(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	updatedAt := time.Now().Add(time.Duration(-1) * time.Minute)
	dbTransaction := db.Transaction{
		State:       constants.TRANSACTION_STATE_FAILED,
		Type:        constants.TRANSACTION_TYPE_OUTGOING,
		PaymentHash: tests.MockLNClientTransaction.PaymentHash,
		AmountMsat:  123000,
		UpdatedAt:   updatedAt,
	}
	svc.DB.Create(&dbTransaction)

	mockEventConsumer := tests.NewMockEventConsumer()
	svc.EventPublisher.RegisterSubscriber(mockEventConsumer)
	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	err = svc.DB.Transaction(func(tx *gorm.DB) error {
		return transactionsService.markPaymentFailed(tx, &dbTransaction, "some routing error")
	})

	assert.Nil(t, err)
	assert.Equal(t, updatedAt, dbTransaction.UpdatedAt)
	assert.Zero(t, len(mockEventConsumer.GetConsumeEvents()))
}

func TestSendPaymentSync_FailedRemovesFeeReserve(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	svc.LNClient.(*tests.MockLn).PayInvoiceErrors = append(svc.LNClient.(*tests.MockLn).PayInvoiceErrors, errors.New("Some error"))
	svc.LNClient.(*tests.MockLn).PayInvoiceResponses = append(svc.LNClient.(*tests.MockLn).PayInvoiceResponses, nil)

	mockEventConsumer := tests.NewMockEventConsumer()
	svc.EventPublisher.RegisterSubscriber(mockEventConsumer)

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockLNClientTransaction.Invoice, svc.LNClient, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, transaction)

	transactionType := constants.TRANSACTION_TYPE_OUTGOING
	transaction, err = transactionsService.LookupTransaction(ctx, tests.MockLNClientTransaction.PaymentHash, &transactionType, svc.LNClient, nil)
	assert.Nil(t, err)

	assert.Equal(t, uint64(123000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_FAILED, transaction.State)
	assert.Zero(t, transaction.FeeReserveMsat)
	assert.Nil(t, transaction.Preimage)

	assert.Equal(t, 1, len(mockEventConsumer.GetConsumeEvents()))
	assert.Equal(t, "nwc_payment_failed", mockEventConsumer.GetConsumeEvents()[0].Event)
}

func TestSendPaymentSync_PendingHasFeeReserve(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	// timeout will leave the payment as pending
	svc.LNClient.(*tests.MockLn).PayInvoiceErrors = append(svc.LNClient.(*tests.MockLn).PayInvoiceErrors, lnclient.NewTimeoutError())
	svc.LNClient.(*tests.MockLn).PayInvoiceResponses = append(svc.LNClient.(*tests.MockLn).PayInvoiceResponses, nil)

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockLNClientTransaction.Invoice, svc.LNClient, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, transaction)

	transactionType := constants.TRANSACTION_TYPE_OUTGOING
	transaction, err = transactionsService.LookupTransaction(ctx, tests.MockLNClientTransaction.PaymentHash, &transactionType, svc.LNClient, nil)
	assert.Nil(t, err)

	assert.Equal(t, uint64(123000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_PENDING, transaction.State)
	assert.Equal(t, uint64(10000), transaction.FeeReserveMsat)
	assert.Nil(t, transaction.Preimage)
}
