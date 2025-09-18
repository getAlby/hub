package transactions

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/db/queries"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/tests"
)

func TestSendKeysend(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	mockEventConsumer := tests.NewMockEventConsumer()
	svc.EventPublisher.RegisterSubscriber(mockEventConsumer)

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendKeysend(uint64(1000), "fake destination", nil, "", svc.LNClient, nil, nil)
	assert.NoError(t, err)

	var metadata lnclient.Metadata
	err = json.Unmarshal(transaction.Metadata, &metadata)
	assert.NoError(t, err)

	assert.Equal(t, "fake destination", metadata["destination"])
	assert.Nil(t, metadata["tlv_records"])
	assert.Equal(t, uint64(1000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_TYPE_OUTGOING, transaction.Type)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Zero(t, transaction.FeeReserveMsat)
	assert.NotNil(t, transaction.Preimage)
	assert.Equal(t, 64, len(*transaction.Preimage))

	assert.Equal(t, 1, len(mockEventConsumer.GetConsumedEvents()))
	assert.Equal(t, "nwc_payment_sent", mockEventConsumer.GetConsumedEvents()[0].Event)
	settledTransaction := mockEventConsumer.GetConsumedEvents()[0].Properties.(*db.Transaction)
	assert.Equal(t, transaction, settledTransaction)
}
func TestSendKeysend_CustomPreimage(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	customPreimage := "018465013e2337234a7e5530a21c4a8cf70d84231f4a8ff0b1e2cce3cb2bd03b"
	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendKeysend(uint64(1000), "fake destination", nil, customPreimage, svc.LNClient, nil, nil)
	assert.NoError(t, err)

	var metadata lnclient.Metadata
	err = json.Unmarshal(transaction.Metadata, &metadata)
	assert.NoError(t, err)

	assert.Equal(t, "fake destination", metadata["destination"])
	assert.Nil(t, metadata["tlv_records"])
	assert.Equal(t, uint64(1000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_TYPE_OUTGOING, transaction.Type)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Zero(t, transaction.FeeReserveMsat)
	assert.NotNil(t, transaction.Preimage)
	assert.Equal(t, customPreimage, *transaction.Preimage)
}

func TestSendKeysend_App_NoPermission(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendKeysend(uint64(1000), "fake destination", nil, "", svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.Error(t, err)
	assert.Equal(t, "app does not have pay_invoice scope", err.Error())
	assert.Nil(t, transaction)
}

func TestSendKeysend_App_WithPermission(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

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

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendKeysend(uint64(1000), "fake destination", nil, "", svc.LNClient, &app.ID, &dbRequestEvent.ID)
	assert.NoError(t, err)

	var metadata lnclient.Metadata
	err = json.Unmarshal(transaction.Metadata, &metadata)
	assert.NoError(t, err)

	assert.Equal(t, "fake destination", metadata["destination"])
	assert.Nil(t, metadata["tlv_records"])
	assert.Equal(t, uint64(1000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_TYPE_OUTGOING, transaction.Type)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, app.ID, *transaction.AppId)
	assert.Equal(t, dbRequestEvent.ID, *transaction.RequestEventId)
	assert.Zero(t, transaction.FeeReserveMsat)
	assert.NotNil(t, transaction.Preimage)
	assert.Equal(t, 64, len(*transaction.Preimage))
}

func TestSendKeysend_App_BudgetExceeded(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

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

	mockEventConsumer := tests.NewMockEventConsumer()
	svc.EventPublisher.RegisterSubscriber(mockEventConsumer)

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendKeysend(uint64(1000), "fake destination", nil, "", svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.ErrorIs(t, err, NewQuotaExceededError())
	assert.Nil(t, transaction)

	assert.Equal(t, 1, len(mockEventConsumer.GetConsumedEvents()))
	assert.Equal(t, "nwc_permission_denied", mockEventConsumer.GetConsumedEvents()[0].Event)
	assert.Equal(t, app.Name, mockEventConsumer.GetConsumedEvents()[0].Properties.(map[string]interface{})["app_name"])
	assert.Equal(t, constants.ERROR_QUOTA_EXCEEDED, mockEventConsumer.GetConsumedEvents()[0].Properties.(map[string]interface{})["code"])
	assert.Equal(t, NewQuotaExceededError().Error(), mockEventConsumer.GetConsumedEvents()[0].Properties.(map[string]interface{})["message"])
}
func TestSendKeysend_App_BudgetNotExceeded(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

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

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendKeysend(uint64(1000), "fake destination", nil, "", svc.LNClient, &app.ID, &dbRequestEvent.ID)
	assert.NoError(t, err)

	var metadata lnclient.Metadata
	err = json.Unmarshal(transaction.Metadata, &metadata)
	assert.NoError(t, err)

	assert.Equal(t, "fake destination", metadata["destination"])
	assert.Nil(t, metadata["tlv_records"])
	assert.Equal(t, uint64(1000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_TYPE_OUTGOING, transaction.Type)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, app.ID, *transaction.AppId)
	assert.Equal(t, dbRequestEvent.ID, *transaction.RequestEventId)
	assert.Zero(t, transaction.FeeReserveMsat)
	assert.NotNil(t, transaction.Preimage)
	assert.Equal(t, 64, len(*transaction.Preimage))
}

func TestSendKeysend_App_BalanceExceeded(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

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

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendKeysend(uint64(1000), "fake destination", nil, "", svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.ErrorIs(t, err, NewInsufficientBalanceError())
	assert.Nil(t, transaction)
}

func TestSendKeysend_App_BalanceSufficient(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

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

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendKeysend(uint64(1000), "fake destination", nil, "", svc.LNClient, &app.ID, &dbRequestEvent.ID)
	assert.NoError(t, err)

	var metadata lnclient.Metadata
	err = json.Unmarshal(transaction.Metadata, &metadata)
	assert.NoError(t, err)

	assert.Equal(t, "fake destination", metadata["destination"])
	assert.Nil(t, metadata["tlv_records"])
	assert.Equal(t, uint64(1000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_TYPE_OUTGOING, transaction.Type)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, app.ID, *transaction.AppId)
	assert.Equal(t, dbRequestEvent.ID, *transaction.RequestEventId)
	assert.NotNil(t, transaction.Preimage)
	assert.Equal(t, 64, len(*transaction.Preimage))
	assert.Zero(t, transaction.FeeReserveMsat)
}

func TestSendKeysend_TLVs(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendKeysend(uint64(1000), "fake destination", []lnclient.TLVRecord{
		{
			Type:  7629169,
			Value: "7b22616374696f6e223a22626f6f7374222c2276616c75655f6d736174223a313030302c2276616c75655f6d7361745f746f74616c223a313030302c226170705f6e616d65223a22e29aa1205765624c4e2044656d6f222c226170705f76657273696f6e223a22312e30222c22666565644944223a2268747470733a2f2f66656564732e706f6463617374696e6465782e6f72672f706332302e786d6c222c22706f6463617374223a22506f6463617374696e6720322e30222c22657069736f6465223a22457069736f6465203130343a2041204e65772044756d70222c227473223a32312c226e616d65223a22e29aa1205765624c4e2044656d6f222c2273656e6465725f6e616d65223a225361746f736869204e616b616d6f746f222c226d657373616765223a22476f20706f6463617374696e6721227d",
		},
	}, "", svc.LNClient, nil, nil)
	assert.NoError(t, err)

	var metadata lnclient.Metadata
	err = json.Unmarshal(transaction.Metadata, &metadata)
	assert.NoError(t, err)

	assert.Equal(t, "fake destination", metadata["destination"])
	assert.NotNil(t, metadata["tlv_records"])

	var boostagram Boostagram
	err = json.Unmarshal(transaction.Boostagram, &boostagram)
	assert.NoError(t, err)

	assert.Equal(t, "⚡ WebLN Demo", boostagram.AppName)
	assert.Equal(t, "⚡ WebLN Demo", boostagram.Name)
	assert.Equal(t, "Podcasting 2.0", boostagram.Podcast)
	assert.Equal(t, "Episode 104: A New Dump", boostagram.Episode.String())
	assert.Equal(t, "https://feeds.podcastindex.org/pc20.xml", boostagram.FeedId.String())
	assert.Equal(t, int64(21), boostagram.Timestamp)
	assert.Equal(t, "Go podcasting!", boostagram.Message)
	assert.Equal(t, "Satoshi Nakamoto", boostagram.SenderName)
	assert.Equal(t, "boost", boostagram.Action)
	assert.Equal(t, int64(1000), boostagram.ValueMsatTotal)

	assert.Equal(t, "Go podcasting!", transaction.Description)

	assert.Equal(t, uint64(1000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_TYPE_OUTGOING, transaction.Type)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.NotNil(t, transaction.Preimage)
	assert.Equal(t, 64, len(*transaction.Preimage))
	assert.Zero(t, transaction.FeeReserveMsat)
}

func TestSendKeysend_IsolatedAppToNoApp(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	// setup for self payment
	svc.LNClient.(*tests.MockLn).Pubkey = "03cbd788f5b22bd56e2714bff756372d2293504c064e03250ed16a4dd80ad70e2c"

	app, _, err := tests.CreateApp(svc)
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
		AmountMsat: 133000, // payment is 123000 msat, but we also calculate fee reserves max of(10 sats or 1%)
	})

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	mockPreimage := "c8aeb44ae8eb269c8dbfb7ec5c263f0bfa3d755bc0ca641b8ee118673afda657"

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendKeysend(123000, "03cbd788f5b22bd56e2714bff756372d2293504c064e03250ed16a4dd80ad70e2c", []lnclient.TLVRecord{}, mockPreimage, svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.NoError(t, err)
	assert.NotNil(t, transaction)
	assert.Equal(t, uint64(123000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, mockPreimage, *transaction.Preimage)
	assert.Equal(t, app.ID, *transaction.AppId)
	assert.Equal(t, dbRequestEvent.ID, *transaction.RequestEventId)
	assert.True(t, transaction.SelfPayment)

	transactionType := constants.TRANSACTION_TYPE_INCOMING
	incomingTransaction, err := transactionsService.LookupTransaction(context.TODO(), transaction.PaymentHash, &transactionType, svc.LNClient, nil)
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

func TestSendKeysend_IsolatedAppToIsolatedApp(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	// setup for self payment
	svc.LNClient.(*tests.MockLn).Pubkey = "03cbd788f5b22bd56e2714bff756372d2293504c064e03250ed16a4dd80ad70e2c"

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)
	app.Isolated = true
	err = svc.DB.Save(&app).Error
	assert.NoError(t, err)

	app2, _, err := tests.CreateApp(svc)
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
		AmountMsat: 133000, // payment is 123000 msat, but we also calculate fee reserves max of(10 sats or 1%)
	})

	dbRequestEvent := &db.RequestEvent{}
	err = svc.DB.Create(&dbRequestEvent).Error
	assert.NoError(t, err)

	mockPreimage := "c8aeb44ae8eb269c8dbfb7ec5c263f0bfa3d755bc0ca641b8ee118673afda657"

	// Keysend from app 1 to app 2
	tlvRecords := []lnclient.TLVRecord{
		{
			Type:  696969,
			Value: hex.EncodeToString([]byte(strconv.FormatUint(uint64(app2.ID), 10))),
		},
		{
			Type:  7629169,
			Value: "7b22616374696f6e223a22626f6f7374222c2276616c75655f6d736174223a313030302c2276616c75655f6d7361745f746f74616c223a313030302c226170705f6e616d65223a22e29aa1205765624c4e2044656d6f222c226170705f76657273696f6e223a22312e30222c22666565644944223a2268747470733a2f2f66656564732e706f6463617374696e6465782e6f72672f706332302e786d6c222c22706f6463617374223a22506f6463617374696e6720322e30222c22657069736f6465223a22457069736f6465203130343a2041204e65772044756d70222c227473223a32312c226e616d65223a22e29aa1205765624c4e2044656d6f222c2273656e6465725f6e616d65223a225361746f736869204e616b616d6f746f222c226d657373616765223a22476f20706f6463617374696e6721227d",
		},
	}

	mockEventConsumer := tests.NewMockEventConsumer()
	svc.EventPublisher.RegisterSubscriber(mockEventConsumer)

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendKeysend(123000, "03cbd788f5b22bd56e2714bff756372d2293504c064e03250ed16a4dd80ad70e2c", tlvRecords, mockPreimage, svc.LNClient, &app.ID, &dbRequestEvent.ID)

	assert.NoError(t, err)
	assert.NotNil(t, transaction)
	assert.Equal(t, uint64(123000), transaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, mockPreimage, *transaction.Preimage)
	assert.Equal(t, app.ID, *transaction.AppId)
	assert.Equal(t, dbRequestEvent.ID, *transaction.RequestEventId)
	assert.True(t, transaction.SelfPayment)

	transactionType := constants.TRANSACTION_TYPE_INCOMING
	incomingTransaction, err := transactionsService.LookupTransaction(context.TODO(), transaction.PaymentHash, &transactionType, svc.LNClient, &app2.ID)
	assert.NoError(t, err)
	assert.Equal(t, uint64(123000), incomingTransaction.AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, incomingTransaction.State)
	assert.Equal(t, mockPreimage, *incomingTransaction.Preimage)
	assert.Equal(t, app2.ID, *incomingTransaction.AppId)
	assert.True(t, incomingTransaction.SelfPayment)

	// receiving app should have the same data as what was sent
	assert.Equal(t, transaction.Description, incomingTransaction.Description)
	assert.Equal(t, transaction.Metadata, incomingTransaction.Metadata)
	assert.Equal(t, transaction.Boostagram, incomingTransaction.Boostagram)

	transactions := []db.Transaction{}
	result := svc.DB.Find(&transactions)
	assert.Equal(t, int64(3), result.RowsAffected)
	// expect balance to be decreased
	assert.Equal(t, int64(10000), queries.GetIsolatedBalance(svc.DB, app.ID))

	// expect app2 to receive the payment
	assert.Equal(t, int64(123000), queries.GetIsolatedBalance(svc.DB, app2.ID))

	// check notifications
	assert.Equal(t, 2, len(mockEventConsumer.GetConsumedEvents()))

	assert.Equal(t, "nwc_payment_sent", mockEventConsumer.GetConsumedEvents()[1].Event)
	settledTransaction := mockEventConsumer.GetConsumedEvents()[1].Properties.(*db.Transaction)
	assert.Equal(t, transaction.ID, settledTransaction.ID)

	assert.Equal(t, "nwc_payment_received", mockEventConsumer.GetConsumedEvents()[0].Event)
	receivedTransaction := mockEventConsumer.GetConsumedEvents()[0].Properties.(*db.Transaction)
	assert.Equal(t, incomingTransaction.ID, receivedTransaction.ID)
}
