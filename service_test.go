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

const nip47GetJson = `
{
	"method": "get_balance"
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
	assert.Error(t, err)
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
	assert.Equal(t, received.Result.(*Nip47PayResponse).Preimage, "123preimage")
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
	assert.Equal(t, received.Error.Code, NIP_47_ERROR_QUOTA_EXCEEDED)
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
	assert.Equal(t, received.Error.Code, NIP_47_ERROR_EXPIRED)
	assert.NotNil(t, res)
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
	assert.Equal(t, received.Error.Code, NIP_47_ERROR_RESTRICTED)
	assert.NotNil(t, res)
	// get_balance: without permission
	newPayload, err = nip04.Encrypt(nip47GetJson, ss)
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
	assert.Equal(t, received.Error.Code, NIP_47_ERROR_RESTRICTED)
	assert.NotNil(t, res)
	// get_balance: with permission
	err = svc.db.Model(&AppPermission{}).Where("app_id = ?", app.ID).Update("request_method", "pay_invoice").Error
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
	assert.Equal(t, received.Result.(*Nip47BalanceResponse).Balance, int64(21))
	assert.Equal(t, received.Result.(*Nip47BalanceResponse).MaxAmount, 100)
	assert.Equal(t, received.Result.(*Nip47BalanceResponse).BudgetRenewal, "never")
}

func createTestService(t *testing.T) (svc *Service, ln *MockLn) {
	db, err := gorm.Open(sqlite.Open(testDB), &gorm.Config{})
	assert.NoError(t, err)
	err = db.AutoMigrate(&User{}, &App{}, &AppPermission{}, &NostrEvent{}, &Payment{}, &Identity{})
	assert.NoError(t, err)
	ln = &MockLn{}
	sk := nostr.GeneratePrivateKey()
	pk, err := nostr.GetPublicKey(sk)
	assert.NoError(t, err)
	return &Service{
		cfg: &Config{
			NostrSecretKey: sk,
			IdentityPubkey: pk,
		},
		db:          db,
		lnClient:    ln,
		ReceivedEOS: false,
		Logger:      &logrus.Logger{},
	}, ln
}

type MockLn struct {
}

func (mln *MockLn) SendPaymentSync(ctx context.Context, senderPubkey string, payReq string) (preimage string, err error) {
	//todo more advanced behaviour
	return "123preimage", nil
}

func (mln *MockLn) GetBalance(ctx context.Context, senderPubkey string) (balance int64, err error) {
	return 21, nil
}
