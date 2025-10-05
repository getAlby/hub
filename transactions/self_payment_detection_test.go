package transactions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/tests"
)

func TestSendPaymentSync_SelfPaymentDetection_WithIncomingTransaction(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

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
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockInvoice, nil, nil, svc.LNClient, nil, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, transaction)
	assert.True(t, transaction.SelfPayment)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, uint64(123000), transaction.AmountMsat)
	assert.Equal(t, mockPreimage, *transaction.Preimage)
	assert.Zero(t, transaction.FeeMsat)
}

func TestSendPaymentSync_SelfPaymentDetection_WithoutIncomingTransaction(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	svc.LNClient.(*tests.MockLn).Pubkey = "03cbd788f5b22bd56e2714bff756372d2293504c064e03250ed16a4dd80ad70e2c"

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockInvoice, nil, nil, svc.LNClient, nil, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, transaction)
	assert.False(t, transaction.SelfPayment)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, uint64(123000), transaction.AmountMsat)
}

func TestSendPaymentSync_SelfPaymentDetection_DifferentPubkey(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	svc.LNClient.(*tests.MockLn).Pubkey = "different_pubkey_123"

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
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockInvoice, nil, nil, svc.LNClient, nil, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, transaction)
	assert.False(t, transaction.SelfPayment)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
}

func TestSendPaymentSync_SelfPaymentDetection_MultipleIncomingTransactions(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	svc.LNClient.(*tests.MockLn).Pubkey = "03cbd788f5b22bd56e2714bff756372d2293504c064e03250ed16a4dd80ad70e2c"

	mockPreimage := "123preimage"
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockInvoice,
		PaymentHash:    tests.MockPaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockInvoice,
		PaymentHash:    tests.MockPaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockInvoice, nil, nil, svc.LNClient, nil, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, transaction)
	assert.True(t, transaction.SelfPayment)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
}

func TestSendPaymentSync_SelfPaymentDetection_OutgoingTransactionOnly(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	svc.LNClient.(*tests.MockLn).Pubkey = "03cbd788f5b22bd56e2714bff756372d2293504c064e03250ed16a4dd80ad70e2c"

	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_FAILED,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockInvoice,
		PaymentHash:    tests.MockPaymentHash,
		AmountMsat:     123000,
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockInvoice, nil, nil, svc.LNClient, nil, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, transaction)
	assert.False(t, transaction.SelfPayment)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
}

func TestSendPaymentSync_SelfPaymentDetection_DatabaseError(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	svc.LNClient.(*tests.MockLn).Pubkey = "03cbd788f5b22bd56e2714bff756372d2293504c064e03250ed16a4dd80ad70e2c"

	sqlDB, err := svc.DB.DB()
	require.NoError(t, err)
	sqlDB.Close()

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockInvoice, nil, nil, svc.LNClient, nil, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, transaction)
}

func TestSendPaymentSync_SelfPaymentDetection_WithApp(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	svc.LNClient.(*tests.MockLn).Pubkey = "03cbd788f5b22bd56e2714bff756372d2293504c064e03250ed16a4dd80ad70e2c"

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: constants.PAY_INVOICE_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
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
	transaction, err := transactionsService.SendPaymentSync(ctx, tests.MockInvoice, nil, nil, svc.LNClient, &app.ID, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, transaction)
	assert.True(t, transaction.SelfPayment)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, app.ID, *transaction.AppId)
}
