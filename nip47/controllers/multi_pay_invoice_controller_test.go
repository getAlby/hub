package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
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
				"invoice": "lntbs1230n1pnvxqc2dqqnp4q0w0f29u6f7yrrpr5y6wj45gtnyhtch9u2m2j7qrws8eevrw90c72pp57gnea9rwqh9c62dl67akgyhuxm7dd3fgwufyuyctgx3awuv8f7cqsp56rtp7kryxssfp3lk7h79uv7n55dc4nwuvslva64caxz45ysefmeq9qyysgqcqpcxqyz5vq7trlnnrjjtfkaw3evfgqh7nxayppkvlkxa2nzhg39zs372j7hff8kht7j40hl0elh2ukhu26nzawvk3aqszdl8ppxhzsgtumemewtccq3xryqt"
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

func TestHandleMultiPayInvoiceEvent_Success(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	var preimages = []string{"123preimage", "123preimage2"}

	svc.LNClient.(*tests.MockLn).PayInvoiceResponses = []*lnclient.PayInvoiceResponse{{
		Preimage: preimages[0],
	}, {
		Preimage: preimages[1],
	}}
	svc.LNClient.(*tests.MockLn).PayInvoiceErrors = []error{nil, nil}

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
	transactionsSvc := transactions.NewTransactionsService(svc.DB, svc.EventPublisher)
	NewNip47Controller(svc.LNClient, svc.DB, svc.EventPublisher, permissionsSvc, transactionsSvc).
		HandleMultiPayInvoiceEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	var paymentHashes = []string{
		"320c2c5a1492ccfd5bc7aa4ad9b657d6aaec3cfcc0d1d98413a29af4ac772ccf",
		"f2279e946e05cb8d29bfd7bb6412fc36fcd6c52877124e130b41a3d771874fb0",
	}

	assert.Equal(t, 2, len(responses))

	// we can't guarantee which request was processed first
	// so swap them if they are back to front
	if dTags[0].GetFirst([]string{"d"}).Value() != paymentHashes[0] {
		responses[0], responses[1] = responses[1], responses[0]
		dTags[0], dTags[1] = dTags[1], dTags[0]
		preimages[0], preimages[1] = preimages[1], preimages[0]
	}

	for i := 0; i < len(responses); i++ {
		assert.Equal(t, preimages[i], responses[i].Result.(payResponse).Preimage)
		assert.Equal(t, paymentHashes[i], dTags[i].GetFirst([]string{"d"}).Value())
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
	transactionsSvc := transactions.NewTransactionsService(svc.DB, svc.EventPublisher)
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
	assert.Equal(t, constants.ERROR_INTERNAL, responses[0].Error.Code)
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
	transactionsSvc := transactions.NewTransactionsService(svc.DB, svc.EventPublisher)
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

	// we cannot guarantee which payment will be made first,
	// so ensure we have results for both payment hashes
	var paymentHashes = []string{
		"320c2c5a1492ccfd5bc7aa4ad9b657d6aaec3cfcc0d1d98413a29af4ac772ccf",
		"f2279e946e05cb8d29bfd7bb6412fc36fcd6c52877124e130b41a3d771874fb0",
	}

	assert.NotEqual(t, dTags[0].GetFirst([]string{"d"}).Value(), dTags[1].GetFirst([]string{"d"}).Value())

	assert.Contains(t, paymentHashes, dTags[0].GetFirst([]string{"d"}).Value())
	assert.Equal(t, "123preimage", responses[0].Result.(payResponse).Preimage)
	assert.Nil(t, responses[0].Error)

	assert.Contains(t, paymentHashes, dTags[1].GetFirst([]string{"d"}).Value())
	assert.Nil(t, responses[1].Result)
	assert.Equal(t, constants.ERROR_INSUFFICIENT_BALANCE, responses[1].Error.Code)
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
		logger.Logger.WithFields(logrus.Fields{
			"response": response,
			"tags":     tags,
		}).Info("Publish response")
		mu.Lock()
		defer mu.Unlock()
		responses = append(responses, response)
		dTags = append(dTags, tags)
	}

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB, svc.EventPublisher)
	NewNip47Controller(svc.LNClient, svc.DB, svc.EventPublisher, permissionsSvc, transactionsSvc).
		HandleMultiPayInvoiceEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	assert.Equal(t, 2, len(responses))
	assert.Equal(t, 2, len(dTags))

	logger.Logger.WithField("dTags", dTags).WithField("responses", responses).Info("Got responses")
	// we can't guarantee which request was processed first
	// so swap them if they are back to front
	if responses[0].Result == nil {
		responses[0], responses[1] = responses[1], responses[0]
		dTags[0], dTags[1] = dTags[1], dTags[0]
	}

	// we cannot guarantee which payment will be made first,
	// so ensure we have results for both payment hashes
	var paymentHashes = []string{
		"320c2c5a1492ccfd5bc7aa4ad9b657d6aaec3cfcc0d1d98413a29af4ac772ccf",
		"f2279e946e05cb8d29bfd7bb6412fc36fcd6c52877124e130b41a3d771874fb0",
	}

	assert.NotEqual(t, dTags[0].GetFirst([]string{"d"}).Value(), dTags[1].GetFirst([]string{"d"}).Value())

	assert.Contains(t, paymentHashes, dTags[0].GetFirst([]string{"d"}).Value())
	assert.Equal(t, "123preimage", responses[0].Result.(payResponse).Preimage)
	assert.Nil(t, responses[0].Error)

	assert.Contains(t, paymentHashes, dTags[1].GetFirst([]string{"d"}).Value())
	assert.Nil(t, responses[1].Result)
	assert.Equal(t, constants.ERROR_INTERNAL, responses[1].Error.Code)
	assert.Equal(t, "Some error", responses[1].Error.Message)
}
