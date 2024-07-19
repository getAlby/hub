package models

import (
	"context"

	"github.com/nbd-wtf/go-nostr"
)

type Relay interface {
	Publish(ctx context.Context, event nostr.Event) error
}
