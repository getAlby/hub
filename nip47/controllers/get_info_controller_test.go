package controllers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/nbd-wtf/go-nostr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/tests"
)

const nip47GetInfoJson = `
{
	"method": "get_info"
}
`

func TestHandleGetInfoEvent_NoPermission(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	metadata := map[string]interface{}{
		"a": 123,
	}

	app, _, err := svc.AppsService.CreateApp("test", "", 0, "monthly", nil, []string{constants.GET_INFO_SCOPE}, false, metadata)
	assert.NoError(t, err)

	lightningAddress := "hello@getalby.com"
	svc.Cfg.SetUpdate("AlbyLightningAddress", lightningAddress, "")

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47GetInfoJson), nip47Request)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	// delete the existing app permissions (the app was created with get_info scope)
	svc.DB.Exec("delete from app_permissions")

	appPermission := &db.AppPermission{
		AppId:     app.ID,
		Scope:     constants.GET_BALANCE_SCOPE,
		ExpiresAt: nil,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandleGetInfoEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	assert.Nil(t, publishedResponse.Error)
	infoResponse := publishedResponse.Result.(*getInfoResponse)
	assert.Nil(t, infoResponse.Alias)
	assert.Nil(t, infoResponse.Color)
	assert.Nil(t, infoResponse.Pubkey)
	assert.Nil(t, infoResponse.Network)
	assert.Nil(t, infoResponse.BlockHeight)
	assert.Nil(t, infoResponse.BlockHash)
	require.NotNil(t, infoResponse.LightningAddress)
	assert.Equal(t, lightningAddress, *infoResponse.LightningAddress)
	// get_info method is always granted, but does not return pubkey
	assert.Contains(t, infoResponse.Methods, models.GET_INFO_METHOD)
	assert.Equal(t, []string{}, infoResponse.Notifications)
	require.NotNil(t, infoResponse.Metadata)
	assert.Equal(t, float64(123), infoResponse.Metadata.(map[string]interface{})["a"])
	assert.Equal(t, app.ID, infoResponse.Metadata.(map[string]interface{})["id"])
	assert.Equal(t, app.Name, infoResponse.Metadata.(map[string]interface{})["name"])
}

func TestHandleGetInfoEvent_SubwalletNoPermission(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	lightningAddress := "hello@getalby.com"

	metadata := map[string]interface{}{
		constants.METADATA_APPSTORE_APP_ID_KEY: constants.SUBWALLET_APPSTORE_APP_ID,
		"lud16":                                lightningAddress,
	}

	svc.Cfg.SetUpdate("LNBackendType", config.LDKBackendType, "")

	app, _, err := svc.AppsService.CreateApp("test", "", 0, "monthly", nil, []string{constants.GET_INFO_SCOPE}, true, metadata)
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47GetInfoJson), nip47Request)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	// delete the existing app permissions (the app was created with get_info scope)
	svc.DB.Exec("delete from app_permissions")

	appPermission := &db.AppPermission{
		AppId:     app.ID,
		Scope:     constants.GET_BALANCE_SCOPE,
		ExpiresAt: nil,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandleGetInfoEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	assert.Nil(t, publishedResponse.Error)
	infoResponse := publishedResponse.Result.(*getInfoResponse)
	assert.Nil(t, infoResponse.Alias)
	assert.Nil(t, infoResponse.Color)
	assert.Nil(t, infoResponse.Pubkey)
	assert.Nil(t, infoResponse.Network)
	assert.Nil(t, infoResponse.BlockHeight)
	assert.Nil(t, infoResponse.BlockHash)
	require.NotNil(t, infoResponse.LightningAddress)
	assert.Equal(t, lightningAddress, *infoResponse.LightningAddress)
	// get_info method is always granted, but does not return pubkey
	assert.Contains(t, infoResponse.Methods, models.GET_INFO_METHOD)
	assert.Equal(t, []string{}, infoResponse.Notifications)
	require.NotNil(t, infoResponse.Metadata)
	assert.Equal(t, lightningAddress, infoResponse.Metadata.(map[string]interface{})["lud16"])
	assert.Equal(t, app.ID, infoResponse.Metadata.(map[string]interface{})["id"])
	assert.Equal(t, app.Name, infoResponse.Metadata.(map[string]interface{})["name"])
}

func TestHandleGetInfoEvent_WithPermission(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47GetInfoJson), nip47Request)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId:     app.ID,
		Scope:     constants.GET_INFO_SCOPE,
		ExpiresAt: nil,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandleGetInfoEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	assert.Nil(t, publishedResponse.Error)
	infoResponse := publishedResponse.Result.(*getInfoResponse)
	assert.Equal(t, tests.MockNodeInfo.Alias, *infoResponse.Alias)
	assert.Equal(t, tests.MockNodeInfo.Color, *infoResponse.Color)
	assert.Equal(t, tests.MockNodeInfo.Pubkey, *infoResponse.Pubkey)
	assert.Equal(t, tests.MockNodeInfo.Network, *infoResponse.Network)
	assert.Equal(t, tests.MockNodeInfo.BlockHeight, *infoResponse.BlockHeight)
	assert.Equal(t, tests.MockNodeInfo.BlockHash, *infoResponse.BlockHash)
	assert.Contains(t, infoResponse.Methods, "get_info")
	assert.Equal(t, []string{}, infoResponse.Notifications)
}

func TestHandleGetInfoEvent_WithMetadata(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	lightningAddress := "hello@getalby.com"
	svc.Cfg.SetUpdate("AlbyLightningAddress", lightningAddress, "")

	metadata := map[string]interface{}{
		"a": 123,
	}

	app, _, err := svc.AppsService.CreateApp("test", "", 0, "monthly", nil, []string{constants.GET_INFO_SCOPE}, false, metadata)
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47GetInfoJson), nip47Request)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId:     app.ID,
		Scope:     constants.GET_INFO_SCOPE,
		ExpiresAt: nil,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandleGetInfoEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	assert.Nil(t, publishedResponse.Error)
	infoResponse := publishedResponse.Result.(*getInfoResponse)
	assert.Equal(t, tests.MockNodeInfo.Alias, *infoResponse.Alias)
	assert.Equal(t, tests.MockNodeInfo.Color, *infoResponse.Color)
	assert.Equal(t, tests.MockNodeInfo.Pubkey, *infoResponse.Pubkey)
	assert.Equal(t, tests.MockNodeInfo.Network, *infoResponse.Network)
	assert.Equal(t, tests.MockNodeInfo.BlockHeight, *infoResponse.BlockHeight)
	assert.Equal(t, tests.MockNodeInfo.BlockHash, *infoResponse.BlockHash)
	assert.Equal(t, lightningAddress, *infoResponse.LightningAddress)
	assert.Contains(t, infoResponse.Methods, "get_info")
	assert.Equal(t, []string{}, infoResponse.Notifications)
	assert.Equal(t, float64(123), infoResponse.Metadata.(map[string]interface{})["a"])
}

func TestHandleGetInfoEvent_SubwalletWithMetadata(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	lightningAddress := "hello@getalby.com"

	metadata := map[string]interface{}{
		constants.METADATA_APPSTORE_APP_ID_KEY: constants.SUBWALLET_APPSTORE_APP_ID,
		"lud16":                                lightningAddress,
		"a":                                    123,
	}

	svc.Cfg.SetUpdate("LNBackendType", config.LDKBackendType, "")
	app, _, err := svc.AppsService.CreateApp("test", "", 0, "monthly", nil, []string{constants.GET_INFO_SCOPE}, true, metadata)
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47GetInfoJson), nip47Request)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId:     app.ID,
		Scope:     constants.GET_INFO_SCOPE,
		ExpiresAt: nil,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandleGetInfoEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	assert.Nil(t, publishedResponse.Error)
	infoResponse := publishedResponse.Result.(*getInfoResponse)
	assert.Equal(t, tests.MockNodeInfo.Alias, *infoResponse.Alias)
	assert.Equal(t, tests.MockNodeInfo.Color, *infoResponse.Color)
	assert.Equal(t, tests.MockNodeInfo.Pubkey, *infoResponse.Pubkey)
	assert.Equal(t, tests.MockNodeInfo.Network, *infoResponse.Network)
	assert.Equal(t, tests.MockNodeInfo.BlockHeight, *infoResponse.BlockHeight)
	assert.Equal(t, tests.MockNodeInfo.BlockHash, *infoResponse.BlockHash)
	assert.Equal(t, lightningAddress, *infoResponse.LightningAddress)
	assert.Contains(t, infoResponse.Methods, "get_info")
	assert.Equal(t, []string{}, infoResponse.Notifications)
	assert.Equal(t, float64(123), infoResponse.Metadata.(map[string]interface{})["a"])
}

func TestHandleGetInfoEvent_WithNotifications(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47GetInfoJson), nip47Request)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId:     app.ID,
		Scope:     constants.GET_INFO_SCOPE,
		ExpiresAt: nil,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	appPermission = &db.AppPermission{
		AppId:     app.ID,
		Scope:     constants.NOTIFICATIONS_SCOPE,
		ExpiresAt: nil,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandleGetInfoEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	assert.Nil(t, publishedResponse.Error)
	infoResponse := publishedResponse.Result.(*getInfoResponse)
	assert.Equal(t, tests.MockNodeInfo.Alias, *infoResponse.Alias)
	assert.Equal(t, tests.MockNodeInfo.Color, *infoResponse.Color)
	assert.Equal(t, tests.MockNodeInfo.Pubkey, *infoResponse.Pubkey)
	assert.Equal(t, tests.MockNodeInfo.Network, *infoResponse.Network)
	assert.Equal(t, tests.MockNodeInfo.BlockHeight, *infoResponse.BlockHeight)
	assert.Equal(t, tests.MockNodeInfo.BlockHash, *infoResponse.BlockHash)
	assert.Contains(t, infoResponse.Methods, "get_info")
	assert.Equal(t, []string{"payment_received", "payment_sent"}, infoResponse.Notifications)
}
