package events

import (
	"context"
	"slices"
	"sync"

	"github.com/getAlby/hub/logger"
	"github.com/sirupsen/logrus"
)

type eventPublisher struct {
	listeners        []EventSubscriber
	subscriberMtx    sync.Mutex
	globalProperties map[string]interface{}
}

func NewEventPublisher() *eventPublisher {
	eventPublisher := &eventPublisher{
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

func (ep *eventPublisher) Publish(event *Event) {
	ep.subscriberMtx.Lock()
	defer ep.subscriberMtx.Unlock()
	logger.Logger.WithFields(logrus.Fields{"event": event}).Info("Logging event")
	for _, listener := range ep.listeners {
		go func(listener EventSubscriber) {
			err := listener.ConsumeEvent(context.Background(), event, ep.globalProperties)
			if err != nil {
				logger.Logger.WithError(err).Error("Failed to consume event")
			}
		}(listener)
	}
}

func (el *eventPublisher) SetGlobalProperty(key string, value interface{}) {
	el.globalProperties[key] = value
}
