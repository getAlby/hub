package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/getAlby/nostr-wallet-connect/events"
	"github.com/getAlby/nostr-wallet-connect/migrations"
	"github.com/getAlby/nostr-wallet-connect/models/config"
	"github.com/getAlby/nostr-wallet-connect/models/lnclient"
	"github.com/glebarez/sqlite"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

const testDB = "test.db"

const nip47GetBalanceJson = `
{
	"method": "get_balance"
}
`

const nip47GetInfoJson = `
{
	"method": "get_info"
}
`

const nip47LookupInvoiceJson = `
{
	"method": "lookup_invoice",
	"params": {
		"payment_hash": "4ad9cd27989b514d868e755178378019903a8d78767e3fceb211af9dd00e7a94"
	}
}
`

const nip47MakeInvoiceJson = `
{
	"method": "make_invoice",
	"params": {
		"amount": 1000,
		"description": "[[\"text/identifier\",\"hello@getalby.com\"],[\"text/plain\",\"Sats for Alby\"]]",
		"expiry": 3600
	}
}
`

const nip47ListTransactionsJson = `
{
	"method": "list_transactions",
	"params": {
		"from": 1693876973,
		"until": 1694876973,
		"limit": 10,
		"offset": 0,
		"type": "incoming"
	}
}
`

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

const nip47PayJson = `
{
	"method": "pay_invoice",
	"params": {
		"invoice": "lntb1230n1pjypux0pp5xgxzcks5jtx06k784f9dndjh664wc08ucrganpqn52d0ftrh9n8sdqyw3jscqzpgxqyz5vqsp5rkx7cq252p3frx8ytjpzc55rkgyx2mfkzzraa272dqvr2j6leurs9qyyssqhutxa24r5hqxstchz5fxlslawprqjnarjujp5sm3xj7ex73s32sn54fthv2aqlhp76qmvrlvxppx9skd3r5ut5xutgrup8zuc6ay73gqmra29m"
	}
}
`

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

const nip47PayWrongMethodJson = `
{
	"method": "get_balance",
	"params": {
		"invoice": "lntb1230n1pjypux0pp5xgxzcks5jtx06k784f9dndjh664wc08ucrganpqn52d0ftrh9n8sdqyw3jscqzpgxqyz5vqsp5rkx7cq252p3frx8ytjpzc55rkgyx2mfkzzraa272dqvr2j6leurs9qyyssqhutxa24r5hqxstchz5fxlslawprqjnarjujp5sm3xj7ex73s32sn54fthv2aqlhp76qmvrlvxppx9skd3r5ut5xutgrup8zuc6ay73gqmra29m"
	}
}
`
const nip47PayJsonNoInvoice = `
{
	"method": "pay_invoice",
	"params": {
		"something": "else"
	}
}
`

// for the invoice:
// lnbcrt5u1pjuywzppp5h69dt59cypca2wxu69sw8ga0g39a3yx7dqug5nthrw3rcqgfdu4qdqqcqzzsxqyz5vqsp5gzlpzszyj2k30qmpme7jsfzr24wqlvt9xdmr7ay34lfelz050krs9qyyssq038x07nh8yuv8hdpjh5y8kqp7zcd62ql9na9xh7pla44htjyy02sz23q7qm2tza6ct4ypljk54w9k9qsrsu95usk8ce726ytep6vhhsq9mhf9a
const mockPaymentHash500 = "be8ad5d0b82071d538dcd160e3a3af444bd890de68388a4d771ba23c01096f2a"

const mockInvoice = "lntb1230n1pjypux0pp5xgxzcks5jtx06k784f9dndjh664wc08ucrganpqn52d0ftrh9n8sdqyw3jscqzpgxqyz5vqsp5rkx7cq252p3frx8ytjpzc55rkgyx2mfkzzraa272dqvr2j6leurs9qyyssqhutxa24r5hqxstchz5fxlslawprqjnarjujp5sm3xj7ex73s32sn54fthv2aqlhp76qmvrlvxppx9skd3r5ut5xutgrup8zuc6ay73gqmra29m"
const mockPaymentHash = "320c2c5a1492ccfd5bc7aa4ad9b657d6aaec3cfcc0d1d98413a29af4ac772ccf" // for the above invoice
var mockNodeInfo = lnclient.NodeInfo{
	Alias:       "bob",
	Color:       "#3399FF",
	Pubkey:      "123pubkey",
	Network:     "testnet",
	BlockHeight: 12,
	BlockHash:   "123blockhash",
}

var mockTime = time.Unix(1693876963, 0)
var mockTimeUnix = mockTime.Unix()

var mockTransactions = []Nip47Transaction{
	{
		Type:            "incoming",
		Invoice:         mockInvoice,
		Description:     "mock invoice 1",
		DescriptionHash: "hash1",
		Preimage:        "preimage1",
		PaymentHash:     "payment_hash_1",
		Amount:          1000,
		FeesPaid:        50,
		SettledAt:       &mockTimeUnix,
		Metadata: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
	},
	{
		Type:            "incoming",
		Invoice:         mockInvoice,
		Description:     "mock invoice 2",
		DescriptionHash: "hash2",
		Preimage:        "preimage2",
		PaymentHash:     "payment_hash_2",
		Amount:          2000,
		FeesPaid:        75,
		SettledAt:       &mockTimeUnix,
	},
}
var mockTransaction = &mockTransactions[0]

// TODO: split each method into separate files (requires moving out of the main package)
// TODO: add E2E tests as well (currently the LNClient and relay are not tested)
// TODO: test a request cannot be processed twice
// TODO: test if an app doesn't exist it returns the right error code
// TODO: test data is stored in the database correctly

func TestHasPermission_NoPermission(t *testing.T) {
	defer os.Remove(testDB)
	mockLn, err := NewMockLn()
	assert.NoError(t, err)
	svc, err := createTestService(mockLn)
	assert.NoError(t, err)

	app, _, err := createApp(svc)
	assert.NoError(t, err)

	result, code, message := svc.hasPermission(app, NIP_47_PAY_INVOICE_METHOD, 100)
	assert.False(t, result)
	assert.Equal(t, NIP_47_ERROR_RESTRICTED, code)
	assert.Equal(t, "This app does not have permission to request pay_invoice", message)
}

func TestHasPermission_Expired(t *testing.T) {
	defer os.Remove(testDB)
	mockLn, err := NewMockLn()
	assert.NoError(t, err)
	svc, err := createTestService(mockLn)
	assert.NoError(t, err)

	app, _, err := createApp(svc)
	assert.NoError(t, err)

	budgetRenewal := "never"
	expiresAt := time.Now().Add(-24 * time.Hour)
	appPermission := &AppPermission{
		AppId:         app.ID,
		App:           *app,
		RequestMethod: NIP_47_PAY_INVOICE_METHOD,
		MaxAmount:     100,
		BudgetRenewal: budgetRenewal,
		ExpiresAt:     expiresAt,
	}
	err = svc.db.Create(appPermission).Error
	assert.NoError(t, err)

	result, code, message := svc.hasPermission(app, NIP_47_PAY_INVOICE_METHOD, 100)
	assert.False(t, result)
	assert.Equal(t, NIP_47_ERROR_EXPIRED, code)
	assert.Equal(t, "This app has expired", message)
}

func TestHasPermission_Exceeded(t *testing.T) {
	defer os.Remove(testDB)
	mockLn, err := NewMockLn()
	assert.NoError(t, err)
	svc, err := createTestService(mockLn)
	assert.NoError(t, err)

	app, _, err := createApp(svc)
	assert.NoError(t, err)

	budgetRenewal := "never"
	expiresAt := time.Now().Add(24 * time.Hour)
	appPermission := &AppPermission{
		AppId:         app.ID,
		App:           *app,
		RequestMethod: NIP_47_PAY_INVOICE_METHOD,
		MaxAmount:     10,
		BudgetRenewal: budgetRenewal,
		ExpiresAt:     expiresAt,
	}
	err = svc.db.Create(appPermission).Error
	assert.NoError(t, err)

	result, code, message := svc.hasPermission(app, NIP_47_PAY_INVOICE_METHOD, 100*1000)
	assert.False(t, result)
	assert.Equal(t, NIP_47_ERROR_QUOTA_EXCEEDED, code)
	assert.Equal(t, "Insufficient budget remaining to make payment", message)
}

func TestHasPermission_OK(t *testing.T) {
	defer os.Remove(testDB)
	mockLn, err := NewMockLn()
	assert.NoError(t, err)
	svc, err := createTestService(mockLn)
	assert.NoError(t, err)

	app, _, err := createApp(svc)
	assert.NoError(t, err)

	budgetRenewal := "never"
	expiresAt := time.Now().Add(24 * time.Hour)
	appPermission := &AppPermission{
		AppId:         app.ID,
		App:           *app,
		RequestMethod: NIP_47_PAY_INVOICE_METHOD,
		MaxAmount:     10,
		BudgetRenewal: budgetRenewal,
		ExpiresAt:     expiresAt,
	}
	err = svc.db.Create(appPermission).Error
	assert.NoError(t, err)

	result, code, message := svc.hasPermission(app, NIP_47_PAY_INVOICE_METHOD, 10*1000)
	assert.True(t, result)
	assert.Empty(t, code)
	assert.Empty(t, message)
}

func TestCreateResponse(t *testing.T) {
	defer os.Remove(testDB)
	mockLn, err := NewMockLn()
	assert.NoError(t, err)
	svc, err := createTestService(mockLn)
	assert.NoError(t, err)

	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)
	assert.NoError(t, err)

	reqEvent := &nostr.Event{
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  reqPubkey,
		Content: "1",
	}

	reqEvent.ID = "12345"

	ss, err := nip04.ComputeSharedSecret(reqPubkey, svc.cfg.NostrSecretKey)
	assert.NoError(t, err)

	nip47Response := &Nip47Response{
		ResultType: NIP_47_GET_BALANCE_METHOD,
		Result: Nip47BalanceResponse{
			Balance: 1000,
		},
	}
	res, err := svc.createResponse(reqEvent, nip47Response, nostr.Tags{}, ss)
	assert.NoError(t, err)
	assert.Equal(t, reqPubkey, res.Tags.GetFirst([]string{"p"}).Value())
	assert.Equal(t, reqEvent.ID, res.Tags.GetFirst([]string{"e"}).Value())
	assert.Equal(t, svc.cfg.NostrPublicKey, res.PubKey)

	decrypted, err := nip04.Decrypt(res.Content, ss)
	assert.NoError(t, err)
	unmarshalledResponse := Nip47Response{
		Result: &Nip47BalanceResponse{},
	}

	err = json.Unmarshal([]byte(decrypted), &unmarshalledResponse)
	assert.NoError(t, err)
	assert.Equal(t, nip47Response.ResultType, unmarshalledResponse.ResultType)
	assert.Equal(t, nip47Response.Result, *unmarshalledResponse.Result.(*Nip47BalanceResponse))
}

func TestHandleEncryption(t *testing.T) {}

func TestHandleMultiPayInvoiceEvent(t *testing.T) {
	ctx := context.TODO()
	defer os.Remove(testDB)
	mockLn, err := NewMockLn()
	assert.NoError(t, err)
	svc, err := createTestService(mockLn)
	assert.NoError(t, err)
	app, ss, err := createApp(svc)
	assert.NoError(t, err)

	request := &Nip47Request{}
	err = json.Unmarshal([]byte(nip47MultiPayJson), request)
	assert.NoError(t, err)

	// without permission
	payload, err := nip04.Encrypt(nip47MultiPayJson, ss)
	assert.NoError(t, err)
	reqEvent := &nostr.Event{
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  app.NostrPubkey,
		Content: payload,
	}
	requestEvent := &RequestEvent{
		Content: reqEvent.Content,
	}

	reqEvent.ID = "multi_pay_invoice_without_permission"
	requestEvent.NostrId = reqEvent.ID

	responses := []*Nip47Response{}
	dTags := []nostr.Tags{}

	publishResponse := func(response *Nip47Response, tags nostr.Tags) {
		responses = append(responses, response)
		dTags = append(dTags, tags)
	}

	svc.HandleMultiPayInvoiceEvent(ctx, request, requestEvent, app, publishResponse)

	assert.Equal(t, 2, len(responses))
	assert.Equal(t, 2, len(dTags))
	for i := 0; i < len(responses); i++ {
		assert.Equal(t, NIP_47_ERROR_RESTRICTED, responses[i].Error.Code)
		assert.Equal(t, mockPaymentHash, dTags[i].GetFirst([]string{"d"}).Value())
	}

	// with permission
	maxAmount := 1000
	budgetRenewal := "never"
	expiresAt := time.Now().Add(24 * time.Hour)
	appPermission := &AppPermission{
		AppId:         app.ID,
		App:           *app,
		RequestMethod: NIP_47_PAY_INVOICE_METHOD,
		MaxAmount:     maxAmount,
		BudgetRenewal: budgetRenewal,
		ExpiresAt:     expiresAt,
	}
	err = svc.db.Create(appPermission).Error
	assert.NoError(t, err)

	reqEvent.ID = "multi_pay_invoice_with_permission"
	requestEvent.NostrId = reqEvent.ID
	responses = []*Nip47Response{}
	dTags = []nostr.Tags{}
	svc.HandleMultiPayInvoiceEvent(ctx, request, requestEvent, app, publishResponse)

	assert.Equal(t, 2, len(responses))
	for i := 0; i < len(responses); i++ {
		assert.Equal(t, responses[i].Result.(Nip47PayResponse).Preimage, "123preimage")
		assert.Equal(t, mockPaymentHash, dTags[i].GetFirst([]string{"d"}).Value())
	}

	// one malformed invoice
	err = json.Unmarshal([]byte(nip47MultiPayOneMalformedInvoiceJson), request)
	assert.NoError(t, err)

	payload, err = nip04.Encrypt(nip47MultiPayOneMalformedInvoiceJson, ss)
	assert.NoError(t, err)
	reqEvent.Content = payload

	reqEvent.ID = "multi_pay_invoice_with_one_malformed_invoice"
	requestEvent.NostrId = reqEvent.ID
	responses = []*Nip47Response{}
	dTags = []nostr.Tags{}
	svc.HandleMultiPayInvoiceEvent(ctx, request, requestEvent, app, publishResponse)

	assert.Equal(t, 2, len(responses))
	assert.Equal(t, "invoiceId123", dTags[0].GetFirst([]string{"d"}).Value())
	assert.Equal(t, responses[0].Error.Code, NIP_47_ERROR_INTERNAL)

	assert.Equal(t, mockPaymentHash, dTags[1].GetFirst([]string{"d"}).Value())
	assert.Equal(t, responses[1].Result.(Nip47PayResponse).Preimage, "123preimage")

	// we've spent 369 till here in three payments

	// budget overflow
	newMaxAmount := 500
	err = svc.db.Model(&AppPermission{}).Where("app_id = ?", app.ID).Update("max_amount", newMaxAmount).Error
	assert.NoError(t, err)

	err = json.Unmarshal([]byte(nip47MultiPayOneOverflowingBudgetJson), request)
	assert.NoError(t, err)

	payload, err = nip04.Encrypt(nip47MultiPayOneOverflowingBudgetJson, ss)
	assert.NoError(t, err)
	reqEvent.Content = payload

	reqEvent.ID = "multi_pay_invoice_with_budget_overflow"
	requestEvent.NostrId = reqEvent.ID
	responses = []*Nip47Response{}
	dTags = []nostr.Tags{}
	svc.HandleMultiPayInvoiceEvent(ctx, request, requestEvent, app, publishResponse)

	// might be flaky because the two requests run concurrently
	// and there's more chance that the failed respons calls the
	// publishResponse as it's called earlier
	assert.Equal(t, responses[0].Error.Code, NIP_47_ERROR_QUOTA_EXCEEDED)
	assert.Equal(t, mockPaymentHash500, dTags[0].GetFirst([]string{"d"}).Value())
	assert.Equal(t, responses[1].Result.(Nip47PayResponse).Preimage, "123preimage")
	assert.Equal(t, mockPaymentHash, dTags[1].GetFirst([]string{"d"}).Value())
}

func TestHandleMultiPayKeysendEvent(t *testing.T) {

	ctx := context.TODO()
	defer os.Remove(testDB)
	mockLn, err := NewMockLn()
	assert.NoError(t, err)
	svc, err := createTestService(mockLn)
	assert.NoError(t, err)
	app, ss, err := createApp(svc)
	assert.NoError(t, err)

	request := &Nip47Request{}
	err = json.Unmarshal([]byte(nip47MultiPayKeysendJson), request)
	assert.NoError(t, err)

	// without permission
	payload, err := nip04.Encrypt(nip47MultiPayKeysendJson, ss)
	assert.NoError(t, err)
	reqEvent := &nostr.Event{
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  app.NostrPubkey,
		Content: payload,
	}
	requestEvent := &RequestEvent{
		Content: reqEvent.Content,
	}

	reqEvent.ID = "multi_pay_keysend_without_permission"
	requestEvent.NostrId = reqEvent.ID

	responses := []*Nip47Response{}
	dTags := []nostr.Tags{}

	publishResponse := func(response *Nip47Response, tags nostr.Tags) {
		responses = append(responses, response)
		dTags = append(dTags, tags)
	}

	svc.HandleMultiPayKeysendEvent(ctx, request, requestEvent, app, publishResponse)

	assert.Equal(t, 2, len(responses))
	for i := 0; i < len(responses); i++ {
		assert.Equal(t, NIP_47_ERROR_RESTRICTED, responses[i].Error.Code)
	}

	// with permission
	maxAmount := 1000
	budgetRenewal := "never"
	expiresAt := time.Now().Add(24 * time.Hour)
	// because we need the same permission for keysend although
	// it works even with NIP_47_PAY_KEYSEND_METHOD, see
	// https://github.com/getAlby/nostr-wallet-connect/issues/189
	appPermission := &AppPermission{
		AppId:         app.ID,
		App:           *app,
		RequestMethod: NIP_47_PAY_INVOICE_METHOD,
		MaxAmount:     maxAmount,
		BudgetRenewal: budgetRenewal,
		ExpiresAt:     expiresAt,
	}
	err = svc.db.Create(appPermission).Error
	assert.NoError(t, err)

	reqEvent.ID = "multi_pay_keysend_with_permission"
	requestEvent.NostrId = reqEvent.ID
	responses = []*Nip47Response{}
	dTags = []nostr.Tags{}
	svc.HandleMultiPayKeysendEvent(ctx, request, requestEvent, app, publishResponse)

	assert.Equal(t, 2, len(responses))
	for i := 0; i < len(responses); i++ {
		assert.Equal(t, responses[i].Result.(Nip47PayResponse).Preimage, "12345preimage")
		assert.Equal(t, "123pubkey", dTags[i].GetFirst([]string{"d"}).Value())
	}

	// we've spent 246 till here in two payments

	// budget overflow
	newMaxAmount := 500
	err = svc.db.Model(&AppPermission{}).Where("app_id = ?", app.ID).Update("max_amount", newMaxAmount).Error
	assert.NoError(t, err)

	err = json.Unmarshal([]byte(nip47MultiPayKeysendOneOverflowingBudgetJson), request)
	assert.NoError(t, err)

	payload, err = nip04.Encrypt(nip47MultiPayKeysendOneOverflowingBudgetJson, ss)
	assert.NoError(t, err)
	reqEvent.Content = payload

	reqEvent.ID = "multi_pay_keysend_with_budget_overflow"
	requestEvent.NostrId = reqEvent.ID
	responses = []*Nip47Response{}
	dTags = []nostr.Tags{}
	svc.HandleMultiPayKeysendEvent(ctx, request, requestEvent, app, publishResponse)

	assert.Equal(t, responses[0].Error.Code, NIP_47_ERROR_QUOTA_EXCEEDED)
	assert.Equal(t, "500pubkey", dTags[0].GetFirst([]string{"d"}).Value())
	assert.Equal(t, responses[1].Result.(Nip47PayResponse).Preimage, "12345preimage")
	assert.Equal(t, "customId", dTags[1].GetFirst([]string{"d"}).Value())
}

func TestHandleGetBalanceEvent(t *testing.T) {
	ctx := context.TODO()
	defer os.Remove(testDB)
	mockLn, err := NewMockLn()
	assert.NoError(t, err)
	svc, err := createTestService(mockLn)
	assert.NoError(t, err)
	app, ss, err := createApp(svc)
	assert.NoError(t, err)

	request := &Nip47Request{}
	err = json.Unmarshal([]byte(nip47GetBalanceJson), request)
	assert.NoError(t, err)

	// without permission
	payload, err := nip04.Encrypt(nip47GetBalanceJson, ss)
	assert.NoError(t, err)
	reqEvent := &nostr.Event{
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  app.NostrPubkey,
		Content: payload,
	}
	requestEvent := &RequestEvent{
		Content: reqEvent.Content,
	}

	reqEvent.ID = "test_get_balance_without_permission"
	requestEvent.NostrId = reqEvent.ID

	responses := []*Nip47Response{}

	publishResponse := func(response *Nip47Response, tags nostr.Tags) {
		responses = append(responses, response)
	}

	err = svc.db.Create(&requestEvent).Error
	assert.NoError(t, err)

	svc.HandleGetBalanceEvent(ctx, request, requestEvent, app, publishResponse)

	assert.Equal(t, responses[0].Error.Code, NIP_47_ERROR_RESTRICTED)

	// with permission
	expiresAt := time.Now().Add(24 * time.Hour)
	appPermission := &AppPermission{
		AppId:         app.ID,
		App:           *app,
		RequestMethod: NIP_47_GET_BALANCE_METHOD,
		ExpiresAt:     expiresAt,
	}
	err = svc.db.Create(appPermission).Error
	assert.NoError(t, err)

	reqEvent.ID = "test_get_balance_with_permission"
	requestEvent.NostrId = reqEvent.ID
	responses = []*Nip47Response{}
	svc.HandleGetBalanceEvent(ctx, request, requestEvent, app, publishResponse)

	assert.Equal(t, responses[0].Result.(*Nip47BalanceResponse).Balance, int64(21000))

	// create pay_invoice permission
	maxAmount := 1000
	budgetRenewal := "never"
	appPermission = &AppPermission{
		AppId:         app.ID,
		App:           *app,
		RequestMethod: NIP_47_PAY_INVOICE_METHOD,
		MaxAmount:     maxAmount,
		BudgetRenewal: budgetRenewal,
		ExpiresAt:     expiresAt,
	}
	err = svc.db.Create(appPermission).Error
	assert.NoError(t, err)

	reqEvent.ID = "test_get_balance_with_budget"
	responses = []*Nip47Response{}
	svc.HandleGetBalanceEvent(ctx, request, requestEvent, app, publishResponse)

	assert.Equal(t, int64(21000), responses[0].Result.(*Nip47BalanceResponse).Balance)
	assert.Equal(t, 1000000, responses[0].Result.(*Nip47BalanceResponse).MaxAmount)
	assert.Equal(t, "never", responses[0].Result.(*Nip47BalanceResponse).BudgetRenewal)
}

func TestHandlePayInvoiceEvent(t *testing.T) {
	ctx := context.TODO()
	defer os.Remove(testDB)
	mockLn, err := NewMockLn()
	assert.NoError(t, err)
	svc, err := createTestService(mockLn)
	assert.NoError(t, err)
	app, ss, err := createApp(svc)
	assert.NoError(t, err)

	request := &Nip47Request{}
	err = json.Unmarshal([]byte(nip47PayJson), request)
	assert.NoError(t, err)

	// without permission
	payload, err := nip04.Encrypt(nip47PayJson, ss)
	assert.NoError(t, err)
	reqEvent := &nostr.Event{
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  app.NostrPubkey,
		Content: payload,
	}
	requestEvent := &RequestEvent{
		Content: reqEvent.Content,
	}

	reqEvent.ID = "pay_invoice_without_permission"
	requestEvent.NostrId = reqEvent.ID

	responses := []*Nip47Response{}

	publishResponse := func(response *Nip47Response, tags nostr.Tags) {
		responses = append(responses, response)
	}

	svc.HandlePayInvoiceEvent(ctx, request, requestEvent, app, publishResponse)

	assert.Equal(t, NIP_47_ERROR_RESTRICTED, responses[0].Error.Code)

	// with permission
	maxAmount := 1000
	budgetRenewal := "never"
	expiresAt := time.Now().Add(24 * time.Hour)
	appPermission := &AppPermission{
		AppId:         app.ID,
		App:           *app,
		RequestMethod: NIP_47_PAY_INVOICE_METHOD,
		MaxAmount:     maxAmount,
		BudgetRenewal: budgetRenewal,
		ExpiresAt:     expiresAt,
	}
	err = svc.db.Create(appPermission).Error
	assert.NoError(t, err)

	reqEvent.ID = "pay_invoice_with_permission"
	requestEvent.NostrId = reqEvent.ID
	responses = []*Nip47Response{}
	svc.HandlePayInvoiceEvent(ctx, request, requestEvent, app, publishResponse)

	assert.Equal(t, responses[0].Result.(Nip47PayResponse).Preimage, "123preimage")

	// malformed invoice
	err = json.Unmarshal([]byte(nip47PayJsonNoInvoice), request)
	assert.NoError(t, err)

	payload, err = nip04.Encrypt(nip47PayJsonNoInvoice, ss)
	assert.NoError(t, err)
	reqEvent.Content = payload

	reqEvent.ID = "pay_invoice_with_malformed_invoice"
	requestEvent.NostrId = reqEvent.ID
	responses = []*Nip47Response{}
	svc.HandlePayInvoiceEvent(ctx, request, requestEvent, app, publishResponse)

	assert.Equal(t, NIP_47_ERROR_INTERNAL, responses[0].Error.Code)

	// wrong method
	err = json.Unmarshal([]byte(nip47PayWrongMethodJson), request)
	assert.NoError(t, err)

	payload, err = nip04.Encrypt(nip47PayWrongMethodJson, ss)
	assert.NoError(t, err)
	reqEvent.Content = payload

	reqEvent.ID = "pay_invoice_with_wrong_request_method"
	requestEvent.NostrId = reqEvent.ID
	responses = []*Nip47Response{}
	svc.HandlePayInvoiceEvent(ctx, request, requestEvent, app, publishResponse)

	assert.Equal(t, NIP_47_ERROR_RESTRICTED, responses[0].Error.Code)

	// budget overflow
	newMaxAmount := 100
	err = svc.db.Model(&AppPermission{}).Where("app_id = ?", app.ID).Update("max_amount", newMaxAmount).Error
	assert.NoError(t, err)

	err = json.Unmarshal([]byte(nip47PayJson), request)
	assert.NoError(t, err)

	payload, err = nip04.Encrypt(nip47PayJson, ss)
	assert.NoError(t, err)
	reqEvent.Content = payload

	reqEvent.ID = "pay_invoice_with_budget_overflow"
	requestEvent.NostrId = reqEvent.ID
	responses = []*Nip47Response{}
	svc.HandlePayInvoiceEvent(ctx, request, requestEvent, app, publishResponse)

	assert.Equal(t, NIP_47_ERROR_QUOTA_EXCEEDED, responses[0].Error.Code)

	// budget expiry
	newExpiry := time.Now().Add(-24 * time.Hour)
	err = svc.db.Model(&AppPermission{}).Where("app_id = ?", app.ID).Update("max_amount", maxAmount).Update("expires_at", newExpiry).Error
	assert.NoError(t, err)

	reqEvent.ID = "pay_invoice_with_budget_expiry"
	requestEvent.NostrId = reqEvent.ID
	responses = []*Nip47Response{}
	svc.HandlePayInvoiceEvent(ctx, request, requestEvent, app, publishResponse)

	assert.Equal(t, NIP_47_ERROR_EXPIRED, responses[0].Error.Code)

	// check again
	err = svc.db.Model(&AppPermission{}).Where("app_id = ?", app.ID).Update("expires_at", nil).Error
	assert.NoError(t, err)

	reqEvent.ID = "pay_invoice_after_change"
	requestEvent.NostrId = reqEvent.ID
	responses = []*Nip47Response{}
	svc.HandlePayInvoiceEvent(ctx, request, requestEvent, app, publishResponse)

	assert.Equal(t, responses[0].Result.(Nip47PayResponse).Preimage, "123preimage")
}

func TestHandlePayKeysendEvent(t *testing.T) {
	ctx := context.TODO()
	defer os.Remove(testDB)
	mockLn, err := NewMockLn()
	assert.NoError(t, err)
	svc, err := createTestService(mockLn)
	assert.NoError(t, err)
	app, ss, err := createApp(svc)
	assert.NoError(t, err)

	request := &Nip47Request{}
	err = json.Unmarshal([]byte(nip47KeysendJson), request)
	assert.NoError(t, err)

	// without permission
	payload, err := nip04.Encrypt(nip47KeysendJson, ss)
	assert.NoError(t, err)
	reqEvent := &nostr.Event{
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  app.NostrPubkey,
		Content: payload,
	}
	requestEvent := &RequestEvent{
		Content: reqEvent.Content,
	}

	reqEvent.ID = "pay_keysend_without_permission"
	requestEvent.NostrId = reqEvent.ID

	responses := []*Nip47Response{}

	publishResponse := func(response *Nip47Response, tags nostr.Tags) {
		responses = append(responses, response)
	}

	svc.HandlePayKeysendEvent(ctx, request, requestEvent, app, publishResponse)

	assert.Equal(t, NIP_47_ERROR_RESTRICTED, responses[0].Error.Code)

	// with permission
	maxAmount := 1000
	budgetRenewal := "never"
	expiresAt := time.Now().Add(24 * time.Hour)
	// because we need the same permission for keysend although
	// it works even with NIP_47_PAY_KEYSEND_METHOD, see
	// https://github.com/getAlby/nostr-wallet-connect/issues/189
	appPermission := &AppPermission{
		AppId:         app.ID,
		App:           *app,
		RequestMethod: NIP_47_PAY_INVOICE_METHOD,
		MaxAmount:     maxAmount,
		BudgetRenewal: budgetRenewal,
		ExpiresAt:     expiresAt,
	}
	err = svc.db.Create(appPermission).Error
	assert.NoError(t, err)

	reqEvent.ID = "pay_keysend_with_permission"
	requestEvent.NostrId = reqEvent.ID
	responses = []*Nip47Response{}
	svc.HandlePayKeysendEvent(ctx, request, requestEvent, app, publishResponse)

	assert.Equal(t, responses[0].Result.(Nip47PayResponse).Preimage, "12345preimage")

	// budget overflow
	newMaxAmount := 100
	err = svc.db.Model(&AppPermission{}).Where("app_id = ?", app.ID).Update("max_amount", newMaxAmount).Error
	assert.NoError(t, err)

	err = json.Unmarshal([]byte(nip47KeysendJson), request)
	assert.NoError(t, err)

	payload, err = nip04.Encrypt(nip47KeysendJson, ss)
	assert.NoError(t, err)
	reqEvent.Content = payload

	reqEvent.ID = "pay_keysend_with_budget_overflow"
	requestEvent.NostrId = reqEvent.ID
	responses = []*Nip47Response{}
	svc.HandlePayKeysendEvent(ctx, request, requestEvent, app, publishResponse)

	assert.Equal(t, NIP_47_ERROR_QUOTA_EXCEEDED, responses[0].Error.Code)
}

func TestHandleLookupInvoiceEvent(t *testing.T) {
	ctx := context.TODO()
	defer os.Remove(testDB)
	mockLn, err := NewMockLn()
	assert.NoError(t, err)
	svc, err := createTestService(mockLn)
	assert.NoError(t, err)
	app, ss, err := createApp(svc)
	assert.NoError(t, err)

	request := &Nip47Request{}
	err = json.Unmarshal([]byte(nip47LookupInvoiceJson), request)
	assert.NoError(t, err)

	// without permission
	payload, err := nip04.Encrypt(nip47LookupInvoiceJson, ss)
	assert.NoError(t, err)
	reqEvent := &nostr.Event{
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  app.NostrPubkey,
		Content: payload,
	}
	requestEvent := &RequestEvent{
		Content: reqEvent.Content,
	}

	reqEvent.ID = "test_lookup_invoice_without_permission"
	requestEvent.NostrId = reqEvent.ID

	responses := []*Nip47Response{}

	publishResponse := func(response *Nip47Response, tags nostr.Tags) {
		responses = append(responses, response)
	}

	svc.HandleLookupInvoiceEvent(ctx, request, requestEvent, app, publishResponse)

	assert.Equal(t, NIP_47_ERROR_RESTRICTED, responses[0].Error.Code)

	// with permission
	expiresAt := time.Now().Add(24 * time.Hour)
	appPermission := &AppPermission{
		AppId:         app.ID,
		App:           *app,
		RequestMethod: NIP_47_LOOKUP_INVOICE_METHOD,
		ExpiresAt:     expiresAt,
	}
	err = svc.db.Create(appPermission).Error
	assert.NoError(t, err)

	reqEvent.ID = "test_lookup_invoice_with_permission"
	requestEvent.NostrId = reqEvent.ID
	responses = []*Nip47Response{}
	svc.HandleLookupInvoiceEvent(ctx, request, requestEvent, app, publishResponse)

	transaction := responses[0].Result.(*Nip47LookupInvoiceResponse)
	assert.Equal(t, mockTransaction.Type, transaction.Type)
	assert.Equal(t, mockTransaction.Invoice, transaction.Invoice)
	assert.Equal(t, mockTransaction.Description, transaction.Description)
	assert.Equal(t, mockTransaction.DescriptionHash, transaction.DescriptionHash)
	assert.Equal(t, mockTransaction.Preimage, transaction.Preimage)
	assert.Equal(t, mockTransaction.PaymentHash, transaction.PaymentHash)
	assert.Equal(t, mockTransaction.Amount, transaction.Amount)
	assert.Equal(t, mockTransaction.FeesPaid, transaction.FeesPaid)
	assert.Equal(t, mockTransaction.SettledAt, transaction.SettledAt)
}

func TestHandleMakeInvoiceEvent(t *testing.T) {
	ctx := context.TODO()
	defer os.Remove(testDB)
	mockLn, err := NewMockLn()
	assert.NoError(t, err)
	svc, err := createTestService(mockLn)
	assert.NoError(t, err)
	app, ss, err := createApp(svc)
	assert.NoError(t, err)

	request := &Nip47Request{}
	err = json.Unmarshal([]byte(nip47MakeInvoiceJson), request)
	assert.NoError(t, err)

	// without permission
	payload, err := nip04.Encrypt(nip47MakeInvoiceJson, ss)
	assert.NoError(t, err)
	reqEvent := &nostr.Event{
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  app.NostrPubkey,
		Content: payload,
	}
	requestEvent := &RequestEvent{
		Content: reqEvent.Content,
	}

	reqEvent.ID = "test_make_invoice_without_permission"
	requestEvent.NostrId = reqEvent.ID

	responses := []*Nip47Response{}

	publishResponse := func(response *Nip47Response, tags nostr.Tags) {
		responses = append(responses, response)
	}

	svc.HandleMakeInvoiceEvent(ctx, request, requestEvent, app, publishResponse)

	assert.Equal(t, NIP_47_ERROR_RESTRICTED, responses[0].Error.Code)

	// with permission
	expiresAt := time.Now().Add(24 * time.Hour)
	appPermission := &AppPermission{
		AppId:         app.ID,
		App:           *app,
		RequestMethod: NIP_47_MAKE_INVOICE_METHOD,
		ExpiresAt:     expiresAt,
	}
	err = svc.db.Create(appPermission).Error
	assert.NoError(t, err)

	reqEvent.ID = "test_make_invoice_with_permission"
	requestEvent.NostrId = reqEvent.ID
	responses = []*Nip47Response{}
	svc.HandleMakeInvoiceEvent(ctx, request, requestEvent, app, publishResponse)

	assert.Equal(t, mockTransaction.Preimage, responses[0].Result.(*Nip47MakeInvoiceResponse).Preimage)
}

func TestHandleListTransactionsEvent(t *testing.T) {
	ctx := context.TODO()
	defer os.Remove(testDB)
	mockLn, err := NewMockLn()
	assert.NoError(t, err)
	svc, err := createTestService(mockLn)
	assert.NoError(t, err)
	app, ss, err := createApp(svc)
	assert.NoError(t, err)

	request := &Nip47Request{}
	err = json.Unmarshal([]byte(nip47ListTransactionsJson), request)
	assert.NoError(t, err)

	// without permission
	payload, err := nip04.Encrypt(nip47ListTransactionsJson, ss)
	assert.NoError(t, err)
	reqEvent := &nostr.Event{
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  app.NostrPubkey,
		Content: payload,
	}
	requestEvent := &RequestEvent{
		Content: reqEvent.Content,
	}

	reqEvent.ID = "test_list_transactions_without_permission"
	requestEvent.NostrId = reqEvent.ID

	responses := []*Nip47Response{}

	publishResponse := func(response *Nip47Response, tags nostr.Tags) {
		responses = append(responses, response)
	}

	svc.HandleListTransactionsEvent(ctx, request, requestEvent, app, publishResponse)

	assert.Equal(t, NIP_47_ERROR_RESTRICTED, responses[0].Error.Code)

	// with permission
	expiresAt := time.Now().Add(24 * time.Hour)
	appPermission := &AppPermission{
		AppId:         app.ID,
		App:           *app,
		RequestMethod: NIP_47_LIST_TRANSACTIONS_METHOD,
		ExpiresAt:     expiresAt,
	}
	err = svc.db.Create(appPermission).Error
	assert.NoError(t, err)

	reqEvent.ID = "test_list_transactions_with_permission"
	requestEvent.NostrId = reqEvent.ID
	responses = []*Nip47Response{}
	svc.HandleListTransactionsEvent(ctx, request, requestEvent, app, publishResponse)

	assert.Equal(t, 2, len(responses[0].Result.(*Nip47ListTransactionsResponse).Transactions))
	transaction := responses[0].Result.(*Nip47ListTransactionsResponse).Transactions[0]
	assert.Equal(t, mockTransactions[0].Type, transaction.Type)
	assert.Equal(t, mockTransactions[0].Invoice, transaction.Invoice)
	assert.Equal(t, mockTransactions[0].Description, transaction.Description)
	assert.Equal(t, mockTransactions[0].DescriptionHash, transaction.DescriptionHash)
	assert.Equal(t, mockTransactions[0].Preimage, transaction.Preimage)
	assert.Equal(t, mockTransactions[0].PaymentHash, transaction.PaymentHash)
	assert.Equal(t, mockTransactions[0].Amount, transaction.Amount)
	assert.Equal(t, mockTransactions[0].FeesPaid, transaction.FeesPaid)
	assert.Equal(t, mockTransactions[0].SettledAt, transaction.SettledAt)
}

func TestHandleGetInfoEvent(t *testing.T) {
	ctx := context.TODO()
	defer os.Remove(testDB)
	mockLn, err := NewMockLn()
	assert.NoError(t, err)
	svc, err := createTestService(mockLn)
	assert.NoError(t, err)
	app, ss, err := createApp(svc)
	assert.NoError(t, err)

	request := &Nip47Request{}
	err = json.Unmarshal([]byte(nip47GetInfoJson), request)
	assert.NoError(t, err)

	// without permission
	payload, err := nip04.Encrypt(nip47GetInfoJson, ss)
	assert.NoError(t, err)
	reqEvent := &nostr.Event{
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  app.NostrPubkey,
		Content: payload,
	}
	requestEvent := &RequestEvent{
		Content: reqEvent.Content,
	}

	reqEvent.ID = "test_get_info_without_permission"
	requestEvent.NostrId = reqEvent.ID

	responses := []*Nip47Response{}

	publishResponse := func(response *Nip47Response, tags nostr.Tags) {
		responses = append(responses, response)
	}

	svc.HandleGetInfoEvent(ctx, request, requestEvent, app, publishResponse)

	assert.Equal(t, NIP_47_ERROR_RESTRICTED, responses[0].Error.Code)

	expiresAt := time.Now().Add(24 * time.Hour)
	appPermission := &AppPermission{
		AppId:         app.ID,
		App:           *app,
		RequestMethod: NIP_47_GET_INFO_METHOD,
		ExpiresAt:     expiresAt,
	}
	err = svc.db.Create(appPermission).Error
	assert.NoError(t, err)

	reqEvent.ID = "test_get_info_with_permission"
	requestEvent.NostrId = reqEvent.ID
	responses = []*Nip47Response{}
	svc.HandleGetInfoEvent(ctx, request, requestEvent, app, publishResponse)

	nodeInfo := responses[0].Result.(*Nip47GetInfoResponse)
	assert.Equal(t, mockNodeInfo.Alias, nodeInfo.Alias)
	assert.Equal(t, mockNodeInfo.Color, nodeInfo.Color)
	assert.Equal(t, mockNodeInfo.Pubkey, nodeInfo.Pubkey)
	assert.Equal(t, mockNodeInfo.Network, nodeInfo.Network)
	assert.Equal(t, mockNodeInfo.BlockHeight, nodeInfo.BlockHeight)
	assert.Equal(t, mockNodeInfo.BlockHash, nodeInfo.BlockHash)
	assert.Equal(t, []string{"get_info"}, nodeInfo.Methods)
}

func createTestService(ln *MockLn) (svc *Service, err error) {
	gormDb, err := gorm.Open(sqlite.Open(testDB), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	err = migrations.Migrate(gormDb, &config.AppConfig{
		Workdir: ".test",
	}, logger)
	if err != nil {
		return nil, err
	}
	sk := nostr.GeneratePrivateKey()
	pk, err := nostr.GetPublicKey(sk)
	if err != nil {
		return nil, err
	}

	return &Service{
		cfg: &Config{
			db:             gormDb,
			NostrSecretKey: sk,
			NostrPublicKey: pk,
		},
		db:          gormDb,
		lnClient:    ln,
		Logger:      logger,
		EventLogger: events.NewEventLogger(logger, false),
	}, nil
}

func createApp(svc *Service) (app *App, ss []byte, err error) {
	senderPrivkey := nostr.GeneratePrivateKey()
	senderPubkey, err := nostr.GetPublicKey(senderPrivkey)
	if err != nil {
		return nil, nil, err
	}

	ss, err = nip04.ComputeSharedSecret(svc.cfg.NostrPublicKey, senderPrivkey)
	if err != nil {
		return nil, nil, err
	}

	app = &App{Name: "test", NostrPubkey: senderPubkey}
	err = svc.db.Create(app).Error
	if err != nil {
		return nil, nil, err
	}

	return app, ss, nil
}

type MockLn struct {
}

func NewMockLn() (*MockLn, error) {
	return &MockLn{}, nil
}

func (mln *MockLn) SendPaymentSync(ctx context.Context, payReq string) (preimage string, err error) {
	return "123preimage", nil
}

func (mln *MockLn) SendKeysend(ctx context.Context, amount int64, destination, preimage string, custom_records []lnclient.TLVRecord) (preImage string, err error) {
	return "12345preimage", nil
}

func (mln *MockLn) GetBalance(ctx context.Context) (balance int64, err error) {
	return 21000, nil
}

func (mln *MockLn) GetInfo(ctx context.Context) (info *lnclient.NodeInfo, err error) {
	return &mockNodeInfo, nil
}

func (mln *MockLn) MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64) (transaction *Nip47Transaction, err error) {
	return mockTransaction, nil
}

func (mln *MockLn) LookupInvoice(ctx context.Context, paymentHash string) (transaction *Nip47Transaction, err error) {
	return mockTransaction, nil
}

func (mln *MockLn) ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaid bool, invoiceType string) (invoices []Nip47Transaction, err error) {
	return mockTransactions, nil
}
func (mln *MockLn) Shutdown() error {
	return nil
}

func (mln *MockLn) ListChannels(ctx context.Context) (channels []lnclient.Channel, err error) {
	return []lnclient.Channel{}, nil
}
func (mln *MockLn) GetNodeConnectionInfo(ctx context.Context) (nodeConnectionInfo *lnclient.NodeConnectionInfo, err error) {
	return nil, nil
}
func (mln *MockLn) ConnectPeer(ctx context.Context, connectPeerRequest *lnclient.ConnectPeerRequest) error {
	return nil
}
func (mln *MockLn) OpenChannel(ctx context.Context, openChannelRequest *lnclient.OpenChannelRequest) (*lnclient.OpenChannelResponse, error) {
	return nil, nil
}
func (mln *MockLn) CloseChannel(ctx context.Context, closeChannelRequest *lnclient.CloseChannelRequest) (*lnclient.CloseChannelResponse, error) {
	return nil, nil
}
func (mln *MockLn) GetNewOnchainAddress(ctx context.Context) (string, error) {
	return "", nil
}
func (mln *MockLn) GetOnchainBalance(ctx context.Context) (*lnclient.OnchainBalanceResponse, error) {
	return nil, nil
}
func (mln *MockLn) RedeemOnchainFunds(ctx context.Context, toAddress string) (txId string, err error) {
	return "", nil
}
func (mln *MockLn) ResetRouter(ctx context.Context) error {
	return nil
}
func (mln *MockLn) SignMessage(ctx context.Context, message string) (string, error) {
	return "", nil
}
