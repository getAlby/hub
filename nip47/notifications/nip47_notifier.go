package notifications

import (
	"context"
	"encoding/json"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/nip47/permissions"
	nostrmodels "github.com/getAlby/hub/nostr/models"
	"github.com/getAlby/hub/service/keys"
	"github.com/getAlby/hub/transactions"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Nip47Notifier struct {
	relay               nostrmodels.Relay
	cfg                 config.Config
	keys                keys.Keys
	lnClient            lnclient.LNClient
	db                  *gorm.DB
	permissionsSvc      permissions.PermissionsService
	transactionsService transactions.TransactionsService
}

func NewNip47Notifier(relay nostrmodels.Relay, db *gorm.DB, cfg config.Config, keys keys.Keys, permissionsSvc permissions.PermissionsService, transactionsService transactions.TransactionsService, lnClient lnclient.LNClient) *Nip47Notifier {
	return &Nip47Notifier{
		relay:               relay,
		cfg:                 cfg,
		db:                  db,
		lnClient:            lnClient,
		permissionsSvc:      permissionsSvc,
		transactionsService: transactionsService,
		keys:                keys,
	}
}

func (notifier *Nip47Notifier) ConsumeEvent(ctx context.Context, event *events.Event) {

	// TODO: should listen to transaction service events instead
	// then self-payments will also trigger NIP-47 notifications
	switch event.Event {
	case "nwc_payment_received":
		lnClientTransaction, ok := event.Properties.(*lnclient.Transaction)
		if !ok {
			logger.Logger.WithField("event", event).Error("Failed to cast event")
			return
		}

		transactionType := constants.TRANSACTION_TYPE_INCOMING
		transaction, err := notifier.transactionsService.LookupTransaction(ctx, lnClientTransaction.PaymentHash, &transactionType, notifier.lnClient, nil)
		if err != nil {
			logger.Logger.
				WithField("paymentHash", lnClientTransaction.PaymentHash).
				WithError(err).
				Error("Failed to lookup transaction by payment hash")
			return
		}
		notification := PaymentReceivedNotification{
			Transaction: *models.ToNip47Transaction(transaction),
		}

		notifier.notifySubscribers(ctx, &Notification{
			Notification:     notification,
			NotificationType: PAYMENT_RECEIVED_NOTIFICATION,
		}, nostr.Tags{}, transaction.AppId)

	case "nwc_payment_sent":
		paymentSentEventProperties, ok := event.Properties.(*lnclient.Transaction)
		if !ok {
			logger.Logger.WithField("event", event).Error("Failed to cast event")
			return
		}

		transactionType := constants.TRANSACTION_TYPE_OUTGOING
		transaction, err := notifier.transactionsService.LookupTransaction(ctx, paymentSentEventProperties.PaymentHash, &transactionType, notifier.lnClient, nil)
		if err != nil {
			logger.Logger.
				WithField("paymentHash", paymentSentEventProperties.PaymentHash).
				WithError(err).
				Error("Failed to lookup invoice by payment hash")
			return
		}
		notification := PaymentSentNotification{
			Transaction: *models.ToNip47Transaction(transaction),
		}

		notifier.notifySubscribers(ctx, &Notification{
			Notification:     notification,
			NotificationType: PAYMENT_SENT_NOTIFICATION,
		}, nostr.Tags{}, transaction.AppId)
	}
}

func (notifier *Nip47Notifier) notifySubscribers(ctx context.Context, notification *Notification, tags nostr.Tags, appId *uint) {
	apps := []db.App{}

	// TODO: join apps and permissions
	notifier.db.Find(&apps)

	for _, app := range apps {
		if app.Isolated && (appId == nil || app.ID != *appId) {
			continue
		}

		hasPermission, _, _ := notifier.permissionsSvc.HasPermission(&app, constants.NOTIFICATIONS_SCOPE)
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
