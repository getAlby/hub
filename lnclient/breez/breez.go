//go:build !skip_breez

package breez

import (
	"context"
	"encoding/hex"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/breez/breez-sdk-go/breez_sdk"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"

	"github.com/getAlby/nostr-wallet-connect/lnclient"
	"github.com/getAlby/nostr-wallet-connect/nip47"
)

type BreezService struct {
	listener *BreezListener
	svc      *breez_sdk.BlockingBreezServices
	logger   *logrus.Logger
}
type BreezListener struct {
	logger *logrus.Logger
}

func (listener BreezListener) Log(l breez_sdk.LogEntry) {
	if l.Level != "TRACE" && l.Level != "DEBUG" {
		listener.logger.WithField("level", l.Level).Print(l.Line)
	}
}

func (BreezListener) OnEvent(e breez_sdk.BreezEvent) {
	log.Printf("received event %#v", e)
}

func NewBreezService(logger *logrus.Logger, mnemonic, apiKey, inviteCode, workDir string) (result lnclient.LNClient, err error) {
	if mnemonic == "" || apiKey == "" || inviteCode == "" || workDir == "" {
		return nil, errors.New("one or more required breez configuration are missing")
	}

	//create dir if not exists
	newpath := filepath.Join(workDir)
	err = os.MkdirAll(newpath, os.ModePerm)
	if err != nil {
		return nil, err
	}
	seed, err := breez_sdk.MnemonicToSeed(mnemonic)
	if err != nil {
		return nil, err
	}
	nodeConfig := breez_sdk.NodeConfigGreenlight{
		Config: breez_sdk.GreenlightNodeConfig{
			InviteCode: &inviteCode,
		},
	}
	listener := BreezListener{
		logger: logger,
	}
	config := breez_sdk.DefaultConfig(breez_sdk.EnvironmentTypeProduction, apiKey, nodeConfig)
	config.WorkingDir = workDir
	breez_sdk.SetLogStream(listener)
	svc, err := breez_sdk.Connect(config, seed, listener)
	if err != nil {
		return nil, err
	}
	healthCheck, err := svc.ServiceHealthCheck()
	if err != nil {
		return nil, err
	}
	if err == nil {
		logger.WithField("status", healthCheck.Status).Info("Current service status")
	}

	nodeInfo, err := svc.NodeInfo()
	if err != nil {
		return nil, err
	}
	logger.WithField("info", nodeInfo).Info("Node info")
	logger.WithFields(logrus.Fields{
		"ln balance":                     nodeInfo.ChannelsBalanceMsat,
		"balance":                        nodeInfo.OnchainBalanceMsat,
		"max_payable_msat":               nodeInfo.MaxPayableMsat,
		"max_receivable_msat":            nodeInfo.MaxReceivableMsat,
		"max_single_payment_amount_msat": nodeInfo.MaxSinglePaymentAmountMsat,
		"connected_peers":                nodeInfo.ConnectedPeers,
		"inbound_liquidity_msats":        nodeInfo.InboundLiquidityMsats,
	}).Info("node balances and peers")

	return &BreezService{
		listener: &listener,
		svc:      svc,
		logger:   logger,
	}, nil
}

func (bs *BreezService) Shutdown() error {
	return bs.svc.Disconnect()
}

func (bs *BreezService) SendPaymentSync(ctx context.Context, payReq string) (*lnclient.PayInvoiceResponse, error) {
	sendPaymentRequest := breez_sdk.SendPaymentRequest{
		Bolt11: payReq,
	}
	resp, err := bs.svc.SendPayment(sendPaymentRequest)
	if err != nil {
		return nil, err
	}
	var lnDetails breez_sdk.PaymentDetailsLn
	if resp.Payment.Details != nil {
		lnDetails, _ = resp.Payment.Details.(breez_sdk.PaymentDetailsLn)
	}
	return &lnclient.PayInvoiceResponse{
		Preimage: lnDetails.Data.PaymentPreimage,
	}, nil

}

func (bs *BreezService) SendKeysend(ctx context.Context, amount int64, destination, preimage string, custom_records []lnclient.TLVRecord) (preImage string, err error) {
	extraTlvs := []breez_sdk.TlvEntry{}
	for _, record := range custom_records {
		extraTlvs = append(extraTlvs, breez_sdk.TlvEntry{
			FieldNumber: record.Type,
			Value:       []uint8(record.Value),
		})
	}

	sendSpontaneousPaymentRequest := breez_sdk.SendSpontaneousPaymentRequest{
		NodeId:     destination,
		AmountMsat: uint64(amount),
		ExtraTlvs:  &extraTlvs,
	}
	resp, err := bs.svc.SendSpontaneousPayment(sendSpontaneousPaymentRequest)
	if err != nil {
		return "", err
	}
	var lnDetails breez_sdk.PaymentDetailsLn
	if resp.Payment.Details != nil {
		lnDetails, _ = resp.Payment.Details.(breez_sdk.PaymentDetailsLn)
	}
	return lnDetails.Data.PaymentPreimage, nil
}

func (bs *BreezService) GetBalance(ctx context.Context) (balance int64, err error) {
	info, err := bs.svc.NodeInfo()
	if err != nil {
		return 0, err
	}
	return int64(info.MaxPayableMsat), nil
}

func (bs *BreezService) MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64) (transaction *nip47.Transaction, err error) {
	expiry32 := uint32(expiry)
	receivePaymentRequest := breez_sdk.ReceivePaymentRequest{
		// amount provided in msat
		AmountMsat:  uint64(amount),
		Description: description,
		Expiry:      &expiry32,
	}
	resp, err := bs.svc.ReceivePayment(receivePaymentRequest)
	if err != nil {
		return nil, err
	}

	tx := &nip47.Transaction{
		Type:        "incoming",
		Invoice:     resp.LnInvoice.Bolt11,
		Preimage:    hex.EncodeToString(resp.LnInvoice.PaymentSecret),
		PaymentHash: resp.LnInvoice.PaymentHash,
		FeesPaid:    0,
		CreatedAt:   int64(resp.LnInvoice.Timestamp),
		SettledAt:   nil,
		Metadata:    nil,
	}

	if resp.LnInvoice.Description != nil {
		tx.Description = *resp.LnInvoice.Description
	}
	if resp.LnInvoice.DescriptionHash != nil {
		tx.DescriptionHash = *resp.LnInvoice.DescriptionHash
	}
	if resp.LnInvoice.AmountMsat != nil {
		tx.Amount = int64(*resp.LnInvoice.AmountMsat)
	}
	invExpiry := int64(resp.LnInvoice.Expiry)
	tx.ExpiresAt = &invExpiry
	return tx, nil
}

func (bs *BreezService) LookupInvoice(ctx context.Context, paymentHash string) (transaction *nip47.Transaction, err error) {
	log.Printf("p: %v", paymentHash)
	payment, err := bs.svc.PaymentByHash(paymentHash)
	if err != nil {
		return nil, err
	}
	if payment != nil {
		log.Printf("p: %v", payment)
		transaction, err = breezPaymentToTransaction(payment)
		if err != nil {
			return nil, err
		}
		return transaction, nil
	} else {
		return nil, errors.New("not found")
	}
}

func (bs *BreezService) ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaid bool, invoiceType string) (transactions []nip47.Transaction, err error) {

	request := breez_sdk.ListPaymentsRequest{}
	if limit > 0 {
		limit32 := uint32(limit)
		request.Limit = &limit32
	}
	if offset > 0 {
		offset32 := uint32(offset)
		request.Offset = &offset32
	}
	if from > 0 {
		from64 := int64(from)
		request.FromTimestamp = &from64
	}
	if until > 0 {
		until64 := int64(until)
		request.ToTimestamp = &until64
	}

	payments, err := bs.svc.ListPayments(request)
	if err != nil {
		return nil, err
	}

	transactions = []nip47.Transaction{}
	for _, payment := range payments {
		if payment.PaymentType != breez_sdk.PaymentTypeReceived && payment.PaymentType != breez_sdk.PaymentTypeSent {
			// skip other types of payments for now
			continue
		}

		transaction, err := breezPaymentToTransaction(&payment)
		if err == nil {
			transactions = append(transactions, *transaction)
		}
	}
	return transactions, nil
}

func (bs *BreezService) GetInfo(ctx context.Context) (info *lnclient.NodeInfo, err error) {
	return &lnclient.NodeInfo{
		Alias:       "breez",
		Color:       "",
		Pubkey:      "",
		Network:     "mainnet",
		BlockHeight: 0,
		BlockHash:   "",
	}, nil
}

func (bs *BreezService) ListChannels(ctx context.Context) ([]lnclient.Channel, error) {
	channels := []lnclient.Channel{}
	return channels, nil
}

func (bs *BreezService) GetNodeConnectionInfo(ctx context.Context) (nodeConnectionInfo *lnclient.NodeConnectionInfo, err error) {
	return &lnclient.NodeConnectionInfo{}, nil
}

func (bs *BreezService) ConnectPeer(ctx context.Context, connectPeerRequest *lnclient.ConnectPeerRequest) error {
	return nil
}
func (bs *BreezService) OpenChannel(ctx context.Context, openChannelRequest *lnclient.OpenChannelRequest) (*lnclient.OpenChannelResponse, error) {
	return nil, nil
}

func (bs *BreezService) CloseChannel(ctx context.Context, closeChannelRequest *lnclient.CloseChannelRequest) (*lnclient.CloseChannelResponse, error) {
	return nil, nil
}

func breezPaymentToTransaction(payment *breez_sdk.Payment) (*nip47.Transaction, error) {
	var lnDetails breez_sdk.PaymentDetailsLn
	if payment.Details != nil {
		lnDetails, _ = payment.Details.(breez_sdk.PaymentDetailsLn)
	}
	var txType string
	if payment.PaymentType == breez_sdk.PaymentTypeSent {
		txType = "outgoing"
	} else {
		txType = "incoming"
	}

	createdAt := payment.PaymentTime
	var expiresAt *int64
	description := lnDetails.Data.Label
	descriptionHash := ""

	if lnDetails.Data.Bolt11 != "" {
		// TODO: Breez should provide these details so we don't need to manually decode the invoice
		paymentRequest, err := decodepay.Decodepay(strings.ToLower(lnDetails.Data.Bolt11))
		if err != nil {
			log.Printf("Failed to decode bolt11 invoice: %v", payment)
			return nil, err
		}

		createdAt = int64(paymentRequest.CreatedAt)
		expiresAtUnix := time.UnixMilli(int64(paymentRequest.CreatedAt) * 1000).Add(time.Duration(paymentRequest.Expiry) * time.Second).Unix()
		expiresAt = &expiresAtUnix
		description = paymentRequest.Description
		descriptionHash = paymentRequest.DescriptionHash
	}

	tx := &nip47.Transaction{
		Type:            txType,
		Invoice:         lnDetails.Data.Bolt11,
		Preimage:        lnDetails.Data.PaymentPreimage,
		PaymentHash:     lnDetails.Data.PaymentHash,
		Amount:          int64(payment.AmountMsat),
		FeesPaid:        int64(payment.FeeMsat),
		CreatedAt:       createdAt,
		ExpiresAt:       expiresAt,
		Metadata:        nil,
		Description:     description,
		DescriptionHash: descriptionHash,
	}
	if payment.Status == breez_sdk.PaymentStatusComplete {
		settledAt := payment.PaymentTime
		tx.SettledAt = &settledAt
	}

	if payment.Description != nil {
		tx.Description = *payment.Description
	}

	return tx, nil
}

func (bs *BreezService) GetNewOnchainAddress(ctx context.Context) (string, error) {
	// The below code works but is not needed and there's no Breez UI to support it.
	// Plus, it creates complexity with the deposit limits.
	/*swapInfo, err := bs.svc.ReceiveOnchain(breez_sdk.ReceiveOnchainRequest{})

	if err != nil {
		bs.logger.Errorf("Failed to get onchain address: %v", err)
		return "", err
	}
	bs.logger.Infof("This address has deposit limits! Min: %d Max: %d", swapInfo.MinAllowedDeposit, swapInfo.MaxAllowedDeposit)

	return swapInfo.BitcoinAddress, nil*/
	return "", nil
}

func (bs *BreezService) GetOnchainBalance(ctx context.Context) (*lnclient.OnchainBalanceResponse, error) {
	response, err := bs.svc.NodeInfo()

	if err != nil {
		bs.logger.Errorf("Failed to get node info: %v", err)
		return nil, err
	}

	return &lnclient.OnchainBalanceResponse{
		Spendable: int64(response.OnchainBalanceMsat) / 1000,
		Total:     int64(response.OnchainBalanceMsat+response.PendingOnchainBalanceMsat) / 1000,
	}, nil
}

func (bs *BreezService) RedeemOnchainFunds(ctx context.Context, toAddress string) (txId string, err error) {
	if toAddress == "" {
		return "", errors.New("No address provided")
	}

	recommendedFees, err := bs.svc.RecommendedFees()
	if err != nil {
		bs.logger.Errorf("Failed to get recommended fees info: %v", err)
		return "", err
	}

	satPerVbyte := uint32(recommendedFees.FastestFee)
	prepareReq := breez_sdk.PrepareRedeemOnchainFundsRequest{SatPerVbyte: satPerVbyte, ToAddress: toAddress}
	prepareRedeemOnchainFundsResponse, err := bs.svc.PrepareRedeemOnchainFunds(prepareReq)

	if err != nil {
		bs.logger.Errorf("Failed to prepare onchain address: %v", err)
		return "", err
	}
	bs.logger.Infof("PrepareRedeemOnchainFunds response: %#v", prepareRedeemOnchainFundsResponse)

	redeemReq := breez_sdk.RedeemOnchainFundsRequest{SatPerVbyte: satPerVbyte, ToAddress: toAddress}
	redeemOnchainFundsResponse, err := bs.svc.RedeemOnchainFunds(redeemReq)

	if err != nil {
		bs.logger.Errorf("Failed to redeem onchain funds: %v", err)
		return "", err
	}

	bs.logger.Infof("RedeemOnchainFunds response: %#v", redeemOnchainFundsResponse)
	return hex.EncodeToString(redeemOnchainFundsResponse.Txid), nil
}

func (bs *BreezService) ResetRouter(key string) error {
	return nil
}

func (bs *BreezService) SendPaymentProbes(ctx context.Context, invoice string) error {
	return nil
}

func (bs *BreezService) SendSpontaneousPaymentProbes(ctx context.Context, amountMsat uint64, nodeId string) error {
	return nil
}

func (bs *BreezService) ListPeers(ctx context.Context) ([]lnclient.PeerDetails, error) {
	return nil, nil
}

func (bs *BreezService) GetLogOutput(ctx context.Context, maxLen int) ([]byte, error) {
	return []byte{}, nil
}

func (bs *BreezService) GetNodeStatus(ctx context.Context) (nodeStatus *lnclient.NodeStatus, err error) {
	return nil, nil
}

func (bs *BreezService) SignMessage(ctx context.Context, message string) (string, error) {
	resp, err := bs.svc.SignMessage(breez_sdk.SignMessageRequest{
		Message: message,
	})
	if err != nil {
		return "", err
	}

	return resp.Signature, nil
}

func (bs *BreezService) GetBalances(ctx context.Context) (*lnclient.BalancesResponse, error) {
	info, err := bs.svc.NodeInfo()
	if err != nil {
		return nil, err
	}

	return &lnclient.BalancesResponse{
		Onchain: lnclient.OnchainBalanceResponse{
			Spendable: int64(info.OnchainBalanceMsat) / 1000,
			Total:     int64(info.OnchainBalanceMsat+info.PendingOnchainBalanceMsat) / 1000,
		},
		Lightning: lnclient.LightningBalanceResponse{
			TotalSpendable:       int64(info.MaxPayableMsat),
			TotalReceivable:      int64(info.MaxReceivableMsat),
			NextMaxSpendable:     int64(info.MaxSinglePaymentAmountMsat),
			NextMaxReceivable:    int64(info.MaxReceivableMsat),
			NextMaxSpendableMPP:  int64(info.MaxSinglePaymentAmountMsat),
			NextMaxReceivableMPP: int64(info.MaxReceivableMsat),
		},
	}, nil
}

func (bs *BreezService) GetStorageDir() (string, error) {
	return "", nil
}

func (bs *BreezService) GetNetworkGraph(nodeIds []string) (lnclient.NetworkGraphResponse, error) {
	return nil, nil
}

func (bs *BreezService) UpdateLastWalletSyncRequest() {}

func (bs *BreezService) DisconnectPeer(ctx context.Context, peerId string) error {
	return nil
}
