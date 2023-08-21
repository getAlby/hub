package main

import (
	"context"
	"encoding/json"
	"fmt"

	//"fmt"

	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

func (svc *Service) HandleMakeInvoiceEvent(ctx context.Context, request *Nip47Request, event *nostr.Event, app App, ss []byte) (result *nostr.Event, err error) {
	
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
	makeInvoiceParams := &Nip47MakeInvoiceParams{}
	err = json.Unmarshal(request.Params, makeInvoiceParams)
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
		"amount":          makeInvoiceParams.Amount,
		"description":     makeInvoiceParams.Description,
		"descriptionHash": makeInvoiceParams.DescriptionHash,
		"expiry":          makeInvoiceParams.Expiry,
	}).Info("Making invoice")

	invoice, paymentHash, err := svc.lnClient.MakeInvoice(ctx, event.PubKey, makeInvoiceParams.Amount, makeInvoiceParams.Description, makeInvoiceParams.DescriptionHash, makeInvoiceParams.Expiry)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":         event.ID,
			"eventKind":       event.Kind,
			"appId":           app.ID,
			"amount":          makeInvoiceParams.Amount,
			"description":     makeInvoiceParams.Description,
			"descriptionHash": makeInvoiceParams.DescriptionHash,
			"expiry":          makeInvoiceParams.Expiry,
		}).Infof("Failed to make invoice: %v", err)
		nostrEvent.State = NOSTR_EVENT_STATE_HANDLER_ERROR
		svc.db.Save(&nostrEvent)
		return svc.createResponse(event, Nip47Response{
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: fmt.Sprintf("Something went wrong while making invoice: %s", err.Error()),
			},
		}, ss)
	}

	responsePayload := &Nip47MakeInvoiceResponse{
		Invoice: invoice,
		PaymentHash: paymentHash,
	}

	nostrEvent.State = NOSTR_EVENT_STATE_HANDLER_EXECUTED
	svc.db.Save(&nostrEvent)
	return svc.createResponse(event, Nip47Response{
		ResultType: NIP_47_MAKE_INVOICE_METHOD,
		Result: responsePayload,
	},
	ss)
}
