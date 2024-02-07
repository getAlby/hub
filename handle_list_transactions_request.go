package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

func (svc *Service) HandleListTransactionsEvent(ctx context.Context, request *Nip47Request, requestEvent *RequestEvent, app *App) (result *Nip47Response, err error) {

	listParams := &Nip47ListTransactionsParams{}
	err = json.Unmarshal(request.Params, listParams)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId": requestEvent.NostrId,
			"appId":   app.ID,
		}).Errorf("Failed to decode nostr event: %v", err)
		return nil, err
	}

	hasPermission, code, message := svc.hasPermission(app, request.Method, 0)

	if !hasPermission {
		svc.Logger.WithFields(logrus.Fields{
			// TODO: log request fields from listParams
			"eventId": requestEvent.NostrId,
			"appId":   app.ID,
		}).Errorf("App does not have permission: %s %s", code, message)

		return &Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    code,
				Message: message,
			}}, nil
	}

	svc.Logger.WithFields(logrus.Fields{
		// TODO: log request fields from listParams
		"eventId": requestEvent.NostrId,
		"appId":   app.ID,
	}).Info("Fetching transactions")

	transactions, err := svc.lnClient.ListTransactions(ctx, listParams.From, listParams.Until, listParams.Limit, listParams.Offset, listParams.Unpaid, listParams.Type)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			// TODO: log request fields from listParams
			"eventId": requestEvent.NostrId,
			"appId":   app.ID,
		}).Infof("Failed to fetch transactions: %v", err)
		return &Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: fmt.Sprintf("Something went wrong while fetching transactions: %s", err.Error()),
			},
		}, nil
	}

	responsePayload := &Nip47ListTransactionsResponse{
		Transactions: transactions,
	}
	// fmt.Println(responsePayload)

	return &Nip47Response{
		ResultType: request.Method,
		Result:     responsePayload,
	}, nil
}
