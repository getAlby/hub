package main

import (
	"context"
	"encoding/hex"
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/breez/breez-sdk-go/breez_sdk"
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

func NewBreezService(nwcSvc *Service, mnemonic, apiKey, inviteCode, workDir string) (result LNClient, err error) {
	// FIXME: split single and multi user app
	//add default user to db
	user := &User{}
	err = nwcSvc.db.FirstOrInit(user, User{AlbyIdentifier: "breez"}).Error
	if err != nil {
		return nil, err
	}
	err = nwcSvc.db.Save(user).Error
	if err != nil {
		return nil, err
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

func (bs *BreezService) SendPaymentSync(ctx context.Context, senderPubkey string, payReq string) (preimage string, err error) {
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

func (bs *BreezService) SendKeysend(ctx context.Context, senderPubkey string, amount int64, destination, preimage string, custom_records []TLVRecord) (preImage string, err error) {
	sendSpontaneousPaymentRequest := breez_sdk.SendSpontaneousPaymentRequest{
		NodeId:     destination,
		AmountMsat: uint64(amount),
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

func (bs *BreezService) GetBalance(ctx context.Context, senderPubkey string) (balance int64, err error) {
	info, err := bs.svc.NodeInfo()
	if err != nil {
		return 0, err
	}
	return int64(info.ChannelsBalanceMsat) / 1000, nil
}

func (bs *BreezService) MakeInvoice(ctx context.Context, senderPubkey string, amount int64, description string, descriptionHash string, expiry int64) (transaction *Nip47Transaction, err error) {
	if expiry == 0 {
		expiry = 60 * 60 * 24
	}
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

func (bs *BreezService) LookupInvoice(ctx context.Context, senderPubkey string, paymentHash string) (transaction *Nip47Transaction, err error) {
	log.Printf("p: %v", paymentHash)
	payment, err := bs.svc.PaymentByHash(paymentHash)
	if err != nil {
		return nil, err
	}
	if payment != nil {
		log.Printf("p: %v", payment)
		transaction = breezPaymentToTransaction(payment)
		return transaction, nil
	} else {
		return nil, errors.New("not found")
	}
}

func (bs *BreezService) ListTransactions(ctx context.Context, senderPubkey string, from, until, limit, offset uint64, unpaid bool, invoiceType string) (transactions []Nip47Transaction, err error) {
	payments, err := bs.svc.ListPayments(breez_sdk.ListPaymentsRequest{})
	if err != nil {
		return nil, err
	}

	transactions = []Nip47Transaction{}
	for _, payment := range payments {
		transaction := breezPaymentToTransaction(&payment)

		transactions = append(transactions, *transaction)
	}
	return transactions, nil
}

func (bs *BreezService) GetInfo(ctx context.Context, senderPubkey string) (info *NodeInfo, err error) {
	return &NodeInfo{
		Alias:       "breez",
		Color:       "",
		Pubkey:      "",
		Network:     "mainnet",
		BlockHeight: 0,
		BlockHash:   "",
	}, nil
}

func breezPaymentToTransaction(payment *breez_sdk.Payment) *Nip47Transaction {
	var lnDetails breez_sdk.PaymentDetailsLn
	if payment.Details != nil {
		lnDetails, _ = payment.Details.(breez_sdk.PaymentDetailsLn)
	}
	var txType string
	if payment.PaymentType == breez_sdk.PaymentTypeSent {
		txType = "outgoing"
	} else if payment.PaymentType == breez_sdk.PaymentTypeSent {
		txType = "incoming"
	}

	tx := &Nip47Transaction{
		Type:        txType,
		Invoice:     lnDetails.Data.Bolt11,
		Preimage:    lnDetails.Data.PaymentPreimage,
		PaymentHash: lnDetails.Data.PaymentHash,
		Amount:      int64(payment.AmountMsat),
		FeesPaid:    int64(payment.FeeMsat),
		CreatedAt:   time.Now().Unix(),
		ExpiresAt:   nil,
		Metadata:    nil,
	}
	if payment.Status == breez_sdk.PaymentStatusComplete {
		settledAt := payment.PaymentTime
		tx.SettledAt = &settledAt
	}

	if payment.Description != nil {
		tx.Description = *payment.Description
	}

	return tx
}
