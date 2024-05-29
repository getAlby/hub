package main

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/getAlby/nostr-wallet-connect/db"
	"github.com/getAlby/nostr-wallet-connect/events"
	"github.com/getAlby/nostr-wallet-connect/nip47"
	"github.com/nbd-wtf/go-nostr"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
)

// TODO: pass a channel instead of publishResponse function
func (svc *Service) HandleMultiPayInvoiceEvent(ctx context.Context, nip47Request *nip47.Request, requestEvent *db.RequestEvent, app *db.App, publishResponse func(*nip47.Response, nostr.Tags)) {

	multiPayParams := &nip47.MultiPayInvoiceParams{}
	resp := svc.decodeNip47Request(nip47Request, requestEvent, app, multiPayParams)
	if resp != nil {
		publishResponse(resp, nostr.Tags{})
		return
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, invoiceInfo := range multiPayParams.Invoices {
		wg.Add(1)
		// TODO: we should call the handle_payment_request (most of this code is duplicated)
		go func(invoiceInfo nip47.MultiPayInvoiceElement) {
			defer wg.Done()
			bolt11 := invoiceInfo.Invoice
			// Convert invoice to lowercase string
			bolt11 = strings.ToLower(bolt11)
			paymentRequest, err := decodepay.Decodepay(bolt11)
			if err != nil {
				svc.logger.WithFields(logrus.Fields{
					"requestEventNostrId": requestEvent.NostrId,
					"appId":               app.ID,
					"bolt11":              bolt11,
				}).Errorf("Failed to decode bolt11 invoice: %v", err)

				// TODO: Decide what to do if id is empty
				dTag := []string{"d", invoiceInfo.Id}
				publishResponse(&nip47.Response{
					ResultType: nip47Request.Method,
					Error: &nip47.Error{
						Code:    nip47.ERROR_INTERNAL,
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

			resp := svc.checkPermission(nip47Request, requestEvent.NostrId, app, paymentRequest.MSatoshi)
			if resp != nil {
				publishResponse(resp, nostr.Tags{dTag})
				return
			}

			payment := db.Payment{App: *app, RequestEventId: requestEvent.ID, PaymentRequest: bolt11, Amount: uint(paymentRequest.MSatoshi / 1000)}
			mu.Lock()
			insertPaymentResult := svc.db.Create(&payment)
			mu.Unlock()
			if insertPaymentResult.Error != nil {
				svc.logger.WithFields(logrus.Fields{
					"requestEventNostrId": requestEvent.NostrId,
					"paymentRequest":      bolt11,
					"invoiceId":           invoiceInfo.Id,
				}).Errorf("Failed to process event: %v", insertPaymentResult.Error)
				return
			}

			svc.logger.WithFields(logrus.Fields{
				"requestEventNostrId": requestEvent.NostrId,
				"appId":               app.ID,
				"bolt11":              bolt11,
			}).Info("Sending payment")

			response, err := svc.lnClient.SendPaymentSync(ctx, bolt11)
			if err != nil {
				svc.logger.WithFields(logrus.Fields{
					"requestEventNostrId": requestEvent.NostrId,
					"appId":               app.ID,
					"bolt11":              bolt11,
				}).Infof("Failed to send payment: %v", err)

				svc.eventPublisher.Publish(&events.Event{
					Event: "nwc_payment_failed",
					Properties: map[string]interface{}{
						// "error":   fmt.Sprintf("%v", err),
						"multi":   true,
						"invoice": bolt11,
						"amount":  paymentRequest.MSatoshi / 1000,
					},
				})

				publishResponse(&nip47.Response{
					ResultType: nip47Request.Method,
					Error: &nip47.Error{
						Code:    nip47.ERROR_INTERNAL,
						Message: err.Error(),
					},
				}, nostr.Tags{dTag})
				return
			}
			payment.Preimage = &response.Preimage
			// TODO: also set fee

			mu.Lock()
			svc.db.Save(&payment)
			mu.Unlock()
			svc.eventPublisher.Publish(&events.Event{
				Event: "nwc_payment_succeeded",
				Properties: map[string]interface{}{
					"multi":  true,
					"amount": paymentRequest.MSatoshi / 1000,
				},
			})
			publishResponse(&nip47.Response{
				ResultType: nip47Request.Method,
				Result: nip47.PayResponse{
					Preimage: response.Preimage,
					FeesPaid: response.Fee,
				},
			}, nostr.Tags{dTag})
		}(invoiceInfo)
	}

	wg.Wait()
}
