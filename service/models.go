package service

import (
	"errors"

	"gorm.io/gorm"

	"github.com/getAlby/hub/alby"
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/service/keys"
	"github.com/getAlby/hub/swaps"
	"github.com/getAlby/hub/transactions"
)

var (
	// ErrAppBusy is returned when another start or stop operation is already in progress.
	ErrAppBusy = errors.New("app is busy")
	// ErrAlreadyStarted is returned when the app is already unlocked and running.
	ErrAlreadyStarted = errors.New("app already started")
	// ErrAppNotStarted is returned when trying to stop an app that is not running.
	ErrAppNotStarted = errors.New("app not started")
	// ErrInvalidPassword is returned when the provided unlock password is incorrect.
	ErrInvalidPassword = errors.New("invalid password")
	// ErrIncompleteWalletData is returned when the unlock password check is missing from the database.
	ErrIncompleteWalletData = errors.New("your wallet data is incomplete and cannot be unlocked. Please restore from a backup")
)

type RelayStatus struct {
	Url    string
	Online bool
}

type Service interface {
	StartApp(encryptionKey string) error
	StopApp() error
	// WithStartLock runs fn while holding the start/stop lock,
	// returning ErrAppBusy if a start or stop is already in progress.
	WithStartLock(fn func() error) error
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
