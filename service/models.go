package service

import (
	"gorm.io/gorm"

	"github.com/getAlby/hub/alby"
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/service/keys"
	"github.com/getAlby/hub/swaps"
	"github.com/getAlby/hub/transactions"
)

type Service interface {
	StartApp(encryptionKey string) error
	StopApp()
	StartAutoSwap() error
	Shutdown()

	// TODO: remove getters (currently used by http / wails services)
	GetAlbyOAuthSvc() alby.AlbyOAuthService
	GetEventPublisher() events.EventPublisher
	GetLNClient() lnclient.LNClient
	GetTransactionsService() transactions.TransactionsService
	GetSwapsService() swaps.SwapsService
	GetDB() *gorm.DB
	GetConfig() config.Config
	GetKeys() keys.Keys
	IsRelayReady() bool
	GetStartupState() string
}
