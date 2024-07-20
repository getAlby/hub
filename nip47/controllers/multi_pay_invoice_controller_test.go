package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/nbd-wtf/go-nostr"
	"github.com/stretchr/testify/assert"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/nip47/permissions"
	"github.com/getAlby/hub/tests"
	"github.com/getAlby/hub/transactions"
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

func TestHandleMultiPayInvoiceEvent(t *testing.T) {
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

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB)
	NewNip47Controller(svc.LNClient, svc.DB, svc.EventPublisher, permissionsSvc, transactionsSvc).
		HandleMultiPayInvoiceEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

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

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: constants.PAY_INVOICE_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
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

	requestEvent := &db.RequestEvent{}
	svc.DB.Save(requestEvent)

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB)
	NewNip47Controller(svc.LNClient, svc.DB, svc.EventPublisher, permissionsSvc, transactionsSvc).
		HandleMultiPayInvoiceEvent(ctx, nip47Request, requestEvent.ID, app, publishResponse)

	assert.Equal(t, 2, len(responses))
	assert.Equal(t, 2, len(dTags))

	// we can't guarantee which request was processed first
	// so swap them if they are back to front
	if responses[0].Result != nil {
		responses[0], responses[1] = responses[1], responses[0]
		dTags[0], dTags[1] = dTags[1], dTags[0]
	}

	assert.Equal(t, "invoiceId123", dTags[0].GetFirst([]string{"d"}).Value())
	assert.Equal(t, models.ERROR_INTERNAL, responses[0].Error.Code)
	assert.Nil(t, responses[0].Result)

	assert.Equal(t, tests.MockPaymentHash, dTags[1].GetFirst([]string{"d"}).Value())
	assert.Equal(t, "123preimage", responses[1].Result.(payResponse).Preimage)
	assert.Nil(t, responses[1].Error)

}

func TestHandleMultiPayInvoiceEvent_IsolatedApp_OneBudgetExceeded(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)
	app.Isolated = true
	svc.DB.Save(&app)

	svc.DB.Create(&db.Transaction{
		AppId: &app.ID,
		State: constants.TRANSACTION_STATE_SETTLED,
		Type:  constants.TRANSACTION_TYPE_INCOMING,
		// invoices paid are 123000 millisats
		AmountMsat: 200000,
	})

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: constants.PAY_INVOICE_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
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

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB)
	NewNip47Controller(svc.LNClient, svc.DB, svc.EventPublisher, permissionsSvc, transactionsSvc).
		HandleMultiPayInvoiceEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	assert.Equal(t, 2, len(responses))
	assert.Equal(t, 2, len(dTags))

	// we can't guarantee which request was processed first
	// so swap them if they are back to front
	if responses[0].Result == nil {
		responses[0], responses[1] = responses[1], responses[0]
		dTags[0], dTags[1] = dTags[1], dTags[0]
	}

	assert.Equal(t, "320c2c5a1492ccfd5bc7aa4ad9b657d6aaec3cfcc0d1d98413a29af4ac772ccf", dTags[0].GetFirst([]string{"d"}).Value())
	assert.Equal(t, "123preimage", responses[0].Result.(payResponse).Preimage)
	assert.Nil(t, responses[0].Error)

	assert.Equal(t, tests.MockPaymentHash, dTags[1].GetFirst([]string{"d"}).Value())
	assert.Nil(t, responses[1].Result)
	assert.Equal(t, models.ERROR_INSUFFICIENT_BALANCE, responses[1].Error.Code)
}

func TestHandleMultiPayInvoiceEvent_LNClient_OnePaymentFailed(t *testing.T) {

	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)
	svc.LNClient.(*tests.MockLn).PayInvoiceResponses = []*lnclient.PayInvoiceResponse{{
		Preimage: "123preimage",
	}, nil}
	svc.LNClient.(*tests.MockLn).PayInvoiceErrors = []error{nil, errors.New("Some error")}

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

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB)
	NewNip47Controller(svc.LNClient, svc.DB, svc.EventPublisher, permissionsSvc, transactionsSvc).
		HandleMultiPayInvoiceEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	assert.Equal(t, 2, len(responses))

	assert.Equal(t, "320c2c5a1492ccfd5bc7aa4ad9b657d6aaec3cfcc0d1d98413a29af4ac772ccf", dTags[0].GetFirst([]string{"d"}).Value())
	assert.Equal(t, "123preimage", responses[0].Result.(payResponse).Preimage)
	assert.Nil(t, responses[0].Error)

	assert.Equal(t, tests.MockPaymentHash, dTags[1].GetFirst([]string{"d"}).Value())
	assert.Nil(t, responses[1].Result)
	assert.Equal(t, models.ERROR_INTERNAL, responses[1].Error.Code)
	assert.Equal(t, "Some error", responses[1].Error.Message)
}
