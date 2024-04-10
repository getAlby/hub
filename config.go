package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"github.com/getAlby/nostr-wallet-connect/models/config"
	dbModels "github.com/getAlby/nostr-wallet-connect/models/db"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Config struct {
	Env            *config.AppConfig
	CookieSecret   string
	NostrSecretKey string
	NostrPublicKey string
	db             *gorm.DB
	logger         *logrus.Logger
}

const (
	unlockPasswordCheck = "THIS STRING SHOULD MATCH IF PASSWORD IS CORRECT"
)

func (cfg *Config) Init(db *gorm.DB, env *config.AppConfig, logger *logrus.Logger) {
	cfg.db = db
	cfg.Env = env
	cfg.logger = logger

	if cfg.Env.Relay != "" {
		cfg.SetUpdate("Relay", cfg.Env.Relay, "")
	}
	if cfg.Env.LNBackendType != "" {
		cfg.SetUpdate("LNBackendType", cfg.Env.LNBackendType, "")
	}
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

func (cfg *Config) Get(key string, encryptionKey string) (string, error) {
	return cfg.get(key, encryptionKey, cfg.db)
}

func (cfg *Config) get(key string, encryptionKey string, db *gorm.DB) (string, error) {
	var userConfig dbModels.UserConfig
	db.Where(&dbModels.UserConfig{Key: key}).Limit(1).Find(&userConfig)

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

func (cfg *Config) set(key string, value string, clauses clause.OnConflict, encryptionKey string, db *gorm.DB) error {
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

func (cfg *Config) SetIgnore(key string, value string, encryptionKey string) {
	clauses := clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoNothing: true,
	}
	err := cfg.set(key, value, clauses, encryptionKey, cfg.db)
	if err != nil {
		cfg.logger.Fatalf("Failed to save config: %v", err)
	}
}

func (cfg *Config) SetUpdate(key string, value string, encryptionKey string) {
	clauses := clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}
	err := cfg.set(key, value, clauses, encryptionKey, cfg.db)
	if err != nil {
		cfg.logger.Fatalf("Failed to save config: %v", err)
	}
}

func (cfg *Config) ChangeUnlockPassword(currentUnlockPassword string, newUnlockPassword string) error {
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

func (cfg *Config) CheckUnlockPassword(encryptionKey string) bool {
	decryptedValue, err := cfg.Get("UnlockPasswordCheck", encryptionKey)

	return err == nil && (decryptedValue == "" || decryptedValue == unlockPasswordCheck)
}

func (cfg *Config) SavePasswordCheck(encryptionKey string) {
	cfg.SetUpdate("UnlockPasswordCheck", unlockPasswordCheck, encryptionKey)
}

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
