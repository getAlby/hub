package main

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/getAlby/nostr-wallet-connect/ldk_node" // TODO: include this from an external library
	"github.com/getAlby/nostr-wallet-connect/models/lnclient"
)

type LDKService struct {
	svc     *Service
	workdir string
	node    *ldk_node.LdkNode
}

func NewLDKService(svc *Service, mnemonic, workDir string) (result lnclient.LNClient, err error) {
	if mnemonic == "" || workDir == "" {
		return nil, errors.New("one or more required LDK configuration are missing")
	}

	//create dir if not exists
	newpath := filepath.Join(".", workDir)
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
	config.ListeningAddresses = &listeningAddresses
	config.LogDirPath = &logDirPath
	builder := ldk_node.BuilderFromConfig(config)
	builder.SetEntropyBip39Mnemonic(mnemonic, nil)
	builder.SetNetwork("bitcoin")
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

	gs := LDKService{
		workdir: newpath,
		node:    node,
		//listener: &listener,
		svc: svc,
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
	gs.node.Destroy()

	return nil
}

func (gs *LDKService) SendPaymentSync(ctx context.Context, payReq string) (preimage string, err error) {
	paymentHash, err := gs.node.SendPayment(payReq)
	if err != nil {
		gs.svc.Logger.Errorf("SendPayment failed: %v", err)
		return "", err
	}

	for start := time.Now(); time.Since(start) < time.Second*60; {
		gs.node.WaitNextEvent()

		payment := gs.node.Payment(paymentHash)
		if payment == nil {
			gs.svc.Logger.Errorf("Couldn't find payment by payment hash: %v", paymentHash)
			return "", errors.New("Payment not found")
		}

		if payment.Secret != nil {
			preimage = *payment.Secret
			gs.svc.Logger.Infof("Payment succeeded")
			break
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
	//paymentHash := gs.node.SendSpontaneousPayment(uint64(amount), destination)
	// TODO: get payment by hash
	//return payResponse.Preimage, nil
	return "", errors.New("TODO")
}

func (gs *LDKService) GetBalance(ctx context.Context) (balance int64, err error) {
	channels := gs.node.ListChannels()

	balance = 0
	for _, channel := range channels {
		balance += int64(channel.BalanceMsat)
	}

	return balance, nil
}

func (gs *LDKService) MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64) (transaction *Nip47Transaction, err error) {
	if expiry == 0 {
		expiry = 86400
	}
	invoice, err := gs.node.ReceivePayment(uint64(amount),
		description,
		uint32(expiry))

	if err != nil {
		log.Printf("MakeInvoice failed: %v", err)
		return nil, err
	}

	// TODO: add missing fields
	transaction = &Nip47Transaction{
		Type:    "incoming",
		Invoice: invoice,
		//PaymentHash: invoice.PaymentHash,
		Amount: amount,
		//CreatedAt:   time.Now().Unix(),
		//ExpiresAt:   &invoice.ExpiresAt,
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
		transactions = append(transactions, *ldkPaymentToTransaction(&payment))
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
	err := gs.node.Connect(connectPeerRequest.Pubkey, connectPeerRequest.Address+":"+strconv.Itoa(connectPeerRequest.Port), true)
	if err != nil {
		gs.svc.Logger.Errorf("ConnectPeer failed: %v", err)
		return err
	}

	return nil
}

func (gs *LDKService) OpenChannel(ctx context.Context, openChannelRequest *lnclient.OpenChannelRequest) (*lnclient.OpenChannelResponse, error) {
	peers := gs.node.ListPeers()
	for _, peer := range peers {
		if peer.NodeId == openChannelRequest.Pubkey {
			gs.svc.Logger.Infof("Opening channel with: %v", peer.NodeId)
			err := gs.node.ConnectOpenChannel(peer.NodeId, peer.Address, uint64(openChannelRequest.Amount), nil, nil, true)
			if err != nil {
				gs.svc.Logger.Errorf("OpenChannel failed: %v", err)
				return nil, err
			}
			return nil, nil
		}
	}

	return nil, errors.New("peer not found")
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
	balance, err := gs.node.SpendableOnchainBalanceSats()
	gs.svc.Logger.Infof("SpendableOnchainBalanceSats: %v", balance)
	if err != nil {
		gs.svc.Logger.Errorf("SpendableOnchainBalanceSats failed: %v", err)
		return 0, err
	}
	return int64(balance), nil
}

func ldkPaymentToTransaction(payment *ldk_node.PaymentDetails) *Nip47Transaction {
	transactionType := "incoming"
	if payment.Direction == ldk_node.PaymentDirectionOutbound {
		transactionType = "outgoing"
	}

	preimage := ""
	var settledAt *int64
	if payment.Status == ldk_node.PaymentStatusSucceeded {
		preimage = *payment.Secret
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
