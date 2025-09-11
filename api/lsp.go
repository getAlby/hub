package api

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/getAlby/hub/alby"
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/lsp"
	"github.com/getAlby/hub/version"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
)

func (api *api) RequestLSPOrder(ctx context.Context, request *LSPOrderRequest) (*LSPOrderResponse, error) {

	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}

	if request.LSPType != lsp.LSP_TYPE_LSPS1 {
		return nil, fmt.Errorf("unsupported LSP type: %v", request.LSPType)
	}

	logger.Logger.Info("Requesting own node info")

	nodeInfo, err := api.svc.GetLNClient().GetInfo(ctx)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"lspIdentifier": request.LSPIdentifier,
		}).Error("Failed to request own node info", err)
		return nil, err
	}

	logger.Logger.Info("Requesting LSP info")
	lspInfo, err := api.albyOAuthSvc.GetLSPInfo(ctx, request.LSPIdentifier, nodeInfo.Network)

	if err != nil {
		logger.Logger.WithError(err).Error("Failed to request LSP info")
		return nil, err
	}

	logger.Logger.WithField("lspInfo", lspInfo).Info("Connecting to LSP node as a peer")

	err = api.svc.GetLNClient().ConnectPeer(ctx, &lnclient.ConnectPeerRequest{
		Pubkey:  lspInfo.Pubkey,
		Address: lspInfo.Address,
		Port:    lspInfo.Port,
	})

	if err != nil {
		logger.Logger.WithError(err).Error("Failed to connect to peer")
		return nil, err
	}

	invoice, fee, err := api.requestLSPS1Invoice(ctx, request, nodeInfo.Network, nodeInfo.Pubkey, lspInfo.MaxChannelExpiryBlocks, lspInfo.MinRequiredChannelConfirmations, lspInfo.MinFundingConfirmsWithinBlocks)

	if err != nil {
		logger.Logger.WithError(err).Error("Failed to request invoice")
		return nil, err
	}

	paymentRequest, err := decodepay.Decodepay(invoice)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to decode bolt11 invoice")
		return nil, err
	}

	invoiceAmount := uint64(paymentRequest.MSatoshi / 1000)
	incomingLiquidity := uint64(0)
	outgoingLiquidity := uint64(0)

	if invoiceAmount < request.Amount {
		// assume that the invoice is only the fee
		// and that the user is requesting incoming liquidity (LSPS1)
		incomingLiquidity = request.Amount
	} else {
		outgoingLiquidity = invoiceAmount - fee
	}

	newChannelResponse := &LSPOrderResponse{
		Invoice:           invoice,
		Fee:               fee,
		InvoiceAmount:     invoiceAmount,
		IncomingLiquidity: incomingLiquidity,
		OutgoingLiquidity: outgoingLiquidity,
	}

	logger.Logger.WithFields(logrus.Fields{
		"newChannelResponse": newChannelResponse,
	}).Debug("New Channel response")

	return newChannelResponse, nil
}

func (api *api) requestLSPS1Invoice(ctx context.Context, request *LSPOrderRequest, network, pubkey string, channelExpiryBlocks uint64, minRequiredChannelConfirmations uint64, minFundingConfirmsWithinBlocks uint64) (invoice string, fee uint64, err error) {
	refundAddress, err := api.svc.GetLNClient().GetNewOnchainAddress(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to request onchain address")
		return "", 0, err
	}

	var requiredChannelConfirmations uint64 = 0

	backendType, err := api.cfg.Get("LNBackendType", "")
	if err != nil {
		return "", 0, errors.New("failed to get LN backend type")
	}

	if backendType != config.LDKBackendType {
		// LND does not support 0-conf by default
		requiredChannelConfirmations = 1
	}

	// Some LSPs (e.g. Olympus) require more min confirmations, as per the spec ours must be at least as many blocks
	requiredChannelConfirmations = max(requiredChannelConfirmations, minRequiredChannelConfirmations)

	if request.Public {
		// as per BOLT-7 6 confirmations are required for the channel to be gossiped
		// https://github.com/lightning/bolts/blob/master/07-routing-gossip.md#requirements
		requiredChannelConfirmations = 6
	}

	token := ""
	if request.LSPIdentifier == "olympus" {
		token = "AlbyHub/" + version.Tag
	}

	// set a non-empty token to notify LNServer that we support 0-conf
	// (Pre-v1.17.2 does not support 0-conf)
	if request.LSPIdentifier == "lnserver" {
		token = "AlbyHub/" + version.Tag
	}

	lsps1ChannelRequest := &alby.LSPChannelRequest{
		PublicKey:                    pubkey,
		LSPBalanceSat:                strconv.FormatUint(request.Amount, 10),
		ClientBalanceSat:             "0",
		RequiredChannelConfirmations: requiredChannelConfirmations,
		FundingConfirmsWithinBlocks:  minFundingConfirmsWithinBlocks,
		ChannelExpiryBlocks:          channelExpiryBlocks,
		Token:                        token,
		RefundOnchainAddress:         refundAddress,
		AnnounceChannel:              request.Public,
	}

	channelResponse, err := api.albyOAuthSvc.CreateLSPOrder(ctx, request.LSPIdentifier, network, lsps1ChannelRequest)
	if err != nil {
		return "", 0, err
	}

	invoice = channelResponse.Payment.Bolt11.Invoice
	fee, err = strconv.ParseUint(channelResponse.Payment.Bolt11.FeeTotalSat, 10, 64)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"lspIdentifier": request.LSPIdentifier,
		}).Error("Failed to parse fee")
		return "", 0, fmt.Errorf("failed to parse fee %v", err)
	}

	return invoice, fee, nil
}
