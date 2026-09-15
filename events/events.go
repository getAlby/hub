package events

import (
	"context"
	"slices"
	"sync"

	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/version"
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
	eventPublisher.SetGlobalProperty("version", version.Tag)
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
	ep.publish(event, false)
}
func (ep *eventPublisher) PublishSync(event *Event) {
	ep.publish(event, true)
}

func (ep *eventPublisher) publish(event *Event, sync bool) {
	ep.subscriberMtx.Lock()
	listeners := append([]EventSubscriber(nil), ep.listeners...)
	globalProperties := make(map[string]interface{}, len(ep.globalProperties))
	for key, value := range ep.globalProperties {
		globalProperties[key] = value
	}
	ep.subscriberMtx.Unlock()

	logger.Logger.WithFields(logrus.Fields{"event": event, "global": globalProperties}).Debug("Publishing event")
	for _, listener := range listeners {
		if sync {
			listener.ConsumeEvent(context.Background(), event, globalProperties)
		} else {
			// consume event without blocking thread
			go listener.ConsumeEvent(context.Background(), event, globalProperties)
		}
	}
}

func (ep *eventPublisher) SetGlobalProperty(key string, value interface{}) {
	ep.subscriberMtx.Lock()
	defer ep.subscriberMtx.Unlock()
	ep.globalProperties[key] = value
}
