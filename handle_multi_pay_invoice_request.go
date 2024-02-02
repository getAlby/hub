package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/nbd-wtf/go-nostr"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
)

func (svc *Service) HandleMultiPayInvoiceEvent(ctx context.Context, sub *nostr.Subscription, request *Nip47Request, event *nostr.Event, requestEvent *NostrEvent, app *App, ss []byte) (resps []*nostr.Event, err error) {

	multiPayParams := &Nip47MultiPayInvoiceParams{}
	err = json.Unmarshal(request.Params, multiPayParams)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
			"appId":     app.ID,
		}).Errorf("Failed to decode nostr event: %v", err)
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
		}).Errorf("Failed to process event: %v", err)
		return nil, err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, invoiceInfo := range multiPayParams.Invoices {
		wg.Add(1)
		go func(invoiceInfo Nip47MultiPayInvoiceElement) {
			defer wg.Done()
			bolt11 := invoiceInfo.Invoice
			// Convert invoice to lowercase string
			bolt11 = strings.ToLower(bolt11)
			paymentRequest, err := decodepay.Decodepay(bolt11)
			if err != nil {
				svc.Logger.WithFields(logrus.Fields{
					"eventId":   event.ID,
					"eventKind": event.Kind,
					"appId":     app.ID,
					"bolt11":    bolt11,
				}).Errorf("Failed to decode bolt11 invoice: %v", err)

				// TODO: Decide what to do if id is empty
				dTag := []string{"d", invoiceInfo.Id}
				resp, err := svc.createResponse(event, Nip47Response{
					ResultType: request.Method,
					Error: &Nip47Error{
						Code:    NIP_47_ERROR_INTERNAL,
						Message: fmt.Sprintf("Failed to decode bolt11 invoice: %s", err.Error()),
					},
				}, nostr.Tags{dTag}, ss)
				if err != nil {
					svc.Logger.WithFields(logrus.Fields{
						"eventId":        event.ID,
						"eventKind":      event.Kind,
						"paymentRequest": invoiceInfo.Invoice,
						"invoiceId":      invoiceInfo.Id,
					}).Errorf("Failed to process event: %v", err)
					return
				}

				svc.PublishAndAppend(ctx, sub, requestEvent, resp, app, ss, &mu, &resps)
				return
			}

			invoiceDTagValue := invoiceInfo.Id
			if invoiceDTagValue == "" {
				invoiceDTagValue = paymentRequest.PaymentHash
			}
			dTag := []string{"d", invoiceDTagValue}

			hasPermission, code, message := svc.hasPermission(app, requestEvent, NIP_47_PAY_INVOICE_METHOD, paymentRequest.MSatoshi)

			if !hasPermission {
				svc.Logger.WithFields(logrus.Fields{
					"eventId":   event.ID,
					"eventKind": event.Kind,
					"appId":     app.ID,
				}).Errorf("App does not have permission: %s %s", code, message)

				resp, err := svc.createResponse(event, Nip47Response{
					ResultType: request.Method,
					Error: &Nip47Error{
						Code:    code,
						Message: message,
					},
				}, nostr.Tags{dTag}, ss)
				if err != nil {
					svc.Logger.WithFields(logrus.Fields{
						"eventId":        event.ID,
						"eventKind":      event.Kind,
						"paymentRequest": invoiceInfo.Invoice,
						"invoiceId":      invoiceInfo.Id,
					}).Errorf("Failed to process event: %v", err)
					return
				}
				svc.PublishAndAppend(ctx, sub, requestEvent, resp, app, ss, &mu, &resps)
				return
			}

			payment := Payment{App: *app, NostrEvent: *requestEvent, PaymentRequest: bolt11, Amount: uint(paymentRequest.MSatoshi / 1000)}
			insertPaymentResult := svc.db.Create(&payment)
			if insertPaymentResult.Error != nil {
				svc.Logger.WithFields(logrus.Fields{
					"eventId":        event.ID,
					"eventKind":      event.Kind,
					"paymentRequest": bolt11,
					"invoiceId":      invoiceInfo.Id,
				}).Errorf("Failed to process event: %v", insertPaymentResult.Error)
				return
			}

			svc.Logger.WithFields(logrus.Fields{
				"eventId":   event.ID,
				"eventKind": event.Kind,
				"appId":     app.ID,
				"bolt11":    bolt11,
			}).Info("Sending payment")

			preimage, err := svc.lnClient.SendPaymentSync(ctx, event.PubKey, bolt11)
			if err != nil {
				svc.Logger.WithFields(logrus.Fields{
					"eventId":   event.ID,
					"eventKind": event.Kind,
					"appId":     app.ID,
					"bolt11":    bolt11,
				}).Infof("Failed to send payment: %v", err)

				resp, err := svc.createResponse(event, Nip47Response{
					ResultType: request.Method,
					Error: &Nip47Error{
						Code:    NIP_47_ERROR_INTERNAL,
						Message: fmt.Sprintf("Something went wrong while paying invoice: %s", err.Error()),
					},
				}, nostr.Tags{dTag}, ss)
				if err != nil {
					svc.Logger.WithFields(logrus.Fields{
						"eventId":        event.ID,
						"eventKind":      event.Kind,
						"paymentRequest": invoiceInfo.Invoice,
						"invoiceId":      invoiceInfo.Id,
					}).Errorf("Failed to process event: %v", err)
					return
				}
				svc.PublishAndAppend(ctx, sub, requestEvent, resp, app, ss, &mu, &resps)
				return
			}
			payment.Preimage = &preimage
			svc.db.Save(&payment)
			resp, err := svc.createResponse(event, Nip47Response{
				ResultType: request.Method,
				Result: Nip47PayResponse{
					Preimage: preimage,
				},
			}, nostr.Tags{dTag}, ss)
			if err != nil {
				svc.Logger.WithFields(logrus.Fields{
					"eventId":        event.ID,
					"eventKind":      event.Kind,
					"paymentRequest": invoiceInfo.Invoice,
					"invoiceId":      invoiceInfo.Id,
				}).Errorf("Failed to process event: %v", err)
				return
			}
			svc.PublishAndAppend(ctx, sub, requestEvent, resp, app, ss, &mu, &resps)
		}(invoiceInfo)
	}

	wg.Wait()
	return resps, nil
}
