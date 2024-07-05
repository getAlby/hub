package nip47

import (
	"context"
	"fmt"
	"strings"

	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/nip47/models"
	"github.com/nbd-wtf/go-nostr"
)

func (svc *nip47Service) PublishNip47Info(ctx context.Context, relay *nostr.Relay, lnClient lnclient.LNClient) error {
	capabilities := lnClient.GetSupportedNIP47Methods()
	if len(lnClient.GetSupportedNIP47NotificationTypes()) > 0 {
		capabilities = append(capabilities, "notifications")
	}

	ev := &nostr.Event{}
	ev.Kind = models.INFO_EVENT_KIND
	ev.Content = strings.Join(capabilities, " ")
	ev.CreatedAt = nostr.Now()
	ev.PubKey = svc.keys.GetNostrPublicKey()
	ev.Tags = nostr.Tags{[]string{"notifications", strings.Join(lnClient.GetSupportedNIP47NotificationTypes(), " ")}}
	err := ev.Sign(svc.keys.GetNostrSecretKey())
	if err != nil {
		return err
	}
	err = relay.Publish(ctx, *ev)
	if err != nil {
		return fmt.Errorf("nostr publish not successful: %s", err)
	}
	return nil
}
