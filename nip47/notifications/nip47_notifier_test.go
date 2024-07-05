package notifications

import (
	"context"
	"encoding/json"
	"log"
	"testing"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/nip47/permissions"
	"github.com/getAlby/hub/tests"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/stretchr/testify/assert"
)

func TestSendNotification_PaymentReceived(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	app, ss, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: permissions.NOTIFICATIONS_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	nip47NotificationQueue := NewNip47NotificationQueue()
	svc.EventPublisher.RegisterSubscriber(nip47NotificationQueue)

	testEvent := &events.Event{
		Event: "nwc_payment_received",
		Properties: &events.PaymentReceivedEventProperties{
			PaymentHash: tests.MockPaymentHash,
		},
	}

	svc.EventPublisher.Publish(testEvent)

	receivedEvent := <-nip47NotificationQueue.Channel()
	assert.Equal(t, testEvent, receivedEvent)

	relay := NewMockRelay()

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)

	notifier := NewNip47Notifier(relay, svc.DB, svc.Cfg, svc.Keys, permissionsSvc, svc.LNClient)
	notifier.ConsumeEvent(ctx, receivedEvent)

	assert.NotNil(t, relay.publishedEvent)
	assert.NotEmpty(t, relay.publishedEvent.Content)

	decrypted, err := nip04.Decrypt(relay.publishedEvent.Content, ss)
	assert.NoError(t, err)
	unmarshalledResponse := Notification{
		Notification: &PaymentReceivedNotification{},
	}

	err = json.Unmarshal([]byte(decrypted), &unmarshalledResponse)
	assert.NoError(t, err)
	assert.Equal(t, PAYMENT_RECEIVED_NOTIFICATION, unmarshalledResponse.NotificationType)

	transaction := (unmarshalledResponse.Notification.(*PaymentReceivedNotification))
	assert.Equal(t, tests.MockTransaction.Type, transaction.Type)
	assert.Equal(t, tests.MockTransaction.Invoice, transaction.Invoice)
	assert.Equal(t, tests.MockTransaction.Description, transaction.Description)
	assert.Equal(t, tests.MockTransaction.DescriptionHash, transaction.DescriptionHash)
	assert.Equal(t, tests.MockTransaction.Preimage, transaction.Preimage)
	assert.Equal(t, tests.MockTransaction.PaymentHash, transaction.PaymentHash)
	assert.Equal(t, tests.MockTransaction.Amount, transaction.Amount)
	assert.Equal(t, tests.MockTransaction.FeesPaid, transaction.FeesPaid)
	assert.Equal(t, tests.MockTransaction.SettledAt, transaction.SettledAt)

}
func TestSendNotification_PaymentSent(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	app, ss, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: permissions.NOTIFICATIONS_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	nip47NotificationQueue := NewNip47NotificationQueue()
	svc.EventPublisher.RegisterSubscriber(nip47NotificationQueue)

	testEvent := &events.Event{
		Event: "nwc_payment_sent",
		Properties: &events.PaymentSentEventProperties{
			PaymentHash: tests.MockPaymentHash,
		},
	}

	svc.EventPublisher.Publish(testEvent)

	receivedEvent := <-nip47NotificationQueue.Channel()
	assert.Equal(t, testEvent, receivedEvent)

	relay := NewMockRelay()

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)

	notifier := NewNip47Notifier(relay, svc.DB, svc.Cfg, svc.Keys, permissionsSvc, svc.LNClient)
	notifier.ConsumeEvent(ctx, receivedEvent)

	assert.NotNil(t, relay.publishedEvent)
	assert.NotEmpty(t, relay.publishedEvent.Content)

	decrypted, err := nip04.Decrypt(relay.publishedEvent.Content, ss)
	assert.NoError(t, err)
	unmarshalledResponse := Notification{
		Notification: &PaymentReceivedNotification{},
	}

	err = json.Unmarshal([]byte(decrypted), &unmarshalledResponse)
	assert.NoError(t, err)
	assert.Equal(t, PAYMENT_SENT_NOTIFICATION, unmarshalledResponse.NotificationType)

	transaction := (unmarshalledResponse.Notification.(*PaymentReceivedNotification))
	assert.Equal(t, tests.MockTransaction.Type, transaction.Type)
	assert.Equal(t, tests.MockTransaction.Invoice, transaction.Invoice)
	assert.Equal(t, tests.MockTransaction.Description, transaction.Description)
	assert.Equal(t, tests.MockTransaction.DescriptionHash, transaction.DescriptionHash)
	assert.Equal(t, tests.MockTransaction.Preimage, transaction.Preimage)
	assert.Equal(t, tests.MockTransaction.PaymentHash, transaction.PaymentHash)
	assert.Equal(t, tests.MockTransaction.Amount, transaction.Amount)
	assert.Equal(t, tests.MockTransaction.FeesPaid, transaction.FeesPaid)
	assert.Equal(t, tests.MockTransaction.SettledAt, transaction.SettledAt)
}

func TestSendNotificationNoPermission(t *testing.T) {
	ctx := context.TODO()
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)
	_, _, err = tests.CreateApp(svc)
	assert.NoError(t, err)

	nip47NotificationQueue := NewNip47NotificationQueue()
	svc.EventPublisher.RegisterSubscriber(nip47NotificationQueue)

	testEvent := &events.Event{
		Event: "nwc_payment_received",
		Properties: &events.PaymentReceivedEventProperties{
			PaymentHash: tests.MockPaymentHash,
		},
	}

	svc.EventPublisher.Publish(testEvent)

	receivedEvent := <-nip47NotificationQueue.Channel()
	assert.Equal(t, testEvent, receivedEvent)

	relay := NewMockRelay()

	permissionsSvc := permissions.NewPermissionsService(svc.DB, svc.EventPublisher)

	notifier := NewNip47Notifier(relay, svc.DB, svc.Cfg, svc.Keys, permissionsSvc, svc.LNClient)
	notifier.ConsumeEvent(ctx, receivedEvent)

	assert.Nil(t, relay.publishedEvent)
}

type mockRelay struct {
	publishedEvent *nostr.Event
}

func NewMockRelay() *mockRelay {
	return &mockRelay{}
}

func (relay *mockRelay) Publish(ctx context.Context, event nostr.Event) error {
	log.Printf("Mock Publishing event %+v", event)
	relay.publishedEvent = &event
	return nil
}
