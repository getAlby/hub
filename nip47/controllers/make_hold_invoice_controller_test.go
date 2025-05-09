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

const nip47MakeHoldInvoiceJson = `
{
	"method": "make_hold_invoice",
	"params": {
		"amount": 1000,
		"description": "Hello, world",
		"payment_hash": "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		"expiry": 3600,
		"metadata": {
		  "a": 1,
			"b": "2",
			"c": {
			  "d": 3,
				"e": [{
					"f": "g"
				},{
					"h": "i"
				}]
			}
		}
	}
}
`

func TestHandleMakeHoldInvoiceEvent(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47MakeHoldInvoiceJson), nip47Request)
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
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
		HandleMakeHoldInvoiceEvent(ctx, nip47Request, dbRequestEvent.ID, *dbRequestEvent.AppId, publishResponse)

	expectedMetadata := map[string]interface{}{
		"a": float64(1),
		"b": "2",
		"c": map[string]interface{}{
			"d": float64(3),
			"e": []interface{}{
				map[string]interface{}{"f": "g"},
				map[string]interface{}{"h": "i"},
			},
		},
	}

	assert.Nil(t, publishedResponse.Error)
	assert.Equal(t, tests.MockLNClientHoldTransaction.Invoice, publishedResponse.Result.(*makeHoldInvoiceResponse).Invoice)
	assert.Equal(t, tests.MockLNClientHoldTransaction.PaymentHash, publishedResponse.Result.(*makeHoldInvoiceResponse).PaymentHash)
	assert.Equal(t, expectedMetadata, publishedResponse.Result.(*makeHoldInvoiceResponse).Metadata)
}

const nip47MakeHoldInvoiceMissingPaymentHashJson = `
{
"method": "make_hold_invoice",
"params": {
"amount": 1000,
"description": "Hello, world",
"expiry": 3600
}
}
`

func TestHandleMakeHoldInvoiceEvent_MissingPaymentHash(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47MakeHoldInvoiceMissingPaymentHashJson), nip47Request)
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
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
		HandleMakeHoldInvoiceEvent(ctx, nip47Request, dbRequestEvent.ID, *dbRequestEvent.AppId, publishResponse)

	require.NotNil(t, publishedResponse.Error)
	assert.Equal(t, constants.ERROR_BAD_REQUEST, publishedResponse.Error.Code)
}
