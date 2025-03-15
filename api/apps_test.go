package api

import (
	"testing"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
