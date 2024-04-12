package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/getAlby/nostr-wallet-connect/events"
	"github.com/getAlby/nostr-wallet-connect/nip47"
	"github.com/nbd-wtf/go-nostr"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
)

func (svc *Service) HandlePayInvoiceEvent(ctx context.Context, nip47Request *Nip47Request, requestEvent *RequestEvent, app *App, publishResponse func(*Nip47Response, nostr.Tags)) {

	payParams := &Nip47PayParams{}
	resp := svc.decodeNip47Request(nip47Request, requestEvent, app, payParams)
	if resp != nil {
		publishResponse(resp, nostr.Tags{})
		return
	}

	bolt11 := payParams.Invoice
	// Convert invoice to lowercase string
	bolt11 = strings.ToLower(bolt11)
	paymentRequest, err := decodepay.Decodepay(bolt11)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"requestEventNostrId": requestEvent.NostrId,
			"appId":               app.ID,
			"bolt11":              bolt11,
		}).Errorf("Failed to decode bolt11 invoice: %v", err)

		publishResponse(&Nip47Response{
			ResultType: nip47Request.Method,
			Error: &Nip47Error{
				Code:    nip47.ERROR_INTERNAL,
				Message: fmt.Sprintf("Failed to decode bolt11 invoice: %s", err.Error()),
			},
		}, nostr.Tags{})
		return
	}

	resp = svc.checkPermission(nip47Request, requestEvent.NostrId, app, paymentRequest.MSatoshi)
	if resp != nil {
		publishResponse(resp, nostr.Tags{})
		return
	}

	payment := Payment{App: *app, RequestEvent: *requestEvent, PaymentRequest: bolt11, Amount: uint(paymentRequest.MSatoshi / 1000)}
	err = svc.db.Create(&payment).Error
	if err != nil {
		publishResponse(&Nip47Response{
			ResultType: nip47Request.Method,
			Error: &Nip47Error{
				Code:    nip47.ERROR_INTERNAL,
				Message: err.Error(),
			},
		}, nostr.Tags{})
		return
	}

	svc.Logger.WithFields(logrus.Fields{
		"requestEventNostrId": requestEvent.NostrId,
		"appId":               app.ID,
		"bolt11":              bolt11,
	}).Info("Sending payment")

	preimage, err := svc.lnClient.SendPaymentSync(ctx, bolt11)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"requestEventNostrId": requestEvent.NostrId,
			"appId":               app.ID,
			"bolt11":              bolt11,
		}).Infof("Failed to send payment: %v", err)
		svc.EventPublisher.Publish(&events.Event{
			Event: "nwc_payment_failed",
			Properties: map[string]interface{}{
				// "error":   fmt.Sprintf("%v", err),
				"invoice": bolt11,
				"amount":  paymentRequest.MSatoshi / 1000,
			},
		})
		publishResponse(&Nip47Response{
			ResultType: nip47Request.Method,
			Error: &Nip47Error{
				Code:    nip47.ERROR_INTERNAL,
				Message: err.Error(),
			},
		}, nostr.Tags{})
		return
	}
	payment.Preimage = &preimage
	svc.db.Save(&payment)

	svc.EventPublisher.Publish(&events.Event{
		Event: "nwc_payment_succeeded",
		Properties: map[string]interface{}{
			"bolt11": bolt11,
			"amount": paymentRequest.MSatoshi / 1000,
		},
	})

	publishResponse(&Nip47Response{
		ResultType: nip47Request.Method,
		Result: Nip47PayResponse{
			Preimage: preimage,
		},
	}, nostr.Tags{})
}
