package controllers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/lightningnetwork/lnd/zpay32"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/tests"
)

const nip47MultiPayJson = `
{
	"method": "multi_pay_invoice",
	"params": {
		"invoices": [{
				"invoice": "lntbs1230n1pnkqautdqyw3jsnp4q09a0z84kg4a2m38zjllw43h953fx5zvqe8qxfgw694ymkq26u8zcpp5yvnh6hsnlnj4xnuh2trzlnunx732dv8ta2wjr75pdfxf6p2vlyassp5hyeg97a3ft5u769kjwsn7p0e85h79pzz8kladmnqhpcypz2uawjs9qyysgqcqpcxq8zals8sq9yeg2pa9eywkgj50cyzxd5elatujuc0c0wh6j9nat5mn34pgk8u9ufpgs99tw9ldlfk42cqlkr48au3lmuh09269prg4qkggh4a8cyqpfl0y6j"
			},
			{
				"invoice": "lntbs1230n1pnkq7q2dqqnp4q09a0z84kg4a2m38zjllw43h953fx5zvqe8qxfgw694ymkq26u8zcpp54sde879ktfrwnt4re3t2ckkrt5tr6dgv6cfdjgkar7942ruccvuqsp52qlk3rxr926s630fmnc5mg6sexnng4cyyfas4msrms8j6q28j8ys9qyysgqcqpcxq8zals8sqgjd3a60n6dy92jn7ggtkywhw952sc302qj0cwfupp7gayadznaj5cahvuq7py8p7hnq8yxylru6279urzxta3783cxze2atj9zmwadcq36muep"
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
				"invoice": "lntbs1230n1pnkqautdqyw3jsnp4q09a0z84kg4a2m38zjllw43h953fx5zvqe8qxfgw694ymkq26u8zcpp5yvnh6hsnlnj4xnuh2trzlnunx732dv8ta2wjr75pdfxf6p2vlyassp5hyeg97a3ft5u769kjwsn7p0e85h79pzz8kladmnqhpcypz2uawjs9qyysgqcqpcxq8zals8sq9yeg2pa9eywkgj50cyzxd5elatujuc0c0wh6j9nat5mn34pgk8u9ufpgs99tw9ldlfk42cqlkr48au3lmuh09269prg4qkggh4a8cyqpfl0y6j"
			}
		]
	}
}
`

// the first invoice is expired
const nip47MultiPayOneExpiredInvoiceJson = `
{
	"method": "multi_pay_invoice",
	"params": {
		"invoices": [{
				"invoice": "lntb1230n1pjypux0pp5xgxzcks5jtx06k784f9dndjh664wc08ucrganpqn52d0ftrh9n8sdqyw3jscqzpgxqyz5vqsp5rkx7cq252p3frx8ytjpzc55rkgyx2mfkzzraa272dqvr2j6leurs9qyyssqhutxa24r5hqxstchz5fxlslawprqjnarjujp5sm3xj7ex73s32sn54fthv2aqlhp76qmvrlvxppx9skd3r5ut5xutgrup8zuc6ay73gqmra29m"
			},
			{
				"invoice": "lntbs1230n1pnkqautdqyw3jsnp4q09a0z84kg4a2m38zjllw43h953fx5zvqe8qxfgw694ymkq26u8zcpp5yvnh6hsnlnj4xnuh2trzlnunx732dv8ta2wjr75pdfxf6p2vlyassp5hyeg97a3ft5u769kjwsn7p0e85h79pzz8kladmnqhpcypz2uawjs9qyysgqcqpcxq8zals8sq9yeg2pa9eywkgj50cyzxd5elatujuc0c0wh6j9nat5mn34pgk8u9ufpgs99tw9ldlfk42cqlkr48au3lmuh09269prg4qkggh4a8cyqpfl0y6j"
			}
		]
	}
}
`
const MockExpiredPaymentHash = "320c2c5a1492ccfd5bc7aa4ad9b657d6aaec3cfcc0d1d98413a29af4ac772ccf" // for the expired invoice

func TestHandleMultiPayInvoiceEvent_Success(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

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

	NewTestNip47Controller(svc).
		HandleMultiPayInvoiceEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	var paymentHashes = []string{
		"23277d5e13fce5534f9752c62fcf9337a2a6b0ebea9d21fa816a4c9d054cf93b",
		"ac1b93f8b65a46e9aea3cc56ac5ac35d163d350cd612d922dd1f8b550f98c338",
	}

	assert.Equal(t, 2, len(responses))

	// we can't guarantee which request was processed first
	// so swap them if they are back to front
	if dTags[0].GetFirst([]string{"d"}).Value() != paymentHashes[0] {
		responses[0], responses[1] = responses[1], responses[0]
		dTags[0], dTags[1] = dTags[1], dTags[0]
	}
	// we can't guarantee which request was processed first
	// so swap them if they are back to front
	if responses[0].Result.(payResponse).Preimage != preimages[0] {
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

	NewTestNip47Controller(svc).
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

func TestHandleMultiPayInvoiceEvent_OneExpiredInvoice(t *testing.T) {
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
	err = json.Unmarshal([]byte(nip47MultiPayOneExpiredInvoiceJson), nip47Request)
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

	NewTestNip47Controller(svc).
		HandleMultiPayInvoiceEvent(ctx, nip47Request, requestEvent.ID, app, publishResponse)

	assert.Equal(t, 2, len(responses))
	assert.Equal(t, 2, len(dTags))

	// we can't guarantee which request was processed first
	// so swap them if they are back to front
	if responses[0].Result != nil {
		responses[0], responses[1] = responses[1], responses[0]
		dTags[0], dTags[1] = dTags[1], dTags[0]
	}

	assert.Equal(t, MockExpiredPaymentHash, dTags[0].GetFirst([]string{"d"}).Value())
	assert.Equal(t, constants.ERROR_INTERNAL, responses[0].Error.Code)
	assert.Equal(t, "this invoice has expired", responses[0].Error.Message)
	assert.Nil(t, responses[0].Result)

	assert.Equal(t, tests.MockPaymentHash, dTags[1].GetFirst([]string{"d"}).Value())
	assert.Equal(t, "123preimage", responses[1].Result.(payResponse).Preimage)
	assert.Nil(t, responses[1].Error)
}

func TestHandleMultiPayInvoiceEvent_IsolatedApp_OneBudgetExceeded(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

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

	NewTestNip47Controller(svc).
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
		"23277d5e13fce5534f9752c62fcf9337a2a6b0ebea9d21fa816a4c9d054cf93b",
		"ac1b93f8b65a46e9aea3cc56ac5ac35d163d350cd612d922dd1f8b550f98c338",
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

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()
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

	NewTestNip47Controller(svc).
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
		"23277d5e13fce5534f9752c62fcf9337a2a6b0ebea9d21fa816a4c9d054cf93b",
		"ac1b93f8b65a46e9aea3cc56ac5ac35d163d350cd612d922dd1f8b550f98c338",
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

func TestHandleMultiPayInvoiceEvent_IsolatedApp_ConcurrentPayments(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

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

	tm := time.Now()
	const amountMsat = 123000

	const nInvoices = 80

	type Invoice struct {
		Decoded     *zpay32.Invoice
		Encoded     string
		PaymentHash string
	}

	newDummyInvoice := func(amountMsat uint64, descr string, tm time.Time) (*Invoice, error) {
		random := make([]byte, 32)
		_, err := rand.Read(random)
		if err != nil {
			return nil, fmt.Errorf("failed to read random bytes: %w", err)
		}

		phash := sha256.Sum256(random)

		privKey, err := secp256k1.GeneratePrivateKey()
		if err != nil {
			return nil, fmt.Errorf("failed to generate private key: %w", err)
		}

		inv, err := zpay32.NewInvoice(&chaincfg.TestNet3Params, phash, tm,
			zpay32.Amount(lnwire.MilliSatoshi(amountMsat)),
			zpay32.Description(descr),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create invoice: %w", err)
		}

		invoiceStr, err := inv.Encode(zpay32.MessageSigner{
			SignCompact: func(msg []byte) ([]byte, error) {
				return ecdsa.SignCompact(privKey, msg, true), nil
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to encode invoice: %w", err)
		}

		return &Invoice{
			Decoded:     inv,
			Encoded:     invoiceStr,
			PaymentHash: hex.EncodeToString(phash[:]),
		}, nil
	}

	invoices := make([]string, 0, nInvoices)
	for i := 0; i < nInvoices; i++ {
		descr := fmt.Sprintf("invoice %d", i)
		inv, err := newDummyInvoice(amountMsat, descr, tm)
		require.NoError(t, err)

		invoices = append(invoices, inv.Encoded)
	}

	type multiPayInvoiceElement struct {
		Invoice string `json:"invoice"`
	}

	type multiPayInvoiceParams struct {
		Invoices []multiPayInvoiceElement `json:"invoices"`
	}

	invoiceElems := make([]multiPayInvoiceElement, 0, len(invoices))
	for _, inv := range invoices {
		invoiceElems = append(invoiceElems, multiPayInvoiceElement{Invoice: inv})
	}

	params := multiPayInvoiceParams{Invoices: invoiceElems}
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)

	nip47Request := &models.Request{
		Method: "multi_pay_invoice",
		Params: paramsJSON,
	}

	responses := []*models.Response{}

	var mu sync.Mutex

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		mu.Lock()
		defer mu.Unlock()
		responses = append(responses, response)
	}

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	NewTestNip47Controller(svc).
		HandleMultiPayInvoiceEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	require.Equal(t, nInvoices, len(responses))

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
