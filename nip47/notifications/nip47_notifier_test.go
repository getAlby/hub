package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/getAlby/go-nostr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/nip47/permissions"
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

type rejectingPool struct {
	rejectKinds []int
}

func (pool *rejectingPool) PublishMany(ctx context.Context, relayUrls []string, event nostr.Event) chan nostr.PublishResult {
	channel := make(chan nostr.PublishResult, 1)
	if slices.Contains(pool.rejectKinds, event.Kind) {
		channel <- nostr.PublishResult{RelayURL: "wss://fakerelay.com", Error: errors.New("relay rejected event")}
	} else {
		channel <- nostr.PublishResult{RelayURL: "wss://fakerelay.com"}
	}
	close(channel)
	return channel
}

func (pool *rejectingPool) QuerySingle(ctx context.Context, urls []string, filter nostr.Filter, opts ...nostr.SubscriptionOption) *nostr.RelayEvent {
	return nil
}

func TestSendNotification_RetryOnlyWhenNothingPublished(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	scopes := []string{constants.NOTIFICATIONS_SCOPE}
	// valid hex but not a valid secp256k1 point, so the cipher fails for this app
	_, _, err = svc.AppsService.CreateApp("invalid pubkey", strings.Repeat("ff", 32), 0, constants.BUDGET_RENEWAL_NEVER, nil, scopes, false, nil)
	require.NoError(t, err)
	_, _, err = svc.AppsService.CreateApp("valid pubkey", "", 0, constants.BUDGET_RENEWAL_NEVER, nil, scopes, false, nil)
	require.NoError(t, err)

	pool := &rejectingPool{}
	notifier := NewNip47Notifier(pool, svc.DB, svc.Cfg, svc.Keys, permissions.NewPermissionsService(svc.DB, svc.EventPublisher))
	event := &events.Event{Event: "nwc_payment_received", Properties: &db.Transaction{}}

	// relays offline: retry
	pool.rejectKinds = []int{models.LEGACY_NOTIFICATION_KIND, models.NOTIFICATION_KIND}
	require.ErrorIs(t, notifier.ConsumeEvent(context.TODO(), event), errPublishFailed)

	// partially published: no retry, it would send duplicates
	pool.rejectKinds = []int{models.LEGACY_NOTIFICATION_KIND}
	require.NoError(t, notifier.ConsumeEvent(context.TODO(), event))
}
