package service

import (
	"github.com/getAlby/nostr-wallet-connect/alby"
	"github.com/getAlby/nostr-wallet-connect/config"
	"github.com/getAlby/nostr-wallet-connect/db"
	"github.com/getAlby/nostr-wallet-connect/lnclient"
)

type Service interface {
	GetLNClient() lnclient.LNClient
	GetConfig() config.Config
	StartApp(encryptionKey string) error
	StopApp()
	StopLNClient() error
	StopDb() error
	GetBudgetUsage(appPermission *db.AppPermission) int64
	GetLogFilePath() string
	GetAlbyOAuthSvc() alby.AlbyOAuthService
}
