package models

import (
	"context"

	"github.com/nbd-wtf/go-nostr"
)

type Relay interface {
	Publish(ctx context.Context, event nostr.Event) error
}

type SimplePool interface {
	PublishMany(ctx context.Context, relayUrls []string, event nostr.Event) chan nostr.PublishResult
}
