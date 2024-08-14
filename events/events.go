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

func (ep *eventPublisher) RegisterSubscriber(listener EventSubscriber) {
	ep.subscriberMtx.Lock()
	defer ep.subscriberMtx.Unlock()
	ep.listeners = append(ep.listeners, listener)
}

func (ep *eventPublisher) RemoveSubscriber(listenerToRemove EventSubscriber) {
	ep.subscriberMtx.Lock()
	defer ep.subscriberMtx.Unlock()

	for i, listener := range ep.listeners {
		// delete the listener from the listeners array
		if listener == listenerToRemove {
			ep.listeners[i] = ep.listeners[len(ep.listeners)-1]
			ep.listeners = slices.Delete(ep.listeners, len(ep.listeners)-1, len(ep.listeners))
			break
		}
	}
}

func (ep *eventPublisher) Publish(event *Event) {
	ep.subscriberMtx.Lock()
	defer ep.subscriberMtx.Unlock()
	logger.Logger.WithFields(logrus.Fields{"event": event, "global": ep.globalProperties}).Debug("Publishing event")
	for _, listener := range ep.listeners {
		// events are consumed in sequence as some listeners depend on earlier consumers
		// (e.g. NIP-47 notifier depends on transactions service updating transactions)
		listener.ConsumeEvent(context.Background(), event, ep.globalProperties)
	}
}

func (ep *eventPublisher) SetGlobalProperty(key string, value interface{}) {
	ep.globalProperties[key] = value
}
