package transactions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/tests"
)

func TestMakeInvoice_NoApp(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	txMetadata := make(map[string]interface{})
	txMetadata["randomkey"] = strings.Repeat("a", constants.INVOICE_METADATA_MAX_LENGTH-16) // json encoding adds 16 characters - {"randomkey":""}

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.MakeInvoice(ctx, 1234, "Hello world", "", 0, txMetadata, svc.LNClient, nil, nil, nil)
	assert.NoError(t, err)

	var metadata map[string]interface{}
	err = json.Unmarshal(transaction.Metadata, &metadata)
	assert.NoError(t, err)

	assert.Equal(t, uint64(tests.MockLNClientTransaction.Amount), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_PENDING, transaction.State)
	assert.Equal(t, tests.MockLNClientTransaction.Preimage, *transaction.Preimage)
	assert.Equal(t, txMetadata["randomkey"], metadata["randomkey"])
}

func TestMakeInvoice_MetadataTooLarge(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	metadata := make(map[string]interface{})
	metadata["randomkey"] = strings.Repeat("a", constants.INVOICE_METADATA_MAX_LENGTH-15) // json encoding adds 16 characters

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.MakeInvoice(ctx, 1234, "Hello world", "", 0, metadata, svc.LNClient, nil, nil, nil)

	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf("encoded invoice metadata provided is too large. Limit: %d Received: %d", constants.INVOICE_METADATA_MAX_LENGTH, constants.INVOICE_METADATA_MAX_LENGTH+1), err.Error())
	assert.Nil(t, transaction)
}

func TestMakeInvoice_App(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.MakeInvoice(ctx, 1234, "Hello world", "", 0, nil, svc.LNClient, &app.ID, &dbRequestEvent.ID, nil)

	assert.NoError(t, err)
	assert.Equal(t, uint64(tests.MockLNClientTransaction.Amount), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_PENDING, transaction.State)
	assert.Equal(t, tests.MockLNClientTransaction.Preimage, *transaction.Preimage)
	assert.Equal(t, app.ID, *transaction.AppId)
	assert.Equal(t, dbRequestEvent.ID, *transaction.RequestEventId)
}
