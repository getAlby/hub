package tests

import (
	"context"

	"github.com/getAlby/hub/events"
)

type mockEventConsumer struct {
	ConsumedEvents []*events.Event
}

func NewMockEventConsumer() *mockEventConsumer {
	return &mockEventConsumer{
		ConsumedEvents: []*events.Event{},
	}
}

func (e *mockEventConsumer) ConsumeEvent(ctx context.Context, event *events.Event, globalProperties map[string]interface{}) {
	e.ConsumedEvents = append(e.ConsumedEvents, event)
}
