package notifications

import (
	"context"
	"errors"

	"github.com/getAlby/nostr-wallet-connect/events"
	"github.com/getAlby/nostr-wallet-connect/logger"
)

type Nip47NotificationQueue interface {
	events.EventSubscriber
	Channel() <-chan *events.Event
}

type nip47NotificationQueue struct {
	channel chan *events.Event
}

/*
Queue events that will be consumed when the relay connection is online
*/
func NewNip47NotificationQueue() *nip47NotificationQueue {
	return &nip47NotificationQueue{
		channel: make(chan *events.Event, 1000),
	}
}

func (q *nip47NotificationQueue) ConsumeEvent(ctx context.Context, event *events.Event, globalProperties map[string]interface{}) error {
	select {
	case q.channel <- event: // Put in the channel unless it is full
		return nil
	default:
		logger.Logger.WithField("event", event).Error("NIP47NotificationQueue channel full. Discarding value")
		return errors.New("nip-47 notification queue full")
	}
}

func (q *nip47NotificationQueue) Channel() <-chan *events.Event {
	return q.channel
}
