package api

import (
	"testing"

	"github.com/getAlby/hub/constants"
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
