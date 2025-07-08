package transactions

import (
	"context"
	"encoding/hex"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/tests"
)

func TestReceiveKeysendWithCustomKey(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	// just to have another app in this test
	_, _, err = tests.CreateApp(svc)

	assert.NoError(t, err)

	tlv := []lnclient.TLVRecord{
		{
			Type:  696969,
			Value: hex.EncodeToString([]byte(strconv.FormatUint(uint64(app.ID), 10))),
		},
	}
	tx := lnclient.Transaction{
		Type:        "incoming",
		Description: "mock invoice 1",
		Preimage:    "9f59b18f80a77c2930deb8be5ff1143eacdd1891c63c23d61bc9f99c64e57325",
		PaymentHash: "ae4277b7be3ca1420cafd24c143866190f52b996856b0e4164763f936e61ea1b",
		Amount:      1000,
		FeesPaid:    50,
		SettledAt:   &tests.MockTimeUnix,
		Metadata: map[string]interface{}{
			"tlv_records": tlv,
		},
	}

	event := events.Event{
		Event:      "nwc_lnclient_payment_received",
		Properties: &tx,
	}
	transactionsService.ConsumeEvent(ctx, &event, map[string]interface{}{})

	transaction, err := transactionsService.LookupTransaction(ctx, tx.PaymentHash, nil, svc.LNClient, nil)
	require.NoError(t, err)
	assert.Equal(t, app.ID, *transaction.AppId)
	assert.Equal(t, uint(1), app.ID)
}

func TestReceiveKeysend(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	_, _, err = tests.CreateApp(svc)
	assert.NoError(t, err)

	tx := tests.MockLNClientTransaction
	event := events.Event{
		Event:      "nwc_lnclient_payment_received",
		Properties: tx,
	}
	transactionsService.ConsumeEvent(ctx, &event, map[string]interface{}{})

	transaction, err := transactionsService.LookupTransaction(ctx, tx.PaymentHash, nil, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Nil(t, transaction.AppId)
}
