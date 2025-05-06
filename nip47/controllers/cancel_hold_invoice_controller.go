package controllers

import (
	"context"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

type cancelHoldInvoiceParams struct {
	PaymentHash string `json:"payment_hash"`
}
type cancelHoldInvoiceResponse struct{}

func (controller *nip47Controller) HandleCancelHoldInvoiceEvent(ctx context.Context, nip47Request *models.Request, requestEventId uint, appId uint, publishResponse func(*models.Response, nostr.Tags)) {
	if nip47Request == nil {
		logger.Logger.WithFields(logrus.Fields{
			"requestEventId": requestEventId,
			"appId":          appId,
		}).Error("Received nil nip47Request in HandleCancelHoldInvoiceEvent")
		publishResponse(&models.Response{
			ResultType: models.CANCEL_HOLD_INVOICE_METHOD,
			Error: &models.Error{
				Code:    constants.ERROR_INTERNAL,
				Message: "Internal server error: received nil request payload",
			},
		}, nostr.Tags{})
		return
	}

	resp := &models.Response{}
	resp.ResultType = models.CANCEL_HOLD_INVOICE_METHOD

	cancelHoldInvoiceParams := &cancelHoldInvoiceParams{}
	decodeErrResp := decodeRequest(nip47Request, cancelHoldInvoiceParams)
	if decodeErrResp != nil {
		publishResponse(decodeErrResp, nostr.Tags{})
		return
	}

	logger.Logger.WithFields(logrus.Fields{
		"requestEventId": requestEventId,
		"appId":          appId,
		"paymentHash":    cancelHoldInvoiceParams.PaymentHash,
	}).Info("Canceling hold invoice")

	err := controller.transactionsService.CancelHoldInvoice(ctx, cancelHoldInvoiceParams.PaymentHash, controller.lnClient)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"request_event_id": requestEventId,
			"appId":            appId,
			"paymentHash":      cancelHoldInvoiceParams.PaymentHash,
		}).Infof("Failed to cancel hold invoice: %v", err)

		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error: &models.Error{
				Code:    constants.ERROR_INTERNAL,
				Message: err.Error(),
			},
		}, nostr.Tags{})
		return
	}

	publishResponse(&models.Response{
		ResultType: nip47Request.Method,
		Result:     &cancelHoldInvoiceResponse{},
	}, nostr.Tags{})
}
