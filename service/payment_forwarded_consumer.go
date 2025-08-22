package service

import (
	"context"

	"gorm.io/gorm"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
)

type paymentForwardedConsumer struct {
	events.EventSubscriber
	db *gorm.DB
}

// When a new app is created, subscribe to it on the relay
func (c *paymentForwardedConsumer) ConsumeEvent(ctx context.Context, event *events.Event, globalProperties map[string]interface{}) {
	if event.Event != "nwc_payment_forwarded" {
		return
	}

	properties, ok := event.Properties.(*lnclient.PaymentForwardedEventProperties)
	if !ok {
		logger.Logger.WithField("event", event).Error("Failed to cast event.Properties to payment forwarded event properties")
		return
	}
	forward := &db.Forward{
		OutboundAmountForwardedMsat: properties.OutboundAmountForwardedMsat,
		TotalFeeEarnedMsat:          properties.TotalFeeEarnedMsat,
	}
	err := c.db.Create(forward).Error
	if err != nil {
		logger.Logger.WithError(err).Error("failed to save forward to db")
	}
}
