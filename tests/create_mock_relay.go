package tests

import (
	"context"

	"github.com/getAlby/hub/logger"
	"github.com/nbd-wtf/go-nostr"
)

type mockRelay struct {
	PublishedEvent *nostr.Event
}

func NewMockRelay() *mockRelay {
	return &mockRelay{}
}

func (relay *mockRelay) Publish(ctx context.Context, event nostr.Event) error {
	logger.Logger.WithField("event", event).Info("Mock Publishing event")
	relay.PublishedEvent = &event
	return nil
}
