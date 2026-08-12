package controllers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/getAlby/go-nostr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/tests"
)

const nip47ReceiveJson = `
{
	"method": "receive",
	"params": {
		"amount": 123000,
		"description": "test receive"
	}
}
`

const nip47ReceiveNoAmountJson = `
{
	"method": "receive",
	"params": {
		"description": "test receive"
	}
}
`

func setupReceiveTest(t *testing.T) (*tests.TestService, *db.App, *db.RequestEvent) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	t.Cleanup(svc.Remove)

	app, _, err := tests.CreateApp(svc)
	require.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: constants.MAKE_INVOICE_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
	require.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	require.NoError(t, err)

	return svc, app, dbRequestEvent
}

func TestHandleReceiveEvent(t *testing.T) {
	ctx := context.TODO()
	svc, app, dbRequestEvent := setupReceiveTest(t)

	nip47Request := &models.Request{}
	err := json.Unmarshal([]byte(nip47ReceiveJson), nip47Request)
	require.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandleReceiveEvent(ctx, nip47Request, dbRequestEvent.ID, app.ID, publishResponse)

	require.Nil(t, publishedResponse.Error)
	result := publishedResponse.Result.(receiveResult)
	assert.Equal(t, "bitcoin:?lightning="+tests.MockInvoice, result.Bip321)
	assert.Equal(t, tests.MockPaymentHash, result.TransactionId)
}

func TestHandleReceiveEvent_NoAmount(t *testing.T) {
	ctx := context.TODO()
	svc, app, dbRequestEvent := setupReceiveTest(t)

	nip47Request := &models.Request{}
	err := json.Unmarshal([]byte(nip47ReceiveNoAmountJson), nip47Request)
	require.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandleReceiveEvent(ctx, nip47Request, dbRequestEvent.ID, app.ID, publishResponse)

	assert.Nil(t, publishedResponse.Result)
	assert.Equal(t, constants.ERROR_BAD_REQUEST, publishedResponse.Error.Code)
}
