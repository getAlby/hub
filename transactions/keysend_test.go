package transactions

import (
	"context"
	"encoding/json"
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
	transaction, err := transactionsService.SendKeysend(ctx, uint64(1000), "fake destination", nil, "", svc.LNClient, nil, nil)
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
}
func TestSendKeysend_CustomPreimage(t *testing.T) {
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	customPreimage := "018465013e2337234a7e5530a21c4a8cf70d84231f4a8ff0b1e2cce3cb2bd03b"
	transactionsService := NewTransactionsService(svc.DB)
	transaction, err := transactionsService.SendKeysend(ctx, uint64(1000), "fake destination", nil, customPreimage, svc.LNClient, nil, nil)
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
	transaction, err := transactionsService.SendKeysend(ctx, uint64(1000), "fake destination", nil, "", svc.LNClient, &app.ID, &dbRequestEvent.ID)

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
	transaction, err := transactionsService.SendKeysend(ctx, uint64(1000), "fake destination", nil, "", svc.LNClient, &app.ID, &dbRequestEvent.ID)
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
	transaction, err := transactionsService.SendKeysend(ctx, uint64(1000), "fake destination", nil, "", svc.LNClient, &app.ID, &dbRequestEvent.ID)

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
	transaction, err := transactionsService.SendKeysend(ctx, uint64(1000), "fake destination", nil, "", svc.LNClient, &app.ID, &dbRequestEvent.ID)
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
	transaction, err := transactionsService.SendKeysend(ctx, uint64(1000), "fake destination", nil, "", svc.LNClient, &app.ID, &dbRequestEvent.ID)

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
	transaction, err := transactionsService.SendKeysend(ctx, uint64(1000), "fake destination", nil, "", svc.LNClient, &app.ID, &dbRequestEvent.ID)
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
	ctx := context.TODO()

	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	transactionsService := NewTransactionsService(svc.DB)
	transaction, err := transactionsService.SendKeysend(ctx, uint64(1000), "fake destination", []lnclient.TLVRecord{
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
	assert.Equal(t, "Episode 104: A New Dump", boostagram.Episode)
	assert.Equal(t, "https://feeds.podcastindex.org/pc20.xml", boostagram.FeedId)
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
