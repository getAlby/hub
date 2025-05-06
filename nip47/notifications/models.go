package notifications

import "github.com/getAlby/hub/nip47/models"

type Notification struct {
	Notification     interface{} `json:"notification,omitempty"`
	NotificationType string      `json:"notification_type"`
}

const (
	PAYMENT_RECEIVED_NOTIFICATION      = "payment_received"
	PAYMENT_SENT_NOTIFICATION          = "payment_sent"
	HOLD_INVOICE_ACCEPTED_NOTIFICATION = "hold_invoice_accepted"
)

type PaymentSentNotification struct {
	models.Transaction
}

type PaymentReceivedNotification struct {
	models.Transaction
}

type HoldInvoiceAcceptedNotification struct {
	models.Transaction
}
