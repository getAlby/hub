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

	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/lsp"
	"github.com/getAlby/hub/utils"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
)

type lspInfo struct {
	Pubkey                 string
	Address                string
	Port                   uint16
	MaxChannelExpiryBlocks uint64
}

func (api *api) NewInstantChannelInvoice(ctx context.Context, request *NewInstantChannelInvoiceRequest) (*NewInstantChannelInvoiceResponse, error) {

	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}

	if request.LSPType != lsp.LSP_TYPE_LSPS1 && request.Public {
		return nil, errors.New("This LSP option does not support public channels")
	}

	logger.Logger.Infoln("Requesting LSP info")

	var lspInfo *lspInfo
	var err error
	switch request.LSPType {
	case lsp.LSP_TYPE_FLOW_2_0:
		fallthrough
	case lsp.LSP_TYPE_PMLSP:
		lspInfo, err = api.getFlowLSPInfo(request.LSPUrl + "/info")

	case lsp.LSP_TYPE_LSPS1:
		lspInfo, err = api.getLSPS1LSPInfo(request.LSPUrl + "/get_info")

	default:
		return nil, fmt.Errorf("unsupported LSP type: %v", request.LSPType)
	}
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to request LSP info")
		return nil, err
	}

	logger.Logger.Infoln("Requesting own node info")

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

	invoice := ""
	var fee uint64 = 0

	switch request.LSPType {
	case lsp.LSP_TYPE_FLOW_2_0:
		invoice, fee, err = api.requestFlow20WrappedInvoice(ctx, request, nodeInfo.Pubkey)
	case lsp.LSP_TYPE_PMLSP:
		invoice, fee, err = api.requestPMLSPInvoice(request, nodeInfo.Pubkey)
	case lsp.LSP_TYPE_LSPS1:
		invoice, fee, err = api.requestLSPS1Invoice(ctx, request, nodeInfo.Pubkey, lspInfo.MaxChannelExpiryBlocks)
	}
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

	newChannelResponse := &NewInstantChannelInvoiceResponse{
		Invoice:           invoice,
		Fee:               fee,
		InvoiceAmount:     invoiceAmount,
		IncomingLiquidity: incomingLiquidity,
		OutgoingLiquidity: outgoingLiquidity,
	}

	logger.Logger.WithFields(logrus.Fields{
		"newChannelResponse": newChannelResponse,
	}).Info("New Channel response")

	return newChannelResponse, nil
}

func (api *api) getLSPS1LSPInfo(url string) (*lspInfo, error) {

	type lsps1LSPInfo struct {
		MaxChannelExpiryBlocks uint64   `json:"max_channel_expiry_blocks"`
		URIs                   []string `json:"uris"`
	}
	var lsps1LspInfo lsps1LSPInfo
	client := http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to create lsp info request")
		return nil, err
	}

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
		Pubkey:                 parts[1],
		Address:                parts[2],
		Port:                   uint16(port),
		MaxChannelExpiryBlocks: lsps1LspInfo.MaxChannelExpiryBlocks,
	}, nil
}
func (api *api) getFlowLSPInfo(url string) (*lspInfo, error) {
	type FlowLSPConnectionMethod struct {
		Address string `json:"address"`
		Port    uint16 `json:"port"`
		Type    string `json:"type"`
	}
	type FlowLSPInfo struct {
		Pubkey            string                    `json:"pubkey"`
		ConnectionMethods []FlowLSPConnectionMethod `json:"connection_methods"`
	}
	var flowLspInfo FlowLSPInfo
	client := http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to create lsp info request")
		return nil, err
	}

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

	err = json.Unmarshal(body, &flowLspInfo)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to deserialize json")
		return nil, fmt.Errorf("failed to deserialize json %s %s", url, string(body))
	}

	ipIndex := -1
	for i, cm := range flowLspInfo.ConnectionMethods {
		if strings.HasPrefix(cm.Type, "ip") {
			ipIndex = i
			break
		}
	}

	if ipIndex == -1 {
		logger.Logger.Error("No ipv4/ipv6 connection method found in LSP info")
		return nil, errors.New("unexpected LSP connection method")
	}

	return &lspInfo{
		Pubkey:  flowLspInfo.Pubkey,
		Address: flowLspInfo.ConnectionMethods[ipIndex].Address,
		Port:    flowLspInfo.ConnectionMethods[ipIndex].Port,
	}, nil
}

func (api *api) requestFlow20WrappedInvoice(ctx context.Context, request *NewInstantChannelInvoiceRequest, pubkey string) (invoice string, fee uint64, err error) {
	logger.Logger.Infoln("Requesting fee information")

	type FeeRequest struct {
		AmountMsat uint64 `json:"amount_msat"`
		Pubkey     string `json:"pubkey"`
	}
	type FeeResponse struct {
		FeeAmountMsat uint64 `json:"fee_amount_msat"`
		Id            string `json:"id"`
	}

	var feeResponse FeeResponse
	{
		client := http.Client{
			Timeout: time.Second * 10,
		}
		payloadBytes, err := json.Marshal(FeeRequest{
			AmountMsat: request.Amount * 1000,
			Pubkey:     pubkey,
		})
		if err != nil {
			return "", 0, err
		}
		bodyReader := bytes.NewReader(payloadBytes)

		req, err := http.NewRequest(http.MethodPost, request.LSPUrl+"/fee", bodyReader)
		if err != nil {
			logger.Logger.WithError(err).WithFields(logrus.Fields{
				"url": request.LSPUrl,
			}).Error("Failed to create lsp fee request")
			return "", 0, err
		}

		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			logger.Logger.WithError(err).WithFields(logrus.Fields{
				"url": request.LSPUrl,
			}).Error("Failed to request lsp fee")
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
				"body":       string(body),
				"statusCode": res.StatusCode,
			}).Error("fee endpoint returned non-success code")
			return "", 0, fmt.Errorf("fee endpoint returned non-success code: %s", string(body))
		}

		err = json.Unmarshal(body, &feeResponse)
		if err != nil {
			logger.Logger.WithError(err).WithFields(logrus.Fields{
				"url": request.LSPUrl,
			}).Error("Failed to deserialize json")
			return "", 0, fmt.Errorf("failed to deserialize json %s %s", request.LSPUrl, string(body))
		}

		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url":         request.LSPUrl,
			"feeResponse": feeResponse,
		}).Info("Got fee response")
		if feeResponse.Id == "" {
			logger.Logger.WithError(err).WithFields(logrus.Fields{
				"feeResponse": feeResponse,
			}).Error("No fee id in fee response")
			return "", 0, fmt.Errorf("no fee id in fee response %v", feeResponse)
		}
		fee = feeResponse.FeeAmountMsat / 1000
	}

	// because we don't want the sender to pay the fee
	// see: https://docs.voltage.cloud/voltage-lsp#gqBqV
	makeInvoiceResponse, err := api.svc.GetLNClient().MakeInvoice(ctx, int64(request.Amount)*1000-int64(feeResponse.FeeAmountMsat), "", "", 60*60)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to request own invoice")
		return "", 0, fmt.Errorf("failed to request own invoice %v", err)
	}

	type ProposalRequest struct {
		Bolt11 string `json:"bolt11"`
		FeeId  string `json:"fee_id"`
	}
	type ProposalResponse struct {
		Bolt11 string `json:"jit_bolt11"`
	}

	logger.Logger.Infoln("Proposing invoice")

	var proposalResponse ProposalResponse
	{
		client := http.Client{
			Timeout: time.Second * 10,
		}
		payloadBytes, err := json.Marshal(ProposalRequest{
			Bolt11: makeInvoiceResponse.Invoice,
			FeeId:  feeResponse.Id,
		})
		if err != nil {
			return "", 0, err
		}
		bodyReader := bytes.NewReader(payloadBytes)

		req, err := http.NewRequest(http.MethodPost, request.LSPUrl+"/proposal", bodyReader)
		if err != nil {
			logger.Logger.WithError(err).WithFields(logrus.Fields{
				"url": request.LSPUrl,
			}).Error("Failed to create lsp fee request")
			return "", 0, err
		}

		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			logger.Logger.WithError(err).WithFields(logrus.Fields{
				"url": request.LSPUrl,
			}).Error("Failed to request lsp fee")
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
				"body":       string(body),
				"statusCode": res.StatusCode,
			}).Error("proposal endpoint returned non-success code")
			return "", 0, fmt.Errorf("proposal endpoint returned non-success code: %s", string(body))
		}

		err = json.Unmarshal(body, &proposalResponse)
		if err != nil {
			logger.Logger.WithError(err).WithFields(logrus.Fields{
				"url": request.LSPUrl,
			}).Error("Failed to deserialize json")
			return "", 0, fmt.Errorf("failed to deserialize json %s %s", request.LSPUrl, string(body))
		}
		logger.Logger.WithField("proposalResponse", proposalResponse).Info("Got proposal response")
		if proposalResponse.Bolt11 == "" {
			logger.Logger.WithError(err).WithFields(logrus.Fields{
				"url":              request.LSPUrl,
				"proposalResponse": proposalResponse,
			}).Error("No bolt11 in proposal response")
			return "", 0, fmt.Errorf("no bolt11 in proposal response %v", proposalResponse)
		}
	}
	invoice = proposalResponse.Bolt11

	return invoice, fee, nil
}

func (api *api) requestPMLSPInvoice(request *NewInstantChannelInvoiceRequest, pubkey string) (invoice string, fee uint64, err error) {
	type NewInstantChannelRequest struct {
		Amount uint64 `json:"amount"`
		Pubkey string `json:"pubkey"`
	}

	client := http.Client{
		Timeout: time.Second * 10,
	}
	payloadBytes, err := json.Marshal(NewInstantChannelRequest{
		Amount: request.Amount,
		Pubkey: pubkey,
	})
	if err != nil {
		return "", 0, err
	}
	bodyReader := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest(http.MethodPost, request.LSPUrl+"/new-channel", bodyReader)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": request.LSPUrl,
		}).Error("Failed to create new channel request")
		return "", 0, err
	}

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
			"body":       string(body),
			"statusCode": res.StatusCode,
		}).Error("new-channel endpoint returned non-success code")
		return "", 0, fmt.Errorf("new-channel endpoint returned non-success code: %s", string(body))
	}

	type newInstantChannelResponse struct {
		FeeAmountMsat uint64 `json:"fee_amount_msat"`
		Invoice       string `json:"invoice"`
	}

	var newChannelResponse newInstantChannelResponse

	err = json.Unmarshal(body, &newChannelResponse)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": request.LSPUrl,
		}).Error("Failed to deserialize json")
		return "", 0, fmt.Errorf("failed to deserialize json %s %s", request.LSPUrl, string(body))
	}

	invoice = newChannelResponse.Invoice
	fee = newChannelResponse.FeeAmountMsat / 1000

	return invoice, fee, nil
}

func (api *api) requestLSPS1Invoice(ctx context.Context, request *NewInstantChannelInvoiceRequest, pubkey string, channelExpiryBlocks uint64) (invoice string, fee uint64, err error) {
	client := http.Client{
		Timeout: time.Second * 10,
	}

	type NewLSPS1ChannelRequest struct {
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

	if request.Public {
		// as per BOLT-7 6 confirmations are required for the channel to be gossiped
		// https://github.com/lightning/bolts/blob/master/07-routing-gossip.md#requirements
		requiredChannelConfirmations = 6
	}

	newLSPS1ChannelRequest := NewLSPS1ChannelRequest{
		PublicKey:                    pubkey,
		LSPBalanceSat:                strconv.FormatUint(request.Amount, 10),
		ClientBalanceSat:             "0",
		RequiredChannelConfirmations: requiredChannelConfirmations,
		FundingConfirmsWithinBlocks:  6,
		ChannelExpiryBlocks:          channelExpiryBlocks,
		Token:                        "",
		RefundOnchainAddress:         refundAddress,
		AnnounceChannel:              request.Public,
	}

	payloadBytes, err := json.Marshal(newLSPS1ChannelRequest)
	if err != nil {
		return "", 0, err
	}
	bodyReader := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest(http.MethodPost, request.LSPUrl+"/create_order", bodyReader)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": request.LSPUrl,
		}).Error("Failed to create new channel request")
		return "", 0, err
	}

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
		Payment newLSPS1ChannelPayment `json:"payment"`
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
