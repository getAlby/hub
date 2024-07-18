package transactions

import (
	"context"
	"testing"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/tests"
	"github.com/stretchr/testify/assert"
)

func TestMakeInvoice_NoApp(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	transactionsService := NewTransactionsService(svc.DB)
	transaction, err := transactionsService.MakeInvoice(ctx, 1234, "Hello world", "", 0, svc.LNClient, nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, uint64(tests.MockLNClientTransaction.Amount), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_PENDING, transaction.State)
	assert.Equal(t, tests.MockLNClientTransaction.Preimage, *transaction.Preimage)
}

func TestSendPaymentSync_NoApp(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	transactionsService := NewTransactionsService(svc.DB)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockLNClientTransaction.Invoice, svc.LNClient, nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, "123preimage", *transaction.Preimage)
}

// TODO: apps & events
// TODO: isolated apps & events
// TODO: self payments

// TODO: keysend
// TODO: lookup invoice
// TODO: list transactions
