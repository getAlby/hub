package keys

import (
	"strconv"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/tests/db"

	"github.com/nbd-wtf/go-nostr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

func TestUseExistingMnemonic(t *testing.T) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	gormDb, err := db.NewDB(t)
	require.NoError(t, err)
	defer db.CloseDB(gormDb)

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

	// ensure backup key uses correct derivation path
	derivedKeyFromKeys, err := keys.DeriveKey([]uint32{bip32.FirstHardenedChild})
	require.NoError(t, err)

	masterKey, err := bip32.NewMasterKey(bip39.NewSeed(mnemonic, ""))
	assert.NoError(t, err)

	appKey, err := masterKey.NewChildKey(bip32.FirstHardenedChild + 128029 /* 🐝 */)
	assert.NoError(t, err)

	encryptedChannelsBackupKey, err := appKey.NewChildKey(bip32.FirstHardenedChild)
	assert.NoError(t, err)

	assert.Equal(t, encryptedChannelsBackupKey.String(), derivedKeyFromKeys.String())

	// get a wallet key for app ID 2, expect it is derived correctly
	appWalletPrivateKey, err := keys.GetAppWalletKey(2)
	require.NoError(t, err)
	appWalletPubkey, err := nostr.GetPublicKey(appWalletPrivateKey)
	require.NoError(t, err)

	assert.Equal(t, "dd9e304d24f29f3481d5cf18a76c85ca3e95931aee3c997a27f267e975e72976", appWalletPubkey)
}

func TestGenerateNewMnemonic(t *testing.T) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	gormDb, err := db.NewDB(t)
	require.NoError(t, err)
	defer db.CloseDB(gormDb)

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

func TestGenerateSwapMnemonic(t *testing.T) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))
	gormDb, err := db.NewDB(t)
	require.NoError(t, err)
	defer db.CloseDB(gormDb)

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

	seed := bip39.NewSeed(mnemonicFromConfig, "")
	masterKey, err := bip32.NewMasterKey(seed)
	require.NoError(t, err)

	swapMnemonic, err := keys.GenerateSwapMnemonic(masterKey)
	require.NoError(t, err)

	// this matches https://iancoleman.io/bip39/ -> check "Show BIP85" and set BIP85 Index to 128260
	expectedSwapMnemonic := "truth cargo pluck prefer mosquito symptom review kitchen exile fit corn vault"
	assert.Equal(t, expectedSwapMnemonic, swapMnemonic)
}
