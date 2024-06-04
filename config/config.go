package config

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	dbModels "github.com/getAlby/nostr-wallet-connect/db"
)

type config struct {
	Env            *AppConfig
	CookieSecret   string
	NostrSecretKey string
	NostrPublicKey string
	db             *gorm.DB
	logger         *logrus.Logger
}

const (
	unlockPasswordCheck = "THIS STRING SHOULD MATCH IF PASSWORD IS CORRECT"
)

func NewConfig(db *gorm.DB, env *AppConfig, logger *logrus.Logger) *config {
	cfg := &config{}
	cfg.init(db, env, logger)
	return cfg
}

func (cfg *config) init(db *gorm.DB, env *AppConfig, logger *logrus.Logger) {
	cfg.db = db
	cfg.Env = env
	cfg.logger = logger

	if cfg.Env.Relay != "" {
		cfg.SetUpdate("Relay", cfg.Env.Relay, "")
	}
	if cfg.Env.LNBackendType != "" {
		cfg.SetUpdate("LNBackendType", cfg.Env.LNBackendType, "")
	}

	// LND specific to support env variables
	if cfg.Env.LNDAddress != "" {
		cfg.SetUpdate("LNDAddress", cfg.Env.LNDAddress, "")
	}
	if cfg.Env.LNDCertFile != "" {
		certBytes, err := os.ReadFile(cfg.Env.LNDCertFile)
		if err != nil {
			logger.Fatalf("Failed to read LND cert file: %v", err)
		}
		certHex := hex.EncodeToString(certBytes)
		cfg.SetUpdate("LNDCertHex", certHex, "")
	}
	if cfg.Env.LNDMacaroonFile != "" {
		macBytes, err := os.ReadFile(cfg.Env.LNDMacaroonFile)
		if err != nil {
			logger.Fatalf("Failed to read LND macaroon file: %v", err)
		}
		macHex := hex.EncodeToString(macBytes)
		cfg.SetUpdate("LNDMacaroonHex", macHex, "")
	}
	// Phoenix specific to support env variables
	if cfg.Env.PhoenixdAddress != "" {
		cfg.SetUpdate("PhoenixdAddress", cfg.Env.PhoenixdAddress, "")
	}
	if cfg.Env.PhoenixdAuthorization != "" {
		cfg.SetUpdate("PhoenixdAuthorization", cfg.Env.PhoenixdAuthorization, "")
	}

	// set the cookie secret to the one from the env
	// if no cookie secret is configured we create a random one and store it in the DB
	cfg.CookieSecret = cfg.Env.CookieSecret
	if cfg.CookieSecret == "" {
		hex, err := randomHex(20)
		if err == nil {
			cfg.SetIgnore("CookieSecret", hex, "")
		}
		cfg.CookieSecret, _ = cfg.Get("CookieSecret", "")
	}
}

func (cfg *config) GetNostrPublicKey() string {
	return cfg.NostrPublicKey
}

func (cfg *config) GetNostrSecretKey() string {
	return cfg.NostrSecretKey
}

func (cfg *config) GetCookieSecret() string {
	return cfg.CookieSecret
}

func (cfg *config) GetRelayUrl() string {
	relayUrl, _ := cfg.Get("Relay", "")
	return relayUrl
}

func (cfg *config) Get(key string, encryptionKey string) (string, error) {
	return cfg.get(key, encryptionKey, cfg.db)
}

func (cfg *config) get(key string, encryptionKey string, db *gorm.DB) (string, error) {
	var userConfig dbModels.UserConfig
	err := db.Where(&dbModels.UserConfig{Key: key}).Limit(1).Find(&userConfig).Error
	if err != nil {
		return "", fmt.Errorf("failed to get configuration value: %w", db.Error)
	}

	value := userConfig.Value
	if userConfig.Value != "" && encryptionKey != "" && userConfig.Encrypted {
		decrypted, err := AesGcmDecrypt(value, encryptionKey)
		if err != nil {
			return "", err
		}
		value = decrypted
	}
	return value, nil
}

func (cfg *config) set(key string, value string, clauses clause.OnConflict, encryptionKey string, db *gorm.DB) error {
	if encryptionKey != "" {
		encrypted, err := AesGcmEncrypt(value, encryptionKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt: %v", err)
		}
		value = encrypted
	}
	userConfig := dbModels.UserConfig{Key: key, Value: value, Encrypted: encryptionKey != ""}
	result := db.Clauses(clauses).Create(&userConfig)

	if result.Error != nil {
		return fmt.Errorf("failed to save key to config: %v", result.Error)
	}
	return nil
}

func (cfg *config) SetIgnore(key string, value string, encryptionKey string) {
	clauses := clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoNothing: true,
	}
	err := cfg.set(key, value, clauses, encryptionKey, cfg.db)
	if err != nil {
		cfg.logger.Fatalf("Failed to save config: %v", err)
	}
}

func (cfg *config) SetUpdate(key string, value string, encryptionKey string) {
	clauses := clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}
	err := cfg.set(key, value, clauses, encryptionKey, cfg.db)
	if err != nil {
		cfg.logger.Fatalf("Failed to save config: %v", err)
	}
}

func (cfg *config) ChangeUnlockPassword(currentUnlockPassword string, newUnlockPassword string) error {
	if !cfg.CheckUnlockPassword(currentUnlockPassword) {
		return errors.New("incorrect password")
	}
	err := cfg.db.Transaction(func(tx *gorm.DB) error {

		var encryptedUserConfigs []dbModels.UserConfig
		err := tx.Where(&dbModels.UserConfig{Encrypted: true}).Find(&encryptedUserConfigs).Error
		if err != nil {
			return err
		}

		cfg.logger.WithField("count", len(encryptedUserConfigs)).Info("Updating encrypted entries")

		for _, userConfig := range encryptedUserConfigs {
			decryptedValue, err := cfg.get(userConfig.Key, currentUnlockPassword, tx)
			if err != nil {
				cfg.logger.WithField("key", userConfig.Key).WithError(err).Error("Failed to decrypt key")
				return err
			}
			clauses := clause.OnConflict{
				Columns:   []clause.Column{{Name: "key"}},
				DoUpdates: clause.AssignmentColumns([]string{"value"}),
			}
			err = cfg.set(userConfig.Key, decryptedValue, clauses, newUnlockPassword, tx)
			if err != nil {
				cfg.logger.WithField("key", userConfig.Key).WithError(err).Error("Failed to encrypt key")
				return err
			}
			cfg.logger.WithField("key", userConfig.Key).Info("re-encrypted key")
		}

		// commit transaction
		return nil
	})

	if err != nil {
		cfg.logger.WithError(err).Error("failed to execute db transaction")
		return err
	}

	return nil
}

func (cfg *config) CheckUnlockPassword(encryptionKey string) bool {
	decryptedValue, err := cfg.Get("UnlockPasswordCheck", encryptionKey)

	return err == nil && (decryptedValue == "" || decryptedValue == unlockPasswordCheck)
}

func (cfg *config) Setup(encryptionKey string) {
	cfg.SetUpdate("UnlockPasswordCheck", unlockPasswordCheck, encryptionKey)
}

func (cfg *config) Start(encryptionKey string) error {
	nostrSecretKey, _ := cfg.Get("NostrSecretKey", encryptionKey)

	if nostrSecretKey == "" {
		nostrSecretKey = nostr.GeneratePrivateKey()
		cfg.SetUpdate("NostrSecretKey", nostrSecretKey, encryptionKey)
	}
	nostrPublicKey, err := nostr.GetPublicKey(nostrSecretKey)
	if err != nil {
		cfg.logger.WithError(err).Error("Error converting nostr privkey to pubkey")
		return err
	}
	cfg.NostrSecretKey = nostrSecretKey
	cfg.NostrPublicKey = nostrPublicKey
	return nil
}

func (cfg *config) GetEnv() *AppConfig {
	return cfg.Env
}

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
