package permissions

import (
	"fmt"
	"slices"
	"time"

	"github.com/getAlby/nostr-wallet-connect/db"
	"github.com/getAlby/nostr-wallet-connect/events"
	"github.com/getAlby/nostr-wallet-connect/logger"
	"github.com/getAlby/nostr-wallet-connect/nip47/models"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// TODO: move other permissions here (e.g. all payment methods use pay_invoice)
const (
	NOTIFICATIONS_PERMISSION = "notifications"
)

type permissionsService struct {
	db             *gorm.DB
	eventPublisher events.EventPublisher
}

// TODO: does this need to be a service?
type PermissionsService interface {
	HasPermission(app *db.App, requestMethod string, amount uint64) (result bool, code string, message string)
	GetBudgetUsage(appPermission *db.AppPermission) uint64
	GetPermittedMethods(app *db.App) []string
}

func NewPermissionsService(db *gorm.DB, eventPublisher events.EventPublisher) *permissionsService {
	return &permissionsService{
		db:             db,
		eventPublisher: eventPublisher,
	}
}

func (svc *permissionsService) HasPermission(app *db.App, requestMethod string, amountMsat uint64) (result bool, code string, message string) {
	switch requestMethod {
	case models.PAY_INVOICE_METHOD, models.PAY_KEYSEND_METHOD, models.MULTI_PAY_INVOICE_METHOD, models.MULTI_PAY_KEYSEND_METHOD:
		requestMethod = models.PAY_INVOICE_METHOD
	}

	appPermission := db.AppPermission{}
	findPermissionResult := svc.db.Find(&appPermission, &db.AppPermission{
		AppId:         app.ID,
		RequestMethod: requestMethod,
	})
	if findPermissionResult.RowsAffected == 0 {
		// No permission for this request method
		return false, models.ERROR_RESTRICTED, fmt.Sprintf("This app does not have permission to request %s", requestMethod)
	}
	expiresAt := appPermission.ExpiresAt
	if expiresAt != nil && expiresAt.Before(time.Now()) {
		logger.Logger.WithFields(logrus.Fields{
			"requestMethod": requestMethod,
			"expiresAt":     expiresAt.Unix(),
			"appId":         app.ID,
			"pubkey":        app.NostrPubkey,
		}).Info("This pubkey is expired")

		return false, models.ERROR_EXPIRED, "This app has expired"
	}

	if requestMethod == models.PAY_INVOICE_METHOD {
		maxAmount := appPermission.MaxAmount
		if maxAmount != 0 {
			budgetUsage := svc.GetBudgetUsage(&appPermission)

			if budgetUsage+amountMsat/1000 > uint64(maxAmount) {
				return false, models.ERROR_QUOTA_EXCEEDED, "Insufficient budget remaining to make payment"
			}
		}
	}
	return true, "", ""
}

func (svc *permissionsService) GetBudgetUsage(appPermission *db.AppPermission) uint64 {
	var result struct {
		Sum uint64
	}
	// TODO: discard failed payments from this check instead of checking payments that have a preimage
	svc.db.Table("payments").Select("SUM(amount) as sum").Where("app_id = ? AND preimage IS NOT NULL AND created_at > ?", appPermission.AppId, getStartOfBudget(appPermission.BudgetRenewal)).Scan(&result)
	return result.Sum
}

func (svc *permissionsService) GetPermittedMethods(app *db.App) []string {
	appPermissions := []db.AppPermission{}
	svc.db.Find(&appPermissions, &db.AppPermission{
		AppId: app.ID,
	})
	requestMethods := make([]string, 0, len(appPermissions))
	for _, appPermission := range appPermissions {
		requestMethods = append(requestMethods, appPermission.RequestMethod)
	}
	if slices.Contains(requestMethods, models.PAY_INVOICE_METHOD) {
		// all payment methods are tied to the pay_invoice permission
		requestMethods = append(requestMethods, models.PAY_KEYSEND_METHOD, models.MULTI_PAY_INVOICE_METHOD, models.MULTI_PAY_KEYSEND_METHOD)
	}

	return requestMethods
}

func getStartOfBudget(budget_type string) time.Time {
	now := time.Now()
	switch budget_type {
	case models.BUDGET_RENEWAL_DAILY:
		// TODO: Use the location of the user, instead of the server
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case models.BUDGET_RENEWAL_WEEKLY:
		weekday := now.Weekday()
		var startOfWeek time.Time
		if weekday == 0 {
			startOfWeek = now.AddDate(0, 0, -6)
		} else {
			startOfWeek = now.AddDate(0, 0, -int(weekday)+1)
		}
		return time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())
	case models.BUDGET_RENEWAL_MONTHLY:
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	case models.BUDGET_RENEWAL_YEARLY:
		return time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location())
	default: //"never"
		return time.Time{}
	}
}
