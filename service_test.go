package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

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
const nip47LookupInvoiceJson = `
{
	"method": "lookup_invoice",
	"params": {
		"payment_hash": "4ad9cd27989b514d868e755178378019903a8d78767e3fceb211af9dd00e7a94"
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

const nip47PayJson = `
{
	"method": "pay_invoice",
	"params": {
		"invoice": "lntb1230n1pjypux0pp5xgxzcks5jtx06k784f9dndjh664wc08ucrganpqn52d0ftrh9n8sdqyw3jscqzpgxqyz5vqsp5rkx7cq252p3frx8ytjpzc55rkgyx2mfkzzraa272dqvr2j6leurs9qyyssqhutxa24r5hqxstchz5fxlslawprqjnarjujp5sm3xj7ex73s32sn54fthv2aqlhp76qmvrlvxppx9skd3r5ut5xutgrup8zuc6ay73gqmra29m"
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

const mockInvoice = "lnbc10n1pjdy9aepp5ftvu6fucndg5mp5ww4ghsduqrxgr4rtcwelrln4jzxhem5qw022qhp50kncf9zk35xg4lxewt4974ry6mudygsztsz8qn3ar8pn3mtpe50scqzzsxqyz5vqsp5zyzp3dyn98g7sjlgy4nvujq3rh9xxsagytcyx050mf3rtrx3sn4s9qyyssq7af24ljqf5uzgnz4ualxhvffryh3kpkvvj76ah9yrgdvu494lmfrdf36h5wuhshzemkvrtg2zu70uk0fd9fcmxgl3j9748dvvx9ey9gqpr4jjd"
const mockPaymentHash = "4ad9cd27989b514d868e755178378019903a8d78767e3fceb211af9dd00e7a94" // for the above invoice
var mockNodeInfo = NodeInfo{
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

func TestBackwardsCompatibility(t *testing.T) {
	// Test without adding a single permission
}

// TODO: add tests for HandleEvent method as these
// only cover the cases after the event is processed
// and the corresponding app is found

func TestHandleGetBalanceEvent(t *testing.T) {
	ctx := context.TODO()
	defer os.Remove(testDB)
	mockLn, err := NewMockLn()
	assert.NoError(t, err)
	svc, err := createTestService(mockLn)
	assert.NoError(t, err)
	_, app, ss, err := createUserWithApp(svc)
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

	reqEvent.ID = "test_get_balance_without_permission"
	res, err := svc.HandleGetBalanceEvent(ctx, request, reqEvent, *app, ss)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	response, err := decryptResponse(res, ss, nil)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, NIP_47_ERROR_RESTRICTED, response.Error.Code)

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
	res, err = svc.HandleGetBalanceEvent(ctx, request, reqEvent, *app, ss)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	response, err = decryptResponse(res, ss, &Nip47BalanceResponse{})
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, int64(21000), response.Result.(*Nip47BalanceResponse).Balance)

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
	res, err = svc.HandleGetBalanceEvent(ctx, request, reqEvent, *app, ss)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	response, err = decryptResponse(res, ss, &Nip47BalanceResponse{})
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, int64(21000), response.Result.(*Nip47BalanceResponse).Balance)
	assert.Equal(t, 1000000, response.Result.(*Nip47BalanceResponse).MaxAmount)
	assert.Equal(t, "never", response.Result.(*Nip47BalanceResponse).BudgetRenewal)
}

func TestHandlePayInvoiceEvent(t *testing.T) {
	ctx := context.TODO()
	defer os.Remove(testDB)
	mockLn, err := NewMockLn()
	assert.NoError(t, err)
	svc, err := createTestService(mockLn)
	assert.NoError(t, err)
	_, app, ss, err := createUserWithApp(svc)
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

	reqEvent.ID = "pay_invoice_without_permission"
	res, err := svc.HandlePayInvoiceEvent(ctx, request, reqEvent, *app, ss)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	response, err := decryptResponse(res, ss, nil)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, NIP_47_ERROR_RESTRICTED, response.Error.Code)

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
	res, err = svc.HandlePayInvoiceEvent(ctx, request, reqEvent, *app, ss)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	response, err = decryptResponse(res, ss, &Nip47PayResponse{})
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, response.Result.(*Nip47PayResponse).Preimage, "123preimage")

	// malformed invoice
	err = json.Unmarshal([]byte(nip47PayJsonNoInvoice), request)
	assert.NoError(t, err)

	payload, err = nip04.Encrypt(nip47PayJsonNoInvoice, ss)
	assert.NoError(t, err)
	reqEvent.Content = payload

	reqEvent.ID = "pay_invoice_with_malformed_invoice"
	res, err = svc.HandlePayInvoiceEvent(ctx, request, reqEvent, *app, ss)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	response, err = decryptResponse(res, ss, &Nip47Error{})
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, NIP_47_ERROR_INTERNAL, response.Error.Code)

	// wrong method
	err = json.Unmarshal([]byte(nip47PayWrongMethodJson), request)
	assert.NoError(t, err)

	payload, err = nip04.Encrypt(nip47PayWrongMethodJson, ss)
	assert.NoError(t, err)
	reqEvent.Content = payload

	reqEvent.ID = "pay_invoice_with_wrong_request_method"
	res, err = svc.HandlePayInvoiceEvent(ctx, request, reqEvent, *app, ss)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	response, err = decryptResponse(res, ss, &Nip47Error{})
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, NIP_47_ERROR_RESTRICTED, response.Error.Code)

	// budget overflow
	newMaxAmount := 100
	err = svc.db.Model(&AppPermission{}).Where("app_id = ?", app.ID).Update("max_amount", newMaxAmount).Error

	err = json.Unmarshal([]byte(nip47PayJson), request)
	assert.NoError(t, err)

	payload, err = nip04.Encrypt(nip47PayJson, ss)
	assert.NoError(t, err)
	reqEvent.Content = payload

	reqEvent.ID = "pay_invoice_with_budget_overflow"
	res, err = svc.HandlePayInvoiceEvent(ctx, request, reqEvent, *app, ss)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	response, err = decryptResponse(res, ss, &Nip47Error{})
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, NIP_47_ERROR_QUOTA_EXCEEDED, response.Error.Code)

	// budget expiry
	newExpiry := time.Now().Add(-24 * time.Hour)
	err = svc.db.Model(&AppPermission{}).Where("app_id = ?", app.ID).Update("max_amount", maxAmount).Update("expires_at", newExpiry).Error

	reqEvent.ID = "pay_invoice_with_budget_expiry"
	res, err = svc.HandlePayInvoiceEvent(ctx, request, reqEvent, *app, ss)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	response, err = decryptResponse(res, ss, &Nip47Error{})
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, NIP_47_ERROR_EXPIRED, response.Error.Code)

	// check again
	err = svc.db.Model(&AppPermission{}).Where("app_id = ?", app.ID).Update("expires_at", nil).Error

	reqEvent.ID = "pay_invoice_with_budget_overflow"
	res, err = svc.HandlePayInvoiceEvent(ctx, request, reqEvent, *app, ss)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	response, err = decryptResponse(res, ss, &Nip47PayResponse{})
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, response.Result.(*Nip47PayResponse).Preimage, "123preimage")
}

func TestHandlePayKeysendEvent(t *testing.T) {
	ctx := context.TODO()
	defer os.Remove(testDB)
	mockLn, err := NewMockLn()
	assert.NoError(t, err)
	svc, err := createTestService(mockLn)
	assert.NoError(t, err)
	_, app, ss, err := createUserWithApp(svc)
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

	reqEvent.ID = "pay_keysend_without_permission"
	res, err := svc.HandlePayKeysendEvent(ctx, request, reqEvent, *app, ss)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	response, err := decryptResponse(res, ss, nil)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, NIP_47_ERROR_RESTRICTED, response.Error.Code)

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
	res, err = svc.HandlePayKeysendEvent(ctx, request, reqEvent, *app, ss)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	response, err = decryptResponse(res, ss, &Nip47PayResponse{})
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, response.Result.(*Nip47PayResponse).Preimage, "12345preimage")

	// budget overflow
	newMaxAmount := 100
	err = svc.db.Model(&AppPermission{}).Where("app_id = ?", app.ID).Update("max_amount", newMaxAmount).Error

	err = json.Unmarshal([]byte(nip47KeysendJson), request)
	assert.NoError(t, err)

	payload, err = nip04.Encrypt(nip47KeysendJson, ss)
	assert.NoError(t, err)
	reqEvent.Content = payload

	reqEvent.ID = "pay_keysend_with_budget_overflow"
	res, err = svc.HandlePayKeysendEvent(ctx, request, reqEvent, *app, ss)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	response, err = decryptResponse(res, ss, &Nip47Error{})
	fmt.Println("response")
	fmt.Println(response.Result)
	fmt.Println("response")
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, NIP_47_ERROR_QUOTA_EXCEEDED, response.Error.Code)
}

func TestHandleMakeInvoiceEvent(t *testing.T) {
	ctx := context.TODO()
	defer os.Remove(testDB)
	mockLn, err := NewMockLn()
	assert.NoError(t, err)
	svc, err := createTestService(mockLn)
	assert.NoError(t, err)
	_, app, ss, err := createUserWithApp(svc)
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

	reqEvent.ID = "test_make_invoice_without_permission"
	res, err := svc.HandleMakeInvoiceEvent(ctx, request, reqEvent, *app, ss)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	response, err := decryptResponse(res, ss, nil)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, NIP_47_ERROR_RESTRICTED, response.Error.Code)

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
	res, err = svc.HandleMakeInvoiceEvent(ctx, request, reqEvent, *app, ss)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	response, err = decryptResponse(res, ss, &Nip47MakeInvoiceResponse{})
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, mockTransaction.Preimage, response.Result.(*Nip47MakeInvoiceResponse).Preimage)
}

func TestHandleListTransactionsEvent(t *testing.T) {
	ctx := context.TODO()
	defer os.Remove(testDB)
	mockLn, err := NewMockLn()
	assert.NoError(t, err)
	svc, err := createTestService(mockLn)
	assert.NoError(t, err)
	_, app, ss, err := createUserWithApp(svc)
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

	reqEvent.ID = "test_list_transactions_without_permission"
	res, err := svc.HandleListTransactionsEvent(ctx, request, reqEvent, *app, ss)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	response, err := decryptResponse(res, ss, nil)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, NIP_47_ERROR_RESTRICTED, response.Error.Code)

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
	res, err = svc.HandleListTransactionsEvent(ctx, request, reqEvent, *app, ss)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	response, err = decryptResponse(res, ss, &Nip47ListTransactionsResponse{})
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, 2, len(response.Result.(*Nip47ListTransactionsResponse).Transactions))
	transaction := response.Result.(*Nip47ListTransactionsResponse).Transactions[0]
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
	_, app, ss, err := createUserWithApp(svc)
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

	reqEvent.ID = "test_get_info_without_permission"
	res, err := svc.HandleGetInfoEvent(ctx, request, reqEvent, *app, ss)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	response, err := decryptResponse(res, ss, nil)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, NIP_47_ERROR_RESTRICTED, response.Error.Code)

	// with permission
	err = svc.db.Exec("delete from app_permissions").Error
	assert.NoError(t, err)

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
	res, err = svc.HandleGetInfoEvent(ctx, request, reqEvent, *app, ss)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	response, err = decryptResponse(res, ss, &Nip47GetInfoResponse{})
	assert.NoError(t, err)
	assert.NotNil(t, response)

	assert.Equal(t, mockNodeInfo.Alias, response.Result.(*Nip47GetInfoResponse).Alias)
	assert.Equal(t, mockNodeInfo.Color, response.Result.(*Nip47GetInfoResponse).Color)
	assert.Equal(t, mockNodeInfo.Pubkey, response.Result.(*Nip47GetInfoResponse).Pubkey)
	assert.Equal(t, mockNodeInfo.Network, response.Result.(*Nip47GetInfoResponse).Network)
	assert.Equal(t, mockNodeInfo.BlockHeight, response.Result.(*Nip47GetInfoResponse).BlockHeight)
	assert.Equal(t, mockNodeInfo.BlockHash, response.Result.(*Nip47GetInfoResponse).BlockHash)
	assert.Equal(t, []string{"get_info"}, response.Result.(*Nip47GetInfoResponse).Methods)
}

func createTestService(ln *MockLn) (svc *Service, err error) {
	db, err := gorm.Open(sqlite.Open(testDB), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&User{}, &App{}, &AppPermission{}, &NostrEvent{}, &Payment{}, &Identity{})
	if err != nil {
		return nil, err
	}
	sk := nostr.GeneratePrivateKey()
	pk, err := nostr.GetPublicKey(sk)
	if err != nil {
		return nil, err
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)

	return &Service{
		cfg: &Config{
			NostrSecretKey: sk,
			IdentityPubkey: pk,
		},
		db:          db,
		lnClient:    ln,
		ReceivedEOS: false,
		Logger:      logger,
	}, nil
}

func createUserWithApp(svc *Service) (user *User, app *App, ss []byte, err error) {
	user = &User{ID: 0, AlbyIdentifier: "dummy"}
	err = svc.db.Create(user).Error
	if err != nil {
		return nil, nil, nil, err
	}

	senderPrivkey := nostr.GeneratePrivateKey()
	senderPubkey, err := nostr.GetPublicKey(senderPrivkey)

	ss, err = nip04.ComputeSharedSecret(svc.cfg.IdentityPubkey, senderPrivkey)
	if err != nil {
		return nil, nil, nil, err
	}

	app = &App{Name: "test", NostrPubkey: senderPubkey}
	if err != nil {
		return nil, nil, nil, err
	}
	err = svc.db.Model(&user).Association("Apps").Append(app)
	if err != nil {
		return nil, nil, nil, err
	}

	// creating this permission because if no permissions
	// are created for an app, it can do anything
	appPermission := &AppPermission{
		AppId:         app.ID,
		App:           *app,
		RequestMethod: "UNKNOWN_METHOD",
	}
	err = svc.db.Create(appPermission).Error

	return user, app, ss, nil
}

func decryptResponse(res *nostr.Event, ss []byte, resultType interface{}) (resp *Nip47Response, err error) {
	decrypted, err := nip04.Decrypt(res.Content, ss)
	if err != nil {
		return nil, err
	}
	response := &Nip47Response{
		Result: resultType,
	}
	err = json.Unmarshal([]byte(decrypted), response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

type MockLn struct {
}

func NewMockLn() (*MockLn, error) {
	return &MockLn{}, nil
}

func (mln *MockLn) SendPaymentSync(ctx context.Context, senderPubkey string, payReq string) (preimage string, err error) {
	return "123preimage", nil
}

func (mln *MockLn) SendKeysend(ctx context.Context, senderPubkey string, amount int64, destination, preimage string, custom_records []TLVRecord) (preImage string, err error) {
	return "12345preimage", nil
}

func (mln *MockLn) GetBalance(ctx context.Context, senderPubkey string) (balance int64, err error) {
	return 21, nil
}

func (mln *MockLn) GetInfo(ctx context.Context, senderPubkey string) (info *NodeInfo, err error) {
	return &mockNodeInfo, nil
}

func (mln *MockLn) MakeInvoice(ctx context.Context, senderPubkey string, amount int64, description string, descriptionHash string, expiry int64) (transaction *Nip47Transaction, err error) {
	return mockTransaction, nil
}

func (mln *MockLn) LookupInvoice(ctx context.Context, senderPubkey string, paymentHash string) (transaction *Nip47Transaction, err error) {
	return mockTransaction, nil
}

func (mln *MockLn) ListTransactions(ctx context.Context, senderPubkey string, from, until, limit, offset uint64, unpaid bool, invoiceType string) (invoices []Nip47Transaction, err error) {
	return mockTransactions, nil
}
