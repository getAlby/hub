package main

import (
	"context"
	"slices"

	"github.com/getAlby/ldk-node-go/ldk_node"
	"github.com/sirupsen/logrus"
)

// based on https://betterprogramming.pub/how-to-broadcast-messages-in-go-using-channels-b68f42bdf32e
type ldkEventBroadcastServer struct {
	logger         *logrus.Logger
	source         <-chan *ldk_node.Event
	listeners      []chan *ldk_node.Event
	addListener    chan chan *ldk_node.Event
	removeListener chan (<-chan *ldk_node.Event)
}

type LDKEventBroadcaster interface {
	Subscribe() <-chan *ldk_node.Event
	CancelSubscription(<-chan *ldk_node.Event)
}

func NewLDKEventBroadcaster(logger *logrus.Logger, ctx context.Context, source <-chan *ldk_node.Event) LDKEventBroadcaster {
	service := &ldkEventBroadcastServer{
		logger:         logger,
		source:         source,
		listeners:      make([]chan *ldk_node.Event, 0),
		addListener:    make(chan chan *ldk_node.Event),
		removeListener: make(chan (<-chan *ldk_node.Event)),
	}
	go service.serve(ctx)
	return service
}

func (s *ldkEventBroadcastServer) Subscribe() <-chan *ldk_node.Event {
	newListener := make(chan *ldk_node.Event)
	s.addListener <- newListener
	return newListener
}

func (s *ldkEventBroadcastServer) CancelSubscription(channel <-chan *ldk_node.Event) {
	s.removeListener <- channel
}

func (s *ldkEventBroadcastServer) serve(ctx context.Context) {
	defer func() {
		for _, listener := range s.listeners {
			if listener != nil {
				close(listener)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case newListener := <-s.addListener:
			s.listeners = append(s.listeners, newListener)
		case listenerToRemove := <-s.removeListener:
			for i, listener := range s.listeners {
				if listener == listenerToRemove {
					s.listeners[i] = s.listeners[len(s.listeners)-1]
					s.listeners = slices.Delete(s.listeners, len(s.listeners)-1, len(s.listeners))
					close(listener)
					break
				}
			}
		case event := <-s.source:
			for _, listener := range s.listeners {
				listener <- event
			}
		}
	}
}
