package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/getAlby/nostr-wallet-connect/events"
	"github.com/nbd-wtf/go-nostr"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
)

// TODO: pass a channel instead of publishResponse function
func (svc *Service) HandleMultiPayInvoiceEvent(ctx context.Context, request *Nip47Request, requestEvent *RequestEvent, app *App, publishResponse func(*Nip47Response, nostr.Tags)) (err error) {

	multiPayParams := &Nip47MultiPayInvoiceParams{}
	err = json.Unmarshal(request.Params, multiPayParams)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId": requestEvent.NostrId,
			"appId":   app.ID,
		}).Errorf("Failed to decode nostr event: %v", err)
		svc.Logger.WithFields(logrus.Fields{
			"eventId": requestEvent.NostrId,
		}).Errorf("Failed to process event: %v", err)
		return err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, invoiceInfo := range multiPayParams.Invoices {
		wg.Add(1)
		// TODO: we should call the handle_payment_request (most of this code is duplicated)
		go func(invoiceInfo Nip47MultiPayInvoiceElement) {
			defer wg.Done()
			bolt11 := invoiceInfo.Invoice
			// Convert invoice to lowercase string
			bolt11 = strings.ToLower(bolt11)
			paymentRequest, err := decodepay.Decodepay(bolt11)
			if err != nil {
				svc.Logger.WithFields(logrus.Fields{
					"eventId": requestEvent.NostrId,
					"appId":   app.ID,
					"bolt11":  bolt11,
				}).Errorf("Failed to decode bolt11 invoice: %v", err)

				// TODO: Decide what to do if id is empty
				dTag := []string{"d", invoiceInfo.Id}
				publishResponse(&Nip47Response{
					ResultType: request.Method,
					Error: &Nip47Error{
						Code:    NIP_47_ERROR_INTERNAL,
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

			hasPermission, code, message := svc.hasPermission(app, NIP_47_PAY_INVOICE_METHOD, paymentRequest.MSatoshi)

			if !hasPermission {
				svc.Logger.WithFields(logrus.Fields{
					"eventId": requestEvent.NostrId,
					"appId":   app.ID,
				}).Errorf("App does not have permission: %s %s", code, message)

				publishResponse(&Nip47Response{
					ResultType: request.Method,
					Error: &Nip47Error{
						Code:    code,
						Message: message,
					},
				}, nostr.Tags{dTag})
				return
			}

			payment := Payment{App: *app, RequestEventId: requestEvent.ID, PaymentRequest: bolt11, Amount: uint(paymentRequest.MSatoshi / 1000)}
			mu.Lock()
			insertPaymentResult := svc.db.Create(&payment)
			mu.Unlock()
			if insertPaymentResult.Error != nil {
				svc.Logger.WithFields(logrus.Fields{
					"eventId":        requestEvent.NostrId,
					"paymentRequest": bolt11,
					"invoiceId":      invoiceInfo.Id,
				}).Errorf("Failed to process event: %v", insertPaymentResult.Error)
				return
			}

			svc.Logger.WithFields(logrus.Fields{
				"eventId": requestEvent.NostrId,
				"appId":   app.ID,
				"bolt11":  bolt11,
			}).Info("Sending payment")

			preimage, err := svc.lnClient.SendPaymentSync(ctx, bolt11)
			if err != nil {
				svc.Logger.WithFields(logrus.Fields{
					"eventId": requestEvent.NostrId,
					"appId":   app.ID,
					"bolt11":  bolt11,
				}).Infof("Failed to send payment: %v", err)

				svc.EventLogger.Log(&events.Event{
					Event: "nwc_payment_failed",
					Properties: map[string]interface{}{
						"error":   fmt.Sprintf("%v", err),
						"multi":   true,
						"invoice": bolt11,
						"amount":  paymentRequest.MSatoshi / 1000,
					},
				})

				publishResponse(&Nip47Response{
					ResultType: request.Method,
					Error: &Nip47Error{
						Code:    NIP_47_ERROR_INTERNAL,
						Message: err.Error(),
					},
				}, nostr.Tags{dTag})
				return
			}
			payment.Preimage = &preimage
			mu.Lock()
			svc.db.Save(&payment)
			mu.Unlock()
			svc.EventLogger.Log(&events.Event{
				Event: "nwc_payment_succeeded",
				Properties: map[string]interface{}{
					"multi":  true,
					"amount": paymentRequest.MSatoshi / 1000,
				},
			})
			publishResponse(&Nip47Response{
				ResultType: request.Method,
				Result: Nip47PayResponse{
					Preimage: preimage,
				},
			}, nostr.Tags{dTag})
		}(invoiceInfo)
	}

	wg.Wait()
	return nil
}
