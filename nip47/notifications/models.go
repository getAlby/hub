package notifications

import "github.com/getAlby/nostr-wallet-connect/nip47/models"

type Notification struct {
	Notification     interface{} `json:"notification,omitempty"`
	NotificationType string      `json:"notification_type"`
}

const (
	NOTIFICATION_TYPES            = "payment_received payment_sent"
	PAYMENT_RECEIVED_NOTIFICATION = "payment_received"
	PAYMENT_SENT_NOTIFICATION     = "payment_sent"
)

type PaymentSentNotification struct {
	models.Transaction
}

type PaymentReceivedNotification struct {
	models.Transaction
}
