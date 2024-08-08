package config

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/logger"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type config struct {
	Env          *AppConfig
	CookieSecret string
	db           *gorm.DB
}

const (
	unlockPasswordCheck = "THIS STRING SHOULD MATCH IF PASSWORD IS CORRECT"
)

func NewConfig(env *AppConfig, db *gorm.DB) *config {
	cfg := &config{
		db: db,
	}
	cfg.init(env)
	return cfg
}

func (cfg *config) init(env *AppConfig) {
	cfg.Env = env

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
			logger.Logger.Fatalf("Failed to read LND cert file: %v", err)
		}
		certHex := hex.EncodeToString(certBytes)
		cfg.SetUpdate("LNDCertHex", certHex, "")
	}
	if cfg.Env.LNDMacaroonFile != "" {
		macBytes, err := os.ReadFile(cfg.Env.LNDMacaroonFile)
		if err != nil {
			logger.Logger.Fatalf("Failed to read LND macaroon file: %v", err)
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

func (cfg *config) get(key string, encryptionKey string, gormDB *gorm.DB) (string, error) {
	var userConfig db.UserConfig
	err := gormDB.Where(&db.UserConfig{Key: key}).Take(&userConfig).Error
	if err != nil {
		return "", fmt.Errorf("failed to get configuration value: %w", gormDB.Error)
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

func (cfg *config) set(key string, value string, clauses clause.OnConflict, encryptionKey string, gormDB *gorm.DB) error {
	if encryptionKey != "" {
		encrypted, err := AesGcmEncrypt(value, encryptionKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt: %v", err)
		}
		value = encrypted
	}
	userConfig := db.UserConfig{Key: key, Value: value, Encrypted: encryptionKey != ""}
	result := gormDB.Clauses(clauses).Create(&userConfig)

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
		logger.Logger.Fatalf("Failed to save config: %v", err)
	}
}

func (cfg *config) SetUpdate(key string, value string, encryptionKey string) {
	clauses := clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}
	err := cfg.set(key, value, clauses, encryptionKey, cfg.db)
	if err != nil {
		logger.Logger.Fatalf("Failed to save config: %v", err)
	}
}

func (cfg *config) ChangeUnlockPassword(currentUnlockPassword string, newUnlockPassword string) error {
	if !cfg.CheckUnlockPassword(currentUnlockPassword) {
		return errors.New("incorrect password")
	}
	err := cfg.db.Transaction(func(tx *gorm.DB) error {

		var encryptedUserConfigs []db.UserConfig
		err := tx.Where(&db.UserConfig{Encrypted: true}).Find(&encryptedUserConfigs).Error
		if err != nil {
			return err
		}

		logger.Logger.WithField("count", len(encryptedUserConfigs)).Info("Updating encrypted entries")

		for _, userConfig := range encryptedUserConfigs {
			decryptedValue, err := cfg.get(userConfig.Key, currentUnlockPassword, tx)
			if err != nil {
				logger.Logger.WithField("key", userConfig.Key).WithError(err).Error("Failed to decrypt key")
				return err
			}
			clauses := clause.OnConflict{
				Columns:   []clause.Column{{Name: "key"}},
				DoUpdates: clause.AssignmentColumns([]string{"value"}),
			}
			err = cfg.set(userConfig.Key, decryptedValue, clauses, newUnlockPassword, tx)
			if err != nil {
				logger.Logger.WithField("key", userConfig.Key).WithError(err).Error("Failed to encrypt key")
				return err
			}
			logger.Logger.WithField("key", userConfig.Key).Info("re-encrypted key")
		}

		// commit transaction
		return nil
	})

	if err != nil {
		logger.Logger.WithError(err).Error("failed to execute db transaction")
		return err
	}

	return nil
}

func (cfg *config) CheckUnlockPassword(encryptionKey string) bool {
	decryptedValue, err := cfg.Get("UnlockPasswordCheck", encryptionKey)

	return err == nil && (decryptedValue == "" || decryptedValue == unlockPasswordCheck)
}

// TODO: rename
func (cfg *config) Setup(encryptionKey string) {
	cfg.SetUpdate("UnlockPasswordCheck", unlockPasswordCheck, encryptionKey)
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
