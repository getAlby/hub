package permissions

import (
	"fmt"
	"slices"
	"time"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type permissionsService struct {
	db             *gorm.DB
	eventPublisher events.EventPublisher
}

// TODO: does this need to be a service?
type PermissionsService interface {
	HasPermission(app *db.App, requestMethod string) (result bool, code string, message string)
	GetPermittedMethods(app *db.App, lnClient lnclient.LNClient) []string
	PermitsNotifications(app *db.App) bool
}

func NewPermissionsService(db *gorm.DB, eventPublisher events.EventPublisher) *permissionsService {
	return &permissionsService{
		db:             db,
		eventPublisher: eventPublisher,
	}
}

func (svc *permissionsService) HasPermission(app *db.App, scope string) (result bool, code string, message string) {
	appPermission := db.AppPermission{}
	findPermissionResult := svc.db.Limit(1).Find(&appPermission, &db.AppPermission{
		AppId: app.ID,
		Scope: scope,
	})
	if findPermissionResult.RowsAffected == 0 {
		// No permission for this request method
		return false, constants.ERROR_RESTRICTED, fmt.Sprintf("This app does not have the %s scope", scope)
	}
	expiresAt := appPermission.ExpiresAt
	if expiresAt != nil && expiresAt.Before(time.Now()) {
		logger.Logger.WithFields(logrus.Fields{
			"scope":     scope,
			"expiresAt": expiresAt.Unix(),
			"appId":     app.ID,
			"pubkey":    app.AppPubkey,
		}).Info("This pubkey is expired")

		return false, constants.ERROR_EXPIRED, "This app has expired"
	}

	return true, "", ""
}

func (svc *permissionsService) GetPermittedMethods(app *db.App, lnClient lnclient.LNClient) []string {
	appPermissions := []db.AppPermission{}
	svc.db.Where("app_id = ?", app.ID).Find(&appPermissions)
	scopes := make([]string, 0, len(appPermissions))
	for _, appPermission := range appPermissions {
		scopes = append(scopes, appPermission.Scope)
	}

	requestMethods := scopesToRequestMethods(scopes)

	for _, method := range GetAlwaysGrantedMethods() {
		if !slices.Contains(requestMethods, method) {
			requestMethods = append(requestMethods, method)
		}
	}

	// only return methods supported by the lnClient
	lnClientSupportedMethods := lnClient.GetSupportedNIP47Methods()
	requestMethods = utils.Filter(requestMethods, func(requestMethod string) bool {
		// TODO: better way to exclude methods unrelated to the lnclient
		if requestMethod == models.CREATE_CONNECTION_METHOD {
			return true
		}

		return slices.Contains(lnClientSupportedMethods, requestMethod)
	})

	return requestMethods
}

func (svc *permissionsService) PermitsNotifications(app *db.App) bool {
	notificationPermission := db.AppPermission{}
	result := svc.db.Limit(1).Find(&notificationPermission, &db.AppPermission{
		AppId: app.ID,
		Scope: constants.NOTIFICATIONS_SCOPE,
	})

	return result.Error == nil && result.RowsAffected > 0
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
	case constants.PAY_INVOICE_SCOPE:
		return []string{models.PAY_INVOICE_METHOD, models.PAY_KEYSEND_METHOD, models.MULTI_PAY_INVOICE_METHOD, models.MULTI_PAY_KEYSEND_METHOD}
	case constants.GET_BALANCE_SCOPE:
		return []string{models.GET_BALANCE_METHOD}
	case constants.GET_INFO_SCOPE:
		return []string{models.GET_INFO_METHOD}
	case constants.MAKE_INVOICE_SCOPE:
		return []string{models.MAKE_INVOICE_METHOD, models.MAKE_HOLD_INVOICE_METHOD, models.SETTLE_HOLD_INVOICE_METHOD, models.CANCEL_HOLD_INVOICE_METHOD}
	case constants.LOOKUP_INVOICE_SCOPE:
		return []string{models.LOOKUP_INVOICE_METHOD}
	case constants.LIST_TRANSACTIONS_SCOPE:
		return []string{models.LIST_TRANSACTIONS_METHOD}
	case constants.SIGN_MESSAGE_SCOPE:
		return []string{models.SIGN_MESSAGE_METHOD}
	case constants.SUPERUSER_SCOPE:
		return []string{models.CREATE_CONNECTION_METHOD}
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
		if scope != "" && !slices.Contains(scopes, scope) {
			scopes = append(scopes, scope)
		}
	}
	return scopes, nil
}

func RequestMethodToScope(requestMethod string) (string, error) {
	switch requestMethod {
	case models.PAY_INVOICE_METHOD, models.PAY_KEYSEND_METHOD, models.MULTI_PAY_INVOICE_METHOD, models.MULTI_PAY_KEYSEND_METHOD:
		return constants.PAY_INVOICE_SCOPE, nil
	case models.GET_BALANCE_METHOD:
		return constants.GET_BALANCE_SCOPE, nil
	case models.GET_BUDGET_METHOD:
		return "", nil
	case models.GET_INFO_METHOD:
		return constants.GET_INFO_SCOPE, nil
	case models.MAKE_INVOICE_METHOD:
		return constants.MAKE_INVOICE_SCOPE, nil
	case models.LOOKUP_INVOICE_METHOD:
		return constants.LOOKUP_INVOICE_SCOPE, nil
	case models.LIST_TRANSACTIONS_METHOD:
		return constants.LIST_TRANSACTIONS_SCOPE, nil
	case models.SIGN_MESSAGE_METHOD:
		return constants.SIGN_MESSAGE_SCOPE, nil
	case models.MAKE_HOLD_INVOICE_METHOD, models.SETTLE_HOLD_INVOICE_METHOD, models.CANCEL_HOLD_INVOICE_METHOD:
		return constants.MAKE_INVOICE_SCOPE, nil
	case models.CREATE_CONNECTION_METHOD:
		return constants.SUPERUSER_SCOPE, nil
	}
	logger.Logger.WithField("request_method", requestMethod).Error("Unsupported request method")
	return "", fmt.Errorf("unsupported request method: %s", requestMethod)
}

func AllScopes() []string {
	return []string{
		constants.PAY_INVOICE_SCOPE,
		constants.GET_BALANCE_SCOPE,
		constants.GET_INFO_SCOPE,
		constants.MAKE_INVOICE_SCOPE,
		constants.LOOKUP_INVOICE_SCOPE,
		constants.LIST_TRANSACTIONS_SCOPE,
		constants.SIGN_MESSAGE_SCOPE,
		constants.NOTIFICATIONS_SCOPE,
		constants.SUPERUSER_SCOPE,
	}
}

func GetAlwaysGrantedMethods() []string {
	return []string{models.GET_INFO_METHOD, models.GET_BUDGET_METHOD}
}
