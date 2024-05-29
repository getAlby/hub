package nip47

import (
	"encoding/json"

	"github.com/getAlby/nostr-wallet-connect/lnclient"
)

const (
	INFO_EVENT_KIND            = 13194
	REQUEST_KIND               = 23194
	RESPONSE_KIND              = 23195
	NOTIFICATION_KIND          = 23196
	PAY_INVOICE_METHOD         = "pay_invoice"
	GET_BALANCE_METHOD         = "get_balance"
	GET_INFO_METHOD            = "get_info"
	MAKE_INVOICE_METHOD        = "make_invoice"
	LOOKUP_INVOICE_METHOD      = "lookup_invoice"
	LIST_TRANSACTIONS_METHOD   = "list_transactions"
	PAY_KEYSEND_METHOD         = "pay_keysend"
	MULTI_PAY_INVOICE_METHOD   = "multi_pay_invoice"
	MULTI_PAY_KEYSEND_METHOD   = "multi_pay_keysend"
	SIGN_MESSAGE_METHOD        = "sign_message"
	ERROR_INTERNAL             = "INTERNAL"
	ERROR_NOT_IMPLEMENTED      = "NOT_IMPLEMENTED"
	ERROR_QUOTA_EXCEEDED       = "QUOTA_EXCEEDED"
	ERROR_INSUFFICIENT_BALANCE = "INSUFFICIENT_BALANCE"
	ERROR_UNAUTHORIZED         = "UNAUTHORIZED"
	ERROR_EXPIRED              = "EXPIRED"
	ERROR_RESTRICTED           = "RESTRICTED"
	ERROR_BAD_REQUEST          = "BAD_REQUEST"
	OTHER                      = "OTHER"
	CAPABILITIES               = "pay_invoice pay_keysend get_balance get_info make_invoice lookup_invoice list_transactions multi_pay_invoice multi_pay_keysend sign_message notifications"
	NOTIFICATION_TYPES         = "payment_received" // same format as above e.g. "payment_received balance_updated payment_sent channel_opened channel_closed ..."
)

// TODO: move other permissions here (e.g. all payment methods use pay_invoice)
const (
	NOTIFICATIONS_PERMISSION = "notifications"
)

const (
	PAYMENT_RECEIVED_NOTIFICATION = "payment_received"
)

const (
	BUDGET_RENEWAL_DAILY   = "daily"
	BUDGET_RENEWAL_WEEKLY  = "weekly"
	BUDGET_RENEWAL_MONTHLY = "monthly"
	BUDGET_RENEWAL_YEARLY  = "yearly"
	BUDGET_RENEWAL_NEVER   = "never"
)

type Transaction = lnclient.Transaction

type PayRequest struct {
	Invoice string `json:"invoice"`
}

type Request struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type Response struct {
	Error      *Error      `json:"error,omitempty"`
	Result     interface{} `json:"result,omitempty"`
	ResultType string      `json:"result_type"`
}

type Notification struct {
	Notification     interface{} `json:"notification,omitempty"`
	NotificationType string      `json:"notification_type"`
}

type Error struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type PaymentReceivedNotification struct {
	Transaction
}

type PayParams struct {
	Invoice string `json:"invoice"`
}
type PayResponse struct {
	Preimage string  `json:"preimage"`
	FeesPaid *uint64 `json:"fees_paid"`
}

type MultiPayKeysendParams struct {
	Keysends []MultiPayKeysendElement `json:"keysends"`
}

type MultiPayKeysendElement struct {
	KeysendParams
	Id string `json:"id"`
}

type MultiPayInvoiceParams struct {
	Invoices []MultiPayInvoiceElement `json:"invoices"`
}

type MultiPayInvoiceElement struct {
	PayParams
	Id string `json:"id"`
}

type KeysendParams struct {
	Amount     int64                `json:"amount"`
	Pubkey     string               `json:"pubkey"`
	Preimage   string               `json:"preimage"`
	TLVRecords []lnclient.TLVRecord `json:"tlv_records"`
}

type BalanceResponse struct {
	Balance       int64  `json:"balance"`
	MaxAmount     int    `json:"max_amount"`
	BudgetRenewal string `json:"budget_renewal"`
}

type GetInfoResponse struct {
	Alias       string   `json:"alias"`
	Color       string   `json:"color"`
	Pubkey      string   `json:"pubkey"`
	Network     string   `json:"network"`
	BlockHeight uint32   `json:"block_height"`
	BlockHash   string   `json:"block_hash"`
	Methods     []string `json:"methods"`
}

type MakeInvoiceParams struct {
	Amount          int64  `json:"amount"`
	Description     string `json:"description"`
	DescriptionHash string `json:"description_hash"`
	Expiry          int64  `json:"expiry"`
}
type MakeInvoiceResponse struct {
	Transaction
}

type LookupInvoiceParams struct {
	Invoice     string `json:"invoice"`
	PaymentHash string `json:"payment_hash"`
}

type LookupInvoiceResponse struct {
	Transaction
}

type ListTransactionsParams struct {
	From   uint64 `json:"from,omitempty"`
	Until  uint64 `json:"until,omitempty"`
	Limit  uint64 `json:"limit,omitempty"`
	Offset uint64 `json:"offset,omitempty"`
	Unpaid bool   `json:"unpaid,omitempty"`
	Type   string `json:"type,omitempty"`
}

type ListTransactionsResponse struct {
	Transactions []Transaction `json:"transactions"`
}

type SignMessageParams struct {
	Message string `json:"message"`
}

type SignMessageResponse struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
}
