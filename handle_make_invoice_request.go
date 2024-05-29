package main

import (
	"context"

	"github.com/getAlby/nostr-wallet-connect/db"
	"github.com/getAlby/nostr-wallet-connect/nip47"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

func (svc *Service) HandleMakeInvoiceEvent(ctx context.Context, nip47Request *nip47.Request, requestEvent *db.RequestEvent, app *db.App, publishResponse func(*nip47.Response, nostr.Tags)) {

	makeInvoiceParams := &nip47.MakeInvoiceParams{}
	resp := svc.decodeNip47Request(nip47Request, requestEvent, app, makeInvoiceParams)
	if resp != nil {
		publishResponse(resp, nostr.Tags{})
		return
	}

	resp = svc.checkPermission(nip47Request, requestEvent.NostrId, app, 0)
	if resp != nil {
		publishResponse(resp, nostr.Tags{})
		return
	}

	svc.logger.WithFields(logrus.Fields{
		"requestEventNostrId": requestEvent.NostrId,
		"appId":               app.ID,
		"amount":              makeInvoiceParams.Amount,
		"description":         makeInvoiceParams.Description,
		"descriptionHash":     makeInvoiceParams.DescriptionHash,
		"expiry":              makeInvoiceParams.Expiry,
	}).Info("Making invoice")

	expiry := makeInvoiceParams.Expiry
	if expiry == 0 {
		expiry = 86400
	}

	transaction, err := svc.lnClient.MakeInvoice(ctx, makeInvoiceParams.Amount, makeInvoiceParams.Description, makeInvoiceParams.DescriptionHash, expiry)
	if err != nil {
		svc.logger.WithFields(logrus.Fields{
			"requestEventNostrId": requestEvent.NostrId,
			"appId":               app.ID,
			"amount":              makeInvoiceParams.Amount,
			"description":         makeInvoiceParams.Description,
			"descriptionHash":     makeInvoiceParams.DescriptionHash,
			"expiry":              makeInvoiceParams.Expiry,
		}).Infof("Failed to make invoice: %v", err)

		publishResponse(&nip47.Response{
			ResultType: nip47Request.Method,
			Error: &nip47.Error{
				Code:    nip47.ERROR_INTERNAL,
				Message: err.Error(),
			},
		}, nostr.Tags{})
		return
	}

	responsePayload := &nip47.MakeInvoiceResponse{
		Transaction: *transaction,
	}

	publishResponse(&nip47.Response{
		ResultType: nip47Request.Method,
		Result:     responsePayload,
	}, nostr.Tags{})
}
