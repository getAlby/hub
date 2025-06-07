package bark

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"

	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	bindings "github.com/getAlby/hub/second_ark"
)

const barkDB = "bark.sqlite"
const vtxoRefreshInterval = 1 * time.Hour

const nodeCommandPubkey = "pubkey"
const nodeCommandMaintenance = "maintenance"

type BarkService struct {
	wallet *bindings.Wallet
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

func NewBarkService(ctx context.Context, mnemonic, workdir string) (*BarkService, error) {
	err := os.MkdirAll(workdir, 0755)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to create Bark working dir")
		return nil, err
	}

	var wallet *bindings.Wallet

	dbFilePath := filepath.Join(workdir, barkDB)
	if _, err := os.Stat(dbFilePath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			logger.Logger.WithError(err).Error("Failed to check Bark database file")
			return nil, err
		}

		barkConfig := bindings.Config{
			Network:              "signet",
			Birthday:             nil,
			AspAddress:           "https://ark.signet.2nd.dev",
			EsploraAddress:       "https://esplora.signet.2nd.dev",
			VtxoRefreshThreshold: 0,
		}

		wallet, err = checkBindingsErr(bindings.CreateWallet(dbFilePath, mnemonic, barkConfig))
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to create Bark wallet")
			return nil, err
		}
	} else {
		wallet, err = checkBindingsErr(bindings.OpenWallet(dbFilePath, mnemonic))
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to open Bark wallet")
			return nil, err
		}
	}

	logger.Logger.Info("Performing wallet maintenance")
	if err := wallet.Maintenance().AsError(); err != nil {
		logger.Logger.WithError(err).Error("Failed to perform wallet maintenance")
		return nil, err
	}

	pk := wallet.OorPubkey()
	logger.Logger.Info("Ark public key: ", pk)

	cctx, cancel := context.WithCancel(ctx)
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(vtxoRefreshInterval)
		defer ticker.Stop()

		// Perform initial refresh on startup.
		logger.Logger.Info("Refreshing vtxos")
		if err := wallet.RefreshAll().AsError(); err != nil {
			logger.Logger.WithError(err).Error("Failed to refresh vtxos")
		}

		for {
			select {
			case <-ticker.C:
				logger.Logger.Info("Refreshing vtxos")
				if err := wallet.RefreshAll().AsError(); err != nil {
					logger.Logger.WithError(err).Error("Failed to refresh vtxos")
				}
			case <-cctx.Done():
				return
			}
		}
	}()

	return &BarkService{
		wallet: wallet,
		cancel: cancel,
	}, nil
}

func (s *BarkService) Shutdown() error {
	logger.Logger.Info("Shutting down Ark client")

	s.cancel()

	logger.Logger.Info("Waiting for Ark procs to finish")
	s.wg.Wait()

	logger.Logger.Info("Ark shutdown complete")

	return nil
}

func (s *BarkService) SendPaymentSync(ctx context.Context, invoice string, amount *uint64, timeoutSeconds *int64) (*lnclient.PayInvoiceResponse, error) {
	paymentRequest, err := decodepay.Decodepay(invoice)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"bolt11": invoice,
		}).WithError(err).Error("Failed to decode bolt11 invoice")

		return nil, err
	}

	// Bark won't allow setting amount on invoices that already have an amount set.
	// Bark also does not support msats, offering rounding up or down to the nearest sat.
	// FIXME: Rounding up here to avoid underpayment; however, this may lead to unexpected spending.
	var customAmount *uint64
	if paymentRequest.MSatoshi == 0 && amount != nil {
		customAmountSat := (*amount + 999) / 1000
		customAmount = &customAmountSat
	}

	preimage, err := checkBindingsErr(s.wallet.PayBolt11(invoice, customAmount))
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"bolt11": invoice,
			"amount": amount,
		}).WithError(err).Error("Failed to pay bolt11 invoice")

		return nil, err
	}

	logger.Logger.WithFields(logrus.Fields{
		"bolt11":   invoice,
		"amount":   amount,
		"preimage": preimage,
	}).Info("Successfully paid bolt11 invoice")

	return &lnclient.PayInvoiceResponse{
		Preimage: preimage,
		Fee:      0,
	}, nil
}

func (s *BarkService) SendKeysend(ctx context.Context, amount uint64, destination string, customRecords []lnclient.TLVRecord, preimage string) (*lnclient.PayKeysendResponse, error) {
	return nil, errors.New("not implemented")
}

func (s *BarkService) GetPubkey() string {
	return s.wallet.OorPubkey()
}

func (s *BarkService) GetInfo(ctx context.Context) (info *lnclient.NodeInfo, err error) {
	pk := s.wallet.OorPubkey()
	arkInfo, err := checkBindingsErr(s.wallet.ArkInfo())
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to get Ark info")
		return nil, err
	}

	return &lnclient.NodeInfo{
		Pubkey:  pk,
		Network: arkInfo.Network,
	}, nil
}

func (s *BarkService) MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64) (transaction *lnclient.Transaction, err error) {
	return nil, errors.New("not implemented")
}

func (s *BarkService) MakeHoldInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64, paymentHash string) (transaction *lnclient.Transaction, err error) {
	return nil, errors.New("not implemented")
}

func (s *BarkService) SettleHoldInvoice(ctx context.Context, preimage string) (err error) {
	return errors.New("not implemented")
}

func (s *BarkService) CancelHoldInvoice(ctx context.Context, paymentHash string) (err error) {
	return errors.New("not implemented")
}

func (s *BarkService) LookupInvoice(ctx context.Context, paymentHash string) (transaction *lnclient.Transaction, err error) {
	return nil, errors.New("not implemented")
}

func (s *BarkService) ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaid bool, invoiceType string) (transactions []lnclient.Transaction, err error) {
	return nil, errors.New("not implemented")
}

func (s *BarkService) ListOnchainTransactions(ctx context.Context) ([]lnclient.OnchainTransaction, error) {
	return nil, errors.New("not implemented")
}

func (s *BarkService) ListChannels(ctx context.Context) (channels []lnclient.Channel, err error) {
	return []lnclient.Channel{}, nil
}

func (s *BarkService) GetNodeConnectionInfo(ctx context.Context) (nodeConnectionInfo *lnclient.NodeConnectionInfo, err error) {
	return &lnclient.NodeConnectionInfo{
		Pubkey: s.wallet.OorPubkey(),
	}, nil
}

func (s *BarkService) GetNodeStatus(ctx context.Context) (nodeStatus *lnclient.NodeStatus, err error) {
	return &lnclient.NodeStatus{
		IsReady: true,
	}, nil
}

func (s *BarkService) ConnectPeer(ctx context.Context, connectPeerRequest *lnclient.ConnectPeerRequest) error {
	return errors.New("not implemented")
}

func (s *BarkService) OpenChannel(ctx context.Context, openChannelRequest *lnclient.OpenChannelRequest) (*lnclient.OpenChannelResponse, error) {
	return nil, errors.New("not implemented")
}

func (s *BarkService) CloseChannel(ctx context.Context, closeChannelRequest *lnclient.CloseChannelRequest) (*lnclient.CloseChannelResponse, error) {
	return nil, errors.New("not implemented")
}

func (s *BarkService) UpdateChannel(ctx context.Context, updateChannelRequest *lnclient.UpdateChannelRequest) error {
	return errors.New("not implemented")
}

func (s *BarkService) DisconnectPeer(ctx context.Context, peerId string) error {
	return errors.New("not implemented")
}

func (s *BarkService) GetNewOnchainAddress(ctx context.Context) (string, error) {
	return "", errors.New("not implemented")
}

func (s *BarkService) ResetRouter(key string) error {
	return errors.New("not implemented")
}

func (s *BarkService) GetOnchainBalance(ctx context.Context) (*lnclient.OnchainBalanceResponse, error) {
	balance, err := checkBindingsErr(s.wallet.Balance())
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to get Bark wallet balance")
		return nil, err
	}

	return &lnclient.OnchainBalanceResponse{
		Spendable:                          int64(balance.OnchainSat * 1000),
		PendingBalancesFromChannelClosures: balance.PendingExitSat * 1000,
	}, nil
}

func (s *BarkService) GetBalances(ctx context.Context, includeInactiveChannels bool) (*lnclient.BalancesResponse, error) {
	balance, err := checkBindingsErr(s.wallet.Balance())
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to get Bark wallet balance")
		return nil, err
	}

	return &lnclient.BalancesResponse{
		Onchain: lnclient.OnchainBalanceResponse{
			Spendable:                          int64(balance.OnchainSat * 1000),
			PendingBalancesFromChannelClosures: balance.PendingExitSat * 1000,
		},
		Lightning: lnclient.LightningBalanceResponse{
			TotalSpendable: int64(balance.OffchainSat * 1000),
		},
	}, nil
}

func (s *BarkService) RedeemOnchainFunds(ctx context.Context, toAddress string, amount uint64, sendAll bool) (txId string, err error) {
	return "", errors.New("not implemented")
}

func (s *BarkService) SendPaymentProbes(ctx context.Context, invoice string) error {
	return errors.New("not implemented")
}

func (s *BarkService) SendSpontaneousPaymentProbes(ctx context.Context, amountMsat uint64, nodeId string) error {
	return errors.New("not implemented")
}

func (s *BarkService) ListPeers(ctx context.Context) ([]lnclient.PeerDetails, error) {
	return nil, errors.New("not implemented")
}

func (s *BarkService) GetLogOutput(ctx context.Context, maxLen int) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (s *BarkService) SignMessage(ctx context.Context, message string) (string, error) {
	return "", errors.New("not implemented")
}

func (s *BarkService) GetStorageDir() (string, error) {
	return "", errors.New("not implemented")
}

func (s *BarkService) GetNetworkGraph(ctx context.Context, nodeIds []string) (lnclient.NetworkGraphResponse, error) {
	return nil, errors.New("not implemented")
}

func (s *BarkService) UpdateLastWalletSyncRequest() {
}

func (s *BarkService) GetSupportedNIP47Methods() []string {
	return []string{"pay_invoice", "get_balance", "get_info"}
}

func (s *BarkService) GetSupportedNIP47NotificationTypes() []string {
	return []string{}
}

func (s *BarkService) GetCustomNodeCommandDefinitions() []lnclient.CustomNodeCommandDef {
	return []lnclient.CustomNodeCommandDef{
		{
			Name:        nodeCommandPubkey,
			Description: "Get Arc pubkey of the wallet.",
			Args:        nil,
		},
		{
			Name:        nodeCommandMaintenance,
			Description: "Run Bark wallet maintenance.",
			Args:        nil,
		},
	}
}

func (s *BarkService) ExecuteCustomNodeCommand(ctx context.Context, command *lnclient.CustomNodeCommandRequest) (*lnclient.CustomNodeCommandResponse, error) {
	switch command.Name {
	case nodeCommandPubkey:
		return &lnclient.CustomNodeCommandResponse{
			Response: map[string]interface{}{
				"pubkey": s.wallet.OorPubkey(),
			},
		}, nil
	case nodeCommandMaintenance:
		if err := s.wallet.Maintenance().AsError(); err != nil {
			logger.Logger.WithError(err).Error("Failed to perform wallet maintenance")
			return nil, err
		}
		return lnclient.NewCustomNodeCommandResponseEmpty(), nil
	}

	return nil, lnclient.ErrUnknownCustomNodeCommand
}

func checkBindingsErr[T any](val T, err error) (T, error) {
	if e, ok := err.(bindings.NativeError); ok {
		return val, e.AsError()
	}
	return val, err
}
