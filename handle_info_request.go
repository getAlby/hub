package main

import (
	"context"

	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

func (svc *Service) HandleGetInfoEvent(ctx context.Context, request *Nip47Request, requestEvent *RequestEvent, app *App, publishResponse func(*Nip47Response, nostr.Tags)) {

	resp := svc.checkPermission(request, requestEvent.NostrId, app, 0)
	if resp != nil {
		publishResponse(resp, nostr.Tags{})
		return
	}

	svc.Logger.WithFields(logrus.Fields{
		"requestEventNostrId": requestEvent.NostrId,
		"appId":               app.ID,
	}).Info("Fetching node info")

	info, err := svc.lnClient.GetInfo(ctx)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"requestEventNostrId": requestEvent.NostrId,
			"appId":               app.ID,
		}).Infof("Failed to fetch node info: %v", err)

		publishResponse(&Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: err.Error(),
			},
		}, nostr.Tags{})
		return
	}

	responsePayload := &Nip47GetInfoResponse{
		Alias:       info.Alias,
		Color:       info.Color,
		Pubkey:      info.Pubkey,
		Network:     info.Network,
		BlockHeight: info.BlockHeight,
		BlockHash:   info.BlockHash,
		Methods:     svc.GetMethods(app),
	}

	publishResponse(&Nip47Response{
		ResultType: request.Method,
		Result:     responsePayload,
	}, nostr.Tags{})
}
