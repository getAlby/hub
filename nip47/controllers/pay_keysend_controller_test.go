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

const nip47KeysendJson = `
{
	"method": "pay_keysend",
	"params": {
		"amount": 123000,
		"pubkey": "123pubkey",
		"tlv_records": [{
			"type": 5482373484,
			"value": "fajsn341414fq"
		}]
	}
}
`

func TestHandlePayKeysendEvent_NoPermission(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47KeysendJson), nip47Request)
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

	NewPayKeysendController(svc.LNClient, svc.DB, svc.EventPublisher).
		HandlePayKeysendEvent(ctx, nip47Request, dbRequestEvent.ID, app, checkPermission, publishResponse, nostr.Tags{})

	assert.Nil(t, publishedResponse.Result)
	assert.Equal(t, models.ERROR_RESTRICTED, publishedResponse.Error.Code)
}

func TestHandlePayKeysendEvent_WithPermission(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47KeysendJson), nip47Request)
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

	NewPayKeysendController(svc.LNClient, svc.DB, svc.EventPublisher).
		HandlePayKeysendEvent(ctx, nip47Request, dbRequestEvent.ID, app, checkPermission, publishResponse, nostr.Tags{})

	assert.Nil(t, publishedResponse.Error)
	assert.Equal(t, "12345preimage", publishedResponse.Result.(payResponse).Preimage)
}
