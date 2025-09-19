package transactions

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/tests"
)

func TestNotifications_ReceivedKnownPayment(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	transactionsService.ConsumeEvent(ctx, &events.Event{
		Event:      "nwc_lnclient_payment_received",
		Properties: tests.MockLNClientTransaction,
	}, map[string]interface{}{})

	incomingTransaction, err := transactionsService.LookupTransaction(ctx, tests.MockLNClientTransaction.PaymentHash, nil, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), incomingTransaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, incomingTransaction.State)
	assert.Equal(t, tests.MockLNClientTransaction.Preimage, *incomingTransaction.Preimage)
	assert.Zero(t, incomingTransaction.FeeReserveMsat)

	transactions := []db.Transaction{}
	result := svc.DB.Find(&transactions)
	assert.Equal(t, int64(1), result.RowsAffected)
}

func TestNotifications_ReceivedUnknownPayment(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	transactionsService.ConsumeEvent(ctx, &events.Event{
		Event:      "nwc_lnclient_payment_received",
		Properties: tests.MockLNClientTransaction,
	}, map[string]interface{}{})

	transactionType := constants.TRANSACTION_TYPE_INCOMING
	incomingTransaction, err := transactionsService.LookupTransaction(ctx, tests.MockLNClientTransaction.PaymentHash, &transactionType, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(tests.MockLNClientTransaction.Amount), incomingTransaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, incomingTransaction.State)
	assert.Equal(t, tests.MockLNClientTransaction.Preimage, *incomingTransaction.Preimage)
	assert.Zero(t, incomingTransaction.FeeReserveMsat)

	transactions := []db.Transaction{}
	result := svc.DB.Find(&transactions)
	assert.Equal(t, int64(1), result.RowsAffected)
}

func TestNotifications_ReceivedKeysend(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	metadata := map[string]interface{}{}

	metadata["tlv_records"] = []lnclient.TLVRecord{
		{
			Type:  7629169,
			Value: "7b22616374696f6e223a22626f6f7374222c2276616c75655f6d736174223a313030302c2276616c75655f6d7361745f746f74616c223a313030302c226170705f6e616d65223a22e29aa1205765624c4e2044656d6f222c226170705f76657273696f6e223a22312e30222c22666565644944223a2268747470733a2f2f66656564732e706f6463617374696e6465782e6f72672f706332302e786d6c222c22706f6463617374223a22506f6463617374696e6720322e30222c22657069736f6465223a22457069736f6465203130343a2041204e65772044756d70222c227473223a32312c226e616d65223a22e29aa1205765624c4e2044656d6f222c2273656e6465725f6e616d65223a225361746f736869204e616b616d6f746f222c226d657373616765223a22476f20706f6463617374696e6721227d",
		},
	}

	transaction := &lnclient.Transaction{
		Type:            "incoming",
		Invoice:         tests.MockInvoice,
		Description:     "",
		DescriptionHash: "",
		Preimage:        tests.MockLNClientTransaction.Preimage,
		PaymentHash:     tests.MockLNClientTransaction.PaymentHash,
		Amount:          2000,
		FeesPaid:        75,
		SettledAt:       &tests.MockTimeUnix,
		Metadata:        metadata,
	}

	transactionsService.ConsumeEvent(ctx, &events.Event{
		Event:      "nwc_lnclient_payment_received",
		Properties: transaction,
	}, map[string]interface{}{})

	transactionType := constants.TRANSACTION_TYPE_INCOMING
	incomingTransaction, err := transactionsService.LookupTransaction(ctx, tests.MockLNClientTransaction.PaymentHash, &transactionType, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(2000), incomingTransaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, incomingTransaction.State)
	assert.Equal(t, tests.MockLNClientTransaction.Preimage, *incomingTransaction.Preimage)
	assert.Zero(t, incomingTransaction.FeeReserveMsat)

	var boostagram Boostagram
	err = json.Unmarshal(incomingTransaction.Boostagram, &boostagram)
	assert.NoError(t, err)

	assert.Equal(t, "⚡ WebLN Demo", boostagram.AppName)
	assert.Equal(t, "⚡ WebLN Demo", boostagram.Name)
	assert.Equal(t, "Podcasting 2.0", boostagram.Podcast)
	assert.Equal(t, "Episode 104: A New Dump", boostagram.Episode.String())
	assert.Equal(t, "https://feeds.podcastindex.org/pc20.xml", boostagram.FeedId.String())
	assert.Equal(t, int64(21), boostagram.Timestamp)
	assert.Equal(t, "Go podcasting!", boostagram.Message)
	assert.Equal(t, "Satoshi Nakamoto", boostagram.SenderName)
	assert.Equal(t, "boost", boostagram.Action)
	assert.Equal(t, int64(1000), boostagram.ValueMsatTotal)

	assert.Equal(t, "Go podcasting!", incomingTransaction.Description)

	transactions := []db.Transaction{}
	result := svc.DB.Find(&transactions)
	assert.Equal(t, int64(1), result.RowsAffected)
}

func TestNotifications_SentKnownPayment(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		AmountMsat:     123000,
		FeeReserveMsat: uint64(10000),
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	transactionsService.ConsumeEvent(ctx, &events.Event{
		Event:      "nwc_lnclient_payment_sent",
		Properties: tests.MockLNClientTransaction,
	}, map[string]interface{}{})

	transactionType := constants.TRANSACTION_TYPE_OUTGOING
	outgoingTransaction, err := transactionsService.LookupTransaction(ctx, tests.MockLNClientTransaction.PaymentHash, &transactionType, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), outgoingTransaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, outgoingTransaction.State)
	assert.Equal(t, tests.MockLNClientTransaction.Preimage, *outgoingTransaction.Preimage)
	assert.Zero(t, outgoingTransaction.FeeReserveMsat)

	transactions := []db.Transaction{}
	result := svc.DB.Find(&transactions)
	assert.Equal(t, int64(1), result.RowsAffected)
}

func TestNotifications_SentUnknownPayment(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	transactions := []db.Transaction{}
	result := svc.DB.Find(&transactions)
	assert.Equal(t, int64(0), result.RowsAffected)

	transactionsService.ConsumeEvent(ctx, &events.Event{
		Event:      "nwc_lnclient_payment_sent",
		Properties: tests.MockLNClientTransaction,
	}, map[string]interface{}{})

	transactionType := constants.TRANSACTION_TYPE_OUTGOING
	outgoingTransaction, err := transactionsService.LookupTransaction(ctx, tests.MockLNClientTransaction.PaymentHash, &transactionType, svc.LNClient, nil)
	assert.Nil(t, outgoingTransaction)
	assert.ErrorIs(t, err, NewNotFoundError())
}

func TestNotifications_FailedKnownPendingPayment(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		AmountMsat:     123000,
		FeeReserveMsat: uint64(10000),
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	transactionsService.ConsumeEvent(ctx, &events.Event{
		Event: "nwc_lnclient_payment_failed",
		Properties: &lnclient.PaymentFailedEventProperties{
			Transaction: tests.MockLNClientTransaction,
			Reason:      "Some failure reason",
		},
	}, map[string]interface{}{})

	transactionType := constants.TRANSACTION_TYPE_OUTGOING
	outgoingTransaction, err := transactionsService.LookupTransaction(ctx, tests.MockLNClientTransaction.PaymentHash, &transactionType, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, constants.TRANSACTION_STATE_FAILED, outgoingTransaction.State)
	assert.Nil(t, outgoingTransaction.Preimage)
	assert.Zero(t, outgoingTransaction.FeeReserveMsat)

	transactions := []db.Transaction{}
	result := svc.DB.Find(&transactions)
	assert.Equal(t, int64(1), result.RowsAffected)
}

func TestNotifications_FailedKnownPendingAndExistingFailedPayment(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	// in this test, a user tries to pay again, and the payment fails again.
	// The second (pending) payment should be marked as failed.

	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_FAILED,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		AmountMsat:     123000,
		FeeReserveMsat: uint64(10000),
	})

	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		AmountMsat:     123000,
		FeeReserveMsat: uint64(10000),
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	transactionsService.ConsumeEvent(ctx, &events.Event{
		Event: "nwc_lnclient_payment_failed",
		Properties: &lnclient.PaymentFailedEventProperties{
			Transaction: tests.MockLNClientTransaction,
			Reason:      "Some failure reason",
		},
	}, map[string]interface{}{})

	transactions, totalCount, err := transactionsService.ListTransactions(ctx, uint64(0), uint64(0), uint64(0), uint64(0), true, false, nil, svc.LNClient, nil, false)
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), totalCount)
	for _, transaction := range transactions {
		assert.Equal(t, constants.TRANSACTION_STATE_FAILED, transaction.State)
	}
}

func TestNotifications_SentAfterMarkedPaymentFailed(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	// in this test, we marked a payment as failed (for whatever reason) but then later received a payment successful event.
	// The second (pending) payment should be marked as failed.

	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_FAILED,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		AmountMsat:     123000,
		FeeReserveMsat: uint64(10000),
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	transactionsService.ConsumeEvent(ctx, &events.Event{
		Event:      "nwc_lnclient_payment_sent",
		Properties: tests.MockLNClientTransaction,
	}, map[string]interface{}{})

	transactions, totalCount, err := transactionsService.ListTransactions(ctx, uint64(0), uint64(0), uint64(0), uint64(0), true, false, nil, svc.LNClient, nil, false)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), totalCount)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transactions[0].State)
}

func TestNotifications_SentAfterMarkedTwoPaymentsFailed(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	// in this test, we failed to pay twice but then later received a payment successful event.
	// The second (latest) failed payment should be marked as settled.

	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_FAILED,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		AmountMsat:     123000,
		FeeReserveMsat: uint64(10000),
		UpdatedAt:      time.Now().Add(-1 * time.Second),
	})

	latestFailedTransaction := &db.Transaction{
		State:          constants.TRANSACTION_STATE_FAILED,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		AmountMsat:     123000,
		FeeReserveMsat: uint64(10000),
		UpdatedAt:      time.Now(),
	}
	svc.DB.Create(latestFailedTransaction)

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	transactionsService.ConsumeEvent(ctx, &events.Event{
		Event:      "nwc_lnclient_payment_sent",
		Properties: tests.MockLNClientTransaction,
	}, map[string]interface{}{})

	transactions, totalCount, err := transactionsService.ListTransactions(ctx, uint64(0), uint64(0), uint64(0), uint64(0), true, false, nil, svc.LNClient, nil, false)
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), totalCount)
	assert.Equal(t, latestFailedTransaction.ID, transactions[0].ID)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transactions[0].State)
	assert.Equal(t, constants.TRANSACTION_STATE_FAILED, transactions[1].State)
}

func TestNotifications_SentWithFailedAndPendingPayment(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	// in this test, we failed to pay once and retried (second attempt pending) then later received a payment successful event.
	// The pending payment should be marked as settled.

	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_FAILED,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		AmountMsat:     123000,
		FeeReserveMsat: uint64(10000),
	})

	pendingTransaction := &db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		AmountMsat:     123000,
		FeeReserveMsat: uint64(10000),
	}
	svc.DB.Create(pendingTransaction)

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	transactionsService.ConsumeEvent(ctx, &events.Event{
		Event:      "nwc_lnclient_payment_sent",
		Properties: tests.MockLNClientTransaction,
	}, map[string]interface{}{})

	transactions, totalCount, err := transactionsService.ListTransactions(ctx, uint64(0), uint64(0), uint64(0), uint64(0), true, false, nil, svc.LNClient, nil, false)
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), totalCount)
	assert.Equal(t, pendingTransaction.ID, transactions[0].ID)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transactions[0].State)
	assert.Equal(t, constants.TRANSACTION_STATE_FAILED, transactions[1].State)
}
