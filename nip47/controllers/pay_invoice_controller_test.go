package controllers

import (
	"context"
	"encoding/json"
	"testing"

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

const nip47PayInvoiceJson = `
{
	"method": "pay_invoice",
	"params": {
		"invoice": "lntb1230n1pjypux0pp5xgxzcks5jtx06k784f9dndjh664wc08ucrganpqn52d0ftrh9n8sdqyw3jscqzpgxqyz5vqsp5rkx7cq252p3frx8ytjpzc55rkgyx2mfkzzraa272dqvr2j6leurs9qyyssqhutxa24r5hqxstchz5fxlslawprqjnarjujp5sm3xj7ex73s32sn54fthv2aqlhp76qmvrlvxppx9skd3r5ut5xutgrup8zuc6ay73gqmra29m",
		"metadata": {"a": 123}
	}
}
`
const nip47PayInvoice0AmountJson = `
{
	"method": "pay_invoice",
	"params": {
		"invoice": "lnbc1pn428a6pp5njn604kl6x7eruycwzpwapwe97uer8et084kru4r0hsql9ttpmusdp82pshjgr5dusyymrfde4jq4mpd3kx2apq24ek2uscqzpuxqr8pqsp50nvadnrqlvghx44ftl89gjvx9lvrfpy5sypt2zahmpehvex4lp7q9p4gqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqpqysgq395pa893uux2agv8jlqxsppyyx5e4rfdjn9dhynlzxhuq3sgwnv48akdqxut7s4fqrre8adnt8kmfmjaexdcevajqlxvmtjpp823cmspud7udk",
		"amount": 1234
	}
}
`

const nip47PayJsonNoInvoice = `
{
	"method": "pay_invoice",
	"params": {
		"something": "else"
	}
}
`

func TestHandlePayInvoiceEvent(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: constants.PAY_INVOICE_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47PayInvoiceJson), nip47Request)
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
		HandlePayInvoiceEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse, nostr.Tags{})

	assert.Equal(t, "123preimage", publishedResponse.Result.(payResponse).Preimage)

	transactionType := constants.TRANSACTION_TYPE_OUTGOING
	transaction, err := transactionsSvc.LookupTransaction(ctx, "320c2c5a1492ccfd5bc7aa4ad9b657d6aaec3cfcc0d1d98413a29af4ac772ccf", &transactionType, svc.LNClient, &app.ID)
	assert.NoError(t, err)

	type dummyMetadata struct {
		A int `json:"a"`
	}
	var decodedMetadata dummyMetadata
	err = json.Unmarshal(transaction.Metadata, &decodedMetadata)
	assert.NoError(t, err)
	assert.Equal(t, 123, decodedMetadata.A)
}

func TestHandlePayInvoiceEvent_0Amount(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: constants.PAY_INVOICE_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47PayInvoice0AmountJson), nip47Request)
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
		HandlePayInvoiceEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse, nostr.Tags{})

	assert.Equal(t, "123preimage", publishedResponse.Result.(payResponse).Preimage)

	transactionType := constants.TRANSACTION_TYPE_OUTGOING
	transaction, err := transactionsSvc.LookupTransaction(ctx, "9ca7a7d6dfd1bd91f0987082ee85d92fb9919f2b79eb61f2a37de00f956b0ef9", &transactionType, svc.LNClient, &app.ID)
	assert.NoError(t, err)
	// from the request amount
	assert.Equal(t, uint64(1234), transaction.AmountMsat)
}

func TestHandlePayInvoiceEvent_MalformedInvoice(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: constants.PAY_INVOICE_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47PayJsonNoInvoice), nip47Request)
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
		HandlePayInvoiceEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse, nostr.Tags{})

	assert.Nil(t, publishedResponse.Result)
	assert.Equal(t, constants.ERROR_INTERNAL, publishedResponse.Error.Code)
	assert.Equal(t, "Failed to decode bolt11 invoice: bolt11 too short", publishedResponse.Error.Message)
}
