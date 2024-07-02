package controllers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/nbd-wtf/go-nostr"
	"github.com/stretchr/testify/assert"

	"github.com/getAlby/nostr-wallet-connect/db"
	"github.com/getAlby/nostr-wallet-connect/nip47/models"
	permissions "github.com/getAlby/nostr-wallet-connect/nip47/permissions"
	"github.com/getAlby/nostr-wallet-connect/tests"
	"github.com/getAlby/nostr-wallet-connect/transactions"
)

const nip47ListTransactionsJson = `
{
	"method": "list_transactions",
	"params": {
		"from": 1693876973,
		"until": 1694876973,
		"limit": 10,
		"offset": 0,
		"type": "incoming"
	}
}
`

func TestHandleListTransactionsEvent_NoPermission(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47ListTransactionsJson), nip47Request)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	checkPermission := func(amountMsat uint64) *models.Response {
		return &models.Response{
			ResultType: nip47Request.Method,
			Error: &models.Error{
				Code: models.ERROR_RESTRICTED,
			},
		}
	}

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB)
	NewNip47Controller(svc.LNClient, svc.DB, svc.EventPublisher, permissionsSvc, transactionsSvc).
		HandleListTransactionsEvent(ctx, nip47Request, dbRequestEvent.ID, checkPermission, publishResponse)

	assert.Nil(t, publishedResponse.Result)
	assert.Equal(t, models.ERROR_RESTRICTED, publishedResponse.Error.Code)
}

func TestHandleListTransactionsEvent_WithPermission(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47ListTransactionsJson), nip47Request)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	checkPermission := func(amountMsat uint64) *models.Response {
		return nil
	}

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB)
	NewNip47Controller(svc.LNClient, svc.DB, svc.EventPublisher, permissionsSvc, transactionsSvc).
		HandleListTransactionsEvent(ctx, nip47Request, dbRequestEvent.ID, checkPermission, publishResponse)

	assert.Nil(t, publishedResponse.Error)

	assert.Equal(t, 2, len(publishedResponse.Result.(*listTransactionsResponse).Transactions))
	transaction := publishedResponse.Result.(*listTransactionsResponse).Transactions[0]
	assert.Equal(t, tests.MockTransactions[0].Type, transaction.Type)
	assert.Equal(t, tests.MockTransactions[0].Invoice, transaction.Invoice)
	assert.Equal(t, tests.MockTransactions[0].Description, transaction.Description)
	assert.Equal(t, tests.MockTransactions[0].DescriptionHash, transaction.DescriptionHash)
	assert.Equal(t, tests.MockTransactions[0].Preimage, transaction.Preimage)
	assert.Equal(t, tests.MockTransactions[0].PaymentHash, transaction.PaymentHash)
	assert.Equal(t, tests.MockTransactions[0].Amount, transaction.Amount)
	assert.Equal(t, tests.MockTransactions[0].FeesPaid, transaction.FeesPaid)
	assert.Equal(t, tests.MockTransactions[0].SettledAt, transaction.SettledAt)
}
