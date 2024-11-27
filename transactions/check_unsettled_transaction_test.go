package transactions

import (
	"context"
	"testing"
	"time"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckUnsettledTransaction(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

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
	settledAt := time.Now().Unix()
	svc.LNClient.(*tests.MockLn).MockTransaction = &lnclient.Transaction{
		SettledAt: &settledAt,
		Preimage:  "dummy",
	}

	// do not allow checking unsettled transactions if notifications are supported
	transactionsService.checkUnsettledTransaction(context.TODO(), &dbTransaction, svc.LNClient)
	assert.Equal(t, constants.TRANSACTION_STATE_PENDING, dbTransaction.State)

	svc.LNClient.(*tests.MockLn).SupportedNotificationTypes = &[]string{}
	transactionsService.checkUnsettledTransaction(context.TODO(), &dbTransaction, svc.LNClient)

	assert.NoError(t, err)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, dbTransaction.State)
	assert.Equal(t, 1, len(mockEventConsumer.GetConsumedEvents()))
	assert.Equal(t, "nwc_payment_sent", mockEventConsumer.GetConsumedEvents()[0].Event)
	settledTransaction := mockEventConsumer.GetConsumedEvents()[0].Properties.(*db.Transaction)
	assert.Equal(t, &dbTransaction, settledTransaction)
}

func TestCheckUnsettledTransactions(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	dbTransaction := db.Transaction{
		State:       constants.TRANSACTION_STATE_PENDING,
		Type:        constants.TRANSACTION_TYPE_OUTGOING,
		PaymentHash: tests.MockLNClientTransaction.PaymentHash,
		AmountMsat:  123000,
		CreatedAt:   time.Now(),
	}
	svc.DB.Create(&dbTransaction)

	mockEventConsumer := tests.NewMockEventConsumer()
	svc.EventPublisher.RegisterSubscriber(mockEventConsumer)
	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	settledAt := time.Now().Unix()

	svc.LNClient.(*tests.MockLn).MockTransaction = &lnclient.Transaction{
		SettledAt: &settledAt,
		Preimage:  "dummy",
	}

	// do not allow checking unsettled transactions if notifications are supported
	transactionsService.checkUnsettledTransactions(context.TODO(), svc.LNClient)

	svc.DB.Find(&dbTransaction, db.Transaction{
		ID: dbTransaction.ID,
	})
	assert.Equal(t, constants.TRANSACTION_STATE_PENDING, dbTransaction.State)

	svc.LNClient.(*tests.MockLn).SupportedNotificationTypes = &[]string{}
	transactionsService.checkUnsettledTransactions(context.TODO(), svc.LNClient)

	svc.DB.Find(&dbTransaction, db.Transaction{
		ID: dbTransaction.ID,
	})
	assert.NoError(t, err)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, dbTransaction.State)
	assert.Equal(t, 1, len(mockEventConsumer.GetConsumedEvents()))
	assert.Equal(t, "nwc_payment_sent", mockEventConsumer.GetConsumedEvents()[0].Event)
	settledTransaction := mockEventConsumer.GetConsumedEvents()[0].Properties.(*db.Transaction)
	assert.Equal(t, dbTransaction.ID, settledTransaction.ID)
}
