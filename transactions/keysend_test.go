package transactions

import (
	"context"
	"testing"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/tests"
	"github.com/stretchr/testify/assert"
)

func TestSendKeysend(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	transactionsService := NewTransactionsService(svc.DB)
	transaction, err := transactionsService.SendKeysend(ctx, uint64(1000), "fake destination", []lnclient.TLVRecord{}, svc.LNClient, nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, `{"destination":"fake destination","tlv_records":[]}`, transaction.Metadata)
	assert.Equal(t, uint64(1000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_TYPE_OUTGOING, transaction.Type)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Zero(t, *transaction.FeeReserveMsat)
}

func TestSendKeysend_App_NoPermission(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	transactionsService := NewTransactionsService(svc.DB)
	transaction, err := transactionsService.SendKeysend(ctx, uint64(1000), "fake destination", []lnclient.TLVRecord{}, svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.Error(t, err)
	assert.Equal(t, "app does not have pay_invoice scope", err.Error())
	assert.Nil(t, transaction)
}

func TestSendKeysend_App_WithPermission(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: constants.PAY_INVOICE_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	transactionsService := NewTransactionsService(svc.DB)
	transaction, err := transactionsService.SendKeysend(ctx, uint64(1000), "fake destination", []lnclient.TLVRecord{}, svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.NoError(t, err)
	assert.Equal(t, `{"destination":"fake destination","tlv_records":[]}`, transaction.Metadata)
	assert.Equal(t, uint64(1000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_TYPE_OUTGOING, transaction.Type)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, app.ID, *transaction.AppId)
	assert.Equal(t, dbRequestEvent.ID, *transaction.RequestEventId)
	assert.Zero(t, *transaction.FeeReserveMsat)
}

func TestSendKeysend_App_BudgetExceeded(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId:        app.ID,
		App:          *app,
		Scope:        constants.PAY_INVOICE_SCOPE,
		MaxAmountSat: 10, // not enough for the fee reserve
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	transactionsService := NewTransactionsService(svc.DB)
	transaction, err := transactionsService.SendKeysend(ctx, uint64(1000), "fake destination", []lnclient.TLVRecord{}, svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.ErrorIs(t, err, NewQuotaExceededError())
	assert.Nil(t, transaction)
}
func TestSendKeysend_App_BudgetNotExceeded(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId:        app.ID,
		App:          *app,
		Scope:        constants.PAY_INVOICE_SCOPE,
		MaxAmountSat: 11, // fee reserve (10) + keysend amount (1)
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	transactionsService := NewTransactionsService(svc.DB)
	transaction, err := transactionsService.SendKeysend(ctx, uint64(1000), "fake destination", []lnclient.TLVRecord{}, svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.NoError(t, err)
	assert.Equal(t, `{"destination":"fake destination","tlv_records":[]}`, transaction.Metadata)
	assert.Equal(t, uint64(1000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_TYPE_OUTGOING, transaction.Type)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, app.ID, *transaction.AppId)
	assert.Equal(t, dbRequestEvent.ID, *transaction.RequestEventId)
	assert.Zero(t, *transaction.FeeReserveMsat)
}

func TestSendKeysend_App_BalanceExceeded(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)
	app.Isolated = true
	svc.DB.Save(&app)

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: constants.PAY_INVOICE_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	svc.DB.Create(&db.Transaction{
		AppId:      &app.ID,
		State:      constants.TRANSACTION_STATE_SETTLED,
		Type:       constants.TRANSACTION_TYPE_INCOMING,
		AmountMsat: 10000, // invoice is 1000 msat, but we also calculate fee reserves max of(10 sats or 1%)
	})

	transactionsService := NewTransactionsService(svc.DB)
	transaction, err := transactionsService.SendKeysend(ctx, uint64(1000), "fake destination", []lnclient.TLVRecord{}, svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.ErrorIs(t, err, NewInsufficientBalanceError())
	assert.Nil(t, transaction)
}

func TestSendKeysend_App_BalanceSufficient(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)
	app.Isolated = true
	svc.DB.Save(&app)

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: constants.PAY_INVOICE_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	svc.DB.Create(&db.Transaction{
		AppId:      &app.ID,
		State:      constants.TRANSACTION_STATE_SETTLED,
		Type:       constants.TRANSACTION_TYPE_INCOMING,
		AmountMsat: 11000, // invoice is 1000 msat, but we also calculate fee reserves max of(10 sats or 1%)
	})

	transactionsService := NewTransactionsService(svc.DB)
	transaction, err := transactionsService.SendKeysend(ctx, uint64(1000), "fake destination", []lnclient.TLVRecord{}, svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.NoError(t, err)
	assert.Equal(t, `{"destination":"fake destination","tlv_records":[]}`, transaction.Metadata)
	assert.Equal(t, uint64(1000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_TYPE_OUTGOING, transaction.Type)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, app.ID, *transaction.AppId)
	assert.Equal(t, dbRequestEvent.ID, *transaction.RequestEventId)
	assert.Zero(t, *transaction.FeeReserveMsat)
}

func TestSendKeysend_TLVs(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	transactionsService := NewTransactionsService(svc.DB)
	transaction, err := transactionsService.SendKeysend(ctx, uint64(1000), "fake destination", []lnclient.TLVRecord{
		{
			Type:  7629169,
			Value: "48656C6C6F2C20776F726C64",
		},
	}, svc.LNClient, nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, `{"destination":"fake destination","tlv_records":[{"type":7629169,"value":"48656C6C6F2C20776F726C64"}]}`, transaction.Metadata)
	assert.Equal(t, uint64(1000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_TYPE_OUTGOING, transaction.Type)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Zero(t, *transaction.FeeReserveMsat)
}
