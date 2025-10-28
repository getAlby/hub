package service

import (
	"context"

	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/logger"
	"github.com/nbd-wtf/go-nostr"
)

type updateAppConsumer struct {
	events.EventSubscriber
	svc *service
}

// When a app is updated, re-publish the nip47 info event
func (s *updateAppConsumer) ConsumeEvent(ctx context.Context, event *events.Event, globalProperties map[string]interface{}) {
	if event.Event != "nwc_app_updated" {
		return
	}

	properties, ok := event.Properties.(map[string]interface{})
	if !ok {
		logger.Logger.WithField("event", event).Error("Failed to cast event.Properties to map")
		return
	}
	id, ok := properties["id"].(uint)
	if !ok {
		logger.Logger.WithField("event", event).Error("Failed to get app id")
		return
	}
	walletPrivKey, err := s.svc.keys.GetAppWalletKey(id)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to calculate app wallet priv key")
		return
	}
	walletPubKey, err := nostr.GetPublicKey(walletPrivKey)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to calculate app wallet pub key")
		return
	}

	if s.svc.keys.GetNostrPublicKey() != walletPubKey {
		// only need to re-publish the nip47 event info if it is not a legacy app connection (shared wallet pubkey)
		// (legacy app connection can be used for multiple apps - so it cannot be app-specific)
		for _, relayUrl := range s.svc.cfg.GetRelayUrls() {
			s.svc.nip47Service.EnqueueNip47InfoPublishRequest(id, walletPubKey, walletPrivKey, relayUrl)
		}
	}
}
