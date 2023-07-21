package main

import (
	"context"

	"github.com/nbd-wtf/go-nostr"
)

func (svc *Service) HandleGetBalanceEvent(ctx context.Context, request *Nip47Request, event *nostr.Event, app App, ss []byte) (result *nostr.Event, err error) {
	//no need to look at the parameters here
	balance, err := svc.lnClient.GetBalance(ctx, event.PubKey)
	if err != nil {
		//todo
	}
	//todo: add budget info from permission db call
	//construct response
	responsePayload := &Nip47BalanceResponse{
		Balance: balance,
	}
	return svc.createResponse(event, responsePayload, ss)
}
