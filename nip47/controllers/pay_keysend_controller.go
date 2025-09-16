package controllers

import (
	"context"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

type payKeysendParams struct {
	Amount     uint64               `json:"amount"`
	Pubkey     string               `json:"pubkey"`
	Preimage   string               `json:"preimage"`
	TLVRecords []lnclient.TLVRecord `json:"tlv_records"`
}

func (controller *nip47Controller) HandlePayKeysendEvent(ctx context.Context, nip47Request *models.Request, requestEventId uint, app *db.App, publishResponse publishFunc, tags nostr.Tags) {
	payKeysendParams := &payKeysendParams{}
	resp := decodeRequest(nip47Request, payKeysendParams)
	if resp != nil {
		publishResponse(resp, tags)
		return
	}
	controller.payKeysend(ctx, payKeysendParams, nip47Request, requestEventId, app, publishResponse, tags)
}

func (controller *nip47Controller) payKeysend(ctx context.Context, payKeysendParams *payKeysendParams, nip47Request *models.Request, requestEventId uint, app *db.App, publishResponse publishFunc, tags nostr.Tags) {
	logger.Logger.WithFields(logrus.Fields{
		"request_event_id": requestEventId,
		"appId":            app.ID,
		"senderPubkey":     payKeysendParams.Pubkey,
	}).Info("Sending keysend payment")

	transaction, err := controller.transactionsService.SendKeysend(payKeysendParams.Amount, payKeysendParams.Pubkey, payKeysendParams.TLVRecords, payKeysendParams.Preimage, controller.lnClient, &app.ID, &requestEventId)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"request_event_id": requestEventId,
			"appId":            app.ID,
			"recipientPubkey":  payKeysendParams.Pubkey,
		}).Infof("Failed to send keysend payment: %v", err)
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
