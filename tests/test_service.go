package tests

import (
	"github.com/getAlby/hub/apps"
	"os"
	"strconv"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/service/keys"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const testDB = "test.db"

func CreateTestService() (svc *TestService, err error) {
	gormDb, err := db.NewDB(testDB, true)
	if err != nil {
		return nil, err
	}

	mockLn, err := NewMockLn()
	if err != nil {
		return nil, err
	}

	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))

	appConfig := &config.AppConfig{
		Workdir: ".test",
	}

	cfg, err := config.NewConfig(
		appConfig,
		gormDb,
	)
	if err != nil {
		return nil, err
	}

	keys := keys.NewKeys()
	keys.Init(cfg, "")

	eventPublisher := events.NewEventPublisher()

	appsService := apps.NewAppsService(gormDb, eventPublisher, keys)

	return &TestService{
		Cfg:            cfg,
		LNClient:       mockLn,
		EventPublisher: eventPublisher,
		DB:             gormDb,
		Keys:           keys,
		AppsService:    appsService,
	}, nil
}

type TestService struct {
	Keys           keys.Keys
	Cfg            config.Config
	LNClient       lnclient.LNClient
	EventPublisher events.EventPublisher
	AppsService    apps.AppsService
	DB             *gorm.DB
}

func RemoveTestService() {
	os.Remove(testDB)
}
