package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

func (svc *Service) HandleMultiPayKeysendEvent(ctx context.Context, sub *nostr.Subscription, request *Nip47Request, event *nostr.Event, app App, ss []byte) {

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

	multiPayParams := &Nip47MultiPayKeysendParams{}
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
	for _, keysendInfo := range multiPayParams.Invoices {
		wg.Add(1)
		go func(keysendInfo Nip47MultiPayKeysendElement) {
			defer wg.Done()

			keysendDTagValue := keysendInfo.Id
			if keysendDTagValue == "" {
				keysendDTagValue = keysendInfo.Pubkey
			}
			dTag := []string{"d", keysendDTagValue}

			hasPermission, code, message := svc.hasPermission(&app, event, NIP_47_PAY_INVOICE_METHOD, keysendInfo.Amount)

			if !hasPermission {
				svc.Logger.WithFields(logrus.Fields{
					"eventId":      event.ID,
					"eventKind":    event.Kind,
					"appId":        app.ID,
					"senderPubkey": keysendInfo.Pubkey,
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
						"eventId":      event.ID,
						"eventKind":    event.Kind,
						"senderPubkey": keysendInfo.Pubkey,
						"keysendId":    keysendInfo.Id,
					}).Errorf("Failed to process event: %v", err)
					return
				}
				svc.PublishEvent(ctx, sub, event, resp)
				return
			}

			payment := Payment{App: app, NostrEvent: nostrEvent, Amount: uint(keysendInfo.Amount / 1000)}
			insertPaymentResult := svc.db.Create(&payment)
			if insertPaymentResult.Error != nil {
				svc.Logger.WithFields(logrus.Fields{
					"eventId":      event.ID,
					"eventKind":    event.Kind,
					"senderPubkey": keysendInfo.Pubkey,
					"keysendId":    keysendInfo.Id,
				}).Errorf("Failed to process event: %v", insertPaymentResult.Error)
				return
			}

			svc.Logger.WithFields(logrus.Fields{
				"eventId":      event.ID,
				"eventKind":    event.Kind,
				"appId":        app.ID,
				"senderPubkey": keysendInfo.Pubkey,
			}).Info("Sending payment")

			preimage, err := svc.lnClient.SendKeysend(ctx, event.PubKey, keysendInfo.Amount/1000, keysendInfo.Pubkey, keysendInfo.Preimage, keysendInfo.TLVRecords)
			if err != nil {
				svc.Logger.WithFields(logrus.Fields{
					"eventId":      event.ID,
					"eventKind":    event.Kind,
					"appId":        app.ID,
					"senderPubkey": keysendInfo.Pubkey,
				}).Infof("Failed to send payment: %v", err)
				// TODO: https://github.com/getAlby/nostr-wallet-connect/issues/231
				nostrEvent.State = NOSTR_EVENT_STATE_HANDLER_ERROR
				svc.db.Save(&nostrEvent)

				resp, err := svc.createResponse(event, Nip47Response{
					ResultType: request.Method,
					Error: &Nip47Error{
						Code:    NIP_47_ERROR_INTERNAL,
						Message: fmt.Sprintf("Something went wrong while paying invoice: %s", err.Error()),
					},
				}, nostr.Tags{dTag}, ss)
				if err != nil {
					svc.Logger.WithFields(logrus.Fields{
						"eventId":      event.ID,
						"eventKind":    event.Kind,
						"senderPubkey": keysendInfo.Pubkey,
						"keysendId":    keysendInfo.Id,
					}).Errorf("Failed to process event: %v", err)
					return
				}
				svc.PublishEvent(ctx, sub, event, resp)
				return
			}
			payment.Preimage = &preimage
			// TODO: https://github.com/getAlby/nostr-wallet-connect/issues/231
			nostrEvent.State = NOSTR_EVENT_STATE_HANDLER_EXECUTED
			svc.db.Save(&nostrEvent)
			svc.db.Save(&payment)
			resp, err := svc.createResponse(event, Nip47Response{
				ResultType: request.Method,
				Result: Nip47PayResponse{
					Preimage: preimage,
				},
			}, nostr.Tags{dTag}, ss)
			if err != nil {
				svc.Logger.WithFields(logrus.Fields{
					"eventId":      event.ID,
					"eventKind":    event.Kind,
					"senderPubkey": keysendInfo.Pubkey,
					"keysendId":    keysendInfo.Id,
				}).Errorf("Failed to process event: %v", err)
				return
			}
			svc.PublishEvent(ctx, sub, event, resp)
		}(keysendInfo)
	}

	wg.Wait()
	return
}
