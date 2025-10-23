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
	"github.com/sirupsen/logrus"
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
	StartNotifier(ctx context.Context, pool *nostr.SimplePool)
	StartNip47InfoPublisher(ctx context.Context, pool *nostr.SimplePool, lnClient lnclient.LNClient)
	HandleEvent(ctx context.Context, pool nostrmodels.SimplePool, event *nostr.Event, lnClient lnclient.LNClient)
	GetNip47Info(ctx context.Context, relay *nostr.Relay, appWalletPubKey string) (*nostr.Event, error)
	PublishNip47Info(ctx context.Context, pool nostrmodels.SimplePool, appId uint, appWalletPubKey string, appWalletPrivKey string, relayUrl string, lnClient lnclient.LNClient) (*nostr.Event, error)
	PublishNip47InfoDeletion(ctx context.Context, relay nostrmodels.Relay, appWalletPubKey string, appWalletPrivKey string, infoEventId string) error
	CreateResponse(initialEvent *nostr.Event, content interface{}, tags nostr.Tags, cipher *cipher.Nip47Cipher, walletPrivKey string) (result *nostr.Event, err error)
	EnqueueNip47InfoPublishRequest(appId uint, appWalletPubKey, appWalletPrivKey, relayUrl string)
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

// FIXME: this is not how it works now since we use a pool that handles reconnection logic
// The notifier is decoupled from the notification queue
// so that if Alby Hub disconnects from the relay, it will wait to reconnect
// to send notifications rather than dropping them
func (svc *nip47Service) StartNotifier(ctx context.Context, pool *nostr.SimplePool) {
	nip47Notifier := notifications.NewNip47Notifier(pool, svc.db, svc.cfg, svc.keys, svc.permissionsService)
	go func() {
		for {
			select {
			case <-ctx.Done():
				// app exited
				return
			case event := <-svc.nip47NotificationQueue.Channel():
				logger.Logger.WithField("event", event).Debug("Consuming event from notification queue")
				err := nip47Notifier.ConsumeEvent(ctx, event)
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

func (svc *nip47Service) EnqueueNip47InfoPublishRequest(appId uint, appWalletPubKey, appWalletPrivKey, relayUrl string) {
	svc.enqueueNip47InfoPublishRequestWithAttempt(appId, appWalletPubKey, appWalletPrivKey, relayUrl, 0)
}

func (svc *nip47Service) enqueueNip47InfoPublishRequestWithAttempt(appId uint, appWalletPubKey, appWalletPrivKey, relayUrl string, attempt uint32) {
	svc.nip47InfoPublishQueue.AddToQueue(&Nip47InfoPublishRequest{
		AppId:            appId,
		AppWalletPubKey:  appWalletPubKey,
		AppWalletPrivKey: appWalletPrivKey,
		RelayUrl:         relayUrl,
		Attempt:          attempt,
	})
}

func (svc *nip47Service) StartNip47InfoPublisher(ctx context.Context, pool *nostr.SimplePool, lnClient lnclient.LNClient) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				// relay disconnected
				return
			case req := <-svc.nip47InfoPublishQueue.Channel():
				_, err := svc.PublishNip47Info(ctx, pool, req.AppId, req.AppWalletPubKey, req.AppWalletPrivKey, req.RelayUrl, lnClient)
				if err != nil {
					logger.Logger.WithError(err).WithFields(logrus.Fields{
						"wallet_pubkey": req.AppWalletPubKey,
						"relay_url":     req.RelayUrl,
					}).Error("Failed to publish NIP47 info from queue")

					// wait and then re-add the item to the queue
					// done async to ensure an offline relay does not delay
					// the publishing of newly created app connections
					go func() {
						time.Sleep((5 * time.Duration(req.Attempt+1)) * time.Second)
						svc.enqueueNip47InfoPublishRequestWithAttempt(req.AppId, req.AppWalletPubKey, req.AppWalletPrivKey, req.RelayUrl, req.Attempt+1)
					}()
				}
			}
		}
	}()
}
