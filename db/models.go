package db

import "time"

type UserConfig struct {
	ID        uint
	Key       string
	Value     string
	Encrypted bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type App struct {
	ID          uint
	Name        string `validate:"required"`
	Description string
	NostrPubkey string `validate:"required"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type AppPermission struct {
	ID            uint
	AppId         uint `validate:"required"`
	App           App
	RequestMethod string `validate:"required"`
	MaxAmount     int
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

type Payment struct {
	ID             uint
	AppId          uint `validate:"required"`
	App            App
	RequestEventId uint `validate:"required"`
	RequestEvent   RequestEvent
	Amount         uint // in sats
	PaymentRequest string
	Preimage       *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type DBService interface {
	CreateApp(name string, pubkey string, maxAmount uint64, budgetRenewal string, expiresAt *time.Time, requestMethods []string) (*App, string, error)
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
