package breez

import (
	"context"
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
	return &BreezService{
		listener: &listener,
		svc:      svc,
	}, nil
}

func (bs *BreezService) SendPaymentSync(ctx context.Context, senderPubkey string, payReq string) (preimage string, err error) {
	resp, err := bs.svc.SendPayment(payReq, nil)
	if err != nil {
		return "", err
	}
	var lnDetails breez_sdk.PaymentDetailsLn
	if resp.Details != nil {
		lnDetails, _ = resp.Details.(breez_sdk.PaymentDetailsLn)
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
	resp, err := bs.svc.ReceivePayment(uint64(amount), description)
	if err != nil {
		return "", "", err
	}
	return resp.Bolt11, resp.PaymentHash, nil
}
