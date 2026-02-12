package controllers

import (
	"context"

	"github.com/getAlby/hub/db/queries"
	"github.com/nbd-wtf/go-nostr"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	"github.com/sirupsen/logrus"
)

type getBudgetResponse struct {
	UsedBudget    uint64  `json:"used_budget"`
	TotalBudget   uint64  `json:"total_budget"`
	RenewsAt      *uint64 `json:"renews_at,omitempty"`
	RenewalPeriod string  `json:"renewal_period"`
}

func (controller *nip47Controller) HandleGetBudgetEvent(ctx context.Context, nip47Request *models.Request, requestEventId uint, app *db.App, publishResponse publishFunc) {

	logger.Logger.WithFields(logrus.Fields{
		"request_event_id": requestEventId,
	}).Debug("Getting budget")

	appPermission := db.AppPermission{}
	controller.db.Where("app_id = ? AND scope = ?", app.ID, models.PAY_INVOICE_METHOD).First(&appPermission)

	maxAmount := appPermission.MaxAmountSat
	if maxAmount == 0 {
		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Result:     struct{}{},
		}, nostr.Tags{})
		return
	}

	usedBudget, err := queries.GetBudgetUsageSat(controller.db, &appPermission)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"request_event_id": requestEventId,
		}).WithError(err).Error("Failed to fetch budget usage")
		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error:      mapNip47Error(err),
		}, nostr.Tags{})
		return
	}

	responsePayload := &getBudgetResponse{
		TotalBudget:   uint64(maxAmount * 1000),
		UsedBudget:    usedBudget * 1000,
		RenewalPeriod: appPermission.BudgetRenewal,
		RenewsAt:      queries.GetBudgetRenewsAt(appPermission.BudgetRenewal),
	}

	publishResponse(&models.Response{
		ResultType: nip47Request.Method,
		Result:     responsePayload,
	}, nostr.Tags{})
}
