package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/nbd-wtf/go-nostr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/nip47/permissions"
	"github.com/getAlby/hub/tests"
	"github.com/getAlby/hub/transactions"
)

func TestHandleCreateConnectionEvent(t *testing.T) {
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
		"methods": ["get_info"]
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

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB, svc.EventPublisher)
	NewNip47Controller(svc.LNClient, svc.DB, svc.EventPublisher, permissionsSvc, transactionsSvc, svc.AppsService).
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
	assert.Equal(t, 1, len(permissions))
	assert.Equal(t, constants.GET_INFO_SCOPE, permissions[0].Scope)
}

func TestHandleCreateConnectionEvent_PubkeyAlreadyExists(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	pairingSecretKey := nostr.GeneratePrivateKey()
	pairingPublicKey, err := nostr.GetPublicKey(pairingSecretKey)
	require.NoError(t, err)

	_, _, err = svc.AppsService.CreateApp("Existing App", pairingPublicKey, 0, constants.BUDGET_RENEWAL_NEVER, nil, []string{models.GET_INFO_METHOD}, false, nil)

	nip47CreateConnectionJson := fmt.Sprintf(`
{
	"method": "create_connection",
	"params": {
		"pubkey": "%s",
		"name": "Test 123",
		"methods": ["get_info"]
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

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB, svc.EventPublisher)
	NewNip47Controller(svc.LNClient, svc.DB, svc.EventPublisher, permissionsSvc, transactionsSvc, svc.AppsService).
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

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB, svc.EventPublisher)
	NewNip47Controller(svc.LNClient, svc.DB, svc.EventPublisher, permissionsSvc, transactionsSvc, svc.AppsService).
		HandleCreateConnectionEvent(ctx, nip47Request, dbRequestEvent.ID, publishResponse)

	assert.NotNil(t, publishedResponse.Error)
	assert.Equal(t, constants.ERROR_INTERNAL, publishedResponse.Error.Code)
	assert.Equal(t, "No methods provided", publishedResponse.Error.Message)
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
		"methods": ["non_existent"]
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

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB, svc.EventPublisher)
	NewNip47Controller(svc.LNClient, svc.DB, svc.EventPublisher, permissionsSvc, transactionsSvc, svc.AppsService).
		HandleCreateConnectionEvent(ctx, nip47Request, dbRequestEvent.ID, publishResponse)

	assert.NotNil(t, publishedResponse.Error)
	assert.Equal(t, constants.ERROR_INTERNAL, publishedResponse.Error.Code)
	assert.Equal(t, "One or more methods are not supported by the current LNClient", publishedResponse.Error.Message)
	assert.Equal(t, models.CREATE_CONNECTION_METHOD, publishedResponse.ResultType)
	assert.Nil(t, publishedResponse.Result)
}

func TestHandleCreateConnectionEvent_CreateConnectionMethod(t *testing.T) {
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
		"methods": ["create_connection"]
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

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB, svc.EventPublisher)
	NewNip47Controller(svc.LNClient, svc.DB, svc.EventPublisher, permissionsSvc, transactionsSvc, svc.AppsService).
		HandleCreateConnectionEvent(ctx, nip47Request, dbRequestEvent.ID, publishResponse)

	assert.NotNil(t, publishedResponse.Error)
	assert.Equal(t, constants.ERROR_INTERNAL, publishedResponse.Error.Code)
	assert.Equal(t, "cannot create a new app that has create_connection permission via NWC", publishedResponse.Error.Message)
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
		"methods": ["create_connection"]
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

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB, svc.EventPublisher)
	NewNip47Controller(svc.LNClient, svc.DB, svc.EventPublisher, permissionsSvc, transactionsSvc, svc.AppsService).
		HandleCreateConnectionEvent(ctx, nip47Request, dbRequestEvent.ID, publishResponse)

	assert.NotNil(t, publishedResponse.Error)
	assert.Equal(t, constants.ERROR_INTERNAL, publishedResponse.Error.Code)
	assert.Equal(t, "One or more methods are not supported by the current LNClient", publishedResponse.Error.Message)
	assert.Equal(t, models.CREATE_CONNECTION_METHOD, publishedResponse.ResultType)
	assert.Nil(t, publishedResponse.Result)
}
