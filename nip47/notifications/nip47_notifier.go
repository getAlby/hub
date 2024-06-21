package notifications

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/getAlby/nostr-wallet-connect/config"
	"github.com/getAlby/nostr-wallet-connect/db"
	"github.com/getAlby/nostr-wallet-connect/events"
	"github.com/getAlby/nostr-wallet-connect/lnclient"
	"github.com/getAlby/nostr-wallet-connect/logger"
	"github.com/getAlby/nostr-wallet-connect/nip47/models"
	"github.com/getAlby/nostr-wallet-connect/nip47/permissions"
	"github.com/getAlby/nostr-wallet-connect/service/keys"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Relay interface {
	Publish(ctx context.Context, event nostr.Event) error
}

type Nip47Notifier struct {
	relay          Relay
	cfg            config.Config
	keys           keys.Keys
	lnClient       lnclient.LNClient
	db             *gorm.DB
	permissionsSvc permissions.PermissionsService
}

func NewNip47Notifier(relay Relay, db *gorm.DB, cfg config.Config, keys keys.Keys, permissionsSvc permissions.PermissionsService, lnClient lnclient.LNClient) *Nip47Notifier {
	return &Nip47Notifier{
		relay:          relay,
		cfg:            cfg,
		db:             db,
		lnClient:       lnClient,
		permissionsSvc: permissionsSvc,
		keys:           keys,
	}
}

func (notifier *Nip47Notifier) ConsumeEvent(ctx context.Context, event *events.Event) error {
	switch event.Event {
	case "nwc_payment_received":
		paymentReceivedEventProperties, ok := event.Properties.(*events.PaymentReceivedEventProperties)
		if !ok {
			logger.Logger.WithField("event", event).Error("Failed to cast event")
			return errors.New("failed to cast event")
		}

		transaction, err := notifier.lnClient.LookupInvoice(ctx, paymentReceivedEventProperties.PaymentHash)
		if err != nil {
			logger.Logger.
				WithField("paymentHash", paymentReceivedEventProperties.PaymentHash).
				WithError(err).
				Error("Failed to lookup invoice by payment hash")
			return err
		}
		notification := PaymentReceivedNotification{
			Transaction: *transaction,
		}

		notifier.notifySubscribers(ctx, &Notification{
			Notification:     notification,
			NotificationType: PAYMENT_RECEIVED_NOTIFICATION,
		}, nostr.Tags{})

	case "nwc_payment_sent":
		paymentSentEventProperties, ok := event.Properties.(*events.PaymentSentEventProperties)
		if !ok {
			logger.Logger.WithField("event", event).Error("Failed to cast event")
			return errors.New("failed to cast event")
		}

		transaction, err := notifier.lnClient.LookupInvoice(ctx, paymentSentEventProperties.PaymentHash)
		if err != nil {
			logger.Logger.
				WithField("paymentHash", paymentSentEventProperties.PaymentHash).
				WithError(err).
				Error("Failed to lookup invoice by payment hash")
			return err
		}
		notification := PaymentSentNotification{
			Transaction: *transaction,
		}

		notifier.notifySubscribers(ctx, &Notification{
			Notification:     notification,
			NotificationType: PAYMENT_SENT_NOTIFICATION,
		}, nostr.Tags{})
	}

	return nil
}

func (notifier *Nip47Notifier) notifySubscribers(ctx context.Context, notification *Notification, tags nostr.Tags) {
	apps := []db.App{}

	// TODO: join apps and permissions
	notifier.db.Find(&apps)

	for _, app := range apps {
		hasPermission, _, _ := notifier.permissionsSvc.HasPermission(&app, permissions.NOTIFICATIONS_PERMISSION, 0)
		if !hasPermission {
			continue
		}
		notifier.notifySubscriber(ctx, &app, notification, tags)
	}
}

func (notifier *Nip47Notifier) notifySubscriber(ctx context.Context, app *db.App, notification *Notification, tags nostr.Tags) {
	logger.Logger.WithFields(logrus.Fields{
		"notification": notification,
		"appId":        app.ID,
	}).Info("Notifying subscriber")

	ss, err := nip04.ComputeSharedSecret(app.NostrPubkey, notifier.keys.GetNostrSecretKey())
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"notification": notification,
			"appId":        app.ID,
		}).WithError(err).Error("Failed to compute shared secret")
		return
	}

	payloadBytes, err := json.Marshal(notification)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"notification": notification,
			"appId":        app.ID,
		}).WithError(err).Error("Failed to stringify notification")
		return
	}
	msg, err := nip04.Encrypt(string(payloadBytes), ss)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"notification": notification,
			"appId":        app.ID,
		}).WithError(err).Error("Failed to encrypt notification payload")
		return
	}

	allTags := nostr.Tags{[]string{"p", app.NostrPubkey}}
	allTags = append(allTags, tags...)

	event := &nostr.Event{
		PubKey:    notifier.keys.GetNostrPublicKey(),
		CreatedAt: nostr.Now(),
		Kind:      models.NOTIFICATION_KIND,
		Tags:      allTags,
		Content:   msg,
	}
	err = event.Sign(notifier.keys.GetNostrSecretKey())
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"notification": notification,
			"appId":        app.ID,
		}).WithError(err).Error("Failed to sign event")
		return
	}

	err = notifier.relay.Publish(ctx, *event)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"notification": notification,
			"appId":        app.ID,
		}).WithError(err).Error("Failed to publish notification")
		return
	}
	logger.Logger.WithFields(logrus.Fields{
		"notification": notification,
		"appId":        app.ID,
	}).Info("Published notification event")

}
