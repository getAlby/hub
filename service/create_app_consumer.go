package service

import (
	"context"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/logger"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

type createAppConsumer struct {
	events.EventSubscriber
	svc   *service
	relay *nostr.Relay
}

// When a new app is created, subscribe to it on the relay
func (s *createAppConsumer) ConsumeEvent(ctx context.Context, event *events.Event, globalProperties map[string]interface{}) {
	if event.Event != "nwc_app_created" {
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
	nwaSecret, ok := properties["nwa_secret"].(string)
	if !ok {
		logger.Logger.WithField("event", event).Error("Failed to get nwa secret")
		return
	}

	app := db.App{}
	err := s.svc.db.First(&app, &db.App{
		ID: id,
	}).Error
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"id": id,
		}).WithError(err).Error("Failed to find app for id")
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

	go func() {
		_, err := s.svc.GetNip47Service().PublishNip47Info(ctx, s.relay, walletPubKey, walletPrivKey, s.svc.lnClient)
		if err != nil {
			logger.Logger.WithError(err).Error("Could not publish NIP47 info")
		}
		if nwaSecret != "" {
			_, err = s.svc.GetNip47Service().PublishNWAEvent(ctx, s.relay, nwaSecret, app.AppPubkey, walletPubKey, walletPrivKey, s.svc.lnClient)
			if err != nil {
				logger.Logger.WithError(err).Error("Could not publish NWA event")
			}
		}
		err = s.svc.startAppWalletSubscription(ctx, s.relay, walletPubKey)
		if err != nil {
			logger.Logger.WithError(err).WithFields(logrus.Fields{
				"app_id": id}).Error("Failed to subscribe to wallet")
		}
		logger.Logger.WithFields(logrus.Fields{
			"app_id": id}).Info("App Nostr Subscription ended")
	}()
}
