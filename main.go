package main

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"os/signal"
	"path"
	"sync"

	"github.com/getAlby/nostr-wallet-connect/migrations"
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
	appConfig := &AppConfig{}
	err := envconfig.Process("", appConfig)
	if err != nil {
		return nil, err
	}

	logger := log.New()
	logger.SetFormatter(&log.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(log.InfoLevel)

	hook, err := lumberjackrus.NewHook(
		&lumberjackrus.LogFile{
			Filename: path.Join(appConfig.Workdir, "log/nwc-general.log"),
		},
		log.InfoLevel,
		&log.JSONFormatter{},
		&lumberjackrus.LogFileOpts{
			log.InfoLevel: &lumberjackrus.LogFile{
				Filename:   path.Join(appConfig.Workdir, "log/nwc-info.log"),
				MaxAge:     1,
				MaxBackups: 2,
			},
			log.ErrorLevel: &lumberjackrus.LogFile{
				Filename:   path.Join(appConfig.Workdir, "log/nwc-error.log"),
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
	db, err = gorm.Open(sqlite.Open(appConfig.DatabaseUri), &gorm.Config{})
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

	err = migrations.Migrate(db)
	if err != nil {
		logger.Errorf("Failed to migrate: %v", err)
		return nil, err
	}

	ctx, _ = signal.NotifyContext(ctx, os.Interrupt)

	cfg := &Config{}
	cfg.Init(db, appConfig)

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

func (svc *Service) launchLNBackend(encryptionKey string) error {
	if svc.lnClient != nil {
		err := svc.lnClient.Shutdown()
		if err != nil {
			return err
		}
		svc.lnClient = nil
	}

	lndBackend, _ := svc.cfg.Get("LNBackendType", "")
	if lndBackend == "" {
		return errors.New("No LNBackendType specified")
	}

	svc.Logger.Infof("Launching LN Backend: %s", lndBackend)
	var lnClient LNClient
	var err error
	switch lndBackend {
	case LNDBackendType:
		LNDAddress, _ := svc.cfg.Get("LNDAddress", encryptionKey)
		LNDCertHex, _ := svc.cfg.Get("LNDCertHex", encryptionKey)
		LNDMacaroonHex, _ := svc.cfg.Get("LNDMacaroonHex", encryptionKey)

		lnClient, err = NewLNDService(svc, LNDAddress, LNDCertHex, LNDMacaroonHex)
	case lndBackend:
		BreezMnemonic, _ := svc.cfg.Get("BreezMnemonic", encryptionKey)
		BreezAPIKey, _ := svc.cfg.Get("BreezAPIKey", encryptionKey)
		GreenlightInviteCode, _ := svc.cfg.Get("GreenlightInviteCode", encryptionKey)
		BreezWorkdir := path.Join(svc.cfg.Env.Workdir, "breez")

		lnClient, err = NewBreezService(BreezMnemonic, BreezAPIKey, GreenlightInviteCode, BreezWorkdir)
	default:
		svc.Logger.Fatalf("Unsupported LNBackendType: %v", lndBackend)
	}
	if err != nil {
		svc.Logger.Errorf("Failed to launch LN backend: %v", err)
		return err
	}
	svc.lnClient = lnClient
	return nil
}

func (svc *Service) createFilters(identityPubkey string) nostr.Filters {
	filter := nostr.Filter{
		Tags:  nostr.TagMap{"p": []string{identityPubkey}},
		Kinds: []int{NIP_47_REQUEST_KIND},
	}
	return []nostr.Filter{filter}
}

func (svc *Service) noticeHandler(notice string) {
	svc.Logger.Infof("Received a notice %s", notice)
}
