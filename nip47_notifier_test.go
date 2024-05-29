package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"testing"

	"github.com/getAlby/nostr-wallet-connect/db"
	"github.com/getAlby/nostr-wallet-connect/events"
	"github.com/getAlby/nostr-wallet-connect/nip47"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/stretchr/testify/assert"
)

func TestSendNotification(t *testing.T) {
	ctx := context.TODO()
	defer os.Remove(testDB)
	mockLn, err := NewMockLn()
	assert.NoError(t, err)
	svc, err := createTestService(mockLn)
	assert.NoError(t, err)
	app, ss, err := createApp(svc)
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId:         app.ID,
		App:           *app,
		RequestMethod: nip47.NOTIFICATIONS_PERMISSION,
	}
	err = svc.db.Create(appPermission).Error
	assert.NoError(t, err)

	svc.nip47NotificationQueue = nip47.NewNip47NotificationQueue(svc.logger)
	svc.eventPublisher.RegisterSubscriber(svc.nip47NotificationQueue)

	testEvent := &events.Event{
		Event: "nwc_payment_received",
		Properties: &events.PaymentReceivedEventProperties{
			PaymentHash: mockPaymentHash,
			Amount:      uint64(mockTransaction.Amount),
			NodeType:    "LDK",
		},
	}

	svc.eventPublisher.Publish(testEvent)

	receivedEvent := <-svc.nip47NotificationQueue.Channel()
	assert.Equal(t, testEvent, receivedEvent)

	relay := NewMockRelay()

	n := NewNip47Notifier(svc, relay)
	n.ConsumeEvent(ctx, receivedEvent)

	assert.NotNil(t, relay.publishedEvent)
	assert.NotEmpty(t, relay.publishedEvent.Content)

	decrypted, err := nip04.Decrypt(relay.publishedEvent.Content, ss)
	assert.NoError(t, err)
	unmarshalledResponse := nip47.Notification{
		Notification: &nip47.PaymentReceivedNotification{},
	}

	err = json.Unmarshal([]byte(decrypted), &unmarshalledResponse)
	assert.NoError(t, err)
	assert.Equal(t, nip47.PAYMENT_RECEIVED_NOTIFICATION, unmarshalledResponse.NotificationType)

	transaction := (unmarshalledResponse.Notification.(*nip47.PaymentReceivedNotification))
	assert.Equal(t, mockTransaction.Type, transaction.Type)
	assert.Equal(t, mockTransaction.Invoice, transaction.Invoice)
	assert.Equal(t, mockTransaction.Description, transaction.Description)
	assert.Equal(t, mockTransaction.DescriptionHash, transaction.DescriptionHash)
	assert.Equal(t, mockTransaction.Preimage, transaction.Preimage)
	assert.Equal(t, mockTransaction.PaymentHash, transaction.PaymentHash)
	assert.Equal(t, mockTransaction.Amount, transaction.Amount)
	assert.Equal(t, mockTransaction.FeesPaid, transaction.FeesPaid)
	assert.Equal(t, mockTransaction.SettledAt, transaction.SettledAt)
}

func TestSendNotificationNoPermission(t *testing.T) {
	ctx := context.TODO()
	defer os.Remove(testDB)
	mockLn, err := NewMockLn()
	assert.NoError(t, err)
	svc, err := createTestService(mockLn)
	assert.NoError(t, err)
	_, _, err = createApp(svc)
	assert.NoError(t, err)

	svc.nip47NotificationQueue = nip47.NewNip47NotificationQueue(svc.logger)
	svc.eventPublisher.RegisterSubscriber(svc.nip47NotificationQueue)

	testEvent := &events.Event{
		Event: "nwc_payment_received",
		Properties: &events.PaymentReceivedEventProperties{
			PaymentHash: mockPaymentHash,
			Amount:      uint64(mockTransaction.Amount),
			NodeType:    "LDK",
		},
	}

	svc.eventPublisher.Publish(testEvent)

	receivedEvent := <-svc.nip47NotificationQueue.Channel()
	assert.Equal(t, testEvent, receivedEvent)

	relay := NewMockRelay()

	n := NewNip47Notifier(svc, relay)
	n.ConsumeEvent(ctx, receivedEvent)

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
