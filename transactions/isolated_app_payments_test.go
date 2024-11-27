package transactions

import (
	"context"
	"testing"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendPaymentSync_IsolatedApp_NoBalance(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
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

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockLNClientTransaction.Invoice, nil, svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, NewInsufficientBalanceError())
	assert.Nil(t, transaction)
}

func TestSendPaymentSync_IsolatedApp_BalanceInsufficient(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
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

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	svc.DB.Create(&db.Transaction{
		AppId:      &app.ID,
		State:      constants.TRANSACTION_STATE_SETTLED,
		Type:       constants.TRANSACTION_TYPE_INCOMING,
		AmountMsat: 132000, // invoice is 123000 msat, but we also calculate fee reserves max of(10 sats or 1%)
	})

	mockEventConsumer := tests.NewMockEventConsumer()
	svc.EventPublisher.RegisterSubscriber(mockEventConsumer)

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockLNClientTransaction.Invoice, nil, svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, NewInsufficientBalanceError())
	assert.Nil(t, transaction)

	assert.Equal(t, 1, len(mockEventConsumer.GetConsumedEvents()))
	assert.Equal(t, "nwc_permission_denied", mockEventConsumer.GetConsumedEvents()[0].Event)
	assert.Equal(t, app.Name, mockEventConsumer.GetConsumedEvents()[0].Properties.(map[string]interface{})["app_name"])
	assert.Equal(t, constants.ERROR_INSUFFICIENT_BALANCE, mockEventConsumer.GetConsumedEvents()[0].Properties.(map[string]interface{})["code"])
	expectedMessage := NewInsufficientBalanceError().Error() + " te" // invoice description is "te" in the mock invoice
	assert.Equal(t, expectedMessage, mockEventConsumer.GetConsumedEvents()[0].Properties.(map[string]interface{})["message"])
}

func TestSendPaymentSync_IsolatedApp_BalanceSufficient(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
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

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	svc.DB.Create(&db.Transaction{
		AppId:      &app.ID,
		State:      constants.TRANSACTION_STATE_SETTLED,
		Type:       constants.TRANSACTION_TYPE_INCOMING,
		AmountMsat: 133000, // invoice is 123000 msat, but we also calculate fee reserves max of(10 sats or 1%)
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockLNClientTransaction.Invoice, nil, svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, "123preimage", *transaction.Preimage)
	assert.Equal(t, app.ID, *transaction.AppId)
	assert.Equal(t, dbRequestEvent.ID, *transaction.RequestEventId)
}

func TestSendPaymentSync_IsolatedApp_BalanceInsufficient_OutstandingPayment(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
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

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	svc.DB.Create(&db.Transaction{
		AppId:      &app.ID,
		State:      constants.TRANSACTION_STATE_SETTLED,
		Type:       constants.TRANSACTION_TYPE_INCOMING,
		AmountMsat: 133000, // invoice is 123000 msat, but we also calculate fee reserves max of(10 sats or 1%)
	})

	svc.DB.Create(&db.Transaction{
		AppId:      &app.ID,
		State:      constants.TRANSACTION_STATE_PENDING,
		Type:       constants.TRANSACTION_TYPE_OUTGOING,
		AmountMsat: 1000,
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockLNClientTransaction.Invoice, nil, svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, NewInsufficientBalanceError())
	assert.Nil(t, transaction)
}

func TestSendPaymentSync_IsolatedApp_BalanceInsufficient_SettledPayment(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
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

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	svc.DB.Create(&db.Transaction{
		AppId:      &app.ID,
		State:      constants.TRANSACTION_STATE_SETTLED,
		Type:       constants.TRANSACTION_TYPE_INCOMING,
		AmountMsat: 133000, // invoice is 123000 msat, but we also calculate fee reserves max of(10 sats or 1%)
	})

	svc.DB.Create(&db.Transaction{
		AppId:      &app.ID,
		State:      constants.TRANSACTION_STATE_SETTLED,
		Type:       constants.TRANSACTION_TYPE_OUTGOING,
		AmountMsat: 1000,
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockLNClientTransaction.Invoice, nil, svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, NewInsufficientBalanceError())
	assert.Nil(t, transaction)
}

func TestSendPaymentSync_IsolatedApp_BalanceSufficient_UnrelatedPayment(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
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

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	svc.DB.Create(&db.Transaction{
		AppId:      &app.ID,
		State:      constants.TRANSACTION_STATE_SETTLED,
		Type:       constants.TRANSACTION_TYPE_INCOMING,
		AmountMsat: 133000, // invoice is 123000 msat, but we also calculate fee reserves max of(10 sats or 1%)
	})
	svc.DB.Create(&db.Transaction{
		AppId:      nil, // unrelated to this app
		State:      constants.TRANSACTION_STATE_SETTLED,
		Type:       constants.TRANSACTION_TYPE_OUTGOING,
		AmountMsat: 1000,
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockLNClientTransaction.Invoice, nil, svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, "123preimage", *transaction.Preimage)
	assert.Equal(t, app.ID, *transaction.AppId)
	assert.Equal(t, dbRequestEvent.ID, *transaction.RequestEventId)
}

func TestSendPaymentSync_IsolatedApp_BalanceSufficient_FailedPayment(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
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

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	svc.DB.Create(&db.Transaction{
		AppId:      &app.ID,
		State:      constants.TRANSACTION_STATE_SETTLED,
		Type:       constants.TRANSACTION_TYPE_INCOMING,
		AmountMsat: 133000, // invoice is 123000 msat, but we also calculate fee reserves max of(10 sats or 1%)
	})
	svc.DB.Create(&db.Transaction{
		AppId:      &app.ID,
		State:      constants.TRANSACTION_STATE_FAILED,
		Type:       constants.TRANSACTION_TYPE_OUTGOING,
		AmountMsat: 1000,
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockLNClientTransaction.Invoice, nil, svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, "123preimage", *transaction.Preimage)
	assert.Equal(t, app.ID, *transaction.AppId)
	assert.Equal(t, dbRequestEvent.ID, *transaction.RequestEventId)
}
