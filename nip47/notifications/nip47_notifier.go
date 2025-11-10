package notifications

import (
	"context"
	"encoding/json"
	"errors"

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
	"github.com/nbd-wtf/go-nostr"
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

func (notifier *Nip47Notifier) ConsumeEvent(ctx context.Context, event *events.Event) error {
	switch event.Event {
	case "nwc_payment_received":
		transaction, ok := event.Properties.(*db.Transaction)
		if !ok {
			logger.Logger.WithField("event", event).Error("Failed to cast event")
			return errors.New("failed to cast event")
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
			logger.Logger.WithField("event", event).Error("Failed to cast event")
			return errors.New("failed to cast event")
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
			logger.Logger.WithField("event", event).Error("Failed to cast event properties to db.Transaction for hold invoice accepted")
			return errors.New("failed to cast event")
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
	return nil
}

func (notifier *Nip47Notifier) notifySubscribers(ctx context.Context, notification *Notification, tags nostr.Tags, appId *uint) error {
	apps := []db.App{}

	// TODO: join apps and permissions
	err := notifier.db.Find(&apps).Error
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to list apps")
		return errors.New("failed to list apps")
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
				logger.Logger.WithFields(logrus.Fields{
					"notification": notification,
					"appId":        app.ID,
				}).WithError(err).Error("error deriving child key")
				return errors.New("failed to derive child key")
			}
		}

		appWalletPubKey, err := nostr.GetPublicKey(appWalletPrivKey)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"notification": notification,
				"appId":        app.ID,
			}).WithError(err).Error("Failed to calculate app wallet pub key")
			return errors.New("failed to calculate app wallet pubkey")
		}

		err = notifier.notifySubscriber(ctx, &app, notification, tags, appWalletPubKey, appWalletPrivKey, constants.ENCRYPTION_TYPE_NIP04)
		if err != nil {
			logger.Logger.WithError(err).Error("failed to notify subscriber (NIP-04)")
			return err
		}
		err = notifier.notifySubscriber(ctx, &app, notification, tags, appWalletPubKey, appWalletPrivKey, constants.ENCRYPTION_TYPE_NIP44_V2)
		if err != nil {
			logger.Logger.WithError(err).Error("failed to notify subscriber (NIP-44)")
			return err
		}
	}
	return nil
}

func (notifier *Nip47Notifier) notifySubscriber(ctx context.Context, app *db.App, notification *Notification, tags nostr.Tags, appWalletPubKey, appWalletPrivKey string, encryption string) error {
	logger.Logger.WithFields(logrus.Fields{
		"notification": notification,
		"appId":        app.ID,
		"encryption":   encryption,
	}).Debug("Notifying subscriber")

	var err error

	payloadBytes, err := json.Marshal(notification)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"notification": notification,
			"appId":        app.ID,
			"encryption":   encryption,
		}).WithError(err).Error("Failed to stringify notification")
		return err
	}

	nip47Cipher, err := cipher.NewNip47Cipher(encryption, app.AppPubkey, appWalletPrivKey)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"notification": notification,
			"appId":        app.ID,
			"encryption":   encryption,
		}).WithError(err).Error("Failed to initialize cipher")
		return err
	}

	msg, err := nip47Cipher.Encrypt(string(payloadBytes))
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"notification": notification,
			"appId":        app.ID,
			"encryption":   encryption,
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
		logger.Logger.WithFields(logrus.Fields{
			"notification": notification,
			"appId":        app.ID,
			"encryption":   encryption,
		}).WithError(err).Error("Failed to sign event")
		return err
	}

	publishResultChannel := notifier.pool.PublishMany(ctx, notifier.cfg.GetRelayUrls(), *event)

	publishSuccessful := false
	for result := range publishResultChannel {
		if result.Error == nil {
			publishSuccessful = true
		} else {
			logger.Logger.WithFields(logrus.Fields{
				"notification": notification,
				"appId":        app.ID,
				"relay":        result.RelayURL,
			}).WithError(result.Error).Error("failed to publish notification to relay")
		}
	}

	if !publishSuccessful {
		logger.Logger.WithFields(logrus.Fields{
			"notification": notification,
			"appId":        app.ID,
			"encryption":   encryption,
		}).WithError(err).Error("Failed to publish notification")
		return err
	}
	logger.Logger.WithFields(logrus.Fields{
		"appId":      app.ID,
		"encryption": encryption,
	}).Debug("Published notification event")
	return nil
}
