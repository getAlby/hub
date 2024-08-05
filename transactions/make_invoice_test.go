package transactions

import (
	"context"
	"strings"
	"testing"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/tests"
	"github.com/stretchr/testify/assert"
)

func TestMakeInvoice_NoApp(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	metadata := strings.Repeat("a", constants.INVOICE_METADATA_MAX_LENGTH-2) // json encoding adds 2 characters

	transactionsService := NewTransactionsService(svc.DB)
	transaction, err := transactionsService.MakeInvoice(ctx, 1234, "Hello world", "", 0, metadata, svc.LNClient, nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, uint64(tests.MockLNClientTransaction.Amount), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_PENDING, transaction.State)
	assert.Equal(t, tests.MockLNClientTransaction.Preimage, *transaction.Preimage)
	assert.Equal(t, `"`+metadata+`"`, transaction.Metadata) // json-encoded
}

func TestMakeInvoice_MetadataTooLarge(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)
	metadata := strings.Repeat("a", constants.INVOICE_METADATA_MAX_LENGTH-1) // json encoding adds 2 characters

	transactionsService := NewTransactionsService(svc.DB)
	transaction, err := transactionsService.MakeInvoice(ctx, 1234, "Hello world", "", 0, metadata, svc.LNClient, nil, nil)

	assert.Error(t, err)
	assert.Equal(t, "encoded invoice metadata provided is too large. Limit: 2048 Received: 2049", err.Error())
	assert.Nil(t, transaction)
}

func TestMakeInvoice_App(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	transactionsService := NewTransactionsService(svc.DB)
	transaction, err := transactionsService.MakeInvoice(ctx, 1234, "Hello world", "", 0, nil, svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.NoError(t, err)
	assert.Equal(t, uint64(tests.MockLNClientTransaction.Amount), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_PENDING, transaction.State)
	assert.Equal(t, tests.MockLNClientTransaction.Preimage, *transaction.Preimage)
	assert.Equal(t, app.ID, *transaction.AppId)
	assert.Equal(t, dbRequestEvent.ID, *transaction.RequestEventId)
}
