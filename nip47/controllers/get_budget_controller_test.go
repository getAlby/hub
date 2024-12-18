package controllers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/nip47/permissions"
	"github.com/getAlby/hub/tests"
	"github.com/getAlby/hub/transactions"
)

const nip47GetBudgetJson = `
{
	"method": "get_budget"
}
`

func TestHandleGetBudgetEvent_NoRenewal(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47GetBudgetJson), nip47Request)
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId:         app.ID,
		App:           *app,
		Scope:         constants.PAY_INVOICE_SCOPE,
		MaxAmountSat:  400,
		BudgetRenewal: constants.BUDGET_RENEWAL_NEVER,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB, svc.EventPublisher)
	NewNip47Controller(svc.LNClient, svc.DB, svc.EventPublisher, permissionsSvc, transactionsSvc, svc.AppsService).
		HandleGetBudgetEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	assert.Equal(t, uint64(400000), publishedResponse.Result.(*getBudgetResponse).TotalBudget)
	assert.Equal(t, uint64(0), publishedResponse.Result.(*getBudgetResponse).UsedBudget)
	assert.Nil(t, publishedResponse.Result.(*getBudgetResponse).RenewsAt)
	assert.Equal(t, constants.BUDGET_RENEWAL_NEVER, publishedResponse.Result.(*getBudgetResponse).RenewalPeriod)
	assert.Nil(t, publishedResponse.Error)
}

func TestHandleGetBudgetEvent_NoneUsed(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47GetBudgetJson), nip47Request)
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)
	now := time.Now()

	appPermission := &db.AppPermission{
		AppId:         app.ID,
		App:           *app,
		Scope:         constants.PAY_INVOICE_SCOPE,
		MaxAmountSat:  400,
		BudgetRenewal: constants.BUDGET_RENEWAL_MONTHLY,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB, svc.EventPublisher)
	NewNip47Controller(svc.LNClient, svc.DB, svc.EventPublisher, permissionsSvc, transactionsSvc, svc.AppsService).
		HandleGetBudgetEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	assert.Equal(t, uint64(400000), publishedResponse.Result.(*getBudgetResponse).TotalBudget)
	assert.Equal(t, uint64(0), publishedResponse.Result.(*getBudgetResponse).UsedBudget)
	renewsAt := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, 1, 0).Unix()
	assert.Equal(t, uint64(renewsAt), *publishedResponse.Result.(*getBudgetResponse).RenewsAt)
	assert.Equal(t, constants.BUDGET_RENEWAL_MONTHLY, publishedResponse.Result.(*getBudgetResponse).RenewalPeriod)
	assert.Nil(t, publishedResponse.Error)
}

func TestHandleGetBudgetEvent_HalfUsed(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47GetBudgetJson), nip47Request)
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)
	now := time.Now()

	appPermission := &db.AppPermission{
		AppId:         app.ID,
		App:           *app,
		Scope:         constants.PAY_INVOICE_SCOPE,
		MaxAmountSat:  400,
		BudgetRenewal: constants.BUDGET_RENEWAL_MONTHLY,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	svc.DB.Create(&db.Transaction{
		AppId:      &app.ID,
		State:      constants.TRANSACTION_STATE_SETTLED,
		Type:       constants.TRANSACTION_TYPE_OUTGOING,
		AmountMsat: 200000,
	})

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB, svc.EventPublisher)
	NewNip47Controller(svc.LNClient, svc.DB, svc.EventPublisher, permissionsSvc, transactionsSvc, svc.AppsService).
		HandleGetBudgetEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	assert.Equal(t, uint64(400000), publishedResponse.Result.(*getBudgetResponse).TotalBudget)
	assert.Equal(t, uint64(200000), publishedResponse.Result.(*getBudgetResponse).UsedBudget)
	renewsAt := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, 1, 0).Unix()
	assert.Equal(t, uint64(renewsAt), *publishedResponse.Result.(*getBudgetResponse).RenewsAt)
	assert.Equal(t, constants.BUDGET_RENEWAL_MONTHLY, publishedResponse.Result.(*getBudgetResponse).RenewalPeriod)
	assert.Nil(t, publishedResponse.Error)
}

func TestHandleGetBudgetEvent_NoBudget(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47GetBudgetJson), nip47Request)
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: constants.PAY_INVOICE_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	svc.DB.Create(&db.Transaction{
		AppId:      &app.ID,
		State:      constants.TRANSACTION_STATE_SETTLED,
		Type:       constants.TRANSACTION_TYPE_OUTGOING,
		AmountMsat: 200000,
	})

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB, svc.EventPublisher)
	NewNip47Controller(svc.LNClient, svc.DB, svc.EventPublisher, permissionsSvc, transactionsSvc, svc.AppsService).
		HandleGetBudgetEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	assert.Equal(t, struct{}{}, publishedResponse.Result)
	assert.Nil(t, publishedResponse.Error)
}

func TestHandleGetBudgetEvent_NoPayInvoicePermission(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47GetBudgetJson), nip47Request)
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB, svc.EventPublisher)
	NewNip47Controller(svc.LNClient, svc.DB, svc.EventPublisher, permissionsSvc, transactionsSvc, svc.AppsService).
		HandleGetBudgetEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	assert.Equal(t, struct{}{}, publishedResponse.Result)
	assert.Nil(t, publishedResponse.Error)
}
