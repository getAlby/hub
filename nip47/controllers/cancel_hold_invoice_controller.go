package controllers

import (
	"context"

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
		}).WithError(err).Error("Failed to cancel hold invoice")

		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error:      mapNip47Error(err),
		}, nostr.Tags{})
		return
	}

	publishResponse(&models.Response{
		ResultType: nip47Request.Method,
		Result:     &cancelHoldInvoiceResponse{},
	}, nostr.Tags{})
}
