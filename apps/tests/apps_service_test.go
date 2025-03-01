package tests

import (
	"testing"

	"github.com/getAlby/hub/apps"
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleCreateApp_NilScopes(t *testing.T) {
	// ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	appsService := apps.NewAppsService(svc.DB, svc.EventPublisher, svc.Keys, svc.Cfg)
	app, secretKey, err := appsService.CreateApp("Test", "", 0, "monthly", nil, nil, false, nil)

	assert.Nil(t, app)
	assert.Equal(t, "", secretKey)
	require.Error(t, err)
	assert.Equal(t, "no scopes provided", err.Error())
}

func TestHandleCreateApp_EmptyScopes(t *testing.T) {
	// ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	appsService := apps.NewAppsService(svc.DB, svc.EventPublisher, svc.Keys, svc.Cfg)
	app, secretKey, err := appsService.CreateApp("Test", "", 0, "monthly", nil, []string{}, false, nil)

	assert.Nil(t, app)
	assert.Equal(t, "", secretKey)
	require.Error(t, err)
	assert.Equal(t, "no scopes provided", err.Error())
}

func TestHandleCreateApp_IsolatedUnsupportedBackendType(t *testing.T) {
	// ctx := context.TODO()
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()
	svc.Cfg.SetUpdate("BackendType", config.CashuBackendType, "")

	appsService := apps.NewAppsService(svc.DB, svc.EventPublisher, svc.Keys, svc.Cfg)
	app, secretKey, err := appsService.CreateApp("Test", "", 0, "monthly", nil, []string{constants.GET_INFO_SCOPE}, true, nil)

	assert.Nil(t, app)
	assert.Equal(t, "", secretKey)
	require.Error(t, err)
	assert.Equal(t, "sub-wallets are currently not supported on your node backend. Try LDK or LND", err.Error())
}
