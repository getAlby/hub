package nip47

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/stretchr/testify/assert"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/nip47/permissions"
	"github.com/getAlby/hub/tests"
)

func TestHandleResponse_LegacyApp_WithPermission(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	assert.NoError(t, err)
	defer svc.Remove()
	nip47svc := NewNip47Service(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)

	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)
	assert.NoError(t, err)

	app, ss, err := tests.CreateLegacyApp(svc, reqPrivateKey)
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: constants.GET_BALANCE_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	content := map[string]interface{}{
		"method": models.GET_INFO_METHOD,
	}

	payloadBytes, err := json.Marshal(content)
	assert.NoError(t, err)

	msg, err := nip04.Encrypt(string(payloadBytes), ss)
	assert.NoError(t, err)

	reqEvent := &nostr.Event{
		Kind:      models.REQUEST_KIND,
		PubKey:    reqPubkey,
		CreatedAt: nostr.Now(),
		Tags:      nostr.Tags{},
		Content:   msg,
	}
	err = reqEvent.Sign(reqPrivateKey)
	assert.NoError(t, err)

	relay := tests.NewMockRelay()

	nip47svc.HandleEvent(context.TODO(), relay, reqEvent, svc.LNClient)

	assert.NotNil(t, relay.PublishedEvent)
	assert.NotEmpty(t, relay.PublishedEvent.Content)

	decrypted, err := nip04.Decrypt(relay.PublishedEvent.Content, ss)
	assert.NoError(t, err)

	type getInfoResult struct {
		Methods []string `json:"methods"`
	}

	type getInfoResponseWrapper struct {
		models.Response
		Result getInfoResult `json:"result"`
	}

	unmarshalledResponse := getInfoResponseWrapper{}

	err = json.Unmarshal([]byte(decrypted), &unmarshalledResponse)
	assert.NoError(t, err)
	assert.Nil(t, unmarshalledResponse.Error)
	assert.Equal(t, models.GET_INFO_METHOD, unmarshalledResponse.ResultType)
	expectedMethods := slices.Concat([]string{constants.GET_BALANCE_SCOPE}, permissions.GetAlwaysGrantedMethods())
	assert.Equal(t, expectedMethods, unmarshalledResponse.Result.Methods)
}

func TestHandleResponse_LegacyApp_NoPermission(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	assert.NoError(t, err)
	defer svc.Remove()
	nip47svc := NewNip47Service(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)

	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)
	assert.NoError(t, err)

	_, ss, err := tests.CreateLegacyApp(svc, reqPrivateKey)
	assert.NoError(t, err)

	content := map[string]interface{}{
		"method": models.GET_BALANCE_METHOD,
	}

	payloadBytes, err := json.Marshal(content)
	assert.NoError(t, err)

	msg, err := nip04.Encrypt(string(payloadBytes), ss)
	assert.NoError(t, err)

	reqEvent := &nostr.Event{
		Kind:      models.REQUEST_KIND,
		PubKey:    reqPubkey,
		CreatedAt: nostr.Now(),
		Tags:      nostr.Tags{},
		Content:   msg,
	}
	err = reqEvent.Sign(reqPrivateKey)
	assert.NoError(t, err)

	relay := tests.NewMockRelay()

	nip47svc.HandleEvent(context.TODO(), relay, reqEvent, svc.LNClient)

	assert.NotNil(t, relay.PublishedEvent)
	assert.NotEmpty(t, relay.PublishedEvent.Content)

	decrypted, err := nip04.Decrypt(relay.PublishedEvent.Content, ss)
	assert.NoError(t, err)

	unmarshalledResponse := models.Response{}

	err = json.Unmarshal([]byte(decrypted), &unmarshalledResponse)
	assert.NoError(t, err)
	assert.Nil(t, unmarshalledResponse.Result)
	assert.Equal(t, models.GET_BALANCE_METHOD, unmarshalledResponse.ResultType)
	assert.Equal(t, "RESTRICTED", unmarshalledResponse.Error.Code)
	assert.Equal(t, "This app does not have the get_balance scope", unmarshalledResponse.Error.Message)
}

func TestHandleResponse_LegacyApp_IncorrectPubkey(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	assert.NoError(t, err)
	defer svc.Remove()
	nip47svc := NewNip47Service(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)

	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)
	assert.NoError(t, err)

	reqPrivateKey2 := nostr.GeneratePrivateKey()

	app, ss, err := tests.CreateLegacyApp(svc, reqPrivateKey)
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: constants.GET_BALANCE_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	content := map[string]interface{}{
		"method": models.GET_BALANCE_METHOD,
	}

	payloadBytes, err := json.Marshal(content)
	assert.NoError(t, err)

	msg, err := nip04.Encrypt(string(payloadBytes), ss)
	assert.NoError(t, err)

	reqEvent := &nostr.Event{
		Kind:      models.REQUEST_KIND,
		CreatedAt: nostr.Now(),
		Tags:      nostr.Tags{},
		Content:   msg,
	}
	err = reqEvent.Sign(reqPrivateKey2)
	assert.NoError(t, err)

	// set a different pubkey (this will not pass validation)
	reqEvent.PubKey = reqPubkey

	relay := tests.NewMockRelay()

	nip47svc.HandleEvent(context.TODO(), relay, reqEvent, svc.LNClient)

	assert.Nil(t, relay.PublishedEvent)
}
