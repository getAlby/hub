package constants

// shared constants used by multiple packages

const (
	TRANSACTION_TYPE_INCOMING = "incoming"
	TRANSACTION_TYPE_OUTGOING = "outgoing"

	TRANSACTION_STATE_PENDING = "PENDING"
	TRANSACTION_STATE_SETTLED = "SETTLED"
	TRANSACTION_STATE_FAILED  = "FAILED"
)

const (
	BUDGET_RENEWAL_DAILY   = "daily"
	BUDGET_RENEWAL_WEEKLY  = "weekly"
	BUDGET_RENEWAL_MONTHLY = "monthly"
	BUDGET_RENEWAL_YEARLY  = "yearly"
	BUDGET_RENEWAL_NEVER   = "never"
)

const (
	PAY_INVOICE_SCOPE       = "pay_invoice" // also covers pay_keysend and multi_* payment methods
	GET_BALANCE_SCOPE       = "get_balance"
	GET_INFO_SCOPE          = "get_info"
	MAKE_INVOICE_SCOPE      = "make_invoice"
	LOOKUP_INVOICE_SCOPE    = "lookup_invoice"
	LIST_TRANSACTIONS_SCOPE = "list_transactions"
	SIGN_MESSAGE_SCOPE      = "sign_message"
	NOTIFICATIONS_SCOPE     = "notifications" // covers all notification types
)

// limit encoded metadata length, otherwise relays may have trouble listing multiple transactions
// given a relay limit of 512000 bytes and ideally being able to list 25 transactions,
// each transaction would have to have a maximum size of 20480
// accounting for encryption and other metadata in the response, this is set to 4096 characters
const INVOICE_METADATA_MAX_LENGTH = 4096

// errors used by NIP-47 and the transaction service
const (
	ERROR_INTERNAL             = "INTERNAL"
	ERROR_NOT_IMPLEMENTED      = "NOT_IMPLEMENTED"
	ERROR_QUOTA_EXCEEDED       = "QUOTA_EXCEEDED"
	ERROR_INSUFFICIENT_BALANCE = "INSUFFICIENT_BALANCE"
	ERROR_UNAUTHORIZED         = "UNAUTHORIZED"
	ERROR_EXPIRED              = "EXPIRED"
	ERROR_RESTRICTED           = "RESTRICTED"
	ERROR_BAD_REQUEST          = "BAD_REQUEST"
	ERROR_NOT_FOUND            = "NOT_FOUND"
	ERROR_UNSUPPORTED_VERSION  = "UNSUPPORTED_VERSION"
	ERROR_OTHER                = "OTHER"
)
