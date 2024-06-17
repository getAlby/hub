package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/getAlby/nostr-wallet-connect/lnclient"
	"github.com/getAlby/nostr-wallet-connect/logger"
	"github.com/getAlby/nostr-wallet-connect/nip47/models"
	"github.com/nbd-wtf/go-nostr"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
)

type lookupInvoiceParams struct {
	Invoice     string `json:"invoice"`
	PaymentHash string `json:"payment_hash"`
}

type lookupInvoiceResponse struct {
	models.Transaction
}

type lookupInvoiceController struct {
	lnClient lnclient.LNClient
}

func NewLookupInvoiceController(lnClient lnclient.LNClient) *lookupInvoiceController {
	return &lookupInvoiceController{
		lnClient: lnClient,
	}
}

func (controller *lookupInvoiceController) HandleLookupInvoiceEvent(ctx context.Context, nip47Request *models.Request, requestEventId uint, checkPermission checkPermissionFunc, publishResponse publishFunc) {
	// basic permissions check
	resp := checkPermission(0)
	if resp != nil {
		publishResponse(resp, nostr.Tags{})
		return
	}

	lookupInvoiceParams := &lookupInvoiceParams{}
	resp = decodeRequest(nip47Request, lookupInvoiceParams)
	if resp != nil {
		publishResponse(resp, nostr.Tags{})
		return
	}

	logger.Logger.WithFields(logrus.Fields{
		"invoice":          lookupInvoiceParams.Invoice,
		"payment_hash":     lookupInvoiceParams.PaymentHash,
		"request_event_id": requestEventId,
	}).Info("Looking up invoice")

	paymentHash := lookupInvoiceParams.PaymentHash

	if paymentHash == "" {
		paymentRequest, err := decodepay.Decodepay(strings.ToLower(lookupInvoiceParams.Invoice))
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"request_event_id": requestEventId,
				"invoice":          lookupInvoiceParams.Invoice,
			}).WithError(err).Error("Failed to decode bolt11 invoice")

			publishResponse(&models.Response{
				ResultType: nip47Request.Method,
				Error: &models.Error{
					Code:    models.ERROR_INTERNAL,
					Message: fmt.Sprintf("Failed to decode bolt11 invoice: %s", err.Error()),
				},
			}, nostr.Tags{})
			return
		}
		paymentHash = paymentRequest.PaymentHash
	}

	transaction, err := controller.lnClient.LookupInvoice(ctx, paymentHash)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"request_event_id": requestEventId,
			"invoice":          lookupInvoiceParams.Invoice,
			"payment_hash":     paymentHash,
		}).Infof("Failed to lookup invoice: %v", err)

		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error: &models.Error{
				Code:    models.ERROR_INTERNAL,
				Message: err.Error(),
			},
		}, nostr.Tags{})
		return
	}

	responsePayload := &lookupInvoiceResponse{
		Transaction: *transaction,
	}

	publishResponse(&models.Response{
		ResultType: nip47Request.Method,
		Result:     responsePayload,
	}, nostr.Tags{})
}
