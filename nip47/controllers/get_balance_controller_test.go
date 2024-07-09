package controllers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/nbd-wtf/go-nostr"
	"github.com/stretchr/testify/assert"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/nip47/permissions"
	"github.com/getAlby/hub/tests"
	"github.com/getAlby/hub/transactions"
)

const nip47GetBalanceJson = `
{
	"method": "get_balance"
}
`

func TestHandleGetBalanceEvent_NoPermission(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47GetBalanceJson), nip47Request)
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
		HandleGetBalanceEvent(ctx, nip47Request, dbRequestEvent.ID, *dbRequestEvent.AppId, checkPermission, publishResponse)

	assert.Nil(t, publishedResponse.Result)
	assert.Equal(t, models.ERROR_RESTRICTED, publishedResponse.Error.Code)
}

func TestHandleGetBalanceEvent_WithPermission(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47GetBalanceJson), nip47Request)
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
		HandleGetBalanceEvent(ctx, nip47Request, dbRequestEvent.ID, *dbRequestEvent.AppId, checkPermission, publishResponse)

	assert.Equal(t, int64(21000), publishedResponse.Result.(*getBalanceResponse).Balance)
	assert.Nil(t, publishedResponse.Error)
}

// create pay_invoice permission
// maxAmount := 1000
// budgetRenewal := "never"
// appPermission = &db.AppPermission{
// 	AppId:         app.ID,
// 	App:           *app,
// 	RequestMethod: models.PAY_INVOICE_METHOD,
// 	MaxAmount:     maxAmount,
// 	BudgetRenewal: budgetRenewal,
// 	ExpiresAt:     &expiresAt,
// }
// err = svc.DB.Create(appPermission).Error
// assert.NoError(t, err)

// reqEvent.ID = "test_get_balance_with_budget"
// responses = []*models.Response{}
// svc.nip47Svc.HandleGetBalanceEvent(ctx, nip47Request, dbRequestEvent, app, publishResponse)

// assert.Equal(t, int64(21000), responses[0].Result.(*getBalanceResponse).Balance)
// assert.Equal(t, 1000000, responses[0].Result.(*getBalanceResponse).MaxAmount)
// assert.Equal(t, "never", responses[0].Result.(*getBalanceResponse).BudgetRenewal)
