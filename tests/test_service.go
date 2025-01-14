package tests

import (
	"strconv"
	"testing"

	"github.com/getAlby/hub/apps"
	"github.com/getAlby/hub/tests/db"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/service/keys"
)

func CreateTestService(t *testing.T) (svc *TestService, err error) {
	return CreateTestServiceWithMnemonic(t, "", "")
}

func CreateTestServiceWithMnemonic(t *testing.T, mnemonic string, unlockPassword string) (svc *TestService, err error) {
	gormDb, err := db.NewDB(t)
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

	if mnemonic != "" {
		if err = cfg.SetUpdate("Mnemonic", mnemonic, unlockPassword); err != nil {
			return nil, err
		}
	}

	if err = keys.Init(cfg, unlockPassword); err != nil {
		return nil, err
	}

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

func (s *TestService) Remove() {
	db.CloseDB(s.DB)
}
