package transactions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/tests"
)

// TestReconcileMissedEvents_SettlesStuckPendingTransactions proves the
// end-to-end recovery for https://github.com/getAlby/hub/issues/1188:
// payments that settled on the LN node while the hub was offline stay pending
// because the notification streams never redeliver them and the unsettled
// polling fallback is disabled for backends with notification support.
// Replaying the missed terminal states as the same events the streams publish
// must converge every stuck row exactly once.
func TestReconcileMissedEvents_SettlesStuckPendingTransactions(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	sentHash := "reconcile-sent-hash"
	receivedHash := "reconcile-received-hash"
	failedHash := "reconcile-failed-hash"

	svc.DB.Create(&db.Transaction{
		State:       constants.TRANSACTION_STATE_PENDING,
		Type:        constants.TRANSACTION_TYPE_OUTGOING,
		PaymentHash: sentHash,
		AmountMsat:  123000,
	})
	svc.DB.Create(&db.Transaction{
		State:       constants.TRANSACTION_STATE_PENDING,
		Type:        constants.TRANSACTION_TYPE_INCOMING,
		PaymentHash: receivedHash,
		AmountMsat:  456000,
	})
	svc.DB.Create(&db.Transaction{
		State:       constants.TRANSACTION_STATE_PENDING,
		Type:        constants.TRANSACTION_TYPE_OUTGOING,
		PaymentHash: failedHash,
		AmountMsat:  789000,
	})

	mockEventConsumer := tests.NewMockEventConsumer()
	svc.EventPublisher.RegisterSubscriber(mockEventConsumer)
	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	// precondition: with notification support enabled (e.g. LND backend),
	// the unsettled polling fallback is skipped so the rows stay stuck.
	transactionsService.checkUnsettledTransactions(ctx, svc.LNClient)
	assertPendingWithHash(t, transactionsService, sentHash, constants.TRANSACTION_TYPE_OUTGOING)
	assertPendingWithHash(t, transactionsService, receivedHash, constants.TRANSACTION_TYPE_INCOMING)
	assertPendingWithHash(t, transactionsService, failedHash, constants.TRANSACTION_TYPE_OUTGOING)

	settledAt := tests.MockTimeUnix
	preimage := "reconcile-preimage"

	// replay the missed terminal states exactly as the startup reconciliation does.
	transactionsService.ConsumeEvent(ctx, &events.Event{
		Event: "nwc_lnclient_payment_sent",
		Properties: &lnclient.Transaction{
			Type:        "outgoing",
			PaymentHash: sentHash,
			AmountMsat:  123000,
			Preimage:    preimage,
			SettledAt:   &settledAt,
		},
	}, map[string]interface{}{})
	transactionsService.ConsumeEvent(ctx, &events.Event{
		Event: "nwc_lnclient_payment_received",
		Properties: &lnclient.Transaction{
			Type:        "incoming",
			PaymentHash: receivedHash,
			AmountMsat:  456000,
			Preimage:    preimage,
			SettledAt:   &settledAt,
		},
	}, map[string]interface{}{})
	transactionsService.ConsumeEvent(ctx, &events.Event{
		Event: "nwc_lnclient_payment_failed",
		Properties: &lnclient.PaymentFailedEventProperties{
			Transaction: &lnclient.Transaction{
				Type:        "outgoing",
				PaymentHash: failedHash,
				AmountMsat:  789000,
			},
			Reason: "FAILURE_REASON_NO_ROUTE",
		},
	}, map[string]interface{}{})

	assertStateWithHash(t, transactionsService, sentHash, constants.TRANSACTION_TYPE_OUTGOING, constants.TRANSACTION_STATE_SETTLED)
	assertStateWithHash(t, transactionsService, receivedHash, constants.TRANSACTION_TYPE_INCOMING, constants.TRANSACTION_STATE_SETTLED)
	assertStateWithHash(t, transactionsService, failedHash, constants.TRANSACTION_TYPE_OUTGOING, constants.TRANSACTION_STATE_FAILED)

	consumedEvents := mockEventConsumer.WaitForConsumedEvents(3)
	require.Len(t, consumedEvents, 3)
	consumedByEvent := map[string]int{}
	for _, consumedEvent := range consumedEvents {
		consumedByEvent[consumedEvent.Event]++
	}
	assert.Equal(t, map[string]int{
		"nwc_payment_sent":     1,
		"nwc_payment_received": 1,
		"nwc_payment_failed":   1,
	}, consumedByEvent)

	// duplicate redelivery (stream finally catching up after the reconcile
	// pass) must not transition or notify twice.
	transactionsService.ConsumeEvent(ctx, &events.Event{
		Event: "nwc_lnclient_payment_sent",
		Properties: &lnclient.Transaction{
			Type:        "outgoing",
			PaymentHash: sentHash,
			AmountMsat:  123000,
			Preimage:    preimage,
			SettledAt:   &settledAt,
		},
	}, map[string]interface{}{})
	transactionsService.ConsumeEvent(ctx, &events.Event{
		Event: "nwc_lnclient_payment_failed",
		Properties: &lnclient.PaymentFailedEventProperties{
			Transaction: &lnclient.Transaction{
				Type:        "outgoing",
				PaymentHash: failedHash,
				AmountMsat:  789000,
			},
			Reason: "FAILURE_REASON_NO_ROUTE",
		},
	}, map[string]interface{}{})

	assertStateWithHash(t, transactionsService, sentHash, constants.TRANSACTION_TYPE_OUTGOING, constants.TRANSACTION_STATE_SETTLED)
	assertStateWithHash(t, transactionsService, failedHash, constants.TRANSACTION_TYPE_OUTGOING, constants.TRANSACTION_STATE_FAILED)
	assert.Len(t, mockEventConsumer.GetConsumedEvents(), 3)
}

func assertPendingWithHash(t *testing.T, transactionsService *transactionsService, paymentHash string, transactionType string) {
	t.Helper()
	assertStateWithHash(t, transactionsService, paymentHash, transactionType, constants.TRANSACTION_STATE_PENDING)
}

func assertStateWithHash(t *testing.T, transactionsService *transactionsService, paymentHash string, transactionType string, state string) {
	t.Helper()
	var transaction db.Transaction
	result := transactionsService.db.Limit(1).Find(&transaction, &db.Transaction{
		Type:        transactionType,
		PaymentHash: paymentHash,
	})
	require.Equal(t, int64(1), result.RowsAffected)
	assert.Equal(t, state, transaction.State)
}
