package events

import (
	"context"

	"github.com/sirupsen/logrus"
)

type EventListener interface {
	Log(ctx context.Context, event *Event) error
}

type Event struct {
	Event      string      `json:"event"`
	Properties interface{} `json:"properties,omitempty"`
}

type eventLogger struct {
	logger    *logrus.Logger
	listeners []EventListener
}

type EventLogger interface {
	Subscribe(eventListener EventListener)
	Log(event *Event)
}

func NewEventLogger(logger *logrus.Logger) *eventLogger {
	eventLogger := &eventLogger{
		logger:    logger,
		listeners: []EventListener{},
	}
	return eventLogger
}

func (el *eventLogger) Subscribe(listener EventListener) {
	el.listeners = append(el.listeners, listener)
}

func (el *eventLogger) Log(event *Event) {
	el.logger.WithField("event", event).Info("Logging event")
	for _, listener := range el.listeners {
		go func(listener EventListener) {
			err := listener.Log(context.Background(), event)
			if err != nil {
				el.logger.WithError(err).Error("Failed to log event")
			}
		}(listener)
	}
}
