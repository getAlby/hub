package cln

import (
	"context"
	"net"
	"testing"
	"time"

	clngrpc "github.com/getAlby/hub/lnclient/cln/clngrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

// hungNodeServer answers ListPays only after a long delay, emulating a CLN
// connection that accepts the RPC but never finishes it.
type hungNodeServer struct {
	clngrpc.UnimplementedNodeServer
}

func (s *hungNodeServer) ListPays(ctx context.Context, in *clngrpc.ListpaysRequest) (*clngrpc.ListpaysResponse, error) {
	select {
	case <-ctx.Done():
		return nil, status.FromContextError(ctx.Err()).Err()
	case <-time.After(30 * time.Second):
		return &clngrpc.ListpaysResponse{}, nil
	}
}

// The listpays deadline must be enforced by the real gRPC transport, not only
// observed in synthetic stubs: a hung (not refused) node connection must
// return DeadlineExceeded to the caller — which fails closed — while the hub
// holds its global payment validation lock.
func TestLookupPaymentStatus_BufconnRealTransportHonoursDeadline(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	clngrpc.RegisterNodeServer(server, &hungNodeServer{})
	go func() { _ = server.Serve(lis) }()
	defer server.Stop()

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	original := lookupPaymentStatusTimeout
	lookupPaymentStatusTimeout = 100 * time.Millisecond
	defer func() { lookupPaymentStatusTimeout = original }()

	svc := &CLNService{ctx: context.Background(), client: clngrpc.NewNodeClient(conn)}

	start := time.Now()
	_, err = svc.LookupPaymentStatus(testHash)
	elapsed := time.Since(start)

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok, "error must carry a gRPC status from the real transport")
	assert.Equal(t, codes.DeadlineExceeded, st.Code())
	assert.Less(t, elapsed, 5*time.Second, "hung RPC must be cut off by the deadline, not left to stall the payment lock")
}
