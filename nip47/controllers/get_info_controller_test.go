package controllers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/nbd-wtf/go-nostr"
	"github.com/stretchr/testify/assert"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/nip47/permissions"
	"github.com/getAlby/hub/tests"
)

const nip47GetInfoJson = `
{
	"method": "get_info"
}
`

// TODO: info event should always return something
func TestHandleGetInfoEvent_NoPermission(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

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
		Scope:     permissions.GET_BALANCE_SCOPE,
		ExpiresAt: nil,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	checkPermission := func(amountMsat uint64) *models.Response {
		return &models.Response{
			ResultType: nip47Request.Method,
			Error: &models.Error{
				Code: models.ERROR_RESTRICTED,
			},
		}
	}

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)

	NewGetInfoController(permissionsSvc, svc.LNClient).
		HandleGetInfoEvent(ctx, nip47Request, dbRequestEvent.ID, app, checkPermission, publishResponse)

	assert.Nil(t, publishedResponse.Error)
	nodeInfo := publishedResponse.Result.(*getInfoResponse)
	assert.Empty(t, nodeInfo.Alias)
	assert.Empty(t, nodeInfo.Color)
	assert.Empty(t, nodeInfo.Pubkey)
	assert.Empty(t, nodeInfo.Network)
	assert.Empty(t, nodeInfo.BlockHeight)
	assert.Empty(t, nodeInfo.BlockHash)
	assert.Equal(t, []string{"get_balance"}, nodeInfo.Methods)
	assert.Equal(t, []string{}, nodeInfo.Notifications)
}

func TestHandleGetInfoEvent_WithPermission(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

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
		Scope:     permissions.GET_INFO_SCOPE,
		ExpiresAt: nil,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	checkPermission := func(amountMsat uint64) *models.Response {
		return nil
	}

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)

	NewGetInfoController(permissionsSvc, svc.LNClient).
		HandleGetInfoEvent(ctx, nip47Request, dbRequestEvent.ID, app, checkPermission, publishResponse)

	assert.Nil(t, publishedResponse.Error)
	nodeInfo := publishedResponse.Result.(*getInfoResponse)
	assert.Equal(t, tests.MockNodeInfo.Alias, nodeInfo.Alias)
	assert.Equal(t, tests.MockNodeInfo.Color, nodeInfo.Color)
	assert.Equal(t, tests.MockNodeInfo.Pubkey, nodeInfo.Pubkey)
	assert.Equal(t, tests.MockNodeInfo.Network, nodeInfo.Network)
	assert.Equal(t, tests.MockNodeInfo.BlockHeight, nodeInfo.BlockHeight)
	assert.Equal(t, tests.MockNodeInfo.BlockHash, nodeInfo.BlockHash)
	assert.Equal(t, []string{"get_info"}, nodeInfo.Methods)
	assert.Equal(t, []string{}, nodeInfo.Notifications)
}

func TestHandleGetInfoEvent_WithNotifications(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

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
		Scope:     permissions.GET_INFO_SCOPE,
		ExpiresAt: nil,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	// TODO: AppPermission RequestMethod needs to change to scope
	appPermission = &db.AppPermission{
		AppId:     app.ID,
		Scope:     permissions.NOTIFICATIONS_SCOPE,
		ExpiresAt: nil,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	checkPermission := func(amountMsat uint64) *models.Response {
		return nil
	}

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)

	NewGetInfoController(permissionsSvc, svc.LNClient).
		HandleGetInfoEvent(ctx, nip47Request, dbRequestEvent.ID, app, checkPermission, publishResponse)

	assert.Nil(t, publishedResponse.Error)
	nodeInfo := publishedResponse.Result.(*getInfoResponse)
	assert.Equal(t, tests.MockNodeInfo.Alias, nodeInfo.Alias)
	assert.Equal(t, tests.MockNodeInfo.Color, nodeInfo.Color)
	assert.Equal(t, tests.MockNodeInfo.Pubkey, nodeInfo.Pubkey)
	assert.Equal(t, tests.MockNodeInfo.Network, nodeInfo.Network)
	assert.Equal(t, tests.MockNodeInfo.BlockHeight, nodeInfo.BlockHeight)
	assert.Equal(t, tests.MockNodeInfo.BlockHash, nodeInfo.BlockHash)
	assert.Equal(t, []string{"get_info"}, nodeInfo.Methods)
	assert.Equal(t, []string{"payment_received", "payment_sent"}, nodeInfo.Notifications)
}
