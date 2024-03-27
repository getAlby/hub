package main

import (
	"context"

	"github.com/sirupsen/logrus"
)

func (svc *Service) HandleListTransactionsEvent(ctx context.Context, request *Nip47Request, requestEvent *RequestEvent, app *App) *Nip47Response {

	listParams := &Nip47ListTransactionsParams{}
	resp := svc.unmarshalRequest(request, requestEvent, app, listParams)
	if resp != nil {
		return resp
	}

	resp = svc.checkPermission(request, requestEvent, app, 0)
	if resp != nil {
		return resp
	}

	svc.Logger.WithFields(logrus.Fields{
		// TODO: log request fields from listParams
		"eventId": requestEvent.NostrId,
		"appId":   app.ID,
	}).Info("Fetching transactions")

	limit := listParams.Limit
	maxLimit := uint64(50)
	if limit == 0 || limit > maxLimit {
		// make sure a sensible limit is passed
		limit = maxLimit
	}
	transactions, err := svc.lnClient.ListTransactions(ctx, listParams.From, listParams.Until, limit, listParams.Offset, listParams.Unpaid, listParams.Type)
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
				Message: err.Error(),
			},
		}
	}

	responsePayload := &Nip47ListTransactionsResponse{
		Transactions: transactions,
	}
	// fmt.Println(responsePayload)

	return &Nip47Response{
		ResultType: request.Method,
		Result:     responsePayload,
	}
}
