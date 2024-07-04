package permissions

import (
	"fmt"
	"slices"
	"time"

	"github.com/getAlby/nostr-wallet-connect/db"
	"github.com/getAlby/nostr-wallet-connect/events"
	"github.com/getAlby/nostr-wallet-connect/lnclient"
	"github.com/getAlby/nostr-wallet-connect/logger"
	"github.com/getAlby/nostr-wallet-connect/nip47/models"
	"github.com/getAlby/nostr-wallet-connect/transactions"
	"github.com/getAlby/nostr-wallet-connect/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	PAY_INVOICE_SCOPE       = "pay_invoice" // also covers pay_keysend and multi_* payment methods
	GET_BALANCE_SCOPE       = "get_balance"
	GET_INFO_SCOPE          = "get_info"
	MAKE_INVOICE_SCOPE      = "make_invoice"
	LOOKUP_INVOICE_SCOPE    = "lookup_invoice"
	LIST_TRANSACTIONS_SCOPE = "list_transactions"
	SIGN_MESSAGE_SCOPE      = "sign_message"
	NOTIFICATIONS_SCOPE     = "notifications" // covers all notification types
)

type permissionsService struct {
	db             *gorm.DB
	eventPublisher events.EventPublisher
}

// TODO: does this need to be a service?
type PermissionsService interface {
	HasPermission(app *db.App, requestMethod string, amount uint64) (result bool, code string, message string)
	GetBudgetUsage(appPermission *db.AppPermission) uint64
	GetPermittedMethods(app *db.App, lnClient lnclient.LNClient) []string
	PermitsNotifications(app *db.App) bool
}

func NewPermissionsService(db *gorm.DB, eventPublisher events.EventPublisher) *permissionsService {
	return &permissionsService{
		db:             db,
		eventPublisher: eventPublisher,
	}
}

func (svc *permissionsService) HasPermission(app *db.App, scope string, amountMsat uint64) (result bool, code string, message string) {

	appPermission := db.AppPermission{}
	findPermissionResult := svc.db.Find(&appPermission, &db.AppPermission{
		AppId: app.ID,
		Scope: scope,
	})
	if findPermissionResult.RowsAffected == 0 {
		// No permission for this request method
		return false, models.ERROR_RESTRICTED, fmt.Sprintf("This app does not have the %s scope", scope)
	}
	expiresAt := appPermission.ExpiresAt
	if expiresAt != nil && expiresAt.Before(time.Now()) {
		logger.Logger.WithFields(logrus.Fields{
			"scope":     scope,
			"expiresAt": expiresAt.Unix(),
			"appId":     app.ID,
			"pubkey":    app.NostrPubkey,
		}).Info("This pubkey is expired")

		return false, models.ERROR_EXPIRED, "This app has expired"
	}

	if scope == PAY_INVOICE_SCOPE {
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
	// TODO: this does not consider unknown fees for pending payments
	svc.db.
		Table("transactions").
		Select("SUM(amount + fee) as sum").
		Where("app_id = ? AND type = ? AND (state = ? OR state = ?) AND created_at > ?", appPermission.AppId, transactions.TRANSACTION_TYPE_OUTGOING, transactions.TRANSACTION_STATE_SETTLED, transactions.TRANSACTION_STATE_PENDING, getStartOfBudget(appPermission.BudgetRenewal)).Scan(&result)
	return result.Sum / 1000
}

func (svc *permissionsService) GetPermittedMethods(app *db.App, lnClient lnclient.LNClient) []string {
	appPermissions := []db.AppPermission{}
	svc.db.Where("app_id = ?", app.ID).Find(&appPermissions)
	scopes := make([]string, 0, len(appPermissions))
	for _, appPermission := range appPermissions {
		scopes = append(scopes, appPermission.Scope)
	}

	requestMethods := scopesToRequestMethods(scopes)

	// only return methods supported by the lnClient
	lnClientSupportedMethods := lnClient.GetSupportedNIP47Methods()
	requestMethods = utils.Filter(requestMethods, func(requestMethod string) bool {
		return slices.Contains(lnClientSupportedMethods, requestMethod)
	})

	return requestMethods
}

func (svc *permissionsService) PermitsNotifications(app *db.App) bool {
	notificationPermission := db.AppPermission{}
	err := svc.db.First(&notificationPermission, &db.AppPermission{
		AppId: app.ID,
		Scope: NOTIFICATIONS_SCOPE,
	}).Error
	if err != nil {
		return false
	}

	return true
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

func scopesToRequestMethods(scopes []string) []string {
	requestMethods := []string{}

	for _, scope := range scopes {
		scopeRequestMethods := scopeToRequestMethods(scope)
		requestMethods = append(requestMethods, scopeRequestMethods...)
	}
	return requestMethods
}

func scopeToRequestMethods(scope string) []string {
	switch scope {
	case PAY_INVOICE_SCOPE:
		return []string{models.PAY_INVOICE_METHOD, models.PAY_KEYSEND_METHOD, models.MULTI_PAY_INVOICE_METHOD, models.MULTI_PAY_KEYSEND_METHOD}
	case GET_BALANCE_SCOPE:
		return []string{models.GET_BALANCE_METHOD}
	case GET_INFO_SCOPE:
		return []string{models.GET_INFO_METHOD}
	case MAKE_INVOICE_SCOPE:
		return []string{models.MAKE_INVOICE_METHOD}
	case LOOKUP_INVOICE_SCOPE:
		return []string{models.LOOKUP_INVOICE_METHOD}
	case LIST_TRANSACTIONS_SCOPE:
		return []string{models.LIST_TRANSACTIONS_METHOD}
	case SIGN_MESSAGE_SCOPE:
		return []string{models.SIGN_MESSAGE_METHOD}
	}
	return []string{}
}

func RequestMethodsToScopes(requestMethods []string) ([]string, error) {
	scopes := []string{}

	for _, requestMethod := range requestMethods {
		scope, err := RequestMethodToScope(requestMethod)
		if err != nil {
			return nil, err
		}
		if !slices.Contains(scopes, scope) {
			scopes = append(scopes, scope)
		}
	}
	return scopes, nil
}

func RequestMethodToScope(requestMethod string) (string, error) {
	switch requestMethod {
	case models.PAY_INVOICE_METHOD, models.PAY_KEYSEND_METHOD, models.MULTI_PAY_INVOICE_METHOD, models.MULTI_PAY_KEYSEND_METHOD:
		return PAY_INVOICE_SCOPE, nil
	case models.GET_BALANCE_METHOD:
		return GET_BALANCE_SCOPE, nil
	case models.GET_INFO_METHOD:
		return GET_INFO_SCOPE, nil
	case models.MAKE_INVOICE_METHOD:
		return MAKE_INVOICE_SCOPE, nil
	case models.LOOKUP_INVOICE_METHOD:
		return LOOKUP_INVOICE_SCOPE, nil
	case models.LIST_TRANSACTIONS_METHOD:
		return LIST_TRANSACTIONS_SCOPE, nil
	case models.SIGN_MESSAGE_METHOD:
		return SIGN_MESSAGE_SCOPE, nil
	}
	logger.Logger.WithField("request_method", requestMethod).Error("Unsupported request method")
	return "", fmt.Errorf("unsupported request method: %s", requestMethod)
}

func AllScopes() []string {
	return []string{
		PAY_INVOICE_SCOPE,
		GET_BALANCE_SCOPE,
		GET_INFO_SCOPE,
		MAKE_INVOICE_SCOPE,
		LOOKUP_INVOICE_SCOPE,
		LIST_TRANSACTIONS_SCOPE,
		SIGN_MESSAGE_SCOPE,
		NOTIFICATIONS_SCOPE,
	}
}
