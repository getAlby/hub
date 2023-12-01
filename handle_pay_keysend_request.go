package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

func (svc *Service) HandlePayKeysendEvent(ctx context.Context, request *Nip47Request, event *nostr.Event, app App, ss []byte) (result *nostr.Event, err error) {

	nostrEvent := NostrEvent{App: app, NostrId: event.ID, Content: event.Content, State: "received"}
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

	hasPermission, code, message := svc.hasPermission(&app, event, request.Method, payParams.Amount)

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
			}}, ss)
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

	preimage, paymentHash, err := svc.lnClient.SendKeysend(ctx, event.PubKey, payParams.Amount, payParams.Pubkey, payParams.Message, payParams.Preimage, payParams.TLVRecords)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":      event.ID,
			"eventKind":    event.Kind,
			"appId":        app.ID,
			"senderPubkey": payParams.Pubkey,
		}).Infof("Failed to send payment: %v", err)
		nostrEvent.State = NOSTR_EVENT_STATE_HANDLER_ERROR
		svc.db.Save(&nostrEvent)
		return svc.createResponse(event, Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: fmt.Sprintf("Something went wrong while paying invoice: %s", err.Error()),
			},
		}, ss)
	}
	payment.Preimage = &preimage
	nostrEvent.State = NOSTR_EVENT_STATE_HANDLER_EXECUTED
	svc.db.Save(&nostrEvent)
	svc.db.Save(&payment)
	return svc.createResponse(event, Nip47Response{
		ResultType: request.Method,
		Result: Nip47KeysendResponse{
			Preimage:    preimage,
			PaymentHash: paymentHash,
		},
	}, ss)
}
