package main

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

const (
	MSAT_PER_SAT = 1000
)

func (svc *Service) HandleGetBalanceEvent(ctx context.Context, request *Nip47Request, requestEvent *NostrEvent, app *App) (result *Nip47Response, err error) {

	hasPermission, code, message := svc.hasPermission(app, requestEvent, request.Method, 0)

	if !hasPermission {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   requestEvent.NostrId,
			"eventKind": requestEvent.Kind,
			"appId":     app.ID,
		}).Errorf("App does not have permission: %s %s", code, message)

		return &Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    code,
				Message: message,
			}}, nil
	}

	svc.Logger.WithFields(logrus.Fields{
		"eventId":   requestEvent.NostrId,
		"eventKind": requestEvent.Kind,
		"appId":     app.ID,
	}).Info("Fetching balance")

	balance, err := svc.lnClient.GetBalance(ctx, requestEvent.PubKey)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   requestEvent.NostrId,
			"eventKind": requestEvent.Kind,
			"appId":     app.ID,
		}).Infof("Failed to fetch balance: %v", err)
		return &Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: fmt.Sprintf("Something went wrong while fetching balance: %s", err.Error()),
			},
		}, nil
	}

	responsePayload := &Nip47BalanceResponse{
		Balance: balance * MSAT_PER_SAT,
	}

	appPermission := AppPermission{}
	svc.db.Where("app_id = ? AND request_method = ?", app.ID, NIP_47_PAY_INVOICE_METHOD).First(&appPermission)

	maxAmount := appPermission.MaxAmount
	if maxAmount > 0 {
		responsePayload.MaxAmount = maxAmount * MSAT_PER_SAT
		responsePayload.BudgetRenewal = appPermission.BudgetRenewal
	}

	return &Nip47Response{
		ResultType: request.Method,
		Result:     responsePayload,
	}, nil
}
