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

const nip47CancelHoldInvoiceJson = `
{
"method": "cancel_hold_invoice",
"params": {
"payment_hash": "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
}
}
`

const testCancelPaymentHash = "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

type cancelHoldInvoiceTestSetup struct {
	ctx            context.Context
	svc            *tests.TestService
	nip47Request   *models.Request
	app            *db.App
	dbRequestEvent *db.RequestEvent
	publishCalled  bool
	response       *models.Response
}

func setupCancelHoldInvoiceTest(t *testing.T, paymentHash string, initialTransactionState *string) *cancelHoldInvoiceTestSetup {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	// Not using defer svc.Remove() as it might be called before the test function finishes in some cases.
	// Caller should call svc.Remove()

	nip47Request := &models.Request{}
	requestJson := `
{
"method": "cancel_hold_invoice",
"params": {
"payment_hash": "` + paymentHash + `"
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

	if initialTransactionState != nil {
		transaction := &db.Transaction{
			AppId:       &app.ID,
			PaymentHash: paymentHash,
			Type:        constants.TRANSACTION_TYPE_INCOMING,
			State:       *initialTransactionState,
			AmountMsat:  1000,
		}
		err = svc.DB.Create(transaction).Error
		require.NoError(t, err)
	}

	dbRequestEvent := &db.RequestEvent{
		AppId: &app.ID,
	}
	err = svc.DB.Create(&dbRequestEvent).Error
	require.NoError(t, err)

	setup := &cancelHoldInvoiceTestSetup{
		ctx:            ctx,
		svc:            svc,
		nip47Request:   nip47Request,
		app:            app,
		dbRequestEvent: dbRequestEvent,
	}

	return setup
}

func (s *cancelHoldInvoiceTestSetup) TearDown() {
	s.svc.Remove()
}

func (s *cancelHoldInvoiceTestSetup) PublishResponse(response *models.Response, tags nostr.Tags) {
	s.publishCalled = true
	s.response = response
}

func TestHandleCancelHoldInvoiceEvent(t *testing.T) {
	initialState := constants.TRANSACTION_STATE_ACCEPTED
	setup := setupCancelHoldInvoiceTest(t, testCancelPaymentHash, &initialState)
	defer setup.TearDown()

	NewTestNip47Controller(setup.svc).
		HandleCancelHoldInvoiceEvent(setup.ctx, setup.nip47Request, setup.dbRequestEvent.ID, *setup.dbRequestEvent.AppId, setup.PublishResponse)

	assert.True(t, setup.publishCalled)
	assert.Nil(t, setup.response.Error)
	assert.Equal(t, &cancelHoldInvoiceResponse{}, setup.response.Result)

	var updatedTransaction db.Transaction
	err := setup.svc.DB.First(&updatedTransaction, "payment_hash = ?", testCancelPaymentHash).Error
	assert.NoError(t, err)
	assert.Equal(t, constants.TRANSACTION_STATE_FAILED, updatedTransaction.State)
}

func TestHandleCancelHoldInvoiceEvent_InvoiceNotFound(t *testing.T) {
	nonExistentPaymentHash := "nonexistentpaymenthashnonexistentpaymenthashnonexistentpaymenthash"
	setup := setupCancelHoldInvoiceTest(t, nonExistentPaymentHash, nil) // nil for initialTransactionState means no transaction created
	defer setup.TearDown()

	NewTestNip47Controller(setup.svc).
		HandleCancelHoldInvoiceEvent(setup.ctx, setup.nip47Request, setup.dbRequestEvent.ID, *setup.dbRequestEvent.AppId, setup.PublishResponse)

	assert.True(t, setup.publishCalled)
	require.NotNil(t, setup.response.Error)
	assert.Equal(t, constants.ERROR_NOT_FOUND, setup.response.Error.Code)
}
