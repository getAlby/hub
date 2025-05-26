package transactions

import (
	"context"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
)

type holdInvoiceUpdatedConsumer struct {
	paymentHash     string
	settledChannel  chan<- *db.Transaction
	canceledChannel chan<- *db.Transaction
}

func newHoldInvoiceUpdatedConsumer(paymentHash string, settledChannel chan<- *db.Transaction, canceledChannel chan<- *db.Transaction) *holdInvoiceUpdatedConsumer {
	return &holdInvoiceUpdatedConsumer{
		paymentHash:     paymentHash,
		settledChannel:  settledChannel,
		canceledChannel: canceledChannel,
	}
}

func (consumer *holdInvoiceUpdatedConsumer) ConsumeEvent(ctx context.Context, event *events.Event, globalProperties map[string]interface{}) {
	if event.Event == "nwc_payment_received" && event.Properties.(*db.Transaction).PaymentHash == consumer.paymentHash {
		consumer.settledChannel <- event.Properties.(*db.Transaction)
	}
	if event.Event == "nwc_hold_invoice_canceled" && event.Properties.(*db.Transaction).PaymentHash == consumer.paymentHash {
		consumer.canceledChannel <- event.Properties.(*db.Transaction)
	}
}
