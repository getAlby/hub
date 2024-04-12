package main

import (
	"encoding/json"
	"time"

	"github.com/getAlby/nostr-wallet-connect/models/lnclient"
)

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

// TODO: move to models/db
type App struct {
	ID          uint
	Name        string `validate:"required"`
	Description string
	NostrPubkey string `validate:"required"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TODO: move to models/db
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

// TODO: move to models/db
type RequestEvent struct {
	ID        uint
	AppId     *uint
	App       App
	NostrId   string `validate:"required"`
	Content   string
	State     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TODO: move to models/db
type ResponseEvent struct {
	ID        uint
	NostrId   string `validate:"required"`
	RequestId uint   `validate:"required"`
	Content   string
	State     string
	RepliedAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TODO: move to models/db
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

// TODO: move to models/Nip47
type Nip47Transaction = lnclient.Transaction

type PayRequest struct {
	Invoice string `json:"invoice"`
}

// TODO: move to models/Nip47
type Nip47Request struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type Nip47Response struct {
	Error      *Nip47Error `json:"error,omitempty"`
	Result     interface{} `json:"result,omitempty"`
	ResultType string      `json:"result_type"`
}

type Nip47Notification struct {
	Notification     interface{} `json:"notification,omitempty"`
	NotificationType string      `json:"notification_type"`
}

type Nip47Error struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type Nip47PaymentReceivedNotification struct {
	Nip47Transaction
}

type Nip47PayParams struct {
	Invoice string `json:"invoice"`
}
type Nip47PayResponse struct {
	Preimage string `json:"preimage"`
}

type Nip47MultiPayKeysendParams struct {
	Keysends []Nip47MultiPayKeysendElement `json:"keysends"`
}

type Nip47MultiPayKeysendElement struct {
	Nip47KeysendParams
	Id string `json:"id"`
}

type Nip47MultiPayInvoiceParams struct {
	Invoices []Nip47MultiPayInvoiceElement `json:"invoices"`
}

type Nip47MultiPayInvoiceElement struct {
	Nip47PayParams
	Id string `json:"id"`
}

type Nip47KeysendParams struct {
	Amount     int64                `json:"amount"`
	Pubkey     string               `json:"pubkey"`
	Preimage   string               `json:"preimage"`
	TLVRecords []lnclient.TLVRecord `json:"tlv_records"`
}

type Nip47BalanceResponse struct {
	Balance       int64  `json:"balance"`
	MaxAmount     int    `json:"max_amount"`
	BudgetRenewal string `json:"budget_renewal"`
}

// TODO: move to models/Nip47
type Nip47GetInfoResponse struct {
	Alias       string   `json:"alias"`
	Color       string   `json:"color"`
	Pubkey      string   `json:"pubkey"`
	Network     string   `json:"network"`
	BlockHeight uint32   `json:"block_height"`
	BlockHash   string   `json:"block_hash"`
	Methods     []string `json:"methods"`
}

type Nip47MakeInvoiceParams struct {
	Amount          int64  `json:"amount"`
	Description     string `json:"description"`
	DescriptionHash string `json:"description_hash"`
	Expiry          int64  `json:"expiry"`
}
type Nip47MakeInvoiceResponse struct {
	Nip47Transaction
}

type Nip47LookupInvoiceParams struct {
	Invoice     string `json:"invoice"`
	PaymentHash string `json:"payment_hash"`
}

type Nip47LookupInvoiceResponse struct {
	Nip47Transaction
}

type Nip47ListTransactionsParams struct {
	From   uint64 `json:"from,omitempty"`
	Until  uint64 `json:"until,omitempty"`
	Limit  uint64 `json:"limit,omitempty"`
	Offset uint64 `json:"offset,omitempty"`
	Unpaid bool   `json:"unpaid,omitempty"`
	Type   string `json:"type,omitempty"`
}

type Nip47ListTransactionsResponse struct {
	Transactions []Nip47Transaction `json:"transactions"`
}

type Nip47SignMessageParams struct {
	Message string `json:"message"`
}

type Nip47SignMessageResponse struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
}
