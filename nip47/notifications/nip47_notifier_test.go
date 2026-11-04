package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/getAlby/go-nostr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/nip47/permissions"
	nostrmodels "github.com/getAlby/hub/nostr/models"
	"github.com/getAlby/hub/tests"
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

func doTestSendNotificationPaymentReceived(t *testing.T, svc *tests.TestService, createAppFn tests.CreateAppFn, nip47Encryption string) {
	ctx := context.TODO()

	app, cipher, err := createAppFn(svc, nostr.GeneratePrivateKey(), nip47Encryption)
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: constants.NOTIFICATIONS_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	settledAt := time.Unix(*tests.MockLNClientTransaction.SettledAt, 0)
	initialTransaction := db.Transaction{
		Type:            constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest:  tests.MockLNClientTransaction.Invoice,
		Description:     tests.MockLNClientTransaction.Description,
		DescriptionHash: tests.MockLNClientTransaction.DescriptionHash,
		Preimage:        &tests.MockLNClientTransaction.Preimage,
		PaymentHash:     tests.MockLNClientTransaction.PaymentHash,
		AmountMsat:      uint64(tests.MockLNClientTransaction.AmountMsat),
		FeeMsat:         uint64(tests.MockLNClientTransaction.FeesPaidMsat),
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

	pool := tests.NewMockSimplePool()

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)

	notifier := NewNip47Notifier(pool, svc.DB, svc.Cfg, svc.Keys, permissionsSvc)
	notifier.ConsumeEvent(ctx, receivedEvent)

	var publishedEvent *nostr.Event
	if nip47Encryption == constants.ENCRYPTION_TYPE_NIP04 {
		publishedEvent = pool.PublishedEvents[0]
	} else {
		publishedEvent = pool.PublishedEvents[1]
	}

	assert.NotNil(t, publishedEvent)
	assert.NotEmpty(t, publishedEvent.Content)

	decrypted, err := cipher.Decrypt(publishedEvent.Content)
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
	assert.Equal(t, tests.MockLNClientTransaction.AmountMsat, transaction.Amount)
	assert.Equal(t, tests.MockLNClientTransaction.FeesPaidMsat, transaction.FeesPaid)
	assert.Equal(t, tests.MockLNClientTransaction.SettledAt, transaction.SettledAt)
}

func TestSendNotification_Nip04_PaymentReceived(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestSendNotificationPaymentReceived(t, svc, tests.CreateAppWithPrivateKey, constants.ENCRYPTION_TYPE_NIP04)
}

func TestSendNotification_Nip44_PaymentReceived(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestSendNotificationPaymentReceived(t, svc, tests.CreateAppWithPrivateKey, constants.ENCRYPTION_TYPE_NIP44_V2)
}

func TestSendNotification_SharedWalletPubkey_Nip04_PaymentReceived(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	defer svc.Remove()
	require.NoError(t, err)

	doTestSendNotificationPaymentReceived(t, svc, tests.CreateAppWithSharedWalletPubkey, constants.ENCRYPTION_TYPE_NIP04)
}

func TestSendNotification_SharedWalletPubkey_Nip44_PaymentReceived(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	defer svc.Remove()
	require.NoError(t, err)

	doTestSendNotificationPaymentReceived(t, svc, tests.CreateAppWithSharedWalletPubkey, constants.ENCRYPTION_TYPE_NIP44_V2)
}

func doTestSendNotificationPaymentSent(t *testing.T, svc *tests.TestService, createAppFn tests.CreateAppFn, nip47Encryption string) {
	ctx := context.TODO()

	app, cipher, err := createAppFn(svc, nostr.GeneratePrivateKey(), nip47Encryption)
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: constants.NOTIFICATIONS_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	settledAt := time.Unix(*tests.MockLNClientTransaction.SettledAt, 0)
	initialTransaction := db.Transaction{
		Type:            constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest:  tests.MockLNClientTransaction.Invoice,
		Description:     tests.MockLNClientTransaction.Description,
		DescriptionHash: tests.MockLNClientTransaction.DescriptionHash,
		Preimage:        &tests.MockLNClientTransaction.Preimage,
		PaymentHash:     tests.MockLNClientTransaction.PaymentHash,
		AmountMsat:      uint64(tests.MockLNClientTransaction.AmountMsat),
		FeeMsat:         uint64(tests.MockLNClientTransaction.FeesPaidMsat),
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

	pool := tests.NewMockSimplePool()

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)

	notifier := NewNip47Notifier(pool, svc.DB, svc.Cfg, svc.Keys, permissionsSvc)
	notifier.ConsumeEvent(ctx, receivedEvent)

	var publishedEvent *nostr.Event
	if nip47Encryption == constants.ENCRYPTION_TYPE_NIP04 {
		publishedEvent = pool.PublishedEvents[0]
	} else {
		publishedEvent = pool.PublishedEvents[1]
	}

	assert.NotNil(t, publishedEvent)
	assert.NotEmpty(t, publishedEvent.Content)

	decrypted, err := cipher.Decrypt(publishedEvent.Content)
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
	assert.Equal(t, tests.MockLNClientTransaction.AmountMsat, transaction.Amount)
	assert.Equal(t, tests.MockLNClientTransaction.FeesPaidMsat, transaction.FeesPaid)
	assert.Equal(t, tests.MockLNClientTransaction.SettledAt, transaction.SettledAt)
}

func TestSendNotification_Nip04_PaymentSent(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestSendNotificationPaymentSent(t, svc, tests.CreateAppWithPrivateKey, constants.ENCRYPTION_TYPE_NIP04)
}

func TestSendNotification_Nip44_PaymentSent(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestSendNotificationPaymentSent(t, svc, tests.CreateAppWithPrivateKey, constants.ENCRYPTION_TYPE_NIP44_V2)
}

func TestSendNotification_SharedWalletPubkey_Nip04_PaymentSent(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestSendNotificationPaymentSent(t, svc, tests.CreateAppWithSharedWalletPubkey, constants.ENCRYPTION_TYPE_NIP04)
}

func TestSendNotification_SharedWalletPubkey_Nip44_PaymentSent(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	doTestSendNotificationPaymentSent(t, svc, tests.CreateAppWithSharedWalletPubkey, constants.ENCRYPTION_TYPE_NIP44_V2)
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

	pool := tests.NewMockSimplePool()

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)

	notifier := NewNip47Notifier(pool, svc.DB, svc.Cfg, svc.Keys, permissionsSvc)
	notifier.ConsumeEvent(ctx, receivedEvent)

	assert.Nil(t, pool.PublishedEvents)
}

func TestSendNotification_NoPermission(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()
	_, _, err = tests.CreateAppWithPrivateKey(svc, nostr.GeneratePrivateKey(), constants.ENCRYPTION_TYPE_NIP44_V2)
	assert.NoError(t, err)
	doTestSendNotificationNoPermission(t, svc)
}

func TestSendNotification_SharedWalletPubkey_NoPermission(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()
	_, _, err = tests.CreateAppWithSharedWalletPubkey(svc, nostr.GeneratePrivateKey(), constants.ENCRYPTION_TYPE_NIP44_V2)
	assert.NoError(t, err)
	doTestSendNotificationNoPermission(t, svc)
}

// rejects every event addressed to the given app pubkey
type rejectingPool struct {
	nostrmodels.SimplePool // for QuerySingle
	rejectAppPubkey        string
	PublishedEvents        []*nostr.Event
}

func (pool *rejectingPool) PublishMany(ctx context.Context, relayUrls []string, event nostr.Event) chan nostr.PublishResult {
	pool.PublishedEvents = append(pool.PublishedEvents, &event)
	result := nostr.PublishResult{RelayURL: "wss://fakerelay.com"}
	if event.Tags.GetFirst([]string{"p", pool.rejectAppPubkey}) != nil {
		result.Error = errors.New("rate limited")
	}
	channel := make(chan nostr.PublishResult, 1)
	channel <- result
	close(channel)
	return channel
}

func TestSendNotification_ContinuesAfterPublishFailure(t *testing.T) {
	ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	failingApp, _, err := tests.CreateAppWithPrivateKey(svc, nostr.GeneratePrivateKey(), constants.ENCRYPTION_TYPE_NIP44_V2)
	require.NoError(t, err)
	otherApp, _, err := tests.CreateAppWithPrivateKey(svc, nostr.GeneratePrivateKey(), constants.ENCRYPTION_TYPE_NIP44_V2)
	require.NoError(t, err)
	for _, app := range []*db.App{failingApp, otherApp} {
		require.NoError(t, svc.DB.Create(&db.AppPermission{AppId: app.ID, App: *app, Scope: constants.NOTIFICATIONS_SCOPE}).Error)
	}

	pool := &rejectingPool{SimplePool: tests.NewMockSimplePool(), rejectAppPubkey: failingApp.AppPubkey}
	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)
	notifier := NewNip47Notifier(pool, svc.DB, svc.Cfg, svc.Keys, permissionsSvc)

	notifier.ConsumeEvent(ctx, &events.Event{
		Event:      "nwc_payment_received",
		Properties: &db.Transaction{PaymentHash: tests.MockLNClientTransaction.PaymentHash},
	})

	// both encryptions were attempted for both apps, the other app is still notified
	require.Len(t, pool.PublishedEvents, 4)
	for i, appPubkey := range []string{failingApp.AppPubkey, failingApp.AppPubkey, otherApp.AppPubkey, otherApp.AppPubkey} {
		assert.Equal(t, appPubkey, pool.PublishedEvents[i].Tags.GetFirst([]string{"p"}).Value())
	}
}
