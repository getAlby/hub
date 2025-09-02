package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/lsp"
	"github.com/getAlby/hub/utils"
	"github.com/getAlby/hub/version"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
)

type lspInfo struct {
	Pubkey                          string
	Address                         string
	Port                            uint16
	MaxChannelExpiryBlocks          uint64
	MinRequiredChannelConfirmations uint64
	MinFundingConfirmsWithinBlocks  uint64
}

func (api *api) RequestLSPOrder(ctx context.Context, request *LSPOrderRequest) (*LSPOrderResponse, error) {

	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}

	if request.LSPType != lsp.LSP_TYPE_LSPS1 {
		return nil, fmt.Errorf("unsupported LSP type: %v", request.LSPType)
	}

	logger.Logger.Info("Requesting LSP info")
	lspInfo, err := api.getLSPS1LSPInfo(ctx, request.LSPUrl+"/get_info")

	if err != nil {
		logger.Logger.WithError(err).Error("Failed to request LSP info")
		return nil, err
	}

	logger.Logger.Info("Requesting own node info")

	nodeInfo, err := api.svc.GetLNClient().GetInfo(ctx)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": request.LSPUrl,
		}).Error("Failed to request own node info", err)
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

	invoice, fee, err := api.requestLSPS1Invoice(ctx, request, nodeInfo.Pubkey, lspInfo.MaxChannelExpiryBlocks, lspInfo.MinRequiredChannelConfirmations, lspInfo.MinFundingConfirmsWithinBlocks)

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

func (api *api) getLSPS1LSPInfo(ctx context.Context, url string) (*lspInfo, error) {

	type lsps1LSPInfo struct {
		MinRequiredChannelConfirmations uint64   `json:"min_required_channel_confirmations"`
		MinFundingConfirmsWithinBlocks  uint64   `json:"min_funding_confirms_within_blocks"`
		MaxChannelExpiryBlocks          uint64   `json:"max_channel_expiry_blocks"`
		URIs                            []string `json:"uris"`
	}
	var lsps1LspInfo lsps1LSPInfo
	client := http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to create lsp info request")
		return nil, err
	}
	setDefaultRequestHeaders(req)

	res, err := client.Do(req)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to request lsp info")
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to read response body")
		return nil, errors.New("failed to read response body")
	}

	if res.StatusCode >= 300 {
		logger.Logger.WithFields(logrus.Fields{
			"url":        url,
			"body":       string(body),
			"statusCode": res.StatusCode,
		}).Error("get_info endpoint returned non-success code")
		return nil, fmt.Errorf("get info endpoint returned non-success code: %s", string(body))
	}

	err = json.Unmarshal(body, &lsps1LspInfo)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to deserialize json")
		return nil, fmt.Errorf("failed to deserialize json %s %s", url, string(body))
	}

	httpUris := utils.Filter(lsps1LspInfo.URIs, func(uri string) bool {
		return !strings.Contains(uri, ".onion")
	})
	if len(httpUris) == 0 {
		logger.Logger.WithField("uris", lsps1LspInfo.URIs).WithError(err).Error("Couldn't find HTTP URI")

		return nil, err
	}
	uri := httpUris[0]

	// make sure it's a valid IPv4 URI
	regex := regexp.MustCompile(`^([0-9a-f]+)@([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+):([0-9]+)$`)
	parts := regex.FindStringSubmatch(uri)
	logger.Logger.WithField("parts", parts).Info("Split URI")
	if parts == nil || len(parts) != 4 {
		logger.Logger.WithField("parts", parts).Error("Unsupported URI")
		return nil, errors.New("could not decode LSP URI")
	}

	port, err := strconv.Atoi(parts[3])
	if err != nil {
		logger.Logger.WithField("port", parts[3]).WithError(err).Error("Failed to decode port number")

		return nil, err
	}

	return &lspInfo{
		Pubkey:                          parts[1],
		Address:                         parts[2],
		Port:                            uint16(port),
		MaxChannelExpiryBlocks:          lsps1LspInfo.MaxChannelExpiryBlocks,
		MinRequiredChannelConfirmations: lsps1LspInfo.MinRequiredChannelConfirmations,
		MinFundingConfirmsWithinBlocks:  lsps1LspInfo.MinFundingConfirmsWithinBlocks,
	}, nil
}

func (api *api) requestLSPS1Invoice(ctx context.Context, request *LSPOrderRequest, pubkey string, channelExpiryBlocks uint64, minRequiredChannelConfirmations uint64, minFundingConfirmsWithinBlocks uint64) (invoice string, fee uint64, err error) {
	client := http.Client{
		Timeout: time.Second * 60,
	}

	type lsps1ChannelRequest struct {
		PublicKey                    string `json:"public_key"`
		LSPBalanceSat                string `json:"lsp_balance_sat"`
		ClientBalanceSat             string `json:"client_balance_sat"`
		RequiredChannelConfirmations uint64 `json:"required_channel_confirmations"`
		FundingConfirmsWithinBlocks  uint64 `json:"funding_confirms_within_blocks"`
		ChannelExpiryBlocks          uint64 `json:"channel_expiry_blocks"`
		Token                        string `json:"token"`
		RefundOnchainAddress         string `json:"refund_onchain_address"`
		AnnounceChannel              bool   `json:"announce_channel"`
	}

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
	if request.LSPUrl == "https://lsps1.lnolymp.us/api/v1" {
		token = "AlbyHub/" + version.Tag
	}

	// set a non-empty token to notify LNServer that we support 0-conf
	// (Pre-v1.17.2 does not support 0-conf)
	if request.LSPUrl == "https://www.lnserver.com/lsp/wave" {
		token = "AlbyHub/" + version.Tag
	}

	newLSPS1ChannelRequest := lsps1ChannelRequest{
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

	payloadBytes, err := json.Marshal(newLSPS1ChannelRequest)
	if err != nil {
		return "", 0, err
	}
	bodyReader := bytes.NewReader(payloadBytes)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, request.LSPUrl+"/create_order", bodyReader)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": request.LSPUrl,
		}).Error("Failed to create new channel request")
		return "", 0, err
	}

	setDefaultRequestHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": request.LSPUrl,
		}).Error("Failed to request new channel invoice")
		return "", 0, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": request.LSPUrl,
		}).Error("Failed to read response body")
		return "", 0, errors.New("failed to read response body")
	}

	if res.StatusCode >= 300 {
		logger.Logger.WithFields(logrus.Fields{
			"newLSPS1ChannelRequest": newLSPS1ChannelRequest,
			"body":                   string(body),
			"statusCode":             res.StatusCode,
		}).Error("create_order endpoint returned non-success code")
		return "", 0, fmt.Errorf("create_order endpoint returned non-success code: %s", string(body))
	}

	type newLSPS1ChannelPaymentBolt11 struct {
		Invoice     string `json:"invoice"`
		FeeTotalSat string `json:"fee_total_sat"`
	}

	type newLSPS1ChannelPayment struct {
		Bolt11 newLSPS1ChannelPaymentBolt11 `json:"bolt11"`
		// TODO: add onchain
	}
	type newLSPS1ChannelResponse struct {
		Payment *newLSPS1ChannelPayment `json:"payment"`
	}

	var newChannelResponse newLSPS1ChannelResponse

	err = json.Unmarshal(body, &newChannelResponse)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": request.LSPUrl,
		}).Error("Failed to deserialize json")
		return "", 0, fmt.Errorf("failed to deserialize json %s %s", request.LSPUrl, string(body))
	}

	invoice = newChannelResponse.Payment.Bolt11.Invoice
	fee, err = strconv.ParseUint(newChannelResponse.Payment.Bolt11.FeeTotalSat, 10, 64)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": request.LSPUrl,
		}).Error("Failed to parse fee")
		return "", 0, fmt.Errorf("failed to parse fee %v", err)
	}

	return invoice, fee, nil
}

func setDefaultRequestHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "AlbyHub/"+version.Tag)
}
