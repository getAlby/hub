package nip47

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/getAlby/go-nostr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/alby"
	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/nip47/cipher"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/nip47/permissions"
	nostrmodels "github.com/getAlby/hub/nostr/models"
	"github.com/getAlby/hub/tests"
)

// TODO: test if an app doesn't exist it returns the right error code

func TestCreateResponse_Nip04(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestCreateResponse(t, svc, constants.ENCRYPTION_TYPE_NIP04)
}

func TestCreateResponse_Nip44(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestCreateResponse(t, svc, constants.ENCRYPTION_TYPE_NIP44_V2)
}

func doTestCreateResponse(t *testing.T, svc *tests.TestService, nip47Encryption string) {
	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)
	assert.NoError(t, err)

	reqEvent := &nostr.Event{
		Kind:    models.REQUEST_KIND,
		PubKey:  reqPubkey,
		Content: "1",
	}

	reqEvent.ID = "12345"

	nip47Cipher, err := cipher.NewNip47Cipher(nip47Encryption, reqPubkey, svc.Keys.GetNostrSecretKey())
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

	albyOAuthSvc := alby.NewAlbyOAuthService(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)
	nip47svc := NewNip47Service(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher, albyOAuthSvc)

	res, err := nip47svc.CreateResponse(reqEvent, nip47Response, nostr.Tags{}, nip47Cipher, svc.Keys.GetNostrSecretKey())
	assert.NoError(t, err)
	assert.Equal(t, reqPubkey, res.Tags.Find("p")[1])
	assert.Equal(t, reqEvent.ID, res.Tags.Find("e")[1])
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
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestHandleResponse_WithPermission(t, svc, tests.CreateAppWithPrivateKey, constants.ENCRYPTION_TYPE_NIP04)
}

func TestHandleResponse_Nip44_WithPermission(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestHandleResponse_WithPermission(t, svc, tests.CreateAppWithPrivateKey, constants.ENCRYPTION_TYPE_NIP44_V2)
}

func doTestHandleResponse_WithPermission(t *testing.T, svc *tests.TestService, createAppFn tests.CreateAppFn, encryption string) {
	albyOAuthSvc := alby.NewAlbyOAuthService(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)
	nip47svc := NewNip47Service(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher, albyOAuthSvc)

	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)
	assert.NoError(t, err)

	app, cipher, err := createAppFn(svc, reqPrivateKey, encryption)
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

	if encryption != constants.ENCRYPTION_TYPE_NIP04 {
		reqEvent.Tags = append(reqEvent.Tags, []string{"encryption", encryption})
	}

	err = reqEvent.Sign(reqPrivateKey)
	assert.NoError(t, err)

	pool := tests.NewMockSimplePool()

	nip47svc.HandleEvent(context.TODO(), pool, reqEvent, svc.LNClient)

	assert.NotNil(t, pool.PublishedEvents[0])
	assert.NotEmpty(t, pool.PublishedEvents[0].Content)

	decrypted, err := cipher.Decrypt(pool.PublishedEvents[0].Content)
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
	assert.ElementsMatch(t, expectedMethods, unmarshalledResponse.Result.Methods)
}

func TestHandleResponse_Nip04_DuplicateRequest(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestHandleResponse_DuplicateRequest(t, svc, tests.CreateAppWithPrivateKey, constants.ENCRYPTION_TYPE_NIP04)
}

func TestHandleResponse_Nip44_DuplicateRequest(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestHandleResponse_DuplicateRequest(t, svc, tests.CreateAppWithPrivateKey, constants.ENCRYPTION_TYPE_NIP44_V2)
}

func doTestHandleResponse_DuplicateRequest(t *testing.T, svc *tests.TestService, createAppFn tests.CreateAppFn, encryption string) {
	albyOAuthSvc := alby.NewAlbyOAuthService(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)
	nip47svc := NewNip47Service(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher, albyOAuthSvc)

	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)
	assert.NoError(t, err)

	app, cipher, err := createAppFn(svc, reqPrivateKey, encryption)
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

	if encryption != constants.ENCRYPTION_TYPE_NIP04 {
		reqEvent.Tags = append(reqEvent.Tags, []string{"encryption", encryption})
	}

	err = reqEvent.Sign(reqPrivateKey)
	assert.NoError(t, err)

	pool := tests.NewMockSimplePool()

	nip47svc.HandleEvent(context.TODO(), pool, reqEvent, svc.LNClient)

	assert.NotNil(t, pool.PublishedEvents[0])
	assert.NotEmpty(t, pool.PublishedEvents[0].Content)

	pool.PublishedEvents = nil

	nip47svc.HandleEvent(context.TODO(), pool, reqEvent, svc.LNClient)

	// second time it should not publish
	assert.Nil(t, pool.PublishedEvents)
}

func TestHandleResponse_Nip04_NoPermission(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestHandleResponse_NoPermission(t, svc, tests.CreateAppWithPrivateKey, constants.ENCRYPTION_TYPE_NIP04)
}

func TestHandleResponse_Nip44_NoPermission(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestHandleResponse_NoPermission(t, svc, tests.CreateAppWithPrivateKey, constants.ENCRYPTION_TYPE_NIP44_V2)
}

func doTestHandleResponse_NoPermission(t *testing.T, svc *tests.TestService, createAppFn tests.CreateAppFn, encryption string) {
	albyOAuthSvc := alby.NewAlbyOAuthService(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)
	nip47svc := NewNip47Service(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher, albyOAuthSvc)

	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)
	assert.NoError(t, err)

	_, cipher, err := createAppFn(svc, reqPrivateKey, encryption)
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

	if encryption != constants.ENCRYPTION_TYPE_NIP04 {
		reqEvent.Tags = append(reqEvent.Tags, []string{"encryption", encryption})
	}

	err = reqEvent.Sign(reqPrivateKey)
	assert.NoError(t, err)

	pool := tests.NewMockSimplePool()

	nip47svc.HandleEvent(context.TODO(), pool, reqEvent, svc.LNClient)

	assert.NotNil(t, pool.PublishedEvents[0])
	assert.NotEmpty(t, pool.PublishedEvents[0].Content)

	decrypted, err := cipher.Decrypt(pool.PublishedEvents[0].Content)
	assert.NoError(t, err)

	unmarshalledResponse := models.Response{}

	err = json.Unmarshal([]byte(decrypted), &unmarshalledResponse)
	assert.NoError(t, err)
	assert.Nil(t, unmarshalledResponse.Result)
	assert.Equal(t, models.GET_BALANCE_METHOD, unmarshalledResponse.ResultType)
	assert.Equal(t, "RESTRICTED", unmarshalledResponse.Error.Code)
	assert.Equal(t, "This app does not have the get_balance scope", unmarshalledResponse.Error.Message)
}

func TestHandleResponse_UnknownMethod(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	albyOAuthSvc := alby.NewAlbyOAuthService(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)
	nip47svc := NewNip47Service(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher, albyOAuthSvc)

	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)
	require.NoError(t, err)

	_, cipher, err := tests.CreateAppWithPrivateKey(svc, reqPrivateKey, constants.ENCRYPTION_TYPE_NIP44_V2)
	require.NoError(t, err)

	payloadBytes, err := json.Marshal(map[string]interface{}{
		"method": "unknown_method",
	})
	require.NoError(t, err)

	msg, err := cipher.Encrypt(string(payloadBytes))
	require.NoError(t, err)

	reqEvent := &nostr.Event{
		Kind:      models.REQUEST_KIND,
		PubKey:    reqPubkey,
		CreatedAt: nostr.Now(),
		Tags:      nostr.Tags{[]string{"encryption", constants.ENCRYPTION_TYPE_NIP44_V2}},
		Content:   msg,
	}
	err = reqEvent.Sign(reqPrivateKey)
	require.NoError(t, err)

	pool := tests.NewMockSimplePool()
	nip47svc.HandleEvent(context.TODO(), pool, reqEvent, svc.LNClient)

	require.NotEmpty(t, pool.PublishedEvents)
	decrypted, err := cipher.Decrypt(pool.PublishedEvents[0].Content)
	require.NoError(t, err)

	response := models.Response{}
	err = json.Unmarshal([]byte(decrypted), &response)
	require.NoError(t, err)
	require.NotNil(t, response.Error)
	assert.Equal(t, "unknown_method", response.ResultType)
	assert.Equal(t, constants.ERROR_NOT_IMPLEMENTED, response.Error.Code)
}

func TestHandleResponse_Nip04_OldRequestForPayment(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestHandleResponse_OldRequestForPayment(t, svc, tests.CreateAppWithPrivateKey, constants.ENCRYPTION_TYPE_NIP04)
}

func TestHandleResponse_Nip44_OldRequestForPayment(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestHandleResponse_OldRequestForPayment(t, svc, tests.CreateAppWithPrivateKey, constants.ENCRYPTION_TYPE_NIP44_V2)
}

func doTestHandleResponse_OldRequestForPayment(t *testing.T, svc *tests.TestService, createAppFn tests.CreateAppFn, encryption string) {
	albyOAuthSvc := alby.NewAlbyOAuthService(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)
	nip47svc := NewNip47Service(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher, albyOAuthSvc)

	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)
	assert.NoError(t, err)

	app, cipher, err := createAppFn(svc, reqPrivateKey, encryption)
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

	if encryption != constants.ENCRYPTION_TYPE_NIP04 {
		reqEvent.Tags = append(reqEvent.Tags, []string{"encryption", encryption})
	}

	err = reqEvent.Sign(reqPrivateKey)
	assert.NoError(t, err)

	pool := tests.NewMockSimplePool()

	nip47svc.HandleEvent(context.TODO(), pool, reqEvent, svc.LNClient)

	// it shouldn't return anything for an old request
	assert.Nil(t, pool.PublishedEvents)

	// change the request to now
	reqEvent.CreatedAt = nostr.Now()
	err = reqEvent.Sign(reqPrivateKey)
	assert.NoError(t, err)

	nip47svc.HandleEvent(context.TODO(), pool, reqEvent, svc.LNClient)
	assert.NotNil(t, pool.PublishedEvents)
}

func TestHandleResponse_Nip04_IncorrectPubkey(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestHandleResponse_IncorrectPubkey(t, svc, tests.CreateAppWithPrivateKey, constants.ENCRYPTION_TYPE_NIP04)
}

func TestHandleResponse_Nip44_IncorrectPubkey(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestHandleResponse_IncorrectPubkey(t, svc, tests.CreateAppWithPrivateKey, constants.ENCRYPTION_TYPE_NIP44_V2)
}

func doTestHandleResponse_IncorrectPubkey(t *testing.T, svc *tests.TestService, createAppFn tests.CreateAppFn, encryption string) {
	albyOAuthSvc := alby.NewAlbyOAuthService(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)
	nip47svc := NewNip47Service(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher, albyOAuthSvc)

	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)
	assert.NoError(t, err)

	reqPrivateKey2 := nostr.GeneratePrivateKey()

	app, cipher, err := createAppFn(svc, reqPrivateKey, encryption)
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

	if encryption != constants.ENCRYPTION_TYPE_NIP04 {
		reqEvent.Tags = append(reqEvent.Tags, []string{"encryption", encryption})
	}

	err = reqEvent.Sign(reqPrivateKey2)
	assert.NoError(t, err)

	// set a different pubkey (this will not pass validation)
	reqEvent.PubKey = reqPubkey

	pool := tests.NewMockSimplePool()

	nip47svc.HandleEvent(context.TODO(), pool, reqEvent, svc.LNClient)

	assert.Nil(t, pool.PublishedEvents)
}

func TestHandleResponse_NoApp(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()
	albyOAuthSvc := alby.NewAlbyOAuthService(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)
	nip47svc := NewNip47Service(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher, albyOAuthSvc)

	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)
	assert.NoError(t, err)

	app, cipher, err := tests.CreateAppWithPrivateKey(svc, reqPrivateKey, constants.ENCRYPTION_TYPE_NIP44_V2)
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
		Tags:      nostr.Tags{[]string{"encryption", constants.ENCRYPTION_TYPE_NIP44_V2}},
		Content:   msg,
	}
	err = reqEvent.Sign(reqPrivateKey)
	assert.NoError(t, err)

	pool := tests.NewMockSimplePool()

	nip47svc.HandleEvent(context.TODO(), pool, reqEvent, svc.LNClient)

	// it shouldn't return anything for an invalid app key
	assert.Nil(t, pool.PublishedEvents)
}

func TestHandleResponse_UnknownEncryptionTag(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()
	doTestHandleResponse_UnknownEncryptionTag(t, svc, "nip44")
	doTestHandleResponse_UnknownEncryptionTag(t, svc, "nip44v2")
	doTestHandleResponse_UnknownEncryptionTag(t, svc, "nip44_v3")
	doTestHandleResponse_UnknownEncryptionTag(t, svc, "")
}

func doTestHandleResponse_UnknownEncryptionTag(t *testing.T, svc *tests.TestService, requestEncryptionTag string) {
	albyOAuthSvc := alby.NewAlbyOAuthService(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)
	nip47svc := NewNip47Service(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher, albyOAuthSvc)

	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)
	assert.NoError(t, err)

	app, cipher, err := tests.CreateAppWithPrivateKey(svc, reqPrivateKey, constants.ENCRYPTION_TYPE_NIP44_V2)
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

	// don't pass correct encryption
	reqEvent := &nostr.Event{
		Kind:      models.REQUEST_KIND,
		PubKey:    reqPubkey,
		CreatedAt: nostr.Now(),
		Tags:      nostr.Tags{[]string{"encryption", requestEncryptionTag}},
		Content:   msg,
	}

	err = reqEvent.Sign(reqPrivateKey)
	assert.NoError(t, err)

	pool := tests.NewMockSimplePool()

	nip47svc.HandleEvent(context.TODO(), pool, reqEvent, svc.LNClient)

	assert.NotNil(t, pool.PublishedEvents)
	responseContent := pool.PublishedEvents[0].Content
	msg, err = cipher.Decrypt(responseContent)
	assert.NoError(t, err)
	assert.NotEqual(t, "", msg)

	unmarshalledResponse := models.Response{}

	err = json.Unmarshal([]byte(msg), &unmarshalledResponse)
	assert.NoError(t, err)
	assert.Nil(t, unmarshalledResponse.Result)
	// assert.Equal(t, models.GET_INFO_METHOD, unmarshalledResponse.ResultType)
	assert.Equal(t, constants.ERROR_UNSUPPORTED_ENCRYPTION, unmarshalledResponse.Error.Code)
	assert.Contains(t, unmarshalledResponse.Error.Message, "invalid encryption:")
}

func TestHandleResponse_EncryptionTagDoesNotMatchPayload(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()
	// encryption specifies what cipher will use. If constants.ENCRYPTION_TYPE_NIP44_V2 is passed,
	// cipher must be NIP-44, otherwise cipher MUST be NIP-04
	doTestHandleResponse_EncryptionTagDoesNotMatchPayload(t, svc, constants.ENCRYPTION_TYPE_NIP44_V2, constants.ENCRYPTION_TYPE_NIP04)
	doTestHandleResponse_EncryptionTagDoesNotMatchPayload(t, svc, constants.ENCRYPTION_TYPE_NIP04, constants.ENCRYPTION_TYPE_NIP44_V2)
	doTestHandleResponse_EncryptionTagDoesNotMatchPayload(t, svc, constants.ENCRYPTION_TYPE_NIP44_V2, "")
}

func doTestHandleResponse_EncryptionTagDoesNotMatchPayload(t *testing.T, svc *tests.TestService, requestEncryption, requestEncryptionTag string) {
	albyOAuthSvc := alby.NewAlbyOAuthService(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)
	nip47svc := NewNip47Service(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher, albyOAuthSvc)

	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)
	assert.NoError(t, err)

	app, _, err := tests.CreateAppWithPrivateKey(svc, reqPrivateKey, constants.ENCRYPTION_TYPE_NIP44_V2)
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

	reqCipher, err := cipher.NewNip47Cipher(requestEncryption, *app.WalletPubkey, reqPrivateKey)
	assert.NoError(t, err)
	// whenever we are unable to handle the request encryption, we always respond with our preferred encryption (NIP44)
	nip44Cipher, err := cipher.NewNip47Cipher(constants.ENCRYPTION_TYPE_NIP44_V2, *app.WalletPubkey, reqPrivateKey)
	assert.NoError(t, err)
	msg, err := reqCipher.Encrypt(string(payloadBytes))
	assert.NoError(t, err)

	// don't pass correct encryption
	reqEvent := &nostr.Event{
		Kind:      models.REQUEST_KIND,
		PubKey:    reqPubkey,
		CreatedAt: nostr.Now(),
		Tags:      nostr.Tags{},
		Content:   msg,
	}

	if requestEncryptionTag != "" {
		reqEvent.Tags = append(reqEvent.Tags, []string{"encryption", requestEncryptionTag})
	}

	err = reqEvent.Sign(reqPrivateKey)
	assert.NoError(t, err)

	pool := tests.NewMockSimplePool()

	nip47svc.HandleEvent(context.TODO(), pool, reqEvent, svc.LNClient)

	assert.NotNil(t, pool.PublishedEvents)
	responseContent := pool.PublishedEvents[0].Content
	msg, err = nip44Cipher.Decrypt(responseContent)
	assert.NoError(t, err)
	assert.NotEqual(t, "", msg)

	unmarshalledResponse := models.Response{}
	err = json.Unmarshal([]byte(msg), &unmarshalledResponse)
	assert.NoError(t, err)
	assert.Nil(t, unmarshalledResponse.Result)
	// assert.Equal(t, models.GET_INFO_METHOD, unmarshalledResponse.ResultType)
	assert.Equal(t, constants.ERROR_BAD_REQUEST, unmarshalledResponse.Error.Code)
	assert.Contains(t, unmarshalledResponse.Error.Message, "failed to decrypt:")
}

func TestHandleResponse_MalformedRequest(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	albyOAuthSvc := alby.NewAlbyOAuthService(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)
	nip47svc := NewNip47Service(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher, albyOAuthSvc)

	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)
	assert.NoError(t, err)

	_, cipher, err := tests.CreateAppWithPrivateKey(svc, reqPrivateKey, constants.ENCRYPTION_TYPE_NIP44_V2)
	assert.NoError(t, err)

	// encrypts fine, but isn't a valid NIP-47 request
	msg, err := cipher.Encrypt("not json")
	assert.NoError(t, err)

	reqEvent := &nostr.Event{
		Kind:      models.REQUEST_KIND,
		PubKey:    reqPubkey,
		CreatedAt: nostr.Now(),
		Tags:      nostr.Tags{[]string{"encryption", constants.ENCRYPTION_TYPE_NIP44_V2}},
		Content:   msg,
	}
	err = reqEvent.Sign(reqPrivateKey)
	assert.NoError(t, err)

	pool := tests.NewMockSimplePool()

	nip47svc.HandleEvent(context.TODO(), pool, reqEvent, svc.LNClient)

	require.Len(t, pool.PublishedEvents, 1)

	decrypted, err := cipher.Decrypt(pool.PublishedEvents[0].Content)
	assert.NoError(t, err)

	unmarshalledResponse := models.Response{}
	err = json.Unmarshal([]byte(decrypted), &unmarshalledResponse)
	assert.NoError(t, err)
	assert.Nil(t, unmarshalledResponse.Result)
	assert.Equal(t, constants.ERROR_BAD_REQUEST, unmarshalledResponse.Error.Code)

	requestEvent := db.RequestEvent{}
	err = svc.DB.First(&requestEvent, "nostr_id = ?", reqEvent.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, db.REQUEST_EVENT_STATE_HANDLER_ERROR, requestEvent.State)
}

// no relay accepts any published event
type failingPool struct {
	nostrmodels.SimplePool // for QuerySingle
}

func (pool *failingPool) PublishMany(ctx context.Context, relayUrls []string, event nostr.Event) chan nostr.PublishResult {
	channel := make(chan nostr.PublishResult, 1)
	channel <- nostr.PublishResult{RelayURL: "wss://fakerelay.com", Error: errors.New("connection closed before receiving an OK")}
	close(channel)
	return channel
}

func TestHandleResponse_PublishFailed(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	albyOAuthSvc := alby.NewAlbyOAuthService(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)
	nip47svc := NewNip47Service(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher, albyOAuthSvc)

	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)
	require.NoError(t, err)

	app, cipher, err := tests.CreateAppWithPrivateKey(svc, reqPrivateKey, constants.ENCRYPTION_TYPE_NIP44_V2)
	require.NoError(t, err)

	payloadBytes, err := json.Marshal(map[string]interface{}{"method": models.GET_INFO_METHOD})
	require.NoError(t, err)
	msg, err := cipher.Encrypt(string(payloadBytes))
	require.NoError(t, err)

	reqEvent := &nostr.Event{
		Kind:      models.REQUEST_KIND,
		PubKey:    reqPubkey,
		CreatedAt: nostr.Now(),
		Tags:      nostr.Tags{[]string{"encryption", constants.ENCRYPTION_TYPE_NIP44_V2}},
		Content:   msg,
	}
	require.NoError(t, reqEvent.Sign(reqPrivateKey))

	nip47svc.HandleEvent(context.TODO(), &failingPool{SimplePool: tests.NewMockSimplePool()}, reqEvent, svc.LNClient)

	requestEvent := db.RequestEvent{}
	require.NoError(t, svc.DB.First(&requestEvent, &db.RequestEvent{NostrId: reqEvent.ID}).Error)
	assert.Equal(t, app.ID, *requestEvent.AppId)
	assert.Equal(t, db.REQUEST_EVENT_STATE_HANDLER_ERROR, requestEvent.State)

	responseEvent := db.ResponseEvent{}
	require.NoError(t, svc.DB.First(&responseEvent, &db.ResponseEvent{RequestId: requestEvent.ID}).Error)
	assert.Equal(t, db.RESPONSE_EVENT_STATE_PUBLISH_FAILED, responseEvent.State)
}

type invoiceScopeTestPermission struct {
	scope     string
	expiresAt *time.Time
}

// sendInvoiceScopeTestRequest creates an app with the given permissions, sends it a
// NIP-47 request and returns the decrypted response (nil if no response was published)
func sendInvoiceScopeTestRequest(t *testing.T, svc *tests.TestService, appPermissions []invoiceScopeTestPermission, method string, params map[string]interface{}, createdAt nostr.Timestamp) *models.Response {
	albyOAuthSvc := alby.NewAlbyOAuthService(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)
	nip47svc := NewNip47Service(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher, albyOAuthSvc)

	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)
	require.NoError(t, err)

	app, cipher, err := tests.CreateAppWithPrivateKey(svc, reqPrivateKey, constants.ENCRYPTION_TYPE_NIP44_V2)
	require.NoError(t, err)

	for _, appPermission := range appPermissions {
		err = svc.DB.Create(&db.AppPermission{
			AppId:     app.ID,
			App:       *app,
			Scope:     appPermission.scope,
			ExpiresAt: appPermission.expiresAt,
		}).Error
		require.NoError(t, err)
	}

	payloadBytes, err := json.Marshal(map[string]interface{}{
		"method": method,
		"params": params,
	})
	require.NoError(t, err)

	msg, err := cipher.Encrypt(string(payloadBytes))
	require.NoError(t, err)

	reqEvent := &nostr.Event{
		Kind:      models.REQUEST_KIND,
		PubKey:    reqPubkey,
		CreatedAt: createdAt,
		Tags:      nostr.Tags{[]string{"encryption", constants.ENCRYPTION_TYPE_NIP44_V2}},
		Content:   msg,
	}
	err = reqEvent.Sign(reqPrivateKey)
	require.NoError(t, err)

	pool := tests.NewMockSimplePool()
	nip47svc.HandleEvent(context.TODO(), pool, reqEvent, svc.LNClient)

	if len(pool.PublishedEvents) == 0 {
		return nil
	}

	decrypted, err := cipher.Decrypt(pool.PublishedEvents[0].Content)
	require.NoError(t, err)

	response := models.Response{}
	err = json.Unmarshal([]byte(decrypted), &response)
	require.NoError(t, err)
	return &response
}

func invoiceScopeTestParams(method string) map[string]interface{} {
	switch method {
	case models.MAKE_HOLD_INVOICE_METHOD:
		return map[string]interface{}{"amount": 1000, "payment_hash": "0000000000000000000000000000000000000000000000000000000000000001"}
	case models.SETTLE_HOLD_INVOICE_METHOD:
		return map[string]interface{}{"preimage": "0000000000000000000000000000000000000000000000000000000000000001"}
	case models.CANCEL_HOLD_INVOICE_METHOD:
		return map[string]interface{}{"payment_hash": "0000000000000000000000000000000000000000000000000000000000000001"}
	}
	return map[string]interface{}{"amount": 1000, "description": "test"}
}

var holdInvoiceMethods = []string{models.MAKE_HOLD_INVOICE_METHOD, models.SETTLE_HOLD_INVOICE_METHOD, models.CANCEL_HOLD_INVOICE_METHOD}

// assertPassedPermissionCheck asserts the request was not rejected by the permission check
func assertPassedPermissionCheck(t *testing.T, response *models.Response, method string) {
	require.NotNil(t, response, method)
	assert.Equal(t, method, response.ResultType)
	if response.Error != nil {
		assert.NotEqual(t, constants.ERROR_RESTRICTED, response.Error.Code, method)
		assert.NotEqual(t, constants.ERROR_EXPIRED, response.Error.Code, method)
	}
}

func TestHandleResponse_MakeInvoiceOnlyScope_AllowsMakeInvoice(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	response := sendInvoiceScopeTestRequest(t, svc, []invoiceScopeTestPermission{
		{scope: constants.MAKE_INVOICE_ONLY_SCOPE},
	}, models.MAKE_INVOICE_METHOD, invoiceScopeTestParams(models.MAKE_INVOICE_METHOD), nostr.Now())

	require.NotNil(t, response)
	assert.Equal(t, models.MAKE_INVOICE_METHOD, response.ResultType)
	assert.Nil(t, response.Error)
	assert.NotNil(t, response.Result)
}

func TestHandleResponse_MakeInvoiceOnlyScope_RestrictsHoldInvoiceMethods(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	for _, method := range holdInvoiceMethods {
		response := sendInvoiceScopeTestRequest(t, svc, []invoiceScopeTestPermission{
			{scope: constants.MAKE_INVOICE_ONLY_SCOPE},
			{scope: constants.LOOKUP_INVOICE_SCOPE},
		}, method, invoiceScopeTestParams(method), nostr.Now())

		require.NotNil(t, response, method)
		assert.Equal(t, method, response.ResultType)
		assert.Nil(t, response.Result, method)
		require.NotNil(t, response.Error, method)
		assert.Equal(t, constants.ERROR_RESTRICTED, response.Error.Code, method)
		assert.Equal(t, "This app does not have the make_invoice scope", response.Error.Message, method)
	}
}

func TestHandleResponse_MakeInvoiceScope_AllowsHoldInvoiceMethods(t *testing.T) {
	// regression: the bundled make_invoice scope keeps granting all invoice methods
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	for _, method := range append([]string{models.MAKE_INVOICE_METHOD}, holdInvoiceMethods...) {
		response := sendInvoiceScopeTestRequest(t, svc, []invoiceScopeTestPermission{
			{scope: constants.MAKE_INVOICE_SCOPE},
		}, method, invoiceScopeTestParams(method), nostr.Now())

		assertPassedPermissionCheck(t, response, method)
	}
}

func TestHandleResponse_MakeInvoiceOnlyScope_NoInvoiceScope(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	response := sendInvoiceScopeTestRequest(t, svc, []invoiceScopeTestPermission{
		{scope: constants.LOOKUP_INVOICE_SCOPE},
	}, models.MAKE_INVOICE_METHOD, invoiceScopeTestParams(models.MAKE_INVOICE_METHOD), nostr.Now())

	require.NotNil(t, response)
	require.NotNil(t, response.Error)
	assert.Equal(t, constants.ERROR_RESTRICTED, response.Error.Code)
	assert.Equal(t, "This app does not have the make_invoice scope", response.Error.Message)
}

func TestHandleResponse_MakeInvoiceOnlyScope_Expired(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	expiredAt := time.Now().Add(-time.Hour)
	response := sendInvoiceScopeTestRequest(t, svc, []invoiceScopeTestPermission{
		{scope: constants.MAKE_INVOICE_ONLY_SCOPE, expiresAt: &expiredAt},
	}, models.MAKE_INVOICE_METHOD, invoiceScopeTestParams(models.MAKE_INVOICE_METHOD), nostr.Now())

	require.NotNil(t, response)
	require.NotNil(t, response.Error)
	assert.Equal(t, constants.ERROR_EXPIRED, response.Error.Code)
}

func TestHandleResponse_MakeInvoiceOnlyScope_EitherScopeGrantsMakeInvoice(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	expiredAt := time.Now().Add(-time.Hour)
	validUntil := time.Now().Add(time.Hour)

	// expired make_invoice, valid make_invoice_only
	appPermissions := []invoiceScopeTestPermission{
		{scope: constants.MAKE_INVOICE_SCOPE, expiresAt: &expiredAt},
		{scope: constants.MAKE_INVOICE_ONLY_SCOPE, expiresAt: &validUntil},
	}
	response := sendInvoiceScopeTestRequest(t, svc, appPermissions, models.MAKE_INVOICE_METHOD, invoiceScopeTestParams(models.MAKE_INVOICE_METHOD), nostr.Now())
	require.NotNil(t, response)
	assert.Nil(t, response.Error)
	for _, method := range holdInvoiceMethods {
		// hold invoice methods follow the make_invoice scope only
		response = sendInvoiceScopeTestRequest(t, svc, appPermissions, method, invoiceScopeTestParams(method), nostr.Now())
		require.NotNil(t, response, method)
		require.NotNil(t, response.Error, method)
		assert.Equal(t, constants.ERROR_EXPIRED, response.Error.Code, method)
	}

	// valid make_invoice, expired make_invoice_only
	appPermissions = []invoiceScopeTestPermission{
		{scope: constants.MAKE_INVOICE_SCOPE, expiresAt: &validUntil},
		{scope: constants.MAKE_INVOICE_ONLY_SCOPE, expiresAt: &expiredAt},
	}
	for _, method := range append([]string{models.MAKE_INVOICE_METHOD}, holdInvoiceMethods...) {
		response = sendInvoiceScopeTestRequest(t, svc, appPermissions, method, invoiceScopeTestParams(method), nostr.Now())
		assertPassedPermissionCheck(t, response, method)
	}
}

func TestHandleResponse_MakeInvoiceOnlyScope_OldRequestIgnored(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	appPermissions := []invoiceScopeTestPermission{{scope: constants.MAKE_INVOICE_ONLY_SCOPE}}

	// invoice creation requests older than 6 hours are ignored
	response := sendInvoiceScopeTestRequest(t, svc, appPermissions, models.MAKE_INVOICE_METHOD, invoiceScopeTestParams(models.MAKE_INVOICE_METHOD), nostr.Timestamp(time.Now().Add(-7*time.Hour).Unix()))
	assert.Nil(t, response)

	response = sendInvoiceScopeTestRequest(t, svc, appPermissions, models.MAKE_INVOICE_METHOD, invoiceScopeTestParams(models.MAKE_INVOICE_METHOD), nostr.Now())
	require.NotNil(t, response)
	assert.Nil(t, response.Error)
}
