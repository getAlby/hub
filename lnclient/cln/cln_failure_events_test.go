package cln

import (
	"context"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	clngrpc "github.com/getAlby/hub/lnclient/cln/clngrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
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

// newFailureSvc builds a CLNService whose listpays RPC answers with pays,
// so handleSendpayFailure's terminality check sees the node's authoritative
// payment state.
func newFailureSvc(publisher events.EventPublisher, pays ...*clngrpc.ListpaysPays) *CLNService {
	stub := &stubListPaysClient{listPaysFn: func(_ context.Context, _ *clngrpc.ListpaysRequest, _ ...grpc.CallOption) (*clngrpc.ListpaysResponse, error) {
		return &clngrpc.ListpaysResponse{Pays: pays}, nil
	}}
	return &CLNService{ctx: context.Background(), eventPublisher: publisher, client: stub}
}

// newFailingSvc builds a CLNService whose listpays RPC always errors, so the
// terminality check cannot resolve the node's state.
func newFailingSvc(publisher events.EventPublisher) *CLNService {
	stub := &stubListPaysClient{listPaysFn: func(_ context.Context, _ *clngrpc.ListpaysRequest, _ ...grpc.CallOption) (*clngrpc.ListpaysResponse, error) {
		return nil, errors.New("rpc error: listpays unreachable")
	}}
	return &CLNService{ctx: context.Background(), eventPublisher: publisher, client: stub}
}

// The consumer in transactions/transactions_service.go asserts
// *lnclient.PaymentFailedEventProperties. Events published by this backend
// MUST satisfy that assertion, or the failure is silently dropped
// ("Failed to cast event") and pending transactions are never marked failed.
func TestHandleSendpayFailure_PublishesWrapperProperties(t *testing.T) {
	publisher := &capturingEventPublisher{}
	svc := newFailureSvc(publisher, statusPay(clngrpc.ListpaysPays_FAILED, 100))

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
	svc := newFailureSvc(publisher)

	err := svc.handleSendpayFailure(context.Background(), &clngrpc.SendPayFailureNotification{})
	assert.Error(t, err)
	assert.Empty(t, publisher.drain(), "nothing may be published when conversion fails")
}

// CLN emits sendpay_failure for individual xpay attempts, including transient
// part failures the payment group retries immediately. While the node still
// reports the payment PENDING, publishing payment_failed would mark the live
// hub transaction FAILED and release its fee reserve mid-retry.
func TestHandleSendpayFailure_TransientPartFailureWhileGroupPendingNoPublish(t *testing.T) {
	publisher := &capturingEventPublisher{}
	svc := newFailureSvc(publisher,
		statusPay(clngrpc.ListpaysPays_FAILED, 100),
		statusPay(clngrpc.ListpaysPays_PENDING, 100),
	)

	err := svc.handleSendpayFailure(context.Background(), testSendpayFailure(t))
	require.NoError(t, err)
	assert.Empty(t, publisher.drain(), "a failed part with a pending retry must not fail the payment")
}

// Failed part followed by a successful retry: the node reports the payment
// COMPLETE. The settlement arrives through payment_sent; a payment_failed
// event here would race the settle and mis-report a succeeded payment.
func TestHandleSendpayFailure_FailedPartSettledByRetryNoPublish(t *testing.T) {
	publisher := &capturingEventPublisher{}
	preimage := make([]byte, 32)
	preimage[0] = 0xcd
	svc := newFailureSvc(publisher,
		statusPay(clngrpc.ListpaysPays_FAILED, 100),
		completePay(preimage, 100, 101),
	)

	err := svc.handleSendpayFailure(context.Background(), testSendpayFailure(t))
	require.NoError(t, err)
	assert.Empty(t, publisher.drain(), "a payment already settled on the node must not be reported failed")
}

// If the node's state cannot be resolved, the failure is still published:
// the defect-1 guarantee (async terminal failures mark pending rows FAILED)
// must not silently disappear behind a broken lookup path.
func TestHandleSendpayFailure_LookupErrorStillPublishes(t *testing.T) {
	publisher := &capturingEventPublisher{}
	svc := newFailingSvc(publisher)

	err := svc.handleSendpayFailure(context.Background(), testSendpayFailure(t))
	require.NoError(t, err)

	published := publisher.drain()
	require.Len(t, published, 1)
	assert.Equal(t, "nwc_lnclient_payment_failed", published[0].Event)
}

// A successful listpays with no record (confirmed absence) still publishes:
// there is no pending or settled attempt protecting this payment.
func TestHandleSendpayFailure_ConfirmedAbsenceStillPublishes(t *testing.T) {
	publisher := &capturingEventPublisher{}
	svc := newFailureSvc(publisher)

	err := svc.handleSendpayFailure(context.Background(), testSendpayFailure(t))
	require.NoError(t, err)

	published := publisher.drain()
	require.Len(t, published, 1)
	assert.Equal(t, "nwc_lnclient_payment_failed", published[0].Event)
}

func TestHandleSendpayFailure_ConcurrentAllWrapped(t *testing.T) {
	publisher := &capturingEventPublisher{}
	svc := newFailureSvc(publisher, statusPay(clngrpc.ListpaysPays_FAILED, 100))

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
