package transactions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/tests"
)

func TestSendPaymentSync_NoApp(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	metadata := map[string]interface{}{
		"a": 123,
	}

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(tests.MockLNClientTransaction.Invoice, nil, metadata, svc.LNClient, nil, nil)

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

func TestSendPaymentSync_ZeroAmount(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	metadata := map[string]interface{}{
		"a": 123,
	}

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	amount := uint64(1234)
	transaction, err := transactionsService.SendPaymentSync(tests.MockZeroAmountInvoice, &amount, metadata, svc.LNClient, nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, amount, transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Zero(t, transaction.FeeReserveMsat)
	assert.Equal(t, "123preimage", *transaction.Preimage)
}

func TestSendPaymentSync_AmountOnNonZeroAmountInvoice(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	metadata := map[string]interface{}{
		"a": 123,
	}

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	amount := uint64(1234)
	transaction, err := transactionsService.SendPaymentSync(tests.MockInvoice, &amount, metadata, svc.LNClient, nil, nil)

	assert.NoError(t, err)
	// amount is from the invoice, not what was specified
	assert.Equal(t, uint64(123_000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Zero(t, transaction.FeeReserveMsat)
	assert.Equal(t, "123preimage", *transaction.Preimage)
}

func TestSendPaymentSync_MetadataTooLarge(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	metadata := make(map[string]interface{})
	metadata["randomkey"] = strings.Repeat("a", constants.INVOICE_METADATA_MAX_LENGTH-15) // json encoding adds 16 characters

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(tests.MockLNClientTransaction.Invoice, nil, metadata, svc.LNClient, nil, nil)

	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf("encoded payment metadata provided is too large. Limit: %d Received: %d", constants.INVOICE_METADATA_MAX_LENGTH, constants.INVOICE_METADATA_MAX_LENGTH+1), err.Error())
	assert.Nil(t, transaction)
}

func TestSendPaymentSync_Duplicate_AlreadyPaid(t *testing.T) {
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
	transaction, err := transactionsService.SendPaymentSync(tests.MockLNClientTransaction.Invoice, nil, nil, svc.LNClient, nil, nil)

	assert.Error(t, err)
	assert.Equal(t, "this invoice has already been paid", err.Error())
	assert.Nil(t, transaction)
}

func TestSendPaymentSync_Duplicate_Pending(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	svc.DB.Create(&db.Transaction{
		State:       constants.TRANSACTION_STATE_PENDING,
		Type:        constants.TRANSACTION_TYPE_OUTGOING,
		PaymentHash: tests.MockLNClientTransaction.PaymentHash,
		AmountMsat:  123000,
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(tests.MockLNClientTransaction.Invoice, nil, nil, svc.LNClient, nil, nil)

	assert.Error(t, err)
	assert.Equal(t, "there is already a payment pending for this invoice", err.Error())
	assert.Nil(t, transaction)
}

func TestSendPaymentSync_Duplicate_Failed(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	svc.DB.Create(&db.Transaction{
		State:       constants.TRANSACTION_STATE_FAILED,
		Type:        constants.TRANSACTION_TYPE_OUTGOING,
		PaymentHash: tests.MockLNClientTransaction.PaymentHash,
		AmountMsat:  123000,
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	_, err = transactionsService.SendPaymentSync(tests.MockLNClientTransaction.Invoice, nil, nil, svc.LNClient, nil, nil)

	assert.NoError(t, err)
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

func TestMarkSettled_Twice(t *testing.T) {
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
	var wg sync.WaitGroup
	n := 10
	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			err = svc.DB.Transaction(func(tx *gorm.DB) error {
				time.Sleep(time.Duration(n) * 10 * time.Millisecond)
				_, err = transactionsService.markTransactionSettled(tx, &dbTransaction, "test", 0, false)
				time.Sleep(time.Duration(n) * 10 * time.Millisecond)
				return err
			})
			require.NoError(t, err)
		}()
	}
	wg.Wait()

	// ensure we only mark transaction settled once and only fire
	// settled notifications once
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
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	svc.LNClient.(*tests.MockLn).PayInvoiceErrors = append(svc.LNClient.(*tests.MockLn).PayInvoiceErrors, errors.New("Some error"))
	svc.LNClient.(*tests.MockLn).PayInvoiceResponses = append(svc.LNClient.(*tests.MockLn).PayInvoiceResponses, nil)

	mockEventConsumer := tests.NewMockEventConsumer()
	svc.EventPublisher.RegisterSubscriber(mockEventConsumer)

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(tests.MockLNClientTransaction.Invoice, nil, nil, svc.LNClient, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, transaction)

	transactionType := constants.TRANSACTION_TYPE_OUTGOING
	transaction, err = transactionsService.LookupTransaction(context.TODO(), tests.MockLNClientTransaction.PaymentHash, &transactionType, svc.LNClient, nil)
	assert.NoError(t, err)

	assert.Equal(t, uint64(123000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_FAILED, transaction.State)
	assert.Zero(t, transaction.FeeReserveMsat)
	assert.Nil(t, transaction.Preimage)

	assert.Equal(t, 1, len(mockEventConsumer.GetConsumedEvents()))
	assert.Equal(t, "nwc_payment_failed", mockEventConsumer.GetConsumedEvents()[0].Event)
}

func TestSendPaymentSync_PendingHasFeeReserve(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	// fake a delay to ensure the payment is still pending
	delay := 10 * time.Second
	svc.LNClient.(*tests.MockLn).PaymentDelay = &delay

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	go func() {
		transactionsService.SendPaymentSync(tests.MockLNClientTransaction.Invoice, nil, nil, svc.LNClient, nil, nil)
	}()
	// ensure the goroutine above runs first
	time.Sleep(10 * time.Millisecond)

	transactionType := constants.TRANSACTION_TYPE_OUTGOING
	transaction, err := transactionsService.LookupTransaction(context.TODO(), tests.MockLNClientTransaction.PaymentHash, &transactionType, svc.LNClient, nil)
	assert.NoError(t, err)

	assert.Equal(t, uint64(123000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_PENDING, transaction.State)
	assert.Equal(t, uint64(10000), transaction.FeeReserveMsat)
	assert.Nil(t, transaction.Preimage)
}

func TestConsumeEvent_FailedMarkedAsSuccessful(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	svc.LNClient.(*tests.MockLn).PayInvoiceErrors = append(svc.LNClient.(*tests.MockLn).PayInvoiceErrors, errors.New("some error"))
	svc.LNClient.(*tests.MockLn).PayInvoiceResponses = append(svc.LNClient.(*tests.MockLn).PayInvoiceResponses, nil)

	transaction, err := transactionsService.SendPaymentSync(tests.MockLNClientTransaction.Invoice, nil, nil, svc.LNClient, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, transaction)

	var transactions []db.Transaction
	result := svc.DB.Find(&transactions, &db.Transaction{
		Type:        constants.TRANSACTION_TYPE_OUTGOING,
		PaymentHash: tests.MockLNClientTransaction.PaymentHash,
	})
	assert.NoError(t, result.Error)
	assert.Equal(t, 1, len(transactions))

	transaction = &transactions[0]
	assert.Equal(t, constants.TRANSACTION_STATE_FAILED, transaction.State)

	// Now that we have a failed transaction, we submit a "nwc_lnclient_payment_sent" event.
	// This should be marked as successful as long as there are no pending payments for the
	// same payment hash

	transactionsService.ConsumeEvent(context.TODO(), &events.Event{
		Event: "nwc_lnclient_payment_sent",
		Properties: &lnclient.Transaction{
			Type:            tests.MockLNClientTransaction.Type,
			Invoice:         tests.MockLNClientTransaction.Invoice,
			Description:     tests.MockLNClientTransaction.Description,
			DescriptionHash: tests.MockLNClientTransaction.DescriptionHash,
			Preimage:        tests.MockLNClientTransaction.Preimage,
			PaymentHash:     tests.MockLNClientTransaction.PaymentHash,
			Amount:          tests.MockLNClientTransaction.Amount,
			FeesPaid:        tests.MockLNClientTransaction.FeesPaid,
		},
	}, nil)

	// Re-read transactions and ensure that the single returned transaction
	// is now settled.
	result = svc.DB.Find(&transactions, &db.Transaction{
		Type:        constants.TRANSACTION_TYPE_OUTGOING,
		PaymentHash: tests.MockLNClientTransaction.PaymentHash,
	})
	assert.NoError(t, result.Error)
	assert.Equal(t, 1, len(transactions))

	transaction = &transactions[0]
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
}
