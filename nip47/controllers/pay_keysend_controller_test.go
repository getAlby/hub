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

const nip47KeysendJson = `
{
	"method": "pay_keysend",
	"params": {
		"amount": 123000,
		"pubkey": "038a73de75fdc3c7ec951a5e0b0fa95c5cd353bd7ca72df2086aa228c1f92819a5",
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
		"pubkey": "038a73de75fdc3c7ec951a5e0b0fa95c5cd353bd7ca72df2086aa228c1f92819a5",
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

	NewTestNip47Controller(svc).
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

	NewTestNip47Controller(svc).
		HandlePayKeysendEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse, nostr.Tags{})

	assert.Nil(t, publishedResponse.Error)
	assert.Equal(t, "018465013e2337234a7e5530a21c4a8cf70d84231f4a8ff0b1e2cce3cb2bd03b", publishedResponse.Result.(payResponse).Preimage)
	assert.Equal(t, uint64(1), publishedResponse.Result.(payResponse).FeesPaid)
}

func TestHandlePayKeysendEvent_InvalidParams(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	app, _, err := tests.CreateApp(svc)
	require.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: constants.PAY_INVOICE_SCOPE,
	}
	require.NoError(t, svc.DB.Create(appPermission).Error)

	validPubkey := "038a73de75fdc3c7ec951a5e0b0fa95c5cd353bd7ca72df2086aa228c1f92819a5"

	testCases := []struct {
		name          string
		params        map[string]interface{}
		expectedError string
	}{
		{
			name: "x-only nostr pubkey",
			params: map[string]interface{}{
				"amount": 123000,
				"pubkey": validPubkey[2:],
			},
			expectedError: "destination pubkey must be 33 bytes, got 32",
		},
		{
			name: "preimage too short",
			params: map[string]interface{}{
				"amount":   123000,
				"pubkey":   validPubkey,
				"preimage": "018465013e2337234a7e5530a21c4a8c",
			},
			expectedError: "preimage must be 32 bytes, got 16",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requestJson, err := json.Marshal(map[string]interface{}{
				"method": "pay_keysend",
				"params": tc.params,
			})
			require.NoError(t, err)

			nip47Request := &models.Request{}
			require.NoError(t, json.Unmarshal(requestJson, nip47Request))

			dbRequestEvent := &db.RequestEvent{NostrId: tc.name}
			require.NoError(t, svc.DB.Create(&dbRequestEvent).Error)

			var publishedResponse *models.Response
			publishResponse := func(response *models.Response, tags nostr.Tags) {
				publishedResponse = response
			}

			require.NotPanics(t, func() {
				NewTestNip47Controller(svc).
					HandlePayKeysendEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse, nostr.Tags{})
			})

			require.NotNil(t, publishedResponse)
			assert.Nil(t, publishedResponse.Result)
			require.NotNil(t, publishedResponse.Error)
			assert.Equal(t, constants.ERROR_INTERNAL, publishedResponse.Error.Code)
			assert.Equal(t, tc.expectedError, publishedResponse.Error.Message)
		})
	}
}
