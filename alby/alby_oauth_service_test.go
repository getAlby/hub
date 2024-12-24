package alby

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/tests"
)

func TestExistingEncryptedBackup(t *testing.T) {
	svc, err := tests.CreateTestService()
	require.NoError(t, err)
	defer svc.Remove()

	mnemonic := "limit reward expect search tissue call visa fit thank cream brave jump"
	unlockPassword := "123"
	svc.Cfg.SetUpdate("Mnemonic", mnemonic, unlockPassword)
	err = svc.Keys.Init(svc.Cfg, unlockPassword)
	assert.NoError(t, err)

	encryptedBackup := "3fd21f9a393d8345ddbdd449-ba05c3dbafdfb7eea574373b7763d0c81c599b2cd1735e59a1c5571379498f4da8fe834c3403824ab02b61005abc1f563c638f425c65420e82941efe94794555c8b145a0603733ee115277f860011e6a17fd8c22f1d73a096ff7275582aac19b430940b40a2559c7ff59a063305290ef7c9ba46f9de17b0ddbac9030b0"

	masterKey, err := bip32.NewMasterKey(bip39.NewSeed(mnemonic, ""))
	assert.NoError(t, err)

	appKey, err := masterKey.NewChildKey(bip32.FirstHardenedChild + 128029 /* 🐝 */)
	assert.NoError(t, err)
	encryptedChannelsBackupKey, err := appKey.NewChildKey(bip32.FirstHardenedChild)
	assert.NoError(t, err)

	decrypted, err := config.AesGcmDecryptWithKey(encryptedBackup, encryptedChannelsBackupKey.Key)
	assert.NoError(t, err)

	assert.Equal(t, "{\"node_id\":\"037e702144c4fa485d42f0f69864e943605823763866cf4bf619d2d2cf2eda420b\",\"channels\":[],\"monitors\":[]}\n", decrypted)
}

func TestEncryptedBackup(t *testing.T) {
	mnemonic := "limit reward expect search tissue call visa fit thank cream brave jump"
	unlockPassword := "123"
	svc, err := tests.CreateTestServiceWithMnemonic(mnemonic, unlockPassword)
	require.NoError(t, err)
	defer svc.Remove()

	albyOAuthSvc := NewAlbyOAuthService(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)
	encryptedBackup, err := albyOAuthSvc.createEncryptedChannelBackup(&events.StaticChannelsBackupEvent{
		NodeID:   "037e702144c4fa485d42f0f69864e943605823763866cf4bf619d2d2cf2eda420b",
		Channels: []events.ChannelBackup{},
		Monitors: []events.EncodedChannelMonitorBackup{},
	})

	assert.NoError(t, err)
	assert.Equal(t, "channels_v2", encryptedBackup.Description)

	masterKey, err := bip32.NewMasterKey(bip39.NewSeed(mnemonic, ""))
	assert.NoError(t, err)

	appKey, err := masterKey.NewChildKey(bip32.FirstHardenedChild + 128029 /* 🐝 */)
	assert.NoError(t, err)
	encryptedChannelsBackupKey, err := appKey.NewChildKey(bip32.FirstHardenedChild)
	assert.NoError(t, err)

	decrypted, err := config.AesGcmDecryptWithKey(encryptedBackup.Data, encryptedChannelsBackupKey.Key)
	assert.NoError(t, err)

	assert.Equal(t, "{\"node_id\":\"037e702144c4fa485d42f0f69864e943605823763866cf4bf619d2d2cf2eda420b\",\"channels\":[],\"monitors\":[]}\n", decrypted)
}
