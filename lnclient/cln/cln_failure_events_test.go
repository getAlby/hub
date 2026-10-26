package cln

import (
	"context"
	"encoding/hex"
	"strings"
	"sync"
	"testing"

	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	clngrpc "github.com/getAlby/hub/lnclient/cln/clngrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type capturingEventPublisher struct {
	mu        sync.Mutex
	published []*events.Event
}

func (p *capturingEventPublisher) Publish(event *events.Event) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.published = append(p.published, event)
}

func (p *capturingEventPublisher) PublishSync(event *events.Event)                 { p.Publish(event) }
func (p *capturingEventPublisher) RegisterSubscriber(events.EventSubscriber)       {}
func (p *capturingEventPublisher) RemoveSubscriber(events.EventSubscriber)         {}
func (p *capturingEventPublisher) SetGlobalProperty(key string, value interface{}) {}

func (p *capturingEventPublisher) drain() []*events.Event {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := p.published
	p.published = nil
	return out
}

func u64(v uint64) *uint64 { return &v }
func str(v string) *string { return &v }
func u32(v uint32) *uint32 { return &v }
func hash32(t *testing.T, h string) []byte {
	t.Helper()
	b, err := hex.DecodeString(h)
	require.NoError(t, err)
	return b
}

func testSendpayFailure(t *testing.T) *clngrpc.SendPayFailureNotification {
	status := clngrpc.SendpayFailureData_FAILED
	failcodename := "WIRE_TEMPORARY_CHANNEL_FAILURE"
	return &clngrpc.SendPayFailureNotification{
		Code:    209,
		Message: "xpay: Failed after 23 attempts",
		Data: &clngrpc.SendpayFailureData{
			Id:             u64(421),
			PaymentHash:    hash32(t, "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"),
			Groupid:        u64(8675309876543),
			Partid:         u64(7),
			Status:         &status,
			Failcode:       u32(209),
			Failcodename:   &failcodename,
			AmountMsat:     &clngrpc.Amount{Msat: 1001000},
			AmountSentMsat: &clngrpc.Amount{Msat: 1002500},
			CreatedAt:      u64(1700000000),
			CompletedAt:    u64(1700000100),
			Description:    str("mock description"),
		},
	}
}

// The consumer in transactions/transactions_service.go asserts
// *lnclient.PaymentFailedEventProperties. Events published by this backend
// MUST satisfy that assertion, or the failure is silently dropped
// ("Failed to cast event") and pending transactions are never marked failed.
func TestHandleSendpayFailure_PublishesWrapperProperties(t *testing.T) {
	publisher := &capturingEventPublisher{}
	svc := &CLNService{eventPublisher: publisher}

	err := svc.handleSendpayFailure(context.Background(), testSendpayFailure(t))
	require.NoError(t, err)

	published := publisher.drain()
	require.Len(t, published, 1)
	assert.Equal(t, "nwc_lnclient_payment_failed", published[0].Event)

	props, ok := published[0].Properties.(*lnclient.PaymentFailedEventProperties)
	require.True(t, ok, "consumer asserts *lnclient.PaymentFailedEventProperties; publisher must wrap")

	require.NotNil(t, props.Transaction)
	assert.Equal(t,
		"0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20",
		props.Transaction.PaymentHash,
	)
	assert.Equal(t, "outgoing", props.Transaction.Type)
	assert.NotEmpty(t, props.Reason, "failure reason must be forwarded to markPaymentFailed")
	assert.True(t,
		strings.Contains(props.Reason, "WIRE_TEMPORARY_CHANNEL_FAILURE") ||
			strings.Contains(props.Reason, "FAILED"),
		"reason should carry CLN failure semantics, got %q", props.Reason)
}

func TestHandleSendpayFailure_NilDataIsErrorNoPublish(t *testing.T) {
	publisher := &capturingEventPublisher{}
	svc := &CLNService{eventPublisher: publisher}

	err := svc.handleSendpayFailure(context.Background(), &clngrpc.SendPayFailureNotification{})
	assert.Error(t, err)
	assert.Empty(t, publisher.drain(), "nothing may be published when conversion fails")
}

func TestHandleSendpayFailure_ConcurrentAllWrapped(t *testing.T) {
	publisher := &capturingEventPublisher{}
	svc := &CLNService{eventPublisher: publisher}

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			require.NoError(t, svc.handleSendpayFailure(context.Background(), testSendpayFailure(t)))
		}()
	}
	wg.Wait()

	published := publisher.drain()
	require.Len(t, published, 50)
	for _, e := range published {
		props, ok := e.Properties.(*lnclient.PaymentFailedEventProperties)
		require.True(t, ok)
		require.NotNil(t, props.Transaction)
	}
}
