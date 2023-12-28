package main

import (
	"context"
	"encoding/json"
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
		"amount": 100,
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
		SettledAt:       &mockTime,
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
		SettledAt:       &mockTime,
	},
}
var mockTransaction = &mockTransactions[0]

// TODO: split up into individual tests
func TestHandleEvent(t *testing.T) {
	ctx := context.TODO()
	svc, _ := createTestService(t)
	defer os.Remove(testDB)
	//test not yet receivedEOS
	res, err := svc.HandleEvent(ctx, &nostr.Event{
		Kind: NIP_47_REQUEST_KIND,
	})
	assert.Nil(t, res)
	assert.Nil(t, err)
	//now signal that we are ready to receive events
	svc.ReceivedEOS = true

	senderPrivkey := nostr.GeneratePrivateKey()
	senderPubkey, err := nostr.GetPublicKey(senderPrivkey)
	assert.NoError(t, err)
	//test lnbc.. payload without having an app registered
	ss, err := nip04.ComputeSharedSecret(svc.cfg.IdentityPubkey, senderPrivkey)
	assert.NoError(t, err)
	payload, err := nip04.Encrypt(nip47PayJson, ss)
	assert.NoError(t, err)
	res, err = svc.HandleEvent(ctx, &nostr.Event{
		ID:      "test_event_1",
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  senderPubkey,
		Content: payload,
	})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	received := &Nip47Response{}
	decrypted, err := nip04.Decrypt(res.Content, ss)
	assert.NoError(t, err)
	err = json.Unmarshal([]byte(decrypted), received)
	assert.NoError(t, err)
	assert.Equal(t, received.Error.Code, NIP_47_ERROR_UNAUTHORIZED)
	assert.NotNil(t, res)
	//create user
	user := &User{ID: 0, AlbyIdentifier: "dummy"}
	err = svc.db.Create(user).Error
	assert.NoError(t, err)
	//register app
	app := App{Name: "test", NostrPubkey: senderPubkey}
	err = svc.db.Model(&user).Association("Apps").Append(&app)
	assert.NoError(t, err)
	//test old payload
	res, err = svc.HandleEvent(ctx, &nostr.Event{
		ID:      "test_event_2",
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  senderPubkey,
		Content: payload,
	})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	//test new payload
	newPayload, err := nip04.Encrypt(nip47PayJson, ss)
	assert.NoError(t, err)
	res, err = svc.HandleEvent(ctx, &nostr.Event{
		ID:      "test_event_3",
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  senderPubkey,
		Content: newPayload,
	})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	decrypted, err = nip04.Decrypt(res.Content, ss)
	assert.NoError(t, err)
	received = &Nip47Response{
		Result: &Nip47PayResponse{},
	}
	err = json.Unmarshal([]byte(decrypted), received)
	assert.NoError(t, err)
	assert.Equal(t, received.Result.(*Nip47PayResponse).Preimage, "123preimage")
	malformedPayload, err := nip04.Encrypt(nip47PayJsonNoInvoice, ss)
	assert.NoError(t, err)
	res, err = svc.HandleEvent(ctx, &nostr.Event{
		ID:      "test_event_4",
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  senderPubkey,
		Content: malformedPayload,
	})

	decrypted, err = nip04.Decrypt(res.Content, ss)
	assert.NoError(t, err)
	receivedError := &Nip47Response{
		Result: &Nip47Error{},
	}
	err = json.Unmarshal([]byte(decrypted), receivedError)
	assert.NoError(t, err)
	assert.Equal(t, NIP_47_ERROR_INTERNAL, receivedError.Error.Code)

	//test wrong method
	wrongMethodPayload, err := nip04.Encrypt(nip47PayWrongMethodJson, ss)
	assert.NoError(t, err)
	res, err = svc.HandleEvent(ctx, &nostr.Event{
		ID:      "test_event_5",
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  senderPubkey,
		Content: wrongMethodPayload,
	})
	assert.NoError(t, err)
	//add app permissions
	maxAmount := 1000
	budgetRenewal := "never"
	expiresAt := time.Now().Add(24 * time.Hour)
	appPermission := &AppPermission{
		AppId:         app.ID,
		App:           app,
		RequestMethod: NIP_47_PAY_INVOICE_METHOD,
		MaxAmount:     maxAmount,
		BudgetRenewal: budgetRenewal,
		ExpiresAt:     expiresAt,
	}
	err = svc.db.Create(appPermission).Error
	assert.NoError(t, err)
	// permissions: no limitations
	res, err = svc.HandleEvent(ctx, &nostr.Event{
		ID:      "test_event_6",
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  senderPubkey,
		Content: newPayload,
	})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	decrypted, err = nip04.Decrypt(res.Content, ss)
	assert.NoError(t, err)
	received = &Nip47Response{
		Result: &Nip47PayResponse{},
	}
	err = json.Unmarshal([]byte(decrypted), received)
	assert.NoError(t, err)
	assert.Equal(t, "123preimage", received.Result.(*Nip47PayResponse).Preimage)
	// permissions: budget overflow
	newMaxAmount := 100
	err = svc.db.Model(&AppPermission{}).Where("app_id = ?", app.ID).Update("max_amount", newMaxAmount).Error

	res, err = svc.HandleEvent(ctx, &nostr.Event{
		ID:      "test_event_7",
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  senderPubkey,
		Content: newPayload,
	})
	assert.NoError(t, err)
	assert.NotNil(t, res)

	decrypted, err = nip04.Decrypt(res.Content, ss)
	assert.NoError(t, err)
	received = &Nip47Response{}
	err = json.Unmarshal([]byte(decrypted), received)
	assert.NoError(t, err)
	assert.Equal(t, NIP_47_ERROR_QUOTA_EXCEEDED, received.Error.Code)
	assert.NotNil(t, res)
	// permissions: expired app
	newExpiry := time.Now().Add(-24 * time.Hour)
	err = svc.db.Model(&AppPermission{}).Where("app_id = ?", app.ID).Update("expires_at", newExpiry).Error

	res, err = svc.HandleEvent(ctx, &nostr.Event{
		ID:      "test_event_8",
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  senderPubkey,
		Content: newPayload,
	})
	assert.NoError(t, err)
	assert.NotNil(t, res)

	decrypted, err = nip04.Decrypt(res.Content, ss)
	assert.NoError(t, err)
	received = &Nip47Response{}
	err = json.Unmarshal([]byte(decrypted), received)
	assert.NoError(t, err)
	assert.Equal(t, NIP_47_ERROR_EXPIRED, received.Error.Code)
	assert.NotNil(t, res)

	// remove expiry
	err = svc.db.Model(&AppPermission{}).Where("app_id = ?", app.ID).Update("expires_at", nil).Error
	// permissions: no request method
	err = svc.db.Model(&AppPermission{}).Where("app_id = ?", app.ID).Update("request_method", nil).Error

	res, err = svc.HandleEvent(ctx, &nostr.Event{
		ID:      "test_event_9",
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  senderPubkey,
		Content: newPayload,
	})
	assert.NoError(t, err)
	assert.NotNil(t, res)

	decrypted, err = nip04.Decrypt(res.Content, ss)
	assert.NoError(t, err)
	received = &Nip47Response{}
	err = json.Unmarshal([]byte(decrypted), received)
	assert.NoError(t, err)
	assert.Equal(t, NIP_47_ERROR_RESTRICTED, received.Error.Code)
	assert.NotNil(t, res)

	// pay_keysend: without permission
	newPayload, err = nip04.Encrypt(nip47KeysendJson, ss)
	assert.NoError(t, err)
	res, err = svc.HandleEvent(ctx, &nostr.Event{
		ID:      "test_pay_keysend_event_1",
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  senderPubkey,
		Content: newPayload,
	})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	decrypted, err = nip04.Decrypt(res.Content, ss)
	assert.NoError(t, err)
	received = &Nip47Response{}
	err = json.Unmarshal([]byte(decrypted), received)
	assert.NoError(t, err)
	assert.Equal(t, NIP_47_ERROR_RESTRICTED, received.Error.Code)
	assert.NotNil(t, res)
	// pay_keysend: with permission
	// update the existing permission to pay_invoice so we can have the budget info and increase max amount
	newMaxAmount = 1000
	err = svc.db.Model(&AppPermission{}).Where("app_id = ?", app.ID).Update("request_method", NIP_47_PAY_INVOICE_METHOD).Update("max_amount", newMaxAmount).Error
	assert.NoError(t, err)
	err = svc.db.Create(appPermission).Error
	res, err = svc.HandleEvent(ctx, &nostr.Event{
		ID:      "test_pay_keysend_event_2",
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  senderPubkey,
		Content: newPayload,
	})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	decrypted, err = nip04.Decrypt(res.Content, ss)
	assert.NoError(t, err)
	received = &Nip47Response{
		Result: &Nip47PayResponse{},
	}
	err = json.Unmarshal([]byte(decrypted), received)
	assert.NoError(t, err)
	assert.Equal(t, "123preimage", received.Result.(*Nip47PayResponse).Preimage)

	// keysend: budget overflow
	newMaxAmount = 100
	// we change the budget info in pay_invoice permission
	err = svc.db.Model(&AppPermission{}).Where("request_method = ?", NIP_47_PAY_INVOICE_METHOD).Update("max_amount", newMaxAmount).Error
	res, err = svc.HandleEvent(ctx, &nostr.Event{
		ID:      "test_pay_keysend_event_3",
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  senderPubkey,
		Content: newPayload,
	})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	decrypted, err = nip04.Decrypt(res.Content, ss)
	assert.NoError(t, err)
	received = &Nip47Response{
		Result: &Nip47PayResponse{},
	}
	err = json.Unmarshal([]byte(decrypted), received)
	assert.NoError(t, err)
	assert.Equal(t, NIP_47_ERROR_QUOTA_EXCEEDED, received.Error.Code)
	assert.NotNil(t, res)

	// get_balance: without permission
	newPayload, err = nip04.Encrypt(nip47GetBalanceJson, ss)
	assert.NoError(t, err)
	res, err = svc.HandleEvent(ctx, &nostr.Event{
		ID:      "test_event_10",
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  senderPubkey,
		Content: newPayload,
	})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	decrypted, err = nip04.Decrypt(res.Content, ss)
	assert.NoError(t, err)
	received = &Nip47Response{}
	err = json.Unmarshal([]byte(decrypted), received)
	assert.NoError(t, err)
	assert.Equal(t, NIP_47_ERROR_RESTRICTED, received.Error.Code)
	assert.NotNil(t, res)
	// get_balance: with permission
	// the pay_invoice permmission already exists with the budget info
	// create a second permission to fetch the balance and budget info
	appPermission = &AppPermission{
		AppId:         app.ID,
		App:           app,
		RequestMethod: NIP_47_GET_BALANCE_METHOD,
		ExpiresAt:     expiresAt,
	}
	err = svc.db.Create(appPermission).Error
	res, err = svc.HandleEvent(ctx, &nostr.Event{
		ID:      "test_event_11",
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  senderPubkey,
		Content: newPayload,
	})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	decrypted, err = nip04.Decrypt(res.Content, ss)
	assert.NoError(t, err)
	received = &Nip47Response{
		Result: &Nip47BalanceResponse{},
	}
	err = json.Unmarshal([]byte(decrypted), received)
	assert.NoError(t, err)
	assert.Equal(t, int64(21000), received.Result.(*Nip47BalanceResponse).Balance)
	assert.Equal(t, 100000, received.Result.(*Nip47BalanceResponse).MaxAmount)
	assert.Equal(t, "never", received.Result.(*Nip47BalanceResponse).BudgetRenewal)

	// make_invoice: without permission
	newPayload, err = nip04.Encrypt(nip47MakeInvoiceJson, ss)
	assert.NoError(t, err)
	res, err = svc.HandleEvent(ctx, &nostr.Event{
		ID:      "test_event_12",
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  senderPubkey,
		Content: newPayload,
	})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	decrypted, err = nip04.Decrypt(res.Content, ss)
	assert.NoError(t, err)
	received = &Nip47Response{}
	err = json.Unmarshal([]byte(decrypted), received)
	assert.NoError(t, err)
	assert.Equal(t, NIP_47_ERROR_RESTRICTED, received.Error.Code)
	assert.NotNil(t, res)

	// make_invoice: with permission
	err = svc.db.Model(&AppPermission{}).Where("app_id = ?", app.ID).Update("request_method", NIP_47_MAKE_INVOICE_METHOD).Error
	res, err = svc.HandleEvent(ctx, &nostr.Event{
		ID:      "test_event_13",
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  senderPubkey,
		Content: newPayload,
	})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	decrypted, err = nip04.Decrypt(res.Content, ss)
	assert.NoError(t, err)
	received = &Nip47Response{
		Result: &Nip47MakeInvoiceResponse{},
	}
	err = json.Unmarshal([]byte(decrypted), received)
	assert.NoError(t, err)
	assert.Equal(t, mockTransaction.Preimage, received.Result.(*Nip47MakeInvoiceResponse).Preimage)

	// lookup_invoice: without permission
	newPayload, err = nip04.Encrypt(nip47LookupInvoiceJson, ss)
	assert.NoError(t, err)
	res, err = svc.HandleEvent(ctx, &nostr.Event{
		ID:      "test_event_14",
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  senderPubkey,
		Content: newPayload,
	})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	decrypted, err = nip04.Decrypt(res.Content, ss)
	assert.NoError(t, err)
	received = &Nip47Response{}
	err = json.Unmarshal([]byte(decrypted), received)
	assert.NoError(t, err)
	assert.Equal(t, NIP_47_ERROR_RESTRICTED, received.Error.Code)
	assert.NotNil(t, res)

	// lookup_invoice: with permission
	err = svc.db.Model(&AppPermission{}).Where("app_id = ?", app.ID).Update("request_method", NIP_47_LOOKUP_INVOICE_METHOD).Error
	assert.NoError(t, err)
	res, err = svc.HandleEvent(ctx, &nostr.Event{
		ID:      "test_event_15",
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  senderPubkey,
		Content: newPayload,
	})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	decrypted, err = nip04.Decrypt(res.Content, ss)
	assert.NoError(t, err)
	received = &Nip47Response{
		Result: &Nip47LookupInvoiceResponse{},
	}
	err = json.Unmarshal([]byte(decrypted), received)
	assert.NoError(t, err)
	assert.Equal(t, mockTransaction.Preimage, received.Result.(*Nip47LookupInvoiceResponse).Preimage)

	// list_transactions: without permission
	newPayload, err = nip04.Encrypt(nip47ListTransactionsJson, ss)
	assert.NoError(t, err)
	res, err = svc.HandleEvent(ctx, &nostr.Event{
		ID:      "test_list_transactions_event_1",
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  senderPubkey,
		Content: newPayload,
	})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	decrypted, err = nip04.Decrypt(res.Content, ss)
	assert.NoError(t, err)
	received = &Nip47Response{}
	err = json.Unmarshal([]byte(decrypted), received)
	assert.NoError(t, err)
	assert.Equal(t, NIP_47_ERROR_RESTRICTED, received.Error.Code)
	assert.NotNil(t, res)

	// list_transactions: with permission
	err = svc.db.Model(&AppPermission{}).Where("app_id = ?", app.ID).Update("request_method", NIP_47_LIST_TRANSACTIONS_METHOD).Error
	assert.NoError(t, err)
	res, err = svc.HandleEvent(ctx, &nostr.Event{
		ID:      "test_list_transactions_event_2",
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  senderPubkey,
		Content: newPayload,
	})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	decrypted, err = nip04.Decrypt(res.Content, ss)
	assert.NoError(t, err)
	received = &Nip47Response{
		Result: &Nip47ListTransactionsResponse{},
	}
	err = json.Unmarshal([]byte(decrypted), received)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(received.Result.(*Nip47ListTransactionsResponse).Transactions))
	transaction := received.Result.(*Nip47ListTransactionsResponse).Transactions[0]
	assert.Equal(t, mockTransactions[0].Type, transaction.Type)
	assert.Equal(t, mockTransactions[0].Invoice, transaction.Invoice)
	assert.Equal(t, mockTransactions[0].Description, transaction.Description)
	assert.Equal(t, mockTransactions[0].DescriptionHash, transaction.DescriptionHash)
	assert.Equal(t, mockTransactions[0].Preimage, transaction.Preimage)
	assert.Equal(t, mockTransactions[0].PaymentHash, transaction.PaymentHash)
	assert.Equal(t, mockTransactions[0].Amount, transaction.Amount)
	assert.Equal(t, mockTransactions[0].FeesPaid, transaction.FeesPaid)
	assert.Equal(t, mockTransactions[0].SettledAt.Unix(), transaction.SettledAt.Unix())

	// get_info: without permission
	newPayload, err = nip04.Encrypt(nip47GetInfoJson, ss)
	assert.NoError(t, err)
	res, err = svc.HandleEvent(ctx, &nostr.Event{
		ID:      "test_event_16",
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  senderPubkey,
		Content: newPayload,
	})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	decrypted, err = nip04.Decrypt(res.Content, ss)
	assert.NoError(t, err)
	received = &Nip47Response{}
	err = json.Unmarshal([]byte(decrypted), received)
	assert.NoError(t, err)
	assert.Equal(t, NIP_47_ERROR_RESTRICTED, received.Error.Code)
	assert.NotNil(t, res)

	// delete all permissions
	svc.db.Exec("delete from app_permissions")

	// get_info: with permission
	appPermission = &AppPermission{
		AppId:         app.ID,
		App:           app,
		RequestMethod: NIP_47_GET_INFO_METHOD,
		ExpiresAt:     expiresAt,
	}
	err = svc.db.Create(appPermission).Error
	res, err = svc.HandleEvent(ctx, &nostr.Event{
		ID:      "test_event_17",
		Kind:    NIP_47_REQUEST_KIND,
		PubKey:  senderPubkey,
		Content: newPayload,
	})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	decrypted, err = nip04.Decrypt(res.Content, ss)
	assert.NoError(t, err)
	received = &Nip47Response{
		Result: &Nip47GetInfoResponse{},
	}
	err = json.Unmarshal([]byte(decrypted), received)
	assert.NoError(t, err)
	assert.Equal(t, mockNodeInfo.Alias, received.Result.(*Nip47GetInfoResponse).Alias)
	assert.Equal(t, mockNodeInfo.Color, received.Result.(*Nip47GetInfoResponse).Color)
	assert.Equal(t, mockNodeInfo.Pubkey, received.Result.(*Nip47GetInfoResponse).Pubkey)
	assert.Equal(t, mockNodeInfo.Network, received.Result.(*Nip47GetInfoResponse).Network)
	assert.Equal(t, mockNodeInfo.BlockHeight, received.Result.(*Nip47GetInfoResponse).BlockHeight)
	assert.Equal(t, mockNodeInfo.BlockHash, received.Result.(*Nip47GetInfoResponse).BlockHash)
	assert.Equal(t, []string{"get_info"}, received.Result.(*Nip47GetInfoResponse).Methods)
}

func createTestService(t *testing.T) (svc *Service, ln LNClient) {
	db, err := gorm.Open(sqlite.Open(testDB), &gorm.Config{})
	assert.NoError(t, err)
	err = db.AutoMigrate(&User{}, &App{}, &AppPermission{}, &NostrEvent{}, &Payment{}, &Identity{})
	assert.NoError(t, err)
	ln = &MockLn{}
	sk := nostr.GeneratePrivateKey()
	pk, err := nostr.GetPublicKey(sk)
	assert.NoError(t, err)

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
	}, ln
}

type MockLn struct {
}

func (mln *MockLn) SendPaymentSync(ctx context.Context, senderPubkey string, payReq string) (preimage string, err error) {
	return "123preimage", nil
}

func (mln *MockLn) SendKeysend(ctx context.Context, senderPubkey string, amount int64, destination, preimage string, custom_records []TLVRecord) (preImage string, err error) {
	return "123preimage", nil
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
