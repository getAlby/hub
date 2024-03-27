package main

import (
	"context"

	"github.com/sirupsen/logrus"
)

func (svc *Service) HandleMakeInvoiceEvent(ctx context.Context, request *Nip47Request, requestEvent *RequestEvent, app *App) (result *Nip47Response, err error) {

	makeInvoiceParams := &Nip47MakeInvoiceParams{}
	result = svc.unmarshalRequest(request, requestEvent, app, makeInvoiceParams)
	if result != nil {
		return result, nil
	}

	resp := svc.checkPermission(request, requestEvent, app, 0)
	if resp != nil {
		return resp, nil
	}

	if makeInvoiceParams.Description != "" && makeInvoiceParams.DescriptionHash != "" {
		svc.Logger.WithFields(logrus.Fields{
			"eventId": requestEvent.NostrId,
			"appId":   app.ID,
		}).Errorf("Only one of description, description_hash can be provided")

		return &Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    NIP_47_OTHER,
				Message: "Only one of description, description_hash can be provided",
			},
		}, nil
	}

	svc.Logger.WithFields(logrus.Fields{
		"eventId":         requestEvent.NostrId,
		"appId":           app.ID,
		"amount":          makeInvoiceParams.Amount,
		"description":     makeInvoiceParams.Description,
		"descriptionHash": makeInvoiceParams.DescriptionHash,
		"expiry":          makeInvoiceParams.Expiry,
	}).Info("Making invoice")

	expiry := makeInvoiceParams.Expiry
	if expiry == 0 {
		expiry = 86400
	}

	transaction, err := svc.lnClient.MakeInvoice(ctx, makeInvoiceParams.Amount, makeInvoiceParams.Description, makeInvoiceParams.DescriptionHash, expiry)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":         requestEvent.NostrId,
			"appId":           app.ID,
			"amount":          makeInvoiceParams.Amount,
			"description":     makeInvoiceParams.Description,
			"descriptionHash": makeInvoiceParams.DescriptionHash,
			"expiry":          makeInvoiceParams.Expiry,
		}).Infof("Failed to make invoice: %v", err)
		return &Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: err.Error(),
			},
		}, nil
	}

	responsePayload := &Nip47MakeInvoiceResponse{
		Nip47Transaction: *transaction,
	}

	return &Nip47Response{
		ResultType: request.Method,
		Result:     responsePayload,
	}, nil
}
