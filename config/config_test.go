// see ./tests/config_test.go for config tests with DB coverage for both SQlite & Postgres
package config

import (
	"strconv"
	"testing"

	"github.com/getAlby/hub/db/migrations"
	"github.com/getAlby/hub/logger"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCheckCache_NoEncryptionKey(t *testing.T) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	err = migrations.Migrate(db)
	require.NoError(t, err)

	cfg, err := NewConfig(&AppConfig{
		Workdir: ".test",
	}, db)
	require.NoError(t, err)

	err = cfg.SetUpdate("key", "value", "")
	require.NoError(t, err)

	require.Equal(t, cfg.cache["key"][""], "")

	value, err := cfg.Get("key", "")
	require.NoError(t, err)
	require.Equal(t, "value", value)

	require.Equal(t, cfg.cache["key"][""], "value")

	// test we can access the cached value without the db
	cfg.db = nil
	value, err = cfg.Get("key", "")
	require.NoError(t, err)
	require.Equal(t, "value", value)
}

func TestCheckUnlockPasswordCache(t *testing.T) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	unlockPassword := "123"

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	err = migrations.Migrate(db)
	require.NoError(t, err)

	cfg, err := NewConfig(&AppConfig{
		Workdir: ".test",
	}, db)
	require.NoError(t, err)
	err = cfg.ChangeUnlockPassword("", unlockPassword)
	require.NoError(t, err)
	err = cfg.SaveUnlockPasswordCheck(unlockPassword)
	require.NoError(t, err)

	// check cache
	assert.Nil(t, cfg.cache["UnlockPasswordCheck"])
	// check password
	assert.True(t, cfg.CheckUnlockPassword(unlockPassword))
	assert.NotNil(t, cfg.cache["UnlockPasswordCheck"])

	// check hash cache
	cacheValue, ok := cfg.cache["UnlockPasswordCheck"]
	require.True(t, ok)
	assert.Equal(t, 1, len(cacheValue))

	assert.False(t, cfg.CheckUnlockPassword(unlockPassword+"1"))

	// check hash cache - length should not have changed because decrypt failed with invalid password
	cacheValue2, ok := cfg.cache["UnlockPasswordCheck"]
	require.True(t, ok)
	assert.Equal(t, 1, len(cacheValue2))
	require.Equal(t, cacheValue, cacheValue2)

	// change the password
	newUnlockPassword := unlockPassword + "1"
	err = cfg.ChangeUnlockPassword(unlockPassword, newUnlockPassword)
	require.NoError(t, err)
	assert.Equal(t, 0, len(cfg.cache["UnlockPasswordCheck"]))

	// test we can access the cached value without the db
	assert.True(t, cfg.CheckUnlockPassword(newUnlockPassword))
	assert.NotNil(t, cfg.cache["UnlockPasswordCheck"])
	cfg.db = nil
	assert.True(t, cfg.CheckUnlockPassword(newUnlockPassword))

	// should panic when trying to access the db for an uncached value
	hitPanic := false
	func() {
		defer func() {
			// ensure the app cannot panic if firing events to Alby API fails
			if r := recover(); r != nil {
				hitPanic = true
			}
		}()
		assert.False(t, cfg.CheckUnlockPassword(unlockPassword))
	}()
	assert.True(t, hitPanic)
}
