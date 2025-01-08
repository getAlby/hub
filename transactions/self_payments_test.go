package transactions

import (
	"context"
	"testing"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/db/queries"
	"github.com/getAlby/hub/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendPaymentSync_SelfPayment_NoAppToNoApp(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	// pubkey matches mock invoice = self payment
	svc.LNClient.(*tests.MockLn).Pubkey = "03cbd788f5b22bd56e2714bff756372d2293504c064e03250ed16a4dd80ad70e2c"

	mockPreimage := "123preimage"
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockInvoice,
		PaymentHash:    tests.MockPaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockInvoice, nil, nil, svc.LNClient, nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, mockPreimage, *transaction.Preimage)
	assert.True(t, transaction.SelfPayment)

	transactionType := constants.TRANSACTION_TYPE_INCOMING
	incomingTransaction, err := transactionsService.LookupTransaction(ctx, tests.MockPaymentHash, &transactionType, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), incomingTransaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, incomingTransaction.State)
	assert.Equal(t, mockPreimage, *incomingTransaction.Preimage)
	assert.True(t, incomingTransaction.SelfPayment)

	transactions := []db.Transaction{}
	result := svc.DB.Find(&transactions)
	assert.Equal(t, int64(2), result.RowsAffected)
}

func TestSendPaymentSync_SelfPayment_NoAppToIsolatedApp(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	// pubkey matches mock invoice = self payment
	svc.LNClient.(*tests.MockLn).Pubkey = "03cbd788f5b22bd56e2714bff756372d2293504c064e03250ed16a4dd80ad70e2c"

	app, _, err := tests.CreateApp(svc, "1.0")
	assert.NoError(t, err)
	app.Isolated = true
	svc.DB.Save(&app)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	mockPreimage := "123preimage"
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockInvoice,
		PaymentHash:    tests.MockPaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		AppId:          &app.ID,
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockInvoice, nil, nil, svc.LNClient, nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, mockPreimage, *transaction.Preimage)
	assert.True(t, transaction.SelfPayment)

	transactionType := constants.TRANSACTION_TYPE_INCOMING
	incomingTransaction, err := transactionsService.LookupTransaction(ctx, tests.MockPaymentHash, &transactionType, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), incomingTransaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, incomingTransaction.State)
	assert.Equal(t, mockPreimage, *incomingTransaction.Preimage)
	assert.Equal(t, app.ID, *incomingTransaction.AppId)
	assert.True(t, incomingTransaction.SelfPayment)

	transactions := []db.Transaction{}
	result := svc.DB.Find(&transactions)
	assert.Equal(t, int64(2), result.RowsAffected)
	// expect balance to be increased
	assert.Equal(t, int64(123000), queries.GetIsolatedBalance(svc.DB, app.ID))
}

func TestSendPaymentSync_SelfPayment_NoAppToApp(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	// pubkey matches mock invoice = self payment
	svc.LNClient.(*tests.MockLn).Pubkey = "03cbd788f5b22bd56e2714bff756372d2293504c064e03250ed16a4dd80ad70e2c"

	app, _, err := tests.CreateApp(svc, "1.0")
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	mockPreimage := "123preimage"
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockInvoice,
		PaymentHash:    tests.MockPaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		AppId:          &app.ID,
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockInvoice, nil, nil, svc.LNClient, nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, mockPreimage, *transaction.Preimage)
	assert.True(t, transaction.SelfPayment)

	transactionType := constants.TRANSACTION_TYPE_INCOMING
	incomingTransaction, err := transactionsService.LookupTransaction(ctx, tests.MockPaymentHash, &transactionType, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), incomingTransaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, incomingTransaction.State)
	assert.Equal(t, mockPreimage, *incomingTransaction.Preimage)
	assert.Equal(t, app.ID, *incomingTransaction.AppId)
	assert.True(t, incomingTransaction.SelfPayment)

	transactions := []db.Transaction{}
	result := svc.DB.Find(&transactions)
	assert.Equal(t, int64(2), result.RowsAffected)
}

func TestSendPaymentSync_SelfPayment_IsolatedAppToNoApp(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	// pubkey matches mock invoice = self payment
	svc.LNClient.(*tests.MockLn).Pubkey = "03cbd788f5b22bd56e2714bff756372d2293504c064e03250ed16a4dd80ad70e2c"

	app, _, err := tests.CreateApp(svc, "1.0")
	assert.NoError(t, err)
	app.Isolated = true
	err = svc.DB.Save(&app).Error
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: constants.PAY_INVOICE_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	// give the isolated app 133 sats
	svc.DB.Create(&db.Transaction{
		AppId:      &app.ID,
		State:      constants.TRANSACTION_STATE_SETTLED,
		Type:       constants.TRANSACTION_TYPE_INCOMING,
		AmountMsat: 133000, // invoice is 123000 msat, but we also calculate fee reserves max of(10 sats or 1%)
	})

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	mockPreimage := "123preimage"
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockInvoice,
		PaymentHash:    tests.MockPaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockInvoice, nil, nil, svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, mockPreimage, *transaction.Preimage)
	assert.Equal(t, app.ID, *transaction.AppId)
	assert.Equal(t, dbRequestEvent.ID, *transaction.RequestEventId)
	assert.True(t, transaction.SelfPayment)

	transactionType := constants.TRANSACTION_TYPE_INCOMING
	incomingTransaction, err := transactionsService.LookupTransaction(ctx, tests.MockPaymentHash, &transactionType, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), incomingTransaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, incomingTransaction.State)
	assert.Equal(t, mockPreimage, *incomingTransaction.Preimage)
	assert.True(t, incomingTransaction.SelfPayment)

	transactions := []db.Transaction{}
	result := svc.DB.Find(&transactions)
	assert.Equal(t, int64(3), result.RowsAffected)
	// expect balance to be decreased
	assert.Equal(t, int64(10000), queries.GetIsolatedBalance(svc.DB, app.ID))
}

func TestSendPaymentSync_SelfPayment_IsolatedAppToApp(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	// pubkey matches mock invoice = self payment
	svc.LNClient.(*tests.MockLn).Pubkey = "03cbd788f5b22bd56e2714bff756372d2293504c064e03250ed16a4dd80ad70e2c"

	app, _, err := tests.CreateApp(svc, "1.0")
	assert.NoError(t, err)
	app.Isolated = true
	err = svc.DB.Save(&app).Error
	assert.NoError(t, err)
	app2, _, err := tests.CreateApp(svc, "1.0")
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: constants.PAY_INVOICE_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	// give the isolated app 133 sats
	svc.DB.Create(&db.Transaction{
		AppId:      &app.ID,
		State:      constants.TRANSACTION_STATE_SETTLED,
		Type:       constants.TRANSACTION_TYPE_INCOMING,
		AmountMsat: 133000, // invoice is 123000 msat, but we also calculate fee reserves max of(10 sats or 1%)
	})

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	mockPreimage := "123preimage"
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockInvoice,
		PaymentHash:    tests.MockPaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		AppId:          &app2.ID,
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockInvoice, nil, nil, svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, mockPreimage, *transaction.Preimage)
	assert.Equal(t, app.ID, *transaction.AppId)
	assert.Equal(t, dbRequestEvent.ID, *transaction.RequestEventId)
	assert.True(t, transaction.SelfPayment)

	transactionType := constants.TRANSACTION_TYPE_INCOMING
	incomingTransaction, err := transactionsService.LookupTransaction(ctx, tests.MockPaymentHash, &transactionType, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), incomingTransaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, incomingTransaction.State)
	assert.Equal(t, mockPreimage, *incomingTransaction.Preimage)
	assert.Equal(t, app2.ID, *incomingTransaction.AppId)
	assert.True(t, incomingTransaction.SelfPayment)

	transactions := []db.Transaction{}
	result := svc.DB.Find(&transactions)
	assert.Equal(t, int64(3), result.RowsAffected)
	// expect balance to be decreased
	assert.Equal(t, int64(10000), queries.GetIsolatedBalance(svc.DB, app.ID))
}

func TestSendPaymentSync_SelfPayment_IsolatedAppToIsolatedApp(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	// pubkey matches mock invoice = self payment
	svc.LNClient.(*tests.MockLn).Pubkey = "03cbd788f5b22bd56e2714bff756372d2293504c064e03250ed16a4dd80ad70e2c"

	app, _, err := tests.CreateApp(svc, "1.0")
	assert.NoError(t, err)
	app.Isolated = true
	err = svc.DB.Save(&app).Error
	assert.NoError(t, err)
	app2, _, err := tests.CreateApp(svc, "1.0")
	assert.NoError(t, err)
	app2.Isolated = true
	err = svc.DB.Save(&app2).Error
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: constants.PAY_INVOICE_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	// give the isolated app 133 sats
	svc.DB.Create(&db.Transaction{
		AppId:      &app.ID,
		State:      constants.TRANSACTION_STATE_SETTLED,
		Type:       constants.TRANSACTION_TYPE_INCOMING,
		AmountMsat: 133000, // invoice is 123000 msat, but we also calculate fee reserves max of(10 sats or 1%)
	})

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	mockPreimage := "123preimage"
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockInvoice,
		PaymentHash:    tests.MockPaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		AppId:          &app2.ID,
	})

	mockEventConsumer := tests.NewMockEventConsumer()
	svc.EventPublisher.RegisterSubscriber(mockEventConsumer)

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockInvoice, nil, nil, svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, mockPreimage, *transaction.Preimage)
	assert.Equal(t, app.ID, *transaction.AppId)
	assert.Equal(t, dbRequestEvent.ID, *transaction.RequestEventId)
	assert.True(t, transaction.SelfPayment)

	transactionType := constants.TRANSACTION_TYPE_INCOMING
	incomingTransaction, err := transactionsService.LookupTransaction(ctx, tests.MockPaymentHash, &transactionType, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), incomingTransaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, incomingTransaction.State)
	assert.Equal(t, mockPreimage, *incomingTransaction.Preimage)
	assert.Equal(t, app2.ID, *incomingTransaction.AppId)
	assert.True(t, incomingTransaction.SelfPayment)

	transactions := []db.Transaction{}
	result := svc.DB.Find(&transactions)
	assert.Equal(t, int64(3), result.RowsAffected)
	// expect balance to be decreased
	assert.Equal(t, int64(10000), queries.GetIsolatedBalance(svc.DB, app.ID))

	// check notifications
	assert.Equal(t, 2, len(mockEventConsumer.GetConsumedEvents()))

	assert.Equal(t, "nwc_payment_sent", mockEventConsumer.GetConsumedEvents()[1].Event)
	settledTransaction := mockEventConsumer.GetConsumedEvents()[1].Properties.(*db.Transaction)
	assert.Equal(t, transaction.ID, settledTransaction.ID)

	assert.Equal(t, "nwc_payment_received", mockEventConsumer.GetConsumedEvents()[0].Event)
	receivedTransaction := mockEventConsumer.GetConsumedEvents()[0].Properties.(*db.Transaction)
	assert.Equal(t, incomingTransaction.ID, receivedTransaction.ID)
}

func TestSendPaymentSync_SelfPayment_IsolatedAppToSelf(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	// pubkey matches mock invoice = self payment
	svc.LNClient.(*tests.MockLn).Pubkey = "03cbd788f5b22bd56e2714bff756372d2293504c064e03250ed16a4dd80ad70e2c"

	app, _, err := tests.CreateApp(svc, "1.0")
	assert.NoError(t, err)
	app.Isolated = true
	err = svc.DB.Save(&app).Error
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: constants.PAY_INVOICE_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	// give the isolated app 133 sats
	svc.DB.Create(&db.Transaction{
		AppId:      &app.ID,
		State:      constants.TRANSACTION_STATE_SETTLED,
		Type:       constants.TRANSACTION_TYPE_INCOMING,
		AmountMsat: 133000, // invoice is 123000 msat, but we also calculate fee reserves max of(10 sats or 1%)
	})

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	mockPreimage := "123preimage"
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockInvoice,
		PaymentHash:    tests.MockPaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		AppId:          &app.ID,
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockInvoice, nil, nil, svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, mockPreimage, *transaction.Preimage)
	assert.Equal(t, app.ID, *transaction.AppId)
	assert.Equal(t, dbRequestEvent.ID, *transaction.RequestEventId)
	assert.True(t, transaction.SelfPayment)

	transactionType := constants.TRANSACTION_TYPE_INCOMING
	incomingTransaction, err := transactionsService.LookupTransaction(ctx, tests.MockPaymentHash, &transactionType, svc.LNClient, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), incomingTransaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, incomingTransaction.State)
	assert.Equal(t, mockPreimage, *incomingTransaction.Preimage)
	assert.Equal(t, app.ID, *incomingTransaction.AppId)
	assert.True(t, incomingTransaction.SelfPayment)

	transactions := []db.Transaction{}
	result := svc.DB.Find(&transactions)
	assert.Equal(t, int64(3), result.RowsAffected)

	// expect balance to be unchanged
	assert.Equal(t, int64(133000), queries.GetIsolatedBalance(svc.DB, app.ID))
}
