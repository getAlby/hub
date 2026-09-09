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

// HandleReceiveEvent handles the NWC-321 receive method. Currently only
// BOLT-11 instructions (the "lightning" URI parameter) are returned.
func (controller *nip47Controller) HandleReceiveEvent(ctx context.Context, nip47Request *models.Request, requestEventId uint, appId uint, publishResponse publishFunc) {
	receiveParams := &receiveParams{}
	resp := decodeRequest(nip47Request, receiveParams)
	if resp != nil {
		publishResponse(resp, nostr.Tags{})
		return
	}

	if receiveParams.Amount == nil || *receiveParams.Amount == 0 {
		// variable-amount (zero-amount) invoices are not supported
		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error: &models.Error{
				Code:    constants.ERROR_BAD_REQUEST,
				Message: "amount is required and must be greater than 0",
			},
		}, nostr.Tags{})
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
