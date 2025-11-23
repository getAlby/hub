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

type RelayStatus struct {
	Url    string
	Online bool
}

type Service interface {
	StartApp(encryptionKey string) error
	StopApp()
	Shutdown()

	// TODO: remove getters (currently used by http / wails services)
	GetAlbySvc() alby.AlbyService
	GetAlbyOAuthSvc() alby.AlbyOAuthService
	GetEventPublisher() events.EventPublisher
	GetLNClient() lnclient.LNClient
	GetTransactionsService() transactions.TransactionsService
	GetSwapsService() swaps.SwapsService
	GetDB() *gorm.DB
	GetConfig() config.Config
	GetKeys() keys.Keys
	GetRelayStatuses() []RelayStatus
	GetStartupState() string
}
