package service

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/adrg/xdg"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"

	"github.com/getAlby/hub/alby"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/service/keys"
	"github.com/getAlby/hub/swaps"
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
	swapsService        swaps.SwapsService
	albySvc             alby.AlbyService
	albyOAuthSvc        alby.AlbyOAuthService
	eventPublisher      events.EventPublisher
	ctx                 context.Context
	wg                  *sync.WaitGroup
	nip47Service        nip47.Nip47Service
	appCancelFn         context.CancelFunc
	keys                keys.Keys
	isRelayReady        atomic.Bool
	startupState        string
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

	if appConfig.LogToFile {
		err = logger.AddFileLogger(appConfig.Workdir)
		if err != nil {
			return nil, err
		}
	}

	err = finishRestoreNode(appConfig.Workdir)
	if err != nil {
		logger.Logger.WithError(err).Error("failed to restore backup")
		return nil, err
	}

	// If DATABASE_URI is a URI or a path, leave it unchanged.
	// If it only contains a filename, prepend the workdir.
	if !strings.HasPrefix(appConfig.DatabaseUri, "file:") {
		databasePath, _ := filepath.Split(appConfig.DatabaseUri)
		if databasePath == "" {
			appConfig.DatabaseUri = filepath.Join(appConfig.Workdir, appConfig.DatabaseUri)
		}
	}

	gormDB, err := db.NewDB(appConfig.DatabaseUri, appConfig.LogDBQueries)
	if err != nil {
		return nil, err
	}

	cfg, err := config.NewConfig(appConfig, gormDB)
	if err != nil {
		return nil, err
	}

	// write auto unlock password from env to user config
	if appConfig.AutoUnlockPassword != "" {
		err = cfg.SetUpdate("AutoUnlockPassword", appConfig.AutoUnlockPassword, "")
		if err != nil {
			return nil, err
		}
	}
	autoUnlockPassword, err := cfg.Get("AutoUnlockPassword", "")
	if err != nil {
		return nil, err
	}

	eventPublisher := events.NewEventPublisher()

	keys := keys.NewKeys()

	albySvc := alby.NewAlbyService(cfg)
	albyOAuthSvc := alby.NewAlbyOAuthService(gormDB, cfg, keys, eventPublisher)

	transactionsSvc := transactions.NewTransactionsService(gormDB, eventPublisher)

	var wg sync.WaitGroup
	svc := &service{
		cfg:                 cfg,
		ctx:                 ctx,
		wg:                  &wg,
		eventPublisher:      eventPublisher,
		albySvc:             albySvc,
		albyOAuthSvc:        albyOAuthSvc,
		nip47Service:        nip47.NewNip47Service(gormDB, cfg, keys, eventPublisher, albyOAuthSvc),
		transactionsService: transactionsSvc,
		db:                  gormDB,
		keys:                keys,
	}

	eventPublisher.RegisterSubscriber(svc.transactionsService)
	eventPublisher.RegisterSubscriber(svc.nip47Service)
	eventPublisher.RegisterSubscriber(svc.albyOAuthSvc)
	eventPublisher.RegisterSubscriber(&paymentForwardedConsumer{
		db: gormDB,
	})

	eventPublisher.Publish(&events.Event{
		Event: "nwc_started",
		Properties: map[string]interface{}{
			"version": version.Tag,
		},
	})

	if appConfig.GoProfilerAddr != "" {
		startProfiler(ctx, appConfig.GoProfilerAddr)
	}

	if autoUnlockPassword != "" {
		nodeLastStartTime, _ := cfg.Get("NodeLastStartTime", "")
		if nodeLastStartTime != "" {
			svc.StartApp(autoUnlockPassword)
		}
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(10 * time.Minute)
				svc.removeExcessEvents()
			}
		}
	}()

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

func finishRestoreNode(workDir string) error {
	restoreDir := filepath.Join(workDir, "restore")
	if restoreDirStat, err := os.Stat(restoreDir); err == nil && restoreDirStat.IsDir() {
		logger.Logger.WithField("restoreDir", restoreDir).Infof("Restore directory found. Finishing Node restore")

		existingFiles, err := os.ReadDir(restoreDir)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to read WORK_DIR")
			return err
		}

		for _, file := range existingFiles {
			if file.Name() != "restore" {
				err = os.RemoveAll(filepath.Join(workDir, file.Name()))
				if err != nil {
					logger.Logger.WithField("filename", file.Name()).WithError(err).Error("Failed to remove file")
					return err
				}
				logger.Logger.WithField("filename", file.Name()).Info("removed file")
			}
		}

		files, err := os.ReadDir(restoreDir)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to read restore directory")
			return err
		}
		for _, file := range files {
			err = os.Rename(filepath.Join(restoreDir, file.Name()), filepath.Join(workDir, file.Name()))
			if err != nil {
				logger.Logger.WithField("filename", file.Name()).WithError(err).Error("Failed to move file")
				return err
			}
			logger.Logger.WithField("filename", file.Name()).Info("copied file from restore directory")
		}
		err = os.RemoveAll(restoreDir)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to remove restore directory")
			return err
		}
		logger.Logger.WithField("restoreDir", restoreDir).Info("removed restore directory")
	}
	return nil
}

func (svc *service) Shutdown() {
	svc.StopApp()
	svc.eventPublisher.PublishSync(&events.Event{
		Event: "nwc_stopped",
	})
	db.Stop(svc.db)
}

func (svc *service) GetDB() *gorm.DB {
	return svc.db
}

func (svc *service) GetConfig() config.Config {
	return svc.cfg
}

func (svc *service) GetAlbySvc() alby.AlbyService {
	return svc.albySvc
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

func (svc *service) GetSwapsService() swaps.SwapsService {
	return svc.swapsService
}

func (svc *service) GetKeys() keys.Keys {
	return svc.keys
}

func (svc *service) setRelayReady(ready bool) {
	svc.isRelayReady.Store(ready)
}

func (svc *service) IsRelayReady() bool {
	return svc.isRelayReady.Load()
}

func (svc *service) GetStartupState() string {
	return svc.startupState
}

func (svc *service) removeExcessEvents() {
	logger.Logger.Debug("Cleaning up excess events")

	maxEvents := 1000
	// estimated less than 1 second to delete, it should not lock the DB
	maxEventsToDelete := 5000
	// if we only have a few excess events, don't run the task
	minEventsToDelete := 100

	var events []db.RequestEvent
	err := svc.db.Select("id").Order("id asc").Limit(maxEvents + maxEventsToDelete).Find(&events).Error
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch request events")
	}

	numEventsToDelete := len(events) - maxEvents

	if numEventsToDelete < minEventsToDelete {
		return
	}
	deleteEventsBelowId := events[numEventsToDelete].ID

	logger.Logger.WithFields(logrus.Fields{
		"amount":   numEventsToDelete,
		"below_id": deleteEventsBelowId,
	}).Debug("Removing excess events")

	startTime := time.Now()
	err = svc.db.Exec("delete from request_events where id < ?", deleteEventsBelowId).Error
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"amount":   numEventsToDelete,
			"below_id": deleteEventsBelowId,
		}).Error("Failed to delete excess request events")
		return
	}
	logger.Logger.WithFields(logrus.Fields{
		"amount":           numEventsToDelete,
		"below_id":         deleteEventsBelowId,
		"duration_seconds": time.Since(startTime).Seconds(),
	}).Info("Removed excess events")

	// TODO: REMOVE AFTER 2026-01-01
	// this is needed due to cascading delete previously not working
	err = svc.db.Exec("delete from response_events where request_id < ?", deleteEventsBelowId).Error
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"amount":   numEventsToDelete,
			"below_id": deleteEventsBelowId,
		}).Error("Failed to delete excess response events")
		return
	}
}
