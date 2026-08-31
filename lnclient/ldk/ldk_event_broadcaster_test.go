package ldk

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/getAlby/ldk-node-go/ldk_node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// awaitLdkPaymentEvent mirrors the event-await loops used by SendPaymentSync,
// SendKeysend, OpenChannel and PayOfferSync: block until a payment event
// arrives, the service context is cancelled, or the broadcaster closes the
// subscription on node shutdown.
func awaitLdkPaymentEvent(ctx context.Context, subscription chan *ldk_node.Event) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-subscription:
			if !ok {
				return errors.New("LDK event subscription closed (node shutting down)")
			}
			if _, isPaymentSuccessful := (*event).(ldk_node.EventPaymentSuccessful); isPaymentSuccessful {
				return nil
			}
		}
	}
}

// Regression test: the broadcaster closes all subscription channels when its
// context is cancelled on shutdown (see the deferred close in serve()).
// Subscriber loops that dereference the received event without checking ok
// used to panic with a nil pointer dereference, killing the process during
// lock/stop/restart with in-flight payments or channel opens.
func TestLDKEventBroadcaster_SubscriptionClosedOnShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	broadcaster := NewLDKEventBroadcaster(ctx, make(chan *ldk_node.Event))

	subscription := broadcaster.Subscribe()

	// Simulate node shutdown: serve() returns and its defer closes all
	// listener channels. The receive in the await loop blocks until the
	// channel is closed, so this is deterministic without sleeps.
	cancel()

	require.NotPanics(t, func() {
		// a fresh, non-cancelled context makes the closed subscription the
		// only ready select case - the exact race the payment loops hit on
		// shutdown when the subscription case wins over the ctx case
		err := awaitLdkPaymentEvent(context.Background(), subscription)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "subscription closed")
	})
}

// Cancelling the subscriber context mid-await (no events in flight and no
// shutdown closing the subscription) must return the context error, not panic.
func TestLDKEventBroadcaster_ContextCancelledWhileAwaiting(t *testing.T) {
	broadcasterCtx, cancelBroadcaster := context.WithCancel(context.Background())
	t.Cleanup(cancelBroadcaster)
	broadcaster := NewLDKEventBroadcaster(broadcasterCtx, make(chan *ldk_node.Event))

	subscription := broadcaster.Subscribe()

	awaitCtx, cancelAwait := context.WithCancel(context.Background())
	errChannel := make(chan error, 1)
	go func() {
		errChannel <- awaitLdkPaymentEvent(awaitCtx, subscription)
	}()

	cancelAwait()

	select {
	case err := <-errChannel:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for await loop to return after context cancellation")
	}
}
