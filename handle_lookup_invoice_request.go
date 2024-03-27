package main

import (
	"context"
	"fmt"
	"strings"

	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
)

func (svc *Service) HandleLookupInvoiceEvent(ctx context.Context, request *Nip47Request, requestEvent *RequestEvent, app *App) *Nip47Response {

	lookupInvoiceParams := &Nip47LookupInvoiceParams{}
	resp := svc.unmarshalRequest(request, requestEvent, app, lookupInvoiceParams)
	if resp != nil {
		return resp
	}

	resp = svc.checkPermission(request, requestEvent, app, 0)
	if resp != nil {
		return resp
	}

	svc.Logger.WithFields(logrus.Fields{
		"eventId":     requestEvent.NostrId,
		"appId":       app.ID,
		"invoice":     lookupInvoiceParams.Invoice,
		"paymentHash": lookupInvoiceParams.PaymentHash,
	}).Info("Looking up invoice")

	paymentHash := lookupInvoiceParams.PaymentHash

	if paymentHash == "" {
		paymentRequest, err := decodepay.Decodepay(strings.ToLower(lookupInvoiceParams.Invoice))
		if err != nil {
			svc.Logger.WithFields(logrus.Fields{
				"eventId": requestEvent.NostrId,
				"appId":   app.ID,
				"invoice": lookupInvoiceParams.Invoice,
			}).Errorf("Failed to decode bolt11 invoice: %v", err)

			return &Nip47Response{
				ResultType: request.Method,
				Error: &Nip47Error{
					Code:    NIP_47_ERROR_INTERNAL,
					Message: fmt.Sprintf("Failed to decode bolt11 invoice: %s", err.Error()),
				},
			}
		}
		paymentHash = paymentRequest.PaymentHash
	}

	transaction, err := svc.lnClient.LookupInvoice(ctx, paymentHash)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":     requestEvent.NostrId,
			"appId":       app.ID,
			"invoice":     lookupInvoiceParams.Invoice,
			"paymentHash": lookupInvoiceParams.PaymentHash,
		}).Infof("Failed to lookup invoice: %v", err)
		return &Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: err.Error(),
			},
		}
	}

	responsePayload := &Nip47LookupInvoiceResponse{
		Nip47Transaction: *transaction,
	}

	return &Nip47Response{
		ResultType: request.Method,
		Result:     responsePayload,
	}
}
