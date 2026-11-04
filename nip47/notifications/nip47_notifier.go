package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/getAlby/go-nostr"
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/cipher"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/nip47/permissions"
	nostrmodels "github.com/getAlby/hub/nostr/models"
	"github.com/getAlby/hub/service/keys"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Nip47Notifier struct {
	pool           nostrmodels.SimplePool
	cfg            config.Config
	keys           keys.Keys
	db             *gorm.DB
	permissionsSvc permissions.PermissionsService
}

func NewNip47Notifier(pool nostrmodels.SimplePool, db *gorm.DB, cfg config.Config, keys keys.Keys, permissionsSvc permissions.PermissionsService) *Nip47Notifier {
	return &Nip47Notifier{
		pool:           pool,
		cfg:            cfg,
		db:             db,
		permissionsSvc: permissionsSvc,
		keys:           keys,
	}
}

// ConsumeEvent notifies the subscribed apps. Failures are logged, not returned,
// as they are not retried (see https://github.com/getAlby/hub/issues/2678)
func (notifier *Nip47Notifier) ConsumeEvent(ctx context.Context, event *events.Event) {
	switch event.Event {
	case "nwc_payment_received":
		transaction, ok := event.Properties.(*db.Transaction)
		if !ok {
			logger.Logger.WithFields(EventLogFields(event)).Error("Failed to cast event")
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
		transaction, ok := event.Properties.(*db.Transaction)
		if !ok {
			logger.Logger.WithFields(EventLogFields(event)).Error("Failed to cast event")
			return
		}

		notification := PaymentSentNotification{
			Transaction: *models.ToNip47Transaction(transaction),
		}

		notifier.notifySubscribers(ctx, &Notification{
			Notification:     notification,
			NotificationType: PAYMENT_SENT_NOTIFICATION,
		}, nostr.Tags{}, transaction.AppId)

	case "nwc_hold_invoice_accepted":
		dbTransaction, ok := event.Properties.(*db.Transaction)
		if !ok {
			logger.Logger.WithFields(EventLogFields(event)).Error("Failed to cast event properties to db.Transaction for hold invoice accepted")
			return
		}

		nip47Transaction := models.ToNip47Transaction(dbTransaction)

		notification := HoldInvoiceAcceptedNotification{
			Transaction: *nip47Transaction,
		}

		notifier.notifySubscribers(ctx, &Notification{
			Notification:     notification,
			NotificationType: HOLD_INVOICE_ACCEPTED_NOTIFICATION,
		}, nostr.Tags{}, dbTransaction.AppId)
	}
}

func (notifier *Nip47Notifier) notifySubscribers(ctx context.Context, notification *Notification, tags nostr.Tags, appId *uint) {
	apps := []db.App{}

	// TODO: join apps and permissions
	err := notifier.db.Find(&apps).Error
	if err != nil {
		logger.Logger.WithFields(notification.logFields()).WithError(err).Error("Failed to list apps")
		return
	}

	for _, app := range apps {
		if app.Isolated && (appId == nil || app.ID != *appId) {
			continue
		}

		hasPermission, _, _ := notifier.permissionsSvc.HasPermission(&app, constants.NOTIFICATIONS_SCOPE)
		if !hasPermission {
			continue
		}

		appWalletPrivKey := notifier.keys.GetNostrSecretKey()
		if app.WalletPubkey != nil {
			appWalletPrivKey, err = notifier.keys.GetAppWalletKey(app.ID)
			if err != nil {
				logger.Logger.WithFields(notification.logFields()).WithFields(logrus.Fields{
					"appId": app.ID,
				}).WithError(err).Error("error deriving child key")
				continue
			}
		}

		appWalletPubKey, err := nostr.GetPublicKey(appWalletPrivKey)
		if err != nil {
			logger.Logger.WithFields(notification.logFields()).WithFields(logrus.Fields{
				"appId": app.ID,
			}).WithError(err).Error("Failed to calculate app wallet pub key")
			continue
		}

		for _, encryption := range []string{constants.ENCRYPTION_TYPE_NIP04, constants.ENCRYPTION_TYPE_NIP44_V2} {
			err = notifier.notifySubscriber(ctx, &app, notification, tags, appWalletPubKey, appWalletPrivKey, encryption)
			if err != nil {
				logger.Logger.WithFields(notification.logFields()).WithFields(logrus.Fields{
					"appId":      app.ID,
					"encryption": encryption,
				}).WithError(err).Error("Failed to notify subscriber")
			}
		}
	}
}

func (notifier *Nip47Notifier) notifySubscriber(ctx context.Context, app *db.App, notification *Notification, tags nostr.Tags, appWalletPubKey, appWalletPrivKey string, encryption string) error {
	logger.Logger.WithFields(notification.logFields()).WithFields(logrus.Fields{
		"appId":      app.ID,
		"encryption": encryption,
	}).Debug("Notifying subscriber")

	var err error

	payloadBytes, err := json.Marshal(notification)
	if err != nil {
		logger.Logger.WithFields(notification.logFields()).WithFields(logrus.Fields{
			"appId":      app.ID,
			"encryption": encryption,
		}).WithError(err).Error("Failed to stringify notification")
		return err
	}

	nip47Cipher, err := cipher.NewNip47Cipher(encryption, app.AppPubkey, appWalletPrivKey)
	if err != nil {
		logger.Logger.WithFields(notification.logFields()).WithFields(logrus.Fields{
			"appId":      app.ID,
			"encryption": encryption,
		}).WithError(err).Error("Failed to initialize cipher")
		return err
	}

	msg, err := nip47Cipher.Encrypt(string(payloadBytes))
	if err != nil {
		logger.Logger.WithFields(notification.logFields()).WithFields(logrus.Fields{
			"appId":      app.ID,
			"encryption": encryption,
		}).WithError(err).Error("Failed to encrypt notification payload")
		return err
	}

	allTags := nostr.Tags{[]string{"p", app.AppPubkey}}
	allTags = append(allTags, tags...)

	event := &nostr.Event{
		PubKey:    appWalletPubKey,
		CreatedAt: nostr.Now(),
		Kind:      models.NOTIFICATION_KIND,
		Tags:      allTags,
		Content:   msg,
	}

	if encryption == constants.ENCRYPTION_TYPE_NIP04 {
		event.Kind = models.LEGACY_NOTIFICATION_KIND
	}

	err = event.Sign(appWalletPrivKey)
	if err != nil {
		logger.Logger.WithFields(notification.logFields()).WithFields(logrus.Fields{
			"appId":      app.ID,
			"encryption": encryption,
		}).WithError(err).Error("Failed to sign event")
		return err
	}

	publishResultChannel := notifier.pool.PublishMany(ctx, notifier.cfg.GetRelayUrls(), *event)

	publishSuccessful := false
	var relayErrors []error
	for result := range publishResultChannel {
		if result.Error == nil {
			publishSuccessful = true
		} else {
			relayErrors = append(relayErrors, fmt.Errorf("%s: %w", result.RelayURL, result.Error))
			logger.Logger.WithFields(notification.logFields()).WithFields(logrus.Fields{
				"appId": app.ID,
				"relay": result.RelayURL,
			}).WithError(result.Error).Error("failed to publish notification to relay")
		}
	}

	if !publishSuccessful {
		err = fmt.Errorf("failed to publish notification to any relay: %w", errors.Join(relayErrors...))
		logger.Logger.WithFields(notification.logFields()).WithFields(logrus.Fields{
			"appId":      app.ID,
			"encryption": encryption,
		}).WithError(err).Error("Failed to publish notification")
		return err
	}
	logger.Logger.WithFields(logrus.Fields{
		"appId":      app.ID,
		"encryption": encryption,
	}).Debug("Published notification event")
	return nil
}
