package models

import (
	"context"

	"github.com/getAlby/go-nostr"
)

type SimplePool interface {
	PublishMany(ctx context.Context, relayUrls []string, event nostr.Event) chan nostr.PublishResult
	QuerySingle(
		ctx context.Context,
		urls []string,
		filter nostr.Filter,
		opts ...nostr.SubscriptionOption,
	) *nostr.RelayEvent
}
