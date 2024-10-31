package alby

import (
	"testing"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/tests"
	"github.com/stretchr/testify/assert"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

func TestEncryptedBackup(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	mnemonic := "limit reward expect search tissue call visa fit thank cream brave jump"
	unlockPassword := "123"
	svc.Cfg.SetUpdate("Mnemonic", mnemonic, unlockPassword)
	err = svc.Keys.Init(svc.Cfg, unlockPassword)
	assert.Nil(t, err)

	albyOAuthSvc := NewAlbyOAuthService(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)
	encryptedBackup, err := albyOAuthSvc.createEncryptedChannelBackup(&events.StaticChannelsBackupEvent{
		NodeID:   "037e702144c4fa485d42f0f69864e943605823763866cf4bf619d2d2cf2eda420b",
		Channels: []events.ChannelBackup{},
		Monitors: []events.EncodedChannelMonitorBackup{},
	})

	assert.Nil(t, err)
	assert.Equal(t, "channels_v2", encryptedBackup.Description)

	masterKey, err := bip32.NewMasterKey(bip39.NewSeed(mnemonic, ""))
	assert.Nil(t, err)

	appKey, err := masterKey.NewChildKey(0)
	assert.Nil(t, err)
	encryptedChannelsBackupKey, err := appKey.NewChildKey(0)
	assert.Nil(t, err)

	decrypted, err := config.AesGcmDecryptWithKey(encryptedBackup.Data, encryptedChannelsBackupKey.Key)
	assert.Nil(t, err)

	assert.Equal(t, "{\"node_id\":\"037e702144c4fa485d42f0f69864e943605823763866cf4bf619d2d2cf2eda420b\",\"channels\":[],\"monitors\":[]}\n", decrypted)
}
