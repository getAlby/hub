package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/getAlby/ldk-node-go/ldk_node"
	"github.com/getAlby/nostr-wallet-connect/events"
	"github.com/getAlby/nostr-wallet-connect/models/config"
	"github.com/getAlby/nostr-wallet-connect/models/lnclient"
	"github.com/getAlby/nostr-wallet-connect/models/lsp"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
)

type LDKService struct {
	svc                 *Service
	workdir             string
	node                *ldk_node.Node
	ldkEventBroadcaster LDKEventBroadcaster
	cancel              context.CancelFunc
	network             string
	eventLogger         events.EventLogger
}

func NewLDKService(svc *Service, mnemonic, workDir string, network string, esploraServer string, gossipSource string) (result lnclient.LNClient, err error) {
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
	config := ldk_node.DefaultConfig()
	listeningAddresses := []string{
		"0.0.0.0:9735",
	}
	config.TrustedPeers0conf = []string{
		lsp.VoltageLSP().Pubkey,
		lsp.OlympusLSP().Pubkey,
		lsp.AlbyPlebsLSP().Pubkey,
	}
	config.AnchorChannelsConfig.TrustedPeersNoReserve = []string{
		lsp.OlympusLSP().Pubkey,
		lsp.AlbyPlebsLSP().Pubkey,
	}

	config.ListeningAddresses = &listeningAddresses
	config.LogDirPath = &logDirPath
	logLevel, err := strconv.Atoi(svc.cfg.Env.LDKLogLevel)
	if err == nil {
		config.LogLevel = ldk_node.LogLevel(logLevel)
	}
	builder := ldk_node.BuilderFromConfig(config)
	builder.SetEntropyBip39Mnemonic(mnemonic, nil)
	builder.SetNetwork(network)
	builder.SetEsploraServer(esploraServer)
	builder.SetGossipSourceRgs(gossipSource)
	builder.SetStorageDirPath(filepath.Join(newpath, "./storage"))

	// TODO: remove when https://github.com/lightningdevkit/rust-lightning/issues/2914 is merged
	// LDK default HTLC inflight value is 10% of the channel size. If an LSPS service is configured this will be set to 0.
	builder.SetLiquiditySourceLsps2("52.88.33.119:9735", lsp.VoltageLSP().Pubkey, nil)

	//builder.SetLogDirPath (filepath.Join(newpath, "./logs")); // missing?
	node, err := builder.Build()

	if err != nil {
		svc.Logger.Errorf("Failed to create LDK node: %v", err)
		return nil, err
	}

	ldkEventConsumer := make(chan *ldk_node.Event)
	ctx, cancel := context.WithCancel(svc.ctx)
	ldkEventBroadcaster := NewLDKEventBroadcaster(svc.Logger, ctx, ldkEventConsumer)

	ls := LDKService{
		workdir:             newpath,
		node:                node,
		svc:                 svc,
		cancel:              cancel,
		ldkEventBroadcaster: ldkEventBroadcaster,
		network:             network,
		eventLogger:         svc.EventLogger,
	}

	// check for and forward new LDK events to LDKEventBroadcaster (through ldkEventConsumer)
	go func() {
		for {
			select {
			case <-ctx.Done():
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

				ls.logLdkEvent(ctx, event)
				ldkEventConsumer <- event

				node.EventHandled()
			}
		}
	}()

	err = node.Start()
	if err != nil {
		svc.Logger.Errorf("Failed to start LDK node: %v", err)
		return nil, err
	}

	nodeId := node.NodeId()

	if err != nil {
		return nil, err
	}

	svc.Logger.WithFields(logrus.Fields{
		"nodeId": nodeId,
	}).Info("Connected to LDK node")

	return &ls, nil
}

func (gs *LDKService) Shutdown() error {
	gs.svc.Logger.Infof("shutting down LDK client")
	gs.cancel()
	gs.node.Destroy()

	return nil
}

func (gs *LDKService) SendPaymentSync(ctx context.Context, invoice string) (preimage string, err error) {
	paymentRequest, err := decodepay.Decodepay(invoice)
	if err != nil {
		gs.svc.Logger.WithFields(logrus.Fields{
			"bolt11": invoice,
		}).Errorf("Failed to decode bolt11 invoice: %v", err)

		return "", err
	}

	maxSendable := gs.getMaxSendable()
	if paymentRequest.MSatoshi > maxSendable {
		gs.eventLogger.Log(&events.Event{
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
	ldkEventSubscription := gs.ldkEventBroadcaster.Subscribe()
	defer gs.ldkEventBroadcaster.CancelSubscription(ldkEventSubscription)

	paymentHash, err := gs.node.Bolt11Payment().Send(invoice)
	if err != nil {
		gs.svc.Logger.WithError(err).Error("SendPayment failed")
		return "", err
	}
	fee := uint64(0)

	for start := time.Now(); time.Since(start) < time.Second*60; {
		event := <-ldkEventSubscription

		eventPaymentSuccessful, isEventPaymentSuccessfulEvent := (*event).(ldk_node.EventPaymentSuccessful)
		eventPaymentFailed, isEventPaymentFailedEvent := (*event).(ldk_node.EventPaymentFailed)

		if isEventPaymentSuccessfulEvent && eventPaymentSuccessful.PaymentHash == paymentHash {
			gs.svc.Logger.Info("Got payment success event")
			payment := gs.node.Payment(paymentHash)
			if payment == nil {
				gs.svc.Logger.Errorf("Couldn't find payment by payment hash: %v", paymentHash)
				return "", errors.New("Payment not found")
			}

			bolt11PaymentKind, ok := payment.Kind.(ldk_node.PaymentKindBolt11)

			if !ok {
				gs.svc.Logger.WithFields(logrus.Fields{
					"payment": payment,
				}).Error("Payment is not a bolt11 kind")
			}

			if bolt11PaymentKind.Preimage == nil {
				gs.svc.Logger.Errorf("No payment preimage for payment hash: %v", paymentHash)
				return "", errors.New("Payment preimage not found")
			}
			preimage = *bolt11PaymentKind.Preimage

			if eventPaymentSuccessful.FeePaidMsat != nil {
				fee = *eventPaymentSuccessful.FeePaidMsat
			}
			break
		}
		if isEventPaymentFailedEvent && eventPaymentFailed.PaymentHash == paymentHash {
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

			gs.svc.Logger.WithFields(logrus.Fields{
				"paymentHash":          paymentHash,
				"failureReason":        failureReason,
				"failureReasonMessage": failureReasonMessage,
			}).Error("Received payment failed event")

			return "", fmt.Errorf("received payment failed event: %v %s", failureReason, failureReasonMessage)
		}
	}
	if preimage == "" {
		// TODO: this doesn't necessarily mean it will fail - we should return a different response
		return "", errors.New("Payment timed out")
	}

	gs.svc.Logger.WithFields(logrus.Fields{
		"duration": time.Since(paymentStart).Milliseconds(),
		"fee":      fee,
	}).Info("Successful payment")
	return preimage, nil
}

func (gs *LDKService) SendKeysend(ctx context.Context, amount int64, destination, preimage string, custom_records []lnclient.TLVRecord) (preImage string, err error) {
	paymentStart := time.Now()
	customTlvs := []ldk_node.TlvEntry{}

	for _, customRecord := range custom_records {
		customTlvs = append(customTlvs, ldk_node.TlvEntry{
			Type:  customRecord.Type,
			Value: []uint8(customRecord.Value),
		})
	}

	ldkEventSubscription := gs.ldkEventBroadcaster.Subscribe()
	defer gs.ldkEventBroadcaster.CancelSubscription(ldkEventSubscription)

	paymentHash, err := gs.node.SpontaneousPayment().Send(uint64(amount), destination, customTlvs)
	if err != nil {
		gs.svc.Logger.WithError(err).Error("Keysend failed")
		return "", err
	}

	gs.svc.Logger.Infof("TODO: listen for events %v", paymentHash)

	fee := uint64(0)

	for start := time.Now(); time.Since(start) < time.Second*60; {
		event := <-ldkEventSubscription

		eventPaymentSuccessful, isEventPaymentSuccessfulEvent := (*event).(ldk_node.EventPaymentSuccessful)
		eventPaymentFailed, isEventPaymentFailedEvent := (*event).(ldk_node.EventPaymentFailed)

		if isEventPaymentSuccessfulEvent && eventPaymentSuccessful.PaymentHash == paymentHash {
			gs.svc.Logger.Info("Got payment success event")
			payment := gs.node.Payment(paymentHash)
			if payment == nil {
				gs.svc.Logger.Errorf("Couldn't find payment by payment hash: %v", paymentHash)
				return "", errors.New("Payment not found")
			}

			spontaneousPaymentKind, ok := payment.Kind.(ldk_node.PaymentKindSpontaneous)

			if !ok {
				gs.svc.Logger.WithFields(logrus.Fields{
					"payment": payment,
				}).Error("Payment is not a spontaneous kind")
			}

			if spontaneousPaymentKind.Preimage == nil {
				gs.svc.Logger.Errorf("No payment preimage for payment hash: %v", paymentHash)
				return "", errors.New("Payment preimage not found")
			}
			preimage = *spontaneousPaymentKind.Preimage

			if eventPaymentSuccessful.FeePaidMsat != nil {
				fee = *eventPaymentSuccessful.FeePaidMsat
			}
			break
		}
		if isEventPaymentFailedEvent && eventPaymentFailed.PaymentHash == paymentHash {
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

			gs.svc.Logger.WithFields(logrus.Fields{
				"paymentHash":          paymentHash,
				"failureReason":        failureReason,
				"failureReasonMessage": failureReasonMessage,
			}).Error("Received payment failed event")

			return "", fmt.Errorf("payment failed event: %v %s", failureReason, failureReasonMessage)
		}
	}
	if preimage == "" {
		// TODO: this doesn't necessarily mean it will fail - we should return a different response
		return "", errors.New("keysend payment timed out")
	}

	gs.svc.Logger.WithFields(logrus.Fields{
		"duration": time.Since(paymentStart).Milliseconds(),
		"fee":      fee,
	}).Info("Successful keysend payment")
	return preimage, nil
}

func (gs *LDKService) GetBalance(ctx context.Context) (balance int64, err error) {
	channels := gs.node.ListChannels()

	balance = 0
	for _, channel := range channels {
		if channel.IsUsable {
			balance += int64(channel.OutboundCapacityMsat)
		}
	}

	return balance, nil
}

func (gs *LDKService) getMaxReceivable() int64 {
	var receivable int64 = 0
	channels := gs.node.ListChannels()
	for _, channel := range channels {
		if channel.IsUsable {
			receivable += min(int64(channel.InboundCapacityMsat), int64(*channel.InboundHtlcMaximumMsat))
		}
	}
	return int64(receivable)
}

func (gs *LDKService) getMaxSendable() int64 {
	var spendable int64 = 0
	channels := gs.node.ListChannels()
	for _, channel := range channels {
		if channel.IsUsable {
			spendable += min(int64(channel.OutboundCapacityMsat), int64(*channel.CounterpartyOutboundHtlcMaximumMsat))
		}
	}
	return int64(spendable)
}

func (gs *LDKService) MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64) (transaction *Nip47Transaction, err error) {

	maxReceivable := gs.getMaxReceivable()

	if amount > maxReceivable {
		gs.eventLogger.Log(&events.Event{
			Event: "nwc_incoming_liquidity_required",
			Properties: map[string]interface{}{
				//"amount":         amount / 1000,
				//"max_receivable": maxReceivable,
				//"num_channels":   len(gs.node.ListChannels()),
				"node_type": config.LDKBackendType,
			},
		})
	}

	// TODO: support passing description hash
	invoice, err := gs.node.Bolt11Payment().Receive(uint64(amount),
		description,
		uint32(expiry))

	if err != nil {
		gs.svc.Logger.WithError(err).Error("MakeInvoice failed")
		return nil, err
	}

	var expiresAt *int64
	paymentRequest, err := decodepay.Decodepay(invoice)
	if err != nil {
		gs.svc.Logger.WithFields(logrus.Fields{
			"bolt11": invoice,
		}).Errorf("Failed to decode bolt11 invoice: %v", err)

		return nil, err
	}
	expiresAtUnix := time.UnixMilli(int64(paymentRequest.CreatedAt) * 1000).Add(time.Duration(paymentRequest.Expiry) * time.Second).Unix()
	expiresAt = &expiresAtUnix
	description = paymentRequest.Description
	descriptionHash = paymentRequest.DescriptionHash

	transaction = &Nip47Transaction{
		Type:            "incoming",
		Invoice:         invoice,
		PaymentHash:     paymentRequest.PaymentHash,
		Amount:          amount,
		CreatedAt:       int64(paymentRequest.CreatedAt),
		ExpiresAt:       expiresAt,
		Description:     description,
		DescriptionHash: descriptionHash,
	}

	return transaction, nil
}

func (gs *LDKService) LookupInvoice(ctx context.Context, paymentHash string) (transaction *Nip47Transaction, err error) {

	payment := gs.node.Payment(paymentHash)
	if payment == nil {
		gs.svc.Logger.Errorf("Couldn't find payment by payment hash: %v", paymentHash)
		return nil, errors.New("Payment not found")
	}

	transaction, err = gs.ldkPaymentToTransaction(payment)

	if err != nil {
		gs.svc.Logger.Errorf("Failed to map transaction: %v", err)
		return nil, err
	}

	return transaction, nil
}

func (gs *LDKService) ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaid bool, invoiceType string) (transactions []Nip47Transaction, err error) {
	transactions = []Nip47Transaction{}

	// TODO: support pagination
	payments := gs.node.ListPayments()

	for _, payment := range payments {
		if payment.Status == ldk_node.PaymentStatusSucceeded {
			transaction, err := gs.ldkPaymentToTransaction(&payment)

			if err != nil {
				gs.svc.Logger.Errorf("Failed to map transaction: %v", err)
				continue
			}

			transactions = append(transactions, *transaction)
		}
	}

	// sort by created date descending
	sort.SliceStable(transactions, func(i, j int) bool {
		return transactions[i].CreatedAt > transactions[j].CreatedAt
	})

	// locally limit for now
	if len(transactions) > int(limit) {
		transactions = transactions[:limit]
	}

	return transactions, nil
}

func (gs *LDKService) GetInfo(ctx context.Context) (info *lnclient.NodeInfo, err error) {
	// TODO: should alias, color be configured in LDK-node? or can we manage them in NWC?
	// an alias is only needed if the user has public channels and wants their node to be publicly visible?
	status := gs.node.Status()
	return &lnclient.NodeInfo{
		Alias:       "NWC",
		Color:       "#897FFF",
		Pubkey:      gs.node.NodeId(),
		Network:     gs.network,
		BlockHeight: status.CurrentBestBlock.Height,
		BlockHash:   status.CurrentBestBlock.BlockHash,
	}, nil
}

func (gs *LDKService) ListChannels(ctx context.Context) ([]lnclient.Channel, error) {

	ldkChannels := gs.node.ListChannels()

	channels := []lnclient.Channel{}

	gs.svc.Logger.WithFields(logrus.Fields{
		"channels": ldkChannels,
	}).Debug("Listed Channels")

	for _, ldkChannel := range ldkChannels {
		channels = append(channels, lnclient.Channel{
			LocalBalance:  int64(ldkChannel.OutboundCapacityMsat),
			RemoteBalance: int64(ldkChannel.InboundCapacityMsat),
			RemotePubkey:  ldkChannel.CounterpartyNodeId,
			Id:            ldkChannel.UserChannelId, // CloseChannel takes the UserChannelId
			Active:        ldkChannel.IsUsable,      // superset of ldkChannel.IsReady
			Public:        ldkChannel.IsPublic,
		})
	}

	return channels, nil
}

func (gs *LDKService) GetNodeConnectionInfo(ctx context.Context) (nodeConnectionInfo *lnclient.NodeConnectionInfo, err error) {
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
		gs.svc.Logger.WithError(err).Error("ConnectPeer failed")
		return nil, err
	}*/

	return &lnclient.NodeConnectionInfo{
		Pubkey: gs.node.NodeId(),
		//Address: parts[0],
		//Port:    port,
	}, nil
}

func (gs *LDKService) ConnectPeer(ctx context.Context, connectPeerRequest *lnclient.ConnectPeerRequest) error {
	err := gs.node.Connect(connectPeerRequest.Pubkey, connectPeerRequest.Address+":"+strconv.Itoa(int(connectPeerRequest.Port)), true)
	if err != nil {
		gs.svc.Logger.WithError(err).Error("ConnectPeer failed")
		return err
	}

	return nil
}

func (gs *LDKService) OpenChannel(ctx context.Context, openChannelRequest *lnclient.OpenChannelRequest) (*lnclient.OpenChannelResponse, error) {
	peers := gs.node.ListPeers()
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

	ldkEventSubscription := gs.ldkEventBroadcaster.Subscribe()
	defer gs.ldkEventBroadcaster.CancelSubscription(ldkEventSubscription)

	gs.svc.Logger.Infof("Opening channel with: %v", foundPeer.NodeId)
	userChannelId, err := gs.node.ConnectOpenChannel(foundPeer.NodeId, foundPeer.Address, uint64(openChannelRequest.Amount), nil, nil, openChannelRequest.Public)
	if err != nil {
		gs.svc.Logger.WithError(err).Error("OpenChannel failed")
		return nil, err
	}

	// userChannelId allows to locally keep track of the channel (and is also used to close the channel)
	gs.svc.Logger.Infof("Funded channel: %v", userChannelId)

	for start := time.Now(); time.Since(start) < time.Second*60; {
		event := <-ldkEventSubscription

		channelPendingEvent, isChannelPendingEvent := (*event).(ldk_node.EventChannelPending)
		channelClosedEvent, isChannelClosedEvent := (*event).(ldk_node.EventChannelClosed)

		if isChannelClosedEvent {
			gs.svc.Logger.WithFields(logrus.Fields{
				"event": channelClosedEvent,
			})
			return nil, fmt.Errorf("failed to open channel: %+v", *channelClosedEvent.Reason)
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

func (gs *LDKService) CloseChannel(ctx context.Context, closeChannelRequest *lnclient.CloseChannelRequest) (*lnclient.CloseChannelResponse, error) {
	gs.svc.Logger.WithFields(logrus.Fields{
		"request": closeChannelRequest,
	}).Info("Closing Channel")
	// TODO: support passing force option
	err := gs.node.CloseChannel(closeChannelRequest.ChannelId, closeChannelRequest.NodeId, false)
	if err != nil {
		gs.svc.Logger.WithError(err).Error("CloseChannel failed")
		return nil, err
	}
	return &lnclient.CloseChannelResponse{}, nil
}

func (gs *LDKService) GetNewOnchainAddress(ctx context.Context) (string, error) {
	address, err := gs.node.OnchainPayment().NewAddress()
	if err != nil {
		gs.svc.Logger.WithError(err).Error("NewOnchainAddress failed")
		return "", err
	}
	return address, nil
}

func (gs *LDKService) GetOnchainBalance(ctx context.Context) (*lnclient.OnchainBalanceResponse, error) {
	balances := gs.node.ListBalances()
	gs.svc.Logger.WithFields(logrus.Fields{
		"balances": balances,
	}).Debug("Listed Balances")
	return &lnclient.OnchainBalanceResponse{
		Spendable: int64(balances.SpendableOnchainBalanceSats),
		Total:     int64(balances.TotalOnchainBalanceSats),
	}, nil
}

func (gs *LDKService) RedeemOnchainFunds(ctx context.Context, toAddress string) (string, error) {
	txId, err := gs.node.OnchainPayment().SendAllToAddress(toAddress)
	if err != nil {
		gs.svc.Logger.WithError(err).Error("SendAllToOnchainAddress failed")
		return "", err
	}
	return txId, nil
}

func (ls *LDKService) ResetRouter(ctx context.Context) error {
	// HACK: to ensure the router is reset correctly we must stop the node first.
	err := ls.node.Stop()
	if err != nil {
		ls.svc.Logger.WithError(err).Error("Failed to stop the node")
	}

	err = ls.node.ResetRouter()
	if err != nil {
		ls.svc.Logger.WithError(err).Error("ResetRouter failed")
	}

	return err
}

func (gs *LDKService) ldkPaymentToTransaction(payment *ldk_node.PaymentDetails) (*Nip47Transaction, error) {
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

	bolt11PaymentKind, isBolt11PaymentKind := payment.Kind.(ldk_node.PaymentKindBolt11)

	if isBolt11PaymentKind && bolt11PaymentKind.Bolt11Invoice != nil {
		bolt11Invoice = *bolt11PaymentKind.Bolt11Invoice
		paymentRequest, err := decodepay.Decodepay(strings.ToLower(bolt11Invoice))
		if err != nil {
			gs.svc.Logger.WithFields(logrus.Fields{
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

	var amount uint64 = 0
	if payment.AmountMsat != nil {
		amount = *payment.AmountMsat
	}

	return &Nip47Transaction{
		Type:        transactionType,
		Preimage:    preimage,
		PaymentHash: paymentHash,
		SettledAt:   settledAt,
		Amount:      int64(amount),
		Invoice:     bolt11Invoice,
		//FeesPaid:        payment.FeeMsat,
		CreatedAt:       createdAt,
		Description:     description,
		DescriptionHash: descriptionHash,
		ExpiresAt:       expiresAt,
	}, nil
}

func (ls *LDKService) logLdkEvent(ctx context.Context, event *ldk_node.Event) {
	ls.svc.Logger.WithFields(logrus.Fields{
		"event": event,
	}).Info("Received LDK event")

	switch v := (*event).(type) {
	case ldk_node.EventChannelReady:
		ls.eventLogger.Log(&events.Event{
			Event: "nwc_channel_ready",
			Properties: map[string]interface{}{
				// "counterparty_node_id": v.CounterpartyNodeId,
				"node_type": config.LDKBackendType,
			},
		})
	case ldk_node.EventChannelClosed:
		ls.eventLogger.Log(&events.Event{
			Event: "nwc_channel_closed",
			Properties: map[string]interface{}{
				// "counterparty_node_id": v.CounterpartyNodeId,
				// "reason":               fmt.Sprintf("%+v", v.Reason),
				"node_type": config.LDKBackendType,
			},
		})
	case ldk_node.EventPaymentReceived:
		ls.eventLogger.Log(&events.Event{
			Event: "nwc_payment_received",
			Properties: map[string]interface{}{
				"payment_hash": v.PaymentHash,
				"amount":       v.AmountMsat / 1000,
				"node_type":    config.LDKBackendType,
			},
		})
	}

}
