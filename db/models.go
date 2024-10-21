package db

import (
	"time"

	"gorm.io/datatypes"
)

type UserConfig struct {
	ID        uint
	Key       string
	Value     string
	Encrypted bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type App struct {
	ID           uint
	Name         string `validate:"required"`
	Description  string
	NostrPubkey  string `validate:"required"`
	WalletPubkey string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Isolated     bool
	Metadata     datatypes.JSON
}

type AppPermission struct {
	ID            uint
	AppId         uint `validate:"required"`
	App           App
	Scope         string `validate:"required"`
	MaxAmountSat  int
	BudgetRenewal string
	ExpiresAt     *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type RequestEvent struct {
	ID          uint
	AppId       *uint
	App         App
	NostrId     string `validate:"required"`
	ContentData string
	Method      string
	State       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ResponseEvent struct {
	ID        uint
	NostrId   string `validate:"required"`
	RequestId uint   `validate:"required"`
	State     string
	RepliedAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Transaction struct {
	ID              uint
	AppId           *uint
	App             *App
	RequestEventId  *uint
	RequestEvent    *RequestEvent
	Type            string
	State           string
	AmountMsat      uint64
	FeeMsat         uint64
	FeeReserveMsat  uint64
	PaymentRequest  string
	PaymentHash     string
	Description     string
	DescriptionHash string
	Preimage        *string
	CreatedAt       time.Time
	ExpiresAt       *time.Time
	UpdatedAt       time.Time
	SettledAt       *time.Time
	Metadata        datatypes.JSON
	SelfPayment     bool
	Boostagram      datatypes.JSON
	FailureReason   string
}

type DBService interface {
	CreateApp(name string, pubkey string, maxAmountSat uint64, budgetRenewal string, expiresAt *time.Time, scopes []string, isolated bool, metadata map[string]interface{}, walletChildPrivKeyGeneratorFunc func(uint32) (string, error)) (*App, string, error)
	DeleteApp(app *App) error
}

const (
	REQUEST_EVENT_STATE_HANDLER_EXECUTING = "executing"
	REQUEST_EVENT_STATE_HANDLER_EXECUTED  = "executed"
	REQUEST_EVENT_STATE_HANDLER_ERROR     = "error"
)
const (
	RESPONSE_EVENT_STATE_PUBLISH_CONFIRMED   = "confirmed"
	RESPONSE_EVENT_STATE_PUBLISH_FAILED      = "failed"
	RESPONSE_EVENT_STATE_PUBLISH_UNCONFIRMED = "unconfirmed"
)
