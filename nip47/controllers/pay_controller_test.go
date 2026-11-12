package controllers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/getAlby/go-nostr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/tests"
)

const nip47PayJson = `
{
	"method": "pay",
	"params": {
		"payment": "bitcoin:?lightning=` + tests.MockInvoice + `"
	}
}
`

const nip47PayZeroAmountNoAmountJson = `
{
	"method": "pay",
	"params": {
		"payment": "bitcoin:?lightning=` + tests.MockZeroAmountInvoice + `"
	}
}
`

const nip47PayBolt12Json = `
{
	"method": "pay",
	"params": {
		"payment": "bitcoin:?lno=lno1zrxq8pjw7qjlm68mtp7e3yvxee4y5xrgjhhyf2fxhlphpckrvevh50u0qf"
	}
}
`

const nip47PayInvalidUriJson = `
{
	"method": "pay",
	"params": {
		"payment": "lightning:lnbc123"
	}
}
`

const nip47PayPayerNoteJson = `
{
	"method": "pay",
	"params": {
		"payment": "bitcoin:?lightning=` + tests.MockInvoice + `",
		"payer_note": "hello"
	}
}
`

const nip47PayRequiredParamJson = `
{
	"method": "pay",
	"params": {
		"payment": "bitcoin:?lightning=` + tests.MockInvoice + `&req-something=1"
	}
}
`

func setupPayTest(t *testing.T) (*tests.TestService, *db.App, *db.RequestEvent) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	t.Cleanup(svc.Remove)

	app, _, err := tests.CreateApp(svc)
	require.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: constants.PAY_INVOICE_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
	require.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	require.NoError(t, err)

	return svc, app, dbRequestEvent
}

func TestHandlePayEvent(t *testing.T) {
	ctx := context.TODO()
	svc, app, dbRequestEvent := setupPayTest(t)

	nip47Request := &models.Request{}
	err := json.Unmarshal([]byte(nip47PayJson), nip47Request)
	require.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandlePayEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	require.Nil(t, publishedResponse.Error)
	result := publishedResponse.Result.(payResult)
	assert.Equal(t, "settled", result.State)
	assert.Equal(t, "bolt11", result.InstructionType)
	assert.Equal(t, "123preimage", result.Preimage)
	assert.Equal(t, tests.MockPaymentHash, result.PaymentHash)
	assert.Equal(t, tests.MockPaymentHash, result.TransactionId)
	assert.Equal(t, uint64(123000), result.Amount)
	assert.NotNil(t, result.SettledAt)
}

func TestHandlePayEvent_ZeroAmountInvoiceWithoutAmount(t *testing.T) {
	ctx := context.TODO()
	svc, app, dbRequestEvent := setupPayTest(t)

	nip47Request := &models.Request{}
	err := json.Unmarshal([]byte(nip47PayZeroAmountNoAmountJson), nip47Request)
	require.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandlePayEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	assert.Nil(t, publishedResponse.Result)
	assert.Equal(t, constants.ERROR_BAD_REQUEST, publishedResponse.Error.Code)
}

func TestHandlePayEvent_Bolt12Unsupported(t *testing.T) {
	ctx := context.TODO()
	svc, app, dbRequestEvent := setupPayTest(t)

	nip47Request := &models.Request{}
	err := json.Unmarshal([]byte(nip47PayBolt12Json), nip47Request)
	require.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandlePayEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	assert.Nil(t, publishedResponse.Result)
	assert.Equal(t, constants.ERROR_UNSUPPORTED_PAYMENT_INSTRUCTION, publishedResponse.Error.Code)
}

func runPayTest(t *testing.T, requestJson string) *models.Response {
	t.Helper()
	ctx := context.TODO()
	svc, app, dbRequestEvent := setupPayTest(t)

	nip47Request := &models.Request{}
	err := json.Unmarshal([]byte(requestJson), nip47Request)
	require.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandlePayEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	require.NotNil(t, publishedResponse)
	return publishedResponse
}

func TestHandlePayEvent_WrongNetwork(t *testing.T) {
	// regtest invoice (lnbcrt...) on a signet mock node
	response := runPayTest(t, `
{
	"method": "pay",
	"params": {
		"payment": "bitcoin:?lightning=lnbcrt5u1pjuywzppp5h69dt59cypca2wxu69sw8ga0g39a3yx7dqug5nthrw3rcqgfdu4qdqqcqzzsxqyz5vqsp5gzlpzszyj2k30qmpme7jsfzr24wqlvt9xdmr7ay34lfelz050krs9qyyssq038x07nh8yuv8hdpjh5y8kqp7zcd62ql9na9xh7pla44htjyy02sz23q7qm2tza6ct4ypljk54w9k9qsrsu95usk8ce726ytep6vhhsq9mhf9a"
	}
}
`)
	assert.Nil(t, response.Result)
	assert.Equal(t, constants.ERROR_UNSUPPORTED_NETWORK, response.Error.Code)
}

func TestHandlePayEvent_ConflictingParamAmount(t *testing.T) {
	// MockInvoice has an amount of 123000 msat
	response := runPayTest(t, `
{
	"method": "pay",
	"params": {
		"payment": "bitcoin:?lightning=`+tests.MockInvoice+`",
		"amount": 999
	}
}
`)
	assert.Nil(t, response.Result)
	assert.Equal(t, constants.ERROR_BAD_REQUEST, response.Error.Code)
}

func TestHandlePayEvent_ConflictingUriAmount(t *testing.T) {
	// MockInvoice has an amount of 123000 msat, URI says 124000 msat
	response := runPayTest(t, `
{
	"method": "pay",
	"params": {
		"payment": "bitcoin:?lightning=`+tests.MockInvoice+`&amount=0.00000124"
	}
}
`)
	assert.Nil(t, response.Result)
	assert.Equal(t, constants.ERROR_BAD_REQUEST, response.Error.Code)
}

func TestHandlePayEvent_MatchingUriAmount(t *testing.T) {
	// MockInvoice has an amount of 123000 msat = 0.00000123 BTC
	response := runPayTest(t, `
{
	"method": "pay",
	"params": {
		"payment": "bitcoin:?lightning=`+tests.MockInvoice+`&amount=0.00000123"
	}
}
`)
	require.Nil(t, response.Error)
	result := response.Result.(payResult)
	assert.Equal(t, "settled", result.State)
	assert.Equal(t, uint64(123000), result.Amount)
}

func TestHandlePayEvent_ZeroAmountInvoiceWithUriAmount(t *testing.T) {
	// zero-amount invoice funded by the BIP-321 amount parameter (1234 msat)
	response := runPayTest(t, `
{
	"method": "pay",
	"params": {
		"payment": "bitcoin:?lightning=`+tests.MockZeroAmountInvoice+`&amount=0.00000001234"
	}
}
`)
	require.Nil(t, response.Error)
	result := response.Result.(payResult)
	assert.Equal(t, "settled", result.State)
	assert.Equal(t, uint64(1234), result.Amount)
}

func TestHandlePayEvent_PayerNoteUnsupported(t *testing.T) {
	ctx := context.TODO()
	svc, app, dbRequestEvent := setupPayTest(t)

	nip47Request := &models.Request{}
	err := json.Unmarshal([]byte(nip47PayPayerNoteJson), nip47Request)
	require.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandlePayEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	assert.Nil(t, publishedResponse.Result)
	assert.Equal(t, constants.ERROR_BAD_REQUEST, publishedResponse.Error.Code)
}

func TestHandlePayEvent_UnknownRequiredParam(t *testing.T) {
	ctx := context.TODO()
	svc, app, dbRequestEvent := setupPayTest(t)

	nip47Request := &models.Request{}
	err := json.Unmarshal([]byte(nip47PayRequiredParamJson), nip47Request)
	require.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandlePayEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	assert.Nil(t, publishedResponse.Result)
	assert.Equal(t, constants.ERROR_BAD_REQUEST, publishedResponse.Error.Code)
}

func TestHandlePayEvent_InvalidUri(t *testing.T) {
	ctx := context.TODO()
	svc, app, dbRequestEvent := setupPayTest(t)

	nip47Request := &models.Request{}
	err := json.Unmarshal([]byte(nip47PayInvalidUriJson), nip47Request)
	require.NoError(t, err)

	var publishedResponse *models.Response

	publishResponse := func(response *models.Response, tags nostr.Tags) {
		publishedResponse = response
	}

	NewTestNip47Controller(svc).
		HandlePayEvent(ctx, nip47Request, dbRequestEvent.ID, app, publishResponse)

	assert.Nil(t, publishedResponse.Result)
	assert.Equal(t, constants.ERROR_BAD_REQUEST, publishedResponse.Error.Code)
}
