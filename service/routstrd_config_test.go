package service

import (
	"testing"

	"github.com/getAlby/hub/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestReadAutoRefillConfigNoRoutstrBlock(t *testing.T) {
	// No routstr metadata at all → nil (unconfigured; callers use defaults).
	app := &db.App{Metadata: datatypes.JSON([]byte(`{}`))}
	cfg := (&RoutstrdService{}).readAutoRefillConfig(app)
	assert.Nil(t, cfg)
}

func TestReadAutoRefillConfigMissingBlockReturnsDefaults(t *testing.T) {
	// routstr block present but no autoRefill → sane defaults, never zeros.
	app := &db.App{Metadata: datatypes.JSON([]byte(`{"routstr":{}}`))}
	cfg := (&RoutstrdService{}).readAutoRefillConfig(app)
	require.NotNil(t, cfg)
	assert.Equal(t, false, cfg.Enabled)
	assert.Equal(t, int64(500), cfg.Threshold)
	assert.Equal(t, int64(1000), cfg.Amount)
	assert.Equal(t, int64(5*60*1000), cfg.CooldownMs)
}

func TestReadAutoRefillConfigEmptyBlockReturnsDefaults(t *testing.T) {
	app := &db.App{Metadata: datatypes.JSON([]byte(`{"routstr":{"autoRefill":{}}}`))}
	cfg := (&RoutstrdService{}).readAutoRefillConfig(app)
	require.NotNil(t, cfg)
	assert.Equal(t, false, cfg.Enabled)
	assert.Equal(t, int64(500), cfg.Threshold)
	assert.Equal(t, int64(1000), cfg.Amount)
}

func TestReadAutoRefillConfigFullBlock(t *testing.T) {
	app := &db.App{Metadata: datatypes.JSON([]byte(
		`{"routstr":{"autoRefill":{"enabled":true,"threshold":200,"amount":500,"cooldownMs":60000}}}`,
	))}
	cfg := (&RoutstrdService{}).readAutoRefillConfig(app)
	require.NotNil(t, cfg)
	assert.Equal(t, true, cfg.Enabled)
	assert.Equal(t, int64(200), cfg.Threshold)
	assert.Equal(t, int64(500), cfg.Amount)
	assert.Equal(t, int64(60000), cfg.CooldownMs)
}

func TestReadAutoRefillConfigPartialBlockFillsDefaults(t *testing.T) {
	// Only threshold set → defaults for the rest.
	app := &db.App{Metadata: datatypes.JSON([]byte(
		`{"routstr":{"autoRefill":{"threshold":123}}}`,
	))}
	cfg := (&RoutstrdService{}).readAutoRefillConfig(app)
	require.NotNil(t, cfg)
	assert.Equal(t, false, cfg.Enabled)
	assert.Equal(t, int64(123), cfg.Threshold)
	assert.Equal(t, int64(1000), cfg.Amount)
	assert.Equal(t, int64(5*60*1000), cfg.CooldownMs)
}

func TestReadAutoRefillConfigMalformedJSON(t *testing.T) {
	app := &db.App{Metadata: datatypes.JSON([]byte(`not-json`))}
	cfg := (&RoutstrdService{}).readAutoRefillConfig(app)
	assert.Nil(t, cfg)
}
