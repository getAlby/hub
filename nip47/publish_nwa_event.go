package nip47

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/nip47/cipher"
	"github.com/getAlby/hub/nip47/models"
	nostrmodels "github.com/getAlby/hub/nostr/models"
	"github.com/nbd-wtf/go-nostr"
)

func (svc *nip47Service) PublishNWAEvent(ctx context.Context, relay nostrmodels.Relay, nwaSecret string, appPubKey string, appWalletPubKey string, appWalletPrivKey string, lnClient lnclient.LNClient) (*nostr.Event, error) {
	cipher, err := cipher.NewNip47Cipher("1.0", appPubKey, appWalletPrivKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %s", err)
	}

	type nwaEvent struct {
		Secret string `json:"secret"`
		// TODO: add other properties
	}

	payloadBytes, err := json.Marshal(&nwaEvent{
		Secret: nwaSecret,
	})
	if err != nil {
		return nil, err
	}

	content, err := cipher.Encrypt(string(payloadBytes))
	if err != nil {
		return nil, err
	}

	ev := &nostr.Event{}
	ev.Kind = models.NWA_EVENT_KIND
	ev.Content = content
	ev.CreatedAt = nostr.Now()
	ev.PubKey = appWalletPubKey
	ev.Tags = nostr.Tags{[]string{"d", appPubKey}}
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
