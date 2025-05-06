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
	"github.com/getAlby/hub/tests"
)

const nip47CancelHoldInvoiceJson = `
{
	"method": "cancel_hold_invoice",
	"params": {
		"payment_hash": "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	}
}
`

func TestHandleCancelHoldInvoiceEvent(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47CancelHoldInvoiceJson), nip47Request)
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	// Grant the hold_invoice scope to the app
	appPermission := &db.AppPermission{
		AppId: app.ID,
		Scope: constants.HOLD_INVOICES_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{
		AppId: &app.ID,
	}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandleCancelHoldInvoiceEvent(ctx, nip47Request, dbRequestEvent.ID, *dbRequestEvent.AppId, publishResponse)

	assert.Nil(t, publishedResponse.Error)
	// Successful response for cancel_hold_invoice has an empty result object (pointer to empty struct)
	assert.Equal(t, &cancelHoldInvoiceResponse{}, publishedResponse.Result)
}
