package service

import (
	"context"
	"time"

	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/adrg/xdg"
	"github.com/nbd-wtf/go-nostr"
	"gorm.io/gorm"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"

	"github.com/getAlby/hub/alby"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/service/keys"
	"github.com/getAlby/hub/transactions"
	"github.com/getAlby/hub/version"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/nip47"
	"github.com/getAlby/hub/nip47/models"
)

type service struct {
	cfg config.Config

	db                  *gorm.DB
	lnClient            lnclient.LNClient
	transactionsService transactions.TransactionsService
	albyOAuthSvc        alby.AlbyOAuthService
	eventPublisher      events.EventPublisher
	ctx                 context.Context
	wg                  *sync.WaitGroup
	nip47Service        nip47.Nip47Service
	appCancelFn         context.CancelFunc
	keys                keys.Keys
}

func NewService(ctx context.Context) (*service, error) {
	// Load config from environment variables / .GetEnv() file
	godotenv.Load(".env")
	appConfig := &config.AppConfig{}
	err := envconfig.Process("", appConfig)
	if err != nil {
		return nil, err
	}

	logger.Init(appConfig.LogLevel)
	logger.Logger.Info("AlbyHub " + version.Tag)

	if appConfig.Workdir == "" {
		appConfig.Workdir = filepath.Join(xdg.DataHome, "/albyhub")
		logger.Logger.WithField("workdir", appConfig.Workdir).Info("No workdir specified, using default")
	}
	// make sure workdir exists
	os.MkdirAll(appConfig.Workdir, os.ModePerm)

	err = logger.AddFileLogger(appConfig.Workdir)
	if err != nil {
		return nil, err
	}

	finishRestoreNode(appConfig.Workdir)

	// If DATABASE_URI is a URI or a path, leave it unchanged.
	// If it only contains a filename, prepend the workdir.
	if !strings.HasPrefix(appConfig.DatabaseUri, "file:") {
		databasePath, _ := filepath.Split(appConfig.DatabaseUri)
		if databasePath == "" {
			appConfig.DatabaseUri = filepath.Join(appConfig.Workdir, appConfig.DatabaseUri)
		}
	}

	gormDB, err := db.NewDB(appConfig.DatabaseUri)
	if err != nil {
		return nil, err
	}

	cfg := config.NewConfig(appConfig, gormDB)

	eventPublisher := events.NewEventPublisher()

	if err != nil {
		logger.Logger.WithError(err).Error("Failed to create Alby OAuth service")
		return nil, err
	}

	keys := keys.NewKeys()

	var wg sync.WaitGroup
	svc := &service{
		cfg:                 cfg,
		ctx:                 ctx,
		wg:                  &wg,
		eventPublisher:      eventPublisher,
		albyOAuthSvc:        alby.NewAlbyOAuthService(gormDB, cfg, keys, eventPublisher),
		nip47Service:        nip47.NewNip47Service(gormDB, cfg, keys, eventPublisher),
		transactionsService: transactions.NewTransactionsService(gormDB),
		db:                  gormDB,
		keys:                keys,
	}

	// Note: order is important here: transactions service will update transactions
	// from payment events, which will then be consumed by the NIP-47 service to send notifications
	// TODO: transactions service should fire its own events
	eventPublisher.RegisterSubscriber(svc.transactionsService)
	eventPublisher.RegisterSubscriber(svc.nip47Service)
	eventPublisher.RegisterSubscriber(svc.albyOAuthSvc)

	eventPublisher.Publish(&events.Event{
		Event: "nwc_started",
		Properties: map[string]interface{}{
			"version": version.Tag,
		},
	})

	if appConfig.GoProfilerAddr != "" {
		startProfiler(ctx, appConfig.GoProfilerAddr)
	}

	if appConfig.DdProfilerEnabled {
		startDataDogProfiler(ctx)
	}

	return svc, nil
}

func (svc *service) createFilters(identityPubkey string) nostr.Filters {
	filter := nostr.Filter{
		Tags:  nostr.TagMap{"p": []string{identityPubkey}},
		Kinds: []int{models.REQUEST_KIND},
	}
	return []nostr.Filter{filter}
}

func (svc *service) noticeHandler(notice string) {
	logger.Logger.Infof("Received a notice %s", notice)
}

func (svc *service) StartSubscription(ctx context.Context, sub *nostr.Subscription) error {
	svc.nip47Service.StartNotifier(ctx, sub.Relay, svc.lnClient)

	go func() {
		// block till EOS is received
		<-sub.EndOfStoredEvents
		logger.Logger.Debug("Received EOS")

		// loop through incoming events
		for event := range sub.Events {
			go svc.nip47Service.HandleEvent(ctx, sub.Relay, event, svc.lnClient)
		}
		logger.Logger.Debug("Relay subscription events channel ended")
	}()

	<-ctx.Done()

	if sub.Relay.ConnectionError != nil {
		logger.Logger.WithField("connectionError", sub.Relay.ConnectionError).Error("Relay error")
		return sub.Relay.ConnectionError
	}
	logger.Logger.Info("Exiting subscription...")
	return nil
}

func finishRestoreNode(workDir string) {
	restoreDir := filepath.Join(workDir, "restore")
	if restoreDirStat, err := os.Stat(restoreDir); err == nil && restoreDirStat.IsDir() {
		logger.Logger.WithField("restoreDir", restoreDir).Infof("Restore directory found. Finishing Node restore")

		existingFiles, err := os.ReadDir(restoreDir)
		if err != nil {
			logger.Logger.WithError(err).Fatal("Failed to read WORK_DIR")
		}

		for _, file := range existingFiles {
			if file.Name() != "restore" {
				err = os.RemoveAll(filepath.Join(workDir, file.Name()))
				if err != nil {
					logger.Logger.WithField("filename", file.Name()).WithError(err).Fatal("Failed to remove file")
				}
				logger.Logger.WithField("filename", file.Name()).Info("removed file")
			}
		}

		files, err := os.ReadDir(restoreDir)
		if err != nil {
			logger.Logger.WithError(err).Fatal("Failed to read restore directory")
		}
		for _, file := range files {
			err = os.Rename(filepath.Join(restoreDir, file.Name()), filepath.Join(workDir, file.Name()))
			if err != nil {
				logger.Logger.WithField("filename", file.Name()).WithError(err).Fatal("Failed to move file")
			}
			logger.Logger.WithField("filename", file.Name()).Info("copied file from restore directory")
		}
		err = os.RemoveAll(restoreDir)
		if err != nil {
			logger.Logger.WithError(err).Fatal("Failed to remove restore directory")
		}
		logger.Logger.WithField("restoreDir", restoreDir).Info("removed restore directory")
	}
}

func (svc *service) Shutdown() {
	svc.StopApp()
	svc.eventPublisher.Publish(&events.Event{
		Event: "nwc_stopped",
	})
	db.Stop(svc.db)
	// wait for any remaining events
	time.Sleep(1 * time.Second)
}

func (svc *service) GetDB() *gorm.DB {
	return svc.db
}

func (svc *service) GetConfig() config.Config {
	return svc.cfg
}

func (svc *service) GetAlbyOAuthSvc() alby.AlbyOAuthService {
	return svc.albyOAuthSvc
}

func (svc *service) GetNip47Service() nip47.Nip47Service {
	return svc.nip47Service
}

func (svc *service) GetEventPublisher() events.EventPublisher {
	return svc.eventPublisher
}

func (svc *service) GetLNClient() lnclient.LNClient {
	return svc.lnClient
}

func (svc *service) GetTransactionsService() transactions.TransactionsService {
	return svc.transactionsService
}

func (svc *service) GetKeys() keys.Keys {
	return svc.keys
}
