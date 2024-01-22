package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

func (svc *Service) HandleListTransactionsEvent(ctx context.Context, request *Nip47Request, event *nostr.Event, app App, ss []byte) (result *nostr.Event, err error) {

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

	listParams := &Nip47ListTransactionsParams{}
	err = json.Unmarshal(request.Params, listParams)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
			"appId":     app.ID,
		}).Errorf("Failed to decode nostr event: %v", err)
		return nil, err
	}

	hasPermission, code, message := svc.hasPermission(&app, event, request.Method, 0)

	if !hasPermission {
		svc.Logger.WithFields(logrus.Fields{
			// TODO: log request fields from listParams
			"eventId":   event.ID,
			"eventKind": event.Kind,
			"appId":     app.ID,
		}).Errorf("App does not have permission: %s %s", code, message)

		return svc.createResponse(event, Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    code,
				Message: message,
			}}, nostr.Tags{}, ss)
	}

	svc.Logger.WithFields(logrus.Fields{
		// TODO: log request fields from listParams
		"eventId":   event.ID,
		"eventKind": event.Kind,
		"appId":     app.ID,
	}).Info("Fetching transactions")

	transactions, err := svc.lnClient.ListTransactions(ctx, event.PubKey, listParams.From, listParams.Until, listParams.Limit, listParams.Offset, listParams.Unpaid, listParams.Type)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			// TODO: log request fields from listParams
			"eventId":   event.ID,
			"eventKind": event.Kind,
			"appId":     app.ID,
		}).Infof("Failed to fetch transactions: %v", err)
		nostrEvent.State = NOSTR_EVENT_STATE_HANDLER_ERROR
		svc.db.Save(&nostrEvent)
		return svc.createResponse(event, Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: fmt.Sprintf("Something went wrong while fetching transactions: %s", err.Error()),
			},
		}, nostr.Tags{}, ss)
	}

	responsePayload := &Nip47ListTransactionsResponse{
		Transactions: transactions,
	}
	// fmt.Println(responsePayload)

	nostrEvent.State = NOSTR_EVENT_STATE_HANDLER_EXECUTED
	svc.db.Save(&nostrEvent)
	return svc.createResponse(event, Nip47Response{
		ResultType: request.Method,
		Result:     responsePayload,
	}, nostr.Tags{}, ss)
}
