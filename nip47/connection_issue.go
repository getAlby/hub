package nip47

import (
	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	"github.com/sirupsen/logrus"
)

const (
	connectionIssueMissingPermission = "missing_permission"
	connectionIssueUnknownMethod     = "unknown_method"
	connectionIssueExpiredConnection = "expired_connection"
	connectionIssueBudgetExceeded    = "budget_exceeded"
	connectionIssueLowBalance        = "low_balance"
	connectionIssuePaymentFailed     = "payment_failed"
)

func categorizeConnectionIssue(method string, responseError *models.Error) (string, bool) {
	if responseError.Code == constants.ERROR_RESTRICTED {
		return connectionIssueMissingPermission, true
	}
	if responseError.Code == constants.ERROR_EXPIRED {
		return connectionIssueExpiredConnection, true
	}
	if responseError.Code == constants.ERROR_QUOTA_EXCEEDED {
		return connectionIssueBudgetExceeded, true
	}
	if responseError.Code == constants.ERROR_INSUFFICIENT_BALANCE {
		return connectionIssueLowBalance, true
	}
	if responseError.Code == constants.ERROR_NOT_IMPLEMENTED {
		return connectionIssueUnknownMethod, true
	}
	if method == models.PAY_INVOICE_METHOD ||
		method == models.MULTI_PAY_INVOICE_METHOD ||
		method == models.PAY_KEYSEND_METHOD ||
		method == models.MULTI_PAY_KEYSEND_METHOD {
		return connectionIssuePaymentFailed, true
	}

	return "", false
}

func (svc *nip47Service) recordConnectionIssue(app *db.App, requestEvent *db.RequestEvent, response *models.Response) {
	if app == nil || response == nil || response.Error == nil {
		return
	}

	category, ok := categorizeConnectionIssue(requestEvent.Method, response.Error)
	if !ok {
		return
	}

	issue := db.ConnectionIssue{
		AppId:          app.ID,
		RequestEventId: requestEvent.ID,
		Method:         requestEvent.Method,
		Category:       category,
		ErrorCode:      response.Error.Code,
		ErrorMessage:   response.Error.Message,
	}

	err := svc.db.Create(&issue).Error
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"appId":          app.ID,
			"requestEventId": requestEvent.ID,
			"method":         requestEvent.Method,
		}).WithError(err).Error("Failed to record connection issue")
	}
}
