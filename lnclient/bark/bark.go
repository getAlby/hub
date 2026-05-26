package bark

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
	bark "gitlab.com/ark-bitcoin/bark-ffi-bindings/golang/bark"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
)

const (
	signetServerAddress  = "https://ark.signet.2nd.dev"
	signetEsploraAddress = "https://esplora.signet.2nd.dev"
	claimInterval        = 5 * time.Second
)

type BarkService struct {
	wallet         *bark.Wallet
	workDir        string
	eventPublisher events.EventPublisher
	pubkey         string
	cancelFn       context.CancelFunc
	loopWg         sync.WaitGroup
}

func NewBarkService(ctx context.Context, eventPublisher events.EventPublisher, workDir, mnemonic string) (lnclient.LNClient, error) {
	if mnemonic == "" {
		return nil, errors.New("no mnemonic configured")
	}
	if workDir == "" {
		return nil, errors.New("no bark work directory configured")
	}

	esplora := signetEsploraAddress
	cfg := bark.Config{
		ServerAddress:  signetServerAddress,
		EsploraAddress: &esplora,
		Network:        bark.NetworkSignet,
	}

	_, statErr := os.Stat(workDir)
	isFirstSetup := statErr != nil && errors.Is(statErr, os.ErrNotExist)

	logger.Logger.WithFields(logrus.Fields{
		"workDir":      workDir,
		"isFirstSetup": isFirstSetup,
	}).Info("Opening Bark wallet")

	var wallet *bark.Wallet
	var err error
	if isFirstSetup {
		wallet, err = bark.WalletCreate(mnemonic, cfg, workDir, false)
	} else {
		wallet, err = bark.WalletOpen(mnemonic, cfg, workDir)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open bark wallet: %w", err)
	}

	// Bark provides a built-in background daemon that periodically syncs
	// with the Ark server and blockchain. Without onchain wallet support
	// (we don't expose onchain in this minimum integration) it still runs
	// the lightning-relevant tasks.
	if err := wallet.RunDaemon(nil); err != nil {
		logger.Logger.WithError(err).Warn("Bark daemon failed to start")
	}

	loopCtx, cancelFn := context.WithCancel(context.Background())
	bs := &BarkService{
		wallet:         wallet,
		workDir:        workDir,
		eventPublisher: eventPublisher,
		pubkey:         wallet.Fingerprint(),
		cancelFn:       cancelFn,
	}

	// Run maintenance immediately on startup so a wallet that was briefly
	// offline refreshes any VTXOs that drifted towards expiry before they are
	// swept by the server. This is fire-and-forget as it may join an Ark round
	// and take some time.
	go func() {
		if err := bs.wallet.Maintenance(); err != nil {
			logger.Logger.WithError(err).Warn("Bark startup maintenance failed")
		}
	}()

	bs.loopWg.Add(1)
	go bs.runClaimLoop(loopCtx)

	return bs, nil
}

// runClaimLoop periodically claims any pending Lightning receives. Bark's
// Lightning receives don't auto-credit — once an HTLC is received the
// preimage must be revealed via TryClaim*, which moves the funds into
// the spendable balance.
func (bs *BarkService) runClaimLoop(ctx context.Context) {
	defer bs.loopWg.Done()
	ticker := time.NewTicker(claimInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			claimed, err := bs.wallet.TryClaimAllLightningReceives(false)
			if err != nil {
				logger.Logger.WithError(err).Debug("Bark TryClaimAllLightningReceives failed")
				continue
			}
			for i := range claimed {
				lr := claimed[i]
				if !lr.PreimageRevealed {
					continue
				}
				tx, txErr := bs.lightningReceiveToTransaction(&lr)
				if txErr != nil {
					logger.Logger.WithError(txErr).WithField("paymentHash", lr.PaymentHash).Warn("Failed to convert claimed Bark receive to transaction")
					continue
				}
				logger.Logger.WithFields(logrus.Fields{
					"paymentHash": lr.PaymentHash,
					"amountSats":  lr.AmountSats,
				}).Info("Bark lightning receive claimed")
				bs.eventPublisher.Publish(&events.Event{
					Event:      "nwc_lnclient_payment_received",
					Properties: tx,
				})
			}
		}
	}
}

func (bs *BarkService) MakeInvoice(ctx context.Context, amountMsat int64, description string, descriptionHash string, expiry int64, throughNodePubkey *string) (*lnclient.Transaction, error) {
	if amountMsat <= 0 {
		return nil, errors.New("0-amount invoices not supported")
	}
	if amountMsat%1000 != 0 {
		return nil, errors.New("amount must be a whole number of sats")
	}

	var desc *string
	if description != "" {
		desc = &description
	}

	invoice, err := bs.wallet.Bolt11Invoice(uint64(amountMsat/1000), desc)
	if err != nil {
		return nil, fmt.Errorf("bark Bolt11Invoice failed: %w", err)
	}

	paymentRequest, err := decodepay.Decodepay(invoice.Invoice)
	if err != nil {
		logger.Logger.WithError(err).WithField("bolt11", invoice.Invoice).Error("Failed to decode bark-generated bolt11 invoice")
		return nil, err
	}

	expiresAtUnix := time.UnixMilli(int64(paymentRequest.CreatedAt) * 1000).Add(time.Duration(paymentRequest.Expiry) * time.Second).Unix()

	// The preimage is generated alongside the invoice but is not returned by
	// Bolt11Invoice. Fetch it via the receive status so consumers can rely on
	// lookup_invoice exposing the real preimage.
	var preimage string
	receive, err := bs.wallet.LightningReceiveStatus(paymentRequest.PaymentHash)
	if err != nil {
		logger.Logger.WithError(err).WithField("paymentHash", paymentRequest.PaymentHash).Error("Failed to fetch bark receive status for preimage")
		return nil, err
	}
	preimage = receive.PaymentPreimage
	if preimage == "" {
		return nil, errors.New("no preimage available")
	}

	return &lnclient.Transaction{
		Type:            constants.TRANSACTION_TYPE_INCOMING,
		Invoice:         invoice.Invoice,
		Preimage:        preimage,
		PaymentHash:     paymentRequest.PaymentHash,
		AmountMsat:      amountMsat,
		CreatedAt:       int64(paymentRequest.CreatedAt),
		ExpiresAt:       &expiresAtUnix,
		Description:     paymentRequest.Description,
		DescriptionHash: paymentRequest.DescriptionHash,
	}, nil
}

func (bs *BarkService) SendPaymentSync(invoice string, amountMsat *uint64) (*lnclient.PayInvoiceResponse, error) {
	// 0-amount invoices not supported initially — keeps the surface minimal.
	if amountMsat != nil {
		return nil, errors.New("0-amount invoices not supported")
	}

	send, err := bs.wallet.PayLightningInvoice(invoice, nil)
	if err != nil {
		return nil, fmt.Errorf("bark PayLightningInvoice failed: %w", err)
	}

	preimage := ""
	if send.Preimage != nil {
		preimage = *send.Preimage
	}

	// PayLightningInvoice can return before the payment settles; in that case
	// LightningSend.Preimage is nil. Wait for completion via CheckLightningPayment
	// with wait=true to fetch the real preimage.
	if preimage == "" {
		paymentRequest, decodeErr := decodepay.Decodepay(invoice)
		if decodeErr != nil {
			return nil, fmt.Errorf("failed to decode invoice while waiting for preimage: %w", decodeErr)
		}
		preimagePtr, checkErr := bs.wallet.CheckLightningPayment(paymentRequest.PaymentHash, true)
		if checkErr != nil {
			return nil, fmt.Errorf("bark CheckLightningPayment failed: %w", checkErr)
		}
		if preimagePtr == nil || *preimagePtr == "" {
			return nil, errors.New("bark payment did not complete")
		}
		preimage = *preimagePtr
	}

	// Bark's LightningSend struct does not expose a fee field. Lightning routing
	// fees inside Ark are absorbed into the Ark protocol economics and not
	// reported per-payment here. Report zero fee for now.
	return &lnclient.PayInvoiceResponse{
		Preimage: preimage,
		FeeMsat:  0,
	}, nil
}

func (bs *BarkService) LookupInvoice(ctx context.Context, paymentHash string) (*lnclient.Transaction, error) {
	receive, err := bs.wallet.LightningReceiveStatus(paymentHash)
	// Try as an incoming receive first.
	if err == nil && receive != nil {
		return bs.lightningReceiveToTransaction(receive)
	}

	// Fall back to outgoing payment lookup.
	preimagePtr, err := bs.wallet.CheckLightningPayment(paymentHash, false)
	if err != nil {
		return nil, fmt.Errorf("bark CheckLightningPayment failed: %w", err)
	}

	tx := &lnclient.Transaction{
		Type:        constants.TRANSACTION_TYPE_OUTGOING,
		PaymentHash: paymentHash,
	}
	if preimagePtr != nil && *preimagePtr != "" {
		tx.Preimage = *preimagePtr
		now := time.Now().Unix()
		tx.SettledAt = &now
	}
	return tx, nil
}

func (bs *BarkService) lightningReceiveToTransaction(receive *bark.LightningReceive) (*lnclient.Transaction, error) {
	paymentRequest, err := decodepay.Decodepay(receive.Invoice)
	if err != nil {
		return nil, err
	}
	expiresAtUnix := time.UnixMilli(int64(paymentRequest.CreatedAt) * 1000).Add(time.Duration(paymentRequest.Expiry) * time.Second).Unix()

	tx := &lnclient.Transaction{
		Type:            constants.TRANSACTION_TYPE_INCOMING,
		Invoice:         receive.Invoice,
		PaymentHash:     receive.PaymentHash,
		AmountMsat:      paymentRequest.MSatoshi,
		CreatedAt:       int64(paymentRequest.CreatedAt),
		ExpiresAt:       &expiresAtUnix,
		Description:     paymentRequest.Description,
		DescriptionHash: paymentRequest.DescriptionHash,
	}
	if receive.PreimageRevealed {
		now := time.Now().Unix()
		tx.SettledAt = &now
		tx.Preimage = receive.PaymentPreimage
	}
	return tx, nil
}

func (bs *BarkService) GetBalances(ctx context.Context, includeInactiveChannels bool) (*lnclient.BalancesResponse, error) {
	// Balance is computed from local state, which goes stale while the wallet is
	// closed. Sync with the Ark server first so we report the current balance
	// rather than a stale (often zero) snapshot. This is the documented pattern:
	// always Sync() before Balance().
	if err := bs.wallet.Sync(); err != nil {
		logger.Logger.WithError(err).Warn("Bark sync failed before reading balance")
	}

	balance, err := bs.wallet.Balance()
	if err != nil {
		return nil, err
	}
	spendableMsat := int64(balance.SpendableSats) * 1000

	return &lnclient.BalancesResponse{
		Onchain: lnclient.OnchainBalanceResponse{
			PendingBalancesDetails:      []lnclient.PendingBalanceDetails{},
			PendingSweepBalancesDetails: []lnclient.PendingBalanceDetails{},
		},
		Lightning: lnclient.LightningBalanceResponse{
			TotalSpendable:          spendableMsat,
			TotalSpendableSat:       spendableMsat / 1000,
			TotalSpendableMsat:      spendableMsat,
			NextMaxSpendable:        spendableMsat,
			NextMaxSpendableSat:     spendableMsat / 1000,
			NextMaxSpendableMsat:    spendableMsat,
			NextMaxSpendableMPP:     spendableMsat,
			NextMaxSpendableMPPSat:  spendableMsat / 1000,
			NextMaxSpendableMPPMsat: spendableMsat,
		},
	}, nil
}

func (bs *BarkService) GetInfo(ctx context.Context) (*lnclient.NodeInfo, error) {
	return &lnclient.NodeInfo{
		Alias:   "Bark",
		Color:   "#897FFF",
		Pubkey:  bs.pubkey,
		Network: "signet",
	}, nil
}

func (bs *BarkService) GetNodeStatus(ctx context.Context) (*lnclient.NodeStatus, error) {
	return &lnclient.NodeStatus{
		IsReady: true,
	}, nil
}

func (bs *BarkService) GetNodeConnectionInfo(ctx context.Context) (*lnclient.NodeConnectionInfo, error) {
	return &lnclient.NodeConnectionInfo{
		Pubkey: bs.pubkey,
	}, nil
}

func (bs *BarkService) GetPubkey() string {
	return bs.pubkey
}

func (bs *BarkService) GetSupportedNIP47Methods() []string {
	return []string{"pay_invoice", "get_balance", "get_budget", "get_info", "make_invoice", "lookup_invoice", "list_transactions", "multi_pay_invoice"}
}

func (bs *BarkService) GetSupportedNIP47NotificationTypes() []string {
	return []string{}
}

func (bs *BarkService) Shutdown() error {
	if bs.cancelFn != nil {
		bs.cancelFn()
		done := make(chan struct{})
		go func() {
			bs.loopWg.Wait()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(claimInterval + 10*time.Second):
			logger.Logger.Warn("Timed out waiting for Bark background loops to stop")
		}
	}
	if err := bs.wallet.StopDaemon(); err != nil {
		logger.Logger.WithError(err).Warn("Bark StopDaemon failed")
	}
	bs.wallet.Destroy()
	return nil
}

// --- unsupported / stubbed methods ---

func (bs *BarkService) SendKeysend(amountMsat uint64, destination string, customRecords []lnclient.TLVRecord, preimage string) (*lnclient.PayKeysendResponse, error) {
	return nil, errors.New("keysend not supported")
}

func (bs *BarkService) MakeHoldInvoice(ctx context.Context, amountMsat int64, description string, descriptionHash string, expiry int64, paymentHash string, minCltvExpiryDelta *uint64) (*lnclient.Transaction, error) {
	return nil, errors.New("not implemented")
}

func (bs *BarkService) SettleHoldInvoice(ctx context.Context, preimage string) error {
	return errors.New("not implemented")
}

func (bs *BarkService) CancelHoldInvoice(ctx context.Context, paymentHash string) error {
	return errors.New("not implemented")
}

func (bs *BarkService) ListChannels(ctx context.Context) ([]lnclient.Channel, error) {
	return []lnclient.Channel{}, nil
}

func (bs *BarkService) ConnectPeer(ctx context.Context, connectPeerRequest *lnclient.ConnectPeerRequest) error {
	return nil
}

func (bs *BarkService) OpenChannel(ctx context.Context, openChannelRequest *lnclient.OpenChannelRequest) (*lnclient.OpenChannelResponse, error) {
	return nil, nil
}

func (bs *BarkService) CloseChannel(ctx context.Context, closeChannelRequest *lnclient.CloseChannelRequest) (*lnclient.CloseChannelResponse, error) {
	return nil, nil
}

func (bs *BarkService) UpdateChannel(ctx context.Context, updateChannelRequest *lnclient.UpdateChannelRequest) error {
	return nil
}

func (bs *BarkService) DisconnectPeer(ctx context.Context, peerId string) error {
	return nil
}

func (bs *BarkService) GetNewOnchainAddress(ctx context.Context) (string, error) {
	return "", errors.New("not implemented")
}

func (bs *BarkService) GetOnchainBalance(ctx context.Context) (*lnclient.OnchainBalanceResponse, error) {
	return &lnclient.OnchainBalanceResponse{}, nil
}

func (bs *BarkService) RedeemOnchainFunds(ctx context.Context, toAddress string, amountSat uint64, feeRate *uint64, sendAll bool) (string, error) {
	return "", errors.New("not implemented")
}

func (bs *BarkService) ListPeers(ctx context.Context) ([]lnclient.PeerDetails, error) {
	return nil, nil
}

func (bs *BarkService) GetNetworkGraph(ctx context.Context, nodeIds []string) (lnclient.NetworkGraphResponse, error) {
	return nil, nil
}

func (bs *BarkService) GetLogOutput(ctx context.Context, maxLen int) ([]byte, error) {
	return []byte{}, nil
}

func (bs *BarkService) SignMessage(ctx context.Context, message string) (string, error) {
	return "", errors.New("not implemented")
}

func (bs *BarkService) GetStorageDir() (string, error) {
	return bs.workDir, nil
}

func (bs *BarkService) ResetRouter(key string) error {
	return errors.New("not implemented")
}

func (bs *BarkService) UpdateLastWalletSyncRequest() {}

func (bs *BarkService) MakeOffer(ctx context.Context, description string) (string, error) {
	return "", errors.New("not supported")
}

func (bs *BarkService) ListOnchainTransactions(ctx context.Context) ([]lnclient.OnchainTransaction, error) {
	return nil, errors.ErrUnsupported
}

func (bs *BarkService) GetCustomNodeCommandDefinitions() []lnclient.CustomNodeCommandDef {
	return nil
}

func (bs *BarkService) ExecuteCustomNodeCommand(ctx context.Context, command *lnclient.CustomNodeCommandRequest) (*lnclient.CustomNodeCommandResponse, error) {
	return nil, lnclient.ErrUnknownCustomNodeCommand
}
