package breez

import (
	"context"
	"log"

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
	return "", nil

}

func (bs *BreezService) GetBalance(ctx context.Context, senderPubkey string) (balance int64, err error) {
	return 0, nil
}

func (bs *BreezService) MakeInvoice(ctx context.Context, senderPubkey string, amount int64, description string, descriptionHash string, expiry int64) (invoice string, paymentHash string, err error) {
	return "", "", nil
}
