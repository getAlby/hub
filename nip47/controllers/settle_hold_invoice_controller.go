package controllers

import (
	"context"

	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

type settleHoldInvoiceParams struct {
	Preimage string `json:"preimage"`
}
type settleHoldInvoiceResponse struct{}

func (controller *nip47Controller) HandleSettleHoldInvoiceEvent(ctx context.Context, nip47Request *models.Request, requestEventId uint, appId uint, publishResponse func(*models.Response, nostr.Tags)) {
	settleHoldInvoiceParams := &settleHoldInvoiceParams{}
	decodeErrResp := decodeRequest(nip47Request, settleHoldInvoiceParams)
	if decodeErrResp != nil {
		publishResponse(decodeErrResp, nostr.Tags{})
		return
	}

	logger.Logger.WithFields(logrus.Fields{
		"requestEventId": requestEventId,
		"appId":          appId,
		"preimage":       settleHoldInvoiceParams.Preimage,
	}).Info("Settling hold invoice")

	_, err := controller.transactionsService.SettleHoldInvoice(ctx, settleHoldInvoiceParams.Preimage, controller.lnClient)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"request_event_id": requestEventId,
			"appId":            appId,
			"preimage":         settleHoldInvoiceParams.Preimage,
		}).WithError(err).Error("Failed to settle hold invoice")

		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error:      mapNip47Error(err),
		}, nostr.Tags{})
		return
	}

	publishResponse(&models.Response{
		ResultType: nip47Request.Method,
		Result:     &settleHoldInvoiceResponse{},
	}, nostr.Tags{})
}
