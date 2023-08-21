package main

import (
	"context"
	"encoding/json"
	"fmt"

	//"fmt"

	"github.com/nbd-wtf/go-nostr"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
)

func (svc *Service) HandleLookupInvoiceEvent(ctx context.Context, request *Nip47Request, event *nostr.Event, app App, ss []byte) (result *nostr.Event, err error) {
	
	// TODO: move to a shared function
	nostrEvent := NostrEvent{App: app, NostrId: event.ID, Content: event.Content, State: "received"}
	insertNostrEventResult := svc.db.Create(&nostrEvent)
	if insertNostrEventResult.Error != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
			"appId":     app.ID,
		}).Errorf("Failed to save nostr event: %v", insertNostrEventResult.Error)
		return nil, insertNostrEventResult.Error
	}

	// TODO: move to a shared function
	hasPermission, code, message := svc.hasPermission(&app, event, request.Method, nil)

	if !hasPermission {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
			"appId":     app.ID,
		}).Errorf("App does not have permission: %s %s", code, message)

		return svc.createResponse(event, Nip47Response{Error: &Nip47Error{
			Code:    code,
			Message: message,
		}}, ss)
	}

	// TODO: move to a shared generic function
	lookupInvoiceParams := &Nip47LookupInvoiceParams{}
	err = json.Unmarshal(request.Params, lookupInvoiceParams)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
			"appId":     app.ID,
		}).Errorf("Failed to decode nostr event: %v", err)
		return nil, err
	}

	svc.Logger.WithFields(logrus.Fields{
		"eventId":         event.ID,
		"eventKind":       event.Kind,
		"appId":           app.ID,
		"invoice":         lookupInvoiceParams.Invoice,
		"paymentHash":     lookupInvoiceParams.PaymentHash,
	}).Info("Looking up invoice")

	paymentHash := lookupInvoiceParams.PaymentHash

	if (paymentHash == "") {
		paymentRequest, err := decodepay.Decodepay(lookupInvoiceParams.Invoice)
		if err != nil {
			svc.Logger.WithFields(logrus.Fields{
				"eventId":   event.ID,
				"eventKind": event.Kind,
				"appId":     app.ID,
				"invoice":    lookupInvoiceParams.Invoice,
			}).Errorf("Failed to decode bolt11 invoice: %v", err)

			return svc.createResponse(event, Nip47Response{
				Error: &Nip47Error{
					Code:    NIP_47_ERROR_INTERNAL,
					Message: fmt.Sprintf("Failed to decode bolt11 invoice: %s", err.Error()),
				},
			}, ss)
		}
		paymentHash = paymentRequest.PaymentHash
	}

	invoice, paid, err := svc.lnClient.LookupInvoice(ctx, event.PubKey, paymentHash)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":         event.ID,
			"eventKind":       event.Kind,
			"appId":           app.ID,
			"invoice":         lookupInvoiceParams.Invoice,
			"paymentHash":     lookupInvoiceParams.PaymentHash,
		}).Infof("Failed to lookup invoice: %v", err)
		nostrEvent.State = NOSTR_EVENT_STATE_HANDLER_ERROR
		svc.db.Save(&nostrEvent)
		return svc.createResponse(event, Nip47Response{
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: fmt.Sprintf("Something went wrong while looking up invoice: %s", err.Error()),
			},
		}, ss)
	}

	responsePayload := &Nip47LookupInvoiceResponse {
		Invoice: invoice,
		Paid: paid,
	}

	nostrEvent.State = NOSTR_EVENT_STATE_HANDLER_EXECUTED
	svc.db.Save(&nostrEvent)
	return svc.createResponse(event, Nip47Response{
		ResultType: NIP_47_LOOKUP_INVOICE_METHOD,
		Result: responsePayload,
	},
	ss)
}
