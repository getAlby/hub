package transactions

import (
	"context"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
)

type holdInvoiceUpdatedConsumer struct {
	paymentRequest  string
	settledChannel  chan<- *db.Transaction
	canceledChannel chan<- *db.Transaction
}

func newHoldInvoiceUpdatedConsumer(paymentRequest string, settledChannel chan<- *db.Transaction, canceledChannel chan<- *db.Transaction) *holdInvoiceUpdatedConsumer {
	return &holdInvoiceUpdatedConsumer{
		paymentRequest:  paymentRequest,
		settledChannel:  settledChannel,
		canceledChannel: canceledChannel,
	}
}

func (consumer *holdInvoiceUpdatedConsumer) ConsumeEvent(ctx context.Context, event *events.Event, globalProperties map[string]interface{}) {
	if event.Event == "nwc_payment_received" && event.Properties.(*db.Transaction).PaymentRequest == consumer.paymentRequest {
		consumer.settledChannel <- event.Properties.(*db.Transaction)
	}
	if event.Event == "nwc_hold_invoice_canceled" && event.Properties.(*db.Transaction).PaymentRequest == consumer.paymentRequest {
		consumer.canceledChannel <- event.Properties.(*db.Transaction)
	}
}
