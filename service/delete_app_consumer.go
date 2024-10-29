package service

import (
	"context"

	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/logger"
	"github.com/nbd-wtf/go-nostr"
)

type deleteAppConsumer struct {
	events.EventSubscriber
	walletPubkey      string
	relay             *nostr.Relay
	nostrSubscription *nostr.Subscription
	svc               *service
	infoEventId       string
}

// When an app is deleted, unsubscribe from events for that app on the relay
// and publish a deletion event for that app's info event
func (s *deleteAppConsumer) ConsumeEvent(ctx context.Context, event *events.Event, globalProperties map[string]interface{}) {
	if event.Event != "app_deleted" {
		return
	}
	properties, ok := event.Properties.(map[string]interface{})
	if !ok {
		logger.Logger.WithField("event", event).Error("Failed to cast event.Properties to map")
		return
	}
	id, _ := properties["id"].(uint)

	walletPrivKey, err := s.svc.keys.GetAppWalletKey(id)
	if err != nil {
		logger.Logger.WithError(err).WithField("id", id).Error("Failed to calculate app wallet priv key")
	}
	walletPubKey, _ := nostr.GetPublicKey(walletPrivKey)
	if s.walletPubkey == walletPubKey {
		s.nostrSubscription.Unsub()
		err := s.svc.nip47Service.PublishNip47InfoDeletion(ctx, s.relay, walletPubKey, walletPrivKey, s.infoEventId)
		if err != nil {
			logger.Logger.WithError(err).WithField("event", event).Error("Failed to publish nip47 info deletion")
		}
	}
}
