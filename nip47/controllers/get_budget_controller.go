package controllers

import (
	"context"
	"errors"

	"github.com/getAlby/go-nostr"
	"github.com/getAlby/hub/db/queries"
	"gorm.io/gorm"

	"github.com/getAlby/hub/constants"
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
	result := controller.db.Where("app_id = ? AND scope = ?", app.ID, constants.PAY_INVOICE_SCOPE).First(&appPermission)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		logger.Logger.WithFields(logrus.Fields{
			"request_event_id": requestEventId,
		}).WithError(result.Error).Error("Failed to fetch pay_invoice permission")
		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error:      mapNip47Error(result.Error),
		}, nostr.Tags{})
		return
	}

	// On ErrRecordNotFound appPermission stays zero-valued and maxAmountSat == 0,
	// which returns the same empty "no budget" response as a permission with no
	// budget set.
	maxAmountSat := appPermission.MaxAmountSat
	if maxAmountSat == 0 {
		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Result:     struct{}{},
		}, nostr.Tags{})
		return
	}

	usedBudgetMsat, err := queries.GetBudgetUsageMsat(controller.db, &appPermission)
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
		TotalBudget:   uint64(maxAmountSat * 1000),
		UsedBudget:    usedBudgetMsat,
		RenewalPeriod: appPermission.BudgetRenewal,
		RenewsAt:      queries.GetBudgetRenewsAt(appPermission.BudgetRenewal),
	}

	publishResponse(&models.Response{
		ResultType: nip47Request.Method,
		Result:     responsePayload,
	}, nostr.Tags{})
}
