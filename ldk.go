package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	//"github.com/getAlby/ldk-node-go/ldk_node"
	"github.com/getAlby/nostr-wallet-connect/ldk_node"
	"github.com/getAlby/nostr-wallet-connect/models/lnclient"
	"github.com/getAlby/nostr-wallet-connect/models/lsp"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
)

type LDKService struct {
	svc                       *Service
	workdir                   string
	node                      *ldk_node.LdkNode
	cancelLdkEventListenerCtx context.CancelFunc
	subscribeLdkEvents        func() chan ldk_node.Event
	unsubscribeLdkEvents      func(chan ldk_node.Event)
}

func NewLDKService(svc *Service, mnemonic, workDir string) (result lnclient.LNClient, err error) {
	if mnemonic == "" || workDir == "" {
		return nil, errors.New("one or more required LDK configuration are missing")
	}

	//create dir if not exists
	newpath := filepath.Join(".", workDir+"-lsp3")
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
	}
	config.AnchorChannelsConfig.TrustedPeersNoReserve = []string{
		lsp.OlympusLSP().Pubkey,
	}

	config.ListeningAddresses = &listeningAddresses
	config.LogDirPath = &logDirPath
	config.LogLevel = ldk_node.LogLevelTrace
	builder := ldk_node.BuilderFromConfig(config)
	builder.SetEntropyBip39Mnemonic(mnemonic, nil)
	builder.SetNetwork("bitcoin")
	builder.SetLiquiditySourceLsps2("52.88.33.119:9735", lsp.VoltageLSP().Pubkey, nil) // Temporary hack to work with voltage, remove when https://github.com/lightningdevkit/rust-lightning/issues/2914 is merged
	builder.SetEsploraServer("https://blockstream.info/api")
	builder.SetGossipSourceRgs("https://rapidsync.lightningdevkit.org/snapshot")
	builder.SetStorageDirPath(filepath.Join(newpath, "./storage"))

	//builder.SetLogDirPath (filepath.Join(newpath, "./logs")); // missing?
	node, err := builder.Build()

	if err != nil {
		svc.Logger.Errorf("Failed to create LDK node: %v", err)
		return nil, err
	}

	err = node.Start()
	if err != nil {
		svc.Logger.Errorf("Failed to start LDK node: %v", err)
		return nil, err
	}

	// TODO: move this event handler code
	ldkEventListenerCtx, cancelLdkEventListenerCtx := context.WithCancel(context.Background())
	ldkEventHandlers := []chan ldk_node.Event{}
	var ldkEventHandlersMutex sync.Mutex

	subscribeLdkEvents := func() chan ldk_node.Event {
		ldkEventHandler := make(chan ldk_node.Event)
		ldkEventHandlersMutex.Lock()
		ldkEventHandlers = append(ldkEventHandlers, ldkEventHandler)
		ldkEventHandlersMutex.Unlock()
		return ldkEventHandler
	}

	unsubscribeLdkEvents := func(eventHandler chan ldk_node.Event) {
		ldkEventHandlersMutex.Lock()
		for i := 0; i < len(ldkEventHandlers); i++ {
			if eventHandler == ldkEventHandlers[i] {
				// Replace the element to be removed with the last element of the slice
				ldkEventHandlers[i] = ldkEventHandlers[len(ldkEventHandlers)-1]
				// Slice off the last element
				ldkEventHandlers = ldkEventHandlers[:len(ldkEventHandlers)-1]
				break
			}
		}
		ldkEventHandlersMutex.Unlock()
	}

	go func() {
		for {
			select {
			case <-ldkEventListenerCtx.Done():
				return
			default:
				event := node.NextEvent()
				if event == nil {
					time.Sleep(100 * time.Millisecond)
					continue
				}
				ldkEventHandlersMutex.Lock()
				svc.Logger.Infof("Received LDK event %+v (%d listeners)", *event, len(ldkEventHandlers))
				for _, eventHandler := range ldkEventHandlers {
					eventHandler <- *event
				}
				ldkEventHandlersMutex.Unlock()

				node.EventHandled()
			}
		}
	}()

	gs := LDKService{
		workdir: newpath,
		node:    node,
		//listener: &listener,
		svc:                       svc,
		cancelLdkEventListenerCtx: cancelLdkEventListenerCtx,
		subscribeLdkEvents:        subscribeLdkEvents,
		unsubscribeLdkEvents:      unsubscribeLdkEvents,
	}

	nodeId := node.NodeId()

	if err != nil {
		return nil, err
	}

	log.Printf("Connected to node ID: %v", nodeId)

	return &gs, nil
}

func (gs *LDKService) Shutdown() error {
	gs.svc.Logger.Infof("shutting down LDK client")
	gs.cancelLdkEventListenerCtx()
	gs.node.Destroy()

	return nil
}

func (gs *LDKService) SendPaymentSync(ctx context.Context, payReq string) (preimage string, err error) {
	paymentHash, err := gs.node.SendPayment(payReq)
	if err != nil {
		gs.svc.Logger.Errorf("SendPayment failed: %v", err)
		return "", err
	}

	eventListener := gs.subscribeLdkEvents()
	defer gs.unsubscribeLdkEvents(eventListener)
	for start := time.Now(); time.Since(start) < time.Second*60; {
		event := <-eventListener

		eventPaymentSuccessful, isEventPaymentSuccessfulEvent := event.(ldk_node.EventPaymentSuccessful)
		eventPaymentFailed, isEventPaymentFailedEvent := event.(ldk_node.EventPaymentFailed)

		if isEventPaymentSuccessfulEvent && eventPaymentSuccessful.PaymentHash == paymentHash {
			gs.svc.Logger.Infof("Got payment success event")
			payment := gs.node.Payment(paymentHash)
			if payment == nil {
				gs.svc.Logger.Errorf("Couldn't find payment by payment hash: %v", paymentHash)
				return "", errors.New("Payment not found")
			}

			if payment.Secret == nil {
				gs.svc.Logger.Errorf("No payment secret for payment hash: %v", paymentHash)
				return "", errors.New("Payment secret not found")
			}
			preimage = *payment.Secret
			break
		}
		if isEventPaymentFailedEvent && eventPaymentFailed.PaymentHash == paymentHash {
			var failureReason ldk_node.PaymentFailureReason = 0 // unset
			if eventPaymentFailed.Reason != nil {
				failureReason = *eventPaymentFailed.Reason
			}
			gs.svc.Logger.Errorf("Payment failed: %v %v", paymentHash, failureReason)
			return "", fmt.Errorf("payment failed: %v", failureReason)
		}
	}
	if preimage == "" {
		return "", errors.New("Payment timed out")
	}

	return preimage, nil
}

func (gs *LDKService) SendKeysend(ctx context.Context, amount int64, destination, preimage string, custom_records []lnclient.TLVRecord) (preImage string, err error) {

	if len(custom_records) > 0 {
		log.Printf("FIXME: TLVs not supported")
	}

	paymentHash, err := gs.node.SendSpontaneousPayment(uint64(amount), destination)
	if err != nil {
		gs.svc.Logger.Errorf("Keysend failed: %v", err)
		return "", err
	}

	payment := gs.node.Payment(paymentHash)
	if payment == nil {
		gs.svc.Logger.Errorf("Couldn't find payment by payment hash: %v", paymentHash)
		return "", errors.New("Payment not found")
	}

	if payment.Preimage == nil {
		gs.svc.Logger.Errorf("No payment preimage found for payment hash: %v", paymentHash)
		return "", errors.New("no preimage in payment")
	}

	return *payment.Preimage, nil
}

func (gs *LDKService) GetBalance(ctx context.Context) (balance int64, err error) {
	channels := gs.node.ListChannels()

	balance = 0
	for _, channel := range channels {
		balance += int64(channel.OutboundCapacityMsat)
	}

	return balance, nil
}

func (gs *LDKService) MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64) (transaction *Nip47Transaction, err error) {

	// TODO: support passing description hash
	invoice, err := gs.node.ReceivePayment(uint64(amount),
		description,
		uint32(expiry))

	if err != nil {
		gs.svc.Logger.Errorf("MakeInvoice failed: %v", err)
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
		CreatedAt:       time.Now().Unix(),
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

	transaction = ldkPaymentToTransaction(payment)

	return transaction, nil
}

func (gs *LDKService) ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaid bool, invoiceType string) (transactions []Nip47Transaction, err error) {
	transactions = []Nip47Transaction{}

	payments := gs.node.ListPayments()

	for _, payment := range payments {
		if payment.Status == ldk_node.PaymentStatusSucceeded {
			transactions = append(transactions, *ldkPaymentToTransaction(&payment))
		}
	}

	// sort by created date descending
	/*sort.SliceStable(transactions, func(i, j int) bool {
		return transactions[i].CreatedAt > transactions[j].CreatedAt
	})*/

	return transactions, nil
}

func (gs *LDKService) GetInfo(ctx context.Context) (info *lnclient.NodeInfo, err error) {
	return &lnclient.NodeInfo{
		// Alias:       nodeInfo.Alias,
		// Color:       nodeInfo.Color,
		Pubkey: gs.node.NodeId(),
		// Network:     nodeInfo.Network,
		// BlockHeight: nodeInfo.BlockHeight,
		BlockHash: "",
	}, nil
}

func (gs *LDKService) ListChannels(ctx context.Context) ([]lnclient.Channel, error) {

	ldkChannels := gs.node.ListChannels()
	channels := []lnclient.Channel{}

	for _, ldkChannel := range ldkChannels {
		channels = append(channels, lnclient.Channel{
			LocalBalance:  int64(ldkChannel.OutboundCapacityMsat),
			RemoteBalance: int64(ldkChannel.InboundCapacityMsat),
			RemotePubkey:  ldkChannel.CounterpartyNodeId,
			Id:            ldkChannel.ChannelId,
			Active:        ldkChannel.IsChannelReady && ldkChannel.IsUsable, // TODO: confirm
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
		gs.svc.Logger.Errorf("ConnectPeer failed: %v", err)
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
		gs.svc.Logger.Errorf("ConnectPeer failed: %v", err)
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

	gs.svc.Logger.Infof("Opening channel with: %v", foundPeer.NodeId)
	userChannelId, err := gs.node.ConnectOpenChannel(foundPeer.NodeId, foundPeer.Address, uint64(openChannelRequest.Amount), nil, nil, openChannelRequest.Public)
	if err != nil {
		gs.svc.Logger.Errorf("OpenChannel failed: %v", err)
		return nil, err
	}

	// userChannelId allows to locally keep track of the channel
	gs.svc.Logger.Infof("Funded channel: %v", userChannelId)

	eventListener := gs.subscribeLdkEvents()
	defer gs.unsubscribeLdkEvents(eventListener)
	for start := time.Now(); time.Since(start) < time.Second*60; {
		event := <-eventListener

		channelPendingEvent, isChannelPendingEvent := event.(ldk_node.EventChannelPending)

		if !isChannelPendingEvent {
			continue
		}

		return &lnclient.OpenChannelResponse{
			FundingTxId: channelPendingEvent.FundingTxo.Txid,
		}, nil
	}

	return nil, errors.New("open channel timeout")
}

func (gs *LDKService) GetNewOnchainAddress(ctx context.Context) (string, error) {
	address, err := gs.node.NewOnchainAddress()
	if err != nil {
		gs.svc.Logger.Errorf("NewOnchainAddress failed: %v", err)
		return "", err
	}
	return address, nil
}

func (gs *LDKService) GetOnchainBalance(ctx context.Context) (int64, error) {
	balances := gs.node.ListBalances()
	gs.svc.Logger.Infof("SpendableOnchainBalanceSats: %v", balances.SpendableOnchainBalanceSats)
	return int64(balances.SpendableOnchainBalanceSats), nil
}

func ldkPaymentToTransaction(payment *ldk_node.PaymentDetails) *Nip47Transaction {
	transactionType := "incoming"
	if payment.Direction == ldk_node.PaymentDirectionOutbound {
		transactionType = "outgoing"
	}

	preimage := ""
	var settledAt *int64
	if payment.Status == ldk_node.PaymentStatusSucceeded {
		if payment.Preimage != nil {

			preimage = *payment.Preimage
		} else if payment.Secret != nil {
			preimage = *payment.Secret
		}
		// TODO: use payment settle time
		now := time.Now().Unix()
		settledAt = &now
	}

	var amount uint64 = 0
	if payment.AmountMsat != nil {
		amount = *payment.AmountMsat
	}

	return &Nip47Transaction{
		Type: transactionType,
		// TODO: get bolt11 invoice from payment
		//Invoice: payment.,
		Preimage:    preimage,
		PaymentHash: payment.Hash,
		SettledAt:   settledAt,
		Amount:      int64(amount),
	}
}
