package ldk

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/getAlby/ldk-node-go/ldk_node"
	"github.com/tyler-smith/go-bip32"

	// "github.com/getAlby/hub/ldk_node"

	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/lsp"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/nip47/notifications"
	"github.com/getAlby/hub/service/keys"
	"github.com/getAlby/hub/transactions"
)

type LDKService struct {
	workdir                            string
	node                               *ldk_node.Node
	ldkEventBroadcaster                LDKEventBroadcaster
	cancel                             context.CancelFunc
	ctx                                context.Context
	network                            string
	eventPublisher                     events.EventPublisher
	syncing                            bool
	lastFullSync                       time.Time
	lastFeeEstimatesSync               time.Time
	cfg                                config.Config
	lastWalletSyncRequest              time.Time
	redeemedOnchainFundsWithinThisSync bool
	pubkey                             string
	shuttingDown                       bool
}

const resetRouterKey = "ResetRouter"
const maxInvoiceExpiry = 24 * time.Hour

func NewLDKService(ctx context.Context, cfg config.Config, eventPublisher events.EventPublisher, mnemonic, workDir string, vssToken string, setStartupState func(startupState string)) (result lnclient.LNClient, err error) {
	if mnemonic == "" || workDir == "" {
		return nil, errors.New("one or more required LDK configuration are missing")
	}

	setStartupState("Configuring node")

	// create dir if not exists
	newpath := filepath.Join(workDir)
	err = os.MkdirAll(newpath, os.ModePerm)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to create LDK working dir")
		return nil, err
	}

	ldkConfig := ldk_node.DefaultConfig()

	ldkConfig.TrustedPeers0conf = []string{
		lsp.OlympusLSP().Pubkey,
		lsp.MegalithLSP().Pubkey,
		"02b4552a7a85274e4da01a7c71ca57407181752e8568b31d51f13c111a2941dce3", // LNServer_Wave
		"038ba8f67ba8ff5c48764cdd3251c33598d55b203546d08a8f0ec9dcd9f27e3637", // Flashsats

		// Mutinynet
		lsp.OlympusMutinynetLSP().Pubkey,
		lsp.MegalithMutinynetLSP().Pubkey,
		"03f726f240f0391448fb31c33e130ecc9708c9137e1f4e77b5d17d5dec74b0dd1e", // flashsats
	}

	// rather than fully trusting our LSPs, we set the channel reserve to 0.
	// this allows us to receive incoming channels without any on-chain balance
	// but if the user has 0 on-chain balance when the channel is closed,
	// we rely on the counterparty to bump the transaction.
	// It's also possible in rare situations the counterparty can take
	// funds if the channel was closed due to a stuck HTLC.
	// Therefore, the user SHOULD add some on-chain funds to prevent this.
	ldkConfig.AnchorChannelsConfig.PerChannelReserveSats = 0
	ldkConfig.AnchorChannelsConfig.TrustedPeersNoReserve = []string{
		/*lsp.OlympusLSP().Pubkey,
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
		"035e8a9034a8c68f219aacadae748c7a3cd719109309db39b09886e5ff17696b1b", // lqwd*/
	}

	listeningAddresses := strings.Split(cfg.GetEnv().LDKListeningAddresses, ",")
	ldkConfig.ListeningAddresses = &listeningAddresses
	if cfg.GetEnv().LDKAnnouncementAddresses != "" {
		announcementAddresses := strings.Split(cfg.GetEnv().LDKAnnouncementAddresses, ",")
		ldkConfig.AnnouncementAddresses = &announcementAddresses
	}

	logLevel, err := strconv.Atoi(cfg.GetEnv().LDKLogLevel)
	if err != nil {
		// If parsing log level fails we default to info log level
		logLevel = int(logrus.InfoLevel)
	}

	ldkLogger, err := NewLDKLogger(logrus.Level(logLevel), cfg.GetEnv().LogToFile, workDir)
	if err != nil {
		return nil, err
	}
	ldkConfig.TransientNetworkGraph = cfg.GetEnv().LDKTransientNetworkGraph

	alias, _ := cfg.Get("NodeAlias", "")
	if alias == "" {
		alias = "Alby Hub"
	}

	builder := ldk_node.BuilderFromConfig(ldkConfig)
	builder.SetCustomLogger(ldkLogger)
	builder.SetNodeAlias(alias)
	builder.SetEntropyBip39Mnemonic(mnemonic, nil)

	network := cfg.GetNetwork()
	switch network {
	case "signet":
		builder.SetNetwork(ldk_node.NetworkSignet)
	case "regtest":
		builder.SetNetwork(ldk_node.NetworkRegtest)
	case "testnet":
		builder.SetNetwork(ldk_node.NetworkSignet)
	default:
		builder.SetNetwork(ldk_node.NetworkBitcoin)
	}

	var chainSource string
	if cfg.GetEnv().LDKBitcoindRpcHost != "" {
		logger.Logger.WithFields(logrus.Fields{
			"rpc_host": cfg.GetEnv().LDKBitcoindRpcHost,
			"rpc_port": cfg.GetEnv().LDKBitcoindRpcPort,
		}).Info("Using LDK node bitcoin RPC chain source")
		port, err := strconv.ParseUint(cfg.GetEnv().LDKBitcoindRpcPort, 10, 16)
		if err != nil {
			return nil, err
		}
		builder.SetChainSourceBitcoindRpc(cfg.GetEnv().LDKBitcoindRpcHost, uint16(port), cfg.GetEnv().LDKBitcoindRpcUser, cfg.GetEnv().LDKBitcoindRpcPassword)
		chainSource = "bitcoind_rpc"
	} else if cfg.GetEnv().LDKElectrumServer != "" {
		builder.SetChainSourceElectrum(cfg.GetEnv().LDKElectrumServer, &ldk_node.ElectrumSyncConfig{
			// turn off background sync - we manage syncs ourselves
			BackgroundSyncConfig: nil,
		})
		chainSource = "electrum"
	} else {
		logger.Logger.WithFields(logrus.Fields{
			"esplora_url": cfg.GetEnv().LDKEsploraServer,
		}).Info("Using LDK node esplora chain source")
		builder.SetChainSourceEsplora(cfg.GetEnv().LDKEsploraServer, &ldk_node.EsploraSyncConfig{
			// turn off background sync - we manage syncs ourselves
			BackgroundSyncConfig: nil,
		})
		chainSource = "esplora"
	}

	if cfg.GetEnv().LDKGossipSource != "" {
		logger.Logger.WithField("gossipSource", cfg.GetEnv().LDKGossipSource).Warn("LDK RGS instance set")
		builder.SetGossipSourceRgs(cfg.GetEnv().LDKGossipSource)
	}
	builder.SetStorageDirPath(filepath.Join(newpath, "./storage"))

	migrateStorage, _ := cfg.Get("LdkMigrateStorage", "")
	clearMigrateStorageConfigValue := false
	if migrateStorage == "VSS" {
		clearMigrateStorageConfigValue = true
		if vssToken == "" {
			return nil, errors.New("migration enabled but no vss token found")
		}
		builder.MigrateStorage(ldk_node.MigrateStorageVss)
	}

	resetStateRequest := getResetStateRequest(cfg)
	if resetStateRequest != nil {
		builder.ResetState(*resetStateRequest)
	}

	logger.Logger.WithFields(logrus.Fields{
		"migrate_storage":     migrateStorage,
		"vss_enabled":         vssToken != "",
		"node_alias":          alias,
		"listening_addresses": listeningAddresses,
		"chain_source":        chainSource,
	}).Info("Creating LDK node")
	setStartupState("Loading node data...")
	var node *ldk_node.Node
	if vssToken != "" {
		node, err = builder.BuildWithVssStoreAndFixedHeaders(cfg.GetEnv().LDKVssUrl, "albyhub", map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", vssToken),
		})
	} else {
		node, err = builder.Build()
	}

	if err != nil {
		logger.Logger.WithError(err).Error("Failed to create LDK node")
		return nil, err
	}

	logger.Logger.WithFields(logrus.Fields{}).Info("LDK node created")

	if clearMigrateStorageConfigValue {
		err = cfg.SetUpdate("LdkMigrateStorage", "", "")
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to clear LDK migrate storage config value")
			return nil, err
		}
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
		ctx:                 ldkCtx,
	}

	eventPublisher.RegisterSubscriber(&ls)

	// TODO: remove after 2026-01-01 - we now log to app logs rather than ldk log files
	// this line is just left to cleanup old logs after the update
	deleteOldLDKLogs(filepath.Join(newpath, "./logs"))

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
					time.Sleep(time.Duration(1000) * time.Millisecond)
					continue
				}

				ls.handleLdkEvent(event)
				ldkEventConsumer <- event

				node.EventHandled()
			}
		}
	}()

	logger.Logger.WithFields(logrus.Fields{
		"nodeId": nodeId,
	}).Info("Starting LDK node...")

	setStartupState("Starting node...")

	err = node.Start()
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to start LDK node")
		return nil, err
	}

	logger.Logger.WithFields(logrus.Fields{
		"nodeId": nodeId,
		"status": node.Status(),
	}).Info("Started LDK node. Syncing wallet...")

	setStartupState("Syncing node...")
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

	if ls.network == "bitcoin" {
		go func() {
			// try to connect to some peers in the background to retrieve P2P gossip data.
			// TODO: Remove once LDK can correctly do gossip with CLN and Eclair nodes
			// see https://github.com/lightningdevkit/rust-lightning/issues/3075
			peers := []string{
				// "035e4ff418fc8b5554c5d9eea66396c227bd429a3251c8cbc711002ba215bfc226@170.75.163.209:9735",  // WoS
				// "02fcc5bfc48e83f06c04483a2985e1c390cb0f35058baa875ad2053858b8e80dbd@35.239.148.251:9735",  // Blink
				// "027100442c3b79f606f80f322d98d499eefcb060599efc5d4ecb00209c2cb54190@3.230.33.224:9735",    // c=

				// Connect to our LSPs for both:
				// - Gossip data
				// - Ability for auto / free channels for users with eligible Alby subscriptions
				"0364913d18a19c671bb36dd04d6ad5be0fe8f2894314c36a9db3f03c2d414907e1@192.243.215.102:9735",  // LQwD
				"031b301307574bbe9b9ac7b79cbe1700e31e544513eae0b5d7497483083f99e581@45.79.192.236:9735",    // Olympus
				"038a9e56512ec98da2b5789761f7af8f280baf98a09282360cd6ff1381b5e889bf@64.23.162.51:9735",     // Megalith LSP
				"02b4552a7a85274e4da01a7c71ca57407181752e8568b31d51f13c111a2941dce3@159.223.176.115:48049", // LNServer_Wave
				"038ba8f67ba8ff5c48764cdd3251c33598d55b203546d08a8f0ec9dcd9f27e3637@52.24.240.84:9735",     // flashsats
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
					logger.Logger.WithFields(logrus.Fields{
						"peer": peer,
					}).WithError(err).Error("Failed to connect to peer")
				}
			}
		}()
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
				balances := ls.node.ListBalances()
				for _, balance := range balances.LightningBalances {
					switch balanceType := (balance).(type) {
					case ldk_node.LightningBalanceContentiousClaimable:
						logger.Logger.WithField("channel_id", balanceType.ChannelId).Debug("Using short sync time while balances are contentious claimable after channel closure")
						ls.lastWalletSyncRequest = time.Now()
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
								"error":     err.Error(),
								"sync_type": "fee_estimates",
								"node_type": config.LDKBackendType,
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
							"error":     err.Error(),
							"sync_type": "full",
							"node_type": config.LDKBackendType,
						},
					})

					// try again at next MIN_SYNC_INTERVAL
					continue
				}

				ls.redeemedOnchainFundsWithinThisSync = false
				ls.lastFullSync = time.Now()
				// fee estimates happens as part of full sync
				ls.lastFeeEstimatesSync = time.Now()

				logger.Logger.WithFields(logrus.Fields{
					"nodeId":   nodeId,
					"status":   node.Status(),
					"duration": math.Ceil(time.Since(syncStartTime).Seconds()),
				}).Info("LDK node synced successfully")

				// delete old payments while node is not syncing
				ls.deleteOldLDKPayments()
			}
		}
	}()

	return &ls, nil
}

var shutdownMutex sync.Mutex

func (ls *LDKService) Shutdown() error {
	shutdownMutex.Lock()
	defer shutdownMutex.Unlock()
	if ls.shuttingDown {
		logger.Logger.Debug("LDK client is already shutting down")
		return nil
	}
	ls.shuttingDown = true
	ls.eventPublisher.RemoveSubscriber(ls)

	logger.Logger.Info("shutting down LDK client")
	logger.Logger.Info("cancelling LDK context")
	ls.cancel()

	maxAttempts := 40
	for i := 0; ls.syncing; i++ {
		logger.Logger.WithField("attempt", i).Warn("Waiting for background sync to finish before stopping LDK node...")
		time.Sleep(1 * time.Second)
		if i > maxAttempts {
			logger.Logger.Error("Timed out waiting for background sync to finish before stopping LDK node")
			break
		}
	}

	logger.Logger.Info("stopping LDK node")
	shutdownChannel := make(chan error)
	go func() {
		shutdownChannel <- ls.node.Stop()
	}()

	select {
	case err := <-shutdownChannel:
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to stop LDK node")
			// do not return error - we still need to destroy the node
		} else {
			logger.Logger.Info("LDK stop node succeeded")
		}
	case <-time.After(5 * time.Minute):
		logger.Logger.Error("Timeout shutting down LDK node after 5 minutes")
	}

	logger.Logger.Debug("Destroying LDK node object")
	ls.node.Destroy()

	logger.Logger.Info("LDK shutdown complete")

	return nil
}

func getMaxTotalRoutingFeeLimit(amountMsat uint64) uint64 {
	return transactions.CalculateFeeReserveMsat(amountMsat)
}

func (ls *LDKService) MakeOffer(ctx context.Context, description string) (string, error) {
	offer, err := ls.node.Bolt12Payment().ReceiveVariableAmount(description, nil)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to generate BOLT12 offer")
		return "", err
	}

	logger.Logger.WithField("offer", offer).Info("Generated BOLT12 offer")
	return offer.String(), nil
}

func (ls *LDKService) SendPaymentSync(invoice string, amount *uint64) (*lnclient.PayInvoiceResponse, error) {
	paymentRequest, err := decodepay.Decodepay(invoice)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"bolt11": invoice,
		}).WithError(err).Error("Failed to decode bolt11 invoice")

		return nil, err
	}

	paymentAmountMsat := uint64(paymentRequest.MSatoshi)
	if amount != nil {
		paymentAmountMsat = *amount
	}

	maxSpendable := ls.getMaxSpendable()
	if paymentAmountMsat > maxSpendable {
		ls.eventPublisher.Publish(&events.Event{
			Event: "nwc_outgoing_liquidity_required",
			Properties: map[string]interface{}{
				// "amount":         amount / 1000,
				// "max_receivable": maxReceivable,
				// "num_channels":   len(gs.node.ListChannels()),
				"node_type": config.LDKBackendType,
			},
		})
	}

	paymentStart := time.Now()
	ldkEventSubscription := ls.ldkEventBroadcaster.Subscribe()
	defer ls.ldkEventBroadcaster.CancelSubscription(ldkEventSubscription)

	saturationPower := ls.cfg.GetEnv().LDKMaxChannelSaturationPowerOfHalf
	maxPathCount := ls.cfg.GetEnv().LDKMaxPathCount
	maxTotalRoutingFeeMsat := getMaxTotalRoutingFeeLimit(paymentAmountMsat)

	routeParameters := &ldk_node.RouteParametersConfig{
		MaxTotalRoutingFeeMsat:          &maxTotalRoutingFeeMsat,
		MaxChannelSaturationPowerOfHalf: saturationPower,
		MaxPathCount:                    maxPathCount,
		MaxTotalCltvExpiryDelta:         1008, // TODO: remove and use default
	}

	invoiceObj, err := ldk_node.Bolt11InvoiceFromStr(invoice)
	if err != nil {
		logger.Logger.WithError(err).Error("ldk failed to parse bolt 11 invoice from string")
		return nil, err
	}

	var paymentHash string
	if amount == nil {
		paymentHash, err = ls.node.Bolt11Payment().Send(invoiceObj, routeParameters)
	} else {
		paymentHash, err = ls.node.Bolt11Payment().SendUsingAmount(invoiceObj, *amount, routeParameters)
	}
	if err != nil {
		logger.Logger.WithError(err).Error("SendPayment failed")
		return nil, err
	}
	fee := uint64(0)
	preimage := ""

	for {
		select {
		case <-ls.ctx.Done():
			return nil, ls.ctx.Err()

		case ev := <-ldkEventSubscription:
			switch event := (*ev).(type) {
			case ldk_node.EventPaymentSuccessful:
				if event.PaymentHash != paymentHash {
					continue
				}
				logger.Logger.WithFields(logrus.Fields{
					"event": event,
				}).Info("Got payment success event")

				if event.PaymentPreimage == nil {
					logger.Logger.WithField("payment_hash", paymentHash).Error("No payment preimage in payment success event")
					return nil, errors.New("payment preimage not found")
				}

				preimage = *event.PaymentPreimage

				if event.FeePaidMsat != nil {
					fee = *event.FeePaidMsat
				}

				logger.Logger.WithFields(logrus.Fields{
					"duration":     time.Since(paymentStart).Milliseconds(),
					"fee":          fee,
					"payment_hash": event.PaymentHash,
				}).Info("Successful payment")

				return &lnclient.PayInvoiceResponse{
					Preimage: preimage,
					Fee:      fee,
				}, nil
			case ldk_node.EventPaymentFailed:
				if event.PaymentHash != nil && *event.PaymentHash == paymentHash {
					failureReasonMessage := ls.getPaymentFailReason(&event)
					logger.Logger.WithFields(logrus.Fields{
						"payment_hash": paymentHash,
						"reason":       failureReasonMessage,
					}).Error("Received payment failed event")
					return nil, fmt.Errorf("received payment failed event: %s", failureReasonMessage)
				}
			}
		}
	}
}

func (ls *LDKService) SendKeysend(amount uint64, destination string, custom_records []lnclient.TLVRecord, preimage string) (*lnclient.PayKeysendResponse, error) {
	paymentStart := time.Now()
	customTlvs := []ldk_node.CustomTlvRecord{}

	for _, customRecord := range custom_records {
		decodedValue, err := hex.DecodeString(customRecord.Value)
		if err != nil {
			return nil, err
		}
		customTlvs = append(customTlvs, ldk_node.CustomTlvRecord{
			TypeNum: customRecord.Type,
			Value:   decodedValue,
		})
	}

	ldkEventSubscription := ls.ldkEventBroadcaster.Subscribe()
	defer ls.ldkEventBroadcaster.CancelSubscription(ldkEventSubscription)

	saturationPower := ls.cfg.GetEnv().LDKMaxChannelSaturationPowerOfHalf
	maxPathCount := ls.cfg.GetEnv().LDKMaxPathCount
	maxTotalRoutingFeeMsat := getMaxTotalRoutingFeeLimit(amount)

	routeParameters := &ldk_node.RouteParametersConfig{
		MaxTotalRoutingFeeMsat:          &maxTotalRoutingFeeMsat,
		MaxChannelSaturationPowerOfHalf: saturationPower,
		MaxPathCount:                    maxPathCount,
		MaxTotalCltvExpiryDelta:         1008, // TODO: remove and use default
	}

	paymentHash, err := ls.node.SpontaneousPayment().SendWithPreimageAndCustomTlvs(amount, destination, customTlvs, preimage, routeParameters)
	if err != nil {
		logger.Logger.WithError(err).Error("Keysend failed")
		return nil, err
	}
	fee := uint64(0)
	for {
		select {
		case <-ls.ctx.Done():
			return nil, ls.ctx.Err()
		case event := <-ldkEventSubscription:

			eventPaymentSuccessful, isEventPaymentSuccessfulEvent := (*event).(ldk_node.EventPaymentSuccessful)
			eventPaymentFailed, isEventPaymentFailedEvent := (*event).(ldk_node.EventPaymentFailed)

			if isEventPaymentSuccessfulEvent && eventPaymentSuccessful.PaymentHash == paymentHash {
				logger.Logger.Info("Got payment success event")

				if eventPaymentSuccessful.FeePaidMsat != nil {
					fee = *eventPaymentSuccessful.FeePaidMsat
				}
				logger.Logger.WithFields(logrus.Fields{
					"duration": time.Since(paymentStart).Milliseconds(),
					"fee":      fee,
				}).Info("Successful keysend payment")
				return &lnclient.PayKeysendResponse{
					Fee: fee,
				}, nil
			}
			if isEventPaymentFailedEvent && eventPaymentFailed.PaymentHash != nil && *eventPaymentFailed.PaymentHash == paymentHash {

				failureReasonMessage := ls.getPaymentFailReason(&eventPaymentFailed)

				logger.Logger.WithFields(logrus.Fields{
					"payment_hash": paymentHash,
					"reason":       failureReasonMessage,
				}).Error("Received payment failed event")

				return nil, fmt.Errorf("payment failed event: %s", failureReasonMessage)
			}
		}
	}

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

func (ls *LDKService) getMaxSpendable() uint64 {
	var spendable uint64 = 0
	channels := ls.node.ListChannels()
	for _, channel := range channels {
		if channel.IsUsable {
			spendable += min(channel.OutboundCapacityMsat, *channel.CounterpartyOutboundHtlcMaximumMsat)
		}
	}
	return spendable
}

func (ls *LDKService) MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64, throughNodePubkey *string) (transaction *lnclient.Transaction, err error) {

	if time.Duration(expiry)*time.Second > maxInvoiceExpiry {
		return nil, errors.New("expiry is too long")
	}

	maxReceivable := ls.getMaxReceivable()

	if amount > maxReceivable {
		ls.eventPublisher.Publish(&events.Event{
			Event: "nwc_incoming_liquidity_required",
			Properties: map[string]interface{}{
				// "amount":         amount / 1000,
				// "max_receivable": maxReceivable,
				// "num_channels":   len(gs.node.ListChannels()),
				"node_type": config.LDKBackendType,
			},
		})
	}

	if expiry == 0 {
		expiry = lnclient.DEFAULT_INVOICE_EXPIRY
	}

	var descriptionType ldk_node.Bolt11InvoiceDescription
	descriptionType = ldk_node.Bolt11InvoiceDescriptionDirect{
		Description: description,
	}
	if description == "" && descriptionHash != "" {
		descriptionType = ldk_node.Bolt11InvoiceDescriptionHash{
			Hash: descriptionHash,
		}
	}

	invoiceObj, err := ls.node.Bolt11Payment().Receive(uint64(amount),
		descriptionType,
		uint32(expiry))

	if err != nil {
		logger.Logger.WithError(err).Error("MakeInvoice failed")
		return nil, err
	}

	payment := ls.node.Payment(invoiceObj.PaymentHash())
	invoice := *payment.Kind.(ldk_node.PaymentKindBolt11).Bolt11Invoice
	paymentRequest, err := decodepay.Decodepay(invoice)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"bolt11": invoice,
		}).WithError(err).Error("Failed to decode bolt11 invoice")

		return nil, err
	}
	expiresAtUnix := time.UnixMilli(int64(paymentRequest.CreatedAt) * 1000).Add(time.Duration(paymentRequest.Expiry) * time.Second).Unix()

	transaction = &lnclient.Transaction{
		Type:            "incoming",
		Invoice:         invoice,
		PaymentHash:     paymentRequest.PaymentHash,
		Preimage:        *payment.Kind.(ldk_node.PaymentKindBolt11).Preimage,
		Amount:          amount,
		CreatedAt:       int64(paymentRequest.CreatedAt),
		ExpiresAt:       &expiresAtUnix,
		Description:     paymentRequest.Description,
		DescriptionHash: paymentRequest.DescriptionHash,
	}

	return transaction, nil
}

func (ls *LDKService) LookupInvoice(ctx context.Context, paymentHash string) (transaction *lnclient.Transaction, err error) {
	// this method shouldn't be any more because this LNClient supports notifications
	return nil, errors.New("this method should not be called")
}

func (ls *LDKService) ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaid bool, invoiceType string) (transactions []lnclient.Transaction, err error) {
	// this method shouldn't be any more because this LNClient supports notifications
	return nil, errors.New("this method should not be called")
}

func (ls *LDKService) ListOnchainTransactions(ctx context.Context) ([]lnclient.OnchainTransaction, error) {
	transactions := []lnclient.OnchainTransaction{}
	for _, payment := range ls.node.ListPayments() {
		onchainPaymentKind, isOnchainPaymentKind := payment.Kind.(ldk_node.PaymentKindOnchain)
		if !isOnchainPaymentKind {
			continue
		}

		transactionType := "incoming"
		if payment.Direction == ldk_node.PaymentDirectionOutbound {
			transactionType = "outgoing"
		}

		var amountMsat uint64
		if payment.AmountMsat != nil {
			amountMsat = *payment.AmountMsat
		}
		var status string
		var height uint32
		var numConfirmations uint32
		switch onchainPaymentStatus := onchainPaymentKind.Status.(type) {
		case ldk_node.ConfirmationStatusConfirmed:
			status = "confirmed"
			height = onchainPaymentStatus.Height
			nodeStatus := ls.node.Status()
			numConfirmations = nodeStatus.CurrentBestBlock.Height - height
		case ldk_node.ConfirmationStatusUnconfirmed:
			status = "unconfirmed"
		}

		createdAt := payment.CreatedAt
		if createdAt == 0 {
			createdAt = payment.LatestUpdateTimestamp
		}

		transactions = append(transactions, lnclient.OnchainTransaction{
			AmountSat:        amountMsat / 1000,
			CreatedAt:        createdAt,
			State:            status,
			Type:             transactionType,
			NumConfirmations: numConfirmations,
			TxId:             onchainPaymentKind.Txid,
		})

	}
	sort.SliceStable(transactions, func(i, j int) bool {
		return transactions[i].CreatedAt > transactions[j].CreatedAt
	})
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
		fundingTxVout := uint32(0)
		if ldkChannel.FundingTxo != nil {
			fundingTxId = ldkChannel.FundingTxo.Txid
			fundingTxVout = ldkChannel.FundingTxo.Vout
		}

		internalChannel := map[string]interface{}{}
		internalChannel["channel"] = ldkChannel
		internalChannel["config"] = map[string]interface{}{
			"AcceptUnderpayingHtlcs":              ldkChannel.Config.AcceptUnderpayingHtlcs,
			"CltvExpiryDelta":                     ldkChannel.Config.CltvExpiryDelta,
			"ForceCloseAvoidanceMaxFeeSatoshis":   ldkChannel.Config.ForceCloseAvoidanceMaxFeeSatoshis,
			"ForwardingFeeBaseMsat":               ldkChannel.Config.ForwardingFeeBaseMsat,
			"ForwardingFeeProportionalMillionths": ldkChannel.Config.ForwardingFeeProportionalMillionths,
			"MaxDustHtlcExposure":                 ldkChannel.Config.MaxDustHtlcExposure,
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
			Public:                                   ldkChannel.IsAnnounced,
			FundingTxId:                              fundingTxId,
			FundingTxVout:                            fundingTxVout,
			Confirmations:                            ldkChannel.Confirmations,
			ConfirmationsRequired:                    ldkChannel.ConfirmationsRequired,
			ForwardingFeeBaseMsat:                    ldkChannel.Config.ForwardingFeeBaseMsat,
			ForwardingFeeProportionalMillionths:      ldkChannel.Config.ForwardingFeeProportionalMillionths,
			UnspendablePunishmentReserve:             unspendablePunishmentReserve,
			CounterpartyUnspendablePunishmentReserve: ldkChannel.CounterpartyUnspendablePunishmentReserve,
			Error:                                    channelError,
			IsOutbound:                               ldkChannel.IsOutbound,
		})
	}

	return channels, nil
}

func (ls *LDKService) GetNodeConnectionInfo(ctx context.Context) (nodeConnectionInfo *lnclient.NodeConnectionInfo, err error) {
	nodeConnectionInfo = &lnclient.NodeConnectionInfo{
		Pubkey: ls.node.NodeId(),
	}

	if ls.cfg.GetEnv().LDKAnnouncementAddresses != "" {
		addresses := strings.Split(ls.cfg.GetEnv().LDKAnnouncementAddresses, ",")
		for _, address := range addresses {
			address = strings.TrimSpace(address)
			if address == "" {
				continue
			}

			var ip string
			var portStr string

			if strings.HasPrefix(address, "[") {
				// IPv6 format: [ipv6]:port
				closeBracket := strings.Index(address, "]")
				if closeBracket > 0 {
					ip = address[0 : closeBracket+1]
					if closeBracket+2 < len(address) && address[closeBracket+1] == ':' {
						portStr = address[closeBracket+2:]
					}
				}
			} else {
				// IPv4 or hostname format: ip:port
				parts := strings.Split(address, ":")
				if len(parts) >= 2 {
					portStr = parts[len(parts)-1]
					ip = strings.Join(parts[:len(parts)-1], ":")
				}
			}

			if portStr != "" {
				if port, parseErr := strconv.Atoi(portStr); parseErr == nil && ip != "" {
					nodeConnectionInfo.Address = ip
					nodeConnectionInfo.Port = port
					break
				}
			}
		}
	}

	return nodeConnectionInfo, nil
}

func (ls *LDKService) ConnectPeer(ctx context.Context, connectPeerRequest *lnclient.ConnectPeerRequest) error {
	peers := ls.node.ListPeers()

	var foundPeer *ldk_node.PeerDetails
	for _, peer := range peers {
		if peer.NodeId == connectPeerRequest.Pubkey {
			foundPeer = &peer
			break
		}
	}

	if foundPeer != nil && !strings.Contains(foundPeer.Address, connectPeerRequest.Address) {
		logger.Logger.WithFields(logrus.Fields{
			"existing_address": foundPeer.Address,
			"new_address":      connectPeerRequest.Address,
		}).Warn("peer address changed, disconnecting first")
		// disconnect first to ensure new IP address is saved in case of re-connecting
		err := ls.node.Disconnect(connectPeerRequest.Pubkey)
		if err != nil {
			// non-critical: only log an error
			logger.Logger.WithField("request", connectPeerRequest).WithError(err).Error("Disconnect failed while connecting peer")
		}
	}

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
	var userChannelId string
	var err error
	if openChannelRequest.Public {
		userChannelId, err = ls.node.OpenAnnouncedChannel(foundPeer.NodeId, foundPeer.Address, uint64(openChannelRequest.AmountSats), nil, nil)
	} else {
		userChannelId, err = ls.node.OpenChannel(foundPeer.NodeId, foundPeer.Address, uint64(openChannelRequest.AmountSats), nil, nil)
	}
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
	existingConfig.ForwardingFeeBaseMsat = updateChannelRequest.ForwardingFeeBaseMsat
	existingConfig.ForwardingFeeProportionalMillionths = updateChannelRequest.ForwardingFeeProportionalMillionths

	if updateChannelRequest.MaxDustHtlcExposureFromFeeRateMultiplier > 0 {
		existingConfig.MaxDustHtlcExposure = ldk_node.MaxDustHtlcExposureFeeRateMultiplier{
			Multiplier: updateChannelRequest.MaxDustHtlcExposureFromFeeRateMultiplier,
		}
	}

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
		err = ls.node.ForceCloseChannel(closeChannelRequest.ChannelId, closeChannelRequest.NodeId, nil)
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
	nodeStatus := ls.node.Status()
	channels := ls.node.ListChannels()
	balances := ls.node.ListBalances()
	logger.Logger.WithFields(logrus.Fields{
		"balances": balances,
	}).Debug("Listed Balances")

	type internalLightningBalance struct {
		BalanceType string
		Balance     ldk_node.LightningBalance
	}

	internalLightningBalances := []internalLightningBalance{}

	pendingBalancesDetails := make([]lnclient.PendingBalanceDetails, 0)

	pendingBalancesFromChannelClosures := uint64(0)
	// increase pending balance from any lightning balances for channels that are pending closure
	// (they do not exist in our list of open channels)
	for _, balance := range balances.LightningBalances {
		increasePendingBalance := func(nodeId, channelId string, amount uint64, fundingTxId ldk_node.Txid, fundingTxIndex uint16) {
			if !slices.ContainsFunc(channels, func(channel ldk_node.ChannelDetails) bool {
				return channel.ChannelId == channelId
			}) {
				pendingBalancesFromChannelClosures += amount
				pendingBalancesDetails = append(pendingBalancesDetails, lnclient.PendingBalanceDetails{
					NodeId:        nodeId,
					ChannelId:     channelId,
					Amount:        amount,
					FundingTxId:   fundingTxId,
					FundingTxVout: uint32(fundingTxIndex),
				})
			}
		}

		// include the balance type as it's useful to know the state of the channel
		internalLightningBalances = append(internalLightningBalances, internalLightningBalance{
			BalanceType: fmt.Sprintf("%T", balance),
			Balance:     balance,
		})
		switch balanceType := (balance).(type) {
		case ldk_node.LightningBalanceClaimableOnChannelClose:
			increasePendingBalance(balanceType.CounterpartyNodeId, balanceType.ChannelId, balanceType.AmountSatoshis, balanceType.FundingTxId, balanceType.FundingTxIndex)
		case ldk_node.LightningBalanceClaimableAwaitingConfirmations:
			increasePendingBalance(balanceType.CounterpartyNodeId, balanceType.ChannelId, balanceType.AmountSatoshis, balanceType.FundingTxId, balanceType.FundingTxIndex)
		case ldk_node.LightningBalanceContentiousClaimable:
			increasePendingBalance(balanceType.CounterpartyNodeId, balanceType.ChannelId, balanceType.AmountSatoshis, balanceType.FundingTxId, balanceType.FundingTxIndex)
		case ldk_node.LightningBalanceMaybeTimeoutClaimableHtlc:
			increasePendingBalance(balanceType.CounterpartyNodeId, balanceType.ChannelId, balanceType.AmountSatoshis, balanceType.FundingTxId, balanceType.FundingTxIndex)
		case ldk_node.LightningBalanceMaybePreimageClaimableHtlc:
			increasePendingBalance(balanceType.CounterpartyNodeId, balanceType.ChannelId, balanceType.AmountSatoshis, balanceType.FundingTxId, balanceType.FundingTxIndex)
		case ldk_node.LightningBalanceCounterpartyRevokedOutputClaimable:
			increasePendingBalance(balanceType.CounterpartyNodeId, balanceType.ChannelId, balanceType.AmountSatoshis, balanceType.FundingTxId, balanceType.FundingTxIndex)
		}
	}

	pendingSweepBalanceDetails := make([]lnclient.PendingBalanceDetails, 0)
	increasePendingBalanceFromClosure := func(nodeId, channelId *string, amount uint64, fundingTxId *ldk_node.Txid, fundingTxIndex *uint16) {
		pendingBalancesFromChannelClosures += amount

		if nodeId != nil && channelId != nil && fundingTxId != nil && fundingTxIndex != nil {
			pendingSweepBalanceDetails = append(pendingSweepBalanceDetails, lnclient.PendingBalanceDetails{
				NodeId:        *nodeId,
				ChannelId:     *channelId,
				Amount:        amount,
				FundingTxId:   *fundingTxId,
				FundingTxVout: uint32(*fundingTxIndex),
			})
		}
	}

	// increase pending balance from any lightning balances for channels that were closed
	for _, balance := range balances.PendingBalancesFromChannelClosures {
		switch pendingType := (balance).(type) {
		case ldk_node.PendingSweepBalancePendingBroadcast:
			increasePendingBalanceFromClosure(pendingType.CounterpartyNodeId, pendingType.ChannelId, pendingType.AmountSatoshis, pendingType.FundingTxId, pendingType.FundingTxIndex)
		case ldk_node.PendingSweepBalanceBroadcastAwaitingConfirmation:
			increasePendingBalanceFromClosure(pendingType.CounterpartyNodeId, pendingType.ChannelId, pendingType.AmountSatoshis, pendingType.FundingTxId, pendingType.FundingTxIndex)
		case ldk_node.PendingSweepBalanceAwaitingThresholdConfirmations:
			if nodeStatus.CurrentBestBlock.Height < pendingType.ConfirmationHeight+6 {
				// LDK now keeps the balance in this state for four weeks even after the funds are confirmed to be swept
				// to confirm the channel monitors are archived before the sweeper entries are dropped
				// so now we just check for 6 confirmations
				increasePendingBalanceFromClosure(pendingType.CounterpartyNodeId, pendingType.ChannelId, pendingType.AmountSatoshis, pendingType.FundingTxId, pendingType.FundingTxIndex)
			}
		}
	}

	return &lnclient.OnchainBalanceResponse{
		Spendable:                          int64(balances.SpendableOnchainBalanceSats),
		Total:                              int64(balances.TotalOnchainBalanceSats - balances.TotalAnchorChannelsReserveSats),
		Reserved:                           int64(balances.TotalAnchorChannelsReserveSats),
		PendingBalancesFromChannelClosures: pendingBalancesFromChannelClosures,
		PendingBalancesDetails:             pendingBalancesDetails,
		PendingSweepBalancesDetails:        pendingSweepBalanceDetails,
		InternalBalances: map[string]interface{}{
			"internal_lightning_balances": internalLightningBalances,
			"all_balances":                balances,
		},
	}, nil
}

func (ls *LDKService) RedeemOnchainFunds(ctx context.Context, toAddress string, amount uint64, feeRate *uint64, sendAll bool) (string, error) {
	if ls.redeemedOnchainFundsWithinThisSync {
		return "", errors.New("please wait a minute for the wallet to sync before doing another on-chain payment")
	}

	var feePtr **ldk_node.FeeRate
	if feeRate != nil {
		fee := ldk_node.FeeRateFromSatPerVbUnchecked(*feeRate)
		feePtr = &fee
	}

	var txId string
	var err error

	if !sendAll {
		// NOTE: this may fail if user does not reserve enough for the onchain transaction
		// and can also drain the anchor reserves if the user provides a too high amount.
		txId, err = ls.node.OnchainPayment().SendToAddress(toAddress, amount, feePtr)
	} else {
		txId, err = ls.node.OnchainPayment().SendAllToAddress(toAddress, false, feePtr)
	}

	if err != nil {
		logger.Logger.WithField("send_all", sendAll).WithError(err).Error("LDK onchain payment to redeem funds failed")
		return "", err
	}

	// make sure we do a sync after sending on-chain funds
	ls.redeemedOnchainFundsWithinThisSync = true
	ls.lastWalletSyncRequest = time.Now()

	// FIXME: remove once LDK-node returns an error if it can't broadcast the transaction

	tryCheckTransactionWasBroadcasted := func() error {
		url := ls.cfg.GetEnv().MempoolApi + "/tx/" + txId

		client := http.Client{
			Timeout: time.Second * 10,
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			logger.Logger.WithError(err).WithFields(logrus.Fields{
				"url": url,
			}).Error("Failed to create http request")
			return err
		}

		res, err := client.Do(req)
		if err != nil {
			logger.Logger.WithError(err).WithFields(logrus.Fields{
				"url": url,
			}).Error("Failed to send request")
			return err
		}

		if res.StatusCode >= 300 {
			// transaction not found
			return errors.New("unexpected status code")
		}

		return nil
	}

	for attempt := 1; attempt < 30; attempt++ {
		err := tryCheckTransactionWasBroadcasted()
		if err != nil {
			logger.Logger.WithError(err).WithField("attempt", attempt).Error("Failed to fetch broadcasted transaction")
			time.Sleep(1 * time.Second)
			continue
		}

		return txId, nil
	}
	return "", errors.New("ran out of attempts to fetch broadcasted transaction")
}

func (ls *LDKService) ResetRouter(key string) error {
	err := ls.cfg.SetUpdate(resetRouterKey, key, "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to set reset router key")
		return err
	}

	return nil
}

func (ls *LDKService) SignMessage(ctx context.Context, message string) (string, error) {
	signedMessage := ls.node.SignMessage([]byte(message))

	return signedMessage, nil
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
			}).WithError(err).Error("Failed to decode bolt11 invoice")

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
			if payment.LatestUpdateTimestamp > 0 {
				lastUpdate := int64(payment.LatestUpdateTimestamp)
				settledAt = &lastUpdate
			}
		}
		paymentHash = bolt11PaymentKind.Hash
	}

	bolt12PaymentKind, isBolt12PaymentKind := payment.Kind.(ldk_node.PaymentKindBolt12Offer)

	if isBolt12PaymentKind {
		createdAt = int64(payment.CreatedAt)

		if bolt12PaymentKind.Hash == nil {
			return nil, errors.New("BOLT-12 payment has no payment hash")
		}
		paymentHash = *bolt12PaymentKind.Hash

		offer := map[string]interface{}{}
		offer["id"] = bolt12PaymentKind.OfferId

		if bolt12PaymentKind.PayerNote != nil {
			offer["payer_note"] = *bolt12PaymentKind.PayerNote
		}

		metadata["offer"] = offer

		if payment.Status == ldk_node.PaymentStatusSucceeded {
			if bolt12PaymentKind.Preimage != nil {
				preimage = *bolt12PaymentKind.Preimage
			}
			lastUpdate := int64(payment.LatestUpdateTimestamp)
			settledAt = &lastUpdate
		}
	}

	spontaneousPaymentKind, isSpontaneousPaymentKind := payment.Kind.(ldk_node.PaymentKindSpontaneous)
	if isSpontaneousPaymentKind {
		// keysend payment
		lastUpdate := int64(payment.LatestUpdateTimestamp)
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
	if payment.FeePaidMsat != nil {
		fee = *payment.FeePaidMsat
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
	return errors.ErrUnsupported
}

func (ls *LDKService) SendSpontaneousPaymentProbes(ctx context.Context, amountMsat uint64, nodeId string) error {
	return errors.ErrUnsupported
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
			if graphNode.Channels != nil {
				for _, channelId := range graphNode.Channels {
					graphChannel := graph.Channel(channelId)
					if graphChannel != nil {
						channels = append(channels, graphChannel)
					}
				}
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
	return []byte("Node logs are now included in application logs"), nil
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

		isTrusted := eventType.CounterpartyNodeId != nil && slices.Contains(ls.node.Config().AnchorChannelsConfig.TrustedPeersNoReserve, *eventType.CounterpartyNodeId)

		channel := channels[channelIndex]
		ls.eventPublisher.Publish(&events.Event{
			Event: "nwc_channel_ready",
			Properties: map[string]interface{}{
				"counterparty_node_id": eventType.CounterpartyNodeId,
				"node_type":            config.LDKBackendType,
				"public":               channel.IsAnnounced,
				"capacity":             channel.ChannelValueSats,
				"is_outbound":          channel.IsOutbound,
				"trusted":              isTrusted,
			},
		})

		ls.backupChannels()

		if eventType.CounterpartyNodeId == nil {
			logger.Logger.WithField("event", eventType).Error("channel ready event has no counterparty node ID")
			return
		}

		maxDustHtlcExposureFromFeeRateMultiplier := uint64(0)
		if isTrusted {
			// avoid closures like "ProcessingError: Peer sent update_fee with a feerate (62500)
			// which may over-expose us to dust-in-flight on our counterparty's transactions (totaling 69348000 msat)"
			maxDustHtlcExposureFromFeeRateMultiplier = 100_000 // default * 10
		}

		// set a super-high forwarding fee of 100K sats by default to disable unwanted routing by default
		forwardingFeeBaseMsat := uint32(100_000_000)

		err := ls.UpdateChannel(context.Background(), &lnclient.UpdateChannelRequest{
			ChannelId:                                eventType.UserChannelId,
			NodeId:                                   *eventType.CounterpartyNodeId,
			MaxDustHtlcExposureFromFeeRateMultiplier: maxDustHtlcExposureFromFeeRateMultiplier,
			ForwardingFeeBaseMsat:                    forwardingFeeBaseMsat,
		})

		if err != nil {
			logger.Logger.WithField("event", eventType).Error("channel ready event has no counterparty node ID")
			return
		}

	case ldk_node.EventChannelClosed:
		// make sure we do a sync after receiving a channel closed event
		ls.lastWalletSyncRequest = time.Now()

		closureReason := ls.getChannelCloseReason(&eventType)
		logger.Logger.WithFields(logrus.Fields{
			"event":  event,
			"reason": closureReason,
		}).Info("Channel closed")
		onchainBalance, err := ls.GetOnchainBalance(context.Background())
		if err != nil {
			logger.Logger.WithError(err).Error("failed to retrieve on-chain balance when closing channel")
		}
		var pendingBalance uint64
		var fundingTxId string
		var fundingTxVout uint32
		var fundingTxUrl string

		if onchainBalance != nil {
			logger.Logger.WithField("onchain_balance", onchainBalance).Info("got on-chain balance when closing channel")

			for _, details := range onchainBalance.PendingBalancesDetails {
				if details.ChannelId == eventType.ChannelId {
					fundingTxId = details.FundingTxId
					fundingTxVout = details.FundingTxVout
					fundingTxUrl = fmt.Sprintf("https://mempool.space/tx/%s#flow=&vout=%d", fundingTxId, fundingTxVout)
					pendingBalance += details.Amount
				}
			}
			for _, details := range onchainBalance.PendingSweepBalancesDetails {
				if details.ChannelId == eventType.ChannelId {
					fundingTxId = details.FundingTxId
					fundingTxVout = details.FundingTxVout
					fundingTxUrl = fmt.Sprintf("https://mempool.space/tx/%s#flow=&vout=%d", fundingTxId, fundingTxVout)
					pendingBalance += details.Amount
				}
			}
		}

		var counterpartyNodeId string
		var counterpartyNodeUrl string
		if eventType.CounterpartyNodeId != nil {
			counterpartyNodeId = *eventType.CounterpartyNodeId
			counterpartyNodeUrl = "https://amboss.space/node/" + counterpartyNodeId
		}

		ls.eventPublisher.Publish(&events.Event{
			Event: "nwc_channel_closed",
			Properties: map[string]interface{}{
				"counterparty_node_id":  counterpartyNodeId,
				"counterparty_node_url": counterpartyNodeUrl,
				"reason":                closureReason,
				"node_type":             config.LDKBackendType,
				"pending_balance":       pendingBalance,
				"funding_tx_id":         fundingTxId,
				"funding_tx_vout":       fundingTxVout,
				"funding_tx_url":        fundingTxUrl,
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
	case ldk_node.EventPaymentForwarded:
		logger.Logger.WithFields(logrus.Fields{
			"total_fee_earned_msat":          eventType.TotalFeeEarnedMsat,
			"outbound_amount_forwarded_msat": eventType.OutboundAmountForwardedMsat,
		}).Info("LDK Payment forwarded")
		if eventType.TotalFeeEarnedMsat == nil || eventType.OutboundAmountForwardedMsat == nil {
			logger.Logger.WithFields(logrus.Fields{
				"earned_msat":                    eventType.TotalFeeEarnedMsat,
				"outbound_amount_forwarded_msat": eventType.OutboundAmountForwardedMsat,
			}).Error("forwarded payment has missing required fields")
			return
		}
		ls.eventPublisher.Publish(&events.Event{
			Event: "nwc_payment_forwarded",
			Properties: &lnclient.PaymentForwardedEventProperties{
				TotalFeeEarnedMsat:          *eventType.TotalFeeEarnedMsat,
				OutboundAmountForwardedMsat: *eventType.OutboundAmountForwardedMsat,
			},
		})

	case ldk_node.EventPaymentClaimable:
		if eventType.ClaimDeadline == nil {
			logger.Logger.WithField("payment_id", eventType.PaymentId).Error("claimable payment has no claim deadline")
			return
		}

		logger.Logger.WithFields(logrus.Fields{
			"claimable_amount_msats": eventType.ClaimableAmountMsat,
			"payment_hash":           eventType.PaymentHash,
			"claim_deadline":         *eventType.ClaimDeadline,
		}).Info("LDK Payment Claimable")

		payment := ls.node.Payment(eventType.PaymentId)
		if payment == nil {
			logger.Logger.WithField("payment_id", eventType.PaymentId).Error("could not find LDK payment")
			return
		}

		transaction, err := ls.ldkPaymentToTransaction(payment)
		if err != nil {
			logger.Logger.WithField("payment_id", eventType.PaymentId).Error("failed to convert LDK payment to transaction")
			return
		}
		transaction.SettleDeadline = eventType.ClaimDeadline
		ls.eventPublisher.Publish(&events.Event{
			Event:      "nwc_lnclient_hold_invoice_accepted",
			Properties: transaction,
		})
	}
}

func (ls *LDKService) backupChannels() {
	ldkChannels := ls.node.ListChannels()
	ldkPeers := ls.node.ListPeers()
	channels := make([]events.ChannelBackup, 0, len(ldkChannels))
	for _, ldkChannel := range ldkChannels {
		var fundingTxId string
		var fundingTxVout uint32
		if ldkChannel.FundingTxo != nil {
			fundingTxId = ldkChannel.FundingTxo.Txid
			fundingTxVout = ldkChannel.FundingTxo.Vout
		}

		var peer *ldk_node.PeerDetails
		for _, matchingPeer := range ldkPeers {
			if matchingPeer.NodeId == ldkChannel.CounterpartyNodeId {
				peer = &matchingPeer
			}
		}
		if peer == nil {
			logger.Logger.WithField("peer_id", ldkChannel.CounterpartyNodeId).Error("failed to find peer for channel")
			continue
		}

		channels = append(channels, events.ChannelBackup{
			ChannelID:         ldkChannel.ChannelId,
			PeerID:            ldkChannel.CounterpartyNodeId,
			PeerSocketAddress: peer.Address,
			ChannelSize:       ldkChannel.ChannelValueSats,
			FundingTxID:       fundingTxId,
			FundingTxVout:     fundingTxVout,
		})
	}

	monitors, err := ls.node.GetEncodedChannelMonitors()
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to list channel monitors")
		return
	}
	encodedMonitors := []events.EncodedChannelMonitorBackup{}

	for _, monitor := range monitors {
		encodedMonitors = append(encodedMonitors, events.EncodedChannelMonitorBackup{
			Key:   monitor.Key,
			Value: hex.EncodeToString(monitor.Value),
		})
	}

	event := &events.StaticChannelsBackupEvent{
		Channels: channels,
		Monitors: encodedMonitors,
		NodeID:   ls.node.NodeId(),
	}

	ls.saveStaticChannelBackupToDisk(event)

	ls.eventPublisher.Publish(&events.Event{
		Event:      "nwc_backup_channels",
		Properties: event,
	})
}

func (ls *LDKService) saveStaticChannelBackupToDisk(event *events.StaticChannelsBackupEvent) {
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

func (ls *LDKService) GetBalances(ctx context.Context, includeInactiveChannels bool) (*lnclient.BalancesResponse, error) {
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
		if channel.IsUsable || includeInactiveChannels {
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

func (ls *LDKService) deleteOldLDKPayments() {
	payments := ls.node.ListPayments()

	now := time.Now()
	for _, payment := range payments {
		paymentCreatedAt := time.Unix(int64(payment.CreatedAt), 0)

		deletablePaymentKind := false
		switch (payment.Kind).(type) {
		case ldk_node.PaymentKindBolt11:
			deletablePaymentKind = true
		case ldk_node.PaymentKindSpontaneous:
			deletablePaymentKind = true
		}
		if !deletablePaymentKind {
			logger.Logger.WithFields(logrus.Fields{
				"created_at": paymentCreatedAt,
				"payment_id": payment.Id,
			}).Debug("Skipping undeletable payment kind")
			continue
		}

		if paymentCreatedAt.Add(maxInvoiceExpiry).Before(now) {
			logger.Logger.WithFields(logrus.Fields{
				"created_at": paymentCreatedAt,
				"payment_id": payment.Id,
			}).Debug("Deleting old payment")
			err := ls.node.RemovePayment(payment.Id)
			if err != nil {
				logger.Logger.WithError(err).WithField("id", payment.Id).Error("failed to delete old payment")
			}
		}
	}
}

func deleteOldLDKLogs(ldkLogDir string) {
	logger.Logger.WithField("ldkLogDir", ldkLogDir).Debug("Deleting old LDK logs")
	files, err := os.ReadDir(ldkLogDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// no log file directory - expected when VSS is enabled
			return
		}
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
	status := ls.node.Status()
	return &lnclient.NodeStatus{
		IsReady:            status.IsRunning,
		InternalNodeStatus: status,
	}, nil
}

func (ls *LDKService) DisconnectPeer(ctx context.Context, peerId string) error {
	return ls.node.Disconnect(peerId)
}

func (ls *LDKService) UpdateLastWalletSyncRequest() {
	ls.lastWalletSyncRequest = time.Now()
}

func (ls *LDKService) GetSupportedNIP47Methods() []string {
	return []string{
		models.PAY_INVOICE_METHOD,
		models.PAY_KEYSEND_METHOD,
		models.GET_BALANCE_METHOD,
		models.GET_BUDGET_METHOD,
		models.GET_INFO_METHOD,
		models.MAKE_INVOICE_METHOD,
		models.LOOKUP_INVOICE_METHOD,
		models.LIST_TRANSACTIONS_METHOD,
		models.MULTI_PAY_INVOICE_METHOD,
		models.MULTI_PAY_KEYSEND_METHOD,
		models.SIGN_MESSAGE_METHOD,
		models.MAKE_HOLD_INVOICE_METHOD,
		models.SETTLE_HOLD_INVOICE_METHOD,
		models.CANCEL_HOLD_INVOICE_METHOD,
	}
}

func (ls *LDKService) GetSupportedNIP47NotificationTypes() []string {
	return []string{
		notifications.PAYMENT_RECEIVED_NOTIFICATION,
		notifications.PAYMENT_SENT_NOTIFICATION,
		notifications.HOLD_INVOICE_ACCEPTED_NOTIFICATION,
	}
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
	case ldk_node.PaymentFailureReasonUnknownRequiredFeatures:
		failureReasonMessage = "UnknownRequiredFeatures"
	case ldk_node.PaymentFailureReasonInvoiceRequestExpired:
		failureReasonMessage = "InvoiceRequestExpired"
	case ldk_node.PaymentFailureReasonInvoiceRequestRejected:
		failureReasonMessage = "InvoiceRequestRejected"
	case ldk_node.PaymentFailureReasonBlindedPathCreationFailed:
		failureReasonMessage = "BlindedPathCreationFailed"
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

func (ls *LDKService) PayOfferSync(ctx context.Context, offer string, amount uint64, payerNote string) (*lnclient.PayOfferResponse, error) {
	// TODO: this is only for testing MakeOffer and needs improvements
	// (+ BOLT-12 payments need to go through transactions service)
	// TODO: send liquidity event if amount too large
	offerObj, err := ldk_node.OfferFromStr(offer)
	if err != nil {
		return nil, err
	}

	paymentStart := time.Now()
	ldkEventSubscription := ls.ldkEventBroadcaster.Subscribe()
	defer ls.ldkEventBroadcaster.CancelSubscription(ldkEventSubscription)

	// TODO: use normal send if no amount is provided
	// TODO: configure sending params to ensure fee reserve is used, etc.
	paymentId, err := ls.node.Bolt12Payment().SendUsingAmount(offerObj, amount, nil, &payerNote, nil)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to initiate BOLT-12 variable amount payment")
		return nil, errors.New("failed to initiate BOLT-12 variable amount payment")
	}

	logger.Logger.WithFields(logrus.Fields{
		"payment_id": paymentId,
	}).Info("Initiated BOLT-12 variable amount payment")

	fee := uint64(0)
	preimage := ""

	payment := ls.node.Payment(paymentId)
	if payment == nil {
		return nil, errors.New("payment not found by payment ID")
	}

	paymentHash := ""

	for start := time.Now(); time.Since(start) < time.Second*60; {
		event := <-ldkEventSubscription

		eventPaymentSuccessful, isEventPaymentSuccessfulEvent := (*event).(ldk_node.EventPaymentSuccessful)
		eventPaymentFailed, isEventPaymentFailedEvent := (*event).(ldk_node.EventPaymentFailed)

		if isEventPaymentSuccessfulEvent && eventPaymentSuccessful.PaymentId != nil && *eventPaymentSuccessful.PaymentId == paymentId {
			logger.Logger.Info("Got payment success event")
			payment := ls.node.Payment(paymentId)
			if payment == nil {
				logger.Logger.Errorf("Couldn't find payment by payment ID: %v", paymentId)
				return nil, errors.New("payment not found")
			}

			bolt12PaymentKind, ok := payment.Kind.(ldk_node.PaymentKindBolt12Offer)

			if !ok {
				logger.Logger.WithFields(logrus.Fields{
					"payment": payment,
				}).Error("Payment is not a BOLT-12 offer kind")
				return nil, errors.New("payment is not a BOLT-12 offer")
			}

			if bolt12PaymentKind.Preimage == nil {
				logger.Logger.Errorf("No payment preimage for payment ID: %v", paymentId)
				return nil, errors.New("payment preimage not found")
			}
			preimage = *bolt12PaymentKind.Preimage

			if bolt12PaymentKind.Hash == nil {
				logger.Logger.Errorf("No payment hash for payment ID: %v", paymentId)
				return nil, errors.New("payment hash not found")
			}
			paymentHash = *bolt12PaymentKind.Hash

			if eventPaymentSuccessful.FeePaidMsat != nil {
				fee = *eventPaymentSuccessful.FeePaidMsat
			}
			break
		}
		if isEventPaymentFailedEvent && eventPaymentFailed.PaymentId != nil && *eventPaymentFailed.PaymentId == paymentId {
			reason := ls.getPaymentFailReason(&eventPaymentFailed)

			logger.Logger.WithFields(logrus.Fields{
				"payment_id": paymentId,
				"reason":     reason,
			}).Error("Received payment failed event")

			return nil, fmt.Errorf("received payment failed event: %s", reason)
		}
	}

	logger.Logger.WithFields(logrus.Fields{
		"duration": time.Since(paymentStart).Milliseconds(),
		"fee":      fee,
	}).Info("Successful BOLT-12 payment")

	return &lnclient.PayOfferResponse{
		PaymentHash: paymentHash,
		Preimage:    preimage,
		Fee:         fee,
	}, nil
}

const nodeCommandPayBOLT12Offer = "pay_bolt12_offer"
const nodeCommandExportPathfindingScores = "export_pathfinding_scores"
const nodeCommandListChannelMonitorSizes = "list_channel_monitor_sizes"

func (ls *LDKService) GetCustomNodeCommandDefinitions() []lnclient.CustomNodeCommandDef {
	return []lnclient.CustomNodeCommandDef{
		{
			Name:        nodeCommandPayBOLT12Offer,
			Description: "Send payments to a BOLT-12 offer. NOTE: this is for testing only. Payment will not show in transaction list.",
			Args: []lnclient.CustomNodeCommandArgDef{
				{
					Name:        "offer",
					Description: "BOLT-12 offer of receiver",
				},
				{
					Name:        "amount",
					Description: "amount to send in millisats",
				},
				{
					Name:        "payer_note",
					Description: "note to the recepient",
				},
			},
		},
		{
			Name:        nodeCommandExportPathfindingScores,
			Description: "Exports pathfinding scores from the LDK node.",
			Args:        []lnclient.CustomNodeCommandArgDef{}, // Assuming no arguments for now
		},
		{
			Name:        nodeCommandListChannelMonitorSizes,
			Description: "List Channel Monitor sizes from the LDK node.",
			Args:        []lnclient.CustomNodeCommandArgDef{},
		},
	}
}

func (ls *LDKService) ExecuteCustomNodeCommand(ctx context.Context, command *lnclient.CustomNodeCommandRequest) (*lnclient.CustomNodeCommandResponse, error) {
	switch command.Name {
	case nodeCommandPayBOLT12Offer:
		var offer string
		var amount uint64
		var payerNote string
		var err error
		for i := range command.Args {
			switch command.Args[i].Name {
			case "offer":
				offer = command.Args[i].Value
			case "amount":
				amount, err = strconv.ParseUint(string(command.Args[i].Value), 10, 64)
			case "payer_note":
				payerNote = command.Args[i].Value
			}
		}
		if err != nil {
			return nil, err
		}

		payOfferResponse, err := ls.PayOfferSync(ctx, offer, amount, payerNote)

		if err != nil {
			return nil, err
		}

		return &lnclient.CustomNodeCommandResponse{
			Response: map[string]interface{}{
				"paymentHash": payOfferResponse.PaymentHash,
				"preimage":    payOfferResponse.Preimage,
				"fee":         payOfferResponse.Fee,
			},
		}, nil
	case nodeCommandExportPathfindingScores:
		scores, err := ls.node.ExportPathfindingScores()
		if err != nil {
			logger.Logger.WithError(err).Error("ExportPathfindingScores command failed")
			return nil, fmt.Errorf("failed to export pathfinding scores: %w", err)
		}
		return &lnclient.CustomNodeCommandResponse{
			Response: map[string]interface{}{
				"scores": hex.EncodeToString(scores),
			},
		}, nil
	case nodeCommandListChannelMonitorSizes:
		channelMonitorSizes := ls.node.ListChannelMonitorSizes()
		channels := ls.node.ListChannels()
		type channelMonitorSizeResponse struct {
			SizeBytes    uint64 `json:"sizeBytes"`
			RemotePubkey string `json:"remotePubkey"`
			HasWarning   bool   `json:"hasWarning"`
		}
		channelMonitorSizesResponse := []channelMonitorSizeResponse{}
		for _, channelMonitorSizeInfo := range channelMonitorSizes {
			for _, channel := range channels {
				if channel.ChannelId == channelMonitorSizeInfo.ChannelId {
					channelMonitorSizesResponse = append(channelMonitorSizesResponse, channelMonitorSizeResponse{
						SizeBytes:    channelMonitorSizeInfo.SizeBytes,
						RemotePubkey: channel.CounterpartyNodeId,
						HasWarning:   channelMonitorSizeInfo.SizeBytes >= ls.cfg.GetEnv().LDKChannelMonitorWarningSizeBytes,
					})
				}
			}
		}

		return &lnclient.CustomNodeCommandResponse{
			Response: channelMonitorSizesResponse,
		}, nil
	}

	return nil, lnclient.ErrUnknownCustomNodeCommand
}

func (ls *LDKService) MakeHoldInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64, paymentHash string) (*lnclient.Transaction, error) {
	if time.Duration(expiry)*time.Second > maxInvoiceExpiry {
		return nil, errors.New("expiry is too long")
	}

	maxReceivable := ls.getMaxReceivable()

	if amount > maxReceivable {
		ls.eventPublisher.Publish(&events.Event{
			Event: "nwc_incoming_liquidity_required",
			Properties: map[string]interface{}{
				"node_type": config.LDKBackendType,
			},
		})
	}

	if expiry == 0 {
		expiry = lnclient.DEFAULT_INVOICE_EXPIRY
	}

	var descriptionType ldk_node.Bolt11InvoiceDescription
	descriptionType = ldk_node.Bolt11InvoiceDescriptionDirect{
		Description: description,
	}
	if description == "" && descriptionHash != "" {
		descriptionType = ldk_node.Bolt11InvoiceDescriptionHash{
			Hash: descriptionHash,
		}
	}

	decodedPaymentHash, err := hex.DecodeString(paymentHash)
	if err != nil {
		logger.Logger.WithError(err).WithField("paymentHash", paymentHash).Error("Failed to decode payment hash for MakeHoldInvoice")
		return nil, fmt.Errorf("failed to decode payment hash: %w", err)
	}
	if len(decodedPaymentHash) != 32 {
		return nil, errors.New("payment hash must be 32 bytes")
	}
	var paymentHash32 [32]byte
	copy(paymentHash32[:], decodedPaymentHash)

	ldkPaymentHash := ldk_node.PaymentHash(hex.EncodeToString(paymentHash32[:]))

	invoiceObj, err := ls.node.Bolt11Payment().ReceiveForHash(uint64(amount),
		descriptionType,
		uint32(expiry),
		ldkPaymentHash)

	if err != nil {
		logger.Logger.WithError(err).Error("MakeHoldInvoice failed")
		return nil, err
	}

	payment := ls.node.Payment(invoiceObj.PaymentHash())
	invoice := *payment.Kind.(ldk_node.PaymentKindBolt11).Bolt11Invoice
	paymentRequest, err := decodepay.Decodepay(invoice)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"bolt11": invoice,
		}).WithError(err).Error("Failed to decode bolt11 invoice")
		return nil, err
	}
	expiresAtUnix := time.UnixMilli(int64(paymentRequest.CreatedAt) * 1000).Add(time.Duration(paymentRequest.Expiry) * time.Second).Unix()

	transaction := &lnclient.Transaction{
		Type:            "incoming",
		Invoice:         *payment.Kind.(ldk_node.PaymentKindBolt11).Bolt11Invoice,
		PaymentHash:     paymentRequest.PaymentHash,
		Amount:          amount,
		CreatedAt:       int64(payment.CreatedAt),
		ExpiresAt:       &expiresAtUnix,
		Description:     paymentRequest.Description,
		DescriptionHash: paymentRequest.DescriptionHash,
	}

	return transaction, nil
}

func (ls *LDKService) CancelHoldInvoice(ctx context.Context, paymentHash string) error {
	_, err := hex.DecodeString(paymentHash)
	if err != nil {
		logger.Logger.WithError(err).WithField("paymentHash", paymentHash).Error("Failed to decode payment hash for CancelHoldInvoice")
		return err
	}

	err = ls.node.Bolt11Payment().FailForHash(paymentHash)
	if err != nil {
		logger.Logger.WithError(err).WithField("paymentHash", paymentHash).Error("CancelHoldInvoice failed")
	}
	return err
}

func (ls *LDKService) SettleHoldInvoice(ctx context.Context, preimage string) error {
	decodedPreimage, err := hex.DecodeString(preimage)
	if err != nil {
		logger.Logger.WithError(err).WithField("preimage", preimage).Error("Failed to decode preimage for SettleHoldInvoice")
		return err
	}
	if len(decodedPreimage) != 32 {
		return errors.New("preimage must be 32 bytes")
	}

	paymentHash256 := sha256.New()
	paymentHash256.Write(decodedPreimage)
	paymentHashBytes := paymentHash256.Sum(nil)
	paymentHash := hex.EncodeToString(paymentHashBytes)

	paymentDetails := ls.node.Payment(paymentHash)

	if paymentDetails == nil {
		logger.Logger.WithField("payment_hash", paymentHash).Error("SettleHoldInvoice: Could not find payment by derived hash")
		return errors.New("payment not found for derived hash")
	}
	if paymentDetails.AmountMsat == nil {
		logger.Logger.WithField("payment_hash", paymentHash).Error("SettleHoldInvoice: Payment has no amount_msat")
		return errors.New("payment has no amount_msat")
	}

	err = ls.node.Bolt11Payment().ClaimForHash(paymentHash, *paymentDetails.AmountMsat, preimage)
	if err != nil {
		logger.Logger.WithError(err).WithField("preimage", preimage).WithField("derived_payment_hash", paymentHash).Error("SettleHoldInvoice failed")
	}
	return err
}

func GetVssNodeIdentifier(keys keys.Keys) (string, error) {
	key, err := keys.DeriveKey([]uint32{bip32.FirstHardenedChild + 2})

	if err != nil {
		return "", err
	}

	// return a 6-character hex string of the hash of a derived key to ensure if same user
	// runs multiple hubs with different mnemonics, they are all
	// saved in the VSS under different user_tokens.
	pubkeyHash256 := sha256.New()
	pubkeyHash256.Write(key.Key)
	pubkeyHashBytes := pubkeyHash256.Sum(nil)
	return hex.EncodeToString(pubkeyHashBytes[0:3]), nil
}

func getResetStateRequest(cfg config.Config) *ldk_node.ResetState {
	resetKey, err := cfg.Get(resetRouterKey, "")
	if err != nil {
		logger.Logger.Error("Failed to retrieve ResetRouter key")
		return nil
	}

	if resetKey == "" {
		return nil
	}

	err = cfg.SetUpdate(resetRouterKey, "", "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to remove reset router key")
		return nil
	}

	var ret ldk_node.ResetState

	switch resetKey {
	case "ALL":
		ret = ldk_node.ResetStateAll
	case "Scorer":
		ret = ldk_node.ResetStateScorer
	case "NetworkGraph":
		ret = ldk_node.ResetStateNetworkGraph
	case "NodeMetrics":
		ret = ldk_node.ResetStateNodeMetrics
	default:
		logger.Logger.WithField("key", resetKey).Error("Unknown reset router key")
		return nil
	}

	return &ret
}

func (ls *LDKService) ConsumeEvent(ctx context.Context, event *events.Event, globalProperties map[string]interface{}) {
	if event.Event == "nwc_alby_account_connected" {
		// backup existing channels to the user's Alby Account on first connect
		ls.backupChannels()
	}
}
