package controllers

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/nbd-wtf/go-nostr"
	"github.com/stretchr/testify/assert"

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
				"pubkey": "123pubkey",
				"tlv_records": [{
					"type": 5482373484,
					"value": "fajsn341414fq"
				}]
			},
			{
				"amount": 123000,
				"pubkey": "123pubkey",
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
				"pubkey": "123pubkey",
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

func TestHandleMultiPayKeysendEvent_NoPermission(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47MultiPayKeysendJson), nip47Request)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	responses := []*models.Response{}
	dTags := []nostr.Tags{}

	checkPermission := func(amountMsat uint64) *models.Response {
		return &models.Response{
			ResultType: nip47Request.Method,
			Error: &models.Error{
				Code: models.ERROR_RESTRICTED,
			},
		}
	}

	var mu sync.Mutex

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		mu.Lock()
		defer mu.Unlock()
		responses = append(responses, response)
		dTags = append(dTags, tags)
	}

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB)
	NewNip47Controller(svc.LNClient, svc.DB, svc.EventPublisher, permissionsSvc, transactionsSvc).
		HandleMultiPayKeysendEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	assert.Equal(t, 2, len(responses))
	for i := 0; i < len(responses); i++ {
		assert.Equal(t, models.ERROR_RESTRICTED, responses[i].Error.Code)
		assert.Nil(t, responses[i].Result)
	}

}

func TestHandleMultiPayKeysendEvent_WithPermission(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47MultiPayKeysendJson), nip47Request)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	responses := []*models.Response{}
	dTags := []nostr.Tags{}

	checkPermission := func(amountMsat uint64) *models.Response {
		return nil
	}

	var mu sync.Mutex

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		mu.Lock()
		defer mu.Unlock()
		responses = append(responses, response)
		dTags = append(dTags, tags)
	}

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB)
	NewNip47Controller(svc.LNClient, svc.DB, svc.EventPublisher, permissionsSvc, transactionsSvc).
		HandleMultiPayKeysendEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	assert.Equal(t, 2, len(responses))
	for i := 0; i < len(responses); i++ {
		assert.Equal(t, "12345preimage", responses[i].Result.(payResponse).Preimage)
		assert.Nil(t, responses[i].Error)
		assert.Equal(t, "123pubkey", dTags[i].GetFirst([]string{"d"}).Value())
	}
}

// TODO: fix and re-enable this as a separate test
// budget overflow
/*newMaxAmount := 500
err = svc.DB.Model(&db.AppPermission{}).Where("app_id = ?", app.ID).Update("max_amount", newMaxAmount).Error
assert.NoError(t, err)

err = json.Unmarshal([]byte(nip47MultiPayKeysendOneOverflowingBudgetJson), request)
assert.NoError(t, err)

payload, err = nip04.Encrypt(nip47MultiPayKeysendOneOverflowingBudgetJson, ss)
assert.NoError(t, err)
reqEvent.Content = payload

reqEvent.ID = "multi_pay_keysend_with_budget_overflow"
requestEvent.NostrId = reqEvent.ID
responses = []*models.Response{}
dTags = []nostr.Tags{}
svc.nip47Svc.HandleMultiPayKeysendEvent(ctx, request, requestEvent, app, publishResponse)

assert.Equal(t, responses[0].Error.Code, models.ERROR_QUOTA_EXCEEDED)
assert.Equal(t, "500pubkey", dTags[0].GetFirst([]string{"d"}).Value())
assert.Equal(t, responses[1].Result.(payResponse).Preimage, "12345preimage")
assert.Equal(t, "customId", dTags[1].GetFirst([]string{"d"}).Value())*/
