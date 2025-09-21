package controllers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/nbd-wtf/go-nostr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

	app, _, err := tests.CreateApp(svc)
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
	nodeInfo := publishedResponse.Result.(*getInfoResponse)
	assert.Nil(t, nodeInfo.Alias)
	assert.Nil(t, nodeInfo.Color)
	assert.Nil(t, nodeInfo.Pubkey)
	assert.Nil(t, nodeInfo.Network)
	assert.Nil(t, nodeInfo.BlockHeight)
	assert.Nil(t, nodeInfo.BlockHash)
	// get_info method is always granted, but does not return pubkey
	assert.Contains(t, nodeInfo.Methods, models.GET_INFO_METHOD)
	assert.Equal(t, []string{}, nodeInfo.Notifications)
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
	nodeInfo := publishedResponse.Result.(*getInfoResponse)
	assert.Equal(t, tests.MockNodeInfo.Alias, *nodeInfo.Alias)
	assert.Equal(t, tests.MockNodeInfo.Color, *nodeInfo.Color)
	assert.Equal(t, tests.MockNodeInfo.Pubkey, *nodeInfo.Pubkey)
	assert.Equal(t, tests.MockNodeInfo.Network, *nodeInfo.Network)
	assert.Equal(t, tests.MockNodeInfo.BlockHeight, *nodeInfo.BlockHeight)
	assert.Equal(t, tests.MockNodeInfo.BlockHash, *nodeInfo.BlockHash)
	assert.Contains(t, nodeInfo.Methods, "get_info")
	assert.Equal(t, []string{}, nodeInfo.Notifications)
}

func TestHandleGetInfoEvent_WithMetadata(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

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
	nodeInfo := publishedResponse.Result.(*getInfoResponse)
	assert.Equal(t, tests.MockNodeInfo.Alias, *nodeInfo.Alias)
	assert.Equal(t, tests.MockNodeInfo.Color, *nodeInfo.Color)
	assert.Equal(t, tests.MockNodeInfo.Pubkey, *nodeInfo.Pubkey)
	assert.Equal(t, tests.MockNodeInfo.Network, *nodeInfo.Network)
	assert.Equal(t, tests.MockNodeInfo.BlockHeight, *nodeInfo.BlockHeight)
	assert.Equal(t, tests.MockNodeInfo.BlockHash, *nodeInfo.BlockHash)
	assert.Contains(t, nodeInfo.Methods, "get_info")
	assert.Equal(t, []string{}, nodeInfo.Notifications)

	assert.NoError(t, err)
	assert.Equal(t, float64(123), nodeInfo.Metadata.(map[string]interface{})["a"])
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
	nodeInfo := publishedResponse.Result.(*getInfoResponse)
	assert.Equal(t, tests.MockNodeInfo.Alias, *nodeInfo.Alias)
	assert.Equal(t, tests.MockNodeInfo.Color, *nodeInfo.Color)
	assert.Equal(t, tests.MockNodeInfo.Pubkey, *nodeInfo.Pubkey)
	assert.Equal(t, tests.MockNodeInfo.Network, *nodeInfo.Network)
	assert.Equal(t, tests.MockNodeInfo.BlockHeight, *nodeInfo.BlockHeight)
	assert.Equal(t, tests.MockNodeInfo.BlockHash, *nodeInfo.BlockHash)
	assert.Contains(t, nodeInfo.Methods, "get_info")
	assert.Equal(t, []string{"payment_received", "payment_sent"}, nodeInfo.Notifications)
}
