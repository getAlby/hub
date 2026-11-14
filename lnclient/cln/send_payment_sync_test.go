package cln

import (
	"context"
	"encoding/hex"
	"os"
	"testing"

	"github.com/getAlby/hub/lnclient"
	clngrpc "github.com/getAlby/hub/lnclient/cln/clngrpc"
	"github.com/getAlby/hub/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestMain(m *testing.M) {
	logger.Init("2") // error level: quiet test output, non-nil global logger
	os.Exit(m.Run())
}

type stubPayClient struct {
	clngrpc.NodeClient
	xpayFn func(ctx context.Context, in *clngrpc.XpayRequest, opts ...grpc.CallOption) (*clngrpc.XpayResponse, error)
}

func (s *stubPayClient) Decode(ctx context.Context, in *clngrpc.DecodeRequest, opts ...grpc.CallOption) (*clngrpc.DecodeResponse, error) {
	return &clngrpc.DecodeResponse{Valid: true}, nil
}

func (s *stubPayClient) Xpay(ctx context.Context, in *clngrpc.XpayRequest, opts ...grpc.CallOption) (*clngrpc.XpayResponse, error) {
	return s.xpayFn(ctx, in, opts...)
}

func newSendSvc(t *testing.T, fn func(context.Context, *clngrpc.XpayRequest, ...grpc.CallOption) (*clngrpc.XpayResponse, error)) *CLNService {
	t.Helper()
	stub := &stubPayClient{xpayFn: fn}
	return &CLNService{ctx: context.Background(), client: stub}
}

const testInvoice = "lnrtestinvoice"

// The xpay RPC has already been issued: the payment may exist on the node by
// the time the response is inspected. A nil response with a nil error must
// NOT panic (crashing the hub mid-payment) and must NOT be reported as
// success. The caller must see an error so the row stays reconcilable from
// backend state rather than being settled from a phantom response.
func TestSendPaymentSync_NilXpayResponse_FailsClosedWithoutPanic(t *testing.T) {
	svc := newSendSvc(t, func(_ context.Context, _ *clngrpc.XpayRequest, _ ...grpc.CallOption) (*clngrpc.XpayResponse, error) {
		return nil, nil
	})

	var resp *lnclient.PayInvoiceResponse
	var err error
	require.NotPanics(t, func() {
		resp, err = svc.SendPaymentSync(testInvoice, nil)
	})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// A non-nil xpay response without a 32-byte preimage does not prove
// settlement. Returning an empty preimage as success would persist a settled
// transaction with no recoverable preimage; the state must be reported as an
// error so the retry path can look the real result up on the node.
func TestSendPaymentSync_MissingPreimage_FailsClosed(t *testing.T) {
	svc := newSendSvc(t, func(_ context.Context, _ *clngrpc.XpayRequest, _ ...grpc.CallOption) (*clngrpc.XpayResponse, error) {
		return &clngrpc.XpayResponse{
			AmountMsat:     amount(100000),
			AmountSentMsat: amount(101000),
		}, nil
	})

	var resp *lnclient.PayInvoiceResponse
	var err error
	require.NotPanics(t, func() {
		resp, err = svc.SendPaymentSync(testInvoice, nil)
	})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

// Positive control: a well-formed xpay success still settles with the correct
// preimage and derived fee.
func TestSendPaymentSync_ValidResponse_Succeeds(t *testing.T) {
	preimage := make([]byte, 32)
	preimage[0] = 0xcd
	svc := newSendSvc(t, func(_ context.Context, _ *clngrpc.XpayRequest, _ ...grpc.CallOption) (*clngrpc.XpayResponse, error) {
		return &clngrpc.XpayResponse{
			PaymentPreimage: preimage,
			AmountMsat:      amount(100000),
			AmountSentMsat:  amount(101500),
		}, nil
	})

	resp, err := svc.SendPaymentSync(testInvoice, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, hex.EncodeToString(preimage), resp.Preimage)
	assert.Equal(t, uint64(1500), resp.FeeMsat)
}
