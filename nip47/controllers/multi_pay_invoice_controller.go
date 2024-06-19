package controllers

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/getAlby/nostr-wallet-connect/db"
	"github.com/getAlby/nostr-wallet-connect/events"
	"github.com/getAlby/nostr-wallet-connect/lnclient"
	"github.com/getAlby/nostr-wallet-connect/logger"
	"github.com/getAlby/nostr-wallet-connect/nip47/models"
	"github.com/nbd-wtf/go-nostr"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type multiPayInvoiceElement struct {
	payInvoiceParams
	Id string `json:"id"`
}

type multiPayInvoiceParams struct {
	Invoices []multiPayInvoiceElement `json:"invoices"`
}

type multiMultiPayInvoiceController struct {
	lnClient       lnclient.LNClient
	db             *gorm.DB
	eventPublisher events.EventPublisher
}

func NewMultiPayInvoiceController(lnClient lnclient.LNClient, db *gorm.DB, eventPublisher events.EventPublisher) *multiMultiPayInvoiceController {
	return &multiMultiPayInvoiceController{
		lnClient:       lnClient,
		db:             db,
		eventPublisher: eventPublisher,
	}
}

func (controller *multiMultiPayInvoiceController) HandleMultiPayInvoiceEvent(ctx context.Context, nip47Request *models.Request, requestEventId uint, app *db.App, checkPermission checkPermissionFunc, publishResponse publishFunc) {
	multiPayParams := &multiPayInvoiceParams{}
	resp := decodeRequest(nip47Request, multiPayParams)
	if resp != nil {
		publishResponse(resp, nostr.Tags{})
		return
	}

	var wg sync.WaitGroup
	for _, invoiceInfo := range multiPayParams.Invoices {
		wg.Add(1)
		go func(invoiceInfo multiPayInvoiceElement) {
			defer wg.Done()
			bolt11 := invoiceInfo.Invoice
			// Convert invoice to lowercase string
			bolt11 = strings.ToLower(bolt11)
			paymentRequest, err := decodepay.Decodepay(bolt11)
			if err != nil {
				logger.Logger.WithFields(logrus.Fields{
					"request_event_id": requestEventId,
					"appId":            app.ID,
					"bolt11":           bolt11,
				}).Errorf("Failed to decode bolt11 invoice: %v", err)

				// TODO: Decide what to do if id is empty
				dTag := []string{"d", invoiceInfo.Id}
				publishResponse(&models.Response{
					ResultType: nip47Request.Method,
					Error: &models.Error{
						Code:    models.ERROR_INTERNAL,
						Message: fmt.Sprintf("Failed to decode bolt11 invoice: %s", err.Error()),
					},
				}, nostr.Tags{dTag})
				return
			}

			invoiceDTagValue := invoiceInfo.Id
			if invoiceDTagValue == "" {
				invoiceDTagValue = paymentRequest.PaymentHash
			}
			dTag := []string{"d", invoiceDTagValue}

			NewPayInvoiceController(controller.lnClient, controller.db, controller.eventPublisher).
				pay(ctx, bolt11, &paymentRequest, nip47Request, requestEventId, app, checkPermission, publishResponse, nostr.Tags{dTag})
		}(invoiceInfo)
	}

	wg.Wait()
}
