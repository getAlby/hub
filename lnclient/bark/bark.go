//go:build (darwin && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (windows && amd64)

package bark

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
	bark "gitlab.com/ark-bitcoin/bark-ffi-bindings/golang/bark"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/notifications"
)

const (
	// Subsystem name reported on movements produced when a lightning receive is
	// claimed (see bark's Subsystem::LIGHTNING_RECEIVE).
	lightningReceiveSubsystem = "lightning_receive"
	// Subsystem name reported on movements produced for outgoing lightning
	// payments (see bark's Subsystem::LIGHTNING_SEND).
	lightningSendSubsystem = "lightning_send"
	// The status a movement is created with; every other status is terminal.
	movementStatusPending = "pending"
	// Movement status reported once a movement has settled. A movement first
	// appears as "pending" and is updated to this once complete.
	movementStatusSuccessful = "successful"
	// LightningReceive.State values at or past preimage reveal.
	// "delivering" (added in bark 0.6.0) sits between preimage reveal and
	// settlement: the claim is recorded and delivery resumes automatically,
	// so the funds are already irrevocably received.
	receiveStatePreimageRevealed = "preimage-revealed"
	receiveStateDelivering       = "delivering"
	receiveStateSettled          = "settled"
	// Grace period to allow the notification loop to unwind on shutdown.
	shutdownGracePeriod = 10 * time.Second
)

// Config holds the user-configurable settings for connecting to an Ark server.
type Config struct {
	// Network is the bitcoin network name (e.g. "signet", "bitcoin").
	Network string
	// ServerAddress is the Ark server URL.
	ServerAddress string
	// EsploraAddress is the Esplora server URL used for chain data.
	EsploraAddress string
	// ServerAccessToken is an optional access token required by some Ark
	// servers (currently used to gate mainnet access ahead of a public launch).
	ServerAccessToken string
	// LogLevel is the logrus level (as an int string, e.g. "3" for Info) used
	// for bark's own internal logs. Defaults to Info if empty/unparseable.
	LogLevel string
	// LogToFile controls whether bark's logs are also written to a dedicated
	// bark.log file alongside the other backend logs.
	LogToFile bool
}

type BarkService struct {
	wallet         *bark.Wallet
	workDir        string
	network        string
	eventPublisher events.EventPublisher
	pubkey         string
	cancelFn       context.CancelFunc
	loopWg         sync.WaitGroup
	// payment_hash -> waiter that handleLightningSendMovement signals.
	inflightSends    map[string]chan sendResult
	inflightSendsMtx sync.Mutex
}

type sendResult struct {
	preimage string
	feeMsat  uint64
	err      error
}

// parseNetwork maps an Alby Hub network name onto a bark network.
func parseNetwork(network string) (bark.Network, error) {
	switch network {
	case "bitcoin", "mainnet":
		return bark.NetworkBitcoin, nil
	case "testnet":
		return bark.NetworkTestnet, nil
	case "signet":
		return bark.NetworkSignet, nil
	case "regtest":
		return bark.NetworkRegtest, nil
	default:
		return 0, fmt.Errorf("unsupported bark network: %q", network)
	}
}

func NewBarkService(ctx context.Context, eventPublisher events.EventPublisher, workDir, mnemonic string, config Config) (lnclient.LNClient, error) {
	if mnemonic == "" {
		return nil, errors.New("no mnemonic configured")
	}
	if workDir == "" {
		return nil, errors.New("no bark work directory configured")
	}
	if config.ServerAddress == "" {
		return nil, errors.New("no bark server address configured")
	}

	network, err := parseNetwork(config.Network)
	if err != nil {
		return nil, err
	}

	// Forward bark's internal logs into a dedicated logger. Done before opening
	// the wallet so any logs emitted during open are captured.
	logLevel, err := strconv.Atoi(config.LogLevel)
	if err != nil {
		logLevel = int(logrus.InfoLevel)
	}
	installBarkLogger(logrus.Level(logLevel), config.LogToFile, workDir)

	// Usually, you have two wait 2 blocks. You can set nb_min_round_confirmations=0 to make it go faster.
	roundTxRequiredConfirmations := uint32(0)

	cfg := bark.Config{
		ServerAddress:                config.ServerAddress,
		RoundTxRequiredConfirmations: &roundTxRequiredConfirmations,
	}
	esploraAddress := config.EsploraAddress
	if esploraAddress != "" {
		cfg.EsploraAddress = &esploraAddress
	}
	if config.ServerAccessToken != "" {
		token := config.ServerAccessToken
		cfg.ServerAccessToken = &token
	}

	logger.Logger.WithField("workDir", workDir).Info("Opening Bark wallet")

	// Bark provides a built-in background daemon that periodically syncs with
	// the Ark server and blockchain, participates in rounds, and — crucially for
	// us — claims incoming lightning receives via the mailbox (it long-polls for
	// payment notifications and reveals the preimage, crediting the balance). We
	// don't poll for receives ourselves; instead we observe the resulting wallet
	// notifications (see runNotificationLoop) to emit payment-received events.
	wallet, err := bark.WalletOpen(network, mnemonic, cfg, bark.WalletOpenArgs{
		Datadir:           workDir,
		RunDaemon:         true,
		CreateIfNotExists: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open bark wallet: %w", err)
	}

	loopCtx, cancelFn := context.WithCancel(context.Background())
	bs := &BarkService{
		wallet:         wallet,
		workDir:        workDir,
		network:        config.Network,
		eventPublisher: eventPublisher,
		pubkey:         wallet.Fingerprint(),
		cancelFn:       cancelFn,
		inflightSends:  make(map[string]chan sendResult),
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
	go bs.runNotificationLoop(loopCtx)

	return bs, nil
}

// runNotificationLoop consumes the wallet's notification stream and publishes a
// payment-received event whenever the daemon claims an incoming lightning
// receive. The daemon does the actual claiming (it long-polls the mailbox and
// reveals the preimage); claiming a receive produces a lightning-receive
// movement, which surfaces here as a MovementCreated notification. This is
// event-driven — NextNotification blocks until something happens — so we no
// longer poll every few seconds.
func (bs *BarkService) runNotificationLoop(ctx context.Context) {
	defer bs.loopWg.Done()

	notifications := bs.wallet.Notifications()
	defer notifications.Destroy()

	// NextNotification blocks; CancelNextNotificationWait unblocks it (returning
	// nil) so the loop can exit promptly on shutdown.
	go func() {
		<-ctx.Done()
		notifications.CancelNextNotificationWait()
	}()

	for {
		if ctx.Err() != nil {
			return
		}
		notif, err := notifications.NextNotification()
		if err != nil {
			logger.Logger.WithError(err).Debug("Bark NextNotification failed")
			// Back off briefly so a persistent error doesn't spin the loop.
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
			continue
		}
		if notif == nil {
			// nil is returned when the wait was cancelled (shutdown) or the
			// notification source was shut down permanently.
			return
		}
		bs.handleNotification(*notif)
	}
}

func (bs *BarkService) handleNotification(notif bark.WalletNotification) {
	logger.Logger.WithFields(notificationLogFields(notif)).Debug("Received Bark notification")

	var movement bark.Movement
	switch n := notif.(type) {
	case bark.WalletNotificationMovementCreated:
		movement = n.Movement
	case bark.WalletNotificationMovementUpdated:
		movement = n.Movement
	default:
		// Channel lagging and other kinds carry no movement to act on.
		return
	}

	switch {
	case strings.Contains(movement.SubsystemName, lightningReceiveSubsystem):
		bs.handleLightningReceiveMovement(movement)
	case strings.Contains(movement.SubsystemName, lightningSendSubsystem):
		bs.handleLightningSendMovement(movement)
	}
}

func (bs *BarkService) handleLightningReceiveMovement(movement bark.Movement) {
	// A receive is only credited once its movement settles. We always hold the
	// preimage for our own receives, so PreimageRevealed isn't a useful signal;
	// the balance is credited when the movement status reaches "successful".
	// An abandoned receive finishes as "canceled": no funds arrived, so there is
	// nothing to report.
	if movement.Status != movementStatusSuccessful {
		return
	}

	paymentHash, ok := paymentHashFromMovement(movement)
	if !ok {
		return
	}

	receive, err := bs.wallet.LightningReceiveState(paymentHash)
	if err != nil {
		logger.Logger.WithError(err).WithField("paymentHash", paymentHash).Warn("Failed to look up claimed Bark receive")
		return
	}

	tx, err := bs.lightningReceiveToTransaction(&receive)
	if err != nil {
		logger.Logger.WithError(err).WithField("paymentHash", receive.PaymentHash).Warn("Failed to convert claimed Bark receive to transaction")
		return
	}
	logger.Logger.WithFields(logrus.Fields{
		"paymentHash": receive.PaymentHash,
		"amountSats":  receive.AmountSats,
	}).Info("Bark lightning receive claimed")
	bs.eventPublisher.Publish(&events.Event{
		Event:      "nwc_lnclient_payment_received",
		Properties: tx,
	})
}

// handleLightningSendMovement delivers a terminal lightning_send outcome to
// the SendPaymentSync waiter for the matching payment_hash. If no waiter is
// registered (e.g. the hub was restarted mid-send and SendPaymentSync's
// goroutine is gone) it falls back to publishing nwc_lnclient_payment_sent /
// _failed so the transactions service can recover the db transaction state.
func (bs *BarkService) handleLightningSendMovement(movement bark.Movement) {
	if movement.Status == movementStatusPending {
		return
	}

	paymentHash, ok := paymentHashFromMovement(movement)
	if !ok {
		return
	}

	// The movement can be canceled or failed so we should just check if it
	// wasn't successful.
	if movement.Status != movementStatusSuccessful {
		reason := fmt.Sprintf("bark lightning send %s", movement.Status)
		logger.Logger.WithFields(logrus.Fields{
			"paymentHash": paymentHash,
			"status":      movement.Status,
			"reason":      reason,
		}).Warn("Bark lightning send did not succeed")
		bs.deliverSendResult(paymentHash, sendResult{err: errors.New(reason)}, func() {
			bs.eventPublisher.Publish(&events.Event{
				Event: "nwc_lnclient_payment_failed",
				Properties: &lnclient.PaymentFailedEventProperties{
					Transaction: &lnclient.Transaction{
						Type:        constants.TRANSACTION_TYPE_OUTGOING,
						PaymentHash: paymentHash,
					},
					Reason: reason,
				},
			})
		})
		return
	}

	preimage, err := bs.getSettledSendPreimage(paymentHash)
	if err != nil {
		logger.Logger.WithError(err).WithField("paymentHash", paymentHash).Error("Bark lightning send reported successful but no preimage is available")
		bs.deliverSendResult(paymentHash, sendResult{err: fmt.Errorf("bark lightning send completed without a preimage: %w", err)}, nil)
		return
	}

	feeMsat := movement.OffchainFeeSats * 1000
	logger.Logger.WithFields(logrus.Fields{
		"paymentHash": paymentHash,
		"feeMsat":     feeMsat,
	}).Info("Bark lightning send completed")

	bs.deliverSendResult(paymentHash, sendResult{preimage: preimage, feeMsat: feeMsat}, func() {
		settledAt := time.Now().Unix()
		bs.eventPublisher.Publish(&events.Event{
			Event: "nwc_lnclient_payment_sent",
			Properties: &lnclient.Transaction{
				Type:         constants.TRANSACTION_TYPE_OUTGOING,
				PaymentHash:  paymentHash,
				Preimage:     preimage,
				FeesPaidMsat: int64(feeMsat),
				SettledAt:    &settledAt,
			},
		})
	})
}

// Reads the preimage from the lightning-send's own state. Bark records the paid
// invoice before finishing the movement, so it is always persisted by the time
// the successful movement is observed.
func (bs *BarkService) getSettledSendPreimage(paymentHash string) (string, error) {
	status, err := bs.wallet.LightningSendState(paymentHash)
	if err != nil {
		return "", fmt.Errorf("failed to look up bark lightning send state: %w", err)
	}
	paid, ok := status.(bark.LightningSendStatusPaid)
	if !ok {
		return "", fmt.Errorf("send is in state %T, expected settled", status)
	}
	if paid.Preimage == "" {
		return "", errors.New("settled send has an empty preimage")
	}
	return paid.Preimage, nil
}

// deliverSendResult delivers to the SendPaymentSync waiter if present, else
// runs fallback (used to publish an event for the hub-restart recovery path).
func (bs *BarkService) deliverSendResult(paymentHash string, res sendResult, fallback func()) {
	if ch, ok := bs.takeInflightSend(paymentHash); ok {
		ch <- res
		return
	}
	if fallback != nil {
		fallback()
	}
}

func paymentHashFromMovement(movement bark.Movement) (string, bool) {
	if movement.PaymentHash == nil || *movement.PaymentHash == "" {
		logger.Logger.WithFields(logrus.Fields{
			"movementId":    movement.Id,
			"subsystemName": movement.SubsystemName,
		}).Debug("Bark lightning movement missing payment_hash")
		return "", false
	}
	return *movement.PaymentHash, true
}

// notificationLogFields turns a Bark wallet notification into structured log
// fields describing its concrete type, rather than logging the raw interface
// pointer (which would just print an address).
func notificationLogFields(notif bark.WalletNotification) logrus.Fields {
	switch n := notif.(type) {
	case bark.WalletNotificationMovementCreated:
		return movementLogFields("movement_created", n.Movement)
	case bark.WalletNotificationMovementUpdated:
		return movementLogFields("movement_updated", n.Movement)
	case bark.WalletNotificationChannelLagging:
		return logrus.Fields{"kind": "channel_lagging"}
	default:
		return logrus.Fields{"kind": fmt.Sprintf("%T", notif)}
	}
}

func movementLogFields(kind string, m bark.Movement) logrus.Fields {
	return logrus.Fields{
		"kind":                 kind,
		"movementId":           m.Id,
		"status":               m.Status,
		"subsystemName":        m.SubsystemName,
		"subsystemKind":        m.SubsystemKind,
		"metadataJson":         m.MetadataJson,
		"intendedBalanceSats":  m.IntendedBalanceSats,
		"effectiveBalanceSats": m.EffectiveBalanceSats,
		"offchainFeeSats":      m.OffchainFeeSats,
		"sentToAddresses":      m.SentToAddresses,
		"receivedOnAddresses":  m.ReceivedOnAddresses,
		"inputVtxoIds":         m.InputVtxoIds,
		"outputVtxoIds":        m.OutputVtxoIds,
		"exitedVtxoIds":        m.ExitedVtxoIds,
		"createdAt":            m.CreatedAt,
		"updatedAt":            m.UpdatedAt,
		"completedAt":          m.CompletedAt,
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

	// The nil argument is an optional anti-DoS token, which we don't use.
	invoice, err := bs.wallet.Bolt11Invoice(uint64(amountMsat/1000), desc, nil)
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
	// Bolt11Invoice. Fetch it via the receive state so consumers can rely on
	// lookup_invoice exposing the real preimage.
	receive, err := bs.wallet.LightningReceiveState(paymentRequest.PaymentHash)
	if err != nil {
		logger.Logger.WithError(err).WithField("paymentHash", paymentRequest.PaymentHash).Error("Failed to fetch bark receive state for preimage")
		return nil, fmt.Errorf("failed to fetch bark receive state for preimage: %w", err)
	}
	if receive.PaymentPreimage == nil || *receive.PaymentPreimage == "" {
		return nil, errors.New("no preimage available")
	}
	preimage := *receive.PaymentPreimage

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

	paymentRequest, decodeErr := decodepay.Decodepay(invoice)
	if decodeErr != nil {
		return nil, fmt.Errorf("failed to decode invoice: %w", decodeErr)
	}
	paymentHash := paymentRequest.PaymentHash

	// Register a waiter BEFORE initiating the send so a notification that
	// arrives before this goroutine reaches the receive cannot be missed.
	resultCh := make(chan sendResult, 1)
	if err := bs.registerInflightSend(paymentHash, resultCh); err != nil {
		return nil, err
	}
	defer bs.clearInflightSend(paymentHash)

	if _, err := bs.wallet.PayLightningInvoice(invoice, nil, false); err != nil {
		return nil, fmt.Errorf("bark PayLightningInvoice failed: %w", err)
	}

	// Block until handleLightningSendMovement delivers a terminal result.
	res := <-resultCh
	if res.err != nil {
		return nil, res.err
	}
	return &lnclient.PayInvoiceResponse{
		Preimage: res.preimage,
		FeeMsat:  res.feeMsat,
	}, nil
}

func (bs *BarkService) registerInflightSend(paymentHash string, ch chan sendResult) error {
	bs.inflightSendsMtx.Lock()
	defer bs.inflightSendsMtx.Unlock()
	if _, exists := bs.inflightSends[paymentHash]; exists {
		return fmt.Errorf("a bark lightning send is already in flight for payment hash %s", paymentHash)
	}
	bs.inflightSends[paymentHash] = ch
	return nil
}

func (bs *BarkService) clearInflightSend(paymentHash string) {
	bs.inflightSendsMtx.Lock()
	defer bs.inflightSendsMtx.Unlock()
	delete(bs.inflightSends, paymentHash)
}

func (bs *BarkService) takeInflightSend(paymentHash string) (chan sendResult, bool) {
	bs.inflightSendsMtx.Lock()
	defer bs.inflightSendsMtx.Unlock()
	ch, ok := bs.inflightSends[paymentHash]
	if ok {
		delete(bs.inflightSends, paymentHash)
	}
	return ch, ok
}

func (bs *BarkService) LookupInvoice(ctx context.Context, paymentHash string) (*lnclient.Transaction, error) {
	return nil, errors.New("this method should not be called")
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
	// Only report the receive as settled when we can include the preimage —
	// a settled transaction without one is rejected by the transactions
	// service.
	if receive.PaymentPreimage != nil && receiveIsPaid(receive.State) {
		tx.Preimage = *receive.PaymentPreimage
		settledAt := time.Now().Unix()
		if receive.SettledAt != nil {
			settledAt = *receive.SettledAt
		}
		tx.SettledAt = &settledAt
	}
	return tx, nil
}

// receiveIsPaid reports whether a receive's state is at or past preimage
// reveal, meaning the payer holds the preimage and the payment is final.
// The state is the only reliable signal: bark generates and stores the
// preimage at invoice creation, so LightningReceive.PaymentPreimage can be
// set long before anything is paid.
func receiveIsPaid(state string) bool {
	switch state {
	case receiveStatePreimageRevealed, receiveStateDelivering, receiveStateSettled:
		return true
	}
	return false
}

func (bs *BarkService) GetBalances(ctx context.Context, includeInactiveChannels bool) (*lnclient.BalancesResponse, error) {
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
			TotalSpendableMsat:      spendableMsat,
			NextMaxSpendableMsat:    spendableMsat,
			NextMaxSpendableMPPMsat: spendableMsat,
		},
	}, nil
}

func (bs *BarkService) GetInfo(ctx context.Context) (*lnclient.NodeInfo, error) {
	return &lnclient.NodeInfo{
		Alias:   "Bark",
		Color:   "#897FFF",
		Pubkey:  bs.pubkey,
		Network: bs.network,
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
	// payment_received is emitted from runNotificationLoop when the daemon
	// claims an incoming receive; payment_sent is emitted by the transactions
	// service when our synchronous SendPaymentSync succeeds.
	return []string{
		notifications.PAYMENT_RECEIVED_NOTIFICATION,
		notifications.PAYMENT_SENT_NOTIFICATION,
	}
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
		case <-time.After(shutdownGracePeriod):
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

func (bs *BarkService) CloseChannel(ctx context.Context, closeChannelRequest *lnclient.CloseChannelRequest) error {
	return nil
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

const (
	nodeCommandDebug                  = "debug"
	nodeCommandClaimLightningReceives = "claimlightningreceives"
	nodeCommandRunMaintenance         = "runmaintenance"
	nodeCommandRecoveryReport         = "recoveryreport"
)

func (bs *BarkService) GetCustomNodeCommandDefinitions() []lnclient.CustomNodeCommandDef {
	return []lnclient.CustomNodeCommandDef{
		{
			Name:        nodeCommandDebug,
			Description: "Dump the wallet's balance breakdown, VTXOs, pending lightning receives, movement history and Ark server info. Useful for debugging a receive that did not credit your balance.",
			Args:        nil,
		},
		{
			Name:        nodeCommandClaimLightningReceives,
			Description: "Attempt to claim any pending/unclaimed lightning receives. Use this if an invoice was paid but the funds have not shown up in your balance.",
			Args:        nil,
		},
		{
			Name:        nodeCommandRunMaintenance,
			Description: "Run wallet maintenance, which progresses pending rounds and refreshes VTXOs. Use this to nudge funds that are stuck 'pending in round'.",
			Args:        nil,
		},
		{
			Name:        nodeCommandRecoveryReport,
			Description: "Show the result of the seed-recovery scan that runs when a wallet is created from an existing recovery phrase. Use this to verify your funds were restored after migrating to a new device.",
			Args:        nil,
		},
	}
}

func (bs *BarkService) ExecuteCustomNodeCommand(ctx context.Context, command *lnclient.CustomNodeCommandRequest) (*lnclient.CustomNodeCommandResponse, error) {
	switch command.Name {
	case nodeCommandDebug:
		return bs.executeCommandDebug()
	case nodeCommandClaimLightningReceives:
		return bs.executeCommandClaimLightningReceives()
	case nodeCommandRunMaintenance:
		return bs.executeCommandRunMaintenance()
	case nodeCommandRecoveryReport:
		return bs.executeCommandRecoveryReport()
	}

	return nil, lnclient.ErrUnknownCustomNodeCommand
}

func (bs *BarkService) executeCommandDebug() (*lnclient.CustomNodeCommandResponse, error) {
	// Sync first so we report current state rather than a stale snapshot (the
	// same pattern GetBalances uses before reading the balance).
	if err := bs.wallet.Sync(); err != nil {
		logger.Logger.WithError(err).Warn("Bark sync failed before collecting debug info")
	}

	response := map[string]interface{}{
		"network": bs.network,
		"pubkey":  bs.pubkey,
	}

	if balance, err := bs.wallet.Balance(); err != nil {
		response["balanceError"] = err.Error()
	} else {
		response["balance"] = balance
	}

	if claimable, err := bs.wallet.ClaimableLightningReceiveBalanceSats(); err != nil {
		response["claimableLightningReceiveSatsError"] = err.Error()
	} else {
		response["claimableLightningReceiveSats"] = claimable
	}

	if vtxos, err := bs.wallet.Vtxos(); err != nil {
		response["vtxosError"] = err.Error()
	} else {
		response["vtxos"] = vtxos
	}

	if spendable, err := bs.wallet.SpendableVtxos(); err != nil {
		response["spendableVtxosError"] = err.Error()
	} else {
		response["spendableVtxos"] = spendable
	}

	if pending, err := bs.wallet.PendingLightningReceives(); err != nil {
		response["pendingLightningReceivesError"] = err.Error()
	} else {
		response["pendingLightningReceives"] = pending
	}

	if history, err := bs.wallet.History(); err != nil {
		response["historyError"] = err.Error()
	} else {
		response["history"] = history
	}

	// Round state explains funds stuck in PendingInRoundSats: such funds sit in a
	// round whose funding tx is waiting for confirmations (6 on mainnet), which
	// the daemon progresses automatically once confirmed.
	if rounds, err := bs.wallet.PendingRoundStates(); err != nil {
		response["pendingRoundStatesError"] = err.Error()
	} else {
		response["pendingRoundStates"] = rounds
	}

	if nextRoundStartTime, err := bs.wallet.NextRoundStartTime(); err != nil {
		response["nextRoundStartTimeError"] = err.Error()
	} else {
		response["nextRoundStartTime"] = nextRoundStartTime
	}

	if arkInfo := bs.wallet.ArkInfo(); arkInfo != nil {
		response["arkInfo"] = arkInfo
	}

	return &lnclient.CustomNodeCommandResponse{
		Response: response,
	}, nil
}

func (bs *BarkService) executeCommandRunMaintenance() (*lnclient.CustomNodeCommandResponse, error) {
	if err := bs.wallet.Maintenance(); err != nil {
		return nil, fmt.Errorf("failed to run maintenance: %w", err)
	}

	logger.Logger.Debug("Ran Bark maintenance")

	balance, err := bs.wallet.Balance()
	if err != nil {
		return nil, fmt.Errorf("maintenance succeeded but failed to read balance: %w", err)
	}

	return &lnclient.CustomNodeCommandResponse{
		Response: map[string]interface{}{
			"message": "Maintenance completed.",
			"balance": balance,
		},
	}, nil
}

func (bs *BarkService) executeCommandClaimLightningReceives() (*lnclient.CustomNodeCommandResponse, error) {
	if err := bs.wallet.Sync(); err != nil {
		logger.Logger.WithError(err).Warn("Bark sync failed before claiming lightning receives")
	}

	// wait=false: attempt to claim what is already claimable without blocking on
	// the server long-polling for not-yet-arrived payments.
	claimed, err := bs.wallet.TryClaimAllLightningReceives(false)
	if err != nil {
		return nil, fmt.Errorf("failed to claim lightning receives: %w", err)
	}

	logger.Logger.WithField("count", len(claimed)).Info("Attempted to claim Bark lightning receives")

	return &lnclient.CustomNodeCommandResponse{
		Response: map[string]interface{}{
			"claimedCount": len(claimed),
			"claimed":      claimed,
		},
	}, nil
}

func (bs *BarkService) executeCommandRecoveryReport() (*lnclient.CustomNodeCommandResponse, error) {
	// The report is produced by the seed-recovery scan bark runs during the
	// wallet open that creates the wallet locally (e.g. when restoring from a
	// recovery phrase on a new device). It is only available in the session
	// that created the wallet; on subsequent starts no scan runs.
	report := bs.wallet.RecoveryReport()
	if report == nil {
		return &lnclient.CustomNodeCommandResponse{
			Response: map[string]interface{}{
				"message": "No recovery scan ran on this wallet start. A scan only runs when the wallet is first created, e.g. after restoring from a recovery phrase.",
			},
		}, nil
	}

	return &lnclient.CustomNodeCommandResponse{
		Response: map[string]interface{}{
			"isComplete": report.IsComplete,
			"recovered":  report.Recovered,
			"skipped":    report.Skipped,
			"foreign":    report.Foreign,
			"failed":     report.Failed,
			"exited":     report.Exited,
		},
	}, nil
}
