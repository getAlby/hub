package service

import (
	"testing"

	"github.com/getAlby/hub/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func routstrApp(id uint, meta string) db.App {
	return db.App{ID: id, Name: "Routstr", Metadata: datatypes.JSON([]byte(meta))}
}

func TestSelectRoutstrAppNone(t *testing.T) {
	apps := []db.App{
		routstrApp(1, `{"app_store_app_id":"other"}`),
		{ID: 2, Name: "plain", Metadata: datatypes.JSON([]byte(`{}`))},
	}
	assert.Nil(t, selectRoutstrApp(apps))
}

func TestSelectRoutstrAppFallbackFirstRoutstrApp(t *testing.T) {
	apps := []db.App{
		routstrApp(1, `{"app_store_app_id":"routstr"}`),
		routstrApp(2, `{"app_store_app_id":"routstr"}`),
	}
	got := selectRoutstrApp(apps)
	require.NotNil(t, got)
	assert.Equal(t, uint(1), got.ID)
}

func TestSelectRoutstrAppConfiguredBeatsFallback(t *testing.T) {
	// App 1: Routstr but no autoRefill block (fallback tier).
	// App 2: Routstr with an autoRefill block but disabled (configured tier).
	// Expect app 2 — configured beats first-found.
	apps := []db.App{
		routstrApp(1, `{"app_store_app_id":"routstr"}`),
		routstrApp(2, `{"app_store_app_id":"routstr","routstr":{"autoRefill":{"enabled":false}}}`),
	}
	got := selectRoutstrApp(apps)
	require.NotNil(t, got)
	assert.Equal(t, uint(2), got.ID)
}

func TestSelectRoutstrAppEnabledBeatsConfigured(t *testing.T) {
	// App 1: Routstr with an autoRefill block, disabled.
	// App 2: Routstr with enabled autoRefill.
	apps := []db.App{
		routstrApp(1, `{"app_store_app_id":"routstr","routstr":{"autoRefill":{"enabled":false}}}`),
		routstrApp(2, `{"app_store_app_id":"routstr","routstr":{"autoRefill":{"enabled":true,"threshold":200,"amount":500}}}`),
	}
	got := selectRoutstrApp(apps)
	require.NotNil(t, got)
	assert.Equal(t, uint(2), got.ID)
}

func TestSelectRoutstrAppBrokenValuesStillConfiguredTier(t *testing.T) {
	// Enabled with zero threshold/amount disqualifies the tier-1 early return
	// (the loop would treat it as stopped) but the block still marks the app
	// as "configured", so it beats a later disabled-but-blocked app.
	apps := []db.App{
		routstrApp(1, `{"app_store_app_id":"routstr","routstr":{"autoRefill":{"enabled":true,"threshold":0,"amount":0}}}`),
		routstrApp(2, `{"app_store_app_id":"routstr","routstr":{"autoRefill":{"enabled":false}}}`),
	}
	got := selectRoutstrApp(apps)
	require.NotNil(t, got)
	assert.Equal(t, uint(1), got.ID)
}

func TestSelectRoutstrAppSkipsMalformedMetadata(t *testing.T) {
	apps := []db.App{
		routstrApp(1, `not-json`),
		routstrApp(2, `{"app_store_app_id":"routstr"}`),
	}
	got := selectRoutstrApp(apps)
	require.NotNil(t, got)
	assert.Equal(t, uint(2), got.ID)
}
