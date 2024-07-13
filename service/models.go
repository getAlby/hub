package service

import (
	"github.com/getAlby/hub/alby"
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/service/keys"
	"gorm.io/gorm"
)

type Service interface {
	StartApp(encryptionKey string) error
	StopApp()
	Shutdown()

	// TODO: remove getters (currently used by http / wails services)
	GetAlbyOAuthSvc() alby.AlbyOAuthService
	GetEventPublisher() events.EventPublisher
	GetLNClient() lnclient.LNClient
	GetDB() *gorm.DB
	GetConfig() config.Config
	GetKeys() keys.Keys
}
