package nip47

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
