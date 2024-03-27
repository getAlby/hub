package main

import (
	"context"

	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

const (
	MSAT_PER_SAT = 1000
)

func (svc *Service) HandleGetBalanceEvent(ctx context.Context, request *Nip47Request, requestEvent *RequestEvent, app *App, publishResponse func(*Nip47Response, *nostr.Tags)) {

	resp := svc.checkPermission(request, requestEvent, app, 0)
	if resp != nil {
		publishResponse(resp, &nostr.Tags{})
		return
	}

	svc.Logger.WithFields(logrus.Fields{
		"eventId": requestEvent.NostrId,
		"appId":   app.ID,
	}).Info("Fetching balance")

	balance, err := svc.lnClient.GetBalance(ctx)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId": requestEvent.NostrId,
			"appId":   app.ID,
		}).Infof("Failed to fetch balance: %v", err)
		publishResponse(&Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: err.Error(),
			},
		}, &nostr.Tags{})
		return
	}

	responsePayload := &Nip47BalanceResponse{
		Balance: balance,
	}

	appPermission := AppPermission{}
	svc.db.Where("app_id = ? AND request_method = ?", app.ID, NIP_47_PAY_INVOICE_METHOD).First(&appPermission)

	maxAmount := appPermission.MaxAmount
	if maxAmount > 0 {
		responsePayload.MaxAmount = maxAmount * MSAT_PER_SAT
		responsePayload.BudgetRenewal = appPermission.BudgetRenewal
	}

	publishResponse(&Nip47Response{
		ResultType: request.Method,
		Result:     responsePayload,
	}, &nostr.Tags{})
}
