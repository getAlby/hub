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

const nip47KeysendJson = `
{
	"method": "pay_keysend",
	"params": {
		"amount": 123000,
		"pubkey": "123pubkey2",
		"tlv_records": [{
			"type": 5482373484,
			"value": "fajsn341414fq"
		}]
	}
}
`

const nip47KeysendJsonWithPreimage = `
{
	"method": "pay_keysend",
	"params": {
		"amount": 123000,
		"pubkey": "123pubkey2",
		"preimage": "018465013e2337234a7e5530a21c4a8cf70d84231f4a8ff0b1e2cce3cb2bd03b",
		"tlv_records": [{
			"type": 5482373484,
			"value": "fajsn341414fq"
		}]
	}
}
`

func TestHandlePayKeysendEvent(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

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
	err = json.Unmarshal([]byte(nip47KeysendJson), nip47Request)
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
		HandlePayKeysendEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse, nostr.Tags{})

	assert.Nil(t, publishedResponse.Error)
	assert.Equal(t, 64, len(publishedResponse.Result.(payResponse).Preimage))
	assert.Equal(t, uint64(1), publishedResponse.Result.(payResponse).FeesPaid)
}
func TestHandlePayKeysendEvent_WithPreimage(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

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
	err = json.Unmarshal([]byte(nip47KeysendJsonWithPreimage), nip47Request)
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
		HandlePayKeysendEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse, nostr.Tags{})

	assert.Nil(t, publishedResponse.Error)
	assert.Equal(t, "018465013e2337234a7e5530a21c4a8cf70d84231f4a8ff0b1e2cce3cb2bd03b", publishedResponse.Result.(payResponse).Preimage)
	assert.Equal(t, uint64(1), publishedResponse.Result.(payResponse).FeesPaid)
}
