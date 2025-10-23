package nip47

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/cipher"
	"github.com/getAlby/hub/nip47/models"
	nostrmodels "github.com/getAlby/hub/nostr/models"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

type Nip47InfoPublishRequest struct {
	AppId            uint
	AppWalletPubKey  string
	AppWalletPrivKey string
	RelayUrl         string
	Attempt          uint32
}

type nip47InfoPublishQueue struct {
	channel chan *Nip47InfoPublishRequest
}

func NewNip47InfoPublishQueue() *nip47InfoPublishQueue {
	return &nip47InfoPublishQueue{
		channel: make(chan *Nip47InfoPublishRequest),
	}
}

func (q *nip47InfoPublishQueue) AddToQueue(req *Nip47InfoPublishRequest) {
	// thread will be blocked if the channel is full, so execute in a separate goroutine
	go func() {
		q.channel <- req
	}()
}

func (q *nip47InfoPublishQueue) Channel() <-chan *Nip47InfoPublishRequest {
	return q.channel
}

func (svc *nip47Service) GetNip47Info(ctx context.Context, relay *nostr.Relay, appWalletPubKey string) (*nostr.Event, error) {
	filter := nostr.Filter{
		Kinds:   []int{models.INFO_EVENT_KIND},
		Authors: []string{appWalletPubKey},
		Limit:   1,
	}

	events, err := relay.QuerySync(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(events) == 0 {
		return nil, nil
	}

	return events[0], nil
}

func (svc *nip47Service) PublishNip47Info(ctx context.Context, pool nostrmodels.SimplePool, appId uint, appWalletPubKey string, appWalletPrivKey string, relayUrl string, lnClient lnclient.LNClient) (*nostr.Event, error) {
	var capabilities []string
	var permitsNotifications bool
	tags := nostr.Tags{[]string{"encryption", cipher.SUPPORTED_ENCRYPTIONS}}

	if svc.keys.GetNostrPublicKey() == appWalletPubKey {
		// legacy app, so return lnClient.GetSupportedNIP47Methods()
		capabilities = lnClient.GetSupportedNIP47Methods()
		permitsNotifications = true
	} else {
		app := db.App{}
		err := svc.db.First(&app, &db.App{
			ID: appId,
		}).Error
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"walletPubKey": appWalletPubKey,
			}).WithError(err).Error("Failed to find app for wallet pubkey")
			return nil, err
		}
		capabilities = svc.permissionsService.GetPermittedMethods(&app, lnClient)
		permitsNotifications = svc.permissionsService.PermitsNotifications(&app)

		// NWA: associate the info event with the app so that the app can receive the wallet pubkey
		tags = append(tags, []string{"p", app.AppPubkey})
	}
	if permitsNotifications && len(lnClient.GetSupportedNIP47NotificationTypes()) > 0 {
		capabilities = append(capabilities, "notifications")
		tags = append(tags, []string{"notifications", strings.Join(lnClient.GetSupportedNIP47NotificationTypes(), " ")})
	}

	ev := &nostr.Event{}
	ev.Kind = models.INFO_EVENT_KIND
	ev.Content = strings.Join(capabilities, " ")
	ev.CreatedAt = nostr.Now()
	ev.PubKey = appWalletPubKey
	ev.Tags = tags
	err := ev.Sign(appWalletPrivKey)
	if err != nil {
		return nil, err
	}

	// publish to a single relay so that we can requeue failed publishes on a relay level
	publishResultChannel := pool.PublishMany(ctx, []string{relayUrl}, *ev)

	publishSuccessful := false
	for v := range publishResultChannel {
		if v.Error == nil {
			publishSuccessful = true
		} else {
			logger.Logger.WithFields(logrus.Fields{
				"appId": appId,
				"relay": v.RelayURL,
			}).WithError(v.Error).Error("failed to publish nip47 info to relay")
		}
	}
	if !publishSuccessful {
		return nil, fmt.Errorf("nostr publish not successful: %s", err)
	}
	logger.Logger.WithField("wallet_pubkey", appWalletPubKey).Debug("published info event")
	return ev, nil
}

func (svc *nip47Service) PublishNip47InfoDeletion(ctx context.Context, relay nostrmodels.Relay, appWalletPubKey string, appWalletPrivKey string, infoEventId string) error {
	ev := &nostr.Event{}
	ev.Kind = nostr.KindDeletion
	ev.Content = "deleting nip47 info since app connection for this key was deleted"
	ev.Tags = nostr.Tags{[]string{"e", infoEventId}, []string{"k", strconv.Itoa(models.INFO_EVENT_KIND)}}
	ev.CreatedAt = nostr.Now()
	ev.PubKey = appWalletPubKey
	err := ev.Sign(appWalletPrivKey)
	if err != nil {
		return err
	}
	err = relay.Publish(ctx, *ev)
	if err != nil {
		return fmt.Errorf("nostr publish not successful: %s", err)
	}
	return nil
}
