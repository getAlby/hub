package nip47

import (
	"context"
	"encoding/json"
	"slices"
	"testing"
	"time"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/nip47/cipher"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/nip47/permissions"
	"github.com/getAlby/hub/tests"
	"github.com/nbd-wtf/go-nostr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: test HandleEvent
// TODO: test a request cannot be processed twice
// TODO: test if an app doesn't exist it returns the right error code

func TestCreateResponse_Nip04(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	doTestCreateResponse(t, svc, "0.0")
}

func TestCreateResponse_Nip44(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	doTestCreateResponse(t, svc, "1.0")
}

func doTestCreateResponse(t *testing.T, svc *tests.TestService, nip47Version string) {
	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)
	assert.NoError(t, err)

	reqEvent := &nostr.Event{
		Kind:    models.REQUEST_KIND,
		PubKey:  reqPubkey,
		Content: "1",
	}

	reqEvent.ID = "12345"

	nip47Cipher, err := cipher.NewNip47Cipher(nip47Version, reqPubkey, svc.Keys.GetNostrSecretKey())
	assert.NoError(t, err)

	type dummyResponse struct {
		Foo int
	}

	nip47Response := &models.Response{
		ResultType: "dummy_method",
		Result: dummyResponse{
			Foo: 1000,
		},
	}

	nip47svc := NewNip47Service(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)

	res, err := nip47svc.CreateResponse(reqEvent, nip47Response, nostr.Tags{}, nip47Cipher, svc.Keys.GetNostrSecretKey())
	assert.NoError(t, err)
	assert.Equal(t, reqPubkey, res.Tags.GetFirst([]string{"p"}).Value())
	assert.Equal(t, reqEvent.ID, res.Tags.GetFirst([]string{"e"}).Value())
	assert.Equal(t, svc.Keys.GetNostrPublicKey(), res.PubKey)

	decrypted, err := nip47Cipher.Decrypt(res.Content)
	assert.NoError(t, err)
	unmarshalledResponse := models.Response{
		Result: &dummyResponse{},
	}

	err = json.Unmarshal([]byte(decrypted), &unmarshalledResponse)
	assert.NoError(t, err)
	assert.Nil(t, nip47Response.Error)
	assert.Equal(t, nip47Response.ResultType, unmarshalledResponse.ResultType)
	assert.Equal(t, nip47Response.Result, *unmarshalledResponse.Result.(*dummyResponse))
}

func TestHandleResponse_Nip04_WithPermission(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	doTestHandleResponse_WithPermission(t, svc, tests.CreateAppWithPrivateKey, "0.0")
}

func TestHandleResponse_Nip44_WithPermission(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	doTestHandleResponse_WithPermission(t, svc, tests.CreateAppWithPrivateKey, "1.0")
}

func doTestHandleResponse_WithPermission(t *testing.T, svc *tests.TestService, createAppFn tests.CreateAppFn, version string) {
	nip47svc := NewNip47Service(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)

	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)
	assert.NoError(t, err)

	app, cipher, err := createAppFn(svc, reqPrivateKey, version)
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

	msg, err := cipher.Encrypt(string(payloadBytes))
	assert.NoError(t, err)

	reqEvent := &nostr.Event{
		Kind:      models.REQUEST_KIND,
		PubKey:    reqPubkey,
		CreatedAt: nostr.Now(),
		Tags:      nostr.Tags{},
		Content:   msg,
	}

	if version != "0.0" {
		reqEvent.Tags = append(reqEvent.Tags, []string{"v", version})
	}

	err = reqEvent.Sign(reqPrivateKey)
	assert.NoError(t, err)

	relay := tests.NewMockRelay()

	nip47svc.HandleEvent(context.TODO(), relay, reqEvent, svc.LNClient)

	assert.NotNil(t, relay.PublishedEvents[0])
	assert.NotEmpty(t, relay.PublishedEvents[0].Content)

	decrypted, err := cipher.Decrypt(relay.PublishedEvents[0].Content)
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

func TestHandleResponse_Nip04_DuplicateRequest(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	doTestHandleResponse_DuplicateRequest(t, svc, tests.CreateAppWithPrivateKey, "0.0")
}

func TestHandleResponse_Nip44_DuplicateRequest(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	doTestHandleResponse_DuplicateRequest(t, svc, tests.CreateAppWithPrivateKey, "1.0")
}

func doTestHandleResponse_DuplicateRequest(t *testing.T, svc *tests.TestService, createAppFn tests.CreateAppFn, version string) {
	nip47svc := NewNip47Service(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)

	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)
	assert.NoError(t, err)

	app, cipher, err := createAppFn(svc, reqPrivateKey, version)
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

	msg, err := cipher.Encrypt(string(payloadBytes))
	assert.NoError(t, err)

	reqEvent := &nostr.Event{
		Kind:      models.REQUEST_KIND,
		PubKey:    reqPubkey,
		CreatedAt: nostr.Now(),
		Tags:      nostr.Tags{},
		Content:   msg,
	}

	if version != "0.0" {
		reqEvent.Tags = append(reqEvent.Tags, []string{"v", version})
	}

	err = reqEvent.Sign(reqPrivateKey)
	assert.NoError(t, err)

	relay := tests.NewMockRelay()

	nip47svc.HandleEvent(context.TODO(), relay, reqEvent, svc.LNClient)

	assert.NotNil(t, relay.PublishedEvents[0])
	assert.NotEmpty(t, relay.PublishedEvents[0].Content)

	relay.PublishedEvents = nil

	nip47svc.HandleEvent(context.TODO(), relay, reqEvent, svc.LNClient)

	// second time it should not publish
	assert.Nil(t, relay.PublishedEvents)
}

func TestHandleResponse_Nip04_NoPermission(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	doTestHandleResponse_NoPermission(t, svc, tests.CreateAppWithPrivateKey, "0.0")
}

func TestHandleResponse_Nip44_NoPermission(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	doTestHandleResponse_NoPermission(t, svc, tests.CreateAppWithPrivateKey, "1.0")
}

func doTestHandleResponse_NoPermission(t *testing.T, svc *tests.TestService, createAppFn tests.CreateAppFn, version string) {
	nip47svc := NewNip47Service(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)

	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)
	assert.NoError(t, err)

	_, cipher, err := createAppFn(svc, reqPrivateKey, version)
	assert.NoError(t, err)

	content := map[string]interface{}{
		"method": models.GET_BALANCE_METHOD,
	}

	payloadBytes, err := json.Marshal(content)
	assert.NoError(t, err)

	msg, err := cipher.Encrypt(string(payloadBytes))
	assert.NoError(t, err)

	reqEvent := &nostr.Event{
		Kind:      models.REQUEST_KIND,
		PubKey:    reqPubkey,
		CreatedAt: nostr.Now(),
		Tags:      nostr.Tags{},
		Content:   msg,
	}

	if version != "0.0" {
		reqEvent.Tags = append(reqEvent.Tags, []string{"v", version})
	}

	err = reqEvent.Sign(reqPrivateKey)
	assert.NoError(t, err)

	relay := tests.NewMockRelay()

	nip47svc.HandleEvent(context.TODO(), relay, reqEvent, svc.LNClient)

	assert.NotNil(t, relay.PublishedEvents[0])
	assert.NotEmpty(t, relay.PublishedEvents[0].Content)

	decrypted, err := cipher.Decrypt(relay.PublishedEvents[0].Content)
	assert.NoError(t, err)

	unmarshalledResponse := models.Response{}

	err = json.Unmarshal([]byte(decrypted), &unmarshalledResponse)
	assert.NoError(t, err)
	assert.Nil(t, unmarshalledResponse.Result)
	assert.Equal(t, models.GET_BALANCE_METHOD, unmarshalledResponse.ResultType)
	assert.Equal(t, "RESTRICTED", unmarshalledResponse.Error.Code)
	assert.Equal(t, "This app does not have the get_balance scope", unmarshalledResponse.Error.Message)
}

func TestHandleResponse_Nip04_OldRequestForPayment(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	doTestHandleResponse_OldRequestForPayment(t, svc, tests.CreateAppWithPrivateKey, "0.0")
}

func TestHandleResponse_Nip44_OldRequestForPayment(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	doTestHandleResponse_OldRequestForPayment(t, svc, tests.CreateAppWithPrivateKey, "1.0")
}

func doTestHandleResponse_OldRequestForPayment(t *testing.T, svc *tests.TestService, createAppFn tests.CreateAppFn, version string) {
	nip47svc := NewNip47Service(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)

	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)
	assert.NoError(t, err)

	app, cipher, err := createAppFn(svc, reqPrivateKey, version)
	assert.NoError(t, err)

	content := map[string]interface{}{
		"method": models.PAY_INVOICE_METHOD,
	}

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: constants.PAY_INVOICE_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	payloadBytes, err := json.Marshal(content)
	assert.NoError(t, err)

	msg, err := cipher.Encrypt(string(payloadBytes))
	assert.NoError(t, err)

	reqEvent := &nostr.Event{
		Kind:      models.REQUEST_KIND,
		PubKey:    reqPubkey,
		CreatedAt: nostr.Timestamp(time.Now().Add(time.Duration(-6) * time.Hour).Unix()),
		Tags:      nostr.Tags{},
		Content:   msg,
	}

	if version != "0.0" {
		reqEvent.Tags = append(reqEvent.Tags, []string{"v", version})
	}

	err = reqEvent.Sign(reqPrivateKey)
	assert.NoError(t, err)

	relay := tests.NewMockRelay()

	nip47svc.HandleEvent(context.TODO(), relay, reqEvent, svc.LNClient)

	// it shouldn't return anything for an old request
	assert.Nil(t, relay.PublishedEvents)

	// change the request to now
	reqEvent.CreatedAt = nostr.Now()
	err = reqEvent.Sign(reqPrivateKey)
	assert.NoError(t, err)

	nip47svc.HandleEvent(context.TODO(), relay, reqEvent, svc.LNClient)
	assert.NotNil(t, relay.PublishedEvents)
}

func TestHandleResponse_Nip04_IncorrectPubkey(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	doTestHandleResponse_IncorrectPubkey(t, svc, tests.CreateAppWithPrivateKey, "0.0")
}

func TestHandleResponse_Nip44_IncorrectPubkey(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	doTestHandleResponse_IncorrectPubkey(t, svc, tests.CreateAppWithPrivateKey, "1.0")
}

func doTestHandleResponse_IncorrectPubkey(t *testing.T, svc *tests.TestService, createAppFn tests.CreateAppFn, version string) {
	nip47svc := NewNip47Service(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)

	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)
	assert.NoError(t, err)

	reqPrivateKey2 := nostr.GeneratePrivateKey()

	app, cipher, err := createAppFn(svc, reqPrivateKey, version)
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

	msg, err := cipher.Encrypt(string(payloadBytes))
	assert.NoError(t, err)

	reqEvent := &nostr.Event{
		Kind:      models.REQUEST_KIND,
		CreatedAt: nostr.Now(),
		Tags:      nostr.Tags{},
		Content:   msg,
	}

	if version != "0.0" {
		reqEvent.Tags = append(reqEvent.Tags, []string{"v", version})
	}

	err = reqEvent.Sign(reqPrivateKey2)
	assert.NoError(t, err)

	// set a different pubkey (this will not pass validation)
	reqEvent.PubKey = reqPubkey

	relay := tests.NewMockRelay()

	nip47svc.HandleEvent(context.TODO(), relay, reqEvent, svc.LNClient)

	assert.Nil(t, relay.PublishedEvents)
}

func TestHandleResponse_NoApp(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)
	nip47svc := NewNip47Service(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)

	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)
	assert.NoError(t, err)

	app, cipher, err := tests.CreateAppWithPrivateKey(svc, reqPrivateKey, "1.0")
	assert.NoError(t, err)

	// delete the app
	err = svc.DB.Delete(app).Error
	assert.NoError(t, err)

	content := map[string]interface{}{
		"method": models.GET_BALANCE_METHOD,
	}

	payloadBytes, err := json.Marshal(content)
	assert.NoError(t, err)

	msg, err := cipher.Encrypt(string(payloadBytes))
	assert.NoError(t, err)

	reqEvent := &nostr.Event{
		Kind:      models.REQUEST_KIND,
		PubKey:    reqPubkey,
		CreatedAt: nostr.Now(),
		Tags:      nostr.Tags{[]string{"v", "1.0"}},
		Content:   msg,
	}
	err = reqEvent.Sign(reqPrivateKey)
	assert.NoError(t, err)

	relay := tests.NewMockRelay()

	nip47svc.HandleEvent(context.TODO(), relay, reqEvent, svc.LNClient)

	// it shouldn't return anything for an invalid app key
	assert.Nil(t, relay.PublishedEvents)
}

func TestHandleResponse_IncorrectVersions(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	doTestHandleResponse_IncorrectVersion(t, svc, "0.0", "1.0")
	doTestHandleResponse_IncorrectVersion(t, svc, "1.0", "0.0")
}

func doTestHandleResponse_IncorrectVersion(t *testing.T, svc *tests.TestService, appVersion, requestVersion string) {
	nip47svc := NewNip47Service(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)

	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)
	assert.NoError(t, err)

	app, cipher, err := tests.CreateAppWithPrivateKey(svc, reqPrivateKey, appVersion)
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

	msg, err := cipher.Encrypt(string(payloadBytes))
	assert.NoError(t, err)

	// don't pass correct version
	reqEvent := &nostr.Event{
		Kind:      models.REQUEST_KIND,
		PubKey:    reqPubkey,
		CreatedAt: nostr.Now(),
		Tags:      nostr.Tags{[]string{"v", requestVersion}},
		Content:   msg,
	}

	if requestVersion != "0.0" {
		reqEvent.Tags = append(reqEvent.Tags, []string{"v", requestVersion})
	}

	err = reqEvent.Sign(reqPrivateKey)
	assert.NoError(t, err)

	relay := tests.NewMockRelay()

	nip47svc.HandleEvent(context.TODO(), relay, reqEvent, svc.LNClient)

	// it shouldn't return anything for an invalid version
	assert.Nil(t, relay.PublishedEvents)
}
