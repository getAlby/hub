package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
)

func (svc *Service) HandlePayInvoiceEvent(ctx context.Context, request *Nip47Request, requestEvent *NostrEvent, app *App) (result *Nip47Response, err error) {

	payParams := &Nip47PayParams{}
	err = json.Unmarshal(request.Params, payParams)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   requestEvent.NostrId,
			"eventKind": requestEvent.Kind,
			"appId":     app.ID,
		}).Errorf("Failed to decode nostr event: %v", err)
		// TODO: why not return a Nip47Response here?
		return nil, err
	}

	bolt11 := payParams.Invoice
	// Convert invoice to lowercase string
	bolt11 = strings.ToLower(bolt11)
	paymentRequest, err := decodepay.Decodepay(bolt11)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   requestEvent.NostrId,
			"eventKind": requestEvent.Kind,
			"appId":     app.ID,
			"bolt11":    bolt11,
		}).Errorf("Failed to decode bolt11 invoice: %v", err)

		return &Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: fmt.Sprintf("Failed to decode bolt11 invoice: %s", err.Error()),
			},
		}, nil
	}

	hasPermission, code, message := svc.hasPermission(app, request.Method, paymentRequest.MSatoshi)

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

	payment := Payment{App: *app, NostrEvent: *requestEvent, PaymentRequest: bolt11, Amount: uint(paymentRequest.MSatoshi / 1000)}
	insertPaymentResult := svc.db.Create(&payment)
	if insertPaymentResult.Error != nil {
		return nil, insertPaymentResult.Error
	}

	svc.Logger.WithFields(logrus.Fields{
		"eventId":   requestEvent.NostrId,
		"eventKind": requestEvent.Kind,
		"appId":     app.ID,
		"bolt11":    bolt11,
	}).Info("Sending payment")

	preimage, err := svc.lnClient.SendPaymentSync(ctx, bolt11)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey": requestEvent.PubKey,
			"eventId":      requestEvent.NostrId,
			"eventKind":    requestEvent.Kind,
			"appId":        app.ID,
			"bolt11":       bolt11,
		}).Infof("Failed to send payment: %v", err)
		return &Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: fmt.Sprintf("Something went wrong while paying invoice: %s", err.Error()),
			},
		}, nil
	}
	payment.Preimage = &preimage
	svc.db.Save(&payment)
	return &Nip47Response{
		ResultType: request.Method,
		Result: Nip47PayResponse{
			Preimage: preimage,
		},
	}, nil
}
