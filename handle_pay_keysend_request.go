package main

import (
	"context"

	"github.com/getAlby/nostr-wallet-connect/events"
	"github.com/getAlby/nostr-wallet-connect/nip47"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

func (svc *Service) HandlePayKeysendEvent(ctx context.Context, nip47Request *Nip47Request, requestEvent *RequestEvent, app *App, publishResponse func(*Nip47Response, nostr.Tags)) {

	payParams := &Nip47KeysendParams{}
	resp := svc.decodeNip47Request(nip47Request, requestEvent, app, payParams)
	if resp != nil {
		publishResponse(resp, nostr.Tags{})
		return
	}

	resp = svc.checkPermission(nip47Request, requestEvent.NostrId, app, payParams.Amount)
	if resp != nil {
		publishResponse(resp, nostr.Tags{})
		return
	}

	payment := Payment{App: *app, RequestEvent: *requestEvent, Amount: uint(payParams.Amount / 1000)}
	err := svc.db.Create(&payment).Error
	if err != nil {
		publishResponse(&Nip47Response{
			ResultType: nip47Request.Method,
			Error: &Nip47Error{
				Code:    nip47.ERROR_INTERNAL,
				Message: err.Error(),
			},
		}, nostr.Tags{})
		return
	}

	svc.Logger.WithFields(logrus.Fields{
		"requestEventNostrId": requestEvent.NostrId,
		"appId":               app.ID,
		"senderPubkey":        payParams.Pubkey,
	}).Info("Sending payment")

	preimage, err := svc.lnClient.SendKeysend(ctx, payParams.Amount, payParams.Pubkey, payParams.Preimage, payParams.TLVRecords)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"requestEventNostrId": requestEvent.NostrId,
			"appId":               app.ID,
			"recipientPubkey":     payParams.Pubkey,
		}).Infof("Failed to send payment: %v", err)
		svc.EventPublisher.Publish(&events.Event{
			Event: "nwc_payment_failed",
			Properties: map[string]interface{}{
				// "error":   fmt.Sprintf("%v", err),
				"keysend": true,
				"amount":  payParams.Amount / 1000,
			},
		})
		publishResponse(&Nip47Response{
			ResultType: nip47Request.Method,
			Error: &Nip47Error{
				Code:    nip47.ERROR_INTERNAL,
				Message: err.Error(),
			},
		}, nostr.Tags{})
		return
	}
	payment.Preimage = &preimage
	svc.db.Save(&payment)
	svc.EventPublisher.Publish(&events.Event{
		Event: "nwc_payment_succeeded",
		Properties: map[string]interface{}{
			"keysend": true,
			"amount":  payParams.Amount / 1000,
		},
	})
	publishResponse(&Nip47Response{
		ResultType: nip47Request.Method,
		Result: Nip47PayResponse{
			Preimage: preimage,
		},
	}, nostr.Tags{})
}
