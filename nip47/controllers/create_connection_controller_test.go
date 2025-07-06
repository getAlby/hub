package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/nbd-wtf/go-nostr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/apps"
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/tests"
)

func TestHandleCreateConnectionEvent(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()
	svc.Cfg.SetUpdate("LNBackendType", config.LDKBackendType, "")

	pairingSecretKey := nostr.GeneratePrivateKey()
	pairingPublicKey, err := nostr.GetPublicKey(pairingSecretKey)
	require.NoError(t, err)

	nip47CreateConnectionJson := fmt.Sprintf(`
{
	"method": "create_connection",
	"params": {
		"pubkey": "%s",
		"name": "Test 123",
		"request_methods": ["get_info", "pay_invoice"],
		"notification_types": ["payment_received"],
		"max_amount": 100000000,
		"budget_renewal": "monthly",
		"isolated": true
	}
}
`, pairingPublicKey)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47CreateConnectionJson), nip47Request)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandleCreateConnectionEvent(ctx, nip47Request, dbRequestEvent.ID, publishResponse)

	assert.Nil(t, publishedResponse.Error)
	assert.Equal(t, models.CREATE_CONNECTION_METHOD, publishedResponse.ResultType)
	createAppResult := publishedResponse.Result.(createConnectionResponse)

	assert.NotNil(t, createAppResult.WalletPubkey)
	app := db.App{}
	err = svc.DB.First(&app).Error
	assert.NoError(t, err)
	assert.Equal(t, pairingPublicKey, app.AppPubkey)
	assert.Equal(t, createAppResult.WalletPubkey, *app.WalletPubkey)

	permissions := []db.AppPermission{}
	err = svc.DB.Find(&permissions).Error
	assert.NoError(t, err)
	assert.Equal(t, 3, len(permissions))
	assert.Equal(t, constants.GET_INFO_SCOPE, permissions[0].Scope)
	assert.Equal(t, constants.PAY_INVOICE_SCOPE, permissions[1].Scope)
	assert.Equal(t, constants.NOTIFICATIONS_SCOPE, permissions[2].Scope)

	assert.True(t, app.Isolated)
	assert.Equal(t, 100_000, permissions[1].MaxAmountSat)
	assert.Equal(t, constants.BUDGET_RENEWAL_MONTHLY, permissions[1].BudgetRenewal)
}

func TestHandleCreateConnectionEvent_IsolatedUnsupportedBackendType(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()
	svc.Cfg.SetUpdate("BackendType", config.CashuBackendType, "")

	pairingSecretKey := nostr.GeneratePrivateKey()
	pairingPublicKey, err := nostr.GetPublicKey(pairingSecretKey)
	require.NoError(t, err)

	nip47CreateConnectionJson := fmt.Sprintf(`
{
	"method": "create_connection",
	"params": {
		"pubkey": "%s",
		"name": "Test 123",
		"request_methods": ["get_info", "pay_invoice"],
		"notification_types": ["payment_received"],
		"max_amount": 100000000,
		"budget_renewal": "monthly",
		"isolated": true
	}
}
`, pairingPublicKey)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47CreateConnectionJson), nip47Request)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandleCreateConnectionEvent(ctx, nip47Request, dbRequestEvent.ID, publishResponse)

	assert.NotNil(t, publishedResponse.Error)
	assert.Equal(t, constants.ERROR_INTERNAL, publishedResponse.Error.Code)
	assert.Equal(t, "sub-wallets are currently not supported on your node backend. Try LDK or LND", publishedResponse.Error.Message)
	assert.Equal(t, models.CREATE_CONNECTION_METHOD, publishedResponse.ResultType)
}

func TestHandleCreateConnectionEvent_PubkeyAlreadyExists(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	pairingSecretKey := nostr.GeneratePrivateKey()
	pairingPublicKey, err := nostr.GetPublicKey(pairingSecretKey)
	require.NoError(t, err)

	appsSvc := apps.NewAppsService(svc.DB, svc.EventPublisher, svc.Keys, svc.Cfg)
	_, _, err = appsSvc.CreateApp("Existing App", pairingPublicKey, 0, constants.BUDGET_RENEWAL_NEVER, nil, []string{models.GET_INFO_METHOD}, false, nil)

	nip47CreateConnectionJson := fmt.Sprintf(`
{
	"method": "create_connection",
	"params": {
		"pubkey": "%s",
		"name": "Test 123",
		"request_methods": ["get_info"]
	}
}
`, pairingPublicKey)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47CreateConnectionJson), nip47Request)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandleCreateConnectionEvent(ctx, nip47Request, dbRequestEvent.ID, publishResponse)

	assert.NotNil(t, publishedResponse.Error)
	assert.Equal(t, constants.ERROR_INTERNAL, publishedResponse.Error.Code)
	assert.Equal(t, "duplicated key not allowed", publishedResponse.Error.Message)
	assert.Equal(t, models.CREATE_CONNECTION_METHOD, publishedResponse.ResultType)
	assert.Nil(t, publishedResponse.Result)
}

func TestHandleCreateConnectionEvent_NoMethods(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	pairingSecretKey := nostr.GeneratePrivateKey()
	pairingPublicKey, err := nostr.GetPublicKey(pairingSecretKey)
	require.NoError(t, err)

	nip47CreateConnectionJson := fmt.Sprintf(`
{
	"method": "create_connection",
	"params": {
		"pubkey": "%s",
		"name": "Test 123"
	}
}
`, pairingPublicKey)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47CreateConnectionJson), nip47Request)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandleCreateConnectionEvent(ctx, nip47Request, dbRequestEvent.ID, publishResponse)

	assert.NotNil(t, publishedResponse.Error)
	assert.Equal(t, constants.ERROR_BAD_REQUEST, publishedResponse.Error.Code)
	assert.Equal(t, "No request methods provided", publishedResponse.Error.Message)
	assert.Equal(t, models.CREATE_CONNECTION_METHOD, publishedResponse.ResultType)
	assert.Nil(t, publishedResponse.Result)
}

func TestHandleCreateConnectionEvent_UnsupportedMethod(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	pairingSecretKey := nostr.GeneratePrivateKey()
	pairingPublicKey, err := nostr.GetPublicKey(pairingSecretKey)
	require.NoError(t, err)

	nip47CreateConnectionJson := fmt.Sprintf(`
{
	"method": "create_connection",
	"params": {
		"pubkey": "%s",
		"name": "Test 123",
		"request_methods": ["non_existent"]
	}
}
`, pairingPublicKey)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47CreateConnectionJson), nip47Request)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandleCreateConnectionEvent(ctx, nip47Request, dbRequestEvent.ID, publishResponse)

	assert.NotNil(t, publishedResponse.Error)
	assert.Equal(t, constants.ERROR_BAD_REQUEST, publishedResponse.Error.Code)
	assert.Equal(t, "One or more methods are not supported by the current LNClient", publishedResponse.Error.Message)
	assert.Equal(t, models.CREATE_CONNECTION_METHOD, publishedResponse.ResultType)
	assert.Nil(t, publishedResponse.Result)
}

func TestHandleCreateConnectionEvent_UnsupportedNotificationType(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	pairingSecretKey := nostr.GeneratePrivateKey()
	pairingPublicKey, err := nostr.GetPublicKey(pairingSecretKey)
	require.NoError(t, err)

	nip47CreateConnectionJson := fmt.Sprintf(`
{
	"method": "create_connection",
	"params": {
		"pubkey": "%s",
		"name": "Test 123",
		"request_methods": ["get_info"],
		"notification_types": ["non_existent"]
	}
}
`, pairingPublicKey)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47CreateConnectionJson), nip47Request)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandleCreateConnectionEvent(ctx, nip47Request, dbRequestEvent.ID, publishResponse)

	assert.NotNil(t, publishedResponse.Error)
	assert.Equal(t, constants.ERROR_BAD_REQUEST, publishedResponse.Error.Code)
	assert.Equal(t, "One or more notification types are not supported by the current LNClient", publishedResponse.Error.Message)
	assert.Equal(t, models.CREATE_CONNECTION_METHOD, publishedResponse.ResultType)
	assert.Nil(t, publishedResponse.Result)
}

func TestHandleCreateConnectionEvent_DoNotAllowCreateConnectionMethod(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	pairingSecretKey := nostr.GeneratePrivateKey()
	pairingPublicKey, err := nostr.GetPublicKey(pairingSecretKey)
	require.NoError(t, err)

	nip47CreateConnectionJson := fmt.Sprintf(`
{
	"method": "create_connection",
	"params": {
		"pubkey": "%s",
		"name": "Test 123",
		"request_methods": ["create_connection"]
	}
}
`, pairingPublicKey)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47CreateConnectionJson), nip47Request)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandleCreateConnectionEvent(ctx, nip47Request, dbRequestEvent.ID, publishResponse)

	assert.NotNil(t, publishedResponse.Error)
	assert.Equal(t, constants.ERROR_BAD_REQUEST, publishedResponse.Error.Code)
	assert.Equal(t, "cannot create a new app that has create_connection permission via NWC", publishedResponse.Error.Message)
	assert.Equal(t, models.CREATE_CONNECTION_METHOD, publishedResponse.ResultType)
	assert.Nil(t, publishedResponse.Result)
}
