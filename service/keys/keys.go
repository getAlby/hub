package keys

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"errors"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"

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
	// Swap rescue key derived from master key using BIP-85
	GetSwapMnemonic() string
	// Derives a BIP32 child key from appKey derived child dedicated for app wallet keys
	GetAppWalletKey(childIndex uint) (string, error)
	// Derives a child BIP-32 key from the app key (derived from the mnemonic)
	DeriveKey(path []uint32) (*bip32.Key, error)
	// Derives a BIP32 child key from appKey derived child dedicated for swaps
	GetSwapKey(childIndex uint) (*btcec.PrivateKey, error)
}

type keys struct {
	nostrSecretKey string
	nostrPublicKey string
	appKey         *bip32.Key
	swapKey        *hdkeychain.ExtendedKey
	swapMnemonic   string
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

	if mnemonic == "" {
		// for backends that don't use a mnemonic, create one anyway for deriving keys
		entropy, err := bip39.NewEntropy(128)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to generate entropy for mnemonic")
			return err
		}
		mnemonic, err = bip39.NewMnemonic(entropy)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to generate mnemonic")
			return err
		}
		err = cfg.SetUpdate("Mnemonic", mnemonic, encryptionKey)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to save mnemonic")
			return err
		}
	}

	masterKey, err := bip32.NewMasterKey(bip39.NewSeed(mnemonic, ""))
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to create seed from mnemonic")
		return err
	}

	albyHubIndex := uint32(bip32.FirstHardenedChild + 128029 /* 🐝 */)
	appKey, err := masterKey.NewChildKey(albyHubIndex)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to derive app key")
		return err
	}
	keys.appKey = appKey

	swapMnemonic, err := keys.GenerateSwapMnemonic(masterKey)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to generate swap mnemonic")
		return err
	}
	keys.swapMnemonic = swapMnemonic

	netParams := &chaincfg.MainNetParams
	network := cfg.GetNetwork()
	if network == "testnet" {
		netParams = &chaincfg.TestNet3Params
	}

	swapKey, err := hdkeychain.NewMaster(bip39.NewSeed(swapMnemonic, ""), netParams)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to create seed from swap mnemonic")
		return err
	}
	keys.swapKey = swapKey

	return nil
}

func (keys *keys) GetSwapMnemonic() string {
	return keys.swapMnemonic
}

func (keys *keys) GetNostrPublicKey() string {
	return keys.nostrPublicKey
}

func (keys *keys) GetNostrSecretKey() string {
	return keys.nostrSecretKey
}

func (keys *keys) GetAppWalletKey(appID uint) (string, error) {
	path := []uint32{bip32.FirstHardenedChild + 1, bip32.FirstHardenedChild + uint32(appID)}
	key, err := keys.DeriveKey(path)
	if err != nil {
		return "", err
	}
	childPrivKey, _ := btcec.PrivKeyFromBytes(key.Key)
	return hex.EncodeToString(childPrivKey.Serialize()), nil
}

func (keys *keys) DeriveKey(path []uint32) (*bip32.Key, error) {
	if len(path) == 0 {
		return nil, errors.New("path must have at least one element")
	}
	if keys.appKey == nil {
		return nil, errors.New("app key not set")
	}
	key := keys.appKey
	for _, index := range path {
		var err error
		key, err = key.NewChildKey(index)
		if err != nil {
			return nil, err
		}
	}

	return key, nil
}

func (keys *keys) GetSwapKey(swapID uint) (*btcec.PrivateKey, error) {
	path := []uint32{44, 0, 0, 0, uint32(swapID)}

	key := keys.swapKey
	for _, index := range path {
		var err error
		key, err = key.Derive(index)
		if err != nil {
			return nil, err
		}
	}

	return key.ECPrivKey()
}

// Taken from https://github.com/e4coder/bip85/blob/main/bip85.go
func (keys *keys) GenerateSwapMnemonic(key *bip32.Key) (string, error) {
	swapIndex := uint32(bip32.FirstHardenedChild + 128260 /* 🔄 */)
	path := []uint32{bip32.FirstHardenedChild + 83696968, bip32.FirstHardenedChild + 39, bip32.FirstHardenedChild + 0, bip32.FirstHardenedChild + 12, swapIndex}

	for _, index := range path {
		var err error
		key, err = key.NewChildKey(index)
		if err != nil {
			return "", err
		}
	}

	hash := hmac.New(sha512.New, []byte("bip-entropy-from-k"))
	_, err := hash.Write(key.Key)
	if err != nil {
		return "", err
	}
	entropy := hash.Sum(nil)
	entropyBIP39 := entropy[:16]
	mnemonic, err := bip39.NewMnemonic(entropyBIP39)
	if err != nil {
		return "", err
	}
	return mnemonic, nil
}
