package ldk

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	// "github.com/getAlby/ldk-node-go/ldk_node"
	"github.com/getAlby/hub/ldk_node"

	"encoding/hex"
	"encoding/json"

	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/lsp"
	"github.com/getAlby/hub/utils"
)

type LDKService struct {
	workdir               string
	node                  *ldk_node.Node
	ldkEventBroadcaster   LDKEventBroadcaster
	cancel                context.CancelFunc
	network               string
	eventPublisher        events.EventPublisher
	syncing               bool
	lastFullSync          time.Time
	lastFeeEstimatesSync  time.Time
	cfg                   config.Config
	lastWalletSyncRequest time.Time
	pubkey                string
}

const resetRouterKey = "ResetRouter"

func NewLDKService(ctx context.Context, cfg config.Config, eventPublisher events.EventPublisher, mnemonic, workDir string, network string) (result lnclient.LNClient, err error) {
	if mnemonic == "" || workDir == "" {
		return nil, errors.New("one or more required LDK configuration are missing")
	}

	//create dir if not exists
	newpath := filepath.Join(workDir)
	err = os.MkdirAll(newpath, os.ModePerm)
	if err != nil {
		log.Printf("Failed to create LDK working dir: %v", err)
		return nil, err
	}

	logDirPath := filepath.Join(newpath, "./logs")

	ldkConfig := ldk_node.DefaultConfig()
	listeningAddresses := []string{
		"0.0.0.0:9735",
		"[::]:9735",
	}
	ldkConfig.TrustedPeers0conf = []string{
		lsp.OlympusLSP().Pubkey,
		lsp.AlbyPlebsLSP().Pubkey,
		lsp.MegalithLSP().Pubkey,

		// Mutinynet
		lsp.AlbyMutinynetPlebsLSP().Pubkey,
		lsp.OlympusMutinynetLSP().Pubkey,
		lsp.MegalithMutinynetLSP().Pubkey,
	}
	ldkConfig.AnchorChannelsConfig.TrustedPeersNoReserve = []string{
		lsp.OlympusLSP().Pubkey,
		lsp.AlbyPlebsLSP().Pubkey,
		lsp.MegalithLSP().Pubkey,
		"02b4552a7a85274e4da01a7c71ca57407181752e8568b31d51f13c111a2941dce3", // LNServer_Wave
		"0296b2db342fcf87ea94d981757fdf4d3e545bd5cef4919f58b5d38dfdd73bf5c9", // blocktank
		"038ba8f67ba8ff5c48764cdd3251c33598d55b203546d08a8f0ec9dcd9f27e3637", // flashsats
		"0370a5392cd7c81ff5128fa656ee6db0c4d11c778fcd6cb98cb6ba3b48394f5705", // lqwd

		// Mutinynet
		lsp.AlbyMutinynetPlebsLSP().Pubkey,
		lsp.OlympusMutinynetLSP().Pubkey,
		lsp.MegalithMutinynetLSP().Pubkey,
		"0296820bbba5bd33719962bafd69996ee89e03ce7164d8f368cbb85463f5f47876", // flashsats
		"035e8a9034a8c68f219aacadae748c7a3cd719109309db39b09886e5ff17696b1b", // lqwd
	}

	ldkConfig.ListeningAddresses = &listeningAddresses
	ldkConfig.LogDirPath = &logDirPath
	logLevel, err := strconv.Atoi(cfg.GetEnv().LDKLogLevel)
	if err == nil {
		// LogLevelGossip is added due to bug in go bindings which uses an enum that starts at 1 instead of 0
		// If LogLevelGossip is changed to 0, this addition can be removed
		ldkConfig.LogLevel = ldk_node.LogLevel(logLevel) + ldk_node.LogLevelGossip
	}
	builder := ldk_node.BuilderFromConfig(ldkConfig)
	builder.SetEntropyBip39Mnemonic(mnemonic, nil)
	builder.SetNetwork(network)
	builder.SetEsploraServer(cfg.GetEnv().LDKEsploraServer)
	if cfg.GetEnv().LDKGossipSource != "" {
		logger.Logger.WithField("gossipSource", cfg.GetEnv().LDKGossipSource).Warn("LDK RGS instance set")
		builder.SetGossipSourceRgs(cfg.GetEnv().LDKGossipSource)
	}
	builder.SetStorageDirPath(filepath.Join(newpath, "./storage"))

	// TODO: remove when https://github.com/lightningdevkit/rust-lightning/issues/2914 is merged
	// LDK default HTLC inflight value is 10% of the channel size. If an LSPS service is configured this will be set to 0.
	// The liquidity source below is not used because we do not use the native LDK-node LSPS2 API.
	builder.SetLiquiditySourceLsps2("52.88.33.119:9735", lsp.OlympusLSP().Pubkey, nil)

	//builder.SetLogDirPath (filepath.Join(newpath, "./logs")); // missing?
	node, err := builder.Build()

	if err != nil {
		logger.Logger.Errorf("Failed to create LDK node: %v", err)
		return nil, err
	}

	ldkEventConsumer := make(chan *ldk_node.Event)
	ldkCtx, cancel := context.WithCancel(ctx)
	ldkEventBroadcaster := NewLDKEventBroadcaster(ldkCtx, ldkEventConsumer)
	nodeId := node.NodeId()

	ls := LDKService{
		workdir:             newpath,
		node:                node,
		cancel:              cancel,
		ldkEventBroadcaster: ldkEventBroadcaster,
		network:             network,
		eventPublisher:      eventPublisher,
		cfg:                 cfg,
		pubkey:              nodeId,
	}

	// TODO: remove when LDK supports this
	deleteOldLDKLogs(logDirPath)
	go func() {
		// delete old LDK logs every 24 hours
		ticker := time.NewTicker(24 * time.Hour)
		for {
			select {
			case <-ticker.C:
				deleteOldLDKLogs(logDirPath)
			case <-ldkCtx.Done():
				return
			}
		}
	}()

	// check for and forward new LDK events to LDKEventBroadcaster (through ldkEventConsumer)
	go func() {
		for {
			select {
			case <-ldkCtx.Done():
				return
			default:
				// NOTE: currently do not use WaitNextEvent() as it can possibly block the LDK thread (to confirm)
				event := node.NextEvent()
				if event == nil {
					// if there is no event, wait before polling again to avoid 100% CPU usage
					// TODO: remove this and use WaitNextEvent()
					time.Sleep(time.Duration(1) * time.Millisecond)
					continue
				}

				ls.handleLdkEvent(event)
				ldkEventConsumer <- event

				node.EventHandled()
			}
		}
	}()

	err = node.Start()
	if err != nil {
		logger.Logger.Errorf("Failed to start LDK node: %v", err)
		return nil, err
	}

	logger.Logger.WithFields(logrus.Fields{
		"nodeId": nodeId,
		"status": node.Status(),
	}).Info("Started LDK node. Syncing wallet...")

	syncStartTime := time.Now()
	err = node.SyncWallets()
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to sync LDK wallets")
		ls.eventPublisher.Publish(&events.Event{
			Event: "nwc_node_sync_failed",
			Properties: map[string]interface{}{
				"error":        err.Error(),
				"sync_type":    "full",
				"initial_sync": true,
				"node_type":    config.LDKBackendType,
				"esplora_url":  ls.cfg.GetEnv().LDKEsploraServer,
			},
		})

		shutdownErr := ls.Shutdown()
		if shutdownErr != nil {
			logger.Logger.WithError(shutdownErr).Error("Failed to shutdown LDK node")
		}

		return nil, err
	}
	ls.lastFullSync = time.Now()
	ls.lastFeeEstimatesSync = time.Now()

	logger.Logger.WithFields(logrus.Fields{
		"nodeId":   nodeId,
		"status":   node.Status(),
		"duration": math.Ceil(time.Since(syncStartTime).Seconds()),
	}).Info("LDK node synced successfully")

	// backup channels after successful startup
	// TODO: consider removing after all hubs are updated past 1.9.1
	ls.backupChannels()

	if ls.network == "bitcoin" {
		// try to connect to some peers to retrieve P2P gossip data. TODO: Remove once LDK can correctly do gossip with CLN and Eclair nodes
		// see https://github.com/lightningdevkit/rust-lightning/issues/3075
		peers := []string{
			"031b301307574bbe9b9ac7b79cbe1700e31e544513eae0b5d7497483083f99e581@45.79.192.236:9735",   // Olympus
			"0364913d18a19c671bb36dd04d6ad5be0fe8f2894314c36a9db3f03c2d414907e1@192.243.215.102:9735", // LQwD
			"035e4ff418fc8b5554c5d9eea66396c227bd429a3251c8cbc711002ba215bfc226@170.75.163.209:9735",  // WoS
			"02fcc5bfc48e83f06c04483a2985e1c390cb0f35058baa875ad2053858b8e80dbd@35.239.148.251:9735",  // Blink
			"027100442c3b79f606f80f322d98d499eefcb060599efc5d4ecb00209c2cb54190@3.230.33.224:9735",    // c=
			"038a9e56512ec98da2b5789761f7af8f280baf98a09282360cd6ff1381b5e889bf@64.23.162.51:9735",    // Megalith LSP
		}
		logger.Logger.Info("Connecting to some peers to retrieve P2P gossip data")
		for _, peer := range peers {
			parts := strings.FieldsFunc(peer, func(r rune) bool { return r == '@' || r == ':' })
			port, err := strconv.ParseUint(parts[2], 10, 16)
			if err != nil {
				logger.Logger.WithError(err).Error("Failed to parse port number")
				continue
			}
			err = ls.ConnectPeer(ctx, &lnclient.ConnectPeerRequest{
				Pubkey:  parts[0],
				Address: parts[1],
				Port:    uint16(port),
			})
			if err != nil {
				logger.Logger.WithField("peer", peer).WithError(err).Error("Failed to connect to peer")
			}
		}
	}

	// setup background sync
	go func() {
		MIN_SYNC_INTERVAL := 1 * time.Minute
		MIN_FEE_ESTIMATES_SYNC_INTERVAL := 5 * time.Minute
		MAX_SYNC_INTERVAL := 1 * time.Hour // NOTE: this could be increased further (possibly to 6 hours)
		for {
			ls.syncing = false
			select {
			case <-ldkCtx.Done():
				return
			case <-time.After(MIN_SYNC_INTERVAL):
				ls.syncing = true

				channels := ls.node.ListChannels()
				for _, channel := range channels {
					if channel.Confirmations != nil && channel.ConfirmationsRequired != nil && *channel.Confirmations < *channel.ConfirmationsRequired {
						logger.Logger.WithField("channel_id", channel.UserChannelId).Debug("Using short sync time while opening channel")
						ls.lastWalletSyncRequest = time.Now()
						break
					}
				}

				if time.Since(ls.lastWalletSyncRequest) > MIN_SYNC_INTERVAL && time.Since(ls.lastFullSync) < MAX_SYNC_INTERVAL {

					if time.Since(ls.lastFeeEstimatesSync) < MIN_FEE_ESTIMATES_SYNC_INTERVAL {
						logger.Logger.Debug("Skipping updating fee estimates")
						continue
					}

					// only update fee estimates
					logger.Logger.Debug("Updating fee estimates")
					err = node.UpdateFeeEstimates()
					if err != nil {
						logger.Logger.WithError(err).Error("Failed to update fee estimates")
						ls.eventPublisher.Publish(&events.Event{
							Event: "nwc_node_sync_failed",
							Properties: map[string]interface{}{
								"error":       err.Error(),
								"sync_type":   "fee_estimates",
								"node_type":   config.LDKBackendType,
								"esplora_url": ls.cfg.GetEnv().LDKEsploraServer,
							},
						})
						continue
					}
					ls.lastFeeEstimatesSync = time.Now()
					continue
				}

				logger.Logger.Debug("Starting full background wallet sync")
				syncStartTime := time.Now()
				err = node.SyncWallets()

				if err != nil {
					logger.Logger.WithError(err).Error("Failed to sync LDK wallets")
					ls.eventPublisher.Publish(&events.Event{
						Event: "nwc_node_sync_failed",
						Properties: map[string]interface{}{
							"error":       err.Error(),
							"sync_type":   "full",
							"node_type":   config.LDKBackendType,
							"esplora_url": ls.cfg.GetEnv().LDKEsploraServer,
						},
					})

					// try again at next MIN_SYNC_INTERVAL
					continue
				}

				ls.lastFullSync = time.Now()
				// fee estimates happens as part of full sync
				ls.lastFeeEstimatesSync = time.Now()

				logger.Logger.WithFields(logrus.Fields{
					"nodeId":   nodeId,
					"status":   node.Status(),
					"duration": math.Ceil(time.Since(syncStartTime).Seconds()),
				}).Info("LDK node synced successfully")
			}
		}
	}()

	return &ls, nil
}

func (ls *LDKService) Shutdown() error {
	if ls.node == nil {
		logger.Logger.Debug("LDK client already shut down")
		return nil
	}
	// make sure nothing else can use it
	node := ls.node
	ls.node = nil

	logger.Logger.Info("shutting down LDK client")
	logger.Logger.Info("cancelling LDK context")
	ls.cancel()

	for ls.syncing {
		logger.Logger.Warn("Waiting for background sync to finish before stopping LDK node...")
		time.Sleep(1 * time.Second)
	}

	logger.Logger.Info("stopping LDK node")
	shutdownChannel := make(chan error)
	go func() {
		shutdownChannel <- node.Stop()
	}()

	select {
	case err := <-shutdownChannel:
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to stop LDK node")
			// do not return error - we still need to destroy the node
		} else {
			logger.Logger.Info("LDK stop node succeeded")
		}
	case <-time.After(120 * time.Second):
		logger.Logger.Error("Timeout shutting down LDK node after 120 seconds")
	}

	logger.Logger.Debug("Destroying LDK node object")
	node.Destroy()

	ls.resetRouterInternal()

	logger.Logger.Info("LDK shutdown complete")

	return nil
}

func (ls *LDKService) resetRouterInternal() {
	key, err := ls.cfg.Get(resetRouterKey, "")

	if err != nil {
		logger.Logger.Error("Failed to retrieve ResetRouter key")
	}

	if key != "" {
		ls.cfg.SetUpdate(resetRouterKey, "", "")
		logger.Logger.WithField("key", key).Info("Resetting router")

		ldkDbPath := filepath.Join(ls.workdir, "storage", "ldk_node_data.sqlite")
		if _, err := os.Stat(ldkDbPath); errors.Is(err, os.ErrNotExist) {
			logger.Logger.Error("Could not find LDK database")
			return
		}
		ldkDb, err := sql.Open("sqlite", ldkDbPath)
		if err != nil {
			logger.Logger.Error("Could not open LDK DB file")
			return
		}

		command := ""

		switch key {
		case "ALL":
			command = "delete from ldk_node_data where key = 'scorer' or key = 'network_graph';VACUUM;"
		case "Scorer":
			command = "delete from ldk_node_data where key = 'scorer';VACUUM;"
		case "NetworkGraph":
			command = "delete from ldk_node_data where key = 'network_graph';VACUUM;"
		default:
			logger.Logger.WithField("key", key).Error("Unknown reset router key")
			return
		}

		result, err := ldkDb.Exec(command)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed execute reset command")
			return
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to get rows affected")
			return
		}
		logger.Logger.WithFields(logrus.Fields{
			"rowsAffected": rowsAffected,
		}).Info("Reset router")
	}
}

func (ls *LDKService) SendPaymentSync(ctx context.Context, invoice string) (*lnclient.PayInvoiceResponse, error) {
	paymentRequest, err := decodepay.Decodepay(invoice)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"bolt11": invoice,
		}).Errorf("Failed to decode bolt11 invoice: %v", err)

		return nil, err
	}

	maxSpendable := ls.getMaxSpendable()
	if paymentRequest.MSatoshi > maxSpendable {
		ls.eventPublisher.Publish(&events.Event{
			Event: "nwc_outgoing_liquidity_required",
			Properties: map[string]interface{}{
				//"amount":         amount / 1000,
				//"max_receivable": maxReceivable,
				//"num_channels":   len(gs.node.ListChannels()),
				"node_type": config.LDKBackendType,
			},
		})
	}

	paymentStart := time.Now()
	ldkEventSubscription := ls.ldkEventBroadcaster.Subscribe()
	defer ls.ldkEventBroadcaster.CancelSubscription(ldkEventSubscription)

	paymentHash, err := ls.node.Bolt11Payment().Send(invoice)
	if err != nil {
		logger.Logger.WithError(err).Error("SendPayment failed")
		return nil, err
	}
	fee := uint64(0)
	preimage := ""

	for start := time.Now(); time.Since(start) < time.Second*60; {
		event := <-ldkEventSubscription

		eventPaymentSuccessful, isEventPaymentSuccessfulEvent := (*event).(ldk_node.EventPaymentSuccessful)
		eventPaymentFailed, isEventPaymentFailedEvent := (*event).(ldk_node.EventPaymentFailed)

		if isEventPaymentSuccessfulEvent && eventPaymentSuccessful.PaymentHash == paymentHash {
			logger.Logger.Info("Got payment success event")
			payment := ls.node.Payment(paymentHash)
			if payment == nil {
				logger.Logger.Errorf("Couldn't find payment by payment hash: %v", paymentHash)
				return nil, errors.New("payment not found")
			}

			bolt11PaymentKind, ok := payment.Kind.(ldk_node.PaymentKindBolt11)

			if !ok {
				logger.Logger.WithFields(logrus.Fields{
					"payment": payment,
				}).Error("Payment is not a bolt11 kind")
			}

			if bolt11PaymentKind.Preimage == nil {
				logger.Logger.Errorf("No payment preimage for payment hash: %v", paymentHash)
				return nil, errors.New("payment preimage not found")
			}
			preimage = *bolt11PaymentKind.Preimage

			if eventPaymentSuccessful.FeePaidMsat != nil {
				fee = *eventPaymentSuccessful.FeePaidMsat
			}
			break
		}
		if isEventPaymentFailedEvent && eventPaymentFailed.PaymentHash == paymentHash {

			failureReasonMessage := ls.getPaymentFailReason(&eventPaymentFailed)

			logger.Logger.WithFields(logrus.Fields{
				"payment_hash": paymentHash,
				"reason":       failureReasonMessage,
			}).Error("Received payment failed event")

			return nil, fmt.Errorf("received payment failed event: %s", failureReasonMessage)
		}
	}
	if preimage == "" {
		logger.Logger.WithFields(logrus.Fields{
			"paymentHash": paymentHash,
		}).Warn("Timed out waiting for payment to be sent")
		return nil, lnclient.NewTimeoutError()
	}

	logger.Logger.WithFields(logrus.Fields{
		"duration": time.Since(paymentStart).Milliseconds(),
		"fee":      fee,
	}).Info("Successful payment")

	return &lnclient.PayInvoiceResponse{
		Preimage: preimage,
		Fee:      fee,
	}, nil
}

func (ls *LDKService) SendKeysend(ctx context.Context, amount uint64, destination string, custom_records []lnclient.TLVRecord, preimage string) (*lnclient.PayKeysendResponse, error) {
	paymentStart := time.Now()
	customTlvs := []ldk_node.TlvEntry{}

	for _, customRecord := range custom_records {
		decodedValue, err := hex.DecodeString(customRecord.Value)
		if err != nil {
			return nil, err
		}
		customTlvs = append(customTlvs, ldk_node.TlvEntry{
			Type:  customRecord.Type,
			Value: decodedValue,
		})
	}

	ldkEventSubscription := ls.ldkEventBroadcaster.Subscribe()
	defer ls.ldkEventBroadcaster.CancelSubscription(ldkEventSubscription)

	paymentHash, err := ls.node.SpontaneousPayment().Send(amount, destination, customTlvs, &preimage)
	if err != nil {
		logger.Logger.WithError(err).Error("Keysend failed")
		return nil, err
	}
	fee := uint64(0)
	paid := false
	for start := time.Now(); time.Since(start) < time.Second*60; {
		event := <-ldkEventSubscription

		eventPaymentSuccessful, isEventPaymentSuccessfulEvent := (*event).(ldk_node.EventPaymentSuccessful)
		eventPaymentFailed, isEventPaymentFailedEvent := (*event).(ldk_node.EventPaymentFailed)

		if isEventPaymentSuccessfulEvent && eventPaymentSuccessful.PaymentHash == paymentHash {
			logger.Logger.Info("Got payment success event")

			paid = true

			if eventPaymentSuccessful.FeePaidMsat != nil {
				fee = *eventPaymentSuccessful.FeePaidMsat
			}
			break
		}
		if isEventPaymentFailedEvent && eventPaymentFailed.PaymentHash == paymentHash {

			failureReasonMessage := ls.getPaymentFailReason(&eventPaymentFailed)

			logger.Logger.WithFields(logrus.Fields{
				"payment_hash": paymentHash,
				"reason":       failureReasonMessage,
			}).Error("Received payment failed event")

			return nil, fmt.Errorf("payment failed event: %s", failureReasonMessage)
		}
	}
	if !paid {
		logger.Logger.WithFields(logrus.Fields{
			"payment_hash": paymentHash,
		}).Warn("Timed out waiting for keysend to be sent")
		return nil, lnclient.NewTimeoutError()
	}

	logger.Logger.WithFields(logrus.Fields{
		"duration": time.Since(paymentStart).Milliseconds(),
		"fee":      fee,
	}).Info("Successful keysend payment")
	return &lnclient.PayKeysendResponse{
		Fee: fee,
	}, nil
}

func (ls *LDKService) getMaxReceivable() int64 {
	var receivable int64 = 0
	channels := ls.node.ListChannels()
	for _, channel := range channels {
		if channel.IsUsable {
			receivable += min(int64(channel.InboundCapacityMsat), int64(*channel.InboundHtlcMaximumMsat))
		}
	}
	return int64(receivable)
}

func (ls *LDKService) getMaxSpendable() int64 {
	var spendable int64 = 0
	channels := ls.node.ListChannels()
	for _, channel := range channels {
		if channel.IsUsable {
			spendable += min(int64(channel.OutboundCapacityMsat), int64(*channel.CounterpartyOutboundHtlcMaximumMsat))
		}
	}
	return int64(spendable)
}

func (ls *LDKService) MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64) (transaction *lnclient.Transaction, err error) {

	maxReceivable := ls.getMaxReceivable()

	if amount > maxReceivable {
		ls.eventPublisher.Publish(&events.Event{
			Event: "nwc_incoming_liquidity_required",
			Properties: map[string]interface{}{
				//"amount":         amount / 1000,
				//"max_receivable": maxReceivable,
				//"num_channels":   len(gs.node.ListChannels()),
				"node_type": config.LDKBackendType,
			},
		})
	}

	if expiry == 0 {
		expiry = lnclient.DEFAULT_INVOICE_EXPIRY
	}

	// TODO: support passing description hash
	invoice, err := ls.node.Bolt11Payment().Receive(uint64(amount),
		description,
		uint32(expiry))

	if err != nil {
		logger.Logger.WithError(err).Error("MakeInvoice failed")
		return nil, err
	}

	var expiresAt *int64
	paymentRequest, err := decodepay.Decodepay(invoice)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"bolt11": invoice,
		}).Errorf("Failed to decode bolt11 invoice: %v", err)

		return nil, err
	}
	expiresAtUnix := time.UnixMilli(int64(paymentRequest.CreatedAt) * 1000).Add(time.Duration(paymentRequest.Expiry) * time.Second).Unix()
	expiresAt = &expiresAtUnix
	description = paymentRequest.Description
	descriptionHash = paymentRequest.DescriptionHash

	payment := ls.node.Payment(paymentRequest.PaymentHash)

	transaction = &lnclient.Transaction{
		Type:            "incoming",
		Invoice:         invoice,
		PaymentHash:     paymentRequest.PaymentHash,
		Preimage:        *payment.Kind.(ldk_node.PaymentKindBolt11).Preimage,
		Amount:          amount,
		CreatedAt:       int64(paymentRequest.CreatedAt),
		ExpiresAt:       expiresAt,
		Description:     description,
		DescriptionHash: descriptionHash,
	}

	return transaction, nil
}

func (ls *LDKService) LookupInvoice(ctx context.Context, paymentHash string) (transaction *lnclient.Transaction, err error) {

	payment := ls.node.Payment(paymentHash)
	if payment == nil {
		logger.Logger.Errorf("Couldn't find payment by payment hash: %v", paymentHash)
		return nil, errors.New("Payment not found")
	}

	transaction, err = ls.ldkPaymentToTransaction(payment)

	if err != nil {
		logger.Logger.Errorf("Failed to map transaction: %v", err)
		return nil, err
	}

	return transaction, nil
}

func (ls *LDKService) ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaid bool, invoiceType string) (transactions []lnclient.Transaction, err error) {
	transactions = []lnclient.Transaction{}

	// TODO: support pagination
	payments := ls.node.ListPayments()

	for _, payment := range payments {
		if payment.Status == ldk_node.PaymentStatusSucceeded || unpaid {
			transaction, err := ls.ldkPaymentToTransaction(&payment)

			if err != nil {
				logger.Logger.WithError(err).Error("Failed to map transaction")
				continue
			}

			// locally filter
			if from != 0 && uint64(transaction.CreatedAt) < from {
				continue
			}
			if until != 0 && uint64(transaction.CreatedAt) > until {
				continue
			}
			if invoiceType != "" && transaction.Type != invoiceType {
				continue
			}

			transactions = append(transactions, *transaction)
		}
	}

	// sort by created date descending
	sort.SliceStable(transactions, func(i, j int) bool {
		return transactions[i].CreatedAt > transactions[j].CreatedAt
	})

	if offset > 0 {
		if offset < uint64(len(transactions)) {
			transactions = transactions[offset:]
		} else {
			transactions = []lnclient.Transaction{}
		}
	}

	if len(transactions) > int(limit) {
		transactions = transactions[:limit]
	}

	// logger.Logger.WithField("transactions", transactions).Debug("Listed transactions")

	return transactions, nil
}

func (ls *LDKService) GetInfo(ctx context.Context) (info *lnclient.NodeInfo, err error) {
	// TODO: should alias, color be configured in LDK-node? or can we manage them in NWC?
	// an alias is only needed if the user has public channels and wants their node to be publicly visible?
	status := ls.node.Status()
	return &lnclient.NodeInfo{
		Alias:       "NWC",
		Color:       "#897FFF",
		Pubkey:      ls.node.NodeId(),
		Network:     ls.network,
		BlockHeight: status.CurrentBestBlock.Height,
		BlockHash:   status.CurrentBestBlock.BlockHash,
	}, nil
}

func (ls *LDKService) ListChannels(ctx context.Context) ([]lnclient.Channel, error) {

	ldkChannels := ls.node.ListChannels()

	channels := []lnclient.Channel{}

	// logger.Logger.WithFields(logrus.Fields{
	// 	"channels": ldkChannels,
	// }).Debug("Listed Channels")

	for _, ldkChannel := range ldkChannels {
		fundingTxId := ""
		if ldkChannel.FundingTxo != nil {
			fundingTxId = ldkChannel.FundingTxo.Txid
		}

		internalChannel := map[string]interface{}{}
		internalChannel["channel"] = ldkChannel
		internalChannel["config"] = map[string]interface{}{
			"AcceptUnderpayingHtlcs":              ldkChannel.Config.AcceptUnderpayingHtlcs(),
			"CltvExpiryDelta":                     ldkChannel.Config.CltvExpiryDelta(),
			"ForceCloseAvoidanceMaxFeeSatoshis":   ldkChannel.Config.ForceCloseAvoidanceMaxFeeSatoshis(),
			"ForwardingFeeBaseMsat":               ldkChannel.Config.ForwardingFeeBaseMsat(),
			"ForwardingFeeProportionalMillionths": ldkChannel.Config.ForwardingFeeProportionalMillionths(),
		}

		unspendablePunishmentReserve := uint64(0)
		if ldkChannel.UnspendablePunishmentReserve != nil {
			unspendablePunishmentReserve = *ldkChannel.UnspendablePunishmentReserve
		}

		var channelError *string

		if fundingTxId == "" {
			channelErrorValue := "This channel has no funding transaction. Please contact support@getalby.com"
			channelError = &channelErrorValue
		} else if ldkChannel.IsUsable && ldkChannel.CounterpartyForwardingInfoFeeBaseMsat == nil {
			// if we don't have this, routing will not work (LND <-> LDK interoperability bug - https://github.com/lightningnetwork/lnd/issues/6870 )
			channelErrorValue := "Counterparty forwarding info not available. Please contact support@getalby.com"
			channelError = &channelErrorValue
		}

		isActive := ldkChannel.IsUsable /* superset of ldkChannel.IsReady */ && channelError == nil

		channels = append(channels, lnclient.Channel{
			InternalChannel:                          internalChannel,
			LocalBalance:                             int64(ldkChannel.ChannelValueSats*1000 - ldkChannel.InboundCapacityMsat - ldkChannel.CounterpartyUnspendablePunishmentReserve*1000),
			LocalSpendableBalance:                    int64(ldkChannel.OutboundCapacityMsat),
			RemoteBalance:                            int64(ldkChannel.InboundCapacityMsat),
			RemotePubkey:                             ldkChannel.CounterpartyNodeId,
			Id:                                       ldkChannel.UserChannelId, // CloseChannel takes the UserChannelId
			Active:                                   isActive,
			Public:                                   ldkChannel.IsPublic,
			FundingTxId:                              fundingTxId,
			Confirmations:                            ldkChannel.Confirmations,
			ConfirmationsRequired:                    ldkChannel.ConfirmationsRequired,
			ForwardingFeeBaseMsat:                    ldkChannel.Config.ForwardingFeeBaseMsat(),
			UnspendablePunishmentReserve:             unspendablePunishmentReserve,
			CounterpartyUnspendablePunishmentReserve: ldkChannel.CounterpartyUnspendablePunishmentReserve,
			Error:                                    channelError,
			IsOutbound:                               ldkChannel.IsOutbound,
		})
	}

	return channels, nil
}

func (ls *LDKService) GetNodeConnectionInfo(ctx context.Context) (nodeConnectionInfo *lnclient.NodeConnectionInfo, err error) {
	/*addresses := gs.node.ListeningAddresses()
	if addresses == nil || len(*addresses) < 1 {
		return nil, errors.New("no available listening addresses")
	}
	firstAddress := (*addresses)[0]
	parts := strings.Split(firstAddress, ":")
	if len(parts) != 2 {
		return nil, errors.New(fmt.Sprintf("invalid address %v", firstAddress))
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		logger.Logger.WithError(err).Error("ConnectPeer failed")
		return nil, err
	}*/

	return &lnclient.NodeConnectionInfo{
		Pubkey: ls.node.NodeId(),
		//Address: parts[0],
		//Port:    port,
	}, nil
}

func (ls *LDKService) ConnectPeer(ctx context.Context, connectPeerRequest *lnclient.ConnectPeerRequest) error {
	err := ls.node.Connect(connectPeerRequest.Pubkey, connectPeerRequest.Address+":"+strconv.Itoa(int(connectPeerRequest.Port)), true)
	if err != nil {
		logger.Logger.WithField("request", connectPeerRequest).WithError(err).Error("ConnectPeer failed")
		return err
	}

	return nil
}

func (ls *LDKService) OpenChannel(ctx context.Context, openChannelRequest *lnclient.OpenChannelRequest) (*lnclient.OpenChannelResponse, error) {
	peers := ls.node.ListPeers()
	var foundPeer *ldk_node.PeerDetails
	for _, peer := range peers {
		if peer.NodeId == openChannelRequest.Pubkey {

			foundPeer = &peer
			break
		}
	}

	if foundPeer == nil {
		return nil, errors.New("node is not peered yet")
	}

	ldkEventSubscription := ls.ldkEventBroadcaster.Subscribe()
	defer ls.ldkEventBroadcaster.CancelSubscription(ldkEventSubscription)

	logger.Logger.WithField("peer_id", foundPeer.NodeId).Info("Opening channel")
	userChannelId, err := ls.node.ConnectOpenChannel(foundPeer.NodeId, foundPeer.Address, uint64(openChannelRequest.Amount), nil, nil, openChannelRequest.Public)
	if err != nil {
		logger.Logger.WithError(err).Error("OpenChannel failed")
		return nil, err
	}

	// userChannelId allows to locally keep track of the channel (and is also used to close the channel)
	logger.Logger.WithFields(logrus.Fields{
		"peer_id":    foundPeer.NodeId,
		"channel_id": userChannelId,
	}).Info("Funded channel")

	for start := time.Now(); time.Since(start) < time.Second*60; {
		event := <-ldkEventSubscription

		channelPendingEvent, isChannelPendingEvent := (*event).(ldk_node.EventChannelPending)
		channelClosedEvent, isChannelClosedEvent := (*event).(ldk_node.EventChannelClosed)

		if isChannelClosedEvent {
			closureReason := ls.getChannelCloseReason(&channelClosedEvent)
			logger.Logger.WithFields(logrus.Fields{
				"event":  channelClosedEvent,
				"reason": closureReason,
			}).Info("Failed to open channel")

			return nil, fmt.Errorf("failed to open channel with %s: %s", foundPeer.NodeId, closureReason)
		}

		if !isChannelPendingEvent {
			continue
		}

		return &lnclient.OpenChannelResponse{
			FundingTxId: channelPendingEvent.FundingTxo.Txid,
		}, nil
	}

	return nil, errors.New("open channel timeout")
}

func (ls *LDKService) UpdateChannel(ctx context.Context, updateChannelRequest *lnclient.UpdateChannelRequest) error {
	channels := ls.node.ListChannels()

	var foundChannel *ldk_node.ChannelDetails
	for _, channel := range channels {
		if channel.UserChannelId == updateChannelRequest.ChannelId && channel.CounterpartyNodeId == updateChannelRequest.NodeId {
			foundChannel = &channel
			break
		}
	}

	if foundChannel == nil {
		logger.Logger.WithField("request", updateChannelRequest).Error("failed to find channel to update")
		return errors.New("channel not found")
	}

	existingConfig := foundChannel.Config
	existingConfig.SetForwardingFeeBaseMsat(updateChannelRequest.ForwardingFeeBaseMsat)

	err := ls.node.UpdateChannelConfig(updateChannelRequest.ChannelId, updateChannelRequest.NodeId, existingConfig)
	if err != nil {
		logger.Logger.WithError(err).Error("UpdateChannelConfig failed")
		return err
	}
	return nil
}

func (ls *LDKService) CloseChannel(ctx context.Context, closeChannelRequest *lnclient.CloseChannelRequest) (*lnclient.CloseChannelResponse, error) {
	logger.Logger.WithFields(logrus.Fields{
		"request": closeChannelRequest,
	}).Info("Closing Channel")

	var err error
	if closeChannelRequest.Force {
		err = ls.node.ForceCloseChannel(closeChannelRequest.ChannelId, closeChannelRequest.NodeId)
	} else {
		err = ls.node.CloseChannel(closeChannelRequest.ChannelId, closeChannelRequest.NodeId)
	}
	if err != nil {
		logger.Logger.WithError(err).Error("CloseChannel failed")
		return nil, err
	}
	return &lnclient.CloseChannelResponse{}, nil
}

func (ls *LDKService) GetNewOnchainAddress(ctx context.Context) (string, error) {
	address, err := ls.node.OnchainPayment().NewAddress()
	if err != nil {
		logger.Logger.WithError(err).Error("NewOnchainAddress failed")
		return "", err
	}
	return address, nil
}

func (ls *LDKService) GetOnchainBalance(ctx context.Context) (*lnclient.OnchainBalanceResponse, error) {
	channels := ls.node.ListChannels()
	balances := ls.node.ListBalances()
	logger.Logger.WithFields(logrus.Fields{
		"balances": balances,
	}).Debug("Listed Balances")

	pendingBalancesFromChannelClosures := uint64(0)
	// increase pending balance from any lightning balances for channels that are pending closure
	// (they do not exist in our list of open channels)
	for _, balance := range balances.LightningBalances {
		increasePendingBalance := func(channelId string, amount uint64) {
			if !slices.ContainsFunc(channels, func(channel ldk_node.ChannelDetails) bool {
				return channel.ChannelId == channelId
			}) {
				pendingBalancesFromChannelClosures += amount
			}
		}

		switch balanceType := (balance).(type) {
		case ldk_node.LightningBalanceClaimableOnChannelClose:
			increasePendingBalance(balanceType.ChannelId, balanceType.AmountSatoshis)
		case ldk_node.LightningBalanceClaimableAwaitingConfirmations:
			increasePendingBalance(balanceType.ChannelId, balanceType.AmountSatoshis)
		case ldk_node.LightningBalanceContentiousClaimable:
			increasePendingBalance(balanceType.ChannelId, balanceType.AmountSatoshis)
		case ldk_node.LightningBalanceMaybeTimeoutClaimableHtlc:
			increasePendingBalance(balanceType.ChannelId, balanceType.AmountSatoshis)
		case ldk_node.LightningBalanceMaybePreimageClaimableHtlc:
			increasePendingBalance(balanceType.ChannelId, balanceType.AmountSatoshis)
		case ldk_node.LightningBalanceCounterpartyRevokedOutputClaimable:
			increasePendingBalance(balanceType.ChannelId, balanceType.AmountSatoshis)
		}
	}

	// increase pending balance from any lightning balances for channels that were closed
	for _, balance := range balances.PendingBalancesFromChannelClosures {
		switch pendingType := (balance).(type) {
		case ldk_node.PendingSweepBalancePendingBroadcast:
			pendingBalancesFromChannelClosures += pendingType.AmountSatoshis
		case ldk_node.PendingSweepBalanceBroadcastAwaitingConfirmation:
			pendingBalancesFromChannelClosures += pendingType.AmountSatoshis
		case ldk_node.PendingSweepBalanceAwaitingThresholdConfirmations:
			pendingBalancesFromChannelClosures += pendingType.AmountSatoshis
		}
	}

	return &lnclient.OnchainBalanceResponse{
		Spendable:                          int64(balances.SpendableOnchainBalanceSats),
		Total:                              int64(balances.TotalOnchainBalanceSats - balances.TotalAnchorChannelsReserveSats),
		Reserved:                           int64(balances.TotalAnchorChannelsReserveSats),
		PendingBalancesFromChannelClosures: pendingBalancesFromChannelClosures,
	}, nil
}

func (ls *LDKService) RedeemOnchainFunds(ctx context.Context, toAddress string, amount uint64, sendAll bool) (string, error) {
	if !sendAll {
		// NOTE: this may fail if user does not reserve enough for the onchain transaction
		// and can also drain the anchor reserves if the user provides a too high amount.
		txId, err := ls.node.OnchainPayment().SendToAddress(toAddress, amount)
		if err != nil {
			logger.Logger.WithError(err).Error("SendToAddress failed")
			return "", err
		}
		return txId, nil
	}

	// TODO: this could be improved to preserve anchor reserves once LDK supports this
	txId, err := ls.node.OnchainPayment().SendAllToAddress(toAddress)
	if err != nil {
		logger.Logger.WithError(err).Error("SendAllToAddress failed")
		return "", err
	}
	return txId, nil
}

func (ls *LDKService) ResetRouter(key string) error {
	ls.cfg.SetUpdate(resetRouterKey, key, "")

	return nil
}

func (ls *LDKService) SignMessage(ctx context.Context, message string) (string, error) {
	sign, err := ls.node.SignMessage([]byte(message))
	if err != nil {
		logger.Logger.Errorf("SignMessage failed: %v", err)
		return "", err
	}

	return sign, nil
}

func (ls *LDKService) ldkPaymentToTransaction(payment *ldk_node.PaymentDetails) (*lnclient.Transaction, error) {
	// logger.Logger.WithField("payment", payment).Debug("Mapping LDK payment to transaction")

	transactionType := "incoming"
	if payment.Direction == ldk_node.PaymentDirectionOutbound {
		transactionType = "outgoing"
	}

	var expiresAt *int64
	var createdAt int64
	var description string
	var descriptionHash string
	var bolt11Invoice string
	var settledAt *int64
	preimage := ""
	paymentHash := ""
	metadata := map[string]interface{}{}

	bolt11PaymentKind, isBolt11PaymentKind := payment.Kind.(ldk_node.PaymentKindBolt11)

	if isBolt11PaymentKind && bolt11PaymentKind.Bolt11Invoice != nil {
		bolt11Invoice = *bolt11PaymentKind.Bolt11Invoice
		paymentRequest, err := decodepay.Decodepay(strings.ToLower(bolt11Invoice))
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"bolt11": bolt11Invoice,
			}).Errorf("Failed to decode bolt11 invoice: %v", err)

			return nil, err
		}
		createdAt = int64(paymentRequest.CreatedAt)
		expiresAtUnix := time.UnixMilli(int64(paymentRequest.CreatedAt) * 1000).Add(time.Duration(paymentRequest.Expiry) * time.Second).Unix()
		expiresAt = &expiresAtUnix
		description = paymentRequest.Description
		descriptionHash = paymentRequest.DescriptionHash
		if payment.Status == ldk_node.PaymentStatusSucceeded {
			if bolt11PaymentKind.Preimage != nil {
				preimage = *bolt11PaymentKind.Preimage
			}
			settledAt = &createdAt // fallback settledAt to created at time
			if payment.LastUpdate > 0 {
				lastUpdate := int64(payment.LastUpdate)
				settledAt = &lastUpdate
			}
		}
		paymentHash = bolt11PaymentKind.Hash
	}

	spontaneousPaymentKind, isSpontaneousPaymentKind := payment.Kind.(ldk_node.PaymentKindSpontaneous)
	if isSpontaneousPaymentKind {
		// keysend payment
		lastUpdate := int64(payment.LastUpdate)
		createdAt = int64(payment.CreatedAt)
		// TODO: remove this check some point in the future
		// all payments after v0.6.2 will have createdAt set
		if createdAt == 0 {
			createdAt = lastUpdate
		}
		if payment.Status == ldk_node.PaymentStatusSucceeded {
			settledAt = &lastUpdate
		}
		paymentHash = spontaneousPaymentKind.Hash
		if spontaneousPaymentKind.Preimage != nil {
			preimage = *spontaneousPaymentKind.Preimage
		}

		tlvRecords := []lnclient.TLVRecord{}
		for _, tlv := range spontaneousPaymentKind.CustomTlvs {
			tlvRecords = append(tlvRecords, lnclient.TLVRecord{
				Type:  tlv.Type,
				Value: hex.EncodeToString(tlv.Value),
			})
		}
		metadata["tlv_records"] = tlvRecords
	}

	var amount uint64 = 0
	if payment.AmountMsat != nil {
		amount = *payment.AmountMsat
	}

	var fee uint64 = 0
	if payment.FeeMsat != nil {
		fee = *payment.FeeMsat
	}

	return &lnclient.Transaction{
		Type:            transactionType,
		Preimage:        preimage,
		PaymentHash:     paymentHash,
		SettledAt:       settledAt,
		Amount:          int64(amount),
		Invoice:         bolt11Invoice,
		FeesPaid:        int64(fee),
		CreatedAt:       createdAt,
		Description:     description,
		DescriptionHash: descriptionHash,
		ExpiresAt:       expiresAt,
		Metadata:        metadata,
	}, nil
}

func (ls *LDKService) SendPaymentProbes(ctx context.Context, invoice string) error {
	err := ls.node.Bolt11Payment().SendProbes(invoice)
	if err != nil {
		logger.Logger.Errorf("Bolt11Payment.SendProbes failed: %v", err)
		return err
	}

	return nil
}

func (ls *LDKService) SendSpontaneousPaymentProbes(ctx context.Context, amountMsat uint64, nodeId string) error {
	err := ls.node.SpontaneousPayment().SendProbes(amountMsat, nodeId)
	if err != nil {
		logger.Logger.Errorf("SpontaneousPayment.SendProbes failed: %v", err)
		return err
	}

	return nil
}

func (ls *LDKService) ListPeers(ctx context.Context) ([]lnclient.PeerDetails, error) {
	peers := ls.node.ListPeers()
	ret := make([]lnclient.PeerDetails, 0, len(peers))
	for _, peer := range peers {
		ret = append(ret, lnclient.PeerDetails{
			NodeId:      peer.NodeId,
			Address:     peer.Address,
			IsPersisted: peer.IsPersisted,
			IsConnected: peer.IsConnected,
		})
	}
	return ret, nil
}

func (ls *LDKService) GetNetworkGraph(ctx context.Context, nodeIds []string) (lnclient.NetworkGraphResponse, error) {
	graph := ls.node.NetworkGraph()

	type NodeInfoWithId struct {
		Node   *ldk_node.NodeInfo `json:"node"`
		NodeId string             `json:"nodeId"`
	}

	nodes := []NodeInfoWithId{}
	channels := []*ldk_node.ChannelInfo{}
	for _, nodeId := range nodeIds {
		_, err := hex.DecodeString(nodeId)
		if err != nil {
			return nil, err
		}
		if len(nodeId) != 66 {
			return nil, errors.New("unexpected node ID length")
		}
		graphNode := graph.Node(nodeId)
		if graphNode != nil {
			nodes = append(nodes, NodeInfoWithId{
				Node:   graphNode,
				NodeId: nodeId,
			})
		}
		for _, channelId := range graphNode.Channels {
			graphChannel := graph.Channel(channelId)
			if graphChannel != nil {
				channels = append(channels, graphChannel)
			}
		}
	}

	networkGraph := map[string]interface{}{
		"nodes":    nodes,
		"channels": channels,
	}
	return networkGraph, nil
}

func (ls *LDKService) GetLogOutput(ctx context.Context, maxLen int) ([]byte, error) {
	config := ls.node.Config()
	logPath := ""
	if config.LogDirPath != nil {
		logPath = *config.LogDirPath
	} else {
		// Default log path if not set explicitly in the config.
		logPath = filepath.Join(config.StorageDirPath, "logs")
	}

	allLogFiles, err := filepath.Glob(filepath.Join(logPath, "ldk_node_*.log"))
	if err != nil {
		logger.Logger.WithError(err).Error("GetLogOutput failed to list log files")
		return nil, err
	}

	if len(allLogFiles) == 0 {
		return []byte{}, nil
	}

	// Log filenames are formatted as ldk_node_YYYY_MM_DD.log, hence they
	// naturally sort by date.
	lastLogFileName := slices.Max(allLogFiles)

	logData, err := utils.ReadFileTail(lastLogFileName, maxLen)
	if err != nil {
		logger.Logger.WithError(err).Error("GetLogOutput failed to read log file")
		return nil, err
	}

	return logData, nil
}

func (ls *LDKService) handleLdkEvent(event *ldk_node.Event) {
	logger.Logger.WithFields(logrus.Fields{
		"event": event,
	}).Info("Received LDK event")

	switch eventType := (*event).(type) {
	case ldk_node.EventChannelReady:
		channels := ls.node.ListChannels()
		channelIndex := slices.IndexFunc(channels, func(c ldk_node.ChannelDetails) bool {
			return c.ChannelId == eventType.ChannelId
		})
		if channelIndex == -1 {
			logger.Logger.WithField("event", eventType).Error("Failed to find channel by ID")
			return
		}
		channel := channels[channelIndex]
		ls.eventPublisher.Publish(&events.Event{
			Event: "nwc_channel_ready",
			Properties: map[string]interface{}{
				"counterparty_node_id": eventType.CounterpartyNodeId,
				"node_type":            config.LDKBackendType,
				"public":               channel.IsPublic,
				"capacity":             channel.ChannelValueSats,
				"is_outbound":          channel.IsOutbound,
			},
		})

		ls.backupChannels()

		if eventType.CounterpartyNodeId == nil {
			logger.Logger.WithField("event", eventType).Error("channel ready event has no counterparty node ID")
			return
		}
		// set a super-high forwarding fee of 100K sats by default to disable unwanted routing
		err := ls.UpdateChannel(context.Background(), &lnclient.UpdateChannelRequest{
			ChannelId:             eventType.UserChannelId,
			NodeId:                *eventType.CounterpartyNodeId,
			ForwardingFeeBaseMsat: 100_000_000,
		})

		if err != nil {
			logger.Logger.WithField("event", eventType).Error("channel ready event has no counterparty node ID")
			return
		}

	case ldk_node.EventChannelClosed:
		closureReason := ls.getChannelCloseReason(&eventType)
		logger.Logger.WithFields(logrus.Fields{
			"event":  event,
			"reason": closureReason,
		}).Info("Channel closed")

		ls.eventPublisher.Publish(&events.Event{
			Event: "nwc_channel_closed",
			Properties: map[string]interface{}{
				"counterparty_node_id": eventType.CounterpartyNodeId,
				"reason":               closureReason,
				"node_type":            config.LDKBackendType,
			},
		})
	case ldk_node.EventPaymentReceived:
		if eventType.PaymentId == nil {
			logger.Logger.WithField("payment_hash", eventType.PaymentHash).Error("payment received event has no payment ID")
			return
		}
		payment := ls.node.Payment(*eventType.PaymentId)
		if payment == nil {
			logger.Logger.WithField("payment_id", *eventType.PaymentId).Error("could not find LDK payment")
			return
		}

		transaction, err := ls.ldkPaymentToTransaction(payment)
		if err != nil {
			logger.Logger.WithField("payment_id", *eventType.PaymentId).Error("failed to convert LDK payment to transaction")
			return
		}

		ls.eventPublisher.Publish(&events.Event{
			Event:      "nwc_lnclient_payment_received",
			Properties: transaction,
		})
	case ldk_node.EventPaymentSuccessful:
		if eventType.PaymentId == nil {
			logger.Logger.WithField("payment_hash", eventType.PaymentHash).Error("payment received event has no payment ID")
			return
		}
		payment := ls.node.Payment(*eventType.PaymentId)
		if payment == nil {
			logger.Logger.WithField("payment_id", *eventType.PaymentId).Error("could not find LDK payment")
			return
		}

		transaction, err := ls.ldkPaymentToTransaction(payment)
		if err != nil {
			logger.Logger.WithField("payment_id", *eventType.PaymentId).Error("failed to convert LDK payment to transaction")
			return
		}

		ls.eventPublisher.Publish(&events.Event{
			Event:      "nwc_lnclient_payment_sent",
			Properties: transaction,
		})
	case ldk_node.EventPaymentFailed:
		if eventType.PaymentId == nil {
			logger.Logger.WithField("payment_hash", eventType.PaymentHash).Error("payment failed event has no payment ID")
			return
		}
		payment := ls.node.Payment(*eventType.PaymentId)
		if payment == nil {
			logger.Logger.WithField("payment_id", *eventType.PaymentId).Error("could not find LDK payment")
			return
		}

		transaction, err := ls.ldkPaymentToTransaction(payment)
		if err != nil {
			logger.Logger.WithField("payment_id", *eventType.PaymentId).Error("failed to convert LDK payment to transaction")
			return
		}

		reason := ls.getPaymentFailReason(&eventType)

		ls.eventPublisher.Publish(&events.Event{
			Event: "nwc_lnclient_payment_failed",
			Properties: &lnclient.PaymentFailedEventProperties{
				Transaction: transaction,
				Reason:      reason,
			},
		})
	}
}

func (ls *LDKService) backupChannels() {
	ldkChannels := ls.node.ListChannels()
	channels := make([]events.ChannelBackupInfo, 0, len(ldkChannels))
	for _, ldkChannel := range ldkChannels {
		var fundingTxId string
		var fundingTxVout uint32
		if ldkChannel.FundingTxo != nil {
			fundingTxId = ldkChannel.FundingTxo.Txid
			fundingTxVout = ldkChannel.FundingTxo.Vout
		}

		channels = append(channels, events.ChannelBackupInfo{
			ChannelID:     ldkChannel.ChannelId,
			NodeID:        ls.node.NodeId(),
			PeerID:        ldkChannel.CounterpartyNodeId,
			ChannelSize:   ldkChannel.ChannelValueSats,
			FundingTxID:   fundingTxId,
			FundingTxVout: fundingTxVout,
		})
	}

	monitors, err := ls.node.GetEncodedChannelMonitors()
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to list channel monitors")
		return
	}
	encodedMonitors := []events.ChannelMonitorBackup{}

	for _, monitor := range monitors {
		encodedMonitors = append(encodedMonitors, events.ChannelMonitorBackup{
			Key:   monitor.Key,
			Value: hex.EncodeToString(monitor.Value),
		})
	}

	event := &events.ChannelBackupEvent{
		Channels: channels,
		Monitors: encodedMonitors,
	}

	ls.saveStaticChannelBackupToDisk(event)

	ls.eventPublisher.Publish(&events.Event{
		Event:      "nwc_backup_channels",
		Properties: event,
	})
}

func (ls *LDKService) saveStaticChannelBackupToDisk(event *events.ChannelBackupEvent) {
	backupDirectory := filepath.Join(ls.workdir, "static_channel_backups")
	err := os.MkdirAll(backupDirectory, os.ModePerm)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to make static channel backup directory")
		return
	}

	backupFilePath := filepath.Join(backupDirectory, time.Now().Format("2006-01-02T15-04-05")+".json")
	eventBytes, err := json.Marshal(event)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to serialize static channel backup to json")
		return
	}
	err = os.WriteFile(backupFilePath, eventBytes, 0644)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to write static channel backup to disk")
		return
	}
	logger.Logger.WithField("backupPath", backupFilePath).Debug("Saved static channel backup to disk")
}

func (ls *LDKService) GetBalances(ctx context.Context) (*lnclient.BalancesResponse, error) {
	onchainBalance, err := ls.GetOnchainBalance(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to retrieve onchain balance")
		return nil, err
	}

	var totalReceivable int64 = 0
	var totalSpendable int64 = 0
	var nextMaxReceivable int64 = 0
	var nextMaxSpendable int64 = 0
	var nextMaxReceivableMPP int64 = 0
	var nextMaxSpendableMPP int64 = 0
	channels := ls.node.ListChannels()
	for _, channel := range channels {
		if channel.IsUsable {
			// spending or receiving amount may be constrained by channel configuration (e.g. ACINQ does this)
			channelConstrainedSpendable := min(int64(channel.OutboundCapacityMsat), int64(*channel.CounterpartyOutboundHtlcMaximumMsat))
			channelConstrainedReceivable := min(int64(channel.InboundCapacityMsat), int64(*channel.InboundHtlcMaximumMsat))

			nextMaxSpendable = max(nextMaxSpendable, channelConstrainedSpendable)
			nextMaxReceivable = max(nextMaxReceivable, channelConstrainedReceivable)

			nextMaxSpendableMPP += channelConstrainedSpendable
			nextMaxReceivableMPP += channelConstrainedReceivable

			// these are what the wallet can send and receive, but not necessarily in one go
			totalSpendable += int64(channel.OutboundCapacityMsat)
			totalReceivable += int64(channel.InboundCapacityMsat)
		}
	}

	return &lnclient.BalancesResponse{
		Onchain: *onchainBalance,
		Lightning: lnclient.LightningBalanceResponse{
			TotalSpendable:       totalSpendable,
			TotalReceivable:      totalReceivable,
			NextMaxSpendable:     nextMaxSpendable,
			NextMaxReceivable:    nextMaxReceivable,
			NextMaxSpendableMPP:  nextMaxSpendableMPP,
			NextMaxReceivableMPP: nextMaxReceivableMPP,
		},
	}, nil
}

func (ls *LDKService) GetStorageDir() (string, error) {
	// Note: the below will return the path including the WORK_DIR which is harder to use,
	// so for now we just return a hardcoded value.
	// cfg := ls.node.Config()
	// return cfg.StorageDirPath, nil
	return "ldk/storage", nil
}

func deleteOldLDKLogs(ldkLogDir string) {
	logger.Logger.WithField("ldkLogDir", ldkLogDir).Debug("Deleting old LDK logs")
	files, err := os.ReadDir(ldkLogDir)
	if err != nil {
		logger.Logger.WithField("path", ldkLogDir).WithError(err).Error("Failed to list ldk log directory")
		return
	}

	for _, file := range files {
		// get files with a date (e.g. ldk_node_2024_03_29.log)
		if strings.HasPrefix(file.Name(), "ldk_node_2") && strings.HasSuffix(file.Name(), ".log") {
			filePath := filepath.Join(ldkLogDir, file.Name())
			fileInfo, err := file.Info()
			if err != nil {
				logger.Logger.WithField("filePath", filePath).WithError(err).Error("Failed to get file info")
				continue
			}
			// delete files last modified over 3 days ago
			if fileInfo.ModTime().Before(time.Now().AddDate(0, 0, -3)) {
				err := os.Remove(filePath)
				if err != nil {
					logger.Logger.WithField("filePath", filePath).WithError(err).Error("Failed to get file info")
					continue
				}
				logger.Logger.WithField("filePath", filePath).Info("Deleted old LDK log file")
			}
		}
	}
}

func (ls *LDKService) GetNodeStatus(ctx context.Context) (nodeStatus *lnclient.NodeStatus, err error) {
	return &lnclient.NodeStatus{
		InternalNodeStatus: ls.node.Status(),
	}, nil
}

func (ls *LDKService) DisconnectPeer(ctx context.Context, peerId string) error {
	return ls.node.Disconnect(peerId)
}

func (ls *LDKService) UpdateLastWalletSyncRequest() {
	ls.lastWalletSyncRequest = time.Now()
}

func (ls *LDKService) GetSupportedNIP47Methods() []string {
	return []string{"pay_invoice", "pay_keysend", "get_balance", "get_info", "make_invoice", "lookup_invoice", "list_transactions", "multi_pay_invoice", "multi_pay_keysend", "sign_message"}
}

func (ls *LDKService) GetSupportedNIP47NotificationTypes() []string {
	return []string{"payment_received", "payment_sent"}
}

func (ls *LDKService) getPaymentFailReason(eventPaymentFailed *ldk_node.EventPaymentFailed) string {
	var failureReason ldk_node.PaymentFailureReason
	var failureReasonMessage string
	if eventPaymentFailed.Reason != nil {
		failureReason = *eventPaymentFailed.Reason
	}
	switch failureReason {
	case ldk_node.PaymentFailureReasonRecipientRejected:
		failureReasonMessage = "RecipientRejected"
	case ldk_node.PaymentFailureReasonUserAbandoned:
		failureReasonMessage = "UserAbandoned"
	case ldk_node.PaymentFailureReasonRetriesExhausted:
		failureReasonMessage = "RetriesExhausted"
	case ldk_node.PaymentFailureReasonPaymentExpired:
		failureReasonMessage = "PaymentExpired"
	case ldk_node.PaymentFailureReasonRouteNotFound:
		failureReasonMessage = "RouteNotFound"
	case ldk_node.PaymentFailureReasonUnexpectedError:
		failureReasonMessage = "UnexpectedError"
	default:
		failureReasonMessage = "UnknownError"
	}
	return failureReasonMessage
}

func (ls *LDKService) getChannelCloseReason(event *ldk_node.EventChannelClosed) string {
	var reason string

	switch reasonType := (*event.Reason).(type) {
	case ldk_node.ClosureReasonCounterpartyForceClosed:
		reason = fmt.Sprintf("CounterpartyForceClosed (Peer message: %s)", reasonType.PeerMsg)
	case ldk_node.ClosureReasonHolderForceClosed:
		reason = "HolderForceClosed"
	case ldk_node.ClosureReasonLegacyCooperativeClosure:
		reason = "LegacyCooperativeClosure"
	case ldk_node.ClosureReasonCounterpartyInitiatedCooperativeClosure:
		reason = "CounterpartyInitiatedCooperativeClosure"
	case ldk_node.ClosureReasonLocallyInitiatedCooperativeClosure:
		reason = "LocallyInitiatedCooperativeClosure"
	case ldk_node.ClosureReasonCommitmentTxConfirmed:
		reason = "CommitmentTxConfirmed"
	case ldk_node.ClosureReasonFundingTimedOut:
		reason = "FundingTimedOut"
	case ldk_node.ClosureReasonProcessingError:
		reason = fmt.Sprintf("ProcessingError: %s", reasonType.Err)
	case ldk_node.ClosureReasonDisconnectedPeer:
		reason = "DisconnectedPeer"
	case ldk_node.ClosureReasonOutdatedChannelManager:
		reason = "OutdatedChannelManager"
	case ldk_node.ClosureReasonCounterpartyCoopClosedUnfundedChannel:
		reason = "CounterpartyCoopClosedUnfundedChannel"
	case ldk_node.ClosureReasonFundingBatchClosure:
		reason = "FundingBatchClosure"
	case ldk_node.ClosureReasonHtlCsTimedOut:
		reason = "HTLCsTimedOut"
	default:
		reason = fmt.Sprintf("Unknown: %s", *event.Reason)
	}

	return reason
}

func (ls *LDKService) GetPubkey() string {
	return ls.pubkey
}
