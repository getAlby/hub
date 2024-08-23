package models

import (
	"encoding/json"
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
)

type Transaction struct {
	Type            string      `json:"type"`
	Invoice         string      `json:"invoice"`
	Description     string      `json:"description"`
	DescriptionHash string      `json:"description_hash"`
	Preimage        string      `json:"preimage"`
	PaymentHash     string      `json:"payment_hash"`
	Amount          int64       `json:"amount"`
	FeesPaid        int64       `json:"fees_paid"`
	CreatedAt       int64       `json:"created_at"`
	ExpiresAt       *int64      `json:"expires_at"`
	SettledAt       *int64      `json:"settled_at"`
	Metadata        interface{} `json:"metadata,omitempty"`
}

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
