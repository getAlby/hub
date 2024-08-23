package controllers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/nbd-wtf/go-nostr"
	"github.com/stretchr/testify/assert"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/nip47/permissions"
	"github.com/getAlby/hub/tests"
	"github.com/getAlby/hub/transactions"
)

const nip47MakeInvoiceJson = `
{
	"method": "make_invoice",
	"params": {
		"amount": 1000,
		"description": "Hello, world",
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

func TestHandleMakeInvoiceEvent(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47MakeInvoiceJson), nip47Request)
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

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB, svc.EventPublisher)
	NewNip47Controller(svc.LNClient, svc.DB, svc.EventPublisher, permissionsSvc, transactionsSvc).
		HandleMakeInvoiceEvent(ctx, nip47Request, dbRequestEvent.ID, *dbRequestEvent.AppId, publishResponse)

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
	assert.Equal(t, tests.MockLNClientTransaction.Invoice, publishedResponse.Result.(*makeInvoiceResponse).Invoice)
	assert.Equal(t, expectedMetadata, publishedResponse.Result.(*makeInvoiceResponse).Metadata)
}
