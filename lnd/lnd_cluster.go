package lnd

import (
	"context"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/ziflex/lecho/v3"
	"google.golang.org/grpc"
)

type LNDCluster struct {
	Nodes               []LightningClientWrapper
	ActiveNode          LightningClientWrapper
	ActiveChannelRatio  float64
	Logger              *lecho.Logger
	LivenessCheckPeriod int
}

func (cluster *LNDCluster) StartLivenessLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(cluster.LivenessCheckPeriod) * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cluster.Logger.Info("Checking cluster status")
			cluster.checkClusterStatus(ctx)
		}
	}
}
func (cluster *LNDCluster) checkClusterStatus(ctx context.Context) {
	//for all nodes
	for _, node := range cluster.Nodes {
		//call getinfo
		resp, err := node.GetInfo(ctx, &lnrpc.GetInfoRequest{})
		//if we get an error here, the node is probably offline
		//so we move to the next node
		if err != nil {
			msg := fmt.Sprintf("Error connecting to node, node id %s, error %s", node.GetMainPubkey(), err.Error())
			cluster.Logger.Infof(msg)
			sentry.CaptureMessage(msg)
			continue
		}
		//if the context has been canceled, return
		if ctx.Err() == context.Canceled {
			return
		}
		//if num_active_channels / num_total_channels < x % (50?)
		//not booted yet, go to next
		nrActiveChannels := resp.NumActiveChannels
		totalChannels := resp.NumActiveChannels + resp.NumInactiveChannels
		activeChannelRatio := float64(nrActiveChannels) / float64(totalChannels)
		if activeChannelRatio < cluster.ActiveChannelRatio {
			msg := fmt.Sprintf("Node does not have enough active channels yet, node id %s, ratio %f, active channels %d, total channels %d", resp.IdentityPubkey, activeChannelRatio, nrActiveChannels, totalChannels)
			cluster.Logger.Infof(msg)
			sentry.CaptureMessage(msg)
			continue
		}
		//node is online and has enough active channels, set this node to active
		//log & send notification to Sentry in case we're switching
		if cluster.ActiveNode != node {
			cluster.ActiveNode = node
			message := fmt.Sprintf("Switched nodes: new node id %s", node.GetMainPubkey())
			cluster.Logger.Info(message)
			sentry.CaptureMessage(message)
		}
		//if we get here, break because we have an active node
		//either the one which was already active
		//or the new one
		break
	}
}
func (cluster *LNDCluster) ListChannels(ctx context.Context, req *lnrpc.ListChannelsRequest, options ...grpc.CallOption) (*lnrpc.ListChannelsResponse, error) {
	return cluster.ActiveNode.ListChannels(ctx, req, options...)
}

func (cluster *LNDCluster) SendPaymentSync(ctx context.Context, req *lnrpc.SendRequest, options ...grpc.CallOption) (*lnrpc.SendResponse, error) {
	return cluster.ActiveNode.SendPaymentSync(ctx, req, options...)
}

func (cluster *LNDCluster) WalletBalance(ctx context.Context, req *lnrpc.WalletBalanceRequest, options ...grpc.CallOption) (*lnrpc.WalletBalanceResponse, error) {
	return cluster.ActiveNode.WalletBalance(ctx, req, options...)
}

func (cluster *LNDCluster) AddInvoice(ctx context.Context, req *lnrpc.Invoice, options ...grpc.CallOption) (*lnrpc.AddInvoiceResponse, error) {
	return cluster.ActiveNode.AddInvoice(ctx, req, options...)
}

func (cluster *LNDCluster) SubscribeInvoices(ctx context.Context, req *lnrpc.InvoiceSubscription, options ...grpc.CallOption) (SubscribeInvoicesWrapper, error) {
	return nil, fmt.Errorf("not implemented")
}

func (cluster *LNDCluster) SubscribePayment(ctx context.Context, req *routerrpc.TrackPaymentRequest, options ...grpc.CallOption) (SubscribePaymentWrapper, error) {
	return nil, fmt.Errorf("not implemented")
}

func (cluster *LNDCluster) GetInfo(ctx context.Context, req *lnrpc.GetInfoRequest, options ...grpc.CallOption) (*lnrpc.GetInfoResponse, error) {
	return cluster.ActiveNode.GetInfo(ctx, req, options...)
}

func (cluster *LNDCluster) DecodeBolt11(ctx context.Context, bolt11 string, options ...grpc.CallOption) (*lnrpc.PayReq, error) {
	return cluster.ActiveNode.DecodeBolt11(ctx, bolt11, options...)
}

func (cluster *LNDCluster) IsIdentityPubkey(pubkey string) (isOurPubkey bool) {
	for _, node := range cluster.Nodes {
		if node.GetMainPubkey() == pubkey {
			return true
		}
	}
	return false
}

func (cluster *LNDCluster) GetMainPubkey() (pubkey string) {
	//the first node should always be our primary node
	//which we will use for our main pubkey
	return cluster.Nodes[0].GetMainPubkey()
}
