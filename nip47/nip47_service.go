package nip47

import (
	"context"

	"github.com/getAlby/hub/alby"
	"github.com/getAlby/hub/apps"
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
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
	cfg                    config.Config
	keys                   keys.Keys
	db                     *gorm.DB
	eventPublisher         events.EventPublisher
}

type Nip47Service interface {
	events.EventSubscriber
	StartNotifier(ctx context.Context, relay *nostr.Relay, lnClient lnclient.LNClient)
	HandleEvent(ctx context.Context, relay nostrmodels.Relay, event *nostr.Event, lnClient lnclient.LNClient)
	GetNip47Info(ctx context.Context, relay *nostr.Relay, appWalletPubKey string) (*nostr.Event, error)
	PublishNip47Info(ctx context.Context, relay nostrmodels.Relay, appWalletPubKey string, appWalletPrivKey string, lnClient lnclient.LNClient) (*nostr.Event, error)
	PublishNip47InfoDeletion(ctx context.Context, relay nostrmodels.Relay, appWalletPubKey string, appWalletPrivKey string, infoEventId string) error
	CreateResponse(initialEvent *nostr.Event, content interface{}, tags nostr.Tags, cipher *cipher.Nip47Cipher, walletPrivKey string) (result *nostr.Event, err error)
}

func NewNip47Service(db *gorm.DB, cfg config.Config, keys keys.Keys, eventPublisher events.EventPublisher, albyOAuthSvc alby.AlbyOAuthService) *nip47Service {
	return &nip47Service{
		nip47NotificationQueue: notifications.NewNip47NotificationQueue(),
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

func (svc *nip47Service) StartNotifier(ctx context.Context, relay *nostr.Relay, lnClient lnclient.LNClient) {
	nip47Notifier := notifications.NewNip47Notifier(relay, svc.db, svc.cfg, svc.keys, svc.permissionsService, svc.transactionsService, lnClient)
	go func() {
		for {
			select {
			case <-ctx.Done():
				// subscription ended
				return
			case event := <-svc.nip47NotificationQueue.Channel():
				nip47Notifier.ConsumeEvent(ctx, event)
			}
		}
	}()
}
