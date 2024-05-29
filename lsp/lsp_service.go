package lsp

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

	"github.com/getAlby/nostr-wallet-connect/lnclient"
	"github.com/getAlby/nostr-wallet-connect/service"
	"github.com/sirupsen/logrus"
)

type lspService struct {
	svc    service.Service
	logger *logrus.Logger
}

type lspConnectionInfo struct {
	Pubkey  string
	Address string
	Port    uint16
}

func NewLSPService(svc service.Service, logger *logrus.Logger) *lspService {
	return &lspService{
		svc:    svc,
		logger: logger,
	}
}

func (ls *lspService) NewInstantChannelInvoice(ctx context.Context, request *NewInstantChannelInvoiceRequest) (*NewInstantChannelInvoiceResponse, error) {
	var selectedLsp LSP
	switch request.LSP {
	case "VOLTAGE":
		selectedLsp = VoltageLSP()
	case "OLYMPUS_FLOW_2_0":
		selectedLsp = OlympusLSP()
	case "OLYMPUS_MUTINYNET_FLOW_2_0":
		selectedLsp = OlympusMutinynetFlowLSP()
	case "OLYMPUS_MUTINYNET_LSPS1":
		selectedLsp = OlympusMutinynetLSPS1LSP()
	case "ALBY":
		selectedLsp = AlbyPlebsLSP()
	case "ALBY_MUTINYNET":
		selectedLsp = AlbyMutinynetPlebsLSP()
	default:
		return nil, errors.New("unknown LSP")
	}

	if ls.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}

	ls.logger.Infoln("Requesting LSP info")

	var lspInfo *lspConnectionInfo
	var err error
	switch selectedLsp.LspType {
	case LSP_TYPE_FLOW_2_0:
		fallthrough
	case LSP_TYPE_PMLSP:
		lspInfo, err = ls.getFlowLSPInfo(selectedLsp.Url + "/info")

	case LSP_TYPE_LSPS1:
		lspInfo, err = ls.getLSPS1LSPInfo(selectedLsp.Url + "/get_info")

	default:
		return nil, fmt.Errorf("unsupported LSP type: %v", selectedLsp.LspType)
	}
	if err != nil {
		ls.logger.WithError(err).Error("Failed to request LSP info")
		return nil, err
	}

	ls.logger.Infoln("Requesting own node info")

	nodeInfo, err := ls.svc.GetLNClient().GetInfo(ctx)
	if err != nil {
		ls.logger.WithError(err).WithFields(logrus.Fields{
			"url": selectedLsp.Url,
		}).Error("Failed to request own node info", err)
		return nil, err
	}

	ls.logger.WithField("lspInfo", lspInfo).Info("Connecting to LSP node as a peer")

	err = ls.svc.GetLNClient().ConnectPeer(ctx, &lnclient.ConnectPeerRequest{
		Pubkey:  lspInfo.Pubkey,
		Address: lspInfo.Address,
		Port:    lspInfo.Port,
	})

	if err != nil {
		ls.logger.WithError(err).Error("Failed to connect to peer")
		return nil, err
	}

	invoice := ""
	var fee uint64 = 0

	switch selectedLsp.LspType {
	case LSP_TYPE_FLOW_2_0:
		invoice, fee, err = ls.requestFlow20WrappedInvoice(ctx, &selectedLsp, request.Amount, nodeInfo.Pubkey)
	case LSP_TYPE_PMLSP:
		invoice, fee, err = ls.requestPMLSPInvoice(&selectedLsp, request.Amount, nodeInfo.Pubkey)
	case LSP_TYPE_LSPS1:
		invoice, fee, err = ls.requestLSPS1Invoice(ctx, &selectedLsp, request.Amount, nodeInfo.Pubkey)

	default:
		return nil, fmt.Errorf("unsupported LSP type: %v", selectedLsp.LspType)
	}
	if err != nil {
		ls.logger.WithError(err).Error("Failed to request invoice")
		return nil, err
	}

	newChannelResponse := &NewInstantChannelInvoiceResponse{
		Invoice: invoice,
		Fee:     fee,
	}

	ls.logger.WithFields(logrus.Fields{
		"newChannelResponse": newChannelResponse,
	}).Info("New Channel response")

	return newChannelResponse, nil
}

func (ls *lspService) getLSPS1LSPInfo(url string) (*lspConnectionInfo, error) {
	type LSPS1LSPInfo struct {
		// TODO: implement options
		Options interface{} `json:"options"`
		URIs    []string    `json:"uris"`
	}
	var lsps1LspInfo LSPS1LSPInfo
	client := http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		ls.logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to create lsp info request")
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		ls.logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to request lsp info")
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		ls.logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to read response body")
		return nil, errors.New("failed to read response body")
	}

	err = json.Unmarshal(body, &lsps1LspInfo)
	if err != nil {
		ls.logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to deserialize json")
		return nil, fmt.Errorf("failed to deserialize json %s %s", url, string(body))
	}

	uri := lsps1LspInfo.URIs[0]

	// make sure it's a valid IPv4 URI
	regex := regexp.MustCompile(`^([0-9a-f]+)@([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+):([0-9]+)$`)
	parts := regex.FindStringSubmatch(uri)
	ls.logger.WithField("parts", parts).Info("Split URI")
	if parts == nil || len(parts) != 4 {
		ls.logger.WithField("parts", parts).Info("Unsupported URI")
		return nil, errors.New("could not decode LSP URI")
	}

	port, err := strconv.Atoi(parts[3])
	if err != nil {
		ls.logger.WithField("port", parts[3]).WithError(err).Info("Failed to decode port number")

		return nil, err
	}

	return &lspConnectionInfo{
		Pubkey:  parts[1],
		Address: parts[2],
		Port:    uint16(port),
	}, nil
}
func (ls *lspService) getFlowLSPInfo(url string) (*lspConnectionInfo, error) {
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
		ls.logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to create lsp info request")
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		ls.logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to request lsp info")
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		ls.logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to read response body")
		return nil, errors.New("failed to read response body")
	}

	err = json.Unmarshal(body, &flowLspInfo)
	if err != nil {
		ls.logger.WithError(err).WithFields(logrus.Fields{
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
		ls.logger.Error("No ipv4/ipv6 connection method found in LSP info")
		return nil, errors.New("unexpected LSP connection method")
	}

	return &lspConnectionInfo{
		Pubkey:  flowLspInfo.Pubkey,
		Address: flowLspInfo.ConnectionMethods[ipIndex].Address,
		Port:    flowLspInfo.ConnectionMethods[ipIndex].Port,
	}, nil
}

func (ls *lspService) requestFlow20WrappedInvoice(ctx context.Context, selectedLsp *LSP, amount uint64, pubkey string) (invoice string, fee uint64, err error) {
	ls.logger.Infoln("Requesting fee information")

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
			AmountMsat: amount * 1000,
			Pubkey:     pubkey,
		})
		if err != nil {
			return "", 0, err
		}
		bodyReader := bytes.NewReader(payloadBytes)

		req, err := http.NewRequest(http.MethodPost, selectedLsp.Url+"/fee", bodyReader)
		if err != nil {
			ls.logger.WithError(err).WithFields(logrus.Fields{
				"url": selectedLsp.Url,
			}).Error("Failed to create lsp fee request")
			return "", 0, err
		}

		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			ls.logger.WithError(err).WithFields(logrus.Fields{
				"url": selectedLsp.Url,
			}).Error("Failed to request lsp fee")
			return "", 0, err
		}

		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			ls.logger.WithError(err).WithFields(logrus.Fields{
				"url": selectedLsp.Url,
			}).Error("Failed to read response body")
			return "", 0, errors.New("failed to read response body")
		}

		if res.StatusCode >= 300 {
			ls.logger.WithFields(logrus.Fields{
				"body":       string(body),
				"statusCode": res.StatusCode,
			}).Error("fee endpoint returned non-success code")
			return "", 0, fmt.Errorf("fee endpoint returned non-success code: %s", string(body))
		}

		err = json.Unmarshal(body, &feeResponse)
		if err != nil {
			ls.logger.WithError(err).WithFields(logrus.Fields{
				"url": selectedLsp.Url,
			}).Error("Failed to deserialize json")
			return "", 0, fmt.Errorf("failed to deserialize json %s %s", selectedLsp.Url, string(body))
		}

		ls.logger.WithError(err).WithFields(logrus.Fields{
			"url":         selectedLsp.Url,
			"feeResponse": feeResponse,
		}).Info("Got fee response")
		if feeResponse.Id == "" {
			ls.logger.WithError(err).WithFields(logrus.Fields{
				"feeResponse": feeResponse,
			}).Error("No fee id in fee response")
			return "", 0, fmt.Errorf("no fee id in fee response %v", feeResponse)
		}
		fee = feeResponse.FeeAmountMsat / 1000
	}

	// because we don't want the sender to pay the fee
	// see: https://docs.voltage.cloud/voltage-lsp#gqBqV
	makeInvoiceResponse, err := ls.svc.GetLNClient().MakeInvoice(ctx, int64(amount)*1000-int64(feeResponse.FeeAmountMsat), "", "", 60*60)
	if err != nil {
		ls.logger.WithError(err).Error("Failed to request own invoice")
		return "", 0, fmt.Errorf("failed to request own invoice %v", err)
	}

	type ProposalRequest struct {
		Bolt11 string `json:"bolt11"`
		FeeId  string `json:"fee_id"`
	}
	type ProposalResponse struct {
		Bolt11 string `json:"jit_bolt11"`
	}

	ls.logger.Infoln("Proposing invoice")

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

		req, err := http.NewRequest(http.MethodPost, selectedLsp.Url+"/proposal", bodyReader)
		if err != nil {
			ls.logger.WithError(err).WithFields(logrus.Fields{
				"url": selectedLsp.Url,
			}).Error("Failed to create lsp fee request")
			return "", 0, err
		}

		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req)
		if err != nil {
			ls.logger.WithError(err).WithFields(logrus.Fields{
				"url": selectedLsp.Url,
			}).Error("Failed to request lsp fee")
			return "", 0, err
		}

		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			ls.logger.WithError(err).WithFields(logrus.Fields{
				"url": selectedLsp.Url,
			}).Error("Failed to read response body")
			return "", 0, errors.New("failed to read response body")
		}

		if res.StatusCode >= 300 {
			ls.logger.WithFields(logrus.Fields{
				"body":       string(body),
				"statusCode": res.StatusCode,
			}).Error("proposal endpoint returned non-success code")
			return "", 0, fmt.Errorf("proposal endpoint returned non-success code: %s", string(body))
		}

		err = json.Unmarshal(body, &proposalResponse)
		if err != nil {
			ls.logger.WithError(err).WithFields(logrus.Fields{
				"url": selectedLsp.Url,
			}).Error("Failed to deserialize json")
			return "", 0, fmt.Errorf("failed to deserialize json %s %s", selectedLsp.Url, string(body))
		}
		ls.logger.WithField("proposalResponse", proposalResponse).Info("Got proposal response")
		if proposalResponse.Bolt11 == "" {
			ls.logger.WithError(err).WithFields(logrus.Fields{
				"url":              selectedLsp.Url,
				"proposalResponse": proposalResponse,
			}).Error("No bolt11 in proposal response")
			return "", 0, fmt.Errorf("no bolt11 in proposal response %v", proposalResponse)
		}
	}
	invoice = proposalResponse.Bolt11

	return invoice, fee, nil
}

func (ls *lspService) requestPMLSPInvoice(selectedLsp *LSP, amount uint64, pubkey string) (invoice string, fee uint64, err error) {
	type NewInstantChannelRequest struct {
		Amount uint64 `json:"amount"`
		Pubkey string `json:"pubkey"`
	}

	client := http.Client{
		Timeout: time.Second * 10,
	}
	payloadBytes, err := json.Marshal(NewInstantChannelRequest{
		Amount: amount,
		Pubkey: pubkey,
	})
	if err != nil {
		return "", 0, err
	}
	bodyReader := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest(http.MethodPost, selectedLsp.Url+"/new-channel", bodyReader)
	if err != nil {
		ls.logger.WithError(err).WithFields(logrus.Fields{
			"url": selectedLsp.Url,
		}).Error("Failed to create new channel request")
		return "", 0, err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		ls.logger.WithError(err).WithFields(logrus.Fields{
			"url": selectedLsp.Url,
		}).Error("Failed to request new channel invoice")
		return "", 0, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		ls.logger.WithError(err).WithFields(logrus.Fields{
			"url": selectedLsp.Url,
		}).Error("Failed to read response body")
		return "", 0, errors.New("failed to read response body")
	}

	if res.StatusCode >= 300 {
		ls.logger.WithFields(logrus.Fields{
			"body":       string(body),
			"statusCode": res.StatusCode,
		}).Error("new-channel endpoint returned non-success code")
		return "", 0, fmt.Errorf("new-channel endpoint returned non-success code: %s", string(body))
	}

	var newChannelResponse NewInstantChannelResponse

	err = json.Unmarshal(body, &newChannelResponse)
	if err != nil {
		ls.logger.WithError(err).WithFields(logrus.Fields{
			"url": selectedLsp.Url,
		}).Error("Failed to deserialize json")
		return "", 0, fmt.Errorf("failed to deserialize json %s %s", selectedLsp.Url, string(body))
	}

	invoice = newChannelResponse.Invoice
	fee = newChannelResponse.FeeAmountMsat / 1000

	return invoice, fee, nil
}

func (ls *lspService) requestLSPS1Invoice(ctx context.Context, selectedLsp *LSP, amount uint64, pubkey string) (invoice string, fee uint64, err error) {
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

	refundAddress, err := ls.svc.GetLNClient().GetNewOnchainAddress(ctx)
	if err != nil {
		ls.logger.WithError(err).Error("Failed to request onchain address")
		return "", 0, err
	}

	newLSPS1ChannelRequest := NewLSPS1ChannelRequest{
		PublicKey:                    pubkey,
		LSPBalanceSat:                strconv.FormatUint(amount, 10),
		ClientBalanceSat:             "0",
		RequiredChannelConfirmations: 0,
		FundingConfirmsWithinBlocks:  6,
		ChannelExpiryBlocks:          13000, // TODO: this should be customizable
		Token:                        "",
		RefundOnchainAddress:         refundAddress,
		AnnounceChannel:              false, // TODO: this should be customizable
	}

	payloadBytes, err := json.Marshal(newLSPS1ChannelRequest)
	if err != nil {
		return "", 0, err
	}
	bodyReader := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest(http.MethodPost, selectedLsp.Url+"/create_order", bodyReader)
	if err != nil {
		ls.logger.WithError(err).WithFields(logrus.Fields{
			"url": selectedLsp.Url,
		}).Error("Failed to create new channel request")
		return "", 0, err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		ls.logger.WithError(err).WithFields(logrus.Fields{
			"url": selectedLsp.Url,
		}).Error("Failed to request new channel invoice")
		return "", 0, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		ls.logger.WithError(err).WithFields(logrus.Fields{
			"url": selectedLsp.Url,
		}).Error("Failed to read response body")
		return "", 0, errors.New("failed to read response body")
	}

	if res.StatusCode >= 300 {
		ls.logger.WithFields(logrus.Fields{
			"newLSPS1ChannelRequest": newLSPS1ChannelRequest,
			"body":                   string(body),
			"statusCode":             res.StatusCode,
		}).Error("create_order endpoint returned non-success code")
		return "", 0, fmt.Errorf("create_order endpoint returned non-success code: %s", string(body))
	}

	type NewLSPS1ChannelPayment struct {
		LightningInvoice string `json:"lightning_invoice"`
		FeeTotalSat      string `json:"fee_total_sat"`
	}
	type NewLSPS1ChannelResponse struct {
		Payment NewLSPS1ChannelPayment `json:"payment"`
	}

	var newChannelResponse NewLSPS1ChannelResponse

	err = json.Unmarshal(body, &newChannelResponse)
	if err != nil {
		ls.logger.WithError(err).WithFields(logrus.Fields{
			"url": selectedLsp.Url,
		}).Error("Failed to deserialize json")
		return "", 0, fmt.Errorf("failed to deserialize json %s %s", selectedLsp.Url, string(body))
	}

	invoice = newChannelResponse.Payment.LightningInvoice
	fee, err = strconv.ParseUint(newChannelResponse.Payment.FeeTotalSat, 10, 64)
	if err != nil {
		ls.logger.WithError(err).WithFields(logrus.Fields{
			"url": selectedLsp.Url,
		}).Error("Failed to parse fee")
		return "", 0, fmt.Errorf("failed to parse fee %v", err)
	}

	return invoice, fee, nil
}
