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
    Balance int64 `json:"balance"`
    // MaxAmount     int    `json:"max_amount"`
    // BudgetRenewal string `json:"budget_renewal"`
}

// TODO: remove checkPermission - can it be a middleware?
func (controller *nip47Controller) HandleGetBalanceEvent(ctx context.Context, nip47Request *models.Request, requestEventId uint, app *db.App, publishResponse publishFunc) {

    logger.Logger.WithFields(logrus.Fields{
        "request_event_id": requestEventId,
    }).Debug("Getting balance")

    balance := int64(0)
    if app.Isolated {
        balance = queries.GetIsolatedBalance(controller.db, app.ID)
    } else {
        balances, err := controller.lnClient.GetBalances(ctx, true)
        balance = balances.Lightning.TotalSpendable
        if err != nil {
            logger.Logger.WithFields(logrus.Fields{
                "request_event_id": requestEventId,
            }).WithError(err).Error("Failed to fetch balance")
            publishResponse(&models.Response{
                ResultType: nip47Request.Method,
                Error:      mapNip47Error(err),
            }, nostr.Tags{})
            return
        }

        // Check if there's a budget set for pay_invoice (which applies to spending)
        var appPermission db.AppPermission
        err = controller.db.Where("app_id = ? AND scope = ?", app.ID, models.PAY_INVOICE_METHOD).First(&appPermission).Error
        if err == nil && appPermission.MaxAmountSat > 0 {
            // Calculate available budget
            totalBudgetMsat := uint64(appPermission.MaxAmountSat) * MSAT_PER_SAT
            usedBudgetMsat := queries.GetBudgetUsageSat(controller.db, &appPermission) * MSAT_PER_SAT
            availableBudgetMsat := totalBudgetMsat - usedBudgetMsat
            // Show the minimum of actual balance and available budget
            if availableBudgetMsat < uint64(balance) {
                balance = int64(availableBudgetMsat)
            }
        }
    }

    logger.Logger.WithFields(logrus.Fields{"balance": balance}).Info("Returning balance")
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