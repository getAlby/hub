package controllers

import (
	"context"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

type makeHoldInvoiceParams struct {
	Amount          uint64                 `json:"amount"`
	PaymentHash     string                 `json:"payment_hash"`
	Description     string                 `json:"description"`
	DescriptionHash string                 `json:"description_hash"`
	Expiry          uint64                 `json:"expiry"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}
type makeHoldInvoiceResponse struct {
	models.Transaction
}

func (controller *nip47Controller) HandleMakeHoldInvoiceEvent(ctx context.Context, nip47Request *models.Request, requestEventId uint, appId uint, publishResponse func(*models.Response, nostr.Tags)) {
	resp := &models.Response{}
	resp.ResultType = models.MAKE_HOLD_INVOICE_METHOD

	makeHoldInvoiceParams := &makeHoldInvoiceParams{}
	decodeErrResp := decodeRequest(nip47Request, makeHoldInvoiceParams)
	if decodeErrResp != nil {
		publishResponse(decodeErrResp, nostr.Tags{})
		return
	}

	logger.Logger.WithFields(logrus.Fields{
		"requestEventId":  requestEventId,
		"appId":           appId,
		"amount":          makeHoldInvoiceParams.Amount,
		"description":     makeHoldInvoiceParams.Description,
		"descriptionHash": makeHoldInvoiceParams.DescriptionHash,
		"expiry":          makeHoldInvoiceParams.Expiry,
		"paymentHash":     makeHoldInvoiceParams.PaymentHash,
		"metadata":        makeHoldInvoiceParams.Metadata,
	}).Info("Making hold invoice")


	requestEventIdUint := uint(requestEventId)
	transaction, err := controller.transactionsService.MakeHoldInvoice(
		ctx,
		makeHoldInvoiceParams.Amount,
		makeHoldInvoiceParams.Description,
		makeHoldInvoiceParams.DescriptionHash,
		makeHoldInvoiceParams.Expiry,
		makeHoldInvoiceParams.PaymentHash,
		makeHoldInvoiceParams.Metadata,
		controller.lnClient,
		&appId,
		&requestEventIdUint,
	)

	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"request_event_id": requestEventId,
			"appId":            appId,
			"amount":           makeHoldInvoiceParams.Amount,
			"description":      makeHoldInvoiceParams.Description,
			"descriptionHash":  makeHoldInvoiceParams.DescriptionHash,
			"expiry":           makeHoldInvoiceParams.Expiry,
			"paymentHash":      makeHoldInvoiceParams.PaymentHash,
		}).Infof("Failed to make invoice: %v", err)

		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error: &models.Error{
				Code:    constants.ERROR_INTERNAL,
				Message: err.Error(),
			},
		}, nostr.Tags{})
		return
	}

	nip47Transaction := models.ToNip47Transaction(transaction)

	responsePayload := &makeHoldInvoiceResponse{
		Transaction: *nip47Transaction,
	}

	resp.Result = responsePayload

	publishResponse(&models.Response{
		ResultType: nip47Request.Method,
		Result:     responsePayload,
	}, nostr.Tags{})
}
