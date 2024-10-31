package keys

import (
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/logger"
	"github.com/nbd-wtf/go-nostr"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

type Keys interface {
	Init(cfg config.Config, encryptionKey string) error
	// Wallet Service Nostr pubkey (DEPRECATED)
	GetNostrPublicKey() string
	// Wallet Service Nostr secret key (DEPRECATED)
	GetNostrSecretKey() string

	EncryptChannelBackupData(channelBackupData string) (string, error)
}

type keys struct {
	nostrSecretKey string
	nostrPublicKey string
	appKey         *bip32.Key
}

func NewKeys() *keys {
	return &keys{}
}

func (keys *keys) Init(cfg config.Config, encryptionKey string) error {
	nostrSecretKey, _ := cfg.Get("NostrSecretKey", encryptionKey)

	if nostrSecretKey == "" {
		nostrSecretKey = nostr.GeneratePrivateKey()
		err := cfg.SetUpdate("NostrSecretKey", nostrSecretKey, encryptionKey)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to save generated nostr secret key")
			return err
		}
	}
	nostrPublicKey, err := nostr.GetPublicKey(nostrSecretKey)
	if err != nil {
		logger.Logger.WithError(err).Error("Error converting nostr privkey to pubkey")
		return err
	}
	keys.nostrSecretKey = nostrSecretKey
	keys.nostrPublicKey = nostrPublicKey

	mnemonic, err := cfg.Get("Mnemonic", encryptionKey)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to decrypt mnemonic")
		return err
	}

	if mnemonic != "" {
		masterKey, err := bip32.NewMasterKey(bip39.NewSeed(mnemonic, ""))
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to create seed from mnemonic")
			return err
		}

		APP_INDEX := uint32(0) // TODO: choose an index
		appKey, err := masterKey.NewChildKey(APP_INDEX)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to create seed from mnemonic")
			return err
		}
		keys.appKey = appKey
	}

	return nil
}

func (keys *keys) GetNostrPublicKey() string {
	return keys.nostrPublicKey
}

func (keys *keys) GetNostrSecretKey() string {
	return keys.nostrSecretKey
}

// TODO: move somewhere else
func (keys *keys) EncryptChannelBackupData(channelBackupData string) (string, error) {

	ENCRYPTED_SCB_INDEX := uint32(0) // TODO: choose an index
	encryptedChannelsBackupKey, err := keys.appKey.NewChildKey(ENCRYPTED_SCB_INDEX)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to create seed from mnemonic")
		return "", err
	}

	return config.AesGcmEncryptWithKey(channelBackupData, encryptedChannelsBackupKey.Key)
}
