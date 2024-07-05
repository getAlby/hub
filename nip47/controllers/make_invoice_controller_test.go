package controllers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/nbd-wtf/go-nostr"
	"github.com/stretchr/testify/assert"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/tests"
)

const nip47MakeInvoiceJson = `
{
	"method": "make_invoice",
	"params": {
		"amount": 1000,
		"description": "[[\"text/identifier\",\"hello@getalby.com\"],[\"text/plain\",\"Sats for Alby\"]]",
		"expiry": 3600
	}
}
`

func TestHandleMakeInvoiceEvent_NoPermission(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47MakeInvoiceJson), nip47Request)
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

	NewMakeInvoiceController(svc.LNClient).
		HandleMakeInvoiceEvent(ctx, nip47Request, dbRequestEvent.ID, checkPermission, publishResponse)

	assert.Nil(t, publishedResponse.Result)
	assert.Equal(t, models.ERROR_RESTRICTED, publishedResponse.Error.Code)
}

func TestHandleMakeInvoiceEvent_WithPermission(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47MakeInvoiceJson), nip47Request)
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

	NewMakeInvoiceController(svc.LNClient).
		HandleMakeInvoiceEvent(ctx, nip47Request, dbRequestEvent.ID, checkPermission, publishResponse)

	assert.Nil(t, publishedResponse.Error)
	assert.Equal(t, tests.MockTransaction.Invoice, publishedResponse.Result.(*makeInvoiceResponse).Invoice)
}
