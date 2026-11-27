package api

import (
	"context"
	"testing"
	"time"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/tests"
	"github.com/getAlby/hub/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildReturnToUrl(t *testing.T) {
	relayUrls := []string{"wss://relay.getalby.com/v1"}
	walletPubkey := "6f8bf1b7d58ac41b2c793837ba528c1d0a4a1cd2e3f5b7c9d0e1f2a3b4c5d6e7"

	assert.Equal(t,
		"https://example.com?pubkey=6f8bf1b7d58ac41b2c793837ba528c1d0a4a1cd2e3f5b7c9d0e1f2a3b4c5d6e7&relay=wss%3A%2F%2Frelay.getalby.com%2Fv1",
		buildReturnToUrl("https://example.com", relayUrls, walletPubkey, "", false))

	// existing query parameters are preserved and lud16 is added
	assert.Equal(t,
		"https://example.com/path?foo=bar&lud16=user%40getalby.com&pubkey=6f8bf1b7d58ac41b2c793837ba528c1d0a4a1cd2e3f5b7c9d0e1f2a3b4c5d6e7&relay=wss%3A%2F%2Frelay.getalby.com%2Fv1",
		buildReturnToUrl("https://example.com/path?foo=bar", relayUrls, walletPubkey, "user@getalby.com", false))

	// isolated apps do not receive a lightning address
	assert.Equal(t,
		"http://example.com?pubkey=6f8bf1b7d58ac41b2c793837ba528c1d0a4a1cd2e3f5b7c9d0e1f2a3b4c5d6e7&relay=wss%3A%2F%2Frelay.getalby.com%2Fv1",
		buildReturnToUrl("http://example.com", relayUrls, walletPubkey, "user@getalby.com", true))

	// only http and https URLs are accepted
	assert.Equal(t, "", buildReturnToUrl("", relayUrls, walletPubkey, "", false))
	assert.Equal(t, "", buildReturnToUrl("example.com/path", relayUrls, walletPubkey, "", false))
	assert.Equal(t, "", buildReturnToUrl("example://app", relayUrls, walletPubkey, "", false))
	assert.Equal(t, "", buildReturnToUrl("javascript:void(0)", relayUrls, walletPubkey, "", false))
	assert.Equal(t, "", buildReturnToUrl("::invalid::", relayUrls, walletPubkey, "", false))
}

func TestCreateApp_SuperuserScopeIncorrectPassword(t *testing.T) {
	cfg := mocks.NewMockConfig(t)
	cfg.On("CheckUnlockPassword", "").Return(false)
	theAPI := &api{svc: mocks.NewMockService(t), cfg: cfg}
	response, err := theAPI.CreateApp(&CreateAppRequest{
		Scopes: []string{constants.SUPERUSER_SCOPE},
	})

	assert.Nil(t, response)
	require.Error(t, err)
	assert.Equal(t, "incorrect unlock password to create app with superuser permission", err.Error())
}

func TestCreateApp_LightningAddress(t *testing.T) {
	for _, isolated := range []bool{false, true} {
		testSvc, err := tests.CreateTestService(t)
		require.NoError(t, err)
		// isolated apps require a backend that supports sub-wallets
		testSvc.Cfg.SetUpdate("LNBackendType", config.LDKBackendType, "")

		albyOAuthSvc := mocks.NewMockAlbyOAuthService(t)
		albyOAuthSvc.On("GetLightningAddress").Return("owner@getalby.com", nil)

		theAPI := &api{
			appsSvc:      testSvc.AppsService,
			cfg:          testSvc.Cfg,
			albyOAuthSvc: albyOAuthSvc,
		}

		response, err := theAPI.CreateApp(&CreateAppRequest{
			Name:     "Test",
			Scopes:   []string{constants.GET_INFO_SCOPE},
			Isolated: isolated,
			ReturnTo: "https://example.com",
		})
		require.NoError(t, err)

		if isolated {
			// isolated apps must not receive the owner's lightning address
			assert.Empty(t, response.Lud16)
			assert.NotContains(t, response.PairingUri, "lud16")
			assert.NotContains(t, response.ReturnTo, "lud16")
		} else {
			assert.Equal(t, "owner@getalby.com", response.Lud16)
			assert.Contains(t, response.PairingUri, "&lud16=owner@getalby.com")
			assert.Contains(t, response.ReturnTo, "lud16=owner%40getalby.com")
		}

		testSvc.Remove()
	}
}

func TestDeleteApp_NonStringLightningAddress(t *testing.T) {
	testSvc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer testSvc.Remove()

	// no expectations: the Alby API must not be called for a non-string lud16
	albyOAuthSvc := mocks.NewMockAlbyOAuthService(t)
	theAPI := &api{
		appsSvc:      testSvc.AppsService,
		cfg:          testSvc.Cfg,
		albyOAuthSvc: albyOAuthSvc,
	}

	metadata := map[string]interface{}{
		constants.METADATA_APPSTORE_APP_ID_KEY: constants.SUBWALLET_APPSTORE_APP_ID,
		"lud16":                                map[string]interface{}{"a": 1},
	}
	app, _, err := testSvc.AppsService.CreateApp("test", "", 0, "monthly", nil, []string{constants.GET_INFO_SCOPE}, false, metadata)
	require.NoError(t, err)

	assert.False(t, testSvc.AppsService.HasLightningAddress(app))

	err = theAPI.DeleteLightningAddress(context.Background(), app.ID)
	assert.EqualError(t, err, "no lightning address set")

	require.NoError(t, theAPI.DeleteApp(app))
	assert.Nil(t, testSvc.AppsService.GetAppById(app.ID))
}

func TestListApps_NameFilter(t *testing.T) {
	testSvc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer testSvc.Remove()

	for _, name := range []string{"Alby Go", "Damus", "Damus (1)", "100%_app", "100xapp", `back\slash`} {
		_, _, err := testSvc.AppsService.CreateApp(name, "", 0, "monthly", nil, []string{constants.GET_INFO_SCOPE}, false, nil)
		require.NoError(t, err)
	}

	theAPI := &api{db: testSvc.DB, keys: testSvc.Keys}

	listNames := func(filters ListAppsFilters) []string {
		response, err := theAPI.ListApps(0, 0, filters, "")
		require.NoError(t, err)
		names := []string{}
		for _, app := range response.Apps {
			names = append(names, app.Name)
		}
		return names
	}

	// name matches anywhere in the name, case-insensitively
	assert.ElementsMatch(t, []string{"Damus", "Damus (1)"}, listNames(ListAppsFilters{Name: "Damus"}))
	assert.ElementsMatch(t, []string{"Damus", "Damus (1)"}, listNames(ListAppsFilters{Name: "damus"}))
	assert.ElementsMatch(t, []string{"Alby Go"}, listNames(ListAppsFilters{Name: "lby"}))
	assert.ElementsMatch(t, []string{"Alby Go"}, listNames(ListAppsFilters{Name: "GO"}))
	assert.ElementsMatch(t, []string{"Damus", "Damus (1)"}, listNames(ListAppsFilters{Name: "amu"}))
	assert.Empty(t, listNames(ListAppsFilters{Name: "xyz"}))

	// LIKE wildcards in user input are matched literally
	assert.ElementsMatch(t, []string{"100%_app"}, listNames(ListAppsFilters{Name: "%_"}))
	assert.Empty(t, listNames(ListAppsFilters{Name: "_x"}))
	// without escaping, `\s` would match any name containing "s"
	assert.ElementsMatch(t, []string{`back\slash`}, listNames(ListAppsFilters{Name: `\s`}))
}

func TestCreateApp_MakeInvoiceOnlyScope(t *testing.T) {
	testSvc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer testSvc.Remove()

	albyOAuthSvc := mocks.NewMockAlbyOAuthService(t)
	albyOAuthSvc.On("GetLightningAddress").Return("", nil)

	theAPI := &api{
		appsSvc:      testSvc.AppsService,
		cfg:          testSvc.Cfg,
		albyOAuthSvc: albyOAuthSvc,
	}

	response, err := theAPI.CreateApp(&CreateAppRequest{
		Name:   "Test",
		Scopes: []string{constants.MAKE_INVOICE_ONLY_SCOPE, constants.LOOKUP_INVOICE_SCOPE},
	})
	require.NoError(t, err)

	appPermissions := []db.AppPermission{}
	err = testSvc.DB.Where("app_id = ?", response.Id).Order("id").Find(&appPermissions).Error
	require.NoError(t, err)
	require.Equal(t, 2, len(appPermissions))
	assert.Equal(t, constants.MAKE_INVOICE_ONLY_SCOPE, appPermissions[0].Scope)
	assert.Equal(t, constants.LOOKUP_INVOICE_SCOPE, appPermissions[1].Scope)

	// unknown scopes are still refused
	response, err = theAPI.CreateApp(&CreateAppRequest{
		Name:   "Test 2",
		Scopes: []string{"make_hold_invoice"},
	})
	assert.Nil(t, response)
	require.Error(t, err)
	assert.Equal(t, "did not recognize requested scope: make_hold_invoice", err.Error())
}

func getAppScopes(t *testing.T, testSvc *tests.TestService, appId uint) []string {
	appPermissions := []db.AppPermission{}
	err := testSvc.DB.Where("app_id = ?", appId).Find(&appPermissions).Error
	require.NoError(t, err)
	scopes := []string{}
	for _, appPermission := range appPermissions {
		scopes = append(scopes, appPermission.Scope)
	}
	return scopes
}

func TestUpdateApp_MakeInvoiceOnlyScope(t *testing.T) {
	testSvc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer testSvc.Remove()

	mockSvc := mocks.NewMockService(t)
	mockSvc.On("GetEventPublisher").Return(testSvc.EventPublisher)

	theAPI := &api{
		db:  testSvc.DB,
		svc: mockSvc,
	}

	app, _, err := testSvc.AppsService.CreateApp("Test", "", 0, "", nil, []string{constants.MAKE_INVOICE_ONLY_SCOPE, constants.LOOKUP_INVOICE_SCOPE}, false, nil)
	require.NoError(t, err)

	// renaming does not change the scopes
	name := "Renamed"
	err = theAPI.UpdateApp(app, &UpdateAppRequest{Name: &name})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{constants.MAKE_INVOICE_ONLY_SCOPE, constants.LOOKUP_INVOICE_SCOPE}, getAppScopes(t, testSvc, app.ID))

	// saving the same scopes (as the UI does when editing expiry) keeps the restricted scope
	expiresAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	err = theAPI.UpdateApp(app, &UpdateAppRequest{
		Scopes:          []string{constants.MAKE_INVOICE_ONLY_SCOPE, constants.LOOKUP_INVOICE_SCOPE},
		ExpiresAt:       &expiresAt,
		UpdateExpiresAt: true,
	})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{constants.MAKE_INVOICE_ONLY_SCOPE, constants.LOOKUP_INVOICE_SCOPE}, getAppScopes(t, testSvc, app.ID))

	// the API writes exactly the scopes it is sent
	err = theAPI.UpdateApp(app, &UpdateAppRequest{
		Scopes: []string{constants.MAKE_INVOICE_SCOPE, constants.LOOKUP_INVOICE_SCOPE},
	})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{constants.MAKE_INVOICE_SCOPE, constants.LOOKUP_INVOICE_SCOPE}, getAppScopes(t, testSvc, app.ID))
}

func TestGetWalletCapabilities_NoMakeInvoiceOnlyScope(t *testing.T) {
	for _, supportsHoldInvoices := range []bool{false, true} {
		mockLn, err := tests.NewMockLn()
		require.NoError(t, err)
		if supportsHoldInvoices {
			mockLn.SupportedMethods = &[]string{"get_info", "make_invoice", "lookup_invoice", "make_hold_invoice", "settle_hold_invoice", "cancel_hold_invoice"}
		}

		mockSvc := mocks.NewMockService(t)
		mockSvc.On("GetLNClient").Return(mockLn)
		theAPI := &api{svc: mockSvc}

		capabilities, err := theAPI.GetWalletCapabilities(context.TODO())
		require.NoError(t, err)
		// the restricted scope is only granted for explicitly requested methods
		assert.Contains(t, capabilities.Scopes, constants.MAKE_INVOICE_SCOPE)
		assert.NotContains(t, capabilities.Scopes, constants.MAKE_INVOICE_ONLY_SCOPE)
	}
}
