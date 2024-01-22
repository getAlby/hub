package main

import (
	"crypto/rand"
	"encoding/hex"
	"os"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	LNDBackendType   = "LND"
	BreezBackendType = "BREEZ"
	CookieName       = "alby_nwc_session"
)

type AppConfig struct {
	Relay           string `envconfig:"RELAY" default:"wss://relay.getalby.com/v1"`
	PublicRelay     string `envconfig:"PUBLIC_RELAY"`
	LNDCertFile     string `envconfig:"LND_CERT_FILE"`
	LNDMacaroonFile string `envconfig:"LND_MACAROON_FILE"`
	Workdir         string `envconfig:"WORK_DIR" default:".data"`
	Port            string `envconfig:"PORT" default:"8080"`
	DatabaseUri     string `envconfig:"DATABASE_URI" default:"nostr-wallet-connect.db"`
	CookieSecret    string `envconfig:"COOKIE_SECRET"`
}

type UserConfig struct {
	ID        uint   `gorm:"primaryKey"`
	Key       string `gorm:"unique;uniqueIndex;not null"`
	Value     string
	Encrypted bool
}

type Config struct {
	AppConfig
	NostrSecretKey string
	NostrPublicKey string
	db             *gorm.DB
}

func (cfg *Config) Init(db *gorm.DB) {
	cfg.db = db

	if cfg.Relay != "" {
		cfg.SetUpdate("Relay", cfg.Relay, "")
	}
	if cfg.LNDCertFile != "" {
		certBytes, err := os.ReadFile(cfg.LNDCertFile)
		if err != nil {
			certHex := hex.EncodeToString(certBytes)
			cfg.SetUpdate("LNDCertHex", certHex, "")
		}
	}
	if cfg.LNDMacaroonFile != "" {
		macBytes, err := os.ReadFile(cfg.LNDMacaroonFile)
		if err != nil {
			macHex := hex.EncodeToString(macBytes)
			cfg.SetUpdate("LNDMacaroonHex", macHex, "")
		}
	}
	if cfg.CookieSecret == "" {
		hex, err := randomHex(20)
		if err == nil {
			cfg.SetIgnore("CookieSecret", hex, "")
		}
		cfg.CookieSecret, _ = cfg.Get("CookieSecret", "")
	}
}

func (cfg *Config) Get(key string, encryptionKey string) (string, error) {
	var userConfig UserConfig
	cfg.db.Where("key = ?", key).Limit(1).Find(&userConfig)

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

func (cfg *Config) set(key string, value string, clauses clause.OnConflict, encryptionKey string) bool {
	if encryptionKey != "" {
		decrypted, err := AesGcmEncrypt(value, encryptionKey)
		if err == nil {
			value = decrypted
		}
	}
	userConfig := UserConfig{Key: key, Value: value, Encrypted: encryptionKey != ""}
	result := cfg.db.Clauses(clauses).Create(&userConfig)

	return result.Error == nil
}

func (cfg *Config) SetIgnore(key string, value string, encryptionKey string) bool {
	clauses := clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoNothing: true,
	}
	return cfg.set(key, value, clauses, encryptionKey)
}

func (cfg *Config) SetUpdate(key string, value string, encryptionKey string) bool {
	clauses := clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}
	return cfg.set(key, value, clauses, encryptionKey)
}

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
