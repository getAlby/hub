package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/getAlby/nostr-wallet-connect/events"
	"github.com/nbd-wtf/go-nostr"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
)

func (svc *Service) HandlePayInvoiceEvent(ctx context.Context, request *Nip47Request, requestEvent *RequestEvent, app *App, publishResponse func(*Nip47Response, *nostr.Tags)) {

	payParams := &Nip47PayParams{}
	resp := svc.unmarshalRequest(request, requestEvent, app, payParams)
	if resp != nil {
		publishResponse(resp, &nostr.Tags{})
		return
	}

	bolt11 := payParams.Invoice
	// Convert invoice to lowercase string
	bolt11 = strings.ToLower(bolt11)
	paymentRequest, err := decodepay.Decodepay(bolt11)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId": requestEvent.NostrId,
			"appId":   app.ID,
			"bolt11":  bolt11,
		}).Errorf("Failed to decode bolt11 invoice: %v", err)

		publishResponse(&Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: fmt.Sprintf("Failed to decode bolt11 invoice: %s", err.Error()),
			},
		}, &nostr.Tags{})
		return
	}

	resp = svc.checkPermission(request, requestEvent, app, paymentRequest.MSatoshi)
	if resp != nil {
		publishResponse(resp, &nostr.Tags{})
		return
	}

	payment := Payment{App: *app, RequestEvent: *requestEvent, PaymentRequest: bolt11, Amount: uint(paymentRequest.MSatoshi / 1000)}
	err = svc.db.Create(&payment).Error
	if err != nil {
		publishResponse(&Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: err.Error(),
			},
		}, &nostr.Tags{})
		return
	}

	svc.Logger.WithFields(logrus.Fields{
		"eventId": requestEvent.NostrId,
		"appId":   app.ID,
		"bolt11":  bolt11,
	}).Info("Sending payment")

	preimage, err := svc.lnClient.SendPaymentSync(ctx, bolt11)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId": requestEvent.NostrId,
			"appId":   app.ID,
			"bolt11":  bolt11,
		}).Infof("Failed to send payment: %v", err)
		svc.EventLogger.Log(ctx, &events.Event{
			Event: "nwc_payment_failed",
			Properties: map[string]interface{}{
				"invoice": bolt11,
				"error":   fmt.Sprintf("%v", err),
				"amount":  paymentRequest.MSatoshi / 1000,
			},
		})
		publishResponse(&Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: err.Error(),
			},
		}, &nostr.Tags{})
		return
	}
	payment.Preimage = &preimage
	svc.db.Save(&payment)

	svc.EventLogger.Log(ctx, &events.Event{
		Event: "nwc_payment_succeeded",
		Properties: map[string]interface{}{
			"bolt11": bolt11,
			"amount": paymentRequest.MSatoshi / 1000,
		},
	})

	publishResponse(&Nip47Response{
		ResultType: request.Method,
		Result: Nip47PayResponse{
			Preimage: preimage,
		},
	}, &nostr.Tags{})
}
