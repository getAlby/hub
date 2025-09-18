package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	"github.com/nbd-wtf/go-nostr"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
)

type payInvoiceParams struct {
	Invoice  string                 `json:"invoice"`
	Amount   *uint64                `json:"amount"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (controller *nip47Controller) HandlePayInvoiceEvent(ctx context.Context, nip47Request *models.Request, requestEventId uint, app *db.App, publishResponse publishFunc, tags nostr.Tags) {
	payParams := &payInvoiceParams{}
	resp := decodeRequest(nip47Request, payParams)
	if resp != nil {
		publishResponse(resp, tags)
		return
	}

	bolt11 := payParams.Invoice
	// Convert invoice to lowercase string
	bolt11 = strings.ToLower(bolt11)
	paymentRequest, err := decodepay.Decodepay(bolt11)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"request_event_id": requestEventId,
			"app_id":           app.ID,
			"bolt11":           bolt11,
		}).WithError(err).Error("Failed to decode bolt11 invoice")

		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error: &models.Error{
				Code:    constants.ERROR_BAD_REQUEST,
				Message: fmt.Sprintf("Failed to decode bolt11 invoice: %s", err.Error()),
			},
		}, tags)
		return
	}

	controller.pay(bolt11, payParams.Amount, payParams.Metadata, &paymentRequest, nip47Request, requestEventId, app, publishResponse, tags)
}

func (controller *nip47Controller) pay(bolt11 string, amount *uint64, metadata map[string]interface{}, paymentRequest *decodepay.Bolt11, nip47Request *models.Request, requestEventId uint, app *db.App, publishResponse publishFunc, tags nostr.Tags) {
	logger.Logger.WithFields(logrus.Fields{
		"request_event_id": requestEventId,
		"app_id":           app.ID,
		"bolt11":           bolt11,
	}).Info("Sending payment")

	transaction, err := controller.transactionsService.SendPaymentSync(bolt11, amount, metadata, controller.lnClient, &app.ID, &requestEventId)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"request_event_id": requestEventId,
			"app_id":           app.ID,
			"bolt11":           bolt11,
		}).WithError(err).Error("Failed to send payment")
		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error:      mapNip47Error(err),
		}, tags)
		return
	}

	publishResponse(&models.Response{
		ResultType: nip47Request.Method,
		Result: payResponse{
			Preimage: *transaction.Preimage,
			FeesPaid: transaction.FeeMsat,
		},
	}, tags)
}
