package events

import (
	"context"
	"slices"

	"github.com/sirupsen/logrus"
)

type EventSubscriber interface {
	ConsumeEvent(ctx context.Context, event *Event) error
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

type eventPublisher struct {
	logger    *logrus.Logger
	listeners []EventSubscriber
}

type EventPublisher interface {
	RegisterSubscriber(eventListener EventSubscriber)
	RemoveSubscriber(eventListener EventSubscriber)
	Publish(event *Event)
}

func NewEventPublisher(logger *logrus.Logger, enabled bool) *eventPublisher {
	eventPublisher := &eventPublisher{
		logger:    logger,
		listeners: []EventSubscriber{},
	}
	return eventPublisher
}

func (el *eventPublisher) RegisterSubscriber(listener EventSubscriber) {
	el.listeners = append(el.listeners, listener)
}

func (el *eventPublisher) RemoveSubscriber(listenerToRemove EventSubscriber) {
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
	el.logger.WithFields(logrus.Fields{"event": event}).Info("Logging event")
	for _, listener := range el.listeners {
		go func(listener EventSubscriber) {
			err := listener.ConsumeEvent(context.Background(), event)
			if err != nil {
				el.logger.WithError(err).Error("Failed to consume event")
			}
		}(listener)
	}
}
