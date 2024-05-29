package wrapper

import (
	"context"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"google.golang.org/grpc"
)

type LightningClientWrapper interface {
	ListChannels(ctx context.Context, req *lnrpc.ListChannelsRequest, options ...grpc.CallOption) (*lnrpc.ListChannelsResponse, error)
	SendPaymentSync(ctx context.Context, req *lnrpc.SendRequest, options ...grpc.CallOption) (*lnrpc.SendResponse, error)
	ChannelBalance(ctx context.Context, req *lnrpc.ChannelBalanceRequest, options ...grpc.CallOption) (*lnrpc.ChannelBalanceResponse, error)
	AddInvoice(ctx context.Context, req *lnrpc.Invoice, options ...grpc.CallOption) (*lnrpc.AddInvoiceResponse, error)
	SubscribeInvoices(ctx context.Context, req *lnrpc.InvoiceSubscription, options ...grpc.CallOption) (SubscribeInvoicesWrapper, error)
	SubscribePayment(ctx context.Context, req *routerrpc.TrackPaymentRequest, options ...grpc.CallOption) (SubscribePaymentWrapper, error)
	LookupInvoice(ctx context.Context, req *lnrpc.PaymentHash, options ...grpc.CallOption) (*lnrpc.Invoice, error)
	GetInfo(ctx context.Context, req *lnrpc.GetInfoRequest, options ...grpc.CallOption) (*lnrpc.GetInfoResponse, error)
	DecodeBolt11(ctx context.Context, bolt11 string, options ...grpc.CallOption) (*lnrpc.PayReq, error)
	IsIdentityPubkey(pubkey string) (isOurPubkey bool)
	GetMainPubkey() (pubkey string)
	SignMessage(ctx context.Context, req *lnrpc.SignMessageRequest, options ...grpc.CallOption) (*lnrpc.SignMessageResponse, error)
}

type SubscribeInvoicesWrapper interface {
	Recv() (*lnrpc.Invoice, error)
}
type SubscribePaymentWrapper interface {
	Recv() (*lnrpc.Payment, error)
}
