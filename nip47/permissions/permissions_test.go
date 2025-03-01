package permissions

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/tests"
)

func TestHasPermission_NoPermission(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	permissionsSvc := NewPermissionsService(svc.DB, svc.EventPublisher)
	result, code, message := permissionsSvc.HasPermission(app, constants.PAY_INVOICE_SCOPE)
	assert.False(t, result)
	assert.Equal(t, constants.ERROR_RESTRICTED, code)
	assert.Equal(t, "This app does not have the pay_invoice scope", message)
}

func TestHasPermission_Expired(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	app, _, err := tests.CreateApp(svc)
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

func TestHasPermission_OK(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	app, _, err := tests.CreateApp(svc)
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
	svc, err := tests.CreateTestService(t)
	assert.NoError(t, err)
	defer svc.Remove()

	scope, err := RequestMethodToScope(models.GET_BUDGET_METHOD)
	assert.NoError(t, err)
	assert.Equal(t, "", scope)
}

func TestRequestMethodsToScopes_GetBudget(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	assert.NoError(t, err)
	defer svc.Remove()

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

func TestRequestMethodToScope_CreateConnection(t *testing.T) {
	scope, err := RequestMethodToScope(models.CREATE_CONNECTION_METHOD)
	assert.NoError(t, err)
	assert.Equal(t, constants.SUPERUSER_SCOPE, scope)
}
func TestScopeToRequestMethods_Superuser(t *testing.T) {
	methods := scopeToRequestMethods(constants.SUPERUSER_SCOPE)
	assert.Equal(t, []string{models.CREATE_CONNECTION_METHOD}, methods)
}

func TestGetPermittedMethods_AlwaysGranted(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)

	permissionsSvc := NewPermissionsService(svc.DB, svc.EventPublisher)
	result := permissionsSvc.GetPermittedMethods(app, svc.LNClient)
	assert.Equal(t, GetAlwaysGrantedMethods(), result)
}

func TestGetPermittedMethods_PayInvoiceScopeGivesAllPaymentMethods(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	app, _, err := tests.CreateApp(svc)
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
