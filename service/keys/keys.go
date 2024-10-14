package keys

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/logger"
	"github.com/nbd-wtf/go-nostr"
	"github.com/tyler-smith/go-bip32"
)

type Keys interface {
	Init(cfg config.Config, encryptionKey string) error
	// Wallet Service Nostr pubkey
	GetNostrPublicKey() string
	// Wallet Service Nostr secret key
	GetNostrSecretKey() string
	// GetBIP32ChildKey derives a BIP32 child key from the nostrSecretKey given a child key index
	GetBIP32ChildKey(childIndex uint32) (*btcec.PrivateKey, error)
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
	return nil
}

func (keys *keys) GetNostrPublicKey() string {
	return keys.nostrPublicKey
}

func (keys *keys) GetNostrSecretKey() string {
	return keys.nostrSecretKey
}

func (keys *keys) GetBIP32ChildKey(childIndex uint32) (*btcec.PrivateKey, error) {
	// Convert nostrSecretKey to btcec private key
	privKeyBytes, err := hex.DecodeString(keys.nostrSecretKey)
	if err != nil {
		return nil, err
	}
	privKey, _ := btcec.PrivKeyFromBytes(privKeyBytes)

	// Create a BIP32 master key from the private key
	masterKey, err := bip32.NewMasterKey(privKey.Serialize())
	if err != nil {
		return nil, err
	}

	// Derive child key
	childKey, err := masterKey.NewChildKey(childIndex)
	if err != nil {
		return nil, err
	}

	// Convert child key to btcec private key
	childPrivKey, _ := btcec.PrivKeyFromBytes(childKey.Key)

	return childPrivKey, nil
}
