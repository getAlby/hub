package notifications

import (
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/logger"
)

type Nip47NotificationQueue interface {
	Channel() <-chan *events.Event
	AddToQueue(event *events.Event)
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

func (q *nip47NotificationQueue) AddToQueue(event *events.Event) {
	select {
	case q.channel <- event: // Put in the channel unless it is full
		// successfully sent to channel
	default:
		// channel full
		logger.Logger.WithField("event", event).Error("NIP47NotificationQueue channel full. Discarding value")
	}
}

func (q *nip47NotificationQueue) Channel() <-chan *events.Event {
	return q.channel
}
