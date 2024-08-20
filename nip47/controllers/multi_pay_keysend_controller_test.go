package controllers

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/nbd-wtf/go-nostr"
	"github.com/stretchr/testify/assert"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/nip47/permissions"
	"github.com/getAlby/hub/tests"
	"github.com/getAlby/hub/transactions"
)

const nip47MultiPayKeysendJson = `
{
	"method": "multi_pay_keysend",
	"params": {
		"keysends": [{
				"amount": 123000,
				"pubkey": "123pubkey2",
				"tlv_records": [{
					"type": 5482373484,
					"value": "fajsn341414fq"
				}]
			},
			{
				"amount": 123000,
				"pubkey": "123pubkey2",
				"tlv_records": [{
					"type": 5482373484,
					"value": "fajsn341414fq"
				}]
			}
		]
	}
}
`

const nip47MultiPayKeysendOneOverflowingBudgetJson = `
{
	"method": "multi_pay_keysend",
	"params": {
		"keysends": [{
				"amount": 123000,
				"pubkey": "123pubkey2",
				"id": "customId",
				"tlv_records": [{
					"type": 5482373484,
					"value": "fajsn341414fq"
				}]
			},
			{
				"amount": 500000,
				"pubkey": "500pubkey",
				"tlv_records": [{
					"type": 5482373484,
					"value": "fajsn341414fq"
				}]
			}
		]
	}
}
`

func TestHandleMultiPayKeysendEvent_Success(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

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
	err = json.Unmarshal([]byte(nip47MultiPayKeysendJson), nip47Request)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	responses := []*models.Response{}
	dTags := []nostr.Tags{}

	var mu sync.Mutex

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		mu.Lock()
		defer mu.Unlock()
		responses = append(responses, response)
		dTags = append(dTags, tags)
	}

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB, svc.EventPublisher)
	NewNip47Controller(svc.LNClient, svc.DB, svc.EventPublisher, permissionsSvc, transactionsSvc).
		HandleMultiPayKeysendEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	assert.Equal(t, 2, len(responses))
	for i := 0; i < len(responses); i++ {
		assert.Equal(t, 64, len(responses[i].Result.(payResponse).Preimage))
		assert.Equal(t, uint64(1), responses[i].Result.(payResponse).FeesPaid)
		assert.Nil(t, responses[i].Error)
		assert.Equal(t, "123pubkey2", dTags[i].GetFirst([]string{"d"}).Value())
	}
}

func TestHandleMultiPayKeysendEvent_OneBudgetExceeded(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId:        app.ID,
		App:          *app,
		Scope:        constants.PAY_INVOICE_SCOPE,
		MaxAmountSat: 400,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47MultiPayKeysendOneOverflowingBudgetJson), nip47Request)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	responses := []*models.Response{}
	dTags := []nostr.Tags{}

	var mu sync.Mutex

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		mu.Lock()
		defer mu.Unlock()
		responses = append(responses, response)
		dTags = append(dTags, tags)
	}

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB, svc.EventPublisher)
	NewNip47Controller(svc.LNClient, svc.DB, svc.EventPublisher, permissionsSvc, transactionsSvc).
		HandleMultiPayKeysendEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	// we can't guarantee which request was processed first
	// so swap them if they are back to front
	if responses[0].Result == nil {
		responses[0], responses[1] = responses[1], responses[0]
		dTags[0], dTags[1] = dTags[1], dTags[0]
	}

	assert.Equal(t, "customId", dTags[0].GetFirst([]string{"d"}).Value())
	assert.Nil(t, responses[0].Error)
	assert.Equal(t, 64, len(responses[0].Result.(payResponse).Preimage))
	assert.Equal(t, uint64(1), responses[0].Result.(payResponse).FeesPaid)

	assert.Nil(t, responses[1].Result)
	assert.Equal(t, constants.ERROR_QUOTA_EXCEEDED, responses[1].Error.Code)
}
