package main

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/getAlby/nostr-wallet-connect/migrations"
	"github.com/getAlby/nostr-wallet-connect/models/db"
	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/nbd-wtf/go-nostr"
	"github.com/orandin/lumberjackrus"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// TODO: move to service.go
func NewService(ctx context.Context) (*Service, error) {
	// Load config from environment variables / .env file
	godotenv.Load(".env")
	cfg := &Config{}
	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, err
	}

	logger := log.New()
	logger.SetFormatter(&log.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(log.InfoLevel)

	hook, err := lumberjackrus.NewHook(
		&lumberjackrus.LogFile{
			Filename: "nwc.general.log",
		},
		log.InfoLevel,
		&log.JSONFormatter{},
		&lumberjackrus.LogFileOpts{
			log.InfoLevel: &lumberjackrus.LogFile{
				Filename:   "./log/nwc-info.log",
				MaxAge:     1,
				MaxBackups: 2,
			},
			log.ErrorLevel: &lumberjackrus.LogFile{
				Filename:   "./log/nwc-error.log",
				MaxAge:     1,
				MaxBackups: 2,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	logger.AddHook(hook)

	var db *gorm.DB
	var sqlDb *sql.DB
	db, err = gorm.Open(sqlite.Open(cfg.DatabaseUri), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	// Enable foreign keys for sqlite
	db.Exec("PRAGMA foreign_keys=ON;")
	sqlDb, err = db.DB()
	if err != nil {
		return nil, err
	}
	sqlDb.SetMaxOpenConns(1)
	sqlDb.SetMaxIdleConns(5)
	sqlDb.SetConnMaxLifetime(time.Duration(1800) * time.Second)

	err = migrations.Migrate(db)
	if err != nil {
		return nil, err
	}

	ctx, _ = signal.NotifyContext(ctx, os.Interrupt)

	var wg sync.WaitGroup
	svc := &Service{
		cfg:    cfg,
		db:     db,
		ctx:    ctx,
		wg:     &wg,
		Logger: logger,
	}

	return svc, nil
}

func (svc *Service) loadConfig() (*db.Config, error) {
	// setup database config from the env on first run
	// after first run, changes to ENV will have no effect for these fields
	// because the database values always take precedence!

	godotenv.Load(".env")
	cfg := &Config{}
	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, err
	}

	var existingDbConfig db.Config
	res := svc.db.Limit(1).Find(&existingDbConfig)

	if res.Error != nil {
		return nil, res.Error
	}

	if res.RowsAffected > 0 {
		// do not overwrite the existing entry
		return &existingDbConfig, nil
	}

	nostrSecretKey := svc.cfg.NostrSecretKey
	if nostrSecretKey == "" {
		nostrSecretKey = nostr.GeneratePrivateKey()
	}

	newDbConfig := db.Config{
		LNBackendType:        svc.cfg.LNBackendType,
		LNDAddress:           svc.cfg.LNDAddress,
		LNDCertHex:           svc.cfg.LNDCertHex,
		LNDMacaroonHex:       svc.cfg.LNDMacaroonHex,
		BreezMnemonic:        svc.cfg.BreezMnemonic,
		BreezAPIKey:          svc.cfg.BreezAPIKey,
		GreenlightInviteCode: svc.cfg.GreenlightInviteCode,
		NostrSecretKey:       nostrSecretKey,
	}
	err := svc.db.Save(&newDbConfig).Error
	if err != nil {
		return nil, err
	}

	return &newDbConfig, nil
}

func (svc *Service) launchLNBackend() error {
	if svc.lnClient != nil {
		err := svc.lnClient.Shutdown()
		if err != nil {
			return err
		}
		svc.lnClient = nil
	}

	dbConfig := db.Config{}
	svc.db.First(&dbConfig)

	if dbConfig.LNBackendType == "" {
		return errors.New("No LNBackendType specified")
	}

	svc.cfg.NostrSecretKey = dbConfig.NostrSecretKey
	identityPubkey, err := nostr.GetPublicKey(svc.cfg.NostrSecretKey)
	if err != nil {
		log.Fatalf("Error converting nostr privkey to pubkey: %v", err)
	}
	// TODO: should not re-set config like this
	svc.cfg.IdentityPubkey = identityPubkey

	svc.Logger.Infof("Launching LN Backend: %s", dbConfig.LNBackendType)
	var lnClient LNClient
	switch dbConfig.LNBackendType {
	case LNDBackendType:
		lnClient, err = NewLNDService(svc, dbConfig.LNDAddress, svc.cfg.LNDCertFile, dbConfig.LNDCertHex, svc.cfg.LNDMacaroonFile, dbConfig.LNDMacaroonHex)
	case BreezBackendType:
		lnClient, err = NewBreezService(dbConfig.BreezMnemonic, dbConfig.BreezAPIKey, dbConfig.GreenlightInviteCode, svc.cfg.BreezWorkdir)
	default:
		svc.Logger.Fatalf("Unsupported LNBackendType: %v", dbConfig.LNBackendType)
	}
	if err != nil {
		svc.Logger.Errorf("Failed to launch LN backend: %v", err)
		return err
	}
	svc.lnClient = lnClient
	return nil
}

func (svc *Service) createFilters() nostr.Filters {
	filter := nostr.Filter{
		Tags:  nostr.TagMap{"p": []string{svc.cfg.IdentityPubkey}},
		Kinds: []int{NIP_47_REQUEST_KIND},
	}
	if svc.cfg.ClientPubkey != "" {
		filter.Authors = []string{svc.cfg.ClientPubkey}
	}
	return []nostr.Filter{filter}
}

func (svc *Service) noticeHandler(notice string) {
	svc.Logger.Infof("Received a notice %s", notice)
}
