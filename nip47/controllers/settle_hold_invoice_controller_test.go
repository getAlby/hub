package controllers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/nbd-wtf/go-nostr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"time"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"

	// lnclient import removed as it's not directly used after mock setup
	"crypto/sha256"
	"encoding/hex"

	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/tests"
	"github.com/getAlby/hub/transactions"
)

const nip47SettleHoldInvoiceJson = `
{
"method": "settle_hold_invoice",
"params": {
"preimage": "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
}
}
`

const testPreimage = "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
const testPaymentHash = "b7e060a60bb7a82f536a73c17bde37a1b6cf5769ee4a8325bff76c55a95b6aa4" // Corrected sha256 of testPreimage

func TestHandleSettleHoldInvoiceEvent(t *testing.T) {
	// Verify that testPaymentHash is correctly derived from testPreimage
	preimageBytesForCheck, err := hex.DecodeString(testPreimage)
	require.NoError(t, err, "Test setup: failed to decode testPreimage for hash verification")
	calculatedHashBytesForCheck := sha256.Sum256(preimageBytesForCheck)
	calculatedPaymentHashForCheck := hex.EncodeToString(calculatedHashBytesForCheck[:])
	assert.Equal(t, testPaymentHash, calculatedPaymentHashForCheck, "Test setup: testPaymentHash constant does not match calculated hash of testPreimage")

	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(nip47SettleHoldInvoiceJson), nip47Request)
	require.NoError(t, err)

	// Unmarshal params into a map to set the preimage
	var params map[string]interface{}
	err = json.Unmarshal(nip47Request.Params, &params)
	require.NoError(t, err)
	params["preimage"] = testPreimage
	rawParams, err := json.Marshal(params)
	require.NoError(t, err)
	nip47Request.Params = rawParams

	app, _, err := tests.CreateApp(svc)
	require.NoError(t, err)

	// Grant the hold_invoice scope to the app
	appPermission := &db.AppPermission{
		AppId: app.ID,
		Scope: constants.HOLD_INVOICES_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
	require.NoError(t, err)

	// Create a dummy hold invoice transaction
	preimageVar := testPreimage // Use a variable to take its address
	expiresAtVar := time.Now().Add(1 * time.Hour)
	appIDForTx := app.ID // Use a separate variable for the transaction AppId pointer
	holdInvoice := &db.Transaction{
		AppId:       &appIDForTx, // Assuming db.Transaction.AppId is *uint
		Type:        constants.TRANSACTION_TYPE_INCOMING,
		State:       constants.TRANSACTION_STATE_ACCEPTED, // Hold invoices are 'ACCEPTED' until settled or canceled
		PaymentHash: testPaymentHash,
		Preimage:    &preimageVar, // Store preimage to allow lookup by it
		AmountMsat:  1000,
		ExpiresAt:   &expiresAtVar,
	}
	err = svc.DB.Create(holdInvoice).Error
	require.NoError(t, err)

	appID := app.ID // Use a variable to take its address
	dbRequestEvent := &db.RequestEvent{
		AppId: &appID,
	}
	err = svc.DB.Create(&dbRequestEvent).Error
	require.NoError(t, err)

	// svc.LNClient is already a *tests.MockLn, so no specific mock setup is needed here for SettleHoldInvoice
	// as the default mock implementation returns nil, which is the success case.

	var publishedResponse *models.Response
	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	controller := NewTestNip47Controller(svc)
	// Ensure the transactionsService within the controller uses the EventPublisher from the TestService
	controller.transactionsService = transactions.NewTransactionsService(svc.DB, svc.EventPublisher) // Explicitly using svc.EventPublisher

	// Log the transaction we expect to find
	var foundTxBeforeCall db.Transaction
	err = svc.DB.Where("payment_hash = ? AND type = ? AND state = ?", testPaymentHash, constants.TRANSACTION_TYPE_INCOMING, constants.TRANSACTION_STATE_ACCEPTED).First(&foundTxBeforeCall).Error
	if err != nil {
		t.Logf("DEBUG: Transaction NOT FOUND in DB before calling HandleSettleHoldInvoiceEvent. PaymentHash: %s, Error: %v", testPaymentHash, err)
	} else {
		t.Logf("DEBUG: Transaction FOUND in DB before calling HandleSettleHoldInvoiceEvent. ID: %d, PaymentHash: %s, State: %s", foundTxBeforeCall.ID, foundTxBeforeCall.PaymentHash, foundTxBeforeCall.State)
	}
	var allTxs []db.Transaction
	svc.DB.Find(&allTxs)
	t.Logf("DEBUG: All transactions in DB before call: %+v", allTxs)

	controller.HandleSettleHoldInvoiceEvent(ctx, nip47Request, dbRequestEvent.ID, *dbRequestEvent.AppId, publishResponse)

	assert.Nil(t, publishedResponse.Error, "Expected no error, got: %v", publishedResponse.Error)
	// Successful response for settle_hold_invoice has an empty result object
	assert.Equal(t, &settleHoldInvoiceResponse{}, publishedResponse.Result)
	// mockLnClient.AssertExpectations(t) removed as MockLn is not a testify mock
}
