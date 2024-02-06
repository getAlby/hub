package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

func (svc *Service) HandleMakeInvoiceEvent(ctx context.Context, request *Nip47Request, requestEvent *NostrEvent, app *App) (result *Nip47Response, err error) {

	// TODO: move to a shared function
	hasPermission, code, message := svc.hasPermission(app, request.Method, 0)

	if !hasPermission {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   requestEvent.NostrId,
			"eventKind": requestEvent.Kind,
			"appId":     app.ID,
		}).Errorf("App does not have permission: %s %s", code, message)

		return &Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    code,
				Message: message,
			}}, nil
	}

	// TODO: move to a shared generic function
	makeInvoiceParams := &Nip47MakeInvoiceParams{}
	err = json.Unmarshal(request.Params, makeInvoiceParams)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   requestEvent.NostrId,
			"eventKind": requestEvent.Kind,
			"appId":     app.ID,
		}).Errorf("Failed to decode nostr event: %v", err)
		return nil, err
	}

	if makeInvoiceParams.Description != "" && makeInvoiceParams.DescriptionHash != "" {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   requestEvent.NostrId,
			"eventKind": requestEvent.Kind,
			"appId":     app.ID,
		}).Errorf("Only one of description, description_hash can be provided")

		return &Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    NIP_47_OTHER,
				Message: "Only one of description, description_hash can be provided",
			},
		}, nil
	}

	svc.Logger.WithFields(logrus.Fields{
		"eventId":         requestEvent.NostrId,
		"eventKind":       requestEvent.Kind,
		"appId":           app.ID,
		"amount":          makeInvoiceParams.Amount,
		"description":     makeInvoiceParams.Description,
		"descriptionHash": makeInvoiceParams.DescriptionHash,
		"expiry":          makeInvoiceParams.Expiry,
	}).Info("Making invoice")

	transaction, err := svc.lnClient.MakeInvoice(ctx, makeInvoiceParams.Amount, makeInvoiceParams.Description, makeInvoiceParams.DescriptionHash, makeInvoiceParams.Expiry)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey":    requestEvent.PubKey,
			"eventId":         requestEvent.NostrId,
			"eventKind":       requestEvent.Kind,
			"appId":           app.ID,
			"amount":          makeInvoiceParams.Amount,
			"description":     makeInvoiceParams.Description,
			"descriptionHash": makeInvoiceParams.DescriptionHash,
			"expiry":          makeInvoiceParams.Expiry,
		}).Infof("Failed to make invoice: %v", err)
		return &Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: fmt.Sprintf("Something went wrong while making invoice: %s", err.Error()),
			},
		}, nil
	}

	responsePayload := &Nip47MakeInvoiceResponse{
		Nip47Transaction: *transaction,
	}

	return &Nip47Response{
		ResultType: request.Method,
		Result:     responsePayload,
	}, nil
}
