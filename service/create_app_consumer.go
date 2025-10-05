package service

import (
	"context"
	"errors"

	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/logger"
)

type createAppConsumer struct {
	events.EventSubscriber
	svc  *service
	pool *nostr.SimplePool
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
	s.svc.nip47Service.EnqueueNip47InfoPublishRequest(id, walletPubKey, walletPrivKey)

	go func() {
		err = s.svc.startAppWalletSubscription(ctx, s.pool, walletPubKey)
		if err != nil && !errors.Is(err, context.Canceled) {
			logger.Logger.WithError(err).WithFields(logrus.Fields{
				"app_id": id}).Error("Failed to subscribe to wallet")
		}
		logger.Logger.WithFields(logrus.Fields{
			"app_id": id}).Info("App Nostr Subscription ended")
	}()
}
