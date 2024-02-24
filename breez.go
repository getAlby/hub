package main

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
	"github.com/getAlby/nostr-wallet-connect/models/lnclient"
	decodepay "github.com/nbd-wtf/ln-decodepay"
)

type BreezService struct {
	listener *BreezListener
	svc      *breez_sdk.BlockingBreezServices
}
type BreezListener struct{}

func (BreezListener) Log(l breez_sdk.LogEntry) {
	if l.Level != "TRACE" {
		log.Printf("%v\n", l.Line)
	}
}

func (BreezListener) OnEvent(e breez_sdk.BreezEvent) {
	log.Printf("received event %#v", e)
}

func NewBreezService(mnemonic, apiKey, inviteCode, workDir string) (result lnclient.LNClient, err error) {
	if mnemonic == "" || apiKey == "" || inviteCode == "" || workDir == "" {
		return nil, errors.New("one or more required breez configuration are missing")
	}

	//create dir if not exists
	newpath := filepath.Join(".", workDir)
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
	listener := BreezListener{}
	config := breez_sdk.DefaultConfig(breez_sdk.EnvironmentTypeProduction, apiKey, nodeConfig)
	config.WorkingDir = workDir
	// breez_sdk.SetLogStream(listener)
	svc, err := breez_sdk.Connect(config, seed, listener)
	if err != nil {
		return nil, err
	}
	healthCheck, err := svc.ServiceHealthCheck()
	if err != nil {
		return nil, err
	}
	if err == nil {
		log.Printf("Current service status is: %v", healthCheck.Status)
	}

	nodeInfo, err := svc.NodeInfo()
	if err != nil {
		return nil, err
	}
	if err == nil {
		log.Printf("Node info: %v", nodeInfo)
		log.Printf("ln balance: %v - onchain balance: %v - max_payable_msat: %v - max_receivable_msat: %v - max_single_payment_amount_msat: %v - connected_peers: %v - inbound_liquidity_msats: %v", nodeInfo.ChannelsBalanceMsat, nodeInfo.OnchainBalanceMsat, nodeInfo.MaxPayableMsat, nodeInfo.MaxReceivableMsat, nodeInfo.MaxSinglePaymentAmountMsat, nodeInfo.ConnectedPeers, nodeInfo.InboundLiquidityMsats)
	}

	return &BreezService{
		listener: &listener,
		svc:      svc,
	}, nil
}

func (bs *BreezService) Shutdown() error {
	return bs.svc.Disconnect()
}

func (bs *BreezService) SendPaymentSync(ctx context.Context, payReq string) (preimage string, err error) {
	sendPaymentRequest := breez_sdk.SendPaymentRequest{
		Bolt11: payReq,
	}
	resp, err := bs.svc.SendPayment(sendPaymentRequest)
	if err != nil {
		return "", err
	}
	var lnDetails breez_sdk.PaymentDetailsLn
	if resp.Payment.Details != nil {
		lnDetails, _ = resp.Payment.Details.(breez_sdk.PaymentDetailsLn)
	}
	return lnDetails.Data.PaymentPreimage, nil

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
	return int64(info.ChannelsBalanceMsat), nil
}

func (bs *BreezService) MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64) (transaction *Nip47Transaction, err error) {
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

	tx := &Nip47Transaction{
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

func (bs *BreezService) LookupInvoice(ctx context.Context, paymentHash string) (transaction *Nip47Transaction, err error) {
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

func (bs *BreezService) ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaid bool, invoiceType string) (transactions []Nip47Transaction, err error) {

	request := breez_sdk.ListPaymentsRequest{}
	if limit == 0 {
		// make sure a sensible limit is passed
		limit = 100
	}
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

	transactions = []Nip47Transaction{}
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

func breezPaymentToTransaction(payment *breez_sdk.Payment) (*Nip47Transaction, error) {
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

	tx := &Nip47Transaction{
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
	return "", nil
}

func (bs *BreezService) GetOnchainBalance(ctx context.Context) (int64, error) {
	return 0, nil
}
