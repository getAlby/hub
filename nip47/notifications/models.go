package notifications

import "github.com/getAlby/nostr-wallet-connect/nip47/models"

type Notification struct {
	Notification     interface{} `json:"notification,omitempty"`
	NotificationType string      `json:"notification_type"`
}

const (
	NOTIFICATION_TYPES            = "payment_received" // e.g. "payment_received payment_sent balance_updated payment_sent channel_opened channel_closed ..."
	PAYMENT_RECEIVED_NOTIFICATION = "payment_received"
)

type PaymentReceivedNotification struct {
	models.Transaction
}
