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

func TestExistingEncryptedBackup(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	mnemonic := "limit reward expect search tissue call visa fit thank cream brave jump"
	unlockPassword := "123"
	svc.Cfg.SetUpdate("Mnemonic", mnemonic, unlockPassword)
	err = svc.Keys.Init(svc.Cfg, unlockPassword)
	assert.Nil(t, err)

	encryptedBackup := "69defec3014a6ab9c6abc612-44266b33b8bfaa3bdee45c03dc60b659416ad957d101068607ed120c27573b6d61ad63fcfddf427d0a4f0a1e488585e57479183acb45cd7f26663d1f2de9c154b84f68b9f01f420e1b6f6ce6ae31d89f327a5b393ff49c3456994355a22fd965725523f37c393afc369001dcaf46ef2d8ef062f4bb17edc263985dfca4"

	masterKey, err := bip32.NewMasterKey(bip39.NewSeed(mnemonic, ""))
	assert.Nil(t, err)

	appKey, err := masterKey.NewChildKey(0)
	assert.Nil(t, err)
	encryptedChannelsBackupKey, err := appKey.NewChildKey(0)
	assert.Nil(t, err)

	decrypted, err := config.AesGcmDecryptWithKey(encryptedBackup, encryptedChannelsBackupKey.Key)
	assert.Nil(t, err)

	assert.Equal(t, "{\"node_id\":\"037e702144c4fa485d42f0f69864e943605823763866cf4bf619d2d2cf2eda420b\",\"channels\":[],\"monitors\":[]}\n", decrypted)
}

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
