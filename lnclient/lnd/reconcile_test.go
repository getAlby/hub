package lnd

import (
	"context"
	"encoding/hex"
	"errors"
	"sync"
	"testing"
	"time"

	"os"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/lightningnetwork/lnd/lnrpc"
)

type stubEventPublisher struct {
	mu     sync.Mutex
	events []*events.Event
}

func TestMain(m *testing.M) {
	logger.Init("error")
	os.Exit(m.Run())
}

func (s *stubEventPublisher) RegisterSubscriber(events.EventSubscriber) {}

func (s *stubEventPublisher) Publish(event *events.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, event)
}

func (s *stubEventPublisher) published() []*events.Event {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]*events.Event{}, s.events...)
}

type stubPaymentsAPI struct {
	pages []*lnrpc.ListPaymentsResponse
	err   error
	calls int
}

func (s *stubPaymentsAPI) ListPayments(ctx context.Context, req *lnrpc.ListPaymentsRequest, _ ...grpc.CallOption) (*lnrpc.ListPaymentsResponse, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	if len(s.pages) == 0 {
		return &lnrpc.ListPaymentsResponse{}, nil
	}
	page := s.pages[0]
	s.pages = s.pages[1:]
	return page, nil
}

type stubInvoicesAPI struct {
	pages []*lnrpc.ListInvoiceResponse
	err   error
	calls int
}

func (s *stubInvoicesAPI) ListInvoices(ctx context.Context, req *lnrpc.ListInvoiceRequest, _ ...grpc.CallOption) (*lnrpc.ListInvoiceResponse, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	if len(s.pages) == 0 {
		return &lnrpc.ListInvoiceResponse{}, nil
	}
	page := s.pages[0]
	s.pages = s.pages[1:]
	return page, nil
}

func settledPaymentFixture(hash string) *lnrpc.Payment {
	return &lnrpc.Payment{
		PaymentHash:     hash,
		Status:          lnrpc.Payment_SUCCEEDED,
		PaymentPreimage: hash,
		ValueMsat:       1000,
		FeeMsat:         10,
		CreationTimeNs:  time.Now().Add(-time.Hour).UnixNano(),
	}
}

func TestReconcileMissedEvents_PublishesTerminalStates(t *testing.T) {
	ctx := context.Background()
	publisher := &stubEventPublisher{}
	payments := &stubPaymentsAPI{
		pages: []*lnrpc.ListPaymentsResponse{
			{
				Payments: []*lnrpc.Payment{
					settledPaymentFixture("aa"),
					{
						PaymentHash:    "bb",
						Status:         lnrpc.Payment_FAILED,
						FailureReason:  lnrpc.PaymentFailureReason_FAILURE_REASON_NO_ROUTE,
						ValueMsat:      2000,
						CreationTimeNs: time.Now().Add(-time.Hour).UnixNano(),
					},
					{
						PaymentHash:    "cc",
						Status:         lnrpc.Payment_IN_FLIGHT,
						ValueMsat:      3000,
						CreationTimeNs: time.Now().Add(-time.Hour).UnixNano(),
					},
				},
				LastIndexOffset: 3,
			},
			{
				Payments:         []*lnrpc.Payment{},
				FirstIndexOffset: 3,
			},
		},
	}
	rHash, _ := hex.DecodeString("dd")
	invoices := &stubInvoicesAPI{
		pages: []*lnrpc.ListInvoiceResponse{
			{
				Invoices: []*lnrpc.Invoice{
					{
						RHash:        rHash,
						RPreimage:    rHash,
						State:        lnrpc.Invoice_SETTLED,
						ValueMsat:    4000,
						CreationDate: time.Now().Add(-time.Hour).Unix(),
						SettleDate:   time.Now().Add(-30 * time.Minute).Unix(),
					},
					{
						State:        lnrpc.Invoice_OPEN,
						ValueMsat:    5000,
						CreationDate: time.Now().Add(-time.Hour).Unix(),
					},
				},
				LastIndexOffset: 2,
			},
			{
				Invoices:         []*lnrpc.Invoice{},
				FirstIndexOffset: 2,
			},
		},
	}

	err := reconcileMissedEvents(ctx, payments, invoices, publisher, time.Now().Add(-24*time.Hour))
	require.NoError(t, err)

	published := publisher.published()
	require.Len(t, published, 3)

	byEvent := map[string]*events.Event{}
	for _, event := range published {
		byEvent[event.Event] = event
	}

	sent, ok := byEvent["nwc_lnclient_payment_sent"]
	require.True(t, ok, "expected payment sent event for settled payment")
	assert.Equal(t, "aa", sent.Properties.(*lnclient.Transaction).PaymentHash)

	failed, ok := byEvent["nwc_lnclient_payment_failed"]
	require.True(t, ok, "expected payment failed event for failed payment")
	failedProps := failed.Properties.(*lnclient.PaymentFailedEventProperties)
	assert.Equal(t, "bb", failedProps.Transaction.PaymentHash)
	assert.NotEmpty(t, failedProps.Reason)

	received, ok := byEvent["nwc_lnclient_payment_received"]
	require.True(t, ok, "expected payment received event for settled invoice")
	assert.Equal(t, "dd", received.Properties.(*lnclient.Transaction).PaymentHash)
}

func TestReconcileMissedEvents_EmptyListsPublishNothing(t *testing.T) {
	publisher := &stubEventPublisher{}
	err := reconcileMissedEvents(context.Background(), &stubPaymentsAPI{}, &stubInvoicesAPI{}, publisher, time.Now().Add(-24*time.Hour))
	require.NoError(t, err)
	assert.Empty(t, publisher.published())
}

func TestReconcileMissedEvents_PaymentsErrorStillReconcilesInvoices(t *testing.T) {
	publisher := &stubEventPublisher{}
	payments := &stubPaymentsAPI{err: errors.New("lnd unavailable")}
	rHash, _ := hex.DecodeString("dd")
	invoices := &stubInvoicesAPI{
		pages: []*lnrpc.ListInvoiceResponse{
			{
				Invoices: []*lnrpc.Invoice{
					{
						RHash:        rHash,
						RPreimage:    rHash,
						State:        lnrpc.Invoice_SETTLED,
						ValueMsat:    4000,
						CreationDate: time.Now().Add(-time.Hour).Unix(),
						SettleDate:   time.Now().Add(-30 * time.Minute).Unix(),
					},
				},
				LastIndexOffset: 1,
			},
			{
				Invoices:         []*lnrpc.Invoice{},
				FirstIndexOffset: 1,
			},
		},
	}

	err := reconcileMissedEvents(context.Background(), payments, invoices, publisher, time.Now().Add(-24*time.Hour))
	require.Error(t, err)
	require.Len(t, publisher.published(), 1)
	assert.Equal(t, "nwc_lnclient_payment_received", publisher.published()[0].Event)
}

func TestReconcileMissedEvents_RespectsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	publisher := &stubEventPublisher{}
	payments := &stubPaymentsAPI{}
	invoices := &stubInvoicesAPI{}

	err := reconcileMissedEvents(ctx, payments, invoices, publisher, time.Now().Add(-24*time.Hour))
	require.Error(t, err)
	assert.Zero(t, payments.calls)
	assert.Zero(t, invoices.calls)
	assert.Empty(t, publisher.published())
}
