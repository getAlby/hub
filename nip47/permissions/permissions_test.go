package permissions

import (
	"testing"
	"time"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasPermission_NoPermission(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	app, _, err := tests.CreateApp(svc, "1.0")
	assert.NoError(t, err)

	permissionsSvc := NewPermissionsService(svc.DB, svc.EventPublisher)
	result, code, message := permissionsSvc.HasPermission(app, constants.PAY_INVOICE_SCOPE)
	assert.False(t, result)
	assert.Equal(t, constants.ERROR_RESTRICTED, code)
	assert.Equal(t, "This app does not have the pay_invoice scope", message)
}

func TestHasPermission_Expired(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	app, _, err := tests.CreateApp(svc, "1.0")
	assert.NoError(t, err)

	budgetRenewal := "never"
	expiresAt := time.Now().Add(-24 * time.Hour)
	appPermission := &db.AppPermission{
		AppId:         app.ID,
		App:           *app,
		Scope:         constants.PAY_INVOICE_SCOPE,
		MaxAmountSat:  100,
		BudgetRenewal: budgetRenewal,
		ExpiresAt:     &expiresAt,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	permissionsSvc := NewPermissionsService(svc.DB, svc.EventPublisher)
	result, code, message := permissionsSvc.HasPermission(app, constants.PAY_INVOICE_SCOPE)
	assert.False(t, result)
	assert.Equal(t, constants.ERROR_EXPIRED, code)
	assert.Equal(t, "This app has expired", message)
}

// TODO: move to transactions service
/*func TestHasPermission_Exceeded(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	app, _, err := tests.CreateApp(svc, "1.0")
	assert.NoError(t, err)

	budgetRenewal := "never"
	expiresAt := time.Now().Add(24 * time.Hour)
	appPermission := &db.AppPermission{
		AppId:         app.ID,
		App:           *app,
		Scope:         constants.PAY_INVOICE_SCOPE,
		MaxAmountSat:  10,
		BudgetRenewal: budgetRenewal,
		ExpiresAt:     &expiresAt,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	permissionsSvc := NewPermissionsService(svc.DB, svc.EventPublisher)
	result, code, message := permissionsSvc.HasPermission(app, PAY_INVOICE_SCOPE, 100*1000)
	assert.False(t, result)
	assert.Equal(t, constants.ERROR_QUOTA_EXCEEDED, code)
	assert.Equal(t, "Insufficient budget remaining to make payment", message)
}*/

func TestHasPermission_OK(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	app, _, err := tests.CreateApp(svc, "1.0")
	assert.NoError(t, err)

	budgetRenewal := "never"
	expiresAt := time.Now().Add(24 * time.Hour)
	appPermission := &db.AppPermission{
		AppId:         app.ID,
		App:           *app,
		Scope:         constants.PAY_INVOICE_SCOPE,
		MaxAmountSat:  10,
		BudgetRenewal: budgetRenewal,
		ExpiresAt:     &expiresAt,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	permissionsSvc := NewPermissionsService(svc.DB, svc.EventPublisher)
	result, code, message := permissionsSvc.HasPermission(app, constants.PAY_INVOICE_SCOPE)
	assert.True(t, result)
	assert.Empty(t, code)
	assert.Empty(t, message)
}

func TestRequestMethodToScope_GetBudget(t *testing.T) {
	defer tests.RemoveTestService()
	_, err := tests.CreateTestService()
	assert.NoError(t, err)

	scope, err := RequestMethodToScope(models.GET_BUDGET_METHOD)
	assert.NoError(t, err)
	assert.Equal(t, "", scope)
}

func TestRequestMethodsToScopes_GetBudget(t *testing.T) {
	defer tests.RemoveTestService()
	_, err := tests.CreateTestService()
	assert.NoError(t, err)

	scopes, err := RequestMethodsToScopes([]string{models.GET_BUDGET_METHOD})
	assert.NoError(t, err)
	assert.Equal(t, []string{}, scopes)
}

func TestRequestMethodToScope_GetInfo(t *testing.T) {
	scope, err := RequestMethodToScope(models.GET_INFO_METHOD)
	assert.NoError(t, err)
	assert.Equal(t, constants.GET_INFO_SCOPE, scope)
}

func TestRequestMethodsToScopes_GetInfo(t *testing.T) {
	scopes, err := RequestMethodsToScopes([]string{models.GET_INFO_METHOD})
	assert.NoError(t, err)
	assert.Equal(t, []string{constants.GET_INFO_SCOPE}, scopes)
}

func TestGetPermittedMethods_AlwaysGranted(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	app, _, err := tests.CreateApp(svc, "1.0")
	assert.NoError(t, err)

	permissionsSvc := NewPermissionsService(svc.DB, svc.EventPublisher)
	result := permissionsSvc.GetPermittedMethods(app, svc.LNClient)
	assert.Equal(t, GetAlwaysGrantedMethods(), result)
}

func TestGetPermittedMethods_PayInvoiceScopeGivesAllPaymentMethods(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	require.NoError(t, err)

	app, _, err := tests.CreateApp(svc, "1.0")
	assert.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId: app.ID,
		App:   *app,
		Scope: constants.PAY_INVOICE_SCOPE,
	}
	err = svc.DB.Create(appPermission).Error
	assert.NoError(t, err)

	permissionsSvc := NewPermissionsService(svc.DB, svc.EventPublisher)
	result := permissionsSvc.GetPermittedMethods(app, svc.LNClient)
	assert.Contains(t, result, models.PAY_INVOICE_METHOD)
	assert.Contains(t, result, models.PAY_KEYSEND_METHOD)
	assert.Contains(t, result, models.MULTI_PAY_INVOICE_METHOD)
	assert.Contains(t, result, models.MULTI_PAY_KEYSEND_METHOD)
}
