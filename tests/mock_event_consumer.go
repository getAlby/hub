package tests

import (
	"context"
	"sync"
	"time"

	"github.com/getAlby/hub/events"
)

type mockEventConsumer struct {
	mtx            sync.Mutex
	consumedEvents []*events.Event
}

func NewMockEventConsumer() *mockEventConsumer {
	return &mockEventConsumer{
		consumedEvents: []*events.Event{},
	}
}

func (e *mockEventConsumer) ConsumeEvent(ctx context.Context, event *events.Event, globalProperties map[string]interface{}) {
	e.mtx.Lock()
	defer e.mtx.Unlock()
	e.consumedEvents = append(e.consumedEvents, event)
}

func (e *mockEventConsumer) GetConsumedEvents() []*events.Event {
	// events are consumed async - give it a bit of time for tests
	time.Sleep(10 * time.Millisecond)
	return e.snapshotConsumedEvents()
}

// WaitForConsumedEvents waits until at least count events have been consumed
// (events are consumed async) and returns them. On timeout it returns the
// events consumed so far, so the caller's assertions fail with a useful message.
func (e *mockEventConsumer) WaitForConsumedEvents(count int) []*events.Event {
	deadline := time.Now().Add(5 * time.Second)
	for {
		consumedEvents := e.snapshotConsumedEvents()
		if len(consumedEvents) >= count || time.Now().After(deadline) {
			return consumedEvents
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (e *mockEventConsumer) snapshotConsumedEvents() []*events.Event {
	e.mtx.Lock()
	defer e.mtx.Unlock()
	return append([]*events.Event{}, e.consumedEvents...)
}
