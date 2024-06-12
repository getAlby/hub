package wrapper

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lightningnetwork/lnd/macaroons"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/macaroon.v2"
)

type LNPayReq struct {
	PayReq  *lnrpc.PayReq
	Keysend bool
}

// LNDoptions are the options for the connection to the lnd node.
type LNDoptions struct {
	Address      string
	CertFile     string
	CertHex      string
	MacaroonFile string
	MacaroonHex  string
}

type LNDWrapper struct {
	client         lnrpc.LightningClient
	routerClient   routerrpc.RouterClient
	IdentityPubkey string
}

func NewLNDclient(lndOptions LNDoptions) (result *LNDWrapper, err error) {
	// Get credentials either from a hex string, a file or the system's certificate store
	var creds credentials.TransportCredentials
	// if a hex string is provided
	if lndOptions.CertHex != "" {
		cp := x509.NewCertPool()
		cert, err := hex.DecodeString(lndOptions.CertHex)
		if err != nil {
			return nil, err
		}
		cp.AppendCertsFromPEM(cert)
		creds = credentials.NewClientTLSFromCert(cp, "")
		// if a path to a cert file is provided
	} else {
		creds = credentials.NewTLS(&tls.Config{})
	}
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}

	var macaroonData []byte
	if lndOptions.MacaroonHex != "" {
		macBytes, err := hex.DecodeString(lndOptions.MacaroonHex)
		if err != nil {
			return nil, err
		}
		macaroonData = macBytes
	} else {
		return nil, errors.New("LND macaroon is missing")
	}

	mac := &macaroon.Macaroon{}
	if err := mac.UnmarshalBinary(macaroonData); err != nil {
		return nil, err
	}
	macCred, err := macaroons.NewMacaroonCredential(mac)
	if err != nil {
		return nil, err
	}
	opts = append(opts, grpc.WithPerRPCCredentials(macCred))

	conn, err := grpc.Dial(lndOptions.Address, opts...)
	if err != nil {
		return nil, err
	}
	lnClient := lnrpc.NewLightningClient(conn)
	return &LNDWrapper{
		client:       lnClient,
		routerClient: routerrpc.NewRouterClient(conn),
	}, nil
}

func (wrapper *LNDWrapper) ListChannels(ctx context.Context, req *lnrpc.ListChannelsRequest, options ...grpc.CallOption) (*lnrpc.ListChannelsResponse, error) {
	return wrapper.client.ListChannels(ctx, req, options...)
}

func (wrapper *LNDWrapper) GetTransactions(ctx context.Context, req *lnrpc.GetTransactionsRequest, options ...grpc.CallOption) (*lnrpc.TransactionDetails, error) {
	return wrapper.client.GetTransactions(ctx, req, options...)
}

func (wrapper *LNDWrapper) PendingChannels(ctx context.Context, req *lnrpc.PendingChannelsRequest, options ...grpc.CallOption) (*lnrpc.PendingChannelsResponse, error) {
	return wrapper.client.PendingChannels(ctx, req, options...)
}

func (wrapper *LNDWrapper) SendPaymentSync(ctx context.Context, req *lnrpc.SendRequest, options ...grpc.CallOption) (*lnrpc.SendResponse, error) {
	return wrapper.client.SendPaymentSync(ctx, req, options...)
}

func (wrapper *LNDWrapper) ChannelBalance(ctx context.Context, req *lnrpc.ChannelBalanceRequest, options ...grpc.CallOption) (*lnrpc.ChannelBalanceResponse, error) {
	return wrapper.client.ChannelBalance(ctx, req, options...)
}

func (wrapper *LNDWrapper) AddInvoice(ctx context.Context, req *lnrpc.Invoice, options ...grpc.CallOption) (*lnrpc.AddInvoiceResponse, error) {
	return wrapper.client.AddInvoice(ctx, req, options...)
}

func (wrapper *LNDWrapper) SubscribeInvoices(ctx context.Context, req *lnrpc.InvoiceSubscription, options ...grpc.CallOption) (SubscribeInvoicesWrapper, error) {
	return wrapper.client.SubscribeInvoices(ctx, req, options...)
}

func (wrapper *LNDWrapper) ListInvoices(ctx context.Context, req *lnrpc.ListInvoiceRequest, options ...grpc.CallOption) (*lnrpc.ListInvoiceResponse, error) {
	return wrapper.client.ListInvoices(ctx, req, options...)
}

func (wrapper *LNDWrapper) ListPayments(ctx context.Context, req *lnrpc.ListPaymentsRequest, options ...grpc.CallOption) (*lnrpc.ListPaymentsResponse, error) {
	return wrapper.client.ListPayments(ctx, req, options...)
}

func (wrapper *LNDWrapper) LookupInvoice(ctx context.Context, req *lnrpc.PaymentHash, options ...grpc.CallOption) (*lnrpc.Invoice, error) {
	return wrapper.client.LookupInvoice(ctx, req, options...)
}

func (wrapper *LNDWrapper) GetInfo(ctx context.Context, req *lnrpc.GetInfoRequest, options ...grpc.CallOption) (*lnrpc.GetInfoResponse, error) {
	return wrapper.client.GetInfo(ctx, req, options...)
}

func (wrapper *LNDWrapper) DecodeBolt11(ctx context.Context, bolt11 string, options ...grpc.CallOption) (*lnrpc.PayReq, error) {
	return wrapper.client.DecodePayReq(ctx, &lnrpc.PayReqString{
		PayReq: bolt11,
	})
}

func (wrapper *LNDWrapper) SubscribePayment(ctx context.Context, req *routerrpc.TrackPaymentRequest, options ...grpc.CallOption) (SubscribePaymentWrapper, error) {
	return wrapper.routerClient.TrackPaymentV2(ctx, req, options...)
}

func (wrapper *LNDWrapper) IsIdentityPubkey(pubkey string) (isOurPubkey bool) {
	return pubkey == wrapper.IdentityPubkey
}

func (wrapper *LNDWrapper) GetMainPubkey() (pubkey string) {
	return wrapper.IdentityPubkey
}

func (wrapper *LNDWrapper) SignMessage(ctx context.Context, req *lnrpc.SignMessageRequest, options ...grpc.CallOption) (*lnrpc.SignMessageResponse, error) {
	return wrapper.client.SignMessage(ctx, req, options...)
}

func (wrapper *LNDWrapper) ConnectPeer(ctx context.Context, req *lnrpc.ConnectPeerRequest, options ...grpc.CallOption) (*lnrpc.ConnectPeerResponse, error) {
	return wrapper.client.ConnectPeer(ctx, req, options...)
}

func (wrapper *LNDWrapper) ListPeers(ctx context.Context, req *lnrpc.ListPeersRequest, options ...grpc.CallOption) (*lnrpc.ListPeersResponse, error) {
	return wrapper.client.ListPeers(ctx, req, options...)
}

func (wrapper *LNDWrapper) OpenChannelSync(ctx context.Context, req *lnrpc.OpenChannelRequest, options ...grpc.CallOption) (*lnrpc.ChannelPoint, error) {
	return wrapper.client.OpenChannelSync(ctx, req, options...)
}

func (wrapper *LNDWrapper) CloseChannel(ctx context.Context, req *lnrpc.CloseChannelRequest, options ...grpc.CallOption) (lnrpc.Lightning_CloseChannelClient, error) {
	return wrapper.client.CloseChannel(ctx, req, options...)
}

func (wrapper *LNDWrapper) WalletBalance(ctx context.Context, req *lnrpc.WalletBalanceRequest, options ...grpc.CallOption) (*lnrpc.WalletBalanceResponse, error) {
	return wrapper.client.WalletBalance(ctx, req, options...)
}

func (wrapper *LNDWrapper) NewAddress(ctx context.Context, req *lnrpc.NewAddressRequest, options ...grpc.CallOption) (*lnrpc.NewAddressResponse, error) {
	return wrapper.client.NewAddress(ctx, req, options...)
}
