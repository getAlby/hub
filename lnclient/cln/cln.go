package cln

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	clngrpc "github.com/getAlby/hub/lnclient/cln/clngrpc"
	clngrpcHold "github.com/getAlby/hub/lnclient/cln/clngrpc_hold"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/nip47/notifications"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type CLNService struct {
	ctx            context.Context
	client         clngrpc.NodeClient
	clientHold     clngrpcHold.HoldClient
	holdEnabled    bool
	conn           *grpc.ClientConn
	connHold       *grpc.ClientConn
	eventPublisher events.EventPublisher
	pubkey         string
	cancel         context.CancelFunc
}

func NewCLNService(ctx context.Context, eventPublisher events.EventPublisher, address, lightningDir, addressHold string) (lnclient.LNClient, error) {
	logger.Logger.WithFields(logrus.Fields{
		"address":      address,
		"lightningDir": lightningDir,
		"addressHold":  addressHold,
	}).Info("Creating new CLN gRPC service")

	// CLN grpc client
	tlsConfig, err := loadTLSCredentials(lightningDir, "cln")
	if err != nil {
		return nil, fmt.Errorf("failed to load CLN TLS credentials: %w", err)
	}

	creds := credentials.NewTLS(tlsConfig)

	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to CLN gRPC: %w", err)
	}

	client := clngrpc.NewNodeClient(conn)

	ctx, cancel := context.WithCancel(ctx)

	var connHold *grpc.ClientConn

	defer func() {
		if err != nil {
			cancel()
			if conn != nil {
				conn.Close()
			}
			if connHold != nil {
				connHold.Close()
			}
		}
	}()

	svc := &CLNService{
		ctx:            ctx,
		client:         client,
		holdEnabled:    false,
		conn:           conn,
		eventPublisher: eventPublisher,
		cancel:         cancel,
	}

	// Cln hold plugin grpc client
	if addressHold != "" {
		tlsConfigHold, err := loadTLSCredentials(lightningDir, "hold")
		if err != nil {
			return nil, fmt.Errorf("failed to load hold pluginTLS credentials: %w", err)
		}

		credsHold := credentials.NewTLS(tlsConfigHold)

		connHold, err = grpc.NewClient(
			addressHold,
			grpc.WithTransportCredentials(credsHold),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to hold plugin gRPC: %w", err)
		}

		clientHold := clngrpcHold.NewHoldClient(connHold)

		svc.connHold = connHold
		svc.clientHold = clientHold

		logger.Logger.Info("Testing CLN hold plugin gRPC connection")
		_, err = svc.clientHold.List(ctx, &clngrpcHold.ListRequest{Constraint: &clngrpcHold.ListRequest_Pagination_{
			Pagination: &clngrpcHold.ListRequest_Pagination{
				IndexStart: 0,
				Limit:      1,
			},
		}})

		if err != nil {
			logger.Logger.WithError(err).Error("Failed to connect to CLN hold plugin")
			return nil, fmt.Errorf("failed to connect to CLN hold plugin: %w", err)
		}
		svc.holdEnabled = true
		logger.Logger.Info("Successfully connected to CLN hold plugin via gRPC")

		go svc.subscribeOpenHoldInvoices(ctx)
	} else {
		logger.Logger.Info("No hold plugin configured")
	}

	logger.Logger.Info("Testing CLN gRPC connection")
	resp, err := svc.GetInfo(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to connect to CLN")
		return nil, fmt.Errorf("failed to connect to CLN: %w", err)
	}

	svc.pubkey = resp.Pubkey

	logger.Logger.Info("Successfully connected to CLN via gRPC")
	return svc, nil
}
func loadTLSCredentials(lightningDir string, serverName string) (*tls.Config, error) {
	if serverName != "cln" {
		lightningDir = filepath.Join(lightningDir, serverName)
	}
	certPath := filepath.Join(lightningDir, "ca.pem")
	clientCertPath := filepath.Join(lightningDir, "client.pem")
	clientKeyPath := filepath.Join(lightningDir, "client-key.pem")

	serverCA, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read server CA cert: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(serverCA) {
		return nil, fmt.Errorf("failed to add server CA cert to pool")
	}

	clientCert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load client cert/key: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
		ServerName:   serverName, // CLN uses "cln" as default ServerName, hold plugin uses "hold"
		MinVersion:   tls.VersionTLS12,
	}, nil
}

func (c *CLNService) subscribeOpenHoldInvoices(ctx context.Context) {
	holdinvoices := make([]*clngrpcHold.Invoice, 0)

	const (
		maxRetries  = 5
		baseBackoff = 500 * time.Millisecond
		maxBackoff  = 5 * time.Second
		pageSize    = 200
	)

	start := int64(1)

	for {
		var lsr *clngrpcHold.ListResponse
		var err error

		for attempt := 0; attempt <= maxRetries; attempt++ {
			lsr, err = c.clientHold.List(ctx, &clngrpcHold.ListRequest{
				Constraint: &clngrpcHold.ListRequest_Pagination_{
					Pagination: &clngrpcHold.ListRequest_Pagination{
						IndexStart: start,
						Limit:      pageSize,
					},
				},
			})

			if err == nil {
				break
			}

			if attempt == maxRetries {
				logger.Logger.WithError(err).
					WithField("start", start).
					Error("List invoices failed after retries")
				return
			}

			backoff := min(baseBackoff*time.Duration(1<<attempt), maxBackoff)

			jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
			sleep := backoff/2 + jitter

			logger.Logger.WithError(err).
				WithFields(logrus.Fields{
					"attempt": attempt + 1,
					"sleep":   sleep,
				}).
				Warn("List invoices failed, retrying")

			select {
			case <-time.After(sleep):
			case <-ctx.Done():
				logger.Logger.WithError(ctx.Err()).
					Warn("Context cancelled during retry backoff")
				return
			}
		}

		if lsr == nil || len(lsr.Invoices) == 0 {
			break
		}

		holdinvoices = append(holdinvoices, lsr.Invoices...)
		start = lsr.Invoices[len(lsr.Invoices)-1].Id + 1
	}

	for _, invoice := range holdinvoices {
		if invoice.State == clngrpcHold.InvoiceState_UNPAID {
			paymentHashHex := hex.EncodeToString(invoice.PaymentHash)
			logger.Logger.WithFields(logrus.Fields{
				"paymentHash": paymentHashHex,
				"addIndex":    invoice.Id,
			}).Info("Resubscribing to pending hold invoice")

			go c.subscribeSingleInvoice(invoice.PaymentHash)
		}
	}
}

func (c *CLNService) subscribeSingleInvoice(paymentHashBytes []byte) {
	// Use the global context for the lifetime of this subscription, but create a cancellable one for this specific task
	// This allows the goroutine to be potentially cancelled externally if needed, though it primarily exits on invoice state change.
	// We use a background context derived from the global one to avoid cancelling if the original request context finishes.
	ctx, cancel := context.WithCancel(c.ctx)
	defer cancel() // Ensure cancellation happens on exit

	paymentHashHex := hex.EncodeToString(paymentHashBytes)
	log := logger.Logger.WithField("paymentHash", paymentHashHex)

	log.Info("Starting subscribeSingleInvoice goroutine")

	subReq := &clngrpcHold.TrackRequest{
		PaymentHash: paymentHashBytes,
	}

	invoiceStream, err := c.clientHold.Track(ctx, subReq)
	if err != nil {
		log.WithError(err).Error("SubscribeSingleInvoice call failed")
		// Goroutine will exit
		return
	}

	log.Info("Successfully subscribed to single invoice stream")

	defer func() {
		log.Info("Exiting subscribeSingleInvoice goroutine")
		if r := recover(); r != nil {
			log.WithField("panic", r).Errorf("PANIC recovered in single invoice stream processing")
		}
	}()

	for {
		trackResponse, err := invoiceStream.Recv()

		if err != nil {
			log.WithError(err).Error("Failed to receive single invoice update from stream")
			return
		}
		if ctx.Err() != nil {
			log.Info("Context cancelled, exiting single invoice subscription loop")
			return
		}

		log.WithFields(logrus.Fields{
			"rawState": trackResponse.State.String(),
		}).Info("Raw update received from single invoice stream")

		switch trackResponse.State {
		case clngrpcHold.InvoiceState_ACCEPTED:
			log.Info("Hold invoice accepted, publishing internal event")

			tx, err := c.buildHoldInvoiceTransaction(ctx, paymentHashBytes)
			if err != nil {
				logger.Logger.WithError(err).Error("failed to build hold invoice transaction")
				return
			}

			c.eventPublisher.Publish(&events.Event{
				Event:      "nwc_lnclient_hold_invoice_accepted",
				Properties: tx,
			})
		case clngrpcHold.InvoiceState_CANCELLED:
			log.Info("Hold invoice canceled, ending subscription")
			return // Invoice reached final state, exit goroutine
		case clngrpcHold.InvoiceState_PAID:
			return // Invoice reached final state, exit goroutine
		case clngrpcHold.InvoiceState_UNPAID:
			// Continue loop
		}
	}
}

func (c *CLNService) buildHoldInvoiceTransaction(ctx context.Context, paymentHash []byte) (*lnclient.Transaction, error) {
	invoice, err := c.fetchHoldInvoice(ctx, paymentHash)
	if err != nil {
		return nil, err
	}

	decodedInvoice, err := c.client.Decode(ctx, &clngrpc.DecodeRequest{
		String_: invoice.Invoice,
	})
	if err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	return clnHoldInvoiceToTransaction(invoice, decodedInvoice)
}

func (c *CLNService) fetchHoldInvoice(ctx context.Context, paymentHash []byte) (*clngrpcHold.Invoice, error) {
	if !c.holdEnabled {
		return nil, errors.New("hold client not configured")
	}

	resp, err := c.clientHold.List(ctx, &clngrpcHold.ListRequest{
		Constraint: &clngrpcHold.ListRequest_PaymentHash{
			PaymentHash: paymentHash,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("hold list failed: %w", err)
	}

	if len(resp.Invoices) != 1 {
		return nil, fmt.Errorf("expected 1 invoice, got %d", len(resp.Invoices))
	}

	return resp.Invoices[0], nil
}

func clnHoldInvoiceToTransaction(invoice *clngrpcHold.Invoice, decodedInvoice *clngrpc.DecodeResponse) (*lnclient.Transaction, error) {
	description := ""
	if decodedInvoice.Description != nil {
		description = *decodedInvoice.Description
	}

	descriptionHash := ""
	if len(decodedInvoice.DescriptionHash) > 0 {
		descriptionHash = hex.EncodeToString(decodedInvoice.DescriptionHash)
	}

	amountMsat := int64(0)
	if decodedInvoice.AmountMsat != nil {
		amountMsat = int64(decodedInvoice.AmountMsat.Msat)
	}

	var minExpiry *uint32
	for _, htlc := range invoice.Htlcs {
		if htlc.CltvExpiry == nil {
			continue
		}

		exp := uint32(*htlc.CltvExpiry)

		if minExpiry == nil || exp < *minExpiry {
			minExpiry = &exp
		}
	}

	tx := &lnclient.Transaction{
		Type:            "incoming",
		Invoice:         invoice.Invoice,
		Description:     description,
		DescriptionHash: descriptionHash,
		Preimage:        hex.EncodeToString(invoice.Preimage),
		PaymentHash:     hex.EncodeToString(invoice.PaymentHash),
		Amount:          amountMsat,
		CreatedAt:       int64(invoice.CreatedAt),
		SettledAt:       nil,
		FeesPaid:        0,
		Metadata:        lnclient.Metadata{},
		SettleDeadline:  minExpiry,
	}

	if decodedInvoice.Expiry != nil {
		expiresAt := int64(invoice.CreatedAt + *decodedInvoice.Expiry)
		tx.ExpiresAt = &expiresAt
	}

	return tx, nil
}

func (c *CLNService) CloseChannel(ctx context.Context, closeChannelRequest *lnclient.CloseChannelRequest) (*lnclient.CloseChannelResponse, error) {
	logger.Logger.WithFields(logrus.Fields{
		"closeChannelRequest": closeChannelRequest,
	}).Debug("Closing Channel")

	req := &clngrpc.CloseRequest{
		Id: closeChannelRequest.ChannelId,
	}

	if closeChannelRequest.Force {
		// There is no force option in CLN, only a Unilateraltimeout after which the channel will be force closed
		// 0 means waiting forever so we choose 1 second
		timeout := uint32(1)
		req.Unilateraltimeout = &timeout
	}

	_, err := c.client.Close(ctx, req)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to close channel")
		return nil, fmt.Errorf("close failed: %w", err)
	}

	return &lnclient.CloseChannelResponse{}, err
}

func (c *CLNService) ConnectPeer(ctx context.Context, connectPeerRequest *lnclient.ConnectPeerRequest) error {
	logger.Logger.WithFields(logrus.Fields{
		"connectPeerRequest": connectPeerRequest,
	}).Debug("Connecting to Peer")

	port := uint32(connectPeerRequest.Port)
	req := &clngrpc.ConnectRequest{
		Id:   connectPeerRequest.Pubkey,
		Host: &connectPeerRequest.Address,
		Port: &port,
	}

	_, err := c.client.ConnectPeer(ctx, req)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to connect peer")
		return err
	}

	return nil
}

func (c *CLNService) DisconnectPeer(ctx context.Context, peerId string) error {
	logger.Logger.WithFields(logrus.Fields{
		"peerId": peerId,
	}).Debug("Disconnecting Peer")

	pubkey, err := hex.DecodeString(peerId)
	if err != nil {
		return err
	}
	req := &clngrpc.DisconnectRequest{
		Id: pubkey,
	}

	_, err = c.client.Disconnect(ctx, req)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to disconnect peer")
		return err
	}

	return nil
}

func (c *CLNService) GetCustomNodeCommandDefinitions() []lnclient.CustomNodeCommandDef {
	return nil
}

func (c *CLNService) ExecuteCustomNodeCommand(ctx context.Context, command *lnclient.CustomNodeCommandRequest) (*lnclient.CustomNodeCommandResponse, error) {
	return nil, nil
}

func (c *CLNService) GetBalances(ctx context.Context, includeInactiveChannels bool) (*lnclient.BalancesResponse, error) {
	logger.Logger.WithFields(logrus.Fields{
		"includeInactiveChannels": includeInactiveChannels,
	}).Debug("Get all Balances")

	onchainBalance, err := c.GetOnchainBalance(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.ListPeerChannels(ctx, &clngrpc.ListpeerchannelsRequest{})
	if err != nil {
		return nil, fmt.Errorf("listpeerchannels failed: %w", err)
	}

	lightning := lnclient.LightningBalanceResponse{}

	for _, ch := range resp.Channels {
		if ch == nil {
			continue
		}

		// Never include closing or closed channels
		if ch.State != clngrpc.ChannelState_ChanneldNormal {
			continue
		}

		// This isn't perfect to determine if a channel is active
		active := ch.PeerConnected
		include := active || includeInactiveChannels
		if !include {
			continue
		}

		if ch.SpendableMsat != nil {
			spendable := int64(ch.SpendableMsat.Msat)
			lightning.TotalSpendable += spendable

			if spendable > lightning.NextMaxSpendable {
				lightning.NextMaxSpendable = spendable
			}
		}

		if ch.ReceivableMsat != nil {
			receivable := int64(ch.ReceivableMsat.Msat)
			lightning.TotalReceivable += receivable

			if receivable > lightning.NextMaxReceivable {
				lightning.NextMaxReceivable = receivable
			}
		}
	}

	lightning.NextMaxSpendableMPP = lightning.TotalSpendable
	lightning.NextMaxReceivableMPP = lightning.TotalReceivable

	return &lnclient.BalancesResponse{
		Onchain:   *onchainBalance,
		Lightning: lightning,
	}, nil
}

func (c *CLNService) GetInfo(ctx context.Context) (*lnclient.NodeInfo, error) {
	resp, err := c.client.Getinfo(ctx, &clngrpc.GetinfoRequest{})
	if err != nil {
		return nil, fmt.Errorf("getinfo failed: %w", err)
	}

	return &lnclient.NodeInfo{
		Alias:       resp.GetAlias(),
		Color:       hex.EncodeToString(resp.Color),
		Pubkey:      hex.EncodeToString(resp.Id),
		Network:     resp.Network,
		BlockHeight: resp.Blockheight,
		BlockHash:   "", // Not directly available
	}, nil
}

func (c *CLNService) GetLogOutput(ctx context.Context, maxLen int) ([]byte, error) {
	return []byte{}, nil
}

func (c *CLNService) GetNetworkGraph(ctx context.Context, nodeIds []string) (lnclient.NetworkGraphResponse, error) {
	logger.Logger.WithFields(logrus.Fields{
		"nodeIds": nodeIds,
	}).Debug("Get Network Graph")

	listnodes := make([]*clngrpc.ListnodesNodes, 0)
	listchannels := make([]*clngrpc.ListchannelsChannels, 0)

	for _, nodeId := range nodeIds {
		nodeIdBytes, err := hex.DecodeString(nodeId)
		if err != nil {
			logger.Logger.WithError(err).Error("failed to decode nodeId string")
			return nil, fmt.Errorf("failed to decode nodeId string: %w", err)
		}

		listnode, err := c.client.ListNodes(ctx, &clngrpc.ListnodesRequest{Id: nodeIdBytes})
		if err != nil {
			logger.Logger.WithError(err).Error("listnodes failed")
			return nil, err
		}
		listnodes = append(listnodes, listnode.Nodes...)

		listchannel, err := c.client.ListChannels(ctx, &clngrpc.ListchannelsRequest{Source: nodeIdBytes})
		if err != nil {
			logger.Logger.WithError(err).Error("listchannels failed")
			return nil, err
		}
		listchannels = append(listchannels, listchannel.Channels...)

		listchannel, err = c.client.ListChannels(ctx, &clngrpc.ListchannelsRequest{Destination: nodeIdBytes})
		if err != nil {
			logger.Logger.WithError(err).Error("listchannels failed")
			return nil, err
		}
		listchannels = append(listchannels, listchannel.Channels...)
	}

	type NetworkNode struct {
		NodeId    string   `json:"nodeId"`
		Alias     string   `json:"alias"`
		Color     string   `json:"color"`
		Addresses []string `json:"addresses"`
		Features  string   `json:"features"`
	}

	type NodeInfoWithId struct {
		Node   *NetworkNode `json:"node"`
		NodeId string       `json:"nodeId"`
	}

	type NetworkChannel struct {
		Scid     string `json:"scid"`
		Node1    string `json:"node1"`
		Node2    string `json:"node2"`
		Capacity uint64 `json:"capacity"`
		Active   bool   `json:"active"`
		Public   bool   `json:"public"`
	}

	nodes := []NodeInfoWithId{}
	channels := []*NetworkChannel{}

	for _, node := range listnodes {
		nodeIdStr := hex.EncodeToString(node.Nodeid)
		addrs := []string{}
		for _, a := range node.Addresses {
			addrs = append(addrs, fmt.Sprintf("%s:%d", a.GetAddress(), a.GetPort()))
		}
		networkNode := NetworkNode{
			NodeId:    nodeIdStr,
			Alias:     node.GetAlias(),
			Color:     hex.EncodeToString(node.Color),
			Addresses: addrs,
			Features:  hex.EncodeToString(node.Features),
		}
		nodes = append(nodes, NodeInfoWithId{
			Node:   &networkNode,
			NodeId: nodeIdStr,
		})

	}

	seen := make(map[string]struct{})
	for _, edge := range listchannels {
		key := fmt.Sprintf("%s:%d", edge.ShortChannelId, edge.Direction)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		channel := NetworkChannel{
			Scid:     edge.ShortChannelId,
			Node1:    hex.EncodeToString(edge.Source),
			Node2:    hex.EncodeToString(edge.Destination),
			Capacity: sat(edge.AmountMsat),
			Active:   edge.Active,
			Public:   edge.Public,
		}
		channels = append(channels, &channel)

	}

	networkGraph := map[string]interface{}{
		"nodes":    nodes,
		"channels": channels,
	}
	return networkGraph, nil
}

func (c *CLNService) GetNewOnchainAddress(ctx context.Context) (string, error) {
	resp, err := c.client.NewAddr(ctx, &clngrpc.NewaddrRequest{})
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to generate onchain address")
		return "", err
	}

	if resp.Bech32 != nil {
		return *resp.Bech32, nil
	}

	if resp.P2Tr != nil {
		return *resp.P2Tr, nil
	}

	logger.Logger.WithField("resp", resp).Error("No known onchain address type returned")
	return "", fmt.Errorf("unknown default onchain address type")
}

func (c *CLNService) GetNodeConnectionInfo(ctx context.Context) (*lnclient.NodeConnectionInfo, error) {
	resp, err := c.client.Getinfo(ctx, &clngrpc.GetinfoRequest{})
	if err != nil {
		return nil, fmt.Errorf("getinfo failed: %w", err)
	}

	var (
		ipv4  *clngrpc.GetinfoAddress
		ipv6  *clngrpc.GetinfoAddress
		torv3 *clngrpc.GetinfoAddress
	)

	for _, addr := range resp.Address {
		if addr == nil {
			continue
		}

		switch addr.ItemType {
		case clngrpc.GetinfoAddress_IPV4:
			if ipv4 == nil {
				ipv4 = addr
			}
		case clngrpc.GetinfoAddress_IPV6:
			if ipv6 == nil {
				ipv6 = addr
			}
		case clngrpc.GetinfoAddress_TORV3:
			if torv3 == nil {
				torv3 = addr
			}
		}
	}

	var selected *clngrpc.GetinfoAddress
	switch {
	case ipv4 != nil:
		selected = ipv4
	case ipv6 != nil:
		selected = ipv6
	case torv3 != nil:
		selected = torv3
	default:
		addr := "not announced"
		selected = &clngrpc.GetinfoAddress{
			Address: &addr,
			Port:    0,
		}
	}

	return &lnclient.NodeConnectionInfo{
		Pubkey:  hex.EncodeToString(resp.Id),
		Address: selected.GetAddress(),
		Port:    int(selected.Port),
	}, nil
}

func (c *CLNService) GetNodeStatus(ctx context.Context) (nodeStatus *lnclient.NodeStatus, err error) {
	resp, err := c.client.Getinfo(ctx, &clngrpc.GetinfoRequest{})
	if err != nil {
		return nil, fmt.Errorf("getinfo failed: %w", err)
	}

	ready := false
	if resp != nil {
		if resp.WarningBitcoindSync == nil && resp.WarningLightningdSync == nil {
			ready = true
		}
	}

	return &lnclient.NodeStatus{
		IsReady:            ready,
		InternalNodeStatus: 0,
	}, nil
}

func (c *CLNService) GetOnchainBalance(ctx context.Context) (*lnclient.OnchainBalanceResponse, error) {
	lf, err := c.client.ListFunds(ctx, &clngrpc.ListfundsRequest{})
	if err != nil {
		return nil, fmt.Errorf("listfunds failed: %w", err)
	}

	lpc, err := c.client.ListPeerChannels(ctx, &clngrpc.ListpeerchannelsRequest{})
	if err != nil {
		return nil, fmt.Errorf("listpeerchannels failed: %w", err)
	}

	chByID := make(map[string]*clngrpc.ListpeerchannelsChannels)
	for _, ch := range lpc.Channels {
		if ch == nil || len(ch.ChannelId) == 0 {
			continue
		}
		chByID[hex.EncodeToString(ch.ChannelId)] = ch
	}

	balances := &lnclient.OnchainBalanceResponse{
		PendingBalancesDetails:      []lnclient.PendingBalanceDetails{},
		PendingSweepBalancesDetails: []lnclient.PendingBalanceDetails{},
	}

	var reservedSats int64

	for _, utxo := range lf.Outputs {
		if utxo == nil || utxo.AmountMsat == nil {
			continue
		}

		amt := satInt64(utxo.AmountMsat)
		balances.Total += amt

		if utxo.Reserved {
			balances.Reserved += amt
			reservedSats += amt
		}

		switch utxo.Status {
		case clngrpc.ListfundsOutputs_CONFIRMED:
			if !utxo.Reserved {
				balances.Spendable += amt
			}

		case clngrpc.ListfundsOutputs_UNCONFIRMED:
			balances.PendingSweepBalancesDetails = append(
				balances.PendingSweepBalancesDetails,
				lnclient.PendingBalanceDetails{
					Amount:        uint64(amt),
					FundingTxId:   hex.EncodeToString(utxo.Txid),
					FundingTxVout: utxo.Output,
				},
			)
		}
	}

	for _, ch := range lf.Channels {
		if ch == nil || ch.OurAmountMsat == nil || !isClosingState(ch.State) {
			continue
		}

		amt := sat(ch.OurAmountMsat)
		balances.PendingBalancesFromChannelClosures += amt
		chanIdStr := hex.EncodeToString(ch.ChannelId)

		detail := lnclient.PendingBalanceDetails{
			ChannelId: chanIdStr,
			NodeId:    hex.EncodeToString(ch.PeerId),
			Amount:    amt,
		}

		if pc, ok := chByID[chanIdStr]; ok {
			if len(pc.FundingTxid) > 0 {
				detail.FundingTxId = hex.EncodeToString(pc.FundingTxid)
			}
			if pc.FundingOutnum != nil {
				detail.FundingTxVout = *pc.FundingOutnum
			}
		}

		balances.PendingBalancesDetails = append(
			balances.PendingBalancesDetails,
			detail,
		)
	}

	balances.InternalBalances = map[string]int64{
		"reserved": reservedSats,
	}

	return balances, nil
}

func isClosingState(state clngrpc.ChannelState) bool {
	switch state {
	case clngrpc.ChannelState_ChanneldShuttingDown,
		clngrpc.ChannelState_ClosingdSigexchange,
		clngrpc.ChannelState_ClosingdComplete,
		clngrpc.ChannelState_AwaitingUnilateral,
		clngrpc.ChannelState_FundingSpendSeen:
		return true
	default:
		return false
	}
}

func isOpeningState(state clngrpc.ChannelState) bool {
	switch state {
	case clngrpc.ChannelState_ChanneldAwaitingLockin,
		clngrpc.ChannelState_DualopendAwaitingLockin,
		clngrpc.ChannelState_DualopendOpenCommittReady,
		clngrpc.ChannelState_DualopendOpenCommitted,
		clngrpc.ChannelState_DualopendOpenInit,
		clngrpc.ChannelState_Openingd:
		return true
	default:
		return false
	}
}

func isConfirmedState(state clngrpc.ChannelState) bool {
	switch state {
	case clngrpc.ChannelState_AwaitingUnilateral,
		clngrpc.ChannelState_ChanneldAwaitingSplice,
		clngrpc.ChannelState_ChanneldNormal,
		clngrpc.ChannelState_ChanneldShuttingDown,
		clngrpc.ChannelState_ClosingdComplete,
		clngrpc.ChannelState_ClosingdSigexchange,
		clngrpc.ChannelState_FundingSpendSeen,
		clngrpc.ChannelState_Onchain:
		return true
	default:
		return false
	}
}

func msatInt64(a *clngrpc.Amount) int64 {
	if a == nil {
		return 0
	}
	return int64(a.Msat)
}

func satInt64(a *clngrpc.Amount) int64 {
	if a == nil {
		return 0
	}
	return int64(a.Msat / 1000)
}

func sat(a *clngrpc.Amount) uint64 {
	if a == nil {
		return 0
	}
	return a.Msat / 1000
}

func localFeeBaseMsat(ch *clngrpc.ListpeerchannelsChannels) uint32 {
	if ch == nil {
		return 0
	}
	u := ch.Updates
	if u == nil {
		return 0
	}
	l := u.Local
	if l == nil {
		return 0
	}
	f := l.FeeBaseMsat
	if f == nil {
		return 0
	}
	return uint32(f.Msat)
}

func localFeePPM(ch *clngrpc.ListpeerchannelsChannels) uint32 {
	if ch == nil {
		return 0
	}
	u := ch.Updates
	if u == nil {
		return 0
	}
	l := u.Local
	if l == nil {
		return 0
	}
	f := l.FeeProportionalMillionths
	return f
}

func (c *CLNService) GetPubkey() string {
	return c.pubkey
}

func (c *CLNService) GetStorageDir() (string, error) {
	return "", nil
}

func (c *CLNService) GetSupportedNIP47Methods() []string {
	logger.Logger.Info("GetSupportedNIP47Methods")
	methods := []string{
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
	}

	if c.holdEnabled {
		methods = append(methods,
			models.MAKE_HOLD_INVOICE_METHOD,
			models.SETTLE_HOLD_INVOICE_METHOD,
			models.CANCEL_HOLD_INVOICE_METHOD,
		)
	}

	return methods
}

func (c *CLNService) GetSupportedNIP47NotificationTypes() []string {
	result := make([]string, 0)

	if c.holdEnabled {
		result = append(result,
			notifications.HOLD_INVOICE_ACCEPTED_NOTIFICATION)
	}
	return result
}

func (c *CLNService) ListChannels(ctx context.Context) (channels []lnclient.Channel, err error) {
	resp, err := c.client.ListPeerChannels(ctx, &clngrpc.ListpeerchannelsRequest{})
	if err != nil {
		logger.Logger.WithError(err).Error("listpeerchannels failed")
		return nil, err
	}

	infoResp, infoErr := c.client.Getinfo(ctx, &clngrpc.GetinfoRequest{})
	if infoErr != nil {
		logger.Logger.WithError(infoErr).Error("getinfo failed")
		return nil, infoErr
	}

	blockheight := infoResp.Blockheight

	reChanHeight := regexp.MustCompile(`(\d+)x.*`)

	for _, channel := range resp.Channels {
		if channel == nil {
			continue
		}

		var errorStrings []string

		// We could check the funding-confirms config but it's only for remote openers
		// In reality channels often confirm with 3 or 6 confirmations
		ConfirmationsRequired := uint32(6)
		if isConfirmedState(channel.State) {
			ConfirmationsRequired = 0
		} else if isOpeningState(channel.State) {
			confRequired, errStr := confirmationsRequiredFromStatus(channel.Status)
			if errStr != nil {
				logger.Logger.Error(*errStr)
				errorStrings = append(errorStrings, *errStr)
			} else {
				ConfirmationsRequired = confRequired
			}

		} else {
			errStr := fmt.Sprintf("unexpected clngrpc.ChannelState: %#v", channel.State)
			logger.Logger.Error(errStr)
			errorStrings = append(errorStrings, errStr)
		}

		var chanBlock *uint32
		if channel.ShortChannelId != nil {
			match := reChanHeight.FindStringSubmatch(*channel.ShortChannelId)
			if len(match) > 1 {
				num, err := strconv.Atoi(match[1])
				if err != nil {
					errStr := fmt.Sprintf("Error converting number: %v", err)
					logger.Logger.Error(errStr)
					errorStrings = append(errorStrings, errStr)
				}
				num32 := uint32(num)
				chanBlock = &num32
			}
		}

		var Confirmations uint32
		if chanBlock != nil {
			if blockheight >= *chanBlock {
				Confirmations = (blockheight - *chanBlock) + 1
			} else {
				Confirmations = 0
			}
		} else {
			Confirmations = 0
		}

		isActive := channel.State == clngrpc.ChannelState_ChanneldNormal && channel.PeerConnected

		var Error *string
		if len(errorStrings) > 0 {
			combined := strings.Join(errorStrings, "; ")
			Error = &combined
		}

		LocalBalance := msatInt64(channel.ToUsMsat)
		TotalBalance := msatInt64(channel.TotalMsat)
		RemoteBalance := int64(0)
		if TotalBalance >= LocalBalance {
			RemoteBalance = TotalBalance - LocalBalance
		}

		channels = append(channels, lnclient.Channel{
			LocalBalance:                             LocalBalance,
			LocalSpendableBalance:                    msatInt64(channel.SpendableMsat),
			RemoteBalance:                            RemoteBalance,
			Id:                                       hex.EncodeToString(channel.ChannelId),
			RemotePubkey:                             hex.EncodeToString(channel.PeerId),
			FundingTxId:                              hex.EncodeToString(channel.FundingTxid),
			FundingTxVout:                            channel.GetFundingOutnum(),
			Active:                                   isActive,
			Public:                                   !channel.GetPrivate(),
			InternalChannel:                          channel,
			Confirmations:                            &Confirmations,
			ConfirmationsRequired:                    &ConfirmationsRequired,
			ForwardingFeeBaseMsat:                    localFeeBaseMsat(channel),
			ForwardingFeeProportionalMillionths:      localFeePPM(channel),
			UnspendablePunishmentReserve:             sat(channel.OurReserveMsat),
			CounterpartyUnspendablePunishmentReserve: sat(channel.TheirReserveMsat),
			Error:                                    Error,
			IsOutbound:                               channel.GetOpener() == clngrpc.ChannelSide_LOCAL,
		})
	}
	return channels, nil
}

func confirmationsRequiredFromStatus(status []string) (uint32, *string) {
	reStatus := regexp.MustCompile(`.*Funding needs (\d+) more confirmations to be ready.*`)

	for _, status := range status {
		match := reStatus.FindStringSubmatch(status)
		if len(match) > 1 {
			num, err := strconv.Atoi(match[1])
			if err != nil {
				errStr := fmt.Sprintf("Error converting number of confirmations required: %v", err)
				return 0, &errStr
			}
			return uint32(num), nil
		}
	}

	errNotFound := "Could not find status indicating number of confirmations required"
	return 0, &errNotFound
}

func (c *CLNService) ListOnchainTransactions(ctx context.Context) ([]lnclient.OnchainTransaction, error) {
	account := "wallet"
	req := &clngrpc.BkprlistaccounteventsRequest{
		Account: &account,
	}
	bkpr, err := c.client.BkprListAccountEvents(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("bkprlistaccountevents failed: %w", err)
	}

	infoResp, infoErr := c.client.Getinfo(ctx, &clngrpc.GetinfoRequest{})
	if infoErr != nil {
		logger.Logger.WithError(infoErr).Error("getinfo failed")
		return nil, infoErr
	}

	blockheight := infoResp.Blockheight

	transactions := make([]lnclient.OnchainTransaction, 0)

	for _, event := range bkpr.Events {
		if event.ItemType != clngrpc.BkprlistaccounteventsEvents_CHAIN {
			continue
		}
		transactionType := "incoming"
		AmountSat := sat(event.CreditMsat)
		debitSat := sat(event.DebitMsat)
		if debitSat > 0 {
			transactionType = "outgoing"
			AmountSat = debitSat
		}

		numConfirmations := uint32(0)
		if event.Blockheight != nil && blockheight >= *event.Blockheight {
			numConfirmations = blockheight - *event.Blockheight
		}

		TxIdHex := hex.EncodeToString(event.Txid)

		transactions = append(transactions, lnclient.OnchainTransaction{
			AmountSat:        AmountSat,
			CreatedAt:        uint64(event.Timestamp),
			State:            "confirmed",
			Type:             transactionType,
			NumConfirmations: numConfirmations,
			TxId:             TxIdHex,
		})
	}

	slices.Reverse(transactions)

	sort.SliceStable(transactions, func(i, j int) bool {
		return transactions[i].CreatedAt > transactions[j].CreatedAt
	})

	return transactions, nil
}

func (c *CLNService) ListPeers(ctx context.Context) ([]lnclient.PeerDetails, error) {
	resp, err := c.client.ListPeerChannels(ctx, &clngrpc.ListpeerchannelsRequest{})
	if err != nil {
		return nil, fmt.Errorf("listpeerchannels failed: %w", err)
	}

	peers := make([]lnclient.PeerDetails, 0, len(resp.Channels))
	for _, peer := range resp.Channels {
		if peer == nil {
			continue
		}

		req_node := &clngrpc.ListnodesRequest{Id: peer.PeerId}

		resp_node, err := c.client.ListNodes(ctx, req_node)
		if err != nil {
			return nil, fmt.Errorf("listnodes failed: %w", err)
		}

		if len(resp_node.Nodes) == 0 {
			addr := "not gossip yet"
			peers = append(peers, lnclient.PeerDetails{
				NodeId:      hex.EncodeToString(peer.PeerId),
				Address:     addr,
				IsPersisted: true,
				IsConnected: peer.PeerConnected,
			})
			continue
		} else {
			var (
				ipv4  *clngrpc.ListnodesNodesAddresses
				ipv6  *clngrpc.ListnodesNodesAddresses
				torv3 *clngrpc.ListnodesNodesAddresses
			)

			for _, addr := range resp_node.Nodes[0].Addresses {
				if addr == nil {
					continue
				}

				switch addr.ItemType {
				case clngrpc.ListnodesNodesAddresses_IPV4:
					if ipv4 == nil {
						ipv4 = addr
					}
				case clngrpc.ListnodesNodesAddresses_IPV6:
					if ipv6 == nil {
						ipv6 = addr
					}
				case clngrpc.ListnodesNodesAddresses_TORV3:
					if torv3 == nil {
						torv3 = addr
					}
				}
			}

			var selected *clngrpc.ListnodesNodesAddresses
			switch {
			case ipv4 != nil:
				selected = ipv4
			case ipv6 != nil:
				selected = ipv6
			case torv3 != nil:
				selected = torv3
			default:
				addr := "not announced"
				selected = &clngrpc.ListnodesNodesAddresses{
					Address: &addr,
					Port:    0,
				}
			}

			peers = append(peers, lnclient.PeerDetails{
				NodeId:      hex.EncodeToString(peer.PeerId),
				Address:     *selected.Address,
				IsPersisted: true,
				IsConnected: peer.PeerConnected,
			})
		}
	}

	return peers, nil

}

func (c *CLNService) LookupInvoice(ctx context.Context, paymentHash string) (transaction *lnclient.Transaction, err error) {
	logger.Logger.WithFields(logrus.Fields{
		"paymentHash": paymentHash,
	}).Debug("Lookup Invoice")

	paymentHashBytes, err := hex.DecodeString(paymentHash)
	if err != nil {
		logger.Logger.WithError(err).Error("failed to decode payment hash")
		return nil, fmt.Errorf("failed to decode payment hash: %w", err)
	}
	req := &clngrpc.ListinvoicesRequest{PaymentHash: paymentHashBytes}

	resp, err := c.client.ListInvoices(ctx, req)
	if err != nil {
		logger.Logger.WithError(err).Error("listinvoices failed")
		return nil, fmt.Errorf("listinvoices failed: %w", err)
	}
	if len(resp.Invoices) == 0 {
		return nil, fmt.Errorf("invoice not found")
	}

	transaction, err = c.clnInvoiceToTransaction(ctx, resp.Invoices[0])
	if err != nil {
		logger.Logger.WithError(err).Error("failed to convert invoice to transaction")
		return nil, fmt.Errorf("failed to convert invoice to transaction: %w", err)
	}

	return transaction, nil
}

func (c *CLNService) clnInvoiceToTransaction(ctx context.Context, invoice *clngrpc.ListinvoicesInvoices) (*lnclient.Transaction, error) {
	var invoiceStr string
	if invoice.Bolt11 != nil {
		invoiceStr = *invoice.Bolt11
	} else if invoice.Bolt12 != nil {
		invoiceStr = *invoice.Bolt12
	} else {
		invoiceStr = ""
	}

	var amount int64
	if invoice.Status == clngrpc.ListinvoicesInvoices_PAID && invoice.AmountReceivedMsat != nil {
		amount = int64(invoice.AmountReceivedMsat.Msat)
	} else if invoice.AmountMsat != nil {
		amount = int64(invoice.AmountMsat.Msat)
	} else {
		amount = 0
	}

	expires_at := int64(invoice.ExpiresAt)

	var paid_at *int64
	if invoice.Status == clngrpc.ListinvoicesInvoices_PAID {
		if invoice.PaidAt == nil {
			return nil, fmt.Errorf("paid_at missing from paid invoice")
		}
		paid_at_int64 := int64(*invoice.PaidAt)
		paid_at = &paid_at_int64
	}

	decoded_invoice, err := c.client.Decode(ctx, &clngrpc.DecodeRequest{String_: invoiceStr})
	if err != nil {
		logger.Logger.WithError(err).Error("decode failed")
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	var created_at int64
	switch decoded_invoice.ItemType {
	case clngrpc.DecodeResponse_BOLT12_INVOICE:
		if decoded_invoice.InvoiceCreatedAt == nil {
			return nil, fmt.Errorf("invoice_created_at missing from bolt12 invoice")
		}
		created_at = int64(*decoded_invoice.InvoiceCreatedAt)
	case clngrpc.DecodeResponse_BOLT11_INVOICE:
		if decoded_invoice.CreatedAt == nil {
			return nil, fmt.Errorf("created_at missing from bolt11 invoice")
		}
		created_at = int64(*decoded_invoice.CreatedAt)

	default:
		return nil, fmt.Errorf("created_at missing from invoice")
	}

	var description string
	if invoice.Description != nil {
		description = *invoice.Description
	} else {
		description = ""
	}

	transaction := &lnclient.Transaction{
		Type:            "incoming",
		Invoice:         invoiceStr,
		Description:     description,
		DescriptionHash: "",
		Preimage:        hex.EncodeToString(invoice.PaymentPreimage),
		PaymentHash:     hex.EncodeToString(invoice.PaymentHash),
		Amount:          amount,
		FeesPaid:        0,
		CreatedAt:       created_at,
		ExpiresAt:       &expires_at,
		SettledAt:       paid_at,
		Metadata:        lnclient.Metadata{},
		SettleDeadline:  nil,
	}
	return transaction, nil
}

func (c *CLNService) MakeHoldInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64, paymentHash string) (transaction *lnclient.Transaction, err error) {
	if !c.holdEnabled {
		return nil, errors.New("hold plugin not configured")
	}

	logger.Logger.WithFields(logrus.Fields{
		"amount":           amount,
		"description":      description,
		"description_hash": descriptionHash,
		"expiry":           expiry,
		"payment_hash":     paymentHash,
	}).Debug("Make Hold Invoice")

	paymentHashBytes, err := hex.DecodeString(paymentHash)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"paymentHash": paymentHash,
		}).WithError(err).Error("Invalid payment hash")
		return nil, fmt.Errorf("Invalid payment hash: %v", err)
	}

	if expiry == 0 {
		expiry = lnclient.DEFAULT_INVOICE_EXPIRY
	}
	expiryUint64 := uint64(expiry)

	req := &clngrpcHold.InvoiceRequest{
		PaymentHash: paymentHashBytes,
		AmountMsat:  uint64(amount),
		Expiry:      &expiryUint64,
	}

	if descriptionHash != "" {
		descriptionHashBytes, err := hex.DecodeString(descriptionHash)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"descriptionHash": descriptionHash,
			}).WithError(err).Error("Invalid description hash")
			return nil, fmt.Errorf("Invalid description hash: %v", err)
		}
		req.Description = &clngrpcHold.InvoiceRequest_Hash{
			Hash: descriptionHashBytes,
		}
	} else {
		req.Description = &clngrpcHold.InvoiceRequest_Memo{
			Memo: description,
		}
	}

	resp, err := c.clientHold.Invoice(ctx, req)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"paymentHash": paymentHash,
		}).WithError(err).Error("Failed to make hold invoice")
		return nil, fmt.Errorf("Failed to make hold invoice: %v", err)
	}

	expiresAt := time.Now().Unix() + expiry

	go c.subscribeSingleInvoice(paymentHashBytes)
	logger.Logger.WithField("paymentHash", paymentHash).Info("Launched single invoice subscription goroutine")

	transaction = &lnclient.Transaction{
		Type:            "incoming",
		Invoice:         resp.Bolt11,
		Description:     description,
		DescriptionHash: descriptionHash,
		Preimage:        "",
		PaymentHash:     paymentHash,
		Amount:          amount,
		FeesPaid:        0,
		CreatedAt:       time.Now().Unix(),
		ExpiresAt:       &expiresAt,
		SettledAt:       nil,
		Metadata:        lnclient.Metadata{},
		SettleDeadline:  nil,
	}
	return transaction, nil
}

func (c *CLNService) SettleHoldInvoice(ctx context.Context, preimage string) (err error) {
	if !c.holdEnabled {
		return errors.New("hold plugin not configured")
	}

	logger.Logger.WithFields(logrus.Fields{
		"preimage": preimage,
	}).Debug("Settle Hold Invoice")

	preimageBytes, err := hex.DecodeString(preimage)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"preimage": preimage,
		}).WithError(err).Error("Invalid preimage")
		return fmt.Errorf("Invalid preimage: %v", err)
	}

	_, err = c.clientHold.Settle(ctx, &clngrpcHold.SettleRequest{
		PaymentPreimage: preimageBytes,
	})
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"preimage": preimage,
		}).WithError(err).Error("Failed to settle hold invoice")
		return fmt.Errorf("Failed to settle hold invoice: %v", err)
	}

	return nil
}

func (c *CLNService) CancelHoldInvoice(ctx context.Context, paymentHash string) (err error) {
	if !c.holdEnabled {
		return errors.New("hold plugin not configured")
	}

	logger.Logger.WithFields(logrus.Fields{
		"paymentHash": paymentHash,
	}).Debug("Cancel Hold Invoice")

	paymentHashBytes, err := hex.DecodeString(paymentHash)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"paymentHash": paymentHash,
		}).WithError(err).Error("Invalid paymentHash")
		return fmt.Errorf("Invalid paymentHash: %v", err)
	}

	_, err = c.clientHold.Cancel(ctx, &clngrpcHold.CancelRequest{
		PaymentHash: paymentHashBytes,
	})
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"paymentHash": paymentHash,
		}).WithError(err).Error("Failed to cancel hold invoice")
		return fmt.Errorf("Failed to cancel hold invoice: %v", err)
	}

	return nil
}

func (c *CLNService) MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64, throughNodePubkey *string) (transaction *lnclient.Transaction, err error) {
	logger.Logger.WithFields(logrus.Fields{
		"amount":              amount,
		"description":         description,
		"description_hash":    descriptionHash,
		"expiry":              expiry,
		"through_node_pubkey": throughNodePubkey,
	}).Debug("Make Invoice")

	label := "AlbyHub-" + uuid.NewString()

	var deschashonly bool
	if descriptionHash != "" {
		if description == "" {
			return nil, fmt.Errorf("Must have description when using description_hash")
		}
		myDescriptionHash := sha256.Sum256([]byte(description))
		if descriptionHash != hex.EncodeToString(myDescriptionHash[:]) {
			return nil, fmt.Errorf("description_hash does not match description")
		}
		deschashonly = true
	}

	if expiry == 0 {
		expiry = lnclient.DEFAULT_INVOICE_EXPIRY
	}
	myExpiry := uint64(expiry)

	Amount := clngrpc.AmountOrAny{
		Value: &clngrpc.AmountOrAny_Amount{Amount: &clngrpc.Amount{Msat: uint64(amount)}}}
	// amount 0 is often used for "any" amount but CLN doesn't support 0 directly
	if amount == 0 {
		Amount = clngrpc.AmountOrAny{
			Value: &clngrpc.AmountOrAny_Any{Any: true}}
	}

	req := &clngrpc.InvoiceRequest{
		Description:  description,
		Label:        label,
		Expiry:       &myExpiry,
		Deschashonly: &deschashonly,
		AmountMsat:   &Amount,
	}

	Exposeprivatechannels := []string{}

	if throughNodePubkey != nil {
		throughNodePubkeyBytes, err := hex.DecodeString(*throughNodePubkey)
		if err != nil {
			return nil, fmt.Errorf("Could not convert throughNodePubkey to bytes")
		}
		lpc, err := c.client.ListPeerChannels(ctx, &clngrpc.ListpeerchannelsRequest{
			Id: throughNodePubkeyBytes,
		})
		if err != nil {
			logger.Logger.WithError(err).Error("listpeerchannels failed")
			return nil, fmt.Errorf("listpeerchannels failed")
		}

		for _, channel := range lpc.Channels {
			if channel.ShortChannelId != nil {
				Exposeprivatechannels = append(Exposeprivatechannels, *channel.ShortChannelId)
				continue
			}
			if channel.Alias != nil {
				if channel.Alias.Remote != nil {
					Exposeprivatechannels = append(Exposeprivatechannels, *channel.Alias.Remote)
				}
			}
		}
	}

	if len(Exposeprivatechannels) > 0 {
		req.Exposeprivatechannels = Exposeprivatechannels
	}

	resp, err := c.client.Invoice(ctx, req)
	if err != nil {
		logger.Logger.WithError(err).Error("invoice failed")
		return nil, fmt.Errorf("invoice failed: %w", err)
	}

	expiresAt := int64(resp.ExpiresAt)

	transaction = &lnclient.Transaction{
		Type:            "incoming",
		Invoice:         resp.Bolt11,
		Description:     description,
		DescriptionHash: descriptionHash,
		Preimage:        "",
		PaymentHash:     hex.EncodeToString(resp.PaymentHash),
		Amount:          amount,
		FeesPaid:        0,
		CreatedAt:       time.Now().Unix(),
		ExpiresAt:       &expiresAt,
		SettledAt:       nil,
		Metadata:        lnclient.Metadata{},
		SettleDeadline:  nil,
	}

	return transaction, nil
}

func (c *CLNService) MakeOffer(ctx context.Context, description string) (string, error) {
	logger.Logger.WithFields(logrus.Fields{
		"description": description,
	}).Debug("Make Offer")

	req := &clngrpc.OfferRequest{
		Description: &description,
	}
	resp, err := c.client.Offer(ctx, req)
	if err != nil {
		logger.Logger.WithError(err).Error("offer failed")
		return "", fmt.Errorf("offer failed: %w", err)
	}
	if resp == nil {
		return "", fmt.Errorf("empty offer response")
	}

	return resp.Bolt12, nil
}

func (c *CLNService) OpenChannel(ctx context.Context, openChannelRequest *lnclient.OpenChannelRequest) (*lnclient.OpenChannelResponse, error) {
	logger.Logger.WithFields(logrus.Fields{
		"openChannelRequest": openChannelRequest,
	}).Debug("Open Channel")

	Amount := clngrpc.AmountOrAll{Value: &clngrpc.AmountOrAll_Amount{
		Amount: &clngrpc.Amount{Msat: uint64(openChannelRequest.AmountSats) * 1000},
	}}

	Id, err := hex.DecodeString(openChannelRequest.Pubkey)
	if err != nil {
		return nil, fmt.Errorf("Could not convert Pubkey to bytes")
	}

	req := &clngrpc.FundchannelRequest{
		Amount:   &Amount,
		Announce: &openChannelRequest.Public,
		Id:       Id,
	}
	resp, err := c.client.FundChannel(ctx, req)
	if err != nil {
		logger.Logger.WithError(err).Error("fundchannel failed")
		return nil, fmt.Errorf("fundchannel failed: %w", err)
	}

	if resp == nil {
		return nil, fmt.Errorf("empty fundchannel response")
	}

	FundingTxId := hex.EncodeToString(resp.Txid)

	return &lnclient.OpenChannelResponse{
		FundingTxId: FundingTxId,
	}, nil

}

func (c *CLNService) RedeemOnchainFunds(ctx context.Context, toAddress string, amount uint64, feeRate *uint64, sendAll bool) (txId string, err error) {
	logger.Logger.WithFields(logrus.Fields{
		"toAddress": toAddress,
		"amount":    amount,
		"feeRate":   feeRate,
		"sendAll":   sendAll,
	}).Debug("Redeem Onchain Funds")

	Satoshi := clngrpc.AmountOrAll{Value: &clngrpc.AmountOrAll_Amount{
		Amount: &clngrpc.Amount{Msat: uint64(amount) * 1000},
	}}
	if sendAll {
		Satoshi = clngrpc.AmountOrAll{Value: &clngrpc.AmountOrAll_All{
			All: true,
		}}
	}

	req := &clngrpc.WithdrawRequest{
		Destination: toAddress,
		Satoshi:     &Satoshi,
	}

	if feeRate != nil {
		if *feeRate > math.MaxUint32/1000 {
			return "", fmt.Errorf("fee rate too high")
		}
		req.Feerate = &clngrpc.Feerate{
			Style: &clngrpc.Feerate_Perkb{
				Perkb: uint32(*feeRate) * 1000,
			},
		}
	}

	resp, err := c.client.Withdraw(ctx, req)
	if err != nil {
		logger.Logger.WithError(err).Error("withdraw failed")
		return "", fmt.Errorf("withdraw failed: %w", err)
	}

	if resp == nil {
		return "", fmt.Errorf("empty withdraw response")
	}

	return hex.EncodeToString(resp.Txid), nil

}

func (c *CLNService) ResetRouter(key string) error {
	return nil
}

func (c *CLNService) SendKeysend(amount uint64, destination string, customRecords []lnclient.TLVRecord, preimage string) (*lnclient.PayKeysendResponse, error) {
	logger.Logger.WithFields(logrus.Fields{
		"amount":        amount,
		"destination":   destination,
		"customRecords": customRecords,
		"preimage":      preimage,
	}).Debug("Send Keysend")

	if preimage != "" {
		return nil, errors.New("preimage not supported for keysends")
	}

	Destination, err := hex.DecodeString(destination)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to decode payee pubkey")
		return nil, err
	}

	req := &clngrpc.KeysendRequest{
		Destination: Destination,
		AmountMsat:  &clngrpc.Amount{Msat: amount},
	}

	if len(customRecords) > 0 {
		Extratlvs := clngrpc.TlvStream{}
		for _, record := range customRecords {
			valueBytes, err := hex.DecodeString(record.Value)
			if err != nil {
				return nil, fmt.Errorf("could not decode TLV value to bytes: %v", record.Value)
			}

			entry := clngrpc.TlvEntry{
				Type:  record.Type,
				Value: valueBytes,
			}

			Extratlvs.Entries = append(Extratlvs.Entries, &entry)
		}
		req.Extratlvs = &Extratlvs
	}

	resp, err := c.client.KeySend(c.ctx, req)
	if err != nil {
		logger.Logger.WithError(err).Error("keysend failed")
		return nil, fmt.Errorf("keysend failed: %w", err)
	}

	if resp == nil {
		return nil, fmt.Errorf("empty keysend response")
	}

	Fee := uint64(0)

	if resp.AmountSentMsat != nil && resp.AmountMsat != nil {
		Fee = resp.AmountSentMsat.Msat - resp.AmountMsat.Msat
	}
	return &lnclient.PayKeysendResponse{Fee: Fee}, nil
}

func (c *CLNService) SendPaymentProbes(ctx context.Context, invoice string) error {
	return nil
}

func (c *CLNService) SendPaymentSync(payReq string, amount *uint64) (*lnclient.PayInvoiceResponse, error) {
	logger.Logger.WithFields(logrus.Fields{
		"payReq": payReq,
		"amount": amount,
	}).Debug("Send Payment Sync")

	dec_req := &clngrpc.DecodeRequest{
		String_: payReq,
	}

	dec_resp, err := c.client.Decode(c.ctx, dec_req)
	if err != nil {
		logger.Logger.WithError(err).Error("decode failed")
		return nil, fmt.Errorf("decode failed: %w", err)
	}
	if dec_resp == nil {
		return nil, fmt.Errorf("decode result empty")
	}
	if !dec_resp.Valid {
		return nil, fmt.Errorf("payReq not valid")
	}

	var amountMsat *clngrpc.Amount
	if amount != nil {
		amountMsat = &clngrpc.Amount{
			Msat: *amount,
		}
	}

	req := &clngrpc.XpayRequest{
		Invstring:  payReq,
		AmountMsat: amountMsat,
	}

	resp, err := c.client.Xpay(c.ctx, req)
	if err != nil {
		logger.Logger.WithError(err).Error("xpay failed")
		return nil, fmt.Errorf("xpay failed: %w", err)
	}

	feePaid := uint64(0)
	if resp.AmountSentMsat != nil {
		if resp.AmountMsat != nil {
			feePaid = resp.AmountSentMsat.Msat - resp.AmountMsat.Msat
		}
	}

	return &lnclient.PayInvoiceResponse{
		Preimage: hex.EncodeToString(resp.PaymentPreimage),
		Fee:      feePaid,
	}, err
}

func (c *CLNService) SendSpontaneousPaymentProbes(ctx context.Context, amountMsat uint64, nodeId string) error {
	return nil
}

func (c *CLNService) Shutdown() error {
	logger.Logger.Info("Cancelling CLN context")
	c.cancel()

	logger.Logger.Info("Closing gRPC connections")
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			logger.Logger.WithError(err).Error("Failed to close CLN gRPC connection")
		}
	}

	if c.connHold != nil {
		if err := c.connHold.Close(); err != nil {
			logger.Logger.WithError(err).Error("Failed to close CLN hold plugin gRPC connection")
		}
	}

	logger.Logger.Info("CLN backend shutdown complete")
	return nil
}

func (c *CLNService) SignMessage(ctx context.Context, message string) (string, error) {
	logger.Logger.WithFields(logrus.Fields{
		"message": message,
	}).Debug("Signing Message")

	req := &clngrpc.SignmessageRequest{
		Message: message,
	}
	resp, err := c.client.SignMessage(ctx, req)
	if err != nil {
		logger.Logger.WithError(err).Error("signmessage failed")
		return "", fmt.Errorf("signmessage failed: %w", err)
	}
	if resp == nil {
		return "", fmt.Errorf("signmessage result empty")
	}

	return resp.Zbase, nil
}

func (c *CLNService) UpdateChannel(ctx context.Context, updateChannelRequest *lnclient.UpdateChannelRequest) error {
	logger.Logger.WithFields(logrus.Fields{
		"updateChannelRequest": updateChannelRequest,
	}).Debug("Updating Channel")

	req := &clngrpc.SetchannelRequest{
		Id:      updateChannelRequest.ChannelId,
		Feebase: &clngrpc.Amount{Msat: uint64(updateChannelRequest.ForwardingFeeBaseMsat)},
		Feeppm:  &updateChannelRequest.ForwardingFeeProportionalMillionths,
	}

	resp, err := c.client.SetChannel(ctx, req)
	if err != nil {
		logger.Logger.WithError(err).Error("setchannel failed")
		return fmt.Errorf("setchannel failed: %w", err)
	}
	if resp == nil {
		return fmt.Errorf("setchannel result empty")
	}

	return nil
}

func (c *CLNService) UpdateLastWalletSyncRequest() {
}
