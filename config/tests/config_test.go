package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/tests"
)

func TestCheckUnlockPasswordCache_InvalidSecond(t *testing.T) {
	unlockPassword := "123"
	svc, err := tests.CreateTestServiceWithMnemonic(t, "", unlockPassword)
	require.NoError(t, err)
	defer svc.Remove()

	err = svc.Cfg.ChangeUnlockPassword("", unlockPassword)
	require.NoError(t, err)
	err = svc.Cfg.SaveUnlockPasswordCheck(unlockPassword)
	require.NoError(t, err)

	value, err := svc.Cfg.Get("UnlockPasswordCheck", unlockPassword)
	require.NoError(t, err)
	require.Equal(t, "THIS STRING SHOULD MATCH IF PASSWORD IS CORRECT", value)

	assert.True(t, svc.Cfg.CheckUnlockPassword(unlockPassword))
	assert.False(t, svc.Cfg.CheckUnlockPassword(unlockPassword+"1"))
}
func TestCheckUnlockPasswordCache_InvalidFirst(t *testing.T) {
	unlockPassword := "123"
	svc, err := tests.CreateTestServiceWithMnemonic(t, "", unlockPassword)
	require.NoError(t, err)
	defer svc.Remove()

	err = svc.Cfg.ChangeUnlockPassword("", unlockPassword)
	require.NoError(t, err)
	err = svc.Cfg.SaveUnlockPasswordCheck(unlockPassword)
	require.NoError(t, err)

	value, err := svc.Cfg.Get("UnlockPasswordCheck", unlockPassword)
	require.NoError(t, err)
	require.Equal(t, "THIS STRING SHOULD MATCH IF PASSWORD IS CORRECT", value)

	value, err = svc.Cfg.Get("UnlockPasswordCheck", unlockPassword+"1")
	require.Error(t, err)

	assert.False(t, svc.Cfg.CheckUnlockPassword(unlockPassword+"1"))
	assert.True(t, svc.Cfg.CheckUnlockPassword(unlockPassword))
}

func TestCheckUnlockPassword_ChangePassword(t *testing.T) {
	unlockPassword := "123"
	svc, err := tests.CreateTestServiceWithMnemonic(t, "", unlockPassword)
	require.NoError(t, err)
	defer svc.Remove()

	err = svc.Cfg.ChangeUnlockPassword("", unlockPassword)
	require.NoError(t, err)
	err = svc.Cfg.SaveUnlockPasswordCheck(unlockPassword)
	require.NoError(t, err)

	value, err := svc.Cfg.Get("UnlockPasswordCheck", unlockPassword)
	require.NoError(t, err)
	require.Equal(t, "THIS STRING SHOULD MATCH IF PASSWORD IS CORRECT", value)

	assert.True(t, svc.Cfg.CheckUnlockPassword(unlockPassword))

	newUnlockPassword := "1234"

	err = svc.Cfg.ChangeUnlockPassword(unlockPassword, newUnlockPassword)
	require.NoError(t, err)

	assert.False(t, svc.Cfg.CheckUnlockPassword(unlockPassword))
	assert.True(t, svc.Cfg.CheckUnlockPassword(newUnlockPassword))
	// test caching
	assert.False(t, svc.Cfg.CheckUnlockPassword(unlockPassword))
	assert.True(t, svc.Cfg.CheckUnlockPassword(newUnlockPassword))
}
