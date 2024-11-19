package transactions

import (
	"context"
	"testing"
	"time"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendPaymentSync_App_NoPermission(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockLNClientTransaction.Invoice, svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.Error(t, err)
	assert.Equal(t, "app does not have pay_invoice scope", err.Error())
	assert.Nil(t, transaction)
}
func TestSendPaymentSync_App_WithPermission(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

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
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockLNClientTransaction.Invoice, svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, "123preimage", *transaction.Preimage)
	assert.Equal(t, app.ID, *transaction.AppId)
	assert.Equal(t, dbRequestEvent.ID, *transaction.RequestEventId)
}

func TestSendPaymentSync_App_BudgetExceeded(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId:        app.ID,
		App:          *app,
		Scope:        constants.PAY_INVOICE_SCOPE,
		MaxAmountSat: 1,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	mockEventConsumer := tests.NewMockEventConsumer()
	svc.EventPublisher.RegisterSubscriber(mockEventConsumer)

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockLNClientTransaction.Invoice, svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, NewQuotaExceededError())
	assert.Nil(t, transaction)

	assert.Equal(t, 1, len(mockEventConsumer.GetConsumeEvents()))
	assert.Equal(t, "nwc_permission_denied", mockEventConsumer.GetConsumeEvents()[0].Event)
	assert.Equal(t, app.Name, mockEventConsumer.GetConsumeEvents()[0].Properties.(map[string]interface{})["app_name"])
	assert.Equal(t, constants.ERROR_QUOTA_EXCEEDED, mockEventConsumer.GetConsumeEvents()[0].Properties.(map[string]interface{})["code"])
	expectedMessage := NewQuotaExceededError().Error() + " Invoice description: te" // invoice description is "te" in the mock invoice
	assert.Equal(t, expectedMessage, mockEventConsumer.GetConsumeEvents()[0].Properties.(map[string]interface{})["message"])
}

func TestSendPaymentSync_App_BudgetExceeded_SettledPayment(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId:        app.ID,
		App:          *app,
		Scope:        constants.PAY_INVOICE_SCOPE,
		MaxAmountSat: 133, // invoice is 123 sats, but we also calculate fee reserves max of(10 sats or 1%)
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	// 1 sat payment pushes app over the limit
	svc.DB.Create(&db.Transaction{
		AppId:      &app.ID,
		State:      constants.TRANSACTION_STATE_SETTLED,
		Type:       constants.TRANSACTION_TYPE_OUTGOING,
		AmountMsat: 1000,
		CreatedAt:  time.Now(),
	})

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockLNClientTransaction.Invoice, svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, NewQuotaExceededError())
	assert.Nil(t, transaction)
}
func TestSendPaymentSync_App_BudgetExceeded_UnsettledPayment(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId:        app.ID,
		App:          *app,
		Scope:        constants.PAY_INVOICE_SCOPE,
		MaxAmountSat: 133, // invoice is 123 sats, but we also calculate fee reserves max of(10 sats or 1%)
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	// 1 sat payment pushes app over the limit
	svc.DB.Create(&db.Transaction{
		AppId:      &app.ID,
		State:      constants.TRANSACTION_STATE_PENDING,
		Type:       constants.TRANSACTION_TYPE_OUTGOING,
		AmountMsat: 1000,
		CreatedAt:  time.Now(),
	})

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockLNClientTransaction.Invoice, svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.Error(t, err)
	assert.ErrorIs(t, err, NewQuotaExceededError())
	assert.Nil(t, transaction)
}

func TestSendPaymentSync_App_BudgetNotExceeded_FailedPayment(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId:        app.ID,
		App:          *app,
		Scope:        constants.PAY_INVOICE_SCOPE,
		MaxAmountSat: 133, // invoice is 123 sats, but we also calculate fee reserves max of(10 sats or 1%)
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	// 1 sat payment would push app over the limit, but it failed so its not counted
	svc.DB.Create(&db.Transaction{
		AppId:      &app.ID,
		State:      constants.TRANSACTION_STATE_FAILED,
		Type:       constants.TRANSACTION_TYPE_OUTGOING,
		AmountMsat: 1000,
		CreatedAt:  time.Now(),
	})

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockLNClientTransaction.Invoice, svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, "123preimage", *transaction.Preimage)
	assert.Equal(t, app.ID, *transaction.AppId)
	assert.Equal(t, dbRequestEvent.ID, *transaction.RequestEventId)
}
