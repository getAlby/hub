package main

import (
	"context"

	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

func (svc *Service) HandleMakeInvoiceEvent(ctx context.Context, request *Nip47Request, requestEvent *RequestEvent, app *App, publishResponse func(*Nip47Response, nostr.Tags)) {

	makeInvoiceParams := &Nip47MakeInvoiceParams{}
	resp := svc.decodeNip47Request(request, requestEvent, app, makeInvoiceParams)
	if resp != nil {
		publishResponse(resp, nostr.Tags{})
		return
	}

	resp = svc.checkPermission(request, requestEvent.NostrId, app, 0)
	if resp != nil {
		publishResponse(resp, nostr.Tags{})
		return
	}

	if makeInvoiceParams.Description != "" && makeInvoiceParams.DescriptionHash != "" {
		svc.Logger.WithFields(logrus.Fields{
			"requestEventNostrId": requestEvent.NostrId,
			"appId":               app.ID,
		}).Errorf("Only one of description, description_hash can be provided")

		publishResponse(&Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    NIP_47_OTHER,
				Message: "Only one of description, description_hash can be provided",
			},
		}, nostr.Tags{})
		return
	}

	svc.Logger.WithFields(logrus.Fields{
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
		svc.Logger.WithFields(logrus.Fields{
			"requestEventNostrId": requestEvent.NostrId,
			"appId":               app.ID,
			"amount":              makeInvoiceParams.Amount,
			"description":         makeInvoiceParams.Description,
			"descriptionHash":     makeInvoiceParams.DescriptionHash,
			"expiry":              makeInvoiceParams.Expiry,
		}).Infof("Failed to make invoice: %v", err)

		publishResponse(&Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: err.Error(),
			},
		}, nostr.Tags{})
		return
	}

	responsePayload := &Nip47MakeInvoiceResponse{
		Nip47Transaction: *transaction,
	}

	publishResponse(&Nip47Response{
		ResultType: request.Method,
		Result:     responsePayload,
	}, nostr.Tags{})
}
