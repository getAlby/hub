package nip47

import (
	"context"
	"time"

	"github.com/getAlby/hub/alby"
	"github.com/getAlby/hub/apps"
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/cipher"
	"github.com/getAlby/hub/nip47/notifications"
	"github.com/getAlby/hub/nip47/permissions"
	nostrmodels "github.com/getAlby/hub/nostr/models"
	"github.com/getAlby/hub/service/keys"
	"github.com/getAlby/hub/transactions"
	"github.com/nbd-wtf/go-nostr"
	"gorm.io/gorm"
)

type nip47Service struct {
	permissionsService     permissions.PermissionsService
	transactionsService    transactions.TransactionsService
	appsService            apps.AppsService
	albyOAuthSvc           alby.AlbyOAuthService
	nip47NotificationQueue notifications.Nip47NotificationQueue
	nip47InfoPublishQueue  *nip47InfoPublishQueue
	cfg                    config.Config
	keys                   keys.Keys
	db                     *gorm.DB
	eventPublisher         events.EventPublisher
}

type Nip47Service interface {
	events.EventSubscriber
	StartNotifier(relay *nostr.Relay)
	StartNip47InfoPublisher(relay *nostr.Relay, lnClient lnclient.LNClient)
	HandleEvent(ctx context.Context, relay nostrmodels.Relay, event *nostr.Event, lnClient lnclient.LNClient)
	GetNip47Info(ctx context.Context, relay *nostr.Relay, appWalletPubKey string) (*nostr.Event, error)
	PublishNip47Info(ctx context.Context, relay nostrmodels.Relay, appWalletPubKey string, appWalletPrivKey string, lnClient lnclient.LNClient) (*nostr.Event, error)
	PublishNip47InfoDeletion(ctx context.Context, relay nostrmodels.Relay, appWalletPubKey string, appWalletPrivKey string, infoEventId string) error
	CreateResponse(initialEvent *nostr.Event, content interface{}, tags nostr.Tags, cipher *cipher.Nip47Cipher, walletPrivKey string) (result *nostr.Event, err error)
	EnqueueNip47InfoPublishRequest(AppWalletPubKey, AppWalletPrivKey string)
}

func NewNip47Service(db *gorm.DB, cfg config.Config, keys keys.Keys, eventPublisher events.EventPublisher, albyOAuthSvc alby.AlbyOAuthService) *nip47Service {
	return &nip47Service{
		nip47NotificationQueue: notifications.NewNip47NotificationQueue(),
		nip47InfoPublishQueue:  NewNip47InfoPublishQueue(),
		cfg:                    cfg,
		db:                     db,
		permissionsService:     permissions.NewPermissionsService(db, eventPublisher),
		transactionsService:    transactions.NewTransactionsService(db, eventPublisher),
		appsService:            apps.NewAppsService(db, eventPublisher, keys, cfg),
		eventPublisher:         eventPublisher,
		keys:                   keys,
		albyOAuthSvc:           albyOAuthSvc,
	}
}

func (svc *nip47Service) ConsumeEvent(ctx context.Context, event *events.Event, globalProperties map[string]interface{}) {
	svc.nip47NotificationQueue.AddToQueue(event)
}

// The notifier is decoupled from the notification queue
// so that if Alby Hub disconnects from the relay, it will wait to reconnect
// to send notifications rather than dropping them
func (svc *nip47Service) StartNotifier(relay *nostr.Relay) {
	nip47Notifier := notifications.NewNip47Notifier(relay, svc.db, svc.cfg, svc.keys, svc.permissionsService)
	go func() {
		for {
			select {
			case <-relay.Context().Done():
				// relay disconnected
				return
			case event := <-svc.nip47NotificationQueue.Channel():
				logger.Logger.WithField("event", event).Debug("Consuming event from notification queue")
				err := nip47Notifier.ConsumeEvent(relay.Context(), event)
				if err != nil {
					logger.Logger.WithError(err).WithField("event", event).Error("Failed to consume event from notification queue")
					// wait and then re-add the item to the queue
					time.Sleep(5 * time.Second)
					svc.nip47NotificationQueue.AddToQueue(event)
				}
			}
		}
	}()
}

func (svc *nip47Service) EnqueueNip47InfoPublishRequest(AppWalletPubKey, AppWalletPrivKey string) {
	svc.nip47InfoPublishQueue.AddToQueue(&Nip47InfoPublishRequest{
		AppWalletPubKey:  AppWalletPubKey,
		AppWalletPrivKey: AppWalletPrivKey,
	})
}

func (svc *nip47Service) StartNip47InfoPublisher(relay *nostr.Relay, lnClient lnclient.LNClient) {
	go func() {
		for {
			select {
			case <-relay.Context().Done():
				// relay disconnected
				return
			case req := <-svc.nip47InfoPublishQueue.Channel():
				_, err := svc.PublishNip47Info(relay.Context(), relay, req.AppWalletPubKey, req.AppWalletPrivKey, lnClient)
				if err != nil {
					logger.Logger.WithError(err).WithField("wallet_pubkey", req.AppWalletPubKey).Error("Failed to publish NIP47 info from queue")
					// wait and then re-add the item to the queue
					time.Sleep(5 * time.Second)
					svc.EnqueueNip47InfoPublishRequest(req.AppWalletPubKey, req.AppWalletPrivKey)
				}
			}
		}
	}()
}
