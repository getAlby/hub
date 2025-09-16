package wrapper

import (
	"context"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/invoicesrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"google.golang.org/grpc"
)

type LightningClientWrapper interface {
	ListChannels(ctx context.Context, req *lnrpc.ListChannelsRequest, options ...grpc.CallOption) (*lnrpc.ListChannelsResponse, error)
	SendPaymentSync(req *lnrpc.SendRequest, options ...grpc.CallOption) (*lnrpc.SendResponse, error)
	ChannelBalance(ctx context.Context, req *lnrpc.ChannelBalanceRequest, options ...grpc.CallOption) (*lnrpc.ChannelBalanceResponse, error)
	AddInvoice(ctx context.Context, req *lnrpc.Invoice, options ...grpc.CallOption) (*lnrpc.AddInvoiceResponse, error)
	AddHoldInvoice(ctx context.Context, req *invoicesrpc.AddHoldInvoiceRequest, options ...grpc.CallOption) (*invoicesrpc.AddHoldInvoiceResp, error)
	SettleInvoice(ctx context.Context, req *invoicesrpc.SettleInvoiceMsg, options ...grpc.CallOption) (*invoicesrpc.SettleInvoiceResp, error)
	CancelInvoice(ctx context.Context, req *invoicesrpc.CancelInvoiceMsg, options ...grpc.CallOption) (*invoicesrpc.CancelInvoiceResp, error)
	SubscribeInvoices(ctx context.Context, req *lnrpc.InvoiceSubscription, options ...grpc.CallOption) (SubscribeInvoicesWrapper, error)
	SubscribeSingleInvoice(ctx context.Context, req *invoicesrpc.SubscribeSingleInvoiceRequest, options ...grpc.CallOption) (SubscribeSingleInvoiceWrapper, error) // Added
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

type SubscribeSingleInvoiceWrapper interface {
	Recv() (*lnrpc.Invoice, error)
}

type SubscribePaymentWrapper interface {
	Recv() (*lnrpc.Payment, error)
}
