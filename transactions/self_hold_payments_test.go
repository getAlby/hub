package transactions

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/tests"
)

func TestSelfHoldPaymentSettled(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	preimage := tests.MockLNClientHoldTransaction.Preimage
	paymentHash := tests.MockLNClientHoldTransaction.PaymentHash

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.MakeHoldInvoice(ctx, 1000, "Hold payment test", "", 0, paymentHash, nil, svc.LNClient, nil, nil)
	require.NoError(t, err)
	assert.True(t, transaction.Hold)
	// use the pubkey from the decoded tests.MockLNClientHoldTransaction invoice
	svc.LNClient.(*tests.MockLn).Pubkey = "038a73de75fdc3c7ec951a5e0b0fa95c5cd353bd7ca72df2086aa228c1f92819a5"

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		result, err := transactionsService.SendPaymentSync(ctx, transaction.PaymentRequest, nil, nil, svc.LNClient, nil, nil)
		assert.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, result.State)
		assert.Equal(t, true, result.SelfPayment)
		assert.Equal(t, false, result.Hold)
	}()

	// wait for payment to be accepted
	time.Sleep(10 * time.Millisecond)
	settledTransaction, err := transactionsService.SettleHoldInvoice(ctx, preimage, svc.LNClient)
	assert.NoError(t, err)
	require.NotNil(t, settledTransaction)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, settledTransaction.State)
	assert.Equal(t, true, settledTransaction.SelfPayment)
	assert.Equal(t, true, settledTransaction.Hold)
	wg.Wait()
}
func TestSelfHoldPaymentCanceled(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	paymentHash := tests.MockLNClientHoldTransaction.PaymentHash

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.MakeHoldInvoice(ctx, 1000, "Hold payment test", "", 0, paymentHash, nil, svc.LNClient, nil, nil)
	require.NoError(t, err)
	assert.True(t, transaction.Hold)
	// use the pubkey from the decoded tests.MockLNClientHoldTransaction invoice
	svc.LNClient.(*tests.MockLn).Pubkey = "038a73de75fdc3c7ec951a5e0b0fa95c5cd353bd7ca72df2086aa228c1f92819a5"

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		result, err := transactionsService.SendPaymentSync(ctx, transaction.PaymentRequest, nil, nil, svc.LNClient, nil, nil)
		assert.ErrorIs(t, err, lnclient.NewHoldInvoiceCanceledError())
		assert.Nil(t, result)

		outgoingTransactionType := constants.TRANSACTION_TYPE_OUTGOING
		failedOutgoingTransaction, err := transactionsService.LookupTransaction(ctx, paymentHash, &outgoingTransactionType, svc.LNClient, nil)
		assert.Nil(t, err)
		require.NotNil(t, failedOutgoingTransaction)
		assert.Equal(t, constants.TRANSACTION_STATE_FAILED, failedOutgoingTransaction.State)
		assert.Equal(t, true, failedOutgoingTransaction.SelfPayment)
		assert.Equal(t, false, failedOutgoingTransaction.Hold)
	}()

	// wait for payment to be accepted
	time.Sleep(10 * time.Millisecond)
	err = transactionsService.CancelHoldInvoice(ctx, paymentHash, svc.LNClient)
	assert.NoError(t, err)

	incomingTransactionType := constants.TRANSACTION_TYPE_INCOMING
	updatedHoldTransaction, err := transactionsService.LookupTransaction(ctx, paymentHash, &incomingTransactionType, svc.LNClient, nil)
	assert.Nil(t, err)
	assert.Equal(t, constants.TRANSACTION_STATE_FAILED, updatedHoldTransaction.State)
	assert.Equal(t, true, updatedHoldTransaction.SelfPayment)
	assert.Equal(t, true, updatedHoldTransaction.Hold)
	wg.Wait()
}
