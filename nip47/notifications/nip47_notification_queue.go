package notifications

import (
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/logger"
)

// A notification waiting to be sent. New notifications go to all subscribed apps;
// retries are scoped to the single app (and encryption) whose delivery failed.
type QueuedNotification struct {
	Event      *events.Event
	AppId      *uint  // nil = all subscribed apps
	Encryption string // empty = all encryption types
	Attempt    uint32
}

type Nip47NotificationQueue interface {
	Channel() <-chan *QueuedNotification
	AddToQueue(notification *QueuedNotification)
}

type nip47NotificationQueue struct {
	channel chan *QueuedNotification
}

/*
Queue events that will be consumed when the relay connection is online
*/
func NewNip47NotificationQueue() *nip47NotificationQueue {
	return &nip47NotificationQueue{
		channel: make(chan *QueuedNotification, 1000),
	}
}

func (q *nip47NotificationQueue) AddToQueue(notification *QueuedNotification) {
	select {
	case q.channel <- notification: // Put in the channel unless it is full
		// successfully sent to channel
	default:
		// channel full
		logger.Logger.WithField("event", notification.Event).Error("NIP47NotificationQueue channel full. Discarding value")
	}
}

func (q *nip47NotificationQueue) Channel() <-chan *QueuedNotification {
	return q.channel
}
