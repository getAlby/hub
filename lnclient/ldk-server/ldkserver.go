package ldkserver

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	ldkapi "github.com/getAlby/hub/lnclient/ldk-server/grpc/api"
	ldkevents "github.com/getAlby/hub/lnclient/ldk-server/grpc/events"
	ldktypes "github.com/getAlby/hub/lnclient/ldk-server/grpc/types"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/nip47/notifications"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"google.golang.org/protobuf/proto"
)

type LDKServerService struct {
	ctx            context.Context
	cancel         context.CancelFunc
	client         *http.Client
	baseURL        string
	address        string
	apiKey         string
	eventPublisher events.EventPublisher
	pubkey         string
	nodeInfo       *lnclient.NodeInfo
}

func NewLDKServerService(ctx context.Context, eventPublisher events.EventPublisher, address, tlsCertPEM, apiKey string) (lnclient.LNClient, error) {
	if address == "" || tlsCertPEM == "" || apiKey == "" {
		return nil, errors.New("one or more required ldk-server configuration values are missing")
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM([]byte(tlsCertPEM)) {
		return nil, errors.New("failed to parse ldk-server TLS certificate")
	}

	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("invalid ldk-server address %q: %w", address, err)
	}

	serviceCtx, cancel := context.WithCancel(ctx)
	svc := &LDKServerService{
		ctx:            serviceCtx,
		cancel:         cancel,
		baseURL:        "https://" + address,
		address:        address,
		apiKey:         apiKey,
		eventPublisher: eventPublisher,
		client: &http.Client{
			Transport: &http2.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
					RootCAs:    certPool,
					ServerName: host,
				},
			},
		},
	}

	info, err := svc.GetInfo(serviceCtx)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to ldk-server: %w", err)
	}
	svc.nodeInfo = info
	svc.pubkey = info.Pubkey

	go svc.subscribeEvents()

	logger.Logger.WithFields(logrus.Fields{
		"address": address,
		"alias":   info.Alias,
		"pubkey":  info.Pubkey,
	}).Info("Connected to ldk-server via gRPC API")

	return svc, nil
}

func (svc *LDKServerService) SendPaymentSync(payReq string, amountMsat *uint64) (*lnclient.PayInvoiceResponse, error) {
	req := &ldkapi.Bolt11SendRequest{
		Invoice: payReq,
	}
	if amountMsat != nil {
		req.AmountMsat = amountMsat
	}
	resp := &ldkapi.Bolt11SendResponse{}
	if err := svc.doUnaryWithTimeout(svc.ctx, 2*time.Minute, ldkapi.LightningNode_Bolt11Send_FullMethodName, req, resp); err != nil {
		return nil, err
	}

	payment, err := svc.waitForPaymentTerminal(resp.PaymentId)
	if err != nil {
		return nil, err
	}

	switch payment.Status {
	case ldktypes.PaymentStatus_SUCCEEDED:
		response := &lnclient.PayInvoiceResponse{}
		if fee := payment.FeePaidMsat; fee != nil {
			response.FeeMsat = *fee
		}
		if kind, ok := payment.Kind.Kind.(*ldktypes.PaymentKind_Bolt11); ok && kind.Bolt11.Preimage != nil {
			response.Preimage = *kind.Bolt11.Preimage
		}
		return response, nil
	case ldktypes.PaymentStatus_FAILED:
		return nil, errors.New("ldk-server reported payment failure")
	default:
		return nil, fmt.Errorf("unexpected payment status: %s", payment.Status.String())
	}
}

func (svc *LDKServerService) SendKeysend(amountMsat uint64, destination string, customRecords []lnclient.TLVRecord, preimage string) (*lnclient.PayKeysendResponse, error) {
	if preimage != "" {
		return nil, errors.New("ldk-server does not support custom keysend preimages")
	}

	req := &ldkapi.SpontaneousSendRequest{
		AmountMsat: amountMsat,
		NodeId:     destination,
	}
	for _, record := range customRecords {
		value, err := hex.DecodeString(record.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to decode keysend TLV value: %w", err)
		}
		req.CustomTlvs = append(req.CustomTlvs, &ldktypes.CustomTlvRecord{
			TypeNum: record.Type,
			Value:   value,
		})
	}

	resp := &ldkapi.SpontaneousSendResponse{}
	if err := svc.doUnaryWithTimeout(svc.ctx, 2*time.Minute, ldkapi.LightningNode_SpontaneousSend_FullMethodName, req, resp); err != nil {
		return nil, err
	}

	payment, err := svc.waitForPaymentTerminal(resp.PaymentId)
	if err != nil {
		return nil, err
	}
	if payment.Status != ldktypes.PaymentStatus_SUCCEEDED {
		return nil, errors.New("ldk-server keysend did not succeed")
	}

	result := &lnclient.PayKeysendResponse{}
	if fee := payment.FeePaidMsat; fee != nil {
		result.FeeMsat = *fee
	}
	return result, nil
}

func (svc *LDKServerService) GetPubkey() string {
	return svc.pubkey
}

func (svc *LDKServerService) GetInfo(ctx context.Context) (*lnclient.NodeInfo, error) {
	resp := &ldkapi.GetNodeInfoResponse{}
	if err := svc.doUnary(ctx, ldkapi.LightningNode_GetNodeInfo_FullMethodName, &ldkapi.GetNodeInfoRequest{}, resp); err != nil {
		return nil, err
	}

	info := &lnclient.NodeInfo{
		Alias:       resp.GetNodeAlias(),
		Pubkey:      resp.NodeId,
		Network:     networkToString(resp.Network),
		BlockHeight: resp.CurrentBestBlock.GetHeight(),
		BlockHash:   resp.CurrentBestBlock.GetBlockHash(),
	}
	svc.nodeInfo = info
	svc.pubkey = info.Pubkey
	return info, nil
}

func (svc *LDKServerService) MakeInvoice(ctx context.Context, amountMsat int64, description string, descriptionHash string, expiry int64, throughNodePubkey *string) (*lnclient.Transaction, error) {
	if expiry == 0 {
		expiry = lnclient.DEFAULT_INVOICE_EXPIRY
	}

	if throughNodePubkey != nil {
		if amountMsat > 0 {
			resp := &ldkapi.Bolt11ReceiveViaJitChannelResponse{}
			if err := svc.doUnary(ctx, ldkapi.LightningNode_Bolt11ReceiveViaJitChannel_FullMethodName, &ldkapi.Bolt11ReceiveViaJitChannelRequest{
				AmountMsat:  uint64(amountMsat),
				Description: newInvoiceDescription(description, descriptionHash),
				ExpirySecs:  uint32(expiry),
			}, resp); err != nil {
				return nil, err
			}
			return svc.transactionFromCreatedInvoice(ctx, resp.Invoice, "")
		}

		resp := &ldkapi.Bolt11ReceiveVariableAmountViaJitChannelResponse{}
		if err := svc.doUnary(ctx, ldkapi.LightningNode_Bolt11ReceiveVariableAmountViaJitChannel_FullMethodName, &ldkapi.Bolt11ReceiveVariableAmountViaJitChannelRequest{
			Description: newInvoiceDescription(description, descriptionHash),
			ExpirySecs:  uint32(expiry),
		}, resp); err != nil {
			return nil, err
		}
		return svc.transactionFromCreatedInvoice(ctx, resp.Invoice, "")
	}

	req := &ldkapi.Bolt11ReceiveRequest{
		Description: newInvoiceDescription(description, descriptionHash),
		ExpirySecs:  uint32(expiry),
	}
	if amountMsat > 0 {
		req.AmountMsat = uint64Ptr(uint64(amountMsat))
	}
	resp := &ldkapi.Bolt11ReceiveResponse{}
	if err := svc.doUnary(ctx, ldkapi.LightningNode_Bolt11Receive_FullMethodName, req, resp); err != nil {
		return nil, err
	}
	return svc.transactionFromCreatedInvoice(ctx, resp.Invoice, resp.PaymentHash)
}

func (svc *LDKServerService) MakeHoldInvoice(ctx context.Context, amountMsat int64, description string, descriptionHash string, expiry int64, paymentHash string, minCltvExpiryDelta *uint64) (*lnclient.Transaction, error) {
	if minCltvExpiryDelta != nil {
		return nil, errors.New("ldk-server does not expose min_cltv_expiry_delta for hold invoices")
	}
	if expiry == 0 {
		expiry = lnclient.DEFAULT_INVOICE_EXPIRY
	}

	req := &ldkapi.Bolt11ReceiveForHashRequest{
		Description: newInvoiceDescription(description, descriptionHash),
		ExpirySecs:  uint32(expiry),
		PaymentHash: paymentHash,
	}
	if amountMsat > 0 {
		req.AmountMsat = uint64Ptr(uint64(amountMsat))
	}
	resp := &ldkapi.Bolt11ReceiveForHashResponse{}
	if err := svc.doUnary(ctx, ldkapi.LightningNode_Bolt11ReceiveForHash_FullMethodName, req, resp); err != nil {
		return nil, err
	}
	transaction, err := svc.transactionFromCreatedInvoice(ctx, resp.Invoice, paymentHash)
	if err != nil {
		return nil, err
	}
	transaction.PaymentHash = paymentHash
	return transaction, nil
}

func (svc *LDKServerService) SettleHoldInvoice(ctx context.Context, preimage string) error {
	if len(preimage) != 64 {
		return errors.New("preimage must be a 32-byte hex string")
	}
	_, err := svc.doUnaryEmpty(ctx, ldkapi.LightningNode_Bolt11ClaimForHash_FullMethodName, &ldkapi.Bolt11ClaimForHashRequest{
		Preimage: preimage,
	})
	return err
}

func (svc *LDKServerService) CancelHoldInvoice(ctx context.Context, paymentHash string) error {
	_, err := svc.doUnaryEmpty(ctx, ldkapi.LightningNode_Bolt11FailForHash_FullMethodName, &ldkapi.Bolt11FailForHashRequest{
		PaymentHash: paymentHash,
	})
	return err
}

func (svc *LDKServerService) LookupInvoice(ctx context.Context, paymentHash string) (*lnclient.Transaction, error) {
	payment, err := svc.findPayment(ctx, func(payment *ldktypes.Payment) bool {
		return paymentHashMatches(payment, paymentHash)
	})
	if err != nil {
		return nil, err
	}
	return paymentToTransaction(payment)
}

func (svc *LDKServerService) ListOnchainTransactions(ctx context.Context) ([]lnclient.OnchainTransaction, error) {
	payments, err := svc.listAllPayments(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]lnclient.OnchainTransaction, 0)
	for _, payment := range payments {
		onchain, ok := payment.Kind.Kind.(*ldktypes.PaymentKind_Onchain)
		if !ok {
			continue
		}
		tx := lnclient.OnchainTransaction{
			CreatedAt: payment.LatestUpdateTimestamp,
			TxId:      onchain.Onchain.Txid,
		}
		if amount := payment.AmountMsat; amount != nil {
			tx.AmountSat = *amount / 1000
		}
		if payment.Direction == ldktypes.PaymentDirection_OUTBOUND {
			tx.Type = "outgoing"
		} else {
			tx.Type = "incoming"
		}
		switch status := onchain.Onchain.Status.Status.(type) {
		case *ldktypes.ConfirmationStatus_Confirmed:
			tx.State = "confirmed"
			tx.NumConfirmations = uint32(max(int64(svc.nodeInfo.BlockHeight)-int64(status.Confirmed.Height)+1, 0))
		default:
			tx.State = "unconfirmed"
		}
		result = append(result, tx)
	}
	return result, nil
}

func (svc *LDKServerService) Shutdown() error {
	svc.cancel()
	return nil
}

func (svc *LDKServerService) ListChannels(ctx context.Context) ([]lnclient.Channel, error) {
	resp := &ldkapi.ListChannelsResponse{}
	if err := svc.doUnary(ctx, ldkapi.LightningNode_ListChannels_FullMethodName, &ldkapi.ListChannelsRequest{}, resp); err != nil {
		return nil, err
	}

	channels := make([]lnclient.Channel, 0, len(resp.Channels))
	for _, channel := range resp.Channels {
		var fundingTxID string
		var fundingTxVout uint32
		if channel.FundingTxo != nil {
			fundingTxID = channel.FundingTxo.Txid
			fundingTxVout = channel.FundingTxo.Vout
		}

		var errText *string
		if channel.IsChannelReady && !channel.IsUsable {
			msg := "channel is not currently usable"
			errText = &msg
		}

		localBalanceMsat := int64(channel.ChannelValueSats*1000) - int64(channel.InboundCapacityMsat) - int64(channel.CounterpartyUnspendablePunishmentReserve*1000)
		channels = append(channels, lnclient.Channel{
			LocalBalanceMsat:                    localBalanceMsat,
			LocalSpendableBalanceMsat:           int64(channel.OutboundCapacityMsat),
			RemoteBalanceMsat:                   int64(channel.InboundCapacityMsat),
			Id:                                  channel.UserChannelId,
			RemotePubkey:                        channel.CounterpartyNodeId,
			FundingTxId:                         fundingTxID,
			FundingTxVout:                       fundingTxVout,
			Active:                              channel.IsUsable,
			Public:                              channel.IsAnnounced,
			InternalChannel:                     channel,
			Confirmations:                       channel.Confirmations,
			ConfirmationsRequired:               channel.ConfirmationsRequired,
			ForwardingFeeBaseMsat:               channel.ChannelConfig.GetForwardingFeeBaseMsat(),
			ForwardingFeeProportionalMillionths: channel.ChannelConfig.GetForwardingFeeProportionalMillionths(),
			UnspendablePunishmentReserveSat:     channel.GetUnspendablePunishmentReserve(),
			CounterpartyUnspendablePunishmentReserveSat: channel.CounterpartyUnspendablePunishmentReserve,
			Error:      errText,
			IsOutbound: channel.IsOutbound,
		})
	}
	return channels, nil
}

func (svc *LDKServerService) GetNodeConnectionInfo(ctx context.Context) (*lnclient.NodeConnectionInfo, error) {
	resp := &ldkapi.GetNodeInfoResponse{}
	if err := svc.doUnary(ctx, ldkapi.LightningNode_GetNodeInfo_FullMethodName, &ldkapi.GetNodeInfoRequest{}, resp); err != nil {
		return nil, err
	}

	for _, uri := range resp.NodeUris {
		pubkey, host, port, err := parseNodeURI(uri)
		if err == nil {
			return &lnclient.NodeConnectionInfo{
				Pubkey:  pubkey,
				Address: host,
				Port:    port,
			}, nil
		}
	}

	for _, addr := range resp.ListeningAddresses {
		host, port, err := splitHostPort(addr)
		if err == nil {
			return &lnclient.NodeConnectionInfo{
				Pubkey:  resp.NodeId,
				Address: host,
				Port:    port,
			}, nil
		}
	}

	return nil, errors.New("ldk-server did not expose a usable node connection address")
}

func (svc *LDKServerService) GetNodeStatus(ctx context.Context) (*lnclient.NodeStatus, error) {
	resp := &ldkapi.GetNodeInfoResponse{}
	if err := svc.doUnary(ctx, ldkapi.LightningNode_GetNodeInfo_FullMethodName, &ldkapi.GetNodeInfoRequest{}, resp); err != nil {
		return nil, err
	}
	return &lnclient.NodeStatus{
		IsReady: true,
		InternalNodeStatus: map[string]interface{}{
			"latest_lightning_wallet_sync_timestamp": resp.LatestLightningWalletSyncTimestamp,
			"latest_onchain_wallet_sync_timestamp":   resp.LatestOnchainWalletSyncTimestamp,
			"latest_fee_rate_cache_update_timestamp": resp.LatestFeeRateCacheUpdateTimestamp,
			"latest_rgs_snapshot_timestamp":          resp.LatestRgsSnapshotTimestamp,
			"latest_node_announcement_timestamp":     resp.LatestNodeAnnouncementBroadcastTimestamp,
		},
	}, nil
}

func (svc *LDKServerService) ConnectPeer(ctx context.Context, connectPeerRequest *lnclient.ConnectPeerRequest) error {
	address := fmt.Sprintf("%s:%d", connectPeerRequest.Address, connectPeerRequest.Port)
	_, err := svc.doUnaryEmpty(ctx, ldkapi.LightningNode_ConnectPeer_FullMethodName, &ldkapi.ConnectPeerRequest{
		NodePubkey: connectPeerRequest.Pubkey,
		Address:    address,
		Persist:    true,
	})
	return err
}

func (svc *LDKServerService) OpenChannel(ctx context.Context, openChannelRequest *lnclient.OpenChannelRequest) (*lnclient.OpenChannelResponse, error) {
	peer, err := svc.findPeer(ctx, openChannelRequest.Pubkey)
	if err != nil {
		return nil, err
	}
	resp := &ldkapi.OpenChannelResponse{}
	if err := svc.doUnary(ctx, ldkapi.LightningNode_OpenChannel_FullMethodName, &ldkapi.OpenChannelRequest{
		NodePubkey:        openChannelRequest.Pubkey,
		Address:           peer.Address,
		ChannelAmountSats: uint64(openChannelRequest.AmountSats),
		AnnounceChannel:   openChannelRequest.Public,
	}, resp); err != nil {
		return nil, err
	}
	fundingTxID, err := svc.waitForFundingTxID(resp.UserChannelId)
	if err != nil {
		return nil, err
	}
	return &lnclient.OpenChannelResponse{
		FundingTxId: fundingTxID,
	}, nil
}

func (svc *LDKServerService) CloseChannel(ctx context.Context, closeChannelRequest *lnclient.CloseChannelRequest) error {
	method := ldkapi.LightningNode_CloseChannel_FullMethodName
	request := proto.Message(&ldkapi.CloseChannelRequest{
		UserChannelId:      closeChannelRequest.ChannelId,
		CounterpartyNodeId: closeChannelRequest.NodeId,
	})
	if closeChannelRequest.Force {
		method = ldkapi.LightningNode_ForceCloseChannel_FullMethodName
		request = &ldkapi.ForceCloseChannelRequest{
			UserChannelId:      closeChannelRequest.ChannelId,
			CounterpartyNodeId: closeChannelRequest.NodeId,
		}
	}
	_, err := svc.doUnaryEmpty(ctx, method, request)
	return err
}

func (svc *LDKServerService) UpdateChannel(ctx context.Context, updateChannelRequest *lnclient.UpdateChannelRequest) error {
	req := &ldkapi.UpdateChannelConfigRequest{
		UserChannelId:      updateChannelRequest.ChannelId,
		CounterpartyNodeId: updateChannelRequest.NodeId,
		ChannelConfig: &ldktypes.ChannelConfig{
			ForwardingFeeBaseMsat:               &updateChannelRequest.ForwardingFeeBaseMsat,
			ForwardingFeeProportionalMillionths: &updateChannelRequest.ForwardingFeeProportionalMillionths,
			MaxDustHtlcExposure: &ldktypes.ChannelConfig_FeeRateMultiplier{
				FeeRateMultiplier: updateChannelRequest.MaxDustHtlcExposureFromFeeRateMultiplier,
			},
		},
	}
	_, err := svc.doUnaryEmpty(ctx, ldkapi.LightningNode_UpdateChannelConfig_FullMethodName, req)
	return err
}

func (svc *LDKServerService) DisconnectPeer(ctx context.Context, peerID string) error {
	_, err := svc.doUnaryEmpty(ctx, ldkapi.LightningNode_DisconnectPeer_FullMethodName, &ldkapi.DisconnectPeerRequest{
		NodePubkey: peerID,
	})
	return err
}

func (svc *LDKServerService) MakeOffer(ctx context.Context, description string) (string, error) {
	resp := &ldkapi.Bolt12ReceiveResponse{}
	if err := svc.doUnary(ctx, ldkapi.LightningNode_Bolt12Receive_FullMethodName, &ldkapi.Bolt12ReceiveRequest{
		Description: description,
	}, resp); err != nil {
		return "", err
	}
	return resp.Offer, nil
}

func (svc *LDKServerService) GetNewOnchainAddress(ctx context.Context) (string, error) {
	resp := &ldkapi.OnchainReceiveResponse{}
	if err := svc.doUnary(ctx, ldkapi.LightningNode_OnchainReceive_FullMethodName, &ldkapi.OnchainReceiveRequest{}, resp); err != nil {
		return "", err
	}
	return resp.Address, nil
}

func (svc *LDKServerService) ResetRouter(key string) error {
	return errors.New("ldk-server does not expose router reset")
}

func (svc *LDKServerService) GetOnchainBalance(ctx context.Context) (*lnclient.OnchainBalanceResponse, error) {
	resp := &ldkapi.GetBalancesResponse{}
	if err := svc.doUnary(ctx, ldkapi.LightningNode_GetBalances_FullMethodName, &ldkapi.GetBalancesRequest{}, resp); err != nil {
		return nil, err
	}

	result := &lnclient.OnchainBalanceResponse{
		SpendableSat:                int64(resp.SpendableOnchainBalanceSats),
		TotalSat:                    int64(resp.TotalOnchainBalanceSats),
		ReservedSat:                 int64(resp.TotalAnchorChannelsReserveSats),
		PendingBalancesDetails:      []lnclient.PendingBalanceDetails{},
		PendingSweepBalancesDetails: []lnclient.PendingBalanceDetails{},
	}
	for _, balance := range resp.LightningBalances {
		switch b := balance.BalanceType.(type) {
		case *ldktypes.LightningBalance_ClaimableOnChannelClose:
			result.PendingBalancesFromChannelClosuresSat += b.ClaimableOnChannelClose.AmountSatoshis
			result.PendingBalancesDetails = append(result.PendingBalancesDetails, lnclient.PendingBalanceDetails{
				ChannelId: b.ClaimableOnChannelClose.ChannelId,
				NodeId:    b.ClaimableOnChannelClose.CounterpartyNodeId,
				AmountSat: b.ClaimableOnChannelClose.AmountSatoshis,
			})
		case *ldktypes.LightningBalance_ClaimableAwaitingConfirmations:
			result.PendingBalancesFromChannelClosuresSat += b.ClaimableAwaitingConfirmations.AmountSatoshis
			result.PendingBalancesDetails = append(result.PendingBalancesDetails, lnclient.PendingBalanceDetails{
				ChannelId: b.ClaimableAwaitingConfirmations.ChannelId,
				NodeId:    b.ClaimableAwaitingConfirmations.CounterpartyNodeId,
				AmountSat: b.ClaimableAwaitingConfirmations.AmountSatoshis,
			})
		case *ldktypes.LightningBalance_ContentiousClaimable:
			result.PendingBalancesFromChannelClosuresSat += b.ContentiousClaimable.AmountSatoshis
			result.PendingBalancesDetails = append(result.PendingBalancesDetails, lnclient.PendingBalanceDetails{
				ChannelId: b.ContentiousClaimable.ChannelId,
				NodeId:    b.ContentiousClaimable.CounterpartyNodeId,
				AmountSat: b.ContentiousClaimable.AmountSatoshis,
			})
		}
	}
	for _, balance := range resp.PendingBalancesFromChannelClosures {
		switch b := balance.BalanceType.(type) {
		case *ldktypes.PendingSweepBalance_PendingBroadcast:
			result.PendingSweepBalancesDetails = append(result.PendingSweepBalancesDetails, lnclient.PendingBalanceDetails{
				ChannelId: b.PendingBroadcast.GetChannelId(),
				AmountSat: b.PendingBroadcast.AmountSatoshis,
			})
		case *ldktypes.PendingSweepBalance_BroadcastAwaitingConfirmation:
			result.PendingSweepBalancesDetails = append(result.PendingSweepBalancesDetails, lnclient.PendingBalanceDetails{
				ChannelId: b.BroadcastAwaitingConfirmation.GetChannelId(),
				AmountSat: b.BroadcastAwaitingConfirmation.AmountSatoshis,
			})
		case *ldktypes.PendingSweepBalance_AwaitingThresholdConfirmations:
			result.PendingSweepBalancesDetails = append(result.PendingSweepBalancesDetails, lnclient.PendingBalanceDetails{
				ChannelId: b.AwaitingThresholdConfirmations.GetChannelId(),
				AmountSat: b.AwaitingThresholdConfirmations.AmountSatoshis,
			})
		}
	}
	return result, nil
}

func (svc *LDKServerService) GetBalances(ctx context.Context, includeInactiveChannels bool) (*lnclient.BalancesResponse, error) {
	onchain, err := svc.GetOnchainBalance(ctx)
	if err != nil {
		return nil, err
	}
	channels, err := svc.ListChannels(ctx)
	if err != nil {
		return nil, err
	}

	lightning := lnclient.LightningBalanceResponse{}
	for _, channel := range channels {
		if !includeInactiveChannels && !channel.Active {
			continue
		}
		lightning.TotalSpendableMsat += channel.LocalSpendableBalanceMsat
		lightning.TotalReceivableMsat += channel.RemoteBalanceMsat
		lightning.NextMaxSpendableMsat = max(lightning.NextMaxSpendableMsat, channel.LocalSpendableBalanceMsat)
		lightning.NextMaxReceivableMsat = max(lightning.NextMaxReceivableMsat, channel.RemoteBalanceMsat)
	}
	lightning.NextMaxSpendableMPPMsat = lightning.TotalSpendableMsat
	lightning.NextMaxReceivableMPPMsat = lightning.TotalReceivableMsat

	return &lnclient.BalancesResponse{
		Onchain:   *onchain,
		Lightning: lightning,
	}, nil
}

func (svc *LDKServerService) RedeemOnchainFunds(ctx context.Context, toAddress string, amountSat uint64, feeRate *uint64, sendAll bool) (string, error) {
	req := &ldkapi.OnchainSendRequest{
		Address: toAddress,
	}
	if sendAll {
		req.SendAll = boolPtr(true)
	} else {
		req.AmountSats = uint64Ptr(amountSat)
	}
	if feeRate != nil {
		req.FeeRateSatPerVb = feeRate
	}
	resp := &ldkapi.OnchainSendResponse{}
	if err := svc.doUnary(ctx, ldkapi.LightningNode_OnchainSend_FullMethodName, req, resp); err != nil {
		return "", err
	}
	return resp.Txid, nil
}

func (svc *LDKServerService) ListPeers(ctx context.Context) ([]lnclient.PeerDetails, error) {
	resp := &ldkapi.ListPeersResponse{}
	if err := svc.doUnary(ctx, ldkapi.LightningNode_ListPeers_FullMethodName, &ldkapi.ListPeersRequest{}, resp); err != nil {
		return nil, err
	}
	peers := make([]lnclient.PeerDetails, 0, len(resp.Peers))
	for _, peer := range resp.Peers {
		peers = append(peers, lnclient.PeerDetails{
			NodeId:      peer.NodeId,
			Address:     peer.Address,
			IsPersisted: peer.IsPersisted,
			IsConnected: peer.IsConnected,
		})
	}
	return peers, nil
}

func (svc *LDKServerService) GetLogOutput(ctx context.Context, maxLen int) ([]byte, error) {
	return nil, errors.New("ldk-server does not expose remote log output")
}

func (svc *LDKServerService) SignMessage(ctx context.Context, message string) (string, error) {
	resp := &ldkapi.SignMessageResponse{}
	if err := svc.doUnary(ctx, ldkapi.LightningNode_SignMessage_FullMethodName, &ldkapi.SignMessageRequest{
		Message: []byte(message),
	}, resp); err != nil {
		return "", err
	}
	return resp.Signature, nil
}

func (svc *LDKServerService) GetStorageDir() (string, error) {
	return "", errors.New("ldk-server storage is managed remotely")
}

func (svc *LDKServerService) GetNetworkGraph(ctx context.Context, nodeIDs []string) (lnclient.NetworkGraphResponse, error) {
	listNodesResp := &ldkapi.GraphListNodesResponse{}
	if err := svc.doUnary(ctx, ldkapi.LightningNode_GraphListNodes_FullMethodName, &ldkapi.GraphListNodesRequest{}, listNodesResp); err != nil {
		return nil, err
	}
	listChannelsResp := &ldkapi.GraphListChannelsResponse{}
	if err := svc.doUnary(ctx, ldkapi.LightningNode_GraphListChannels_FullMethodName, &ldkapi.GraphListChannelsRequest{}, listChannelsResp); err != nil {
		return nil, err
	}

	filteredNodeIDs := listNodesResp.NodeIds
	if len(nodeIDs) > 0 {
		filteredNodeIDs = filteredNodeIDs[:0]
		for _, nodeID := range listNodesResp.NodeIds {
			if slices.Contains(nodeIDs, nodeID) {
				filteredNodeIDs = append(filteredNodeIDs, nodeID)
			}
		}
	}

	nodes := make([]*ldkapi.GraphGetNodeResponse, 0, len(filteredNodeIDs))
	for _, nodeID := range filteredNodeIDs {
		nodeResp := &ldkapi.GraphGetNodeResponse{}
		if err := svc.doUnary(ctx, ldkapi.LightningNode_GraphGetNode_FullMethodName, &ldkapi.GraphGetNodeRequest{
			NodeId: nodeID,
		}, nodeResp); err == nil {
			nodes = append(nodes, nodeResp)
		}
	}

	channels := make([]*ldkapi.GraphGetChannelResponse, 0, len(listChannelsResp.ShortChannelIds))
	for _, shortID := range listChannelsResp.ShortChannelIds {
		channelResp := &ldkapi.GraphGetChannelResponse{}
		if err := svc.doUnary(ctx, ldkapi.LightningNode_GraphGetChannel_FullMethodName, &ldkapi.GraphGetChannelRequest{
			ShortChannelId: shortID,
		}, channelResp); err == nil {
			channels = append(channels, channelResp)
		}
	}

	return map[string]interface{}{
		"nodes":    nodes,
		"channels": channels,
	}, nil
}

func (svc *LDKServerService) UpdateLastWalletSyncRequest() {}

func (svc *LDKServerService) GetSupportedNIP47Methods() []string {
	return []string{
		models.PAY_INVOICE_METHOD,
		models.PAY_KEYSEND_METHOD,
		models.GET_BALANCE_METHOD,
		models.GET_BUDGET_METHOD,
		models.GET_INFO_METHOD,
		models.MAKE_INVOICE_METHOD,
		models.LOOKUP_INVOICE_METHOD,
		models.LIST_TRANSACTIONS_METHOD,
		models.MULTI_PAY_INVOICE_METHOD,
		models.MULTI_PAY_KEYSEND_METHOD,
		models.SIGN_MESSAGE_METHOD,
		models.MAKE_HOLD_INVOICE_METHOD,
		models.SETTLE_HOLD_INVOICE_METHOD,
		models.CANCEL_HOLD_INVOICE_METHOD,
	}
}

func (svc *LDKServerService) GetSupportedNIP47NotificationTypes() []string {
	return []string{
		notifications.PAYMENT_RECEIVED_NOTIFICATION,
		notifications.PAYMENT_SENT_NOTIFICATION,
		notifications.HOLD_INVOICE_ACCEPTED_NOTIFICATION,
	}
}

func (svc *LDKServerService) GetCustomNodeCommandDefinitions() []lnclient.CustomNodeCommandDef {
	return nil
}

func (svc *LDKServerService) ExecuteCustomNodeCommand(ctx context.Context, command *lnclient.CustomNodeCommandRequest) (*lnclient.CustomNodeCommandResponse, error) {
	return nil, lnclient.ErrUnknownCustomNodeCommand
}

func (svc *LDKServerService) doUnaryEmpty(ctx context.Context, path string, req proto.Message) (*struct{}, error) {
	resp := &struct{}{}
	return resp, svc.doUnary(ctx, path, req, nil)
}

func (svc *LDKServerService) doUnaryWithTimeout(ctx context.Context, timeout time.Duration, path string, req proto.Message, resp proto.Message) error {
	requestCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return svc.doUnary(requestCtx, path, req, resp)
}

func (svc *LDKServerService) doUnary(ctx context.Context, path string, req proto.Message, resp proto.Message) error {
	reqBytes, err := proto.Marshal(req)
	if err != nil {
		return err
	}
	framedReq := grpcFrame(reqBytes)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, svc.baseURL+path, bytes.NewReader(framedReq))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/grpc+proto")
	httpReq.Header.Set("TE", "trailers")
	httpReq.Header.Set("X-Auth", svc.authHeader(framedReq))

	httpResp, err := svc.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	body, readErr := io.ReadAll(httpResp.Body)
	if readErr != nil {
		return readErr
	}
	if err := grpcStatusError(httpResp, body); err != nil {
		return err
	}
	if resp == nil || len(body) == 0 {
		return nil
	}
	messageBytes, err := decodeSingleFrame(body)
	if err != nil {
		return err
	}
	return proto.Unmarshal(messageBytes, resp)
}

func (svc *LDKServerService) authHeader(framedReq []byte) string {
	timestamp := uint64(time.Now().Unix())
	mac := hmac.New(sha256.New, []byte(svc.apiKey))
	var timestampBytes [8]byte
	binary.BigEndian.PutUint64(timestampBytes[:], timestamp)
	mac.Write(timestampBytes[:])
	mac.Write(framedReq)
	return fmt.Sprintf("HMAC %d:%s", timestamp, hex.EncodeToString(mac.Sum(nil)))
}

func (svc *LDKServerService) subscribeEvents() {
	for {
		select {
		case <-svc.ctx.Done():
			return
		default:
		}

		framedReq := grpcFrame([]byte{0})
		reqBytes, err := proto.Marshal(&ldkapi.SubscribeEventsRequest{})
		if err == nil {
			framedReq = grpcFrame(reqBytes)
		}

		httpReq, err := http.NewRequestWithContext(svc.ctx, http.MethodPost, svc.baseURL+ldkapi.LightningNode_SubscribeEvents_FullMethodName, bytes.NewReader(framedReq))
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to build ldk-server event stream request")
			return
		}
		httpReq.Header.Set("Content-Type", "application/grpc+proto")
		httpReq.Header.Set("TE", "trailers")
		httpReq.Header.Set("X-Auth", svc.authHeader(framedReq))

		resp, err := svc.client.Do(httpReq)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to subscribe to ldk-server events")
			select {
			case <-svc.ctx.Done():
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}

		reader := frameReader{reader: resp.Body}
		for {
			frame, err := reader.Next()
			if err != nil {
				resp.Body.Close()
				if !errors.Is(err, io.EOF) && !errors.Is(err, context.Canceled) {
					logger.Logger.WithError(err).Error("ldk-server event stream ended")
				}
				if statusErr := grpcStatusError(resp, nil); statusErr != nil && svc.ctx.Err() == nil {
					logger.Logger.WithError(statusErr).Error("ldk-server event stream returned non-OK status")
				}
				select {
				case <-svc.ctx.Done():
					return
				case <-time.After(2 * time.Second):
					goto reconnect
				}
			}

			event := &ldkevents.EventEnvelope{}
			if err := proto.Unmarshal(frame, event); err != nil {
				logger.Logger.WithError(err).Error("Failed to decode ldk-server event")
				continue
			}
			svc.handleEvent(event)
		}

	reconnect:
	}
}

func (svc *LDKServerService) handleEvent(event *ldkevents.EventEnvelope) {
	switch e := event.Event.(type) {
	case *ldkevents.EventEnvelope_PaymentReceived:
		transaction, err := paymentToTransaction(e.PaymentReceived.Payment)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to convert ldk-server payment received event")
			return
		}
		transaction.Metadata = appendCustomRecords(transaction.Metadata, e.PaymentReceived.CustomRecords)
		svc.eventPublisher.Publish(&events.Event{
			Event:      "nwc_lnclient_payment_received",
			Properties: transaction,
		})
	case *ldkevents.EventEnvelope_PaymentSuccessful:
		transaction, err := paymentToTransaction(e.PaymentSuccessful.Payment)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to convert ldk-server payment successful event")
			return
		}
		svc.eventPublisher.Publish(&events.Event{
			Event:      "nwc_lnclient_payment_sent",
			Properties: transaction,
		})
	case *ldkevents.EventEnvelope_PaymentFailed:
		transaction, err := paymentToTransaction(e.PaymentFailed.Payment)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to convert ldk-server payment failed event")
			return
		}
		svc.eventPublisher.Publish(&events.Event{
			Event: "nwc_lnclient_payment_failed",
			Properties: &lnclient.PaymentFailedEventProperties{
				Transaction: transaction,
				Reason:      "PaymentFailed",
			},
		})
	case *ldkevents.EventEnvelope_PaymentForwarded:
		forwarded := e.PaymentForwarded.ForwardedPayment
		if forwarded.TotalFeeEarnedMsat == nil || forwarded.OutboundAmountForwardedMsat == nil {
			return
		}
		svc.eventPublisher.Publish(&events.Event{
			Event: "nwc_payment_forwarded",
			Properties: &lnclient.PaymentForwardedEventProperties{
				TotalFeeEarnedMsat:          *forwarded.TotalFeeEarnedMsat,
				OutboundAmountForwardedMsat: *forwarded.OutboundAmountForwardedMsat,
			},
		})
	case *ldkevents.EventEnvelope_PaymentClaimable:
		transaction, err := paymentToTransaction(e.PaymentClaimable.Payment)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to convert ldk-server payment claimable event")
			return
		}
		transaction.Metadata = appendCustomRecords(transaction.Metadata, e.PaymentClaimable.CustomRecords)
		svc.eventPublisher.Publish(&events.Event{
			Event:      "nwc_lnclient_hold_invoice_accepted",
			Properties: transaction,
		})
	case *ldkevents.EventEnvelope_ChannelStateChanged:
		svc.handleChannelStateChanged(e.ChannelStateChanged)
	}
}

func (svc *LDKServerService) handleChannelStateChanged(event *ldkevents.ChannelStateChanged) {
	switch event.State {
	case ldkevents.ChannelState_CHANNEL_STATE_READY:
		svc.eventPublisher.Publish(&events.Event{
			Event: "nwc_channel_ready",
			Properties: map[string]interface{}{
				"counterparty_node_id": event.GetCounterpartyNodeId(),
				"node_type":            config.LDKServerBackendType,
			},
		})
	case ldkevents.ChannelState_CHANNEL_STATE_CLOSED, ldkevents.ChannelState_CHANNEL_STATE_OPEN_FAILED:
		reason := ""
		if event.Reason != nil {
			reason = event.Reason.Message
		}
		svc.eventPublisher.Publish(&events.Event{
			Event: "nwc_channel_closed",
			Properties: map[string]interface{}{
				"counterparty_node_id": event.GetCounterpartyNodeId(),
				"reason":               reason,
				"node_type":            config.LDKServerBackendType,
			},
		})
	}
}

func (svc *LDKServerService) waitForPaymentTerminal(paymentID string) (*ldktypes.Payment, error) {
	ctx, cancel := context.WithTimeout(svc.ctx, 2*time.Minute)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		resp := &ldkapi.GetPaymentDetailsResponse{}
		if err := svc.doUnary(ctx, ldkapi.LightningNode_GetPaymentDetails_FullMethodName, &ldkapi.GetPaymentDetailsRequest{
			PaymentId: paymentID,
		}, resp); err != nil {
			return nil, err
		}
		if resp.Payment != nil && resp.Payment.Status != ldktypes.PaymentStatus_PENDING {
			return resp.Payment, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}

func (svc *LDKServerService) waitForFundingTxID(userChannelID string) (string, error) {
	ctx, cancel := context.WithTimeout(svc.ctx, 2*time.Minute)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		channels, err := svc.ListChannels(ctx)
		if err != nil {
			return "", err
		}
		for _, channel := range channels {
			if channel.Id == userChannelID && channel.FundingTxId != "" {
				return channel.FundingTxId, nil
			}
		}

		select {
		case <-ctx.Done():
			return "", fmt.Errorf("timed out waiting for ldk-server funding transaction for channel %s", userChannelID)
		case <-ticker.C:
		}
	}
}

func (svc *LDKServerService) listAllPayments(ctx context.Context) ([]*ldktypes.Payment, error) {
	var token *ldktypes.PageToken
	var payments []*ldktypes.Payment
	for {
		resp := &ldkapi.ListPaymentsResponse{}
		req := &ldkapi.ListPaymentsRequest{PageToken: token}
		if err := svc.doUnary(ctx, ldkapi.LightningNode_ListPayments_FullMethodName, req, resp); err != nil {
			return nil, err
		}
		payments = append(payments, resp.Payments...)
		if resp.NextPageToken == nil {
			return payments, nil
		}
		token = resp.NextPageToken
	}
}

func (svc *LDKServerService) findPayment(ctx context.Context, match func(*ldktypes.Payment) bool) (*ldktypes.Payment, error) {
	payments, err := svc.listAllPayments(ctx)
	if err != nil {
		return nil, err
	}
	for _, payment := range payments {
		if match(payment) {
			return payment, nil
		}
	}
	return nil, errors.New("payment not found")
}

func (svc *LDKServerService) transactionFromCreatedInvoice(ctx context.Context, invoice string, paymentHash string) (*lnclient.Transaction, error) {
	decodeResp := &ldkapi.DecodeInvoiceResponse{}
	if err := svc.doUnary(ctx, ldkapi.LightningNode_DecodeInvoice_FullMethodName, &ldkapi.DecodeInvoiceRequest{
		Invoice: invoice,
	}, decodeResp); err != nil {
		return nil, err
	}

	transaction := &lnclient.Transaction{
		Type:            "incoming",
		Invoice:         invoice,
		PaymentHash:     decodeResp.PaymentHash,
		AmountMsat:      int64(decodeResp.GetAmountMsat()),
		CreatedAt:       int64(decodeResp.GetTimestamp()),
		Description:     decodeResp.GetDescription(),
		DescriptionHash: decodeResp.GetDescriptionHash(),
		Metadata:        lnclient.Metadata{},
	}
	if decodeResp.Expiry > 0 {
		expiresAt := int64(decodeResp.GetTimestamp() + decodeResp.GetExpiry())
		transaction.ExpiresAt = &expiresAt
	}
	if paymentHash != "" {
		transaction.PaymentHash = paymentHash
	}

	lookup, err := svc.LookupInvoice(ctx, transaction.PaymentHash)
	if err == nil {
		lookup.Invoice = invoice
		if lookup.Description == "" {
			lookup.Description = transaction.Description
		}
		if lookup.DescriptionHash == "" {
			lookup.DescriptionHash = transaction.DescriptionHash
		}
		if lookup.ExpiresAt == nil {
			lookup.ExpiresAt = transaction.ExpiresAt
		}
		if lookup.Metadata == nil {
			lookup.Metadata = lnclient.Metadata{}
		}
		return lookup, nil
	}

	return transaction, nil
}

func (svc *LDKServerService) findPeer(ctx context.Context, nodeID string) (*ldktypes.Peer, error) {
	resp := &ldkapi.ListPeersResponse{}
	if err := svc.doUnary(ctx, ldkapi.LightningNode_ListPeers_FullMethodName, &ldkapi.ListPeersRequest{}, resp); err != nil {
		return nil, err
	}
	for _, peer := range resp.Peers {
		if peer.NodeId == nodeID {
			return peer, nil
		}
	}
	return nil, fmt.Errorf("peer %s not found; connect it before opening a channel", nodeID)
}

func paymentToTransaction(payment *ldktypes.Payment) (*lnclient.Transaction, error) {
	if payment == nil {
		return nil, errors.New("payment is nil")
	}

	transaction := &lnclient.Transaction{
		AmountMsat:   int64(payment.GetAmountMsat()),
		FeesPaidMsat: int64(payment.GetFeePaidMsat()),
		CreatedAt:    int64(payment.LatestUpdateTimestamp),
		Metadata:     lnclient.Metadata{},
	}
	if payment.Direction == ldktypes.PaymentDirection_OUTBOUND {
		transaction.Type = "outgoing"
	} else {
		transaction.Type = "incoming"
	}
	if payment.Status == ldktypes.PaymentStatus_SUCCEEDED {
		settledAt := int64(payment.LatestUpdateTimestamp)
		transaction.SettledAt = &settledAt
	}

	switch kind := payment.Kind.Kind.(type) {
	case *ldktypes.PaymentKind_Bolt11:
		transaction.PaymentHash = kind.Bolt11.Hash
		if kind.Bolt11.Preimage != nil {
			transaction.Preimage = *kind.Bolt11.Preimage
		}
	case *ldktypes.PaymentKind_Spontaneous:
		transaction.PaymentHash = kind.Spontaneous.Hash
		if kind.Spontaneous.Preimage != nil {
			transaction.Preimage = *kind.Spontaneous.Preimage
		}
	case *ldktypes.PaymentKind_Bolt12Offer:
		if kind.Bolt12Offer.Hash != nil {
			transaction.PaymentHash = *kind.Bolt12Offer.Hash
		}
		if kind.Bolt12Offer.Preimage != nil {
			transaction.Preimage = *kind.Bolt12Offer.Preimage
		}
		transaction.Metadata["offer"] = map[string]interface{}{
			"id":         kind.Bolt12Offer.OfferId,
			"payer_note": kind.Bolt12Offer.GetPayerNote(),
		}
	case *ldktypes.PaymentKind_Bolt12Refund:
		if kind.Bolt12Refund.Hash != nil {
			transaction.PaymentHash = *kind.Bolt12Refund.Hash
		}
		if kind.Bolt12Refund.Preimage != nil {
			transaction.Preimage = *kind.Bolt12Refund.Preimage
		}
	case *ldktypes.PaymentKind_Onchain:
		transaction.PaymentHash = kind.Onchain.Txid
	}
	return transaction, nil
}

func paymentHashMatches(payment *ldktypes.Payment, paymentHash string) bool {
	switch kind := payment.Kind.Kind.(type) {
	case *ldktypes.PaymentKind_Bolt11:
		return kind.Bolt11.Hash == paymentHash
	case *ldktypes.PaymentKind_Spontaneous:
		return kind.Spontaneous.Hash == paymentHash
	case *ldktypes.PaymentKind_Bolt12Offer:
		return kind.Bolt12Offer.Hash != nil && *kind.Bolt12Offer.Hash == paymentHash
	case *ldktypes.PaymentKind_Bolt12Refund:
		return kind.Bolt12Refund.Hash != nil && *kind.Bolt12Refund.Hash == paymentHash
	default:
		return false
	}
}

func appendCustomRecords(metadata lnclient.Metadata, records []*ldktypes.CustomTlvRecord) lnclient.Metadata {
	if metadata == nil {
		metadata = lnclient.Metadata{}
	}
	if len(records) == 0 {
		return metadata
	}
	tlvs := make([]lnclient.TLVRecord, 0, len(records))
	for _, record := range records {
		tlvs = append(tlvs, lnclient.TLVRecord{
			Type:  record.TypeNum,
			Value: hex.EncodeToString(record.Value),
		})
	}
	metadata["custom_records"] = tlvs
	return metadata
}

func newInvoiceDescription(description string, descriptionHash string) *ldktypes.Bolt11InvoiceDescription {
	if descriptionHash != "" {
		return &ldktypes.Bolt11InvoiceDescription{
			Kind: &ldktypes.Bolt11InvoiceDescription_Hash{Hash: descriptionHash},
		}
	}
	return &ldktypes.Bolt11InvoiceDescription{
		Kind: &ldktypes.Bolt11InvoiceDescription_Direct{Direct: description},
	}
}

func networkToString(network ldktypes.Network) string {
	switch network {
	case ldktypes.Network_TESTNET:
		return "testnet"
	case ldktypes.Network_TESTNET4:
		return "testnet4"
	case ldktypes.Network_SIGNET:
		return "signet"
	case ldktypes.Network_REGTEST:
		return "regtest"
	default:
		return "bitcoin"
	}
}

func grpcFrame(msg []byte) []byte {
	frame := make([]byte, 5+len(msg))
	binary.BigEndian.PutUint32(frame[1:5], uint32(len(msg)))
	copy(frame[5:], msg)
	return frame
}

func decodeSingleFrame(data []byte) ([]byte, error) {
	reader := frameReader{reader: bytes.NewReader(data)}
	return reader.Next()
}

type frameReader struct {
	reader io.Reader
}

func (f *frameReader) Next() ([]byte, error) {
	header := make([]byte, 5)
	if _, err := io.ReadFull(f.reader, header); err != nil {
		return nil, err
	}
	if header[0] != 0 {
		return nil, errors.New("compressed gRPC frames are not supported")
	}
	length := binary.BigEndian.Uint32(header[1:5])
	payload := make([]byte, length)
	if _, err := io.ReadFull(f.reader, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func grpcStatusError(resp *http.Response, body []byte) error {
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ldk-server returned HTTP %d", resp.StatusCode)
	}

	statusCode := resp.Trailer.Get("grpc-status")
	if statusCode == "" {
		statusCode = resp.Header.Get("grpc-status")
	}
	if statusCode == "" || statusCode == "0" {
		return nil
	}

	message := resp.Trailer.Get("grpc-message")
	if message == "" {
		message = resp.Header.Get("grpc-message")
	}
	if message != "" {
		if decoded, err := url.QueryUnescape(message); err == nil {
			message = decoded
		}
	}
	if message == "" {
		message = strings.TrimSpace(string(body))
	}
	return fmt.Errorf("ldk-server gRPC error %s: %s", statusCode, message)
}

func parseNodeURI(uri string) (string, string, int, error) {
	parts := strings.SplitN(uri, "@", 2)
	if len(parts) != 2 {
		return "", "", 0, fmt.Errorf("invalid node URI: %s", uri)
	}
	host, port, err := splitHostPort(parts[1])
	if err != nil {
		return "", "", 0, err
	}
	return parts[0], host, port, nil
}

func splitHostPort(address string) (string, int, error) {
	host, portString, err := net.SplitHostPort(address)
	if err != nil {
		return "", 0, err
	}
	port, err := net.LookupPort("tcp", portString)
	if err != nil {
		return "", 0, err
	}
	return host, port, nil
}

func uint64Ptr(v uint64) *uint64 {
	return &v
}

func boolPtr(v bool) *bool {
	return &v
}

func max[T ~int64 | ~uint32](a, b T) T {
	if a > b {
		return a
	}
	return b
}
