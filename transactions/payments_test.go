package transactions

import (
	"context"
	"errors"
	"testing"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/tests"
	"github.com/stretchr/testify/assert"
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

func TestSendPaymentSync_FailedRemovesFeeReserve(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	svc.LNClient.(*tests.MockLn).PayInvoiceErrors = append(svc.LNClient.(*tests.MockLn).PayInvoiceErrors, errors.New("Some error"))
	svc.LNClient.(*tests.MockLn).PayInvoiceResponses = append(svc.LNClient.(*tests.MockLn).PayInvoiceResponses, nil)

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
