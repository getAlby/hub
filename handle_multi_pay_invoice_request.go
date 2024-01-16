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

func (svc *Service) HandleMultiPayInvoiceEvent(ctx context.Context, request *Nip47Request, event *nostr.Event, app App, ss []byte) (results []*nostr.Event, err error) {

	nostrEvent := NostrEvent{App: app, NostrId: event.ID, Content: event.Content, State: "received"}
	err = svc.db.Create(&nostrEvent).Error
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
			"appId":     app.ID,
		}).Errorf("Failed to save nostr event: %v", err)
		return nil, err
	}

	multiPayParams := &Nip47MultiPayParams{}
	err = json.Unmarshal(request.Params, multiPayParams)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
			"appId":     app.ID,
		}).Errorf("Failed to decode nostr event: %v", err)
		return nil, err
	}

	var wg sync.WaitGroup
	for _, invoiceInfo := range multiPayParams.Invoices {
		wg.Add(1)
		go func(invoiceInfo InvoiceInfo) {
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

				results = svc.createAndAppendResponse(event, Nip47Response{
					ResultType: NIP_47_PAY_INVOICE_METHOD,
					Error: &Nip47Error{
						Code:    NIP_47_ERROR_INTERNAL,
						Message: fmt.Sprintf("Failed to decode bolt11 invoice: %s", err.Error()),
					},
				}, ss, invoiceInfo, results)
				return
			}

			hasPermission, code, message := svc.hasPermission(&app, event, request.Method, paymentRequest.MSatoshi)

			if !hasPermission {
				svc.Logger.WithFields(logrus.Fields{
					"eventId":   event.ID,
					"eventKind": event.Kind,
					"appId":     app.ID,
				}).Errorf("App does not have permission: %s %s", code, message)

				results = svc.createAndAppendResponse(event, Nip47Response{
					ResultType: NIP_47_PAY_INVOICE_METHOD,
					Error: &Nip47Error{
						Code:    code,
						Message: message,
					},
				}, ss, invoiceInfo, results)
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

				results = svc.createAndAppendResponse(event, Nip47Response{
					ResultType: NIP_47_PAY_INVOICE_METHOD,
					Error: &Nip47Error{
						Code:    NIP_47_ERROR_INTERNAL,
						Message: fmt.Sprintf("Something went wrong while paying invoice: %s", err.Error()),
					},
				}, ss, invoiceInfo, results)
				return
			}
			payment.Preimage = &preimage
			// TODO: What to do here?
			nostrEvent.State = NOSTR_EVENT_STATE_HANDLER_EXECUTED
			svc.db.Save(&nostrEvent)
			svc.db.Save(&payment)
			results = svc.createAndAppendResponse(event, Nip47Response{
				ResultType: NIP_47_PAY_INVOICE_METHOD,
				Result: Nip47PayResponse{
					Preimage: preimage,
				},
			}, ss, invoiceInfo, results)
			return
		}(invoiceInfo)
	}

	wg.Wait()
	return results, nil
}

func (svc *Service) createAndAppendResponse(initialEvent *nostr.Event, content interface{}, ss []byte, invoiceInfo InvoiceInfo, results []*nostr.Event) (result []*nostr.Event) {
	resp, err := svc.createResponse(initialEvent, content, ss)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":        initialEvent.ID,
			"eventKind":      initialEvent.Kind,
			"paymentRequest": invoiceInfo.Invoice,
			"invoiceId":      invoiceInfo.Id,
		}).Errorf("Failed to process event: %v", err)
		return results
	}
	dTag := []string{"a", fmt.Sprintf("%d:%s:%d", NIP_47_RESPONSE_KIND, initialEvent.PubKey, invoiceInfo.Id)}
	resp.Tags = append(resp.Tags, dTag)
	return append(results, resp)
}
