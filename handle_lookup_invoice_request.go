package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
)

func (svc *Service) HandleLookupInvoiceEvent(ctx context.Context, request *Nip47Request, requestEvent *NostrEvent, app *App) (result *Nip47Response, err error) {

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
	lookupInvoiceParams := &Nip47LookupInvoiceParams{}
	err = json.Unmarshal(request.Params, lookupInvoiceParams)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   requestEvent.NostrId,
			"eventKind": requestEvent.Kind,
			"appId":     app.ID,
		}).Errorf("Failed to decode nostr event: %v", err)
		return nil, err
	}

	svc.Logger.WithFields(logrus.Fields{
		"eventId":     requestEvent.NostrId,
		"eventKind":   requestEvent.Kind,
		"appId":       app.ID,
		"invoice":     lookupInvoiceParams.Invoice,
		"paymentHash": lookupInvoiceParams.PaymentHash,
	}).Info("Looking up invoice")

	paymentHash := lookupInvoiceParams.PaymentHash

	if paymentHash == "" {
		paymentRequest, err := decodepay.Decodepay(strings.ToLower(lookupInvoiceParams.Invoice))
		if err != nil {
			svc.Logger.WithFields(logrus.Fields{
				"eventId":   requestEvent.NostrId,
				"eventKind": requestEvent.Kind,
				"appId":     app.ID,
				"invoice":   lookupInvoiceParams.Invoice,
			}).Errorf("Failed to decode bolt11 invoice: %v", err)

			return &Nip47Response{
				ResultType: request.Method,
				Error: &Nip47Error{
					Code:    NIP_47_ERROR_INTERNAL,
					Message: fmt.Sprintf("Failed to decode bolt11 invoice: %s", err.Error()),
				},
			}, nil
		}
		paymentHash = paymentRequest.PaymentHash
	}

	transaction, err := svc.lnClient.LookupInvoice(ctx, paymentHash)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey": requestEvent.PubKey,
			"eventId":      requestEvent.NostrId,
			"eventKind":    requestEvent.Kind,
			"appId":        app.ID,
			"invoice":      lookupInvoiceParams.Invoice,
			"paymentHash":  lookupInvoiceParams.PaymentHash,
		}).Infof("Failed to lookup invoice: %v", err)
		return &Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: fmt.Sprintf("Something went wrong while looking up invoice: %s", err.Error()),
			},
		}, nil
	}

	responsePayload := &Nip47LookupInvoiceResponse{
		Nip47Transaction: *transaction,
	}

	return &Nip47Response{
		ResultType: request.Method,
		Result:     responsePayload,
	}, nil
}
