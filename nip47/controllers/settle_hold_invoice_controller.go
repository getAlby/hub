package controllers

import (
	"context"

	"github.com/getAlby/hub/constants"
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
	if nip47Request == nil {
		logger.Logger.WithFields(logrus.Fields{
			"requestEventId": requestEventId,
			"appId":          appId,
		}).Error("Received nil nip47Request in HandleSettleHoldInvoiceEvent")
		publishResponse(&models.Response{
			ResultType: models.SETTLE_HOLD_INVOICE_METHOD,
			Error: &models.Error{
				Code:    constants.ERROR_INTERNAL,
				Message: "Internal server error: received nil request payload",
			},
		}, nostr.Tags{})
		return
	}

	resp := &models.Response{}
	resp.ResultType = models.SETTLE_HOLD_INVOICE_METHOD

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
		}).Infof("Failed to settle hold invoice: %v", err)

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
		Result:     &settleHoldInvoiceResponse{},
	}, nostr.Tags{})
}
