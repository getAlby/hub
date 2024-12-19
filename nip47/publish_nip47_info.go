package nip47

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	nostrmodels "github.com/getAlby/hub/nostr/models"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

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

func (svc *nip47Service) PublishNip47Info(ctx context.Context, relay nostrmodels.Relay, appWalletPubKey string, appWalletPrivKey string, lnClient lnclient.LNClient) (*nostr.Event, error) {
	// TODO: should the capabilities be based on the app permissions? (for non-legacy apps)
	app := db.App{}
	err := svc.db.First(&app, &db.App{
		WalletPubkey: &appWalletPubKey,
	}).Error
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"walletPubKey": appWalletPubKey,
		}).WithError(err).Error("Failed to find app for wallet pubkey")
		return nil, err
	}

	capabilities := svc.permissionsService.GetPermittedMethods(&app, lnClient)
	if len(lnClient.GetSupportedNIP47NotificationTypes()) > 0 {
		capabilities = append(capabilities, "notifications")
	}

	ev := &nostr.Event{}
	ev.Kind = models.INFO_EVENT_KIND
	ev.Content = strings.Join(capabilities, " ")
	ev.CreatedAt = nostr.Now()
	ev.PubKey = appWalletPubKey
	ev.Tags = nostr.Tags{[]string{"notifications", strings.Join(lnClient.GetSupportedNIP47NotificationTypes(), " ")}}
	err = ev.Sign(appWalletPrivKey)
	if err != nil {
		return nil, err
	}
	err = relay.Publish(ctx, *ev)
	if err != nil {
		return nil, fmt.Errorf("nostr publish not successful: %s", err)
	}
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
