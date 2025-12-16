package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/config"
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
	assert.False(t, svc.Cfg.CheckUnlockPassword(""))
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

	assert.False(t, svc.Cfg.CheckUnlockPassword(""))
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

func TestSetIgnore_NoEncryptionKey(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	err = svc.Cfg.SetIgnore("key", "value", "")
	require.NoError(t, err)

	value, err := svc.Cfg.Get("key", "")
	assert.Equal(t, "value", value)

	err = svc.Cfg.SetIgnore("key", "value2", "")
	require.NoError(t, err)

	// value should not be updated
	updatedValue, err := svc.Cfg.Get("key", "")
	assert.Equal(t, "value", updatedValue)
}

func TestSetIgnore_EncryptionKey(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	unlockPassword := "123"

	err = svc.Cfg.SetIgnore("key", "value", unlockPassword)
	require.NoError(t, err)

	value, err := svc.Cfg.Get("key", unlockPassword)
	assert.Equal(t, "value", value)

	invalidValue, err := svc.Cfg.Get("key", unlockPassword+"1")
	assert.Error(t, err)
	assert.Equal(t, "", invalidValue)

	err = svc.Cfg.SetIgnore("key", "value2", unlockPassword)
	require.NoError(t, err)

	// value should not be updated
	updatedValue, err := svc.Cfg.Get("key", unlockPassword)
	assert.Equal(t, "value", updatedValue)
}

func TestSetUpdate_NoEncryptionKey(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	err = svc.Cfg.SetUpdate("key", "value", "")
	require.NoError(t, err)

	value, err := svc.Cfg.Get("key", "")
	assert.Equal(t, "value", value)

	err = svc.Cfg.SetUpdate("key", "value2", "")
	require.NoError(t, err)

	// value should be updated
	updatedValue, err := svc.Cfg.Get("key", "")
	assert.Equal(t, "value2", updatedValue)
}

func TestSetUpdate_EncryptionKey(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	unlockPassword := "123"

	err = svc.Cfg.SetUpdate("key", "value", unlockPassword)
	require.NoError(t, err)

	value, err := svc.Cfg.Get("key", unlockPassword)
	assert.Equal(t, "value", value)

	invalidValue, err := svc.Cfg.Get("key", unlockPassword+"1")
	assert.Error(t, err)
	assert.Equal(t, "", invalidValue)

	err = svc.Cfg.SetUpdate("key", "value2", unlockPassword)
	require.NoError(t, err)

	// value should be updated
	updatedValue, err := svc.Cfg.Get("key", unlockPassword)
	assert.Equal(t, "value2", updatedValue)
}

func TestSetUpdate_NoEncryptionKeyToEncryptionKey(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	err = svc.Cfg.SetUpdate("key", "value", "")
	require.NoError(t, err)

	value, err := svc.Cfg.Get("key", "")
	assert.Equal(t, "value", value)

	unlockPassword := "123"

	err = svc.Cfg.SetUpdate("key", "value2", unlockPassword)
	require.NoError(t, err)

	// value should be updated
	updatedValue, err := svc.Cfg.Get("key", unlockPassword)
	assert.Equal(t, "value2", updatedValue)
}

func TestSetUpdate_EncryptionKeyToNoEncryptionKey(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	unlockPassword := "123"

	err = svc.Cfg.SetUpdate("key", "value", unlockPassword)
	require.NoError(t, err)

	value, err := svc.Cfg.Get("key", unlockPassword)
	assert.Equal(t, "value", value)

	err = svc.Cfg.SetUpdate("key", "value2", "")
	require.NoError(t, err)

	// value should be updated
	updatedValue, err := svc.Cfg.Get("key", "")
	assert.Equal(t, "value2", updatedValue)
}

func TestJWTSecret_GeneratedOnUnlock(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	cfg, err := config.NewConfig(&config.AppConfig{}, svc.DB)
	require.NoError(t, err)

	err = cfg.ChangeUnlockPassword("", "123")
	require.NoError(t, err)

	err = cfg.Unlock("123")
	require.NoError(t, err)

	jwtSecret, err := cfg.GetJWTSecret()
	require.NoError(t, err)
	assert.NotEmpty(t, jwtSecret)

	encryptedSecret, err := cfg.Get("JWTSecret", "")
	require.NoError(t, err)
	decryptedSecret, err := cfg.Get("JWTSecret", "123")
	require.NoError(t, err)
	assert.NotEqual(t, encryptedSecret, decryptedSecret)

	// unlock again without doing anything, ensure the same JWT secret is returned
	err = cfg.Unlock("123")
	require.NoError(t, err)
	jwtSecret2, err := cfg.GetJWTSecret()
	require.NoError(t, err)
	assert.Equal(t, jwtSecret, jwtSecret2)
}

func TestJWTSecret_ChangePassword(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	cfg, err := config.NewConfig(&config.AppConfig{}, svc.DB)
	require.NoError(t, err)

	err = cfg.ChangeUnlockPassword("", "123")
	require.NoError(t, err)

	err = cfg.Unlock("123")
	require.NoError(t, err)

	jwtSecret, err := cfg.GetJWTSecret()
	require.NoError(t, err)
	assert.NotEmpty(t, jwtSecret)

	err = cfg.ChangeUnlockPassword("", "1234")
	require.NoError(t, err)

	newJwtSecret, err := cfg.GetJWTSecret()
	require.ErrorContains(t, err, "unlock")

	err = cfg.Unlock("1234")
	require.NoError(t, err)

	// a new JWT secret must be generated after password change
	newJwtSecret, err = cfg.GetJWTSecret()
	require.NoError(t, err)
	assert.NotEmpty(t, newJwtSecret)
	assert.NotEqual(t, newJwtSecret, jwtSecret)
}

func TestJWTSecret_ReplaceUnencryptedSecretOnUnlock(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	cfg, err := config.NewConfig(&config.AppConfig{}, svc.DB)
	require.NoError(t, err)

	err = cfg.ChangeUnlockPassword("", "123")
	require.NoError(t, err)

	// simulate a hub that had an unencrypted JWT secret
	oldJwtSecret := "dummy secret"
	err = svc.Cfg.SetUpdate("JWTSecret", oldJwtSecret, "")
	require.NoError(t, err)

	err = cfg.Unlock("123")
	require.NoError(t, err)

	jwtSecret, err := cfg.GetJWTSecret()
	require.NoError(t, err)
	assert.NotEmpty(t, jwtSecret)
	assert.NotEqual(t, jwtSecret, oldJwtSecret)

	// ensure it is saved to DB
	jwtSecretFromCfg, err := cfg.Get("JWTSecret", "123")
	require.NoError(t, err)
	assert.Equal(t, jwtSecret, jwtSecretFromCfg)
}
