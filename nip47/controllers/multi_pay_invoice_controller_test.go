package controllers

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/nbd-wtf/go-nostr"
	"github.com/stretchr/testify/assert"

	"github.com/getAlby/nostr-wallet-connect/db"
	"github.com/getAlby/nostr-wallet-connect/nip47/models"
	permissions "github.com/getAlby/nostr-wallet-connect/nip47/permissions"
	"github.com/getAlby/nostr-wallet-connect/tests"
	"github.com/getAlby/nostr-wallet-connect/transactions"
)

const nip47MultiPayJson = `
{
	"method": "multi_pay_invoice",
	"params": {
		"invoices": [{
				"invoice": "lntb1230n1pjypux0pp5xgxzcks5jtx06k784f9dndjh664wc08ucrganpqn52d0ftrh9n8sdqyw3jscqzpgxqyz5vqsp5rkx7cq252p3frx8ytjpzc55rkgyx2mfkzzraa272dqvr2j6leurs9qyyssqhutxa24r5hqxstchz5fxlslawprqjnarjujp5sm3xj7ex73s32sn54fthv2aqlhp76qmvrlvxppx9skd3r5ut5xutgrup8zuc6ay73gqmra29m"
			},
			{
				"invoice": "lntb1230n1pjypux0pp5xgxzcks5jtx06k784f9dndjh664wc08ucrganpqn52d0ftrh9n8sdqyw3jscqzpgxqyz5vqsp5rkx7cq252p3frx8ytjpzc55rkgyx2mfkzzraa272dqvr2j6leurs9qyyssqhutxa24r5hqxstchz5fxlslawprqjnarjujp5sm3xj7ex73s32sn54fthv2aqlhp76qmvrlvxppx9skd3r5ut5xutgrup8zuc6ay73gqmra29m"
			}
		]
	}
}
`

const nip47MultiPayOneOverflowingBudgetJson = `
{
	"method": "multi_pay_invoice",
	"params": {
		"invoices": [{
				"invoice": "lnbcrt5u1pjuywzppp5h69dt59cypca2wxu69sw8ga0g39a3yx7dqug5nthrw3rcqgfdu4qdqqcqzzsxqyz5vqsp5gzlpzszyj2k30qmpme7jsfzr24wqlvt9xdmr7ay34lfelz050krs9qyyssq038x07nh8yuv8hdpjh5y8kqp7zcd62ql9na9xh7pla44htjyy02sz23q7qm2tza6ct4ypljk54w9k9qsrsu95usk8ce726ytep6vhhsq9mhf9a"
			},
			{
				"invoice": "lntb1230n1pjypux0pp5xgxzcks5jtx06k784f9dndjh664wc08ucrganpqn52d0ftrh9n8sdqyw3jscqzpgxqyz5vqsp5rkx7cq252p3frx8ytjpzc55rkgyx2mfkzzraa272dqvr2j6leurs9qyyssqhutxa24r5hqxstchz5fxlslawprqjnarjujp5sm3xj7ex73s32sn54fthv2aqlhp76qmvrlvxppx9skd3r5ut5xutgrup8zuc6ay73gqmra29m"
			}
		]
	}
}
`

const nip47MultiPayOneMalformedInvoiceJson = `
{
	"method": "multi_pay_invoice",
	"params": {
		"invoices": [{
				"invoice": "",
				"id": "invoiceId123"
			},
			{
				"invoice": "lntb1230n1pjypux0pp5xgxzcks5jtx06k784f9dndjh664wc08ucrganpqn52d0ftrh9n8sdqyw3jscqzpgxqyz5vqsp5rkx7cq252p3frx8ytjpzc55rkgyx2mfkzzraa272dqvr2j6leurs9qyyssqhutxa24r5hqxstchz5fxlslawprqjnarjujp5sm3xj7ex73s32sn54fthv2aqlhp76qmvrlvxppx9skd3r5ut5xutgrup8zuc6ay73gqmra29m"
			}
		]
	}
}
`

func TestHandleMultiPayInvoiceEvent_NoPermission(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47MultiPayJson), nip47Request)
	assert.NoError(t, err)

	responses := []*models.Response{}
	dTags := []nostr.Tags{}

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
		HandleMultiPayInvoiceEvent(ctx, nip47Request, dbRequestEvent.ID, app, checkPermission, publishResponse)

	assert.Equal(t, 2, len(responses))
	assert.Equal(t, 2, len(dTags))
	for i := 0; i < len(responses); i++ {
		assert.Equal(t, models.ERROR_RESTRICTED, responses[i].Error.Code)
		assert.Equal(t, tests.MockPaymentHash, dTags[i].GetFirst([]string{"d"}).Value())
		assert.Equal(t, responses[i].Result, nil)
	}
}

func TestHandleMultiPayInvoiceEvent_WithPermission(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47MultiPayJson), nip47Request)
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

	checkPermission := func(amountMsat uint64) *models.Response {
		return nil
	}
	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB)
	NewNip47Controller(svc.LNClient, svc.DB, svc.EventPublisher, permissionsSvc, transactionsSvc).
		HandleMultiPayInvoiceEvent(ctx, nip47Request, dbRequestEvent.ID, app, checkPermission, publishResponse)

	assert.Equal(t, 2, len(responses))
	for i := 0; i < len(responses); i++ {
		assert.Equal(t, "123preimage", responses[i].Result.(payResponse).Preimage)
		assert.Equal(t, tests.MockPaymentHash, dTags[i].GetFirst([]string{"d"}).Value())
		assert.Nil(t, responses[i].Error)
	}

}

func TestHandleMultiPayInvoiceEvent_OneMalformedInvoice(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47MultiPayOneMalformedInvoiceJson), nip47Request)
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

	checkPermission := func(amountMsat uint64) *models.Response {
		return nil
	}
	requestEvent := &db.RequestEvent{}
	svc.DB.Save(requestEvent)

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB)
	NewNip47Controller(svc.LNClient, svc.DB, svc.EventPublisher, permissionsSvc, transactionsSvc).
		HandleMultiPayInvoiceEvent(ctx, nip47Request, requestEvent.ID, app, checkPermission, publishResponse)

	assert.Equal(t, 2, len(responses))
	assert.Equal(t, "invoiceId123", dTags[0].GetFirst([]string{"d"}).Value())
	assert.Equal(t, models.ERROR_INTERNAL, responses[0].Error.Code)
	assert.Nil(t, responses[0].Result)

	assert.Equal(t, tests.MockPaymentHash, dTags[1].GetFirst([]string{"d"}).Value())
	assert.Equal(t, "123preimage", responses[1].Result.(payResponse).Preimage)
	assert.Nil(t, responses[1].Error)

}

// TODO: fix and re-enable this as a separate test
// // budget overflow
// newMaxAmount := 500
// err = svc.DB.Model(&db.AppPermission{}).Where("app_id = ?", app.ID).Update("max_amount", newMaxAmount).Error
// assert.NoError(t, err)

// err = json.Unmarshal([]byte(nip47MultiPayOneOverflowingBudgetJson), nip47Request)
// assert.NoError(t, err)

// responses = []*models.Response{}
// dTags = []nostr.Tags{}
// NewMultiPayInvoiceController(svc.LNClient, svc.DB, svc.EventPublisher).
// 	HandleMultiPayInvoiceEvent(ctx, nip47Request, requestEvent.ID, app, checkPermission, publishResponse)

// // might be flaky because the two requests run concurrently
// // and there's more chance that the failed respons calls the
// // publishResponse as it's called earlier
// assert.Equal(t, responses[0].Error.Code, models.ERROR_QUOTA_EXCEEDED)
// assert.Equal(t, tests.MockPaymentHash500, dTags[0].GetFirst([]string{"d"}).Value())
// assert.Equal(t, responses[1].Result.(payResponse).Preimage, "123preimage")
// assert.Equal(t, tests.MockPaymentHash, dTags[1].GetFirst([]string{"d"}).Value())
