package controllers

import (
	"context"
	"encoding/json"
	"slices"
	"sync"
	"testing"

	"github.com/nbd-wtf/go-nostr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/tests"
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

	NewTestNip47Controller(svc).
		HandleMultiPayKeysendEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	assert.Equal(t, 2, len(responses))
	for i := 0; i < len(responses); i++ {
		assert.Equal(t, 64, len(responses[i].Result.(payResponse).Preimage))
		assert.Equal(t, uint64(1), responses[i].Result.(payResponse).FeesPaid)
		assert.Nil(t, responses[i].Error)
		assert.Equal(t, "123pubkey2", dTags[i].Find("d")[1])
	}
}

func TestHandleMultiPayKeysendEvent_OneBudgetExceeded(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

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

	NewTestNip47Controller(svc).
		HandleMultiPayKeysendEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	// we can't guarantee which request was processed first
	// so swap them if they are back to front
	if responses[0].Result == nil {
		responses[0], responses[1] = responses[1], responses[0]
		dTags[0], dTags[1] = dTags[1], dTags[0]
	}

	assert.Equal(t, "customId", dTags[0].Find("d")[1])
	assert.Nil(t, responses[0].Error)
	assert.Equal(t, 64, len(responses[0].Result.(payResponse).Preimage))
	assert.Equal(t, uint64(1), responses[0].Result.(payResponse).FeesPaid)

	assert.Nil(t, responses[1].Result)
	assert.Equal(t, constants.ERROR_QUOTA_EXCEEDED, responses[1].Error.Code)
}

func TestHandleMultiPayKeysendEvent_IsolatedApp_ConcurrentPayments(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	app, _, err := tests.CreateApp(svc)
	app.Isolated = true
	assert.NoError(t, err)
	app.Isolated = true
	svc.DB.Save(&app)

	appPermission := &db.AppPermission{
		AppId:        app.ID,
		App:          *app,
		Scope:        constants.PAY_INVOICE_SCOPE,
		MaxAmountSat: 400,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	svc.DB.Create(&db.Transaction{
		AppId: &app.ID,
		State: constants.TRANSACTION_STATE_SETTLED,
		Type:  constants.TRANSACTION_TYPE_INCOMING,
		// keysends paid are 123000 millisats
		AmountMsat: 200000,
	})

	// force delay inside transaction
	if svc.DB.Dialector.Name() == "postgres" {
		err = svc.DB.Exec(`
CREATE OR REPLACE FUNCTION slow_down_query()
RETURNS TRIGGER AS $slow_down_query$
BEGIN
    -- Introduce a delay of 1 second
    PERFORM pg_sleep(1);
    RETURN NEW;
END;
$slow_down_query$ LANGUAGE plpgsql;

CREATE TRIGGER slow_down_query
AFTER INSERT ON transactions
FOR EACH ROW
EXECUTE PROCEDURE slow_down_query();`).Error

		require.NoError(t, err)
	}

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

	NewTestNip47Controller(svc).
		HandleMultiPayKeysendEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	require.Equal(t, 2, len(responses))

	// we can't guarantee which request was processed first
	// so put the successful one at the front
	successfulIdx := slices.IndexFunc(responses, func(r *models.Response) bool {
		return r.Result != nil
	})
	require.GreaterOrEqual(t, successfulIdx, 0)

	if successfulIdx > 0 {
		responses[0], responses[successfulIdx] = responses[successfulIdx], responses[0]
	}

	for _, response := range responses[1:] {
		require.Nil(t, response.Result)
		assert.Equal(t, constants.ERROR_INSUFFICIENT_BALANCE, response.Error.Code)
	}
}
