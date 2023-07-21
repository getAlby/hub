package main

import (
	"context"

	"github.com/nbd-wtf/go-nostr"
)

func (svc *Service) HandleGetBalanceEvent(ctx context.Context, request *Nip47Request, event *nostr.Event, app App, ss []byte) (result *nostr.Event, err error) {
	return nil, nil
}
