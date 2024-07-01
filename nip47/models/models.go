package models

import (
	"encoding/json"

	"github.com/getAlby/nostr-wallet-connect/lnclient"
)

const (
	INFO_EVENT_KIND   = 13194
	REQUEST_KIND      = 23194
	RESPONSE_KIND     = 23195
	NOTIFICATION_KIND = 23196

	// request methods
	PAY_INVOICE_METHOD       = "pay_invoice"
	GET_BALANCE_METHOD       = "get_balance"
	GET_INFO_METHOD          = "get_info"
	MAKE_INVOICE_METHOD      = "make_invoice"
	LOOKUP_INVOICE_METHOD    = "lookup_invoice"
	LIST_TRANSACTIONS_METHOD = "list_transactions"
	PAY_KEYSEND_METHOD       = "pay_keysend"
	MULTI_PAY_INVOICE_METHOD = "multi_pay_invoice"
	MULTI_PAY_KEYSEND_METHOD = "multi_pay_keysend"
	SIGN_MESSAGE_METHOD      = "sign_message"

	ERROR_INTERNAL             = "INTERNAL"
	ERROR_NOT_IMPLEMENTED      = "NOT_IMPLEMENTED"
	ERROR_QUOTA_EXCEEDED       = "QUOTA_EXCEEDED"
	ERROR_INSUFFICIENT_BALANCE = "INSUFFICIENT_BALANCE"
	ERROR_UNAUTHORIZED         = "UNAUTHORIZED"
	ERROR_EXPIRED              = "EXPIRED"
	ERROR_RESTRICTED           = "RESTRICTED"
	ERROR_BAD_REQUEST          = "BAD_REQUEST"
	OTHER                      = "OTHER"
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

type Error struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}
