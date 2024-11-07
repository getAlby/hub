package keys

import (
	"strings"
	"testing"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

const testDB = "test.db"

func TestUseExistingMnemonic(t *testing.T) {
	gormDb, err := db.NewDB(testDB, true)
	require.NoError(t, err)

	mnemonic := "thought turkey ask pottery head say catalog desk pledge elbow naive mimic"
	unlockPassword := "123"

	config, err := config.NewConfig(&config.AppConfig{}, gormDb)
	require.NoError(t, err)
	config.SetUpdate("Mnemonic", mnemonic, unlockPassword)

	keys := NewKeys()
	err = keys.Init(config, unlockPassword)
	require.NoError(t, err)

	mnemonicFromConfig, err := config.Get("Mnemonic", unlockPassword)
	require.NoError(t, err)
	require.Equal(t, mnemonic, mnemonicFromConfig)

	derivedKeyFromKeys, err := keys.DeriveKey([]uint32{bip32.FirstHardenedChild})
	require.NoError(t, err)

	masterKey, err := bip32.NewMasterKey(bip39.NewSeed(mnemonic, ""))
	assert.NoError(t, err)

	appKey, err := masterKey.NewChildKey(bip32.FirstHardenedChild + 128029 /* 🐝 */)
	assert.NoError(t, err)

	encryptedChannelsBackupKey, err := appKey.NewChildKey(bip32.FirstHardenedChild)
	assert.NoError(t, err)

	assert.Equal(t, encryptedChannelsBackupKey.String(), derivedKeyFromKeys.String())
}

func TestGenerateNewMnemonic(t *testing.T) {
	gormDb, err := db.NewDB(testDB, true)
	require.NoError(t, err)

	unlockPassword := "123"

	config, err := config.NewConfig(&config.AppConfig{}, gormDb)
	require.NoError(t, err)

	keys := NewKeys()
	err = keys.Init(config, unlockPassword)
	require.NoError(t, err)

	mnemonicFromConfig, err := config.Get("Mnemonic", unlockPassword)
	require.NoError(t, err)

	// expect a new 12-word mnemonic to be saved
	assert.Equal(t, 12, len(strings.Split(mnemonicFromConfig, " ")))

	// re-create keys, ensure same mnemonic is used
	keys = NewKeys()
	err = keys.Init(config, unlockPassword)
	require.NoError(t, err)

	mnemonicFromConfig2, err := config.Get("Mnemonic", unlockPassword)
	require.NoError(t, err)
	assert.Equal(t, mnemonicFromConfig, mnemonicFromConfig2)

	// check derivation

	derivedKeyFromKeys, err := keys.DeriveKey([]uint32{bip32.FirstHardenedChild})
	require.NoError(t, err)

	masterKey, err := bip32.NewMasterKey(bip39.NewSeed(mnemonicFromConfig, ""))
	assert.NoError(t, err)

	appKey, err := masterKey.NewChildKey(bip32.FirstHardenedChild + 128029 /* 🐝 */)
	assert.NoError(t, err)

	encryptedChannelsBackupKey, err := appKey.NewChildKey(bip32.FirstHardenedChild)
	assert.NoError(t, err)

	assert.Equal(t, encryptedChannelsBackupKey.String(), derivedKeyFromKeys.String())
}
