package events

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/getAlby/hub/logger"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	logger.Init(strconv.Itoa(int(logrus.PanicLevel)))
	os.Exit(m.Run())
}

type eventSubscriberFunc func(context.Context, *Event, map[string]interface{})

func (fn eventSubscriberFunc) ConsumeEvent(ctx context.Context, event *Event, globalProperties map[string]interface{}) {
	fn(ctx, event, globalProperties)
}

func TestPublishSyncAllowsNestedPublish(t *testing.T) {
	publisher := NewEventPublisher()
	nestedReceived := make(chan struct{}, 1)
	publisher.RegisterSubscriber(eventSubscriberFunc(func(_ context.Context, event *Event, _ map[string]interface{}) {
		if event.Event == "outer" {
			publisher.PublishSync(&Event{Event: "nested"})
		}
		if event.Event == "nested" {
			nestedReceived <- struct{}{}
		}
	}))

	published := make(chan struct{})
	go func() {
		publisher.PublishSync(&Event{Event: "outer"})
		close(published)
	}()

	select {
	case <-published:
	case <-time.After(time.Second):
		t.Fatal("nested synchronous publish deadlocked")
	}

	select {
	case <-nestedReceived:
	case <-time.After(time.Second):
		t.Fatal("nested event was not delivered")
	}
}

func TestPublishUsesGlobalPropertySnapshot(t *testing.T) {
	publisher := NewEventPublisher()
	publisher.SetGlobalProperty("key", "before")
	propertiesReceived := make(chan map[string]interface{}, 1)
	publisher.RegisterSubscriber(eventSubscriberFunc(func(_ context.Context, _ *Event, properties map[string]interface{}) {
		propertiesReceived <- properties
	}))

	publisher.PublishSync(&Event{Event: "event"})
	publisher.SetGlobalProperty("key", "after")

	properties := <-propertiesReceived
	require.Equal(t, "before", properties["key"])
}
