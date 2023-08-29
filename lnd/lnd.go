package lnd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"io/ioutil"

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

func NewLNDclient(lndOptions LNDoptions, ctx context.Context) (result *LNDWrapper, err error) {
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
	} else if lndOptions.CertFile != "" {
		credsFromFile, err := credentials.NewClientTLSFromFile(lndOptions.CertFile, "")
		if err != nil {
			return nil, err
		}
		creds = credsFromFile // make it available outside of the else if block
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
	} else if lndOptions.MacaroonFile != "" {
		macBytes, err := ioutil.ReadFile(lndOptions.MacaroonFile)
		if err != nil {
			return nil, err
		}
		macaroonData = macBytes // make it available outside of the else if block
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
