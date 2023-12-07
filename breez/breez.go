package breez

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/breez/breez-sdk-go/breez_sdk"
)

type BreezService struct {
	listener *BreezListener
	svc      *breez_sdk.BlockingBreezServices
}
type BreezListener struct{}

func (BreezListener) OnEvent(e breez_sdk.BreezEvent) {
	log.Printf("received event %#v", e)
}

func NewBreezService(mnemonic, apiKey, inviteCode, workDir string) (result *BreezService, err error) {
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

func (bs *BreezService) GetBalance(ctx context.Context, senderPubkey string) (balance int64, err error) {
	info, err := bs.svc.NodeInfo()
	if err != nil {
		return 0, err
	}
	return int64(info.ChannelsBalanceMsat) / 1000, nil
}

func (bs *BreezService) MakeInvoice(ctx context.Context, senderPubkey string, amount int64, description string, descriptionHash string, expiry int64) (invoice string, paymentHash string, err error) {
	receivePaymentRequest := breez_sdk.ReceivePaymentRequest{
		// amount provided in msat
		AmountMsat:  uint64(amount),
		Description: description,
	}
	resp, err := bs.svc.ReceivePayment(receivePaymentRequest)
	if err != nil {
		return "", "", err
	}
	return resp.LnInvoice.Bolt11, resp.LnInvoice.PaymentHash, nil
}

func (bs *BreezService) LookupInvoice(ctx context.Context, senderPubkey string, paymentHash string) (invoice string, paid bool, err error) {
	return "", false, errors.New("Not implemented")
}
