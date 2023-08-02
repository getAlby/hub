package main

import (
	"context"
	"fmt"

	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

func (svc *Service) HandleGetBalanceEvent(ctx context.Context, request *Nip47Request, event *nostr.Event, app App, ss []byte) (result *nostr.Event, err error) {

	nostrEvent := NostrEvent{App: app, NostrId: event.ID, Content: event.Content, State: "received"}
	insertNostrEventResult := svc.db.Create(&nostrEvent)
	if insertNostrEventResult.Error != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
			"appId":     app.ID,
		}).Errorf("Failed to save nostr event: %v", insertNostrEventResult.Error)
		return nil, insertNostrEventResult.Error
	}

	hasPermission, code, message := svc.hasPermission(&app, event, request.Method, nil)

	if !hasPermission {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
			"appId":     app.ID,
		}).Errorf("App does not have permission: %s %s", code, message)

		return svc.createResponse(event, Nip47Response{Error: &Nip47Error{
			Code:    code,
			Message: message,
		}}, ss)
	}

	svc.Logger.WithFields(logrus.Fields{
		"eventId":   event.ID,
		"eventKind": event.Kind,
		"appId":     app.ID,
	}).Info("Fetching balance")

	balance, err := svc.lnClient.GetBalance(ctx, event.PubKey)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
			"appId":     app.ID,
		}).Infof("Failed to fetch balance: %v", err)
		nostrEvent.State = "error"
		svc.db.Save(&nostrEvent)
		return svc.createResponse(event, Nip47Response{
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: fmt.Sprintf("Something went wrong while fetching balance: %s", err.Error()),
			},
		}, ss)
	}

	responsePayload := &Nip47BalanceResponse{
		Balance: balance,
	}

	appPermission := AppPermission{}
	svc.db.Where("app_id = ? AND request_method = ?", app.ID, NIP_47_PAY_INVOICE_METHOD).First(&appPermission)

	maxAmount := appPermission.MaxAmount
	if maxAmount > 0 {
		responsePayload.MaxAmount = maxAmount
		responsePayload.BudgetRenewal = appPermission.BudgetRenewal
	}

	nostrEvent.State = "executed"
	svc.db.Save(&nostrEvent)
	return svc.createResponse(event, Nip47Response{
		ResultType: NIP_47_GET_BALANCE_METHOD,
		Result: responsePayload,
	},
	ss)
}
