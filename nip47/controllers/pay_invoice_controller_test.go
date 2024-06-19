package controllers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/nbd-wtf/go-nostr"
	"github.com/stretchr/testify/assert"

	"github.com/getAlby/nostr-wallet-connect/db"
	"github.com/getAlby/nostr-wallet-connect/nip47/models"
	"github.com/getAlby/nostr-wallet-connect/tests"
)

const nip47PayInvoiceJson = `
{
	"method": "pay_invoice",
	"params": {
		"invoice": "lntb1230n1pjypux0pp5xgxzcks5jtx06k784f9dndjh664wc08ucrganpqn52d0ftrh9n8sdqyw3jscqzpgxqyz5vqsp5rkx7cq252p3frx8ytjpzc55rkgyx2mfkzzraa272dqvr2j6leurs9qyyssqhutxa24r5hqxstchz5fxlslawprqjnarjujp5sm3xj7ex73s32sn54fthv2aqlhp76qmvrlvxppx9skd3r5ut5xutgrup8zuc6ay73gqmra29m"
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

func TestHandlePayInvoiceEvent_NoPermission(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47PayInvoiceJson), nip47Request)
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

	NewPayInvoiceController(svc.LNClient, svc.DB, svc.EventPublisher).
		HandlePayInvoiceEvent(ctx, nip47Request, dbRequestEvent.ID, app, checkPermission, publishResponse, nostr.Tags{})

	assert.Equal(t, models.ERROR_RESTRICTED, publishedResponse.Error.Code)

}

func TestHandlePayInvoiceEvent_WithPermission(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47PayInvoiceJson), nip47Request)
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

	NewPayInvoiceController(svc.LNClient, svc.DB, svc.EventPublisher).
		HandlePayInvoiceEvent(ctx, nip47Request, dbRequestEvent.ID, app, checkPermission, publishResponse, nostr.Tags{})

	assert.Equal(t, "123preimage", publishedResponse.Result.(payResponse).Preimage)
}

func TestHandlePayInvoiceEvent_MalformedInvoice(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47PayJsonNoInvoice), nip47Request)
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

	NewPayInvoiceController(svc.LNClient, svc.DB, svc.EventPublisher).
		HandlePayInvoiceEvent(ctx, nip47Request, dbRequestEvent.ID, app, checkPermission, publishResponse, nostr.Tags{})

	assert.Nil(t, publishedResponse.Result)
	assert.Equal(t, models.ERROR_INTERNAL, publishedResponse.Error.Code)
}
