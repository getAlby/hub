package nip47

import (
	"context"
	"fmt"

	"github.com/getAlby/nostr-wallet-connect/nip47/models"
	"github.com/getAlby/nostr-wallet-connect/nip47/notifications"
	"github.com/nbd-wtf/go-nostr"
)

func (svc *nip47Service) PublishNip47Info(ctx context.Context, relay *nostr.Relay) error {
	ev := &nostr.Event{}
	ev.Kind = models.INFO_EVENT_KIND
	ev.Content = models.CAPABILITIES
	ev.CreatedAt = nostr.Now()
	ev.PubKey = svc.keys.GetNostrPublicKey()
	ev.Tags = nostr.Tags{[]string{"notifications", notifications.NOTIFICATION_TYPES}}
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
