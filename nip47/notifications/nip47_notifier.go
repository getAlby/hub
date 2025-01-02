package notifications

import (
	"context"
	"encoding/json"
	"strings"

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
	"github.com/nbd-wtf/go-nostr/nip44"
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
	switch event.Event {
	case "nwc_payment_received":
		transaction, ok := event.Properties.(*db.Transaction)
		if !ok {
			logger.Logger.WithField("event", event).Error("Failed to cast event")
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
			logger.Logger.WithField("event", event).Error("Failed to cast event")
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
	err := notifier.db.Find(&apps).Error
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to list apps")
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
				logger.Logger.WithFields(logrus.Fields{
					"notification": notification,
					"appId":        app.ID,
				}).WithError(err).Error("error deriving child key")
				return
			}
		}

		appWalletPubKey, err := nostr.GetPublicKey(appWalletPrivKey)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"notification": notification,
				"appId":        app.ID,
			}).WithError(err).Error("Failed to calculate app wallet pub key")
			return
		}

		if strings.Contains(models.SUPPORTED_VERSIONS, "0.0") {
			go notifier.notifySubscriber(ctx, &app, notification, tags, appWalletPubKey, appWalletPrivKey, true)
		}
		if strings.Contains(models.SUPPORTED_VERSIONS, "1.0") {
			go notifier.notifySubscriber(ctx, &app, notification, tags, appWalletPubKey, appWalletPrivKey, false)
		}
	}
}

func (notifier *Nip47Notifier) notifySubscriber(ctx context.Context, app *db.App, notification *Notification, tags nostr.Tags, appWalletPubKey, appWalletPrivKey string, legacy bool) {
	logger.Logger.WithFields(logrus.Fields{
		"notification": notification,
		"appId":        app.ID,
	}).Debug("Notifying subscriber")

	var err error

	payloadBytes, err := json.Marshal(notification)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"notification": notification,
			"appId":        app.ID,
		}).WithError(err).Error("Failed to stringify notification")
		return
	}

	var msg string
	if legacy {
		ss, err := nip04.ComputeSharedSecret(app.AppPubkey, appWalletPrivKey)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"notification": notification,
				"appId":        app.ID,
			}).WithError(err).Error("Failed to compute shared secret")
			return
		}
		msg, err = nip04.Encrypt(string(payloadBytes), ss)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"notification": notification,
				"appId":        app.ID,
			}).WithError(err).Error("Failed to encrypt notification payload")
			return
		}
	} else {
		ck, err := nip44.GenerateConversationKey(app.AppPubkey, appWalletPrivKey)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"notification": notification,
				"appId":        app.ID,
			}).WithError(err).Error("Failed to compute conversation key")
			return
		}
		msg, err = nip44.Encrypt(string(payloadBytes), ck)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"notification": notification,
				"appId":        app.ID,
			}).WithError(err).Error("Failed to encrypt notification payload")
			return
		}
	}

	allTags := nostr.Tags{[]string{"p", app.AppPubkey}}
	allTags = append(allTags, tags...)

	event := &nostr.Event{
		PubKey:    appWalletPubKey,
		CreatedAt: nostr.Now(),
		Tags:      allTags,
		Content:   msg,
	}
	if legacy {
		event.Kind = models.LEGACY_NOTIFICATION_KIND
	} else {
		event.Kind = models.NOTIFICATION_KIND
	}

	err = event.Sign(appWalletPrivKey)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"notification": notification,
			"appId":        app.ID,
			"isLegacy":     legacy,
		}).WithError(err).Error("Failed to sign event")
		return
	}

	err = notifier.relay.Publish(ctx, *event)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"notification": notification,
			"appId":        app.ID,
			"isLegacy":     legacy,
		}).WithError(err).Error("Failed to publish notification")
		return
	}
	logger.Logger.WithFields(logrus.Fields{
		"appId":    app.ID,
		"isLegacy": legacy,
	}).Debug("Published notification event")
}
