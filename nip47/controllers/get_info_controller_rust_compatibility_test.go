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
	"github.com/getAlby/hub/tests"
)

func TestHandleGetInfoEvent_EmptyFieldsOmittedForRustClientCompatibility(t *testing.T) {
	// This test verifies the fix for Rust NWC client compatibility:
	// "When using custom permissions, empty fields should be omitted from JSON response 
	// instead of being included as empty strings which break Rust serde deserialization"
	
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

	// Delete the existing app permissions (the app was created with get_info scope)
	svc.DB.Exec("delete from app_permissions")

	// Create an app with ONLY pay_invoice permission (no get_info scope)
	// This means node info fields will be empty and should be omitted from JSON
	appPermission := &db.AppPermission{
		AppId:     app.ID,
		Scope:     constants.PAY_INVOICE_SCOPE, // Only pay_invoice permission
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

	// Get the actual JSON response that would be sent to clients
	jsonResponse, err := json.MarshalIndent(publishedResponse, "", "  ")
	assert.NoError(t, err)
	fmt.Printf("NWC get_info response JSON (pay_invoice only - fixed):\n%s\n", string(jsonResponse))

	// Convert to raw JSON object to check field presence
	var jsonObj map[string]interface{}
	err = json.Unmarshal(jsonResponse, &jsonObj)
	assert.NoError(t, err)
	
	result, ok := jsonObj["result"].(map[string]interface{})
	assert.True(t, ok, "result should be an object")

	// FIXED: Empty fields should be OMITTED from JSON (not present as empty strings)
	_, aliasExists := result["alias"]
	_, colorExists := result["color"] 
	_, pubkeyExists := result["pubkey"]
	_, networkExists := result["network"]
	_, blockHashExists := result["block_hash"]
	_, lud16Exists := result["lud16"]
	_, blockHeightExists := result["block_height"]

	// These fields should NOT be present when empty (omitempty behavior)
	assert.False(t, aliasExists || (result["alias"] != nil && result["alias"] != ""), 
		"alias should be omitted when empty")
	assert.False(t, colorExists || (result["color"] != nil && result["color"] != ""), 
		"color should be omitted when empty")
	assert.False(t, pubkeyExists || (result["pubkey"] != nil && result["pubkey"] != ""), 
		"pubkey should be omitted when empty")
	assert.False(t, networkExists || (result["network"] != nil && result["network"] != ""), 
		"network should be omitted when empty")
	assert.False(t, blockHashExists || (result["block_hash"] != nil && result["block_hash"] != ""), 
		"block_hash should be omitted when empty")
	assert.False(t, lud16Exists || (result["lud16"] != nil && result["lud16"] != ""), 
		"lud16 should be omitted when empty")
	assert.False(t, blockHeightExists || (result["block_height"] != nil && result["block_height"].(float64) != 0), 
		"block_height should be omitted when zero")

	// These required fields MUST always be present
	assert.True(t, result["methods"] != nil, "methods must always be present")
	assert.True(t, result["notifications"] != nil, "notifications must always be present") 

	// Verify the struct values are still empty (internal state)
	assert.Equal(t, "", nodeInfo.Alias)
	assert.Equal(t, "", nodeInfo.Color)  
	assert.Equal(t, "", nodeInfo.Pubkey)
	assert.Equal(t, "", nodeInfo.Network)
	assert.Equal(t, "", nodeInfo.BlockHash)
	assert.Equal(t, "", nodeInfo.LightningAddress)
	assert.Equal(t, uint32(0), nodeInfo.BlockHeight)

	// Verify required fields are populated correctly
	assert.Contains(t, nodeInfo.Methods, models.GET_INFO_METHOD)
	assert.Contains(t, nodeInfo.Methods, constants.PAY_INVOICE_SCOPE)
	assert.Equal(t, []string{}, nodeInfo.Notifications)
}