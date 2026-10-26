package cln

import (
	"context"
	"errors"
	"testing"

	"github.com/getAlby/hub/lnclient"
	clngrpc "github.com/getAlby/hub/lnclient/cln/clngrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type stubListPaysClient struct {
	clngrpc.NodeClient
	lastRequest *clngrpc.ListpaysRequest
	listPaysFn  func(ctx context.Context, in *clngrpc.ListpaysRequest, opts ...grpc.CallOption) (*clngrpc.ListpaysResponse, error)
}

func (s *stubListPaysClient) ListPays(ctx context.Context, in *clngrpc.ListpaysRequest, opts ...grpc.CallOption) (*clngrpc.ListpaysResponse, error) {
	s.lastRequest = in
	return s.listPaysFn(ctx, in, opts...)
}

func amount(msat uint64) *clngrpc.Amount {
	return &clngrpc.Amount{Msat: msat}
}

func completePay(preimage []byte, amountMsat, amountSentMsat uint64) *clngrpc.ListpaysPays {
	return &clngrpc.ListpaysPays{
		Status:         clngrpc.ListpaysPays_COMPLETE,
		Preimage:       preimage,
		AmountMsat:     amount(amountMsat),
		AmountSentMsat: amount(amountSentMsat),
	}
}

func newLookupSvc(t *testing.T, fn func(context.Context, *clngrpc.ListpaysRequest, ...grpc.CallOption) (*clngrpc.ListpaysResponse, error)) (*CLNService, *stubListPaysClient) {
	t.Helper()
	stub := &stubListPaysClient{listPaysFn: fn}
	return &CLNService{ctx: context.Background(), client: stub}, stub
}

const testHash = "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"

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
			{Status: clngrpc.ListpaysPays_FAILED, AmountMsat: amount(100)},
			{Status: clngrpc.ListpaysPays_PENDING, AmountMsat: amount(100)},
		}}, nil
	})

	status, err := svc.LookupPaymentStatus(testHash)
	require.NoError(t, err)
	assert.Equal(t, lnclient.PAYMENT_STATUS_PENDING, status.Status)
}

func TestLookupPaymentStatus_CompleteDominatesAll(t *testing.T) {
	svc, _ := newLookupSvc(t, func(_ context.Context, _ *clngrpc.ListpaysRequest, _ ...grpc.CallOption) (*clngrpc.ListpaysResponse, error) {
		return &clngrpc.ListpaysResponse{Pays: []*clngrpc.ListpaysPays{
			{Status: clngrpc.ListpaysPays_PENDING, AmountMsat: amount(100)},
			completePay([]byte{0x01}, 100, 101),
			{Status: clngrpc.ListpaysPays_FAILED, AmountMsat: amount(100)},
		}}, nil
	})

	status, err := svc.LookupPaymentStatus(testHash)
	require.NoError(t, err)
	assert.Equal(t, lnclient.PAYMENT_STATUS_COMPLETE, status.Status)
	assert.Equal(t, uint64(1), status.FeeMsat)
}

func TestLookupPaymentStatus_AllFailed(t *testing.T) {
	svc, _ := newLookupSvc(t, func(_ context.Context, _ *clngrpc.ListpaysRequest, _ ...grpc.CallOption) (*clngrpc.ListpaysResponse, error) {
		return &clngrpc.ListpaysResponse{Pays: []*clngrpc.ListpaysPays{
			{Status: clngrpc.ListpaysPays_FAILED, AmountMsat: amount(100)},
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

func TestLookupPaymentStatus_InvalidHashIsError(t *testing.T) {
	svc, _ := newLookupSvc(t, func(_ context.Context, _ *clngrpc.ListpaysRequest, _ ...grpc.CallOption) (*clngrpc.ListpaysResponse, error) {
		return &clngrpc.ListpaysResponse{}, nil
	})

	_, err := svc.LookupPaymentStatus("not-hex")
	assert.Error(t, err)
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
