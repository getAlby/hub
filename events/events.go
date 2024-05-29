package events

import (
	"context"
	"slices"
	"sync"

	"github.com/sirupsen/logrus"
)

type eventPublisher struct {
	logger           *logrus.Logger
	listeners        []EventSubscriber
	subscriberMtx    sync.Mutex
	globalProperties map[string]interface{}
}

func NewEventPublisher(logger *logrus.Logger) *eventPublisher {
	eventPublisher := &eventPublisher{
		logger:           logger,
		listeners:        []EventSubscriber{},
		globalProperties: map[string]interface{}{},
	}
	return eventPublisher
}

func (el *eventPublisher) RegisterSubscriber(listener EventSubscriber) {
	el.subscriberMtx.Lock()
	defer el.subscriberMtx.Unlock()
	el.listeners = append(el.listeners, listener)
}

func (el *eventPublisher) RemoveSubscriber(listenerToRemove EventSubscriber) {
	el.subscriberMtx.Lock()
	defer el.subscriberMtx.Unlock()

	for i, listener := range el.listeners {
		// delete the listener from the listeners array
		if listener == listenerToRemove {
			el.listeners[i] = el.listeners[len(el.listeners)-1]
			el.listeners = slices.Delete(el.listeners, len(el.listeners)-1, len(el.listeners))
			break
		}
	}
}

func (el *eventPublisher) Publish(event *Event) {
	el.subscriberMtx.Lock()
	defer el.subscriberMtx.Unlock()
	el.logger.WithFields(logrus.Fields{"event": event}).Info("Logging event")
	for _, listener := range el.listeners {
		go func(listener EventSubscriber) {
			err := listener.ConsumeEvent(context.Background(), event, el.globalProperties)
			if err != nil {
				el.logger.WithError(err).Error("Failed to consume event")
			}
		}(listener)
	}
}

func (el *eventPublisher) SetGlobalProperty(key string, value interface{}) {
	el.globalProperties[key] = value
}
