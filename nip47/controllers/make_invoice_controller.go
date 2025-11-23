package controllers

import (
	"context"

	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

type makeInvoiceParams struct {
	Amount          uint64                 `json:"amount"`
	Description     string                 `json:"description"`
	DescriptionHash string                 `json:"description_hash"`
	Expiry          uint64                 `json:"expiry"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}
type makeInvoiceResponse struct {
	models.Transaction
}

func (controller *nip47Controller) HandleMakeInvoiceEvent(ctx context.Context, nip47Request *models.Request, requestEventId uint, appId uint, publishResponse publishFunc) {

	makeInvoiceParams := &makeInvoiceParams{}
	resp := decodeRequest(nip47Request, makeInvoiceParams)
	if resp != nil {
		publishResponse(resp, nostr.Tags{})
		return
	}

	logger.Logger.WithFields(logrus.Fields{
		"app_id":           appId,
		"request_event_id": requestEventId,
		"amount":           makeInvoiceParams.Amount,
		"description":      makeInvoiceParams.Description,
		"description_hash": makeInvoiceParams.DescriptionHash,
		"expiry":           makeInvoiceParams.Expiry,
		"metadata":         makeInvoiceParams.Metadata,
	}).Debug("Handling make_invoice request")

	expiry := makeInvoiceParams.Expiry

	transaction, err := controller.transactionsService.MakeInvoice(ctx, makeInvoiceParams.Amount, makeInvoiceParams.Description, makeInvoiceParams.DescriptionHash, expiry, makeInvoiceParams.Metadata, controller.lnClient, &appId, &requestEventId, nil)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"request_event_id": requestEventId,
			"amount":           makeInvoiceParams.Amount,
			"description":      makeInvoiceParams.Description,
			"descriptionHash":  makeInvoiceParams.DescriptionHash,
			"expiry":           makeInvoiceParams.Expiry,
		}).Infof("Failed to make invoice: %v", err)

		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error:      mapNip47Error(err),
		}, nostr.Tags{})
		return
	}

	nip47Transaction := models.ToNip47Transaction(transaction)
	responsePayload := &makeInvoiceResponse{
		Transaction: *nip47Transaction,
	}

	publishResponse(&models.Response{
		ResultType: nip47Request.Method,
		Result:     responsePayload,
	}, nostr.Tags{})
}
