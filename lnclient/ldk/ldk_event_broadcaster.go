package ldk

import (
	"context"
	"slices"
	"time"

	"github.com/getAlby/ldk-node-go/ldk_node"
	// "github.com/getAlby/nostr-wallet-connect/ldk_node"
	"github.com/sirupsen/logrus"
)

/*
*
a LDK event broadcaster powered by channels.
There are 3 main channels:
- 1. receives the next LDK event
- 2. adding a subscriber
- 3. removing a subscriber
Based on https://betterprogramming.pub/how-to-broadcast-messages-in-go-using-channels-b68f42bdf32e
*/
type ldkEventBroadcastServer struct {
	logger         *logrus.Logger
	source         <-chan *ldk_node.Event
	listeners      []chan *ldk_node.Event
	addListener    chan chan *ldk_node.Event
	removeListener chan (<-chan *ldk_node.Event)
}

type LDKEventBroadcaster interface {
	Subscribe() chan *ldk_node.Event
	CancelSubscription(chan *ldk_node.Event)
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

func (s *ldkEventBroadcastServer) Subscribe() chan *ldk_node.Event {
	// create a new listener channel and mark it to be added to the listeners array
	newListener := make(chan *ldk_node.Event)
	s.addListener <- newListener
	return newListener
}

func (s *ldkEventBroadcastServer) CancelSubscription(channel chan *ldk_node.Event) {
	// close the channel - this could fail if the channel was already closed (just ignore it)
	func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.WithField("r", r).Error("Failed to close subscription channel")
			}
		}()
		close(channel)
	}()
	// mark the channel to be removed from the listeners array
	s.removeListener <- channel
}

func (s *ldkEventBroadcastServer) serve(ctx context.Context) {
	// close down all listeners when the LDK lnclient is shut down
	defer func() {
		for _, listener := range s.listeners {
			func() {
				defer func() {
					if r := recover(); r != nil {
						s.logger.WithField("r", r).Error("Failed to close subscription channel")
					}
				}()
				close(listener)
			}()
		}
	}()

	/**
	Process events from channels in an infinite (blocking) loop.
	*/
	for {
		select {
		case <-ctx.Done():
			// LDK lnclient was shut down
			return
		case newListener := <-s.addListener:
			s.listeners = append(s.listeners, newListener)
		case listenerToRemove := <-s.removeListener:
			for i, listener := range s.listeners {
				// delete the listener from the listeners array
				if listener == listenerToRemove {
					s.listeners[i] = s.listeners[len(s.listeners)-1]
					s.listeners = slices.Delete(s.listeners, len(s.listeners)-1, len(s.listeners))
					break
				}
			}
		case event := <-s.source:
			// got a new LDK event - send it to all listeners
			s.logger.WithFields(logrus.Fields{
				"event":         event,
				"listenerCount": len(s.listeners),
			}).Debug("Sending LDK event to listeners")
			for _, listener := range s.listeners {
				func() {
					// if we fail to send the event to the listener it was probably closed
					defer func() {
						if r := recover(); r != nil {
							s.logger.WithField("r", r).Error("Failed to send event to listener")
						}
					}()

					// try to send the event to the listener
					// this can fail if the listener is closed (due to unsubscribing)
					// worst case scenario: it times out because the listener is stuck processing an event
					select {
					case listener <- event:
						s.logger.WithFields(logrus.Fields{
							"event": event,
						}).Debug("Sent LDK event to listener")
					case <-time.After(5 * time.Second):
						s.logger.WithFields(logrus.Fields{
							"event": event,
						}).Error("Timeout sending LDK event to listener")
					}
				}()
			}
		}
	}
}
