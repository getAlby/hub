package notifications

import (
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/sirupsen/logrus"
)

// EventLogFields identifies the event's payment without logging the
// preimage, metadata or other private transaction data
func EventLogFields(event *events.Event) logrus.Fields {
	fields := logrus.Fields{"event": event.Event}
	if transaction, ok := event.Properties.(*db.Transaction); ok {
		fields["transaction_id"] = transaction.ID
		fields["payment_hash"] = transaction.PaymentHash
		fields["app_id"] = transaction.AppId
		fields["amount_msat"] = transaction.AmountMsat
	}
	return fields
}

// identifies the notification's payment without logging the preimage,
// metadata or other private transaction data
func (n *Notification) logFields() logrus.Fields {
	fields := logrus.Fields{"notification_type": n.NotificationType}
	switch notification := n.Notification.(type) {
	case PaymentReceivedNotification:
		fields["payment_hash"] = notification.PaymentHash
	case PaymentSentNotification:
		fields["payment_hash"] = notification.PaymentHash
	case HoldInvoiceAcceptedNotification:
		fields["payment_hash"] = notification.PaymentHash
	}
	return fields
}
