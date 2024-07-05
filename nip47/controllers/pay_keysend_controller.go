package controllers

import (
	"context"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type payKeysendParams struct {
	Amount     uint64               `json:"amount"`
	Pubkey     string               `json:"pubkey"`
	Preimage   string               `json:"preimage"`
	TLVRecords []lnclient.TLVRecord `json:"tlv_records"`
}

type payKeysendController struct {
	lnClient       lnclient.LNClient
	db             *gorm.DB
	eventPublisher events.EventPublisher
}

func NewPayKeysendController(lnClient lnclient.LNClient, db *gorm.DB, eventPublisher events.EventPublisher) *payKeysendController {
	return &payKeysendController{
		lnClient:       lnClient,
		db:             db,
		eventPublisher: eventPublisher,
	}
}

func (controller *payKeysendController) HandlePayKeysendEvent(ctx context.Context, nip47Request *models.Request, requestEventId uint, app *db.App, checkPermission checkPermissionFunc, publishResponse publishFunc, tags nostr.Tags) {
	payKeysendParams := &payKeysendParams{}
	resp := decodeRequest(nip47Request, payKeysendParams)
	if resp != nil {
		publishResponse(resp, tags)
		return
	}
	controller.pay(ctx, payKeysendParams, nip47Request, requestEventId, app, checkPermission, publishResponse, tags)
}

func (controller *payKeysendController) pay(ctx context.Context, payKeysendParams *payKeysendParams, nip47Request *models.Request, requestEventId uint, app *db.App, checkPermission checkPermissionFunc, publishResponse publishFunc, tags nostr.Tags) {
	resp := checkPermission(payKeysendParams.Amount)
	if resp != nil {
		publishResponse(resp, tags)
		return
	}

	payment := db.Payment{App: *app, RequestEventId: requestEventId, Amount: uint(payKeysendParams.Amount / 1000)}
	err := controller.db.Create(&payment).Error
	if err != nil {
		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error: &models.Error{
				Code:    models.ERROR_INTERNAL,
				Message: err.Error(),
			},
		}, tags)
		return
	}

	logger.Logger.WithFields(logrus.Fields{
		"request_event_id": requestEventId,
		"appId":            app.ID,
		"senderPubkey":     payKeysendParams.Pubkey,
	}).Info("Sending keysend payment")

	preimage, err := controller.lnClient.SendKeysend(ctx, payKeysendParams.Amount, payKeysendParams.Pubkey, payKeysendParams.Preimage, payKeysendParams.TLVRecords)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"request_event_id": requestEventId,
			"appId":            app.ID,
			"recipientPubkey":  payKeysendParams.Pubkey,
		}).Infof("Failed to send keysend payment: %v", err)
		controller.eventPublisher.Publish(&events.Event{
			Event: "nwc_payment_failed",
			Properties: map[string]interface{}{
				"error":   err.Error(),
				"keysend": true,
				"amount":  payKeysendParams.Amount / 1000,
			},
		})
		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error: &models.Error{
				Code:    models.ERROR_INTERNAL,
				Message: err.Error(),
			},
		}, tags)
		return
	}
	payment.Preimage = &preimage
	controller.db.Save(&payment)
	controller.eventPublisher.Publish(&events.Event{
		Event: "nwc_payment_succeeded",
		Properties: map[string]interface{}{
			"keysend": true,
			"amount":  payKeysendParams.Amount / 1000,
		},
	})
	publishResponse(&models.Response{
		ResultType: nip47Request.Method,
		Result: payResponse{
			Preimage: preimage,
		},
	}, tags)
}
