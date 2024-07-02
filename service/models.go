package service

import (
	"github.com/getAlby/nostr-wallet-connect/alby"
	"github.com/getAlby/nostr-wallet-connect/config"
	"github.com/getAlby/nostr-wallet-connect/events"
	"github.com/getAlby/nostr-wallet-connect/lnclient"
	"github.com/getAlby/nostr-wallet-connect/service/keys"
	"github.com/getAlby/nostr-wallet-connect/transactions"
	"gorm.io/gorm"
)

type Service interface {
	StartApp(encryptionKey string) error
	StopApp()
	StopLNClient() error
	WaitShutdown()

	// TODO: remove getters (currently used by http / wails services)
	GetAlbyOAuthSvc() alby.AlbyOAuthService
	GetEventPublisher() events.EventPublisher
	GetLNClient() lnclient.LNClient
	GetTransactionsService() transactions.TransactionsService
	GetDB() *gorm.DB
	GetConfig() config.Config
	GetKeys() keys.Keys
}
