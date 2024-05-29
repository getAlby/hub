package main

import (
	"context"

	"github.com/getAlby/nostr-wallet-connect/nip47"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

func (svc *Service) HandleGetInfoEvent(ctx context.Context, nip47Request *Nip47Request, requestEvent *RequestEvent, app *App, publishResponse func(*Nip47Response, nostr.Tags)) {

	resp := svc.checkPermission(nip47Request, requestEvent.NostrId, app, 0)
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
			ResultType: nip47Request.Method,
			Error: &Nip47Error{
				Code:    nip47.ERROR_INTERNAL,
				Message: err.Error(),
			},
		}, nostr.Tags{})
		return
	}

	network := info.Network
	// Some implementations return "bitcoin" while NIP47 expects "mainnet"
	if network == "bitcoin" {
		network = "mainnet"
	}

	responsePayload := &Nip47GetInfoResponse{
		Alias:       info.Alias,
		Color:       info.Color,
		Pubkey:      info.Pubkey,
		Network:     network,
		BlockHeight: info.BlockHeight,
		BlockHash:   info.BlockHash,
		Methods:     svc.GetMethods(app),
	}

	publishResponse(&Nip47Response{
		ResultType: nip47Request.Method,
		Result:     responsePayload,
	}, nostr.Tags{})
}
