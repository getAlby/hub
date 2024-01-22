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

func (svc *Service) HandleMultiPayInvoiceEvent(ctx context.Context, sub *nostr.Subscription, request *Nip47Request, event *nostr.Event, app App, ss []byte) {

	nostrEvent := NostrEvent{App: app, NostrId: event.ID, Content: event.Content, State: "received"}
	err := svc.db.Create(&nostrEvent).Error
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
			"appId":     app.ID,
		}).Errorf("Failed to save nostr event: %v", err)
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
		}).Errorf("Failed to process event: %v", err)
		return
	}

	multiPayParams := &Nip47MultiPayParams{}
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
		return
	}

	var wg sync.WaitGroup
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
					ResultType: NIP_47_MULTI_PAY_INVOICE_METHOD,
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

				svc.PublishEvent(ctx, sub, event, resp)
				return
			}

			invoiceDTagValue := invoiceInfo.Id
			if invoiceDTagValue == "" {
				invoiceDTagValue = paymentRequest.PaymentHash
			}
			dTag := []string{"d", invoiceDTagValue}

			hasPermission, code, message := svc.hasPermission(&app, event, NIP_47_PAY_INVOICE_METHOD, paymentRequest.MSatoshi)

			if !hasPermission {
				svc.Logger.WithFields(logrus.Fields{
					"eventId":   event.ID,
					"eventKind": event.Kind,
					"appId":     app.ID,
				}).Errorf("App does not have permission: %s %s", code, message)

				resp, err := svc.createResponse(event, Nip47Response{
					ResultType: NIP_47_MULTI_PAY_INVOICE_METHOD,
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
				svc.PublishEvent(ctx, sub, event, resp)
				return
			}

			payment := Payment{App: app, NostrEvent: nostrEvent, PaymentRequest: bolt11, Amount: uint(paymentRequest.MSatoshi / 1000)}
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
				// TODO: What to do here?
				nostrEvent.State = NOSTR_EVENT_STATE_HANDLER_ERROR
				svc.db.Save(&nostrEvent)

				resp, err := svc.createResponse(event, Nip47Response{
					ResultType: NIP_47_MULTI_PAY_INVOICE_METHOD,
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
				svc.PublishEvent(ctx, sub, event, resp)
				return
			}
			payment.Preimage = &preimage
			// TODO: What to do here?
			nostrEvent.State = NOSTR_EVENT_STATE_HANDLER_EXECUTED
			svc.db.Save(&nostrEvent)
			svc.db.Save(&payment)
			resp, err := svc.createResponse(event, Nip47Response{
				ResultType: NIP_47_MULTI_PAY_INVOICE_METHOD,
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
			svc.PublishEvent(ctx, sub, event, resp)
		}(invoiceInfo)
	}

	wg.Wait()
	return
}
