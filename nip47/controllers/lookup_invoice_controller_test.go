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

const nip47LookupInvoiceJson = `
{
	"method": "lookup_invoice",
	"params": {
		"payment_hash": "4ad9cd27989b514d868e755178378019903a8d78767e3fceb211af9dd00e7a94"
	}
}
`

func TestHandleLookupInvoiceEvent_NoPermission(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47LookupInvoiceJson), nip47Request)
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

	NewLookupInvoiceController(svc.LNClient).
		HandleLookupInvoiceEvent(ctx, nip47Request, dbRequestEvent.ID, checkPermission, publishResponse)

	assert.Nil(t, publishedResponse.Result)
	assert.Equal(t, models.ERROR_RESTRICTED, publishedResponse.Error.Code)
}

func TestHandleLookupInvoiceEvent_WithPermission(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47LookupInvoiceJson), nip47Request)
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

	NewLookupInvoiceController(svc.LNClient).
		HandleLookupInvoiceEvent(ctx, nip47Request, dbRequestEvent.ID, checkPermission, publishResponse)

	assert.Nil(t, publishedResponse.Error)
	transaction := publishedResponse.Result.(*lookupInvoiceResponse)
	assert.Equal(t, tests.MockTransaction.Type, transaction.Type)
	assert.Equal(t, tests.MockTransaction.Invoice, transaction.Invoice)
	assert.Equal(t, tests.MockTransaction.Description, transaction.Description)
	assert.Equal(t, tests.MockTransaction.DescriptionHash, transaction.DescriptionHash)
	assert.Equal(t, tests.MockTransaction.Preimage, transaction.Preimage)
	assert.Equal(t, tests.MockTransaction.PaymentHash, transaction.PaymentHash)
	assert.Equal(t, tests.MockTransaction.Amount, transaction.Amount)
	assert.Equal(t, tests.MockTransaction.FeesPaid, transaction.FeesPaid)
	assert.Equal(t, tests.MockTransaction.SettledAt, transaction.SettledAt)
}
