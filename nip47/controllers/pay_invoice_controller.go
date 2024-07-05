package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	"github.com/nbd-wtf/go-nostr"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type payInvoiceParams struct {
	Invoice string `json:"invoice"`
}

type payInvoiceController struct {
	lnClient       lnclient.LNClient
	db             *gorm.DB
	eventPublisher events.EventPublisher
}

func NewPayInvoiceController(lnClient lnclient.LNClient, db *gorm.DB, eventPublisher events.EventPublisher) *payInvoiceController {
	return &payInvoiceController{
		lnClient:       lnClient,
		db:             db,
		eventPublisher: eventPublisher,
	}
}

func (controller *payInvoiceController) HandlePayInvoiceEvent(ctx context.Context, nip47Request *models.Request, requestEventId uint, app *db.App, checkPermission checkPermissionFunc, publishResponse publishFunc, tags nostr.Tags) {
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
				Code:    models.ERROR_INTERNAL,
				Message: fmt.Sprintf("Failed to decode bolt11 invoice: %s", err.Error()),
			},
		}, tags)
		return
	}

	controller.pay(ctx, bolt11, &paymentRequest, nip47Request, requestEventId, app, checkPermission, publishResponse, tags)
}

func (controller *payInvoiceController) pay(ctx context.Context, bolt11 string, paymentRequest *decodepay.Bolt11, nip47Request *models.Request, requestEventId uint, app *db.App, checkPermission checkPermissionFunc, publishResponse publishFunc, tags nostr.Tags) {
	resp := checkPermission(uint64(paymentRequest.MSatoshi))
	if resp != nil {
		publishResponse(resp, tags)
		return
	}

	payment := db.Payment{App: *app, RequestEventId: requestEventId, PaymentRequest: bolt11, Amount: uint(paymentRequest.MSatoshi / 1000)}
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
		"app_id":           app.ID,
		"bolt11":           bolt11,
	}).Info("Sending payment")

	response, err := controller.lnClient.SendPaymentSync(ctx, bolt11)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"request_event_id": requestEventId,
			"app_id":           app.ID,
			"bolt11":           bolt11,
		}).Infof("Failed to send payment: %v", err)
		controller.eventPublisher.Publish(&events.Event{
			Event: "nwc_payment_failed",
			Properties: map[string]interface{}{
				"error":   err.Error(),
				"invoice": bolt11,
				"amount":  paymentRequest.MSatoshi / 1000,
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
	payment.Preimage = &response.Preimage
	// TODO: save payment fee
	controller.db.Save(&payment)

	controller.eventPublisher.Publish(&events.Event{
		Event: "nwc_payment_succeeded",
		Properties: map[string]interface{}{
			"bolt11": bolt11,
			"amount": paymentRequest.MSatoshi / 1000,
		},
	})

	publishResponse(&models.Response{
		ResultType: nip47Request.Method,
		Result: payResponse{
			Preimage: response.Preimage,
			FeesPaid: response.Fee,
		},
	}, tags)
}
