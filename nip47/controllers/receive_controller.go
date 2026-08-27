package controllers

import (
	"context"

	"github.com/getAlby/go-nostr"
	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	"github.com/sirupsen/logrus"
)

type receiveParams struct {
	Amount      *uint64                `json:"amount"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type receiveResult struct {
	Bip321        string `json:"bip321"`
	TransactionId string `json:"transaction_id,omitempty"`
}

// HandleReceiveEvent handles the NWC-321 receive method. Requests with an
// amount return a BOLT-11 invoice; requests without an amount return a
// BOLT-12 variable-amount offer if the LN backend supports BOLT-12.
func (controller *nip47Controller) HandleReceiveEvent(ctx context.Context, nip47Request *models.Request, requestEventId uint, appId uint, publishResponse publishFunc) {
	receiveParams := &receiveParams{}
	resp := decodeRequest(nip47Request, receiveParams)
	if resp != nil {
		publishResponse(resp, nostr.Tags{})
		return
	}

	if receiveParams.Amount == nil {
		controller.receiveVariableAmount(ctx, nip47Request, requestEventId, appId, receiveParams, publishResponse)
		return
	}

	logger.Logger.WithFields(logrus.Fields{
		"request_event_id": requestEventId,
		"app_id":           appId,
		"amount":           *receiveParams.Amount,
		"description":      receiveParams.Description,
	}).Debug("Handling receive request")

	transaction, err := controller.transactionsService.MakeInvoice(ctx, *receiveParams.Amount, receiveParams.Description, "", 0, receiveParams.Metadata, controller.lnClient, &appId, &requestEventId, nil)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"request_event_id": requestEventId,
			"app_id":           appId,
			"amount":           *receiveParams.Amount,
			"description":      receiveParams.Description,
		}).Infof("Failed to make invoice: %v", err)

		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error:      mapNip47Error(err),
		}, nostr.Tags{})
		return
	}

	publishResponse(&models.Response{
		ResultType: nip47Request.Method,
		Result: receiveResult{
			Bip321:        "bitcoin:?lightning=" + transaction.PaymentRequest,
			TransactionId: transaction.PaymentHash,
		},
	}, nostr.Tags{})
}

// receiveVariableAmount handles a receive request without an amount by
// returning a BOLT-12 variable-amount offer, which lets the payer choose the
// amount. This requires an LN backend with BOLT-12 support.
func (controller *nip47Controller) receiveVariableAmount(ctx context.Context, nip47Request *models.Request, requestEventId uint, appId uint, receiveParams *receiveParams, publishResponse publishFunc) {
	nodeInfo, err := controller.lnClient.GetInfo(ctx)
	if err != nil {
		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error: &models.Error{
				Code:    constants.ERROR_INTERNAL,
				Message: "Failed to get node info: " + err.Error(),
			},
		}, nostr.Tags{})
		return
	}

	if !nodeInfo.SupportsBolt12 {
		// variable-amount (zero-amount) invoices are not supported
		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error: &models.Error{
				Code:    constants.ERROR_BAD_REQUEST,
				Message: "amount is required",
			},
		}, nostr.Tags{})
		return
	}

	offer, err := controller.lnClient.MakeOffer(ctx, receiveParams.Description)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"request_event_id": requestEventId,
			"app_id":           appId,
			"description":      receiveParams.Description,
		}).WithError(err).Error("Failed to create BOLT-12 offer")

		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error: &models.Error{
				Code:    constants.ERROR_INTERNAL,
				Message: "Failed to create offer: " + err.Error(),
			},
		}, nostr.Tags{})
		return
	}

	publishResponse(&models.Response{
		ResultType: nip47Request.Method,
		Result: receiveResult{
			Bip321: "bitcoin:?lno=" + offer,
		},
	}, nostr.Tags{})
}
