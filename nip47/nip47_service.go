package nip47

import (
	"context"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/nip47/notifications"
	permissions "github.com/getAlby/hub/nip47/permissions"
	"github.com/getAlby/hub/service/keys"
	"github.com/getAlby/hub/transactions"
	"github.com/nbd-wtf/go-nostr"
	"gorm.io/gorm"
)

type nip47Service struct {
	permissionsService     permissions.PermissionsService
	transactionsService    transactions.TransactionsService
	nip47NotificationQueue notifications.Nip47NotificationQueue
	cfg                    config.Config
	keys                   keys.Keys
	db                     *gorm.DB
	eventPublisher         events.EventPublisher
}

type Nip47Service interface {
	events.EventSubscriber
	StartNotifier(ctx context.Context, relay *nostr.Relay, lnClient lnclient.LNClient)
	HandleEvent(ctx context.Context, sub *nostr.Subscription, event *nostr.Event, lnClient lnclient.LNClient)
	PublishNip47Info(ctx context.Context, relay *nostr.Relay, lnClient lnclient.LNClient) error
	CreateResponse(initialEvent *nostr.Event, content interface{}, tags nostr.Tags, ss []byte) (result *nostr.Event, err error)
}

func NewNip47Service(db *gorm.DB, cfg config.Config, keys keys.Keys, eventPublisher events.EventPublisher) *nip47Service {
	return &nip47Service{
		nip47NotificationQueue: notifications.NewNip47NotificationQueue(),
		cfg:                    cfg,
		db:                     db,
		permissionsService:     permissions.NewPermissionsService(db, eventPublisher),
		transactionsService:    transactions.NewTransactionsService(db),
		eventPublisher:         eventPublisher,
		keys:                   keys,
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
