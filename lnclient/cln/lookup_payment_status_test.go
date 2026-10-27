package cln

import (
	"context"
	"encoding/hex"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/getAlby/hub/lnclient"
	clngrpc "github.com/getAlby/hub/lnclient/cln/clngrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type stubListPaysClient struct {
	clngrpc.NodeClient
	mu          sync.Mutex
	lastRequest *clngrpc.ListpaysRequest
	listPaysFn  func(ctx context.Context, in *clngrpc.ListpaysRequest, opts ...grpc.CallOption) (*clngrpc.ListpaysResponse, error)
}

func (s *stubListPaysClient) ListPays(ctx context.Context, in *clngrpc.ListpaysRequest, opts ...grpc.CallOption) (*clngrpc.ListpaysResponse, error) {
	s.mu.Lock()
	s.lastRequest = in
	s.mu.Unlock()
	return s.listPaysFn(ctx, in, opts...)
}

func amount(msat uint64) *clngrpc.Amount {
	return &clngrpc.Amount{Msat: msat}
}

const testHash = "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"

func testHashBytes() []byte {
	b, _ := hex.DecodeString(testHash)
	return b
}

func completePay(preimage []byte, amountMsat, amountSentMsat uint64) *clngrpc.ListpaysPays {
	return &clngrpc.ListpaysPays{
		PaymentHash:    testHashBytes(),
		Status:         clngrpc.ListpaysPays_COMPLETE,
		Preimage:       preimage,
		AmountMsat:     amount(amountMsat),
		AmountSentMsat: amount(amountSentMsat),
	}
}

func statusPay(status clngrpc.ListpaysPays_ListpaysPaysStatus, amountMsat uint64) *clngrpc.ListpaysPays {
	return &clngrpc.ListpaysPays{
		PaymentHash: testHashBytes(),
		Status:      status,
		AmountMsat:  amount(amountMsat),
	}
}

func newLookupSvc(t *testing.T, fn func(context.Context, *clngrpc.ListpaysRequest, ...grpc.CallOption) (*clngrpc.ListpaysResponse, error)) (*CLNService, *stubListPaysClient) {
	t.Helper()
	stub := &stubListPaysClient{listPaysFn: fn}
	return &CLNService{ctx: context.Background(), client: stub}, stub
}

func TestLookupPaymentStatus_Complete(t *testing.T) {
	preimage := make([]byte, 32)
	preimage[0] = 0xab
	svc, stub := newLookupSvc(t, func(_ context.Context, in *clngrpc.ListpaysRequest, _ ...grpc.CallOption) (*clngrpc.ListpaysResponse, error) {
		return &clngrpc.ListpaysResponse{Pays: []*clngrpc.ListpaysPays{
			completePay(preimage, 100000, 105000),
		}}, nil
	})

	status, err := svc.LookupPaymentStatus(testHash)
	require.NoError(t, err)
	assert.Equal(t, lnclient.PAYMENT_STATUS_COMPLETE, status.Status)
	assert.Equal(t, "ab"+""+repeatZeroes(62), status.Preimage)
	assert.Equal(t, uint64(5000), status.FeeMsat)
	assert.Equal(t, uint64(100000), status.AmountMsat)

	// request must filter by the raw 32-byte hash
	require.NotNil(t, stub.lastRequest)
	assert.Len(t, stub.lastRequest.PaymentHash, 32)
}

func TestLookupPaymentStatus_PendingBeatsFailed(t *testing.T) {
	svc, _ := newLookupSvc(t, func(_ context.Context, _ *clngrpc.ListpaysRequest, _ ...grpc.CallOption) (*clngrpc.ListpaysResponse, error) {
		return &clngrpc.ListpaysResponse{Pays: []*clngrpc.ListpaysPays{
			statusPay(clngrpc.ListpaysPays_FAILED, 100),
			statusPay(clngrpc.ListpaysPays_PENDING, 100),
		}}, nil
	})

	status, err := svc.LookupPaymentStatus(testHash)
	require.NoError(t, err)
	assert.Equal(t, lnclient.PAYMENT_STATUS_PENDING, status.Status)
}

// Historical FAILED attempts alongside a later settled group: COMPLETE must
// win so the hub reconciles instead of re-paying.
func TestLookupPaymentStatus_CompleteDominatesAll(t *testing.T) {
	preimage := make([]byte, 32)
	preimage[31] = 0x01
	svc, _ := newLookupSvc(t, func(_ context.Context, _ *clngrpc.ListpaysRequest, _ ...grpc.CallOption) (*clngrpc.ListpaysResponse, error) {
		return &clngrpc.ListpaysResponse{Pays: []*clngrpc.ListpaysPays{
			statusPay(clngrpc.ListpaysPays_PENDING, 100),
			completePay(preimage, 100, 101),
			statusPay(clngrpc.ListpaysPays_FAILED, 100),
		}}, nil
	})

	status, err := svc.LookupPaymentStatus(testHash)
	require.NoError(t, err)
	assert.Equal(t, lnclient.PAYMENT_STATUS_COMPLETE, status.Status)
	assert.Equal(t, uint64(1), status.FeeMsat)
	assert.Len(t, status.Preimage, 64)
}

func TestLookupPaymentStatus_AllFailed(t *testing.T) {
	svc, _ := newLookupSvc(t, func(_ context.Context, _ *clngrpc.ListpaysRequest, _ ...grpc.CallOption) (*clngrpc.ListpaysResponse, error) {
		return &clngrpc.ListpaysResponse{Pays: []*clngrpc.ListpaysPays{
			statusPay(clngrpc.ListpaysPays_FAILED, 100),
		}}, nil
	})

	status, err := svc.LookupPaymentStatus(testHash)
	require.NoError(t, err)
	assert.Equal(t, lnclient.PAYMENT_STATUS_FAILED, status.Status)
}

func TestLookupPaymentStatus_NoRecordIsUnknown(t *testing.T) {
	svc, _ := newLookupSvc(t, func(_ context.Context, _ *clngrpc.ListpaysRequest, _ ...grpc.CallOption) (*clngrpc.ListpaysResponse, error) {
		return &clngrpc.ListpaysResponse{Pays: nil}, nil
	})

	status, err := svc.LookupPaymentStatus(testHash)
	require.NoError(t, err)
	assert.Equal(t, lnclient.PAYMENT_STATUS_UNKNOWN, status.Status)
}

func TestLookupPaymentStatus_RpcErrorPropagates(t *testing.T) {
	svc, _ := newLookupSvc(t, func(_ context.Context, _ *clngrpc.ListpaysRequest, _ ...grpc.CallOption) (*clngrpc.ListpaysResponse, error) {
		return nil, errors.New("rpc deadline exceeded")
	})

	_, err := svc.LookupPaymentStatus(testHash)
	assert.Error(t, err)
}

// A nil response with a nil error is malformed, not a confirmed absence.
func TestLookupPaymentStatus_NilResponseIsError(t *testing.T) {
	svc, _ := newLookupSvc(t, func(_ context.Context, _ *clngrpc.ListpaysRequest, _ ...grpc.CallOption) (*clngrpc.ListpaysResponse, error) {
		return nil, nil
	})

	_, err := svc.LookupPaymentStatus(testHash)
	assert.Error(t, err)
}

// Records whose payment_hash does not exactly match the requested hash make
// the result inconsistent — never interpret another payment's state.
func TestLookupPaymentStatus_HashMismatchIsError(t *testing.T) {
	other := make([]byte, 32)
	other[0] = 0xff
	svc, _ := newLookupSvc(t, func(_ context.Context, _ *clngrpc.ListpaysRequest, _ ...grpc.CallOption) (*clngrpc.ListpaysResponse, error) {
		return &clngrpc.ListpaysResponse{Pays: []*clngrpc.ListpaysPays{
			{PaymentHash: other, Status: clngrpc.ListpaysPays_COMPLETE, Preimage: make([]byte, 32), AmountMsat: amount(100), AmountSentMsat: amount(100)},
		}}, nil
	})

	_, err := svc.LookupPaymentStatus(testHash)
	assert.Error(t, err)
}

func TestLookupPaymentStatus_MissingHashInRecordIsError(t *testing.T) {
	svc, _ := newLookupSvc(t, func(_ context.Context, _ *clngrpc.ListpaysRequest, _ ...grpc.CallOption) (*clngrpc.ListpaysResponse, error) {
		return &clngrpc.ListpaysResponse{Pays: []*clngrpc.ListpaysPays{
			{Status: clngrpc.ListpaysPays_PENDING, AmountMsat: amount(100)},
		}}, nil
	})

	_, err := svc.LookupPaymentStatus(testHash)
	assert.Error(t, err)
}

// An enum value outside PENDING/FAILED/COMPLETE is ambiguous state, not
// absence — fail closed so the caller never reads it as permission to pay.
func TestLookupPaymentStatus_UnrecognizedStatusIsError(t *testing.T) {
	svc, _ := newLookupSvc(t, func(_ context.Context, _ *clngrpc.ListpaysRequest, _ ...grpc.CallOption) (*clngrpc.ListpaysResponse, error) {
		return &clngrpc.ListpaysResponse{Pays: []*clngrpc.ListpaysPays{
			statusPay(clngrpc.ListpaysPays_ListpaysPaysStatus(99), 100),
		}}, nil
	})

	_, err := svc.LookupPaymentStatus(testHash)
	assert.Error(t, err)
}

// COMPLETE without a valid 32-byte preimage cannot settle anything: treat the
// response as malformed rather than reconciling a bogus preimage.
func TestLookupPaymentStatus_CompleteWithoutPreimageIsError(t *testing.T) {
	svc, _ := newLookupSvc(t, func(_ context.Context, _ *clngrpc.ListpaysRequest, _ ...grpc.CallOption) (*clngrpc.ListpaysResponse, error) {
		return &clngrpc.ListpaysResponse{Pays: []*clngrpc.ListpaysPays{
			{PaymentHash: testHashBytes(), Status: clngrpc.ListpaysPays_COMPLETE, AmountMsat: amount(100), AmountSentMsat: amount(101)},
		}}, nil
	})

	_, err := svc.LookupPaymentStatus(testHash)
	assert.Error(t, err)
}

func TestLookupPaymentStatus_InvalidHashIsError(t *testing.T) {
	svc, _ := newLookupSvc(t, func(_ context.Context, _ *clngrpc.ListpaysRequest, _ ...grpc.CallOption) (*clngrpc.ListpaysResponse, error) {
		return &clngrpc.ListpaysResponse{}, nil
	})

	_, err := svc.LookupPaymentStatus("not-hex")
	assert.Error(t, err)
}

// The lookup runs while the hub holds its global payment validation lock, so
// the RPC must carry a per-call deadline — a hung (not refused) node socket
// may never block all payment activity.
func TestLookupPaymentStatus_SetsDeadlineOnListPays(t *testing.T) {
	var captured context.Context
	svc, _ := newLookupSvc(t, func(ctx context.Context, _ *clngrpc.ListpaysRequest, _ ...grpc.CallOption) (*clngrpc.ListpaysResponse, error) {
		captured = ctx
		return &clngrpc.ListpaysResponse{}, nil
	})

	_, err := svc.LookupPaymentStatus(testHash)
	require.NoError(t, err)
	require.NotNil(t, captured)
	_, ok := captured.Deadline()
	assert.True(t, ok, "listpays must run under a per-call deadline")
}

func TestLookupPaymentStatus_HungRpcTimesOutFailsClosed(t *testing.T) {
	original := lookupPaymentStatusTimeout
	lookupPaymentStatusTimeout = 20 * time.Millisecond
	defer func() { lookupPaymentStatusTimeout = original }()

	svc, _ := newLookupSvc(t, func(ctx context.Context, _ *clngrpc.ListpaysRequest, _ ...grpc.CallOption) (*clngrpc.ListpaysResponse, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
			return nil, errors.New("test harness: listpays was never cancelled")
		}
	})

	_, err := svc.LookupPaymentStatus(testHash)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

// CLNService must satisfy the optional capability interface used by the
// transactions service dedup check.
func TestCLNServiceImplementsPaymentStatusLookup(t *testing.T) {
	var svc interface{} = &CLNService{}
	_, ok := svc.(lnclient.PaymentStatusLookup)
	assert.True(t, ok)
}

func repeatZeroes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = '0'
	}
	return string(b)
}
