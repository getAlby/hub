package controllers

import (
	"context"

	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

type makeInvoiceParams struct {
	Amount          int64  `json:"amount"`
	Description     string `json:"description"`
	DescriptionHash string `json:"description_hash"`
	Expiry          int64  `json:"expiry"`
}
type makeInvoiceResponse struct {
	models.Transaction
}

type makeInvoiceController struct {
	lnClient lnclient.LNClient
}

func NewMakeInvoiceController(lnClient lnclient.LNClient) *makeInvoiceController {
	return &makeInvoiceController{
		lnClient: lnClient,
	}
}

func (controller *makeInvoiceController) HandleMakeInvoiceEvent(ctx context.Context, nip47Request *models.Request, requestEventId uint, checkPermission checkPermissionFunc, publishResponse publishFunc) {
	// basic permissions check
	resp := checkPermission(0)
	if resp != nil {
		publishResponse(resp, nostr.Tags{})
		return
	}

	makeInvoiceParams := &makeInvoiceParams{}
	resp = decodeRequest(nip47Request, makeInvoiceParams)
	if resp != nil {
		publishResponse(resp, nostr.Tags{})
		return
	}

	logger.Logger.WithFields(logrus.Fields{
		"request_event_id": requestEventId,
		"amount":           makeInvoiceParams.Amount,
		"description":      makeInvoiceParams.Description,
		"description_hash": makeInvoiceParams.DescriptionHash,
		"expiry":           makeInvoiceParams.Expiry,
	}).Info("Making invoice")

	expiry := makeInvoiceParams.Expiry

	transaction, err := controller.lnClient.MakeInvoice(ctx, makeInvoiceParams.Amount, makeInvoiceParams.Description, makeInvoiceParams.DescriptionHash, expiry)
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
			Error: &models.Error{
				Code:    models.ERROR_INTERNAL,
				Message: err.Error(),
			},
		}, nostr.Tags{})
		return
	}

	responsePayload := &makeInvoiceResponse{
		Transaction: *transaction,
	}

	publishResponse(&models.Response{
		ResultType: nip47Request.Method,
		Result:     responsePayload,
	}, nostr.Tags{})
}
