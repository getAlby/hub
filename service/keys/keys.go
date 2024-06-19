package keys

import (
	"github.com/getAlby/nostr-wallet-connect/config"
	"github.com/getAlby/nostr-wallet-connect/logger"
	"github.com/nbd-wtf/go-nostr"
)

type Keys interface {
	Init(cfg config.Config, encryptionKey string) error
	// Wallet Service Nostr pubkey
	GetNostrPublicKey() string
	// Wallet Service Nostr secret key
	GetNostrSecretKey() string
}

type keys struct {
	nostrSecretKey string
	nostrPublicKey string
}

func NewKeys() *keys {
	return &keys{}
}

func (keys *keys) Init(cfg config.Config, encryptionKey string) error {
	nostrSecretKey, _ := cfg.Get("NostrSecretKey", encryptionKey)

	if nostrSecretKey == "" {
		nostrSecretKey = nostr.GeneratePrivateKey()
		cfg.SetUpdate("NostrSecretKey", nostrSecretKey, encryptionKey)
	}
	nostrPublicKey, err := nostr.GetPublicKey(nostrSecretKey)
	if err != nil {
		logger.Logger.WithError(err).Error("Error converting nostr privkey to pubkey")
		return err
	}
	keys.nostrSecretKey = nostrSecretKey
	keys.nostrPublicKey = nostrPublicKey
	return nil
}

func (keys *keys) GetNostrPublicKey() string {
	if keys.nostrPublicKey == "" {
		logger.Logger.Fatal("keys not initialized")
	}
	return keys.nostrPublicKey
}

func (keys *keys) GetNostrSecretKey() string {
	if keys.nostrSecretKey == "" {
		logger.Logger.Fatal("keys not initialized")
	}
	return keys.nostrSecretKey
}
