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

	"crypto/sha256"
	"encoding/hex"

	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/tests"
)

const nip47SettleHoldInvoiceJson = `
{
"method": "settle_hold_invoice",
"params": {
"preimage": "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
}
}
`

const testSettlePreimage = "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
const testSettlePaymentHash = "b7e060a60bb7a82f536a73c17bde37a1b6cf5769ee4a8325bff76c55a95b6aa4"

type settleHoldInvoiceTestSetup struct {
	ctx            context.Context
	svc            *tests.TestService
	nip47Request   *models.Request
	app            *db.App
	dbRequestEvent *db.RequestEvent
	publishCalled  bool
	response       *models.Response
}

func setupSettleHoldInvoiceTest(t *testing.T, preimage string, paymentHashToCreate string, initialTransactionState string) *settleHoldInvoiceTestSetup {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)

	nip47Request := &models.Request{}
	requestJson := `
{
"method": "settle_hold_invoice",
"params": {
"preimage": "` + preimage + `"
}
}
`
	err = json.Unmarshal([]byte(requestJson), nip47Request)
	require.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	require.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId: app.ID,
		Scope: constants.MAKE_INVOICE_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
	require.NoError(t, err)

	if paymentHashToCreate != "" && initialTransactionState != "" {
		expiresAtVar := time.Now().Add(1 * time.Hour)
		appIDForTx := app.ID
		holdInvoice := &db.Transaction{
			AppId:       &appIDForTx,
			Type:        constants.TRANSACTION_TYPE_INCOMING,
			State:       initialTransactionState,
			PaymentHash: paymentHashToCreate,
			AmountMsat:  1000,
			ExpiresAt:   &expiresAtVar,
		}
		err = svc.DB.Create(holdInvoice).Error
		require.NoError(t, err)
	}

	dbRequestEvent := &db.RequestEvent{
		AppId: &app.ID,
	}
	err = svc.DB.Create(&dbRequestEvent).Error
	require.NoError(t, err)

	setup := &settleHoldInvoiceTestSetup{
		ctx:            ctx,
		svc:            svc,
		nip47Request:   nip47Request,
		app:            app,
		dbRequestEvent: dbRequestEvent,
	}
	return setup
}

func (s *settleHoldInvoiceTestSetup) TearDown() {
	s.svc.Remove()
}

func (s *settleHoldInvoiceTestSetup) PublishResponse(response *models.Response, tags nostr.Tags) {
	s.publishCalled = true
	s.response = response
}

func TestHandleSettleHoldInvoiceEvent(t *testing.T) {
	preimageBytesForCheck, err := hex.DecodeString(testSettlePreimage)
	require.NoError(t, err)
	calculatedHashBytesForCheck := sha256.Sum256(preimageBytesForCheck)
	calculatedPaymentHashForCheck := hex.EncodeToString(calculatedHashBytesForCheck[:])
	assert.Equal(t, testSettlePaymentHash, calculatedPaymentHashForCheck)

	setup := setupSettleHoldInvoiceTest(t, testSettlePreimage, testSettlePaymentHash, constants.TRANSACTION_STATE_ACCEPTED)
	defer setup.TearDown()

	controller := NewTestNip47Controller(setup.svc)

	controller.HandleSettleHoldInvoiceEvent(setup.ctx, setup.nip47Request, setup.dbRequestEvent.ID, *setup.dbRequestEvent.AppId, setup.PublishResponse)

	assert.True(t, setup.publishCalled)
	assert.Nil(t, setup.response.Error)
	assert.Equal(t, &settleHoldInvoiceResponse{}, setup.response.Result)

	var settledTx db.Transaction
	err = setup.svc.DB.First(&settledTx, "payment_hash = ?", testSettlePaymentHash).Error
	assert.NoError(t, err)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, settledTx.State)
	assert.NotNil(t, settledTx.Preimage)
	assert.Equal(t, testSettlePreimage, *settledTx.Preimage)
}

func TestHandleSettleHoldInvoiceEvent_InvalidPreimage(t *testing.T) {
	invalidPreimage := "invalidpreimageinvalidpreimageinvalidpreimageinvalidpreimageinvalid"
	setup := setupSettleHoldInvoiceTest(t, invalidPreimage, testSettlePaymentHash, constants.TRANSACTION_STATE_ACCEPTED)
	defer setup.TearDown()

	controller := NewTestNip47Controller(setup.svc)

	controller.HandleSettleHoldInvoiceEvent(setup.ctx, setup.nip47Request, setup.dbRequestEvent.ID, *setup.dbRequestEvent.AppId, setup.PublishResponse)

	assert.True(t, setup.publishCalled)
	require.NotNil(t, setup.response.Error)
	assert.Equal(t, constants.ERROR_INTERNAL, setup.response.Error.Code)
}
