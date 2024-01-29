package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

func (svc *Service) HandlePayKeysendEvent(ctx context.Context, request *Nip47Request, event *nostr.Event, app App, ss []byte) (result *nostr.Event, err error) {

	nostrEvent := NostrEvent{App: app, NostrId: event.ID, Content: event.Content}
	err = svc.db.Create(&nostrEvent).Error
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
			"appId":     app.ID,
		}).Errorf("Failed to save nostr event: %v", err)
		return nil, err
	}

	payParams := &Nip47KeysendParams{}
	err = json.Unmarshal(request.Params, payParams)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
			"appId":     app.ID,
		}).Errorf("Failed to decode nostr event: %v", err)
		return nil, err
	}

	// We use pay_invoice permissions for budget and max amount
	hasPermission, code, message := svc.hasPermission(&app, event, NIP_47_PAY_INVOICE_METHOD, payParams.Amount)

	if !hasPermission {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":      event.ID,
			"eventKind":    event.Kind,
			"appId":        app.ID,
			"senderPubkey": payParams.Pubkey,
		}).Errorf("App does not have permission: %s %s", code, message)

		return svc.createResponse(event, Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    code,
				Message: message,
			}}, nostr.Tags{}, ss)
	}

	payment := Payment{App: app, NostrEvent: nostrEvent, Amount: uint(payParams.Amount / 1000)}
	insertPaymentResult := svc.db.Create(&payment)
	if insertPaymentResult.Error != nil {
		return nil, insertPaymentResult.Error
	}

	svc.Logger.WithFields(logrus.Fields{
		"eventId":      event.ID,
		"eventKind":    event.Kind,
		"appId":        app.ID,
		"senderPubkey": payParams.Pubkey,
	}).Info("Sending payment")

	preimage, err := svc.lnClient.SendKeysend(ctx, event.PubKey, payParams.Amount/1000, payParams.Pubkey, payParams.Preimage, payParams.TLVRecords)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":      event.ID,
			"eventKind":    event.Kind,
			"appId":        app.ID,
			"senderPubkey": payParams.Pubkey,
		}).Infof("Failed to send payment: %v", err)
		return svc.createResponse(event, Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: fmt.Sprintf("Something went wrong while paying invoice: %s", err.Error()),
			},
		}, nostr.Tags{}, ss)
	}
	payment.Preimage = &preimage
	svc.db.Save(&payment)
	return svc.createResponse(event, Nip47Response{
		ResultType: request.Method,
		Result: Nip47PayResponse{
			Preimage: preimage,
		},
	}, nostr.Tags{}, ss)
}
