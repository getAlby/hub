package transactions

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/tests"
)

func TestMarkHoldInvoiceAcceptedWithoutPaymentRequest(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	paymentHash := "unique-hold-payment-hash"
	transaction := db.Transaction{
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		State:          constants.TRANSACTION_STATE_PENDING,
		PaymentRequest: "lnbc-unique-hold-invoice",
		PaymentHash:    paymentHash,
		Hold:           true,
	}
	require.NoError(t, svc.DB.Create(&transaction).Error)

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	settleDeadline := uint32(840_000)
	transactionsService.markHoldInvoiceAccepted("", paymentHash, settleDeadline, false)

	var updated db.Transaction
	require.NoError(t, svc.DB.First(&updated, transaction.ID).Error)
	require.Equal(t, constants.TRANSACTION_STATE_ACCEPTED, updated.State)
	require.Equal(t, &settleDeadline, updated.SettleDeadline)
}

func TestMarkHoldInvoiceAcceptedWithoutPaymentRequestRejectsAmbiguousHash(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	paymentHash := "ambiguous-hold-payment-hash"
	for _, invoice := range []string{"lnbc-first-hold-invoice", "lnbc-second-hold-invoice"} {
		require.NoError(t, svc.DB.Create(&db.Transaction{
			Type:           constants.TRANSACTION_TYPE_INCOMING,
			State:          constants.TRANSACTION_STATE_PENDING,
			PaymentRequest: invoice,
			PaymentHash:    paymentHash,
			Hold:           true,
		}).Error)
	}

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transactionsService.markHoldInvoiceAccepted("", paymentHash, 840_000, false)

	var acceptedCount int64
	require.NoError(t, svc.DB.Model(&db.Transaction{}).
		Where("payment_hash = ? AND state = ?", paymentHash, constants.TRANSACTION_STATE_ACCEPTED).
		Count(&acceptedCount).Error)
	require.Zero(t, acceptedCount)
}
