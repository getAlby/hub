package transactions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/tests"
)

func TestSendPaymentSync_NoApp(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	metadata := map[string]interface{}{
		"a": 123,
	}

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockLNClientTransaction.Invoice, nil, metadata, svc.LNClient, nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Zero(t, transaction.FeeReserveMsat)
	assert.Equal(t, "123preimage", *transaction.Preimage)

	type dummyMetadata struct {
		A int `json:"a"`
	}
	var decodedMetadata dummyMetadata
	err = json.Unmarshal(transaction.Metadata, &decodedMetadata)
	assert.NoError(t, err)
	assert.Equal(t, 123, decodedMetadata.A)
}

func TestSendPaymentSync_0Amount(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	metadata := map[string]interface{}{
		"a": 123,
	}

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	amount := uint64(1234)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.Mock0AmountInvoice, &amount, metadata, svc.LNClient, nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, amount, transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Zero(t, transaction.FeeReserveMsat)
	assert.Equal(t, "123preimage", *transaction.Preimage)
}

func TestSendPaymentSync_MetadataTooLarge(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	metadata := make(map[string]interface{})
	metadata["randomkey"] = strings.Repeat("a", constants.INVOICE_METADATA_MAX_LENGTH-15) // json encoding adds 16 characters

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockLNClientTransaction.Invoice, nil, metadata, svc.LNClient, nil, nil)

	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf("encoded payment metadata provided is too large. Limit: %d Received: %d", constants.INVOICE_METADATA_MAX_LENGTH, constants.INVOICE_METADATA_MAX_LENGTH+1), err.Error())
	assert.Nil(t, transaction)
}

func TestSendPaymentSync_Duplicate(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	svc.DB.Create(&db.Transaction{
		State:       constants.TRANSACTION_STATE_SETTLED,
		Type:        constants.TRANSACTION_TYPE_OUTGOING,
		PaymentHash: tests.MockLNClientTransaction.PaymentHash,
		AmountMsat:  123000,
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockLNClientTransaction.Invoice, nil, nil, svc.LNClient, nil, nil)

	assert.Error(t, err)
	assert.Equal(t, "this invoice has already been paid", err.Error())
	assert.Nil(t, transaction)
}

func TestMarkSettled_Sent(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

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
		_, err = transactionsService.markTransactionSettled(tx, &dbTransaction, "test", 0, false)
		return err
	})

	assert.NoError(t, err)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, dbTransaction.State)
	assert.Equal(t, 1, len(mockEventConsumer.GetConsumedEvents()))
	assert.Equal(t, "nwc_payment_sent", mockEventConsumer.GetConsumedEvents()[0].Event)
	settledTransaction := mockEventConsumer.GetConsumedEvents()[0].Properties.(*db.Transaction)
	assert.Equal(t, &dbTransaction, settledTransaction)
}

func TestMarkSettled_Received(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

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
		_, err = transactionsService.markTransactionSettled(tx, &dbTransaction, "test", 0, false)
		return err
	})

	assert.NoError(t, err)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, dbTransaction.State)
	assert.Equal(t, 1, len(mockEventConsumer.GetConsumedEvents()))
	assert.Equal(t, "nwc_payment_received", mockEventConsumer.GetConsumedEvents()[0].Event)
	settledTransaction := mockEventConsumer.GetConsumedEvents()[0].Properties.(*db.Transaction)
	assert.Equal(t, &dbTransaction, settledTransaction)
}

func TestDoNotMarkSettledTwice(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

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
		_, err = transactionsService.markTransactionSettled(tx, &dbTransaction, "test", 0, false)
		return err
	})

	assert.NoError(t, err)
	assert.Zero(t, len(mockEventConsumer.GetConsumedEvents()))
}

func TestMarkFailed(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

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

	assert.NoError(t, err)
	assert.Equal(t, constants.TRANSACTION_STATE_FAILED, dbTransaction.State)
	assert.Equal(t, 1, len(mockEventConsumer.GetConsumedEvents()))
	assert.Equal(t, "nwc_payment_failed", mockEventConsumer.GetConsumedEvents()[0].Event)
	settledTransaction := mockEventConsumer.GetConsumedEvents()[0].Properties.(*db.Transaction)
	assert.Equal(t, &dbTransaction, settledTransaction)
	assert.Equal(t, "some routing error", settledTransaction.FailureReason)
}

func TestDoNotMarkFailedTwice(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

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

	assert.NoError(t, err)
	assert.Equal(t, updatedAt, dbTransaction.UpdatedAt)
	assert.Zero(t, len(mockEventConsumer.GetConsumedEvents()))
}

func TestSendPaymentSync_FailedRemovesFeeReserve(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	svc.LNClient.(*tests.MockLn).PayInvoiceErrors = append(svc.LNClient.(*tests.MockLn).PayInvoiceErrors, errors.New("Some error"))
	svc.LNClient.(*tests.MockLn).PayInvoiceResponses = append(svc.LNClient.(*tests.MockLn).PayInvoiceResponses, nil)

	mockEventConsumer := tests.NewMockEventConsumer()
	svc.EventPublisher.RegisterSubscriber(mockEventConsumer)

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockLNClientTransaction.Invoice, nil, nil, svc.LNClient, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, transaction)

	transactionType := constants.TRANSACTION_TYPE_OUTGOING
	transaction, err = transactionsService.LookupTransaction(ctx, tests.MockLNClientTransaction.PaymentHash, &transactionType, svc.LNClient, nil)
	assert.NoError(t, err)

	assert.Equal(t, uint64(123000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_FAILED, transaction.State)
	assert.Zero(t, transaction.FeeReserveMsat)
	assert.Nil(t, transaction.Preimage)

	assert.Equal(t, 1, len(mockEventConsumer.GetConsumedEvents()))
	assert.Equal(t, "nwc_payment_failed", mockEventConsumer.GetConsumedEvents()[0].Event)
}

func TestSendPaymentSync_PendingHasFeeReserve(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	// timeout will leave the payment as pending
	svc.LNClient.(*tests.MockLn).PayInvoiceErrors = append(svc.LNClient.(*tests.MockLn).PayInvoiceErrors, lnclient.NewTimeoutError())
	svc.LNClient.(*tests.MockLn).PayInvoiceResponses = append(svc.LNClient.(*tests.MockLn).PayInvoiceResponses, nil)

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockLNClientTransaction.Invoice, nil, nil, svc.LNClient, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, transaction)

	transactionType := constants.TRANSACTION_TYPE_OUTGOING
	transaction, err = transactionsService.LookupTransaction(ctx, tests.MockLNClientTransaction.PaymentHash, &transactionType, svc.LNClient, nil)
	assert.NoError(t, err)

	assert.Equal(t, uint64(123000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_PENDING, transaction.State)
	assert.Equal(t, uint64(10000), transaction.FeeReserveMsat)
	assert.Nil(t, transaction.Preimage)
}
