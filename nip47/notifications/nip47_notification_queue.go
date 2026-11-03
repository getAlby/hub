package notifications

import (
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/logger"
)

type Nip47NotificationQueue interface {
	Channel() <-chan *events.Event
	AddToQueue(event *events.Event) bool
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

func (q *nip47NotificationQueue) AddToQueue(event *events.Event) bool {
	select {
	case q.channel <- event: // Put in the channel unless it is full
		// successfully sent to channel
		return true
	default:
		// channel full
		logger.Logger.WithField("event", event).Error("NIP47NotificationQueue channel full. Discarding value")
		return false
	}
}

func (q *nip47NotificationQueue) Channel() <-chan *events.Event {
	return q.channel
}
