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

func TestRequestMethodToScope_UnsupportedMethod(t *testing.T) {
	_, err := RequestMethodToScope("unknown_method")
	assert.ErrorIs(t, err, ErrUnsupportedMethod)

	scope, err := RequestMethodToScope(models.PAY_INVOICE_METHOD)
	require.NoError(t, err)
	assert.Equal(t, constants.PAY_INVOICE_SCOPE, scope)
}

func TestScopeToRequestMethods_MakeInvoice(t *testing.T) {
	// the bundled make_invoice scope keeps covering the hold invoice methods
	methods := scopeToRequestMethods(constants.MAKE_INVOICE_SCOPE)
	assert.Equal(t, []string{models.MAKE_INVOICE_METHOD, models.MAKE_HOLD_INVOICE_METHOD, models.SETTLE_HOLD_INVOICE_METHOD, models.CANCEL_HOLD_INVOICE_METHOD}, methods)
}

func TestScopeToRequestMethods_MakeInvoiceOnly(t *testing.T) {
	methods := scopeToRequestMethods(constants.MAKE_INVOICE_ONLY_SCOPE)
	assert.Equal(t, []string{models.MAKE_INVOICE_METHOD}, methods)
}

func TestRequestMethodToScope_InvoiceMethods(t *testing.T) {
	// unchanged: every invoice method maps to the bundled make_invoice scope
	for _, method := range []string{models.MAKE_INVOICE_METHOD, models.MAKE_HOLD_INVOICE_METHOD, models.SETTLE_HOLD_INVOICE_METHOD, models.CANCEL_HOLD_INVOICE_METHOD} {
		scope, err := RequestMethodToScope(method)
		require.NoError(t, err)
		assert.Equal(t, constants.MAKE_INVOICE_SCOPE, scope, method)
	}

	scopes, err := RequestMethodsToScopes([]string{models.MAKE_INVOICE_METHOD, models.LOOKUP_INVOICE_METHOD})
	require.NoError(t, err)
	assert.Equal(t, []string{constants.MAKE_INVOICE_SCOPE, constants.LOOKUP_INVOICE_SCOPE}, scopes)
}

func TestRequestMethodsToScopesExact(t *testing.T) {
	testCases := []struct {
		name           string
		requestMethods []string
		expectedScopes []string
	}{
		{
			name:           "make_invoice without hold invoice methods",
			requestMethods: []string{models.MAKE_INVOICE_METHOD, models.LOOKUP_INVOICE_METHOD},
			expectedScopes: []string{constants.MAKE_INVOICE_ONLY_SCOPE, constants.LOOKUP_INVOICE_SCOPE},
		},
		{
			name:           "make_invoice with pay_invoice",
			requestMethods: []string{models.GET_INFO_METHOD, models.MAKE_INVOICE_METHOD, models.PAY_INVOICE_METHOD},
			expectedScopes: []string{constants.GET_INFO_SCOPE, constants.MAKE_INVOICE_ONLY_SCOPE, constants.PAY_INVOICE_SCOPE},
		},
		{
			name:           "make_invoice with make_hold_invoice",
			requestMethods: []string{models.MAKE_INVOICE_METHOD, models.MAKE_HOLD_INVOICE_METHOD},
			expectedScopes: []string{constants.MAKE_INVOICE_SCOPE},
		},
		{
			name:           "make_invoice with all hold invoice methods",
			requestMethods: []string{models.MAKE_INVOICE_METHOD, models.MAKE_HOLD_INVOICE_METHOD, models.SETTLE_HOLD_INVOICE_METHOD, models.CANCEL_HOLD_INVOICE_METHOD, models.LOOKUP_INVOICE_METHOD},
			expectedScopes: []string{constants.MAKE_INVOICE_SCOPE, constants.LOOKUP_INVOICE_SCOPE},
		},
		{
			name:           "only settle_hold_invoice",
			requestMethods: []string{models.SETTLE_HOLD_INVOICE_METHOD},
			expectedScopes: []string{constants.MAKE_INVOICE_SCOPE},
		},
		{
			name:           "only cancel_hold_invoice",
			requestMethods: []string{models.CANCEL_HOLD_INVOICE_METHOD},
			expectedScopes: []string{constants.MAKE_INVOICE_SCOPE},
		},
		{
			name:           "no invoice methods",
			requestMethods: []string{models.GET_INFO_METHOD, models.GET_BALANCE_METHOD, models.GET_BUDGET_METHOD},
			expectedScopes: []string{constants.GET_INFO_SCOPE, constants.GET_BALANCE_SCOPE},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scopes, err := RequestMethodsToScopesExact(tc.requestMethods)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedScopes, scopes)

			// the bundled mapping is unchanged
			bundledScopes, err := RequestMethodsToScopes(tc.requestMethods)
			require.NoError(t, err)
			assert.NotContains(t, bundledScopes, constants.MAKE_INVOICE_ONLY_SCOPE)
		})
	}
}

func TestRequestMethodsToScopesExact_UnsupportedMethod(t *testing.T) {
	_, err := RequestMethodsToScopesExact([]string{models.MAKE_INVOICE_METHOD, "unknown_method"})
	assert.ErrorIs(t, err, ErrUnsupportedMethod)
}

func TestAllScopes_MakeInvoiceOnly(t *testing.T) {
	assert.Contains(t, AllScopes(), constants.MAKE_INVOICE_ONLY_SCOPE)
	assert.Contains(t, AllScopes(), constants.MAKE_INVOICE_SCOPE)
	assert.Equal(t, 10, len(AllScopes()))
}

func TestHasPermission_MakeInvoiceOnlyDoesNotGrantMakeInvoiceScope(t *testing.T) {
	// Older Hub versions only know the make_invoice scope. An app that only has the
	// make_invoice_only scope must not be granted the make_invoice scope (fail closed).
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	app, _, err := tests.CreateApp(svc)
	require.NoError(t, err)

	err = svc.DB.Create(&db.AppPermission{AppId: app.ID, App: *app, Scope: constants.MAKE_INVOICE_ONLY_SCOPE}).Error
	require.NoError(t, err)

	permissionsSvc := NewPermissionsService(svc.DB, svc.EventPublisher)
	result, code, message := permissionsSvc.HasPermission(app, constants.MAKE_INVOICE_SCOPE)
	assert.False(t, result)
	assert.Equal(t, constants.ERROR_RESTRICTED, code)
	assert.Equal(t, "This app does not have the make_invoice scope", message)

	result, code, _ = permissionsSvc.HasPermission(app, constants.MAKE_INVOICE_ONLY_SCOPE)
	assert.True(t, result)
	assert.Empty(t, code)
}

func createInvoiceScopesTestApp(t *testing.T, svc *tests.TestService, scopes ...string) *db.App {
	// a wallet backend that supports hold invoices
	svc.LNClient.(*tests.MockLn).SupportedMethods = &[]string{
		models.GET_INFO_METHOD, models.GET_BUDGET_METHOD, models.MAKE_INVOICE_METHOD, models.LOOKUP_INVOICE_METHOD,
		models.MAKE_HOLD_INVOICE_METHOD, models.SETTLE_HOLD_INVOICE_METHOD, models.CANCEL_HOLD_INVOICE_METHOD,
	}

	app, _, err := tests.CreateApp(svc)
	require.NoError(t, err)
	for _, scope := range scopes {
		err = svc.DB.Create(&db.AppPermission{AppId: app.ID, App: *app, Scope: scope}).Error
		require.NoError(t, err)
	}
	return app
}

func TestGetPermittedMethods_MakeInvoiceOnlyScope(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	app := createInvoiceScopesTestApp(t, svc, constants.MAKE_INVOICE_ONLY_SCOPE, constants.LOOKUP_INVOICE_SCOPE)

	permissionsSvc := NewPermissionsService(svc.DB, svc.EventPublisher)
	result := permissionsSvc.GetPermittedMethods(app, svc.LNClient)
	assert.ElementsMatch(t, []string{models.MAKE_INVOICE_METHOD, models.LOOKUP_INVOICE_METHOD, models.GET_INFO_METHOD, models.GET_BUDGET_METHOD}, result)
}

func TestGetPermittedMethods_MakeInvoiceScopeIncludesHoldInvoiceMethods(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	app := createInvoiceScopesTestApp(t, svc, constants.MAKE_INVOICE_SCOPE, constants.LOOKUP_INVOICE_SCOPE)

	permissionsSvc := NewPermissionsService(svc.DB, svc.EventPublisher)
	result := permissionsSvc.GetPermittedMethods(app, svc.LNClient)
	assert.ElementsMatch(t, []string{
		models.MAKE_INVOICE_METHOD, models.MAKE_HOLD_INVOICE_METHOD, models.SETTLE_HOLD_INVOICE_METHOD, models.CANCEL_HOLD_INVOICE_METHOD,
		models.LOOKUP_INVOICE_METHOD, models.GET_INFO_METHOD, models.GET_BUDGET_METHOD,
	}, result)
}

func TestGetPermittedMethods_BothInvoiceScopes(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	app := createInvoiceScopesTestApp(t, svc, constants.MAKE_INVOICE_ONLY_SCOPE, constants.MAKE_INVOICE_SCOPE)

	permissionsSvc := NewPermissionsService(svc.DB, svc.EventPublisher)
	result := permissionsSvc.GetPermittedMethods(app, svc.LNClient)
	// make_invoice is listed only once
	assert.ElementsMatch(t, []string{
		models.MAKE_INVOICE_METHOD, models.MAKE_HOLD_INVOICE_METHOD, models.SETTLE_HOLD_INVOICE_METHOD, models.CANCEL_HOLD_INVOICE_METHOD,
		models.GET_INFO_METHOD, models.GET_BUDGET_METHOD,
	}, result)
}
