package tests

import (
	"context"
	"time"

	"github.com/getAlby/hub/events"
)

type mockEventConsumer struct {
	consumedEvents []*events.Event
}

func NewMockEventConsumer() *mockEventConsumer {
	return &mockEventConsumer{
		consumedEvents: []*events.Event{},
	}
}

func (e *mockEventConsumer) ConsumeEvent(ctx context.Context, event *events.Event, globalProperties map[string]interface{}) {
	e.consumedEvents = append(e.consumedEvents, event)
}

func (e *mockEventConsumer) GetConsumeEvents() []*events.Event {
	// events are consumed async - give it a bit of time for tests
	time.Sleep(1 * time.Millisecond)
	return e.consumedEvents
}
