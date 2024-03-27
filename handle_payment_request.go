package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/getAlby/nostr-wallet-connect/events"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
)

func (svc *Service) HandlePayInvoiceEvent(ctx context.Context, request *Nip47Request, requestEvent *RequestEvent, app *App) (result *Nip47Response, err error) {

	payParams := &Nip47PayParams{}
	result = svc.unmarshalRequest(request, requestEvent, app, payParams)
	if result != nil {
		return result, nil
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

		return &Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: fmt.Sprintf("Failed to decode bolt11 invoice: %s", err.Error()),
			},
		}, nil
	}

	resp := svc.checkPermission(request, requestEvent, app, paymentRequest.MSatoshi)
	if resp != nil {
		return resp, nil
	}

	payment := Payment{App: *app, RequestEvent: *requestEvent, PaymentRequest: bolt11, Amount: uint(paymentRequest.MSatoshi / 1000)}
	err = svc.db.Create(&payment).Error
	if err != nil {
		return &Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: err.Error(),
			}}, nil
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
		return &Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: err.Error(),
			},
		}, nil
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

	return &Nip47Response{
		ResultType: request.Method,
		Result: Nip47PayResponse{
			Preimage: preimage,
		},
	}, nil
}
