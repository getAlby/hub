package transactions

import (
	"context"
	"testing"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/tests"
	"github.com/stretchr/testify/assert"
)

func TestLookupTransaction_IncomingPayment(t *testing.T) {
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

	incomingTransaction, err := transactionsService.LookupTransaction(ctx, tests.MockLNClientTransaction.PaymentHash, nil, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), incomingTransaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_PENDING, incomingTransaction.State)
	assert.Equal(t, tests.MockLNClientTransaction.Preimage, *incomingTransaction.Preimage)
	assert.Nil(t, incomingTransaction.FeeReserveMsat)
}

func TestLookupTransaction_OutgoingPayment(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
	})

	transactionsService := NewTransactionsService(svc.DB)

	outgoingTransaction, err := transactionsService.LookupTransaction(ctx, tests.MockLNClientTransaction.PaymentHash, nil, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), outgoingTransaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_PENDING, outgoingTransaction.State)
	assert.Equal(t, tests.MockLNClientTransaction.Preimage, *outgoingTransaction.Preimage)
	assert.Nil(t, outgoingTransaction.FeeReserveMsat)
}
