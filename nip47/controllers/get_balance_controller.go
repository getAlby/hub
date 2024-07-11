package controllers

import (
	"context"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/db/queries"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

const (
	MSAT_PER_SAT = 1000
)

type getBalanceResponse struct {
	Balance uint64 `json:"balance"`
	// MaxAmount     int    `json:"max_amount"`
	// BudgetRenewal string `json:"budget_renewal"`
}

// TODO: remove checkPermission - can it be a middleware?
func (controller *nip47Controller) HandleGetBalanceEvent(ctx context.Context, nip47Request *models.Request, requestEventId uint, appId uint, checkPermission checkPermissionFunc, publishResponse publishFunc) {
	// basic permissions check
	resp := checkPermission(0)
	if resp != nil {
		publishResponse(resp, nostr.Tags{})
		return
	}

	logger.Logger.WithFields(logrus.Fields{
		"request_event_id": requestEventId,
	}).Info("Getting balance")

	// TODO: optimize
	var appPermission db.AppPermission
	controller.db.Find(&appPermission, &db.AppPermission{
		AppId: appId,
	})
	balance := uint64(0)
	if appPermission.BalanceType == "isolated" {
		balance = queries.GetIsolatedBalance(controller.db, appPermission.AppId)
	} else {
		balance_signed, err := controller.lnClient.GetBalance(ctx)
		balance = uint64(balance_signed)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"request_event_id": requestEventId,
			}).WithError(err).Error("Failed to fetch balance")
			publishResponse(&models.Response{
				ResultType: nip47Request.Method,
				Error: &models.Error{
					Code:    models.ERROR_INTERNAL,
					Message: err.Error(),
				},
			}, nostr.Tags{})
			return
		}
	}

	responsePayload := &getBalanceResponse{
		Balance: balance,
	}

	// this is not part of the spec and does not seem to be used
	/*appPermission := db.AppPermission{}
	controller.db.Where("app_id = ? AND request_method = ?", app.ID, models.PAY_INVOICE_METHOD).First(&appPermission)

	maxAmount := appPermission.MaxAmount
	if maxAmount > 0 {
		responsePayload.MaxAmount = maxAmount * MSAT_PER_SAT
		responsePayload.BudgetRenewal = appPermission.BudgetRenewal
	}*/

	publishResponse(&models.Response{
		ResultType: nip47Request.Method,
		Result:     responsePayload,
	}, nostr.Tags{})
}
