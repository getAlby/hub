package greenlight

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/encoding/protowire"

	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/lnclient/cln/clngrpc"
	"github.com/getAlby/hub/logger"
	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	logger.Init(strconv.Itoa(int(logrus.WarnLevel)))
	os.Exit(m.Run())
}

// mockNode implements the cln.Node gRPC surface served by Greenlight nodes,
// backed by in-memory state. Only the RPCs the backend actually calls are
// implemented; everything else falls through to UnimplementedNodeServer.
type mockNode struct {
	clngrpc.UnimplementedNodeServer

	mu           sync.Mutex
	invoices     map[string]*clngrpc.ListinvoicesInvoices // key: payment hash hex
	preimages    map[string][]byte                        // key: payment hash hex
	nextPayIndex uint64
	keysendCh    chan []byte // raw IncomingPayment payloads pushed by tests
}

func newMockNode() *mockNode {
	return &mockNode{
		invoices:  map[string]*clngrpc.ListinvoicesInvoices{},
		preimages: map[string][]byte{},
		keysendCh: make(chan []byte, 10),
	}
}

func (m *mockNode) Getinfo(ctx context.Context, req *clngrpc.GetinfoRequest) (*clngrpc.GetinfoResponse, error) {
	return &clngrpc.GetinfoResponse{
		Id:          mustHex("02f1d2e5f1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d"),
		Alias:       "mock-greenlight-node",
		Color:       mustHex("02f1d2"),
		Network:     "bitcoin",
		Blockheight: 840000,
		Version:     "v26.06",
	}, nil
}

func (m *mockNode) Decode(ctx context.Context, req *clngrpc.DecodeRequest) (*clngrpc.DecodeResponse, error) {
	// bolt11 only: "lnbc<amount>...<paymenthash>"
	amountMsat := uint64(0)
	parts := strings.Split(req.String_, ":")
	if len(parts) == 2 {
		amountMsat, _ = strconv.ParseUint(parts[0][5:], 10, 64)
	}
	now := uint64(time.Now().Unix())
	descHash := sha256.Sum256([]byte("mock invoice"))
	return &clngrpc.DecodeResponse{
		ItemType:        clngrpc.DecodeResponse_BOLT11_INVOICE,
		Valid:           true,
		CreatedAt:       &now,
		DescriptionHash: descHash[:],
		AmountMsat:      &clngrpc.Amount{Msat: amountMsat},
	}, nil
}

func (m *mockNode) Xpay(ctx context.Context, req *clngrpc.XpayRequest) (*clngrpc.XpayResponse, error) {
	preimage := []byte("mock-preimage-0123456789abcdef")
	amount := uint64(1000)
	if req.AmountMsat != nil {
		amount = req.AmountMsat.Msat
	}
	return &clngrpc.XpayResponse{
		PaymentPreimage: preimage,
		AmountMsat:      &clngrpc.Amount{Msat: amount},
		AmountSentMsat:  &clngrpc.Amount{Msat: amount + 10},
	}, nil
}

func (m *mockNode) KeySend(ctx context.Context, req *clngrpc.KeysendRequest) (*clngrpc.KeysendResponse, error) {
	return &clngrpc.KeysendResponse{
		PaymentPreimage: []byte("mock-keysend-preimage-1234567890"),
		AmountMsat:      &clngrpc.Amount{Msat: req.AmountMsat.Msat},
		AmountSentMsat:  &clngrpc.Amount{Msat: req.AmountMsat.Msat + 5},
	}, nil
}

func (m *mockNode) Invoice(ctx context.Context, req *clngrpc.InvoiceRequest) (*clngrpc.InvoiceResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	preimage := req.Preimage
	paymentHash := sha256.Sum256(preimage)
	phHex := hex.EncodeToString(paymentHash[:])
	amountMsat := uint64(0)
	if req.AmountMsat.GetAmount() != nil {
		amountMsat = req.AmountMsat.GetAmount().Msat
	}
	expiry := uint64(3600)
	if req.Expiry != nil {
		expiry = *req.Expiry
	}
	bolt11 := fmt.Sprintf("lnbc%s...%s", strconv.FormatUint(amountMsat/1000, 10), phHex)
	desc := ""
	if req.Description != "" {
		desc = req.Description
	}
	now := time.Now().Unix()

	m.invoices[phHex] = &clngrpc.ListinvoicesInvoices{
		Label:       req.Label,
		Description: &desc,
		PaymentHash: paymentHash[:],
		Status:      clngrpc.ListinvoicesInvoices_UNPAID,
		ExpiresAt:   uint64(now) + expiry,
		AmountMsat:  &clngrpc.Amount{Msat: amountMsat},
		Bolt11:      &bolt11,
	}
	m.preimages[phHex] = preimage

	return &clngrpc.InvoiceResponse{
		Bolt11:      bolt11,
		PaymentHash: paymentHash[:],
		ExpiresAt:   uint64(now) + expiry,
	}, nil
}

func (m *mockNode) ListInvoices(ctx context.Context, req *clngrpc.ListinvoicesRequest) (*clngrpc.ListinvoicesResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	results := []*clngrpc.ListinvoicesInvoices{}
	if len(req.PaymentHash) > 0 {
		if inv, ok := m.invoices[hex.EncodeToString(req.PaymentHash)]; ok {
			results = append(results, inv)
		}
	} else {
		for _, inv := range m.invoices {
			results = append(results, inv)
		}
	}
	return &clngrpc.ListinvoicesResponse{Invoices: results}, nil
}

// markPaid simulates the node receiving a payment for the invoice.
func (m *mockNode) markPaid(paymentHashHex string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	inv, ok := m.invoices[paymentHashHex]
	if !ok {
		return
	}
	m.nextPayIndex++
	payIndex := m.nextPayIndex
	paidAt := uint64(time.Now().Unix())
	inv.Status = clngrpc.ListinvoicesInvoices_PAID
	inv.PaidAt = &paidAt
	inv.AmountReceivedMsat = inv.AmountMsat
	inv.PaymentPreimage = m.preimages[paymentHashHex]
	inv.PayIndex = &payIndex
}

func (m *mockNode) WaitAnyInvoice(ctx context.Context, req *clngrpc.WaitanyinvoiceRequest) (*clngrpc.WaitanyinvoiceResponse, error) {
	// cap the wait so tests don't block on the pump's 30s timeout
	wait := 500 * time.Millisecond
	deadline := time.Now().Add(wait)
	for {
		m.mu.Lock()
		var best *clngrpc.ListinvoicesInvoices
		var bestIndex uint64
		for _, inv := range m.invoices {
			if inv.PayIndex == nil || inv.Status != clngrpc.ListinvoicesInvoices_PAID {
				continue
			}
			if *inv.PayIndex > lastpayIndexOrZero(req) && (best == nil || *inv.PayIndex < bestIndex) {
				best = inv
				bestIndex = *inv.PayIndex
			}
		}
		m.mu.Unlock()

		if best != nil {
			resp := &clngrpc.WaitanyinvoiceResponse{
				Label:              best.Label,
				Description:        best.Description,
				PaymentHash:        best.PaymentHash,
				Status:             clngrpc.WaitanyinvoiceResponse_PAID,
				ExpiresAt:          best.ExpiresAt,
				AmountMsat:         best.AmountMsat,
				Bolt11:             best.Bolt11,
				Bolt12:             best.Bolt12,
				PayIndex:           best.PayIndex,
				AmountReceivedMsat: best.AmountReceivedMsat,
				PaidAt:             best.PaidAt,
				PaymentPreimage:    best.PaymentPreimage,
			}
			return resp, nil
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("waitanyinvoice timed out")
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(25 * time.Millisecond):
		}
	}
}

func lastpayIndexOrZero(req *clngrpc.WaitanyinvoiceRequest) uint64 {
	if req.LastpayIndex != nil {
		return *req.LastpayIndex
	}
	return 0
}

func (m *mockNode) ListPeerChannels(ctx context.Context, req *clngrpc.ListpeerchannelsRequest) (*clngrpc.ListpeerchannelsResponse, error) {
	scid := "800000x1x0"
	state := clngrpc.ChannelState_ChanneldNormal
	toUs := uint64(150000)
	total := uint64(200000)
	spendable := uint64(145000)
	receivable := uint64(49000000)
	opener := clngrpc.ChannelSide_LOCAL
	private := false
	connected := true
	feeBase := uint64(1000)
	feePPM := uint32(1)

	updates := &clngrpc.ListpeerchannelsChannelsUpdates{
		Local: &clngrpc.ListpeerchannelsChannelsUpdatesLocal{
			FeeBaseMsat:               &clngrpc.Amount{Msat: feeBase},
			FeeProportionalMillionths: feePPM,
		},
	}

	ch := &clngrpc.ListpeerchannelsChannels{
		ChannelId:        mustHex("ab01cd02ef030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e"),
		PeerId:           mustHex("02f1d2e5f1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d"),
		FundingTxid:      mustHex("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"),
		FundingOutnum:    uint32Ptr(1),
		ShortChannelId:   &scid,
		State:            state,
		PeerConnected:    connected,
		ToUsMsat:         &clngrpc.Amount{Msat: toUs},
		TotalMsat:        &clngrpc.Amount{Msat: total},
		SpendableMsat:    &clngrpc.Amount{Msat: spendable},
		ReceivableMsat:   &clngrpc.Amount{Msat: receivable},
		Opener:           opener,
		Private:          &private,
		OurReserveMsat:   &clngrpc.Amount{Msat: 5000},
		TheirReserveMsat: &clngrpc.Amount{Msat: 5000},
		Updates:          updates,
	}

	return &clngrpc.ListpeerchannelsResponse{Channels: []*clngrpc.ListpeerchannelsChannels{ch}}, nil
}

func (m *mockNode) ListFunds(ctx context.Context, req *clngrpc.ListfundsRequest) (*clngrpc.ListfundsResponse, error) {
	return &clngrpc.ListfundsResponse{
		Outputs: []*clngrpc.ListfundsOutputs{
			{
				AmountMsat: &clngrpc.Amount{Msat: 5000000},
				Status:     clngrpc.ListfundsOutputs_CONFIRMED,
				Reserved:   false,
				Txid:       mustHex("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"),
				Output:     0,
			},
		},
		Channels: []*clngrpc.ListfundsChannels{
			{
				OurAmountMsat: &clngrpc.Amount{Msat: 150000},
				State:         clngrpc.ChannelState_ChanneldNormal,
				ChannelId:     mustHex("ab01cd02ef030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e"),
				PeerId:        mustHex("02f1d2e5f1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d"),
			},
		},
	}, nil
}

func (m *mockNode) BkprListAccountEvents(ctx context.Context, req *clngrpc.BkprlistaccounteventsRequest) (*clngrpc.BkprlistaccounteventsResponse, error) {
	blockheight := uint32(840000)
	return &clngrpc.BkprlistaccounteventsResponse{
		Events: []*clngrpc.BkprlistaccounteventsEvents{
			{
				ItemType:    clngrpc.BkprlistaccounteventsEvents_CHAIN,
				CreditMsat:  &clngrpc.Amount{Msat: 1000000},
				DebitMsat:   nil,
				Blockheight: &blockheight,
				Txid:        mustHex("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"),
				Timestamp:   1717000000,
			},
		},
	}, nil
}

func (m *mockNode) Withdraw(ctx context.Context, req *clngrpc.WithdrawRequest) (*clngrpc.WithdrawResponse, error) {
	return &clngrpc.WithdrawResponse{
		Txid: mustHex("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"),
	}, nil
}

func (m *mockNode) FundChannel(ctx context.Context, req *clngrpc.FundchannelRequest) (*clngrpc.FundchannelResponse, error) {
	return &clngrpc.FundchannelResponse{
		Txid: mustHex("0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"),
	}, nil
}

func (m *mockNode) Close(ctx context.Context, req *clngrpc.CloseRequest) (*clngrpc.CloseResponse, error) {
	return &clngrpc.CloseResponse{}, nil
}

func (m *mockNode) SetChannel(ctx context.Context, req *clngrpc.SetchannelRequest) (*clngrpc.SetchannelResponse, error) {
	return &clngrpc.SetchannelResponse{}, nil
}

func (m *mockNode) ConnectPeer(ctx context.Context, req *clngrpc.ConnectRequest) (*clngrpc.ConnectResponse, error) {
	return &clngrpc.ConnectResponse{}, nil
}

func (m *mockNode) Disconnect(ctx context.Context, req *clngrpc.DisconnectRequest) (*clngrpc.DisconnectResponse, error) {
	return &clngrpc.DisconnectResponse{}, nil
}

func (m *mockNode) Offer(ctx context.Context, req *clngrpc.OfferRequest) (*clngrpc.OfferResponse, error) {
	return &clngrpc.OfferResponse{
		Bolt12: "lno1mockoffertest1234567890",
	}, nil
}

func (m *mockNode) NewAddr(ctx context.Context, req *clngrpc.NewaddrRequest) (*clngrpc.NewaddrResponse, error) {
	addr := "bc1qmockaddress0123456789abcdef"
	return &clngrpc.NewaddrResponse{
		Bech32: &addr,
	}, nil
}

func (m *mockNode) ListPeers(ctx context.Context, req *clngrpc.ListpeersRequest) (*clngrpc.ListpeersResponse, error) {
	return &clngrpc.ListpeersResponse{
		Peers: []*clngrpc.ListpeersPeers{
			{
				Id:          mustHex("02f1d2e5f1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d"),
				NumChannels: 1,
				Connected:   true,
			},
		},
	}, nil
}

func (m *mockNode) ListNodes(ctx context.Context, req *clngrpc.ListnodesRequest) (*clngrpc.ListnodesResponse, error) {
	addr := "1.2.3.4"
	alias := "mock-peer"
	return &clngrpc.ListnodesResponse{
		Nodes: []*clngrpc.ListnodesNodes{
			{
				Nodeid: mustHex("02f1d2e5f1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d"),
				Alias:  &alias,
				Color:  mustHex("02f1d2"),
				Addresses: []*clngrpc.ListnodesNodesAddresses{
					{
						ItemType: clngrpc.ListnodesNodesAddresses_IPV4,
						Address:  &addr,
						Port:     9735,
					},
				},
				Features: mustHex("02aa"),
			},
		},
	}, nil
}

func (m *mockNode) ListChannels(ctx context.Context, req *clngrpc.ListchannelsRequest) (*clngrpc.ListchannelsResponse, error) {
	scid := "800000x1x0"
	active := true
	public := true
	return &clngrpc.ListchannelsResponse{
		Channels: []*clngrpc.ListchannelsChannels{
			{
				ShortChannelId: scid,
				Direction:      0,
				Source:         mustHex("02f1d2e5f1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d"),
				Destination:    mustHex("02f1d2e5f1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d"),
				AmountMsat:     &clngrpc.Amount{Msat: 200000},
				Active:         active,
				Public:         public,
			},
		},
	}, nil
}

func (m *mockNode) SignMessage(ctx context.Context, req *clngrpc.SignmessageRequest) (*clngrpc.SignmessageResponse, error) {
	return &clngrpc.SignmessageResponse{
		Zbase: "mockzbase0123456789",
	}, nil
}

func (m *mockNode) SubscribeChannelStateChanged(req *clngrpc.StreamChannelStateChangedRequest, srv clngrpc.Node_SubscribeChannelStateChangedServer) error {
	<-srv.Context().Done()
	return nil
}

// --- test helpers ---

func mustHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func uint32Ptr(v uint32) *uint32 {
	return &v
}

func startMockNode(t *testing.T, node *mockNode) (clngrpc.NodeClient, *grpc.ClientConn, func()) {
	t.Helper()

	lis := bufconn.Listen(1024 * 1024)
	// rawCodec on the server handles both the generated cln.Node messages
	// (via the proto.Message fallback) and the raw StreamIncoming payloads.
	srv := grpc.NewServer(grpc.ForceServerCodec(rawCodec{}))
	clngrpc.RegisterNodeServer(srv, node)

	// register the greenlight.Node StreamIncoming stream (no generated
	// bindings; the server handler passes raw IncomingPayment bytes through)
	srv.RegisterService(&grpc.ServiceDesc{
		ServiceName: "greenlight.Node",
		HandlerType: (*interface{})(nil),
		Streams: []grpc.StreamDesc{{
			StreamName:    "StreamIncoming",
			Handler:       streamIncomingServerHandler(node),
			ServerStreams: true,
		}},
		Metadata: "greenlight.proto",
	}, nil)

	go func() {
		if err := srv.Serve(lis); err != nil {
			logger.Logger.WithError(err).Error("mock node server stopped")
		}
	}()

	conn, err := grpc.NewClient("passthrough:///bufconn",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to dial mock node: %v", err)
	}

	client := clngrpc.NewNodeClient(conn)

	cleanup := func() {
		conn.Close()
		srv.Stop()
		lis.Close()
	}

	return client, conn, cleanup
}

// streamIncomingServerHandler delivers raw IncomingPayment payloads pushed by
// tests via mockNode.keysendCh, then blocks until the stream is cancelled.
func streamIncomingServerHandler(node *mockNode) grpc.StreamHandler {
	return func(srv interface{}, stream grpc.ServerStream) error {
		req := &emptyMessage{}
		if err := stream.RecvMsg(req); err != nil {
			return err
		}
		for {
			select {
			case <-stream.Context().Done():
				return nil
			case payload := <-node.keysendCh:
				if err := stream.SendMsg(payload); err != nil {
					return err
				}
			}
		}
	}
}

// encodeIncomingPaymentForTest builds the raw protobuf wire bytes of a
// greenlight.IncomingPayment{ offchain { ... } } message.
func encodeIncomingPaymentForTest(preimage, paymentHash []byte, amountMsat uint64, tlvType uint64, tlvValue []byte, bolt11 string) []byte {
	off := protowire.AppendTag(nil, 1, protowire.BytesType)
	off = protowire.AppendString(off, "test-keysend")
	off = protowire.AppendTag(off, 2, protowire.BytesType)
	off = protowire.AppendBytes(off, preimage)
	amt := protowire.AppendTag(nil, 1, protowire.VarintType)
	amt = protowire.AppendVarint(amt, amountMsat)
	off = protowire.AppendTag(off, 3, protowire.BytesType)
	off = protowire.AppendBytes(off, amt)
	if tlvType != 0 {
		tf := protowire.AppendTag(nil, 1, protowire.VarintType)
		tf = protowire.AppendVarint(tf, tlvType)
		tf = protowire.AppendTag(tf, 2, protowire.BytesType)
		tf = protowire.AppendBytes(tf, tlvValue)
		off = protowire.AppendTag(off, 4, protowire.BytesType)
		off = protowire.AppendBytes(off, tf)
	}
	off = protowire.AppendTag(off, 5, protowire.BytesType)
	off = protowire.AppendBytes(off, paymentHash)
	if bolt11 != "" {
		off = protowire.AppendTag(off, 6, protowire.BytesType)
		off = protowire.AppendString(off, bolt11)
	}
	// outer IncomingPayment: oneof details { OffChainPayment offchain = 1; }
	msg := protowire.AppendTag(nil, 1, protowire.BytesType)
	return protowire.AppendBytes(msg, off)
}

func newTestService(t *testing.T, node *mockNode) (*GreenlightService, func()) {
	t.Helper()
	return newTestServiceWithWorkDir(t, node, t.TempDir())
}

func newTestServiceWithWorkDir(t *testing.T, node *mockNode, workDir string) (*GreenlightService, func()) {
	t.Helper()

	client, conn, cleanup := startMockNode(t, node)
	ctx, cancel := context.WithCancel(context.Background())

	svc := &GreenlightService{
		ctx:     ctx,
		cancel:  cancel,
		client:  client,
		conn:    conn,
		config:  Config{},
		workDir: workDir,
	}

	info, err := svc.GetInfo(ctx)
	if err != nil {
		cleanup()
		t.Fatalf("GetInfo failed: %v", err)
	}
	svc.pubkey = info.Pubkey

	return svc, func() {
		cancel()
		cleanup()
	}
}

type capturePublisher struct {
	mu     sync.Mutex
	events []*events.Event
}

func (c *capturePublisher) RegisterSubscriber(listener events.EventSubscriber) {}
func (c *capturePublisher) RemoveSubscriber(listener events.EventSubscriber)   {}
func (c *capturePublisher) Publish(e *events.Event) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, e)
}

func (c *capturePublisher) PublishSync(e *events.Event) {
	c.Publish(e)
}

func (c *capturePublisher) SetGlobalProperty(key string, value interface{}) {}

func (c *capturePublisher) waitForEvent(eventName string, timeout time.Duration) *events.Event {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		c.mu.Lock()
		for i, e := range c.events {
			if e.Event == eventName {
				c.events = append(c.events[:i], c.events[i+1:]...)
				c.mu.Unlock()
				return e
			}
		}
		c.mu.Unlock()
		time.Sleep(25 * time.Millisecond)
	}
	return nil
}

// --- tests ---

func TestGetInfo(t *testing.T) {
	node := newMockNode()
	svc, cleanup := newTestService(t, node)
	defer cleanup()

	info, err := svc.GetInfo(context.Background())
	if err != nil {
		t.Fatalf("GetInfo failed: %v", err)
	}
	if info.Network != "bitcoin" {
		t.Errorf("unexpected network: %s", info.Network)
	}
	if len(info.Pubkey) != 66 {
		t.Errorf("unexpected pubkey length: %s", info.Pubkey)
	}
	if svc.GetPubkey() != info.Pubkey {
		t.Errorf("pubkey mismatch")
	}
}

func TestSendPaymentSync(t *testing.T) {
	svc, cleanup := newTestService(t, newMockNode())
	defer cleanup()

	resp, err := svc.SendPaymentSync("lnbc1000:deadbeef", nil)
	if err != nil {
		t.Fatalf("SendPaymentSync failed: %v", err)
	}
	if resp.Preimage == "" {
		t.Errorf("expected preimage")
	}
	if resp.FeeMsat != 10 {
		t.Errorf("expected fee 10, got %d", resp.FeeMsat)
	}
}

func TestSendKeysend(t *testing.T) {
	svc, cleanup := newTestService(t, newMockNode())
	defer cleanup()

	resp, err := svc.SendKeysend(1000, "02f1d2e5f1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d", nil, "")
	if err != nil {
		t.Fatalf("SendKeysend failed: %v", err)
	}
	if resp.FeeMsat != 5 {
		t.Errorf("expected fee 5, got %d", resp.FeeMsat)
	}
}

func TestMakeAndLookupInvoice(t *testing.T) {
	node := newMockNode()
	svc, cleanup := newTestService(t, node)
	defer cleanup()

	tx, err := svc.MakeInvoice(context.Background(), 100000, "test invoice", "", 3600, nil)
	if err != nil {
		t.Fatalf("MakeInvoice failed: %v", err)
	}
	if tx.PaymentHash == "" {
		t.Errorf("expected payment hash")
	}
	if tx.Preimage == "" {
		t.Errorf("expected preimage")
	}
	if tx.Invoice == "" {
		t.Errorf("expected bolt11")
	}

	// lookup before payment: unpaid
	lookedUp, err := svc.LookupInvoice(context.Background(), tx.PaymentHash)
	if err != nil {
		t.Fatalf("LookupInvoice failed: %v", err)
	}
	if lookedUp.AmountMsat != 100000 {
		t.Errorf("expected 100000 msat, got %d", lookedUp.AmountMsat)
	}
	if lookedUp.SettledAt != nil {
		t.Errorf("expected unpaid invoice")
	}

	// simulate node receiving payment, lookup again: settled
	node.markPaid(tx.PaymentHash)
	lookedUp, err = svc.LookupInvoice(context.Background(), tx.PaymentHash)
	if err != nil {
		t.Fatalf("LookupInvoice failed: %v", err)
	}
	if lookedUp.SettledAt == nil {
		t.Errorf("expected settled invoice")
	}
	if lookedUp.Preimage != tx.Preimage {
		t.Errorf("preimage mismatch")
	}
}

func TestMakeInvoiceWithDescriptionHash(t *testing.T) {
	svc, cleanup := newTestService(t, newMockNode())
	defer cleanup()

	// matching hash: ok
	tx, err := svc.MakeInvoice(context.Background(), 1000, "desc", "97864e878fe129a3d4c35681c3ad4b12743f04f7cd705643f2fa1142dfede601", 3600, nil)
	if err != nil {
		t.Fatalf("MakeInvoice with valid hash failed: %v", err)
	}
	if tx.DescriptionHash == "" {
		t.Errorf("expected description hash")
	}

	// mismatched hash: error
	_, err = svc.MakeInvoice(context.Background(), 1000, "desc", "0000000000000000000000000000000000000000000000000000000000000000", 3600, nil)
	if err == nil {
		t.Errorf("expected error for mismatched description hash")
	}
}

func TestGetBalances(t *testing.T) {
	svc, cleanup := newTestService(t, newMockNode())
	defer cleanup()

	balances, err := svc.GetBalances(context.Background(), false)
	if err != nil {
		t.Fatalf("GetBalances failed: %v", err)
	}
	if balances.Onchain.TotalSat != 5000 {
		t.Errorf("expected onchain total 5000, got %d", balances.Onchain.TotalSat)
	}
	if balances.Lightning.TotalSpendableMsat != 145000 {
		t.Errorf("expected spendable 145000, got %d", balances.Lightning.TotalSpendableMsat)
	}
	if balances.Lightning.NextMaxSpendableMsat != 145000 {
		t.Errorf("expected next max spendable 145000, got %d", balances.Lightning.NextMaxSpendableMsat)
	}
	if balances.Lightning.TotalReceivableMsat != 49000000 {
		t.Errorf("expected receivable 49000000, got %d", balances.Lightning.TotalReceivableMsat)
	}
}

func TestListChannels(t *testing.T) {
	svc, cleanup := newTestService(t, newMockNode())
	defer cleanup()

	channels, err := svc.ListChannels(context.Background())
	if err != nil {
		t.Fatalf("ListChannels failed: %v", err)
	}
	if len(channels) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(channels))
	}
	ch := channels[0]
	if !ch.Active {
		t.Errorf("expected active channel")
	}
	if !ch.IsOutbound {
		t.Errorf("expected outbound channel")
	}
	if ch.LocalBalanceMsat != 150000 {
		t.Errorf("expected local balance 150000, got %d", ch.LocalBalanceMsat)
	}
	if ch.RemoteBalanceMsat != 50000 {
		t.Errorf("expected remote balance 50000, got %d", ch.RemoteBalanceMsat)
	}
	if ch.Confirmations == nil || *ch.Confirmations != 40001 {
		t.Errorf("unexpected confirmations: %v", ch.Confirmations)
	}
}

func TestListOnchainTransactions(t *testing.T) {
	svc, cleanup := newTestService(t, newMockNode())
	defer cleanup()

	txs, err := svc.ListOnchainTransactions(context.Background())
	if err != nil {
		t.Fatalf("ListOnchainTransactions failed: %v", err)
	}
	if len(txs) != 1 {
		t.Fatalf("expected 1 tx, got %d", len(txs))
	}
	if txs[0].Type != "incoming" {
		t.Errorf("expected incoming, got %s", txs[0].Type)
	}
	if txs[0].AmountSat != 1000 {
		t.Errorf("expected 1000 sats, got %d", txs[0].AmountSat)
	}
}

func TestRedeemOnchainFunds(t *testing.T) {
	svc, cleanup := newTestService(t, newMockNode())
	defer cleanup()

	txid, err := svc.RedeemOnchainFunds(context.Background(), "bc1qtest", 1000, nil, false)
	if err != nil {
		t.Fatalf("RedeemOnchainFunds failed: %v", err)
	}
	if len(txid) != 64 {
		t.Errorf("expected txid, got %s", txid)
	}

	// sendAll
	_, err = svc.RedeemOnchainFunds(context.Background(), "bc1qtest", 0, nil, true)
	if err != nil {
		t.Fatalf("RedeemOnchainFunds(sendAll) failed: %v", err)
	}
}

func TestSignMessage(t *testing.T) {
	svc, cleanup := newTestService(t, newMockNode())
	defer cleanup()

	sig, err := svc.SignMessage(context.Background(), "hello")
	if err != nil {
		t.Fatalf("SignMessage failed: %v", err)
	}
	if sig == "" {
		t.Errorf("expected signature")
	}
}

func TestMakeOffer(t *testing.T) {
	svc, cleanup := newTestService(t, newMockNode())
	defer cleanup()

	offer, err := svc.MakeOffer(context.Background(), "test offer")
	if err != nil {
		t.Fatalf("MakeOffer failed: %v", err)
	}
	if !strings.HasPrefix(offer, "lno1") {
		t.Errorf("expected bolt12 offer, got %s", offer)
	}
}

func TestHoldInvoicesUnsupported(t *testing.T) {
	svc, cleanup := newTestService(t, newMockNode())
	defer cleanup()

	_, err := svc.MakeHoldInvoice(context.Background(), 1000, "hold", "", 3600, "00", nil)
	if err == nil {
		t.Errorf("expected hold invoice error")
	}
	if err := svc.SettleHoldInvoice(context.Background(), "00"); err == nil {
		t.Errorf("expected settle hold error")
	}
	if err := svc.CancelHoldInvoice(context.Background(), "00"); err == nil {
		t.Errorf("expected cancel hold error")
	}
}

func TestSupportedMethodsAndNotifications(t *testing.T) {
	svc, cleanup := newTestService(t, newMockNode())
	defer cleanup()

	methods := svc.GetSupportedNIP47Methods()
	if len(methods) == 0 {
		t.Errorf("expected supported methods")
	}
	// push tier: pump covers incoming payments with persisted index catch-up
	notifications := svc.GetSupportedNIP47NotificationTypes()
	if len(notifications) != 1 || notifications[0] != "payment_received" {
		t.Errorf("expected payment_received notification type, got %v", notifications)
	}
}

func TestWaitAnyInvoicePump(t *testing.T) {
	node := newMockNode()
	svc, cleanup := newTestService(t, node)
	defer cleanup()

	cap := &capturePublisher{}
	svc.eventPublisher = cap
	svc.pumpRetryDelay = 100 * time.Millisecond
	go svc.waitForInvoices(svc.ctx)

	tx, err := svc.MakeInvoice(context.Background(), 100000, "pump test", "", 3600, nil)
	if err != nil {
		t.Fatalf("MakeInvoice failed: %v", err)
	}

	// no event before payment
	if event := cap.waitForEvent("nwc_lnclient_payment_received", 300*time.Millisecond); event != nil {
		t.Fatalf("unexpected event before payment")
	}

	node.markPaid(tx.PaymentHash)

	event := cap.waitForEvent("nwc_lnclient_payment_received", 3*time.Second)
	if event == nil {
		t.Fatal("expected nwc_lnclient_payment_received event")
	}
	propTx, ok := event.Properties.(*lnclient.Transaction)
	if !ok {
		t.Fatalf("expected *lnclient.Transaction properties, got %T", event.Properties)
	}
	if propTx.PaymentHash != tx.PaymentHash {
		t.Errorf("payment hash mismatch: %s != %s", propTx.PaymentHash, tx.PaymentHash)
	}
	if propTx.SettledAt == nil {
		t.Errorf("expected settled transaction")
	}
	if propTx.AmountMsat != 100000 {
		t.Errorf("expected 100000 msat, got %d", propTx.AmountMsat)
	}

	// state persisted so a restart resumes from this index
	state, err := svc.loadPumpState()
	if err != nil {
		t.Fatalf("failed to load pump state: %v", err)
	}
	if state == 0 {
		t.Errorf("expected persisted pay index > 0")
	}
}

func TestWaitAnyInvoicePumpRestartCatchUp(t *testing.T) {
	node := newMockNode()
	workDir := t.TempDir()

	// "first run": invoice 1 was paid and processed (pay index 1), then the hub went down
	svc1, cleanup1 := newTestServiceWithWorkDir(t, node, workDir)
	tx1, err := svc1.MakeInvoice(context.Background(), 10000, "processed 1", "", 3600, nil)
	if err != nil {
		t.Fatalf("MakeInvoice failed: %v", err)
	}
	node.markPaid(tx1.PaymentHash) // pay index 1
	if err := svc1.savePumpState(1); err != nil {
		t.Fatalf("failed to save pump state: %v", err)
	}
	// while down, two more invoices got paid (indexes 2 and 3)
	tx2, err := svc1.MakeInvoice(context.Background(), 20000, "catchup 2", "", 3600, nil)
	if err != nil {
		t.Fatalf("MakeInvoice failed: %v", err)
	}
	tx3, err := svc1.MakeInvoice(context.Background(), 30000, "catchup 3", "", 3600, nil)
	if err != nil {
		t.Fatalf("MakeInvoice failed: %v", err)
	}
	node.markPaid(tx2.PaymentHash)
	node.markPaid(tx3.PaymentHash)
	cleanup1()

	// "restart": fresh service, same workdir, pump resumes from index 1
	svc2, cleanup2 := newTestServiceWithWorkDir(t, node, workDir)
	defer cleanup2()

	cap := &capturePublisher{}
	svc2.eventPublisher = cap
	svc2.pumpRetryDelay = 100 * time.Millisecond
	go svc2.waitForInvoices(svc2.ctx)

	// both missed payments must be replayed, in order
	event2 := cap.waitForEvent("nwc_lnclient_payment_received", 3*time.Second)
	if event2 == nil {
		t.Fatal("expected catch-up event for first missed payment")
	}
	if propTx, ok := event2.Properties.(*lnclient.Transaction); !ok || propTx.PaymentHash != tx2.PaymentHash {
		t.Fatalf("expected payment hash %s in first catch-up event", tx2.PaymentHash)
	}

	event3 := cap.waitForEvent("nwc_lnclient_payment_received", 3*time.Second)
	if event3 == nil {
		t.Fatal("expected catch-up event for second missed payment")
	}
	if propTx, ok := event3.Properties.(*lnclient.Transaction); !ok || propTx.PaymentHash != tx3.PaymentHash {
		t.Fatalf("expected payment hash %s in second catch-up event", tx3.PaymentHash)
	}

	// index advanced past both
	state, err := svc2.loadPumpState()
	if err != nil {
		t.Fatalf("failed to load pump state: %v", err)
	}
	if state < 3 {
		t.Errorf("expected pump state >= 3, got %d", state)
	}
}

func TestStreamIncomingKeysend(t *testing.T) {
	node := newMockNode()
	svc, cleanup := newTestService(t, node)
	defer cleanup()

	cap := &capturePublisher{}
	svc.eventPublisher = cap
	go svc.streamIncoming(svc.ctx)

	// give the subscription a moment to establish
	time.Sleep(300 * time.Millisecond)

	preimage := make([]byte, 32)
	for i := range preimage {
		preimage[i] = byte(i)
	}
	paymentHash := make([]byte, 32)
	for i := range paymentHash {
		paymentHash[i] = byte(0xff - i)
	}
	appPubkey := []byte("02f1d2e5f1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d")

	// keysend with app-specific TLV record (NIP-47 apps use type 5482373484)
	node.keysendCh <- encodeIncomingPaymentForTest(preimage, paymentHash, 4242, 5482373484, appPubkey, "")

	event := cap.waitForEvent("nwc_lnclient_payment_received", 3*time.Second)
	if event == nil {
		t.Fatal("expected keysend payment event")
	}
	tx, ok := event.Properties.(*lnclient.Transaction)
	if !ok {
		t.Fatalf("expected *lnclient.Transaction properties, got %T", event.Properties)
	}
	if tx.PaymentHash != hex.EncodeToString(paymentHash) {
		t.Errorf("payment hash mismatch: %s", tx.PaymentHash)
	}
	if tx.Preimage != hex.EncodeToString(preimage) {
		t.Errorf("preimage mismatch: %s", tx.Preimage)
	}
	if tx.AmountMsat != 4242 {
		t.Errorf("expected 4242 msat, got %d", tx.AmountMsat)
	}
	if tx.Type != "incoming" {
		t.Errorf("expected incoming transaction, got %s", tx.Type)
	}
	tlvRecords, ok := tx.Metadata["tlv_records"].([]lnclient.TLVRecord)
	if !ok || len(tlvRecords) != 1 {
		t.Fatalf("expected one tlv record, got %#v", tx.Metadata["tlv_records"])
	}
	if tlvRecords[0].Type != 5482373484 {
		t.Errorf("expected tlv type 5482373484, got %d", tlvRecords[0].Type)
	}
	if tlvRecords[0].Value != hex.EncodeToString(appPubkey) {
		t.Errorf("tlv value mismatch: %s", tlvRecords[0].Value)
	}

	// invoice payments (non-empty bolt11) must NOT be published here:
	// they belong to the WaitAnyInvoice pump, publishing both would create
	// duplicate transactions in the hub
	node.keysendCh <- encodeIncomingPaymentForTest(preimage, paymentHash, 999, 0, nil, "lnbc999...")
	if event := cap.waitForEvent("nwc_lnclient_payment_received", 500*time.Millisecond); event != nil {
		t.Fatal("invoice payment must not be published by the keysend stream")
	}
}

func TestGetNodeConnectionInfoAndStatus(t *testing.T) {
	svc, cleanup := newTestService(t, newMockNode())
	defer cleanup()

	connInfo, err := svc.GetNodeConnectionInfo(context.Background())
	if err != nil {
		t.Fatalf("GetNodeConnectionInfo failed: %v", err)
	}
	if len(connInfo.Pubkey) != 66 {
		t.Errorf("unexpected pubkey: %s", connInfo.Pubkey)
	}

	status, err := svc.GetNodeStatus(context.Background())
	if err != nil {
		t.Fatalf("GetNodeStatus failed: %v", err)
	}
	if !status.IsReady {
		t.Errorf("expected node ready")
	}
}

func TestListPeersAndNetworkGraph(t *testing.T) {
	svc, cleanup := newTestService(t, newMockNode())
	defer cleanup()

	peers, err := svc.ListPeers(context.Background())
	if err != nil {
		t.Fatalf("ListPeers failed: %v", err)
	}
	if len(peers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(peers))
	}
	if peers[0].Address != "1.2.3.4" {
		t.Errorf("unexpected peer address: %s", peers[0].Address)
	}

	graph, err := svc.GetNetworkGraph(context.Background(), []string{"02f1d2e5f1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d"})
	if err != nil {
		t.Fatalf("GetNetworkGraph failed: %v", err)
	}
	if graphMap, ok := graph.(map[string]interface{}); !ok || len(graphMap) == 0 {
		t.Errorf("expected network graph")
	}
}

func TestConnectionManagement(t *testing.T) {
	svc, cleanup := newTestService(t, newMockNode())
	defer cleanup()

	if err := svc.ConnectPeer(context.Background(), &lnclient.ConnectPeerRequest{
		Pubkey:  "02f1d2e5f1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d",
		Address: "1.2.3.4",
		Port:    9735,
	}); err != nil {
		t.Fatalf("ConnectPeer failed: %v", err)
	}
	if err := svc.DisconnectPeer(context.Background(), "02f1d2e5f1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d"); err != nil {
		t.Fatalf("DisconnectPeer failed: %v", err)
	}
}

func TestLoadTLSCredentials(t *testing.T) {
	dir := t.TempDir()

	// generate a self-signed CA + client cert/key
	caPEM := genCert(t, dir, "ca", true)
	_ = genCert(t, dir, "client", false)

	serverCA, err := os.ReadFile(filepath.Join(dir, "ca.pem"))
	if err != nil {
		t.Fatalf("failed to read ca.pem: %v", err)
	}
	if !strings.Contains(string(serverCA), "BEGIN CERTIFICATE") {
		t.Errorf("unexpected ca.pem content")
	}
	_ = caPEM

	tlsConfig, err := loadTLSCredentials(dir, "gl1mocknode.gl.blckstrm.com")
	if err != nil {
		t.Fatalf("loadTLSCredentials failed: %v", err)
	}
	if tlsConfig.ServerName != "gl1mocknode.gl.blckstrm.com" {
		t.Errorf("unexpected server name: %s", tlsConfig.ServerName)
	}
	if len(tlsConfig.Certificates) != 1 {
		t.Errorf("expected 1 client certificate, got %d", len(tlsConfig.Certificates))
	}

	// missing files -> error
	if _, err := loadTLSCredentials(t.TempDir(), "cln"); err == nil {
		t.Errorf("expected error for missing creds")
	}
}

func TestHostOf(t *testing.T) {
	cases := map[string]string{
		"gl1abc.gl.blckstrm.com:443": "gl1abc.gl.blckstrm.com",
		"10.0.0.1:21000":             "10.0.0.1",
		"user@host.example.com:9735": "host.example.com",
	}
	for uri, want := range cases {
		if got := hostOf(uri); got != want {
			t.Errorf("hostOf(%s) = %s, want %s", uri, got, want)
		}
	}
}

func TestNormalizeNodeURI(t *testing.T) {
	cases := map[string]string{
		"gl1abc.gl.blckstrm.com:443":         "gl1abc.gl.blckstrm.com:443",
		"https://gl1abc.gl.blckstrm.com:443": "gl1abc.gl.blckstrm.com:443",
		"http://localhost:45813":             "localhost:45813",
		"https://gl1abc.gl.blckstrm.com":     "gl1abc.gl.blckstrm.com",
		"localhost:36511":                    "localhost:36511",
	}
	for uri, want := range cases {
		if got := normalizeNodeURI(uri); got != want {
			t.Errorf("normalizeNodeURI(%s) = %s, want %s", uri, got, want)
		}
	}
}

func genCert(t *testing.T, dir string, name string, isCA bool) []byte {
	t.Helper()
	args := []string{"req", "-x509", "-newkey", "rsa:2048", "-keyout", filepath.Join(dir, name+"-key.pem"),
		"-out", filepath.Join(dir, name+".pem"), "-days", "1", "-nodes", "-subj", "/CN=" + name}
	out, err := exec.Command("openssl", args...).CombinedOutput()
	if err != nil {
		t.Fatalf("openssl failed: %v: %s", err, out)
	}
	pem, err := os.ReadFile(filepath.Join(dir, name+".pem"))
	if err != nil {
		t.Fatalf("failed to read generated cert: %v", err)
	}
	if isCA {
		// reuse the same cert as CA for the test
		if err := os.WriteFile(filepath.Join(dir, "ca.pem"), pem, 0o600); err != nil {
			t.Fatalf("failed to write ca.pem: %v", err)
		}
	} else {
		if err := os.WriteFile(filepath.Join(dir, "client.pem"), pem, 0o600); err != nil {
			t.Fatalf("failed to write client.pem: %v", err)
		}
		key, err := os.ReadFile(filepath.Join(dir, name+"-key.pem"))
		if err != nil {
			t.Fatalf("failed to read client key: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "client-key.pem"), key, 0o600); err != nil {
			t.Fatalf("failed to write client-key.pem: %v", err)
		}
	}
	return pem
}

func TestChannelManagement(t *testing.T) {
	svc, cleanup := newTestService(t, newMockNode())
	defer cleanup()

	resp, err := svc.OpenChannel(context.Background(), &lnclient.OpenChannelRequest{
		Pubkey:     "02f1d2e5f1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d",
		AmountSats: 100000,
		Public:     true,
	})
	if err != nil {
		t.Fatalf("OpenChannel failed: %v", err)
	}
	if len(resp.FundingTxId) != 64 {
		t.Errorf("expected funding txid, got %s", resp.FundingTxId)
	}

	if err := svc.CloseChannel(context.Background(), &lnclient.CloseChannelRequest{
		ChannelId: "ab01cd02ef030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e",
		Force:     false,
	}); err != nil {
		t.Fatalf("CloseChannel failed: %v", err)
	}

	if err := svc.UpdateChannel(context.Background(), &lnclient.UpdateChannelRequest{
		ChannelId:                           "ab01cd02ef030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e",
		ForwardingFeeBaseMsat:               1000,
		ForwardingFeeProportionalMillionths: 1,
	}); err != nil {
		t.Fatalf("UpdateChannel failed: %v", err)
	}
}
