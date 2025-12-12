package config

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/logger"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type config struct {
	Env        *AppConfig
	db         *gorm.DB
	cache      map[string]map[string]string // key -> encryptionKeyHash -> value
	cacheMutex sync.Mutex
}

const (
	unlockPasswordCheck = "THIS STRING SHOULD MATCH IF PASSWORD IS CORRECT"
)

func NewConfig(env *AppConfig, db *gorm.DB) (*config, error) {
	cfg := &config{
		db:    db,
		cache: map[string]map[string]string{},
	}
	err := cfg.init(env)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *config) init(env *AppConfig) error {
	cfg.Env = env

	if cfg.Env.Relay != "" {
		err := cfg.SetUpdate("Relay", cfg.Env.Relay, "")
		if err != nil {
			return err
		}
	}
	if cfg.Env.LNBackendType != "" {
		err := cfg.SetIgnore("LNBackendType", cfg.Env.LNBackendType, "")
		if err != nil {
			return err
		}
	}

	// LND specific to support env variables
	if cfg.Env.LNDAddress != "" {
		err := cfg.SetUpdate("LNDAddress", cfg.Env.LNDAddress, "")
		if err != nil {
			return err
		}
	}
	if cfg.Env.LNDCertFile != "" {
		certBytes, err := os.ReadFile(cfg.Env.LNDCertFile)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to read LND cert file")
			return err
		}
		certHex := hex.EncodeToString(certBytes)
		err = cfg.SetUpdate("LNDCertHex", certHex, "")
		if err != nil {
			return err
		}
	} else if cfg.Env.LNBackendType == "LND" {
		// If no LNDCertFile is provided, clear any stored certificate
		// hex value so that no certificate is used for TLS verification.
		err := cfg.SetUpdate("LNDCertHex", "", "")
		if err != nil {
			return err
		}
	}
	if cfg.Env.LNDMacaroonFile != "" {
		macBytes, err := os.ReadFile(cfg.Env.LNDMacaroonFile)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to read LND macaroon file")
			return err
		}
		macHex := hex.EncodeToString(macBytes)
		err = cfg.SetUpdate("LNDMacaroonHex", macHex, "")
		if err != nil {
			return err
		}
	}
	// Phoenix specific to support env variables
	if cfg.Env.PhoenixdAddress != "" {
		err := cfg.SetIgnore("PhoenixdAddress", cfg.Env.PhoenixdAddress, "")
		if err != nil {
			return err
		}
	}
	if cfg.Env.PhoenixdAuthorization != "" {
		err := cfg.SetIgnore("PhoenixdAuthorization", cfg.Env.PhoenixdAuthorization, "")
		if err != nil {
			return err
		}
	}

	// set the JWT secret from the env, or generate a new one
	existingSecret, _ := cfg.Get("JWTSecret", "")
	if existingSecret == "" {
		jwtSecret := cfg.Env.JWTSecret
		if jwtSecret == "" {
			hexSecret, err := randomHex(32)
			if err != nil {
				logger.Logger.WithError(err).Error("failed to generate JWT secret")
				return err
			}
			jwtSecret = hexSecret
			logger.Logger.Info("Generated new JWT secret")
		}

		err := cfg.SetIgnore("JWTSecret", jwtSecret, "")
		if err != nil {
			logger.Logger.WithError(err).Error("failed to save JWT secret")
			return err
		}
	}
	return nil
}

func (cfg *config) SetupCompleted() bool {
	// TODO: remove hasLdkDir check after 2025/01/01
	// to give time for users to update to 1.6.0+
	nodeLastStartTime, _ := cfg.Get("NodeLastStartTime", "")
	ldkDir, err := os.Stat(path.Join(cfg.GetEnv().Workdir, "ldk"))
	hasLdkDir := err == nil && ldkDir != nil && ldkDir.IsDir()

	logger.Logger.WithFields(logrus.Fields{
		"has_ldk_dir":              hasLdkDir,
		"has_node_last_start_time": nodeLastStartTime != "",
	}).Debug("Checking if setup is completed")
	return nodeLastStartTime != "" || hasLdkDir
}

func (cfg *config) GetJWTSecret() string {
	secret, err := cfg.Get("JWTSecret", "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to retrieve JWTSecret from database")
		return ""
	}
	return secret
}

func (cfg *config) GetRelayUrls() []string {
	relayUrls, _ := cfg.Get("Relay", "")
	return strings.Split(relayUrls, ",")
}

func (cfg *config) GetNetwork() string {
	env := cfg.GetEnv()

	if env.Network != "" {
		return env.Network
	}

	if env.LDKNetwork != "" {
		return env.LDKNetwork
	}

	return "bitcoin"
}

func (cfg *config) GetMempoolUrl() string {
	mempoolApiUrl := cfg.GetEnv().MempoolApi
	return strings.TrimSuffix(mempoolApiUrl, "/api")
}

func (cfg *config) getEncryptionKeyHash(encryptionKey string) string {
	if encryptionKey == "" {
		return ""
	}
	hash := sha256.Sum256([]byte(encryptionKey))
	// For cache key purposes, 8 bytes (16 hex chars) provides:
    //   2^64 possible values = ~18 quintillion combinations
    //   More than sufficient to avoid collisions for cache keys
	return hex.EncodeToString(hash[:8])
}

func (cfg *config) Get(key string, encryptionKey string) (string, error) {
	cfg.cacheMutex.Lock()
	defer cfg.cacheMutex.Unlock()

	encKeyHash := cfg.getEncryptionKeyHash(encryptionKey)

	if keyCache, ok := cfg.cache[key]; ok {
		if cachedValue, ok := keyCache[encKeyHash]; ok {
			logger.Logger.WithField("key", key).Debug("hit config cache")
			return cachedValue, nil
		}
	}
	logger.Logger.WithField("key", key).Debug("missed config cache")

	value, err := cfg.get(key, encryptionKey, cfg.db)
	if err != nil {
		return "", err
	}

	if cfg.cache[key] == nil {
		cfg.cache[key] = make(map[string]string)
	}
	cfg.cache[key][encKeyHash] = value
	logger.Logger.WithField("key", key).Debug("set config cache")
	return value, nil
}

func (cfg *config) get(key string, encryptionKey string, gormDB *gorm.DB) (string, error) {
	var userConfig db.UserConfig
	err := gormDB.Where(&db.UserConfig{Key: key}).Limit(1).Find(&userConfig).Error
	if err != nil {
		return "", fmt.Errorf("failed to get configuration value: %w", gormDB.Error)
	}

	value := userConfig.Value
	if userConfig.Value != "" && encryptionKey != "" && userConfig.Encrypted {
		decrypted, err := AesGcmDecryptWithPassword(value, encryptionKey)
		if err != nil {
			return "", err
		}
		value = decrypted
	}
	return value, nil
}

func (cfg *config) set(key string, value string, clauses clause.OnConflict, encryptionKey string, gormDB *gorm.DB) error {
	if encryptionKey != "" {
		encrypted, err := AesGcmEncryptWithPassword(value, encryptionKey)
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

	logger.Logger.WithField("key", key).Debug("clearing config cache")
	cfg.cacheMutex.Lock()
	defer cfg.cacheMutex.Unlock()
	delete(cfg.cache, key)

	return nil
}

func (cfg *config) SetIgnore(key string, value string, encryptionKey string) error {
	clauses := clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoNothing: true,
	}
	err := cfg.set(key, value, clauses, encryptionKey, cfg.db)
	if err != nil {
		logger.Logger.WithField("key", key).WithError(err).Error("Failed to set config key with ignore", err)
		return err
	}
	return nil
}

func (cfg *config) SetUpdate(key string, value string, encryptionKey string) error {
	clauses := clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "encrypted"}),
	}
	err := cfg.set(key, value, clauses, encryptionKey, cfg.db)
	if err != nil {
		logger.Logger.WithField("key", key).WithError(err).Error("Failed to set config key with update", err)
		return err
	}
	return nil
}

func (cfg *config) ChangeUnlockPassword(currentUnlockPassword string, newUnlockPassword string) error {
	if newUnlockPassword == "" {
		return errors.New("new unlock password must not be empty")
	}
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

		newSecret, err := randomHex(32)
		if err != nil {
			logger.Logger.WithError(err).Error("failed to generate new JWT secret during password change transaction")
			return fmt.Errorf("failed to generate new JWT secret: %w", err)
		}

		updateClauses := clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoUpdates: clause.AssignmentColumns([]string{"value"}),
		}
		err = cfg.set("JWTSecret", newSecret, updateClauses, "", tx)
		if err != nil {
			logger.Logger.WithError(err).Error("failed to save new JWT secret during password change transaction")
			return fmt.Errorf("failed to save new JWT secret: %w", err)
		}
		logger.Logger.Info("Successfully regenerated JWT secret as part of password change transaction")

		return nil
	})

	if err != nil {
		logger.Logger.WithError(err).Error("failed to execute password change transaction")
		return err
	}
	return nil
}

func (cfg *config) SetAutoUnlockPassword(unlockPassword string) error {
	if unlockPassword != "" && !cfg.CheckUnlockPassword(unlockPassword) {
		return errors.New("incorrect password")
	}

	err := cfg.SetUpdate("AutoUnlockPassword", unlockPassword, "")
	if err != nil {
		logger.Logger.WithError(err).Error("failed to update auto unlock password")
		return err
	}

	return nil
}

func (cfg *config) CheckUnlockPassword(encryptionKey string) bool {
	decryptedValue, err := cfg.Get("UnlockPasswordCheck", encryptionKey)

	return err == nil && (decryptedValue == "" || decryptedValue == unlockPasswordCheck)
}

func (cfg *config) SaveUnlockPasswordCheck(encryptionKey string) error {
	err := cfg.SetUpdate("UnlockPasswordCheck", unlockPasswordCheck, encryptionKey)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to save unlock password check to config")
		return err
	}
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

const defaultCurrency = "USD"
const defaultBitcoinDisplayFormat = constants.BITCOIN_DISPLAY_FORMAT_BIP177

func (cfg *config) GetCurrency() string {
	currency, err := cfg.Get("Currency", "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch currency")
		return defaultCurrency
	}
	if currency == "" {
		return defaultCurrency
	}
	return currency
}

func (cfg *config) SetCurrency(value string) error {
	if value == "" {
		return errors.New("currency value cannot be empty")
	}
	err := cfg.SetUpdate("Currency", value, "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to update currency")
		return err
	}
	return nil
}

func (cfg *config) GetBitcoinDisplayFormat() string {
	format, err := cfg.Get("BitcoinDisplayFormat", "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch bitcoin display format")
		return defaultBitcoinDisplayFormat
	}
	if format == "" {
		return defaultBitcoinDisplayFormat
	}
	return format
}

func (cfg *config) SetBitcoinDisplayFormat(value string) error {
	if value != constants.BITCOIN_DISPLAY_FORMAT_SATS && value != constants.BITCOIN_DISPLAY_FORMAT_BIP177 {
		return fmt.Errorf("bitcoin display format must be '%s' or '%s'", constants.BITCOIN_DISPLAY_FORMAT_SATS, constants.BITCOIN_DISPLAY_FORMAT_BIP177)
	}
	err := cfg.SetUpdate("BitcoinDisplayFormat", value, "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to update bitcoin display format")
		return err
	}
	return nil
}
