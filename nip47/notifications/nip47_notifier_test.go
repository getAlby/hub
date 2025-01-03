package notifications

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/nbd-wtf/go-nostr"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/nip47/cipher"
	"github.com/getAlby/hub/nip47/permissions"
	"github.com/getAlby/hub/tests"
	"github.com/getAlby/hub/transactions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockConsumer struct {
	nip47NotificationQueue Nip47NotificationQueue
}

func NewMockConsumer(nip47NotificationQueue Nip47NotificationQueue) *mockConsumer {
	return &mockConsumer{
		nip47NotificationQueue: nip47NotificationQueue,
	}
}

func (svc *mockConsumer) ConsumeEvent(ctx context.Context, event *events.Event, globalProperties map[string]interface{}) {
	svc.nip47NotificationQueue.AddToQueue(event)
}

func doTestSendNotificationPaymentReceived(t *testing.T, svc *tests.TestService, app *db.App, cipher *cipher.Nip47Cipher) {
	ctx := context.TODO()

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: constants.NOTIFICATIONS_SCOPE,
	}
	err := svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	settledAt := time.Unix(*tests.MockLNClientTransaction.SettledAt, 0)
	initialTransaction := db.Transaction{
		Type:            constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest:  tests.MockLNClientTransaction.Invoice,
		Description:     tests.MockLNClientTransaction.Description,
		DescriptionHash: tests.MockLNClientTransaction.DescriptionHash,
		Preimage:        &tests.MockLNClientTransaction.Preimage,
		PaymentHash:     tests.MockLNClientTransaction.PaymentHash,
		AmountMsat:      uint64(tests.MockLNClientTransaction.Amount),
		FeeMsat:         uint64(tests.MockLNClientTransaction.FeesPaid),
		SettledAt:       &settledAt,
		AppId:           &app.ID,
		State:           constants.TRANSACTION_STATE_SETTLED,
	}
	err = svc.DB.Create(&initialTransaction).Error
	assert.NoError(t, err)

	nip47NotificationQueue := NewNip47NotificationQueue()
	svc.EventPublisher.RegisterSubscriber(NewMockConsumer(nip47NotificationQueue))

	testEvent := &events.Event{
		Event:      "nwc_payment_received",
		Properties: &initialTransaction,
	}

	svc.EventPublisher.Publish(testEvent)

	receivedEvent := <-nip47NotificationQueue.Channel()
	assert.Equal(t, testEvent, receivedEvent)

	relay := tests.NewMockRelay()

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB, svc.EventPublisher)

	notifier := NewNip47Notifier(relay, svc.DB, svc.Cfg, svc.Keys, permissionsSvc, transactionsSvc, svc.LNClient)
	notifier.ConsumeEvent(ctx, receivedEvent)

	assert.NotNil(t, relay.PublishedEvents[1])
	assert.NotEmpty(t, relay.PublishedEvents[1].Content)

	decrypted, err := cipher.Decrypt(relay.PublishedEvents[1].Content)
	assert.NoError(t, err)
	unmarshalledResponse := Notification{
		Notification: &PaymentReceivedNotification{},
	}

	err = json.Unmarshal([]byte(decrypted), &unmarshalledResponse)
	assert.NoError(t, err)
	assert.Equal(t, PAYMENT_RECEIVED_NOTIFICATION, unmarshalledResponse.NotificationType)

	transaction := (unmarshalledResponse.Notification.(*PaymentReceivedNotification))
	assert.Equal(t, constants.TRANSACTION_TYPE_INCOMING, transaction.Type)
	assert.Equal(t, tests.MockLNClientTransaction.Invoice, transaction.Invoice)
	assert.Equal(t, tests.MockLNClientTransaction.Description, transaction.Description)
	assert.Equal(t, tests.MockLNClientTransaction.DescriptionHash, transaction.DescriptionHash)
	assert.Equal(t, tests.MockLNClientTransaction.Preimage, transaction.Preimage)
	assert.Equal(t, tests.MockLNClientTransaction.PaymentHash, transaction.PaymentHash)
	assert.Equal(t, tests.MockLNClientTransaction.Amount, transaction.Amount)
	assert.Equal(t, tests.MockLNClientTransaction.FeesPaid, transaction.FeesPaid)
	assert.Equal(t, tests.MockLNClientTransaction.SettledAt, transaction.SettledAt)
}

func TestSendNotification_PaymentReceived(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	app, cipher, err := tests.CreateApp(svc)
	assert.NoError(t, err)
	doTestSendNotificationPaymentReceived(t, svc, app, cipher)
}

func TestSendNotification_Legacy_PaymentReceived(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	app, cipher, err := tests.CreateLegacyApp(svc, nostr.GeneratePrivateKey())
	assert.NoError(t, err)
	doTestSendNotificationPaymentReceived(t, svc, app, cipher)
}

func doTestSendNotificationPaymentSent(t *testing.T, svc *tests.TestService, app *db.App, cipher *cipher.Nip47Cipher) {
	ctx := context.TODO()

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: constants.NOTIFICATIONS_SCOPE,
	}
	err := svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	settledAt := time.Unix(*tests.MockLNClientTransaction.SettledAt, 0)
	initialTransaction := db.Transaction{
		Type:            constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest:  tests.MockLNClientTransaction.Invoice,
		Description:     tests.MockLNClientTransaction.Description,
		DescriptionHash: tests.MockLNClientTransaction.DescriptionHash,
		Preimage:        &tests.MockLNClientTransaction.Preimage,
		PaymentHash:     tests.MockLNClientTransaction.PaymentHash,
		AmountMsat:      uint64(tests.MockLNClientTransaction.Amount),
		FeeMsat:         uint64(tests.MockLNClientTransaction.FeesPaid),
		SettledAt:       &settledAt,
		AppId:           &app.ID,
	}
	err = svc.DB.Create(&initialTransaction).Error
	assert.NoError(t, err)

	nip47NotificationQueue := NewNip47NotificationQueue()
	svc.EventPublisher.RegisterSubscriber(NewMockConsumer(nip47NotificationQueue))

	testEvent := &events.Event{
		Event:      "nwc_payment_sent",
		Properties: &initialTransaction,
	}

	svc.EventPublisher.Publish(testEvent)

	receivedEvent := <-nip47NotificationQueue.Channel()
	assert.Equal(t, testEvent, receivedEvent)

	relay := tests.NewMockRelay()

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB, svc.EventPublisher)

	notifier := NewNip47Notifier(relay, svc.DB, svc.Cfg, svc.Keys, permissionsSvc, transactionsSvc, svc.LNClient)
	notifier.ConsumeEvent(ctx, receivedEvent)

	assert.NotNil(t, relay.PublishedEvents[1])
	assert.NotEmpty(t, relay.PublishedEvents[1].Content)

	decrypted, err := cipher.Decrypt(relay.PublishedEvents[1].Content)
	assert.NoError(t, err)
	unmarshalledResponse := Notification{
		Notification: &PaymentReceivedNotification{},
	}

	err = json.Unmarshal([]byte(decrypted), &unmarshalledResponse)
	assert.NoError(t, err)
	assert.Equal(t, PAYMENT_SENT_NOTIFICATION, unmarshalledResponse.NotificationType)

	transaction := (unmarshalledResponse.Notification.(*PaymentReceivedNotification))
	assert.Equal(t, constants.TRANSACTION_TYPE_OUTGOING, transaction.Type)
	assert.Equal(t, tests.MockLNClientTransaction.Invoice, transaction.Invoice)
	assert.Equal(t, tests.MockLNClientTransaction.Description, transaction.Description)
	assert.Equal(t, tests.MockLNClientTransaction.DescriptionHash, transaction.DescriptionHash)
	assert.Equal(t, tests.MockLNClientTransaction.Preimage, transaction.Preimage)
	assert.Equal(t, tests.MockLNClientTransaction.PaymentHash, transaction.PaymentHash)
	assert.Equal(t, tests.MockLNClientTransaction.Amount, transaction.Amount)
	assert.Equal(t, tests.MockLNClientTransaction.FeesPaid, transaction.FeesPaid)
	assert.Equal(t, tests.MockLNClientTransaction.SettledAt, transaction.SettledAt)
}

func TestSendNotification_PaymentSent(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	app, ss, err := tests.CreateApp(svc)
	assert.NoError(t, err)
	doTestSendNotificationPaymentSent(t, svc, app, ss)
}

func TestSendNotification_Legacy_PaymentSent(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	app, cipher, err := tests.CreateLegacyApp(svc, nostr.GeneratePrivateKey())
	assert.NoError(t, err)
	doTestSendNotificationPaymentSent(t, svc, app, cipher)
}

func doTestSendNotificationNoPermission(t *testing.T, svc *tests.TestService) {
	ctx := context.TODO()

	svc.DB.Create(&db.Transaction{
		PaymentHash: tests.MockPaymentHash,
	})

	nip47NotificationQueue := NewNip47NotificationQueue()
	svc.EventPublisher.RegisterSubscriber(NewMockConsumer(nip47NotificationQueue))

	testEvent := &events.Event{
		Event: "nwc_payment_received",
		Properties: &lnclient.Transaction{
			PaymentHash: tests.MockPaymentHash,
		},
	}

	svc.EventPublisher.Publish(testEvent)

	receivedEvent := <-nip47NotificationQueue.Channel()
	assert.Equal(t, testEvent, receivedEvent)

	relay := tests.NewMockRelay()

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	transactionsSvc := transactions.NewTransactionsService(svc.DB, svc.EventPublisher)

	notifier := NewNip47Notifier(relay, svc.DB, svc.Cfg, svc.Keys, permissionsSvc, transactionsSvc, svc.LNClient)
	notifier.ConsumeEvent(ctx, receivedEvent)

	assert.Nil(t, relay.PublishedEvents)
}

func TestSendNotification_NoPermission(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)
	_, _, err = tests.CreateApp(svc)
	assert.NoError(t, err)
	doTestSendNotificationNoPermission(t, svc)
}
func TestSendNotification_Legacy_NoPermission(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)
	_, _, err = tests.CreateLegacyApp(svc, nostr.GeneratePrivateKey())
	assert.NoError(t, err)
	doTestSendNotificationNoPermission(t, svc)
}
