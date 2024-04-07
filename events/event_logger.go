package events

import (
	"context"
	"slices"

	"github.com/sirupsen/logrus"
)

type EventListener interface {
	Log(ctx context.Context, event *Event) error
}

type Event struct {
	Event      string      `json:"event"`
	Properties interface{} `json:"properties,omitempty"`
}

type PaymentReceivedEventProperties struct {
	PaymentHash string `json:"payment_hash"`
	Amount      uint64 `json:"amount"`
	NodeType    string `json:"node_type"`
}

type eventLogger struct {
	logger    *logrus.Logger
	listeners []EventListener
}

// TODO: rename this or use an existing pubsub/eventbus implementation
type EventLogger interface {
	Subscribe(eventListener EventListener)
	Unsubscribe(eventListener EventListener)
	Log(event *Event)
}

func NewEventLogger(logger *logrus.Logger, enabled bool) *eventLogger {
	eventLogger := &eventLogger{
		logger:    logger,
		listeners: []EventListener{},
	}
	return eventLogger
}

func (el *eventLogger) Subscribe(listener EventListener) {
	el.listeners = append(el.listeners, listener)
}

func (el *eventLogger) Unsubscribe(listenerToRemove EventListener) {
	for i, listener := range el.listeners {
		// delete the listener from the listeners array
		if listener == listenerToRemove {
			el.listeners[i] = el.listeners[len(el.listeners)-1]
			el.listeners = slices.Delete(el.listeners, len(el.listeners)-1, len(el.listeners))
			break
		}
	}
}

func (el *eventLogger) Log(event *Event) {
	el.logger.WithFields(logrus.Fields{"event": event}).Info("Logging event")
	for _, listener := range el.listeners {
		go func(listener EventListener) {
			err := listener.Log(context.Background(), event)
			if err != nil {
				el.logger.WithError(err).Error("Failed to log event")
			}
		}(listener)
	}
}
