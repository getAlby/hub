package greenlight

import (
	"context"
	crand "crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/lnclient/cln/clngrpc"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
)

type Config struct {
	// CredsPath is the directory containing the extracted device credentials
	// (ca.pem, client.pem, client-key.pem) produced by extract_creds.py
	CredsPath string
	// NodeURI is the greenlight node gRPC endpoint, e.g. gl1<node_id>.gl.blckstrm.com:443
	NodeURI string
	// ServerName is the TLS SNI host. Defaults to the host of NodeURI.
	ServerName string
	Network    string
	// SignerDataDir is the signer's data directory. When set, hub backups
	// include it (covers the encrypted seed when signer and hub share a
	// filesystem). Optional; defaults to the hub workdir.
	SignerDataDir string
	// SignerStatusProvider returns the signer service's current state.
	// Nil in external-signer mode (the harness provides the signer).
	SignerStatusProvider func() SignerStatus
}

// SignerStatus carries the signer service's health state for surfacing
// in the node-status API (so a signer outage doesn't masquerade as a
// node outage).
type SignerStatus struct {
	Running   bool   `json:"running"`
	LastError string `json:"last_error,omitempty"`
}

type GreenlightService struct {
	ctx            context.Context
	cancel         context.CancelFunc
	conn           *grpc.ClientConn
	client         clngrpc.NodeClient
	eventPublisher events.EventPublisher
	pubkey         string
	config         Config
	workDir        string
	pumpRetryDelay time.Duration

	// node health watchdog state (see health.go)
	healthMtx              sync.RWMutex
	nodeHealthy            bool
	lastHealthError        string
	lastHealthCheck        time.Time
	lastInfo               *clngrpc.GetinfoResponse
	lastPumpCallStart      time.Time
	inPumpCall             bool
	healthCheckInterval    time.Duration
	healthCheckTimeout     time.Duration
	healthFailureThreshold int
	pumpStallThreshold     time.Duration
}

func NewGreenlightService(ctx context.Context, eventPublisher events.EventPublisher, workDir string, config Config) (lnclient.LNClient, error) {
	if config.CredsPath == "" {
		return nil, errors.New("greenlight creds path not configured (GREENLIGHT_CREDS_PATH)")
	}
	if config.NodeURI == "" {
		return nil, errors.New("greenlight node URI not configured (GREENLIGHT_NODE_URI)")
	}

	// Normalize: glcli prints node URIs as https://gl1<id>.gl.blckstrm.com:443,
	// but the grpc dialer rejects a scheme and hostOf() needs bare host:port.
	// Strip the scheme defensively so both forms work.
	nodeURI := normalizeNodeURI(config.NodeURI)

	serverName := config.ServerName
	if serverName == "" {
		serverName = hostOf(nodeURI)
	}

	tlsConfig, err := loadTLSCredentials(config.CredsPath, serverName)
	if err != nil {
		return nil, fmt.Errorf("failed to load greenlight TLS credentials: %w", err)
	}

	creds := credentials.NewTLS(tlsConfig)

	conn, err := grpc.NewClient(nodeURI, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to greenlight gRPC: %w", err)
	}

	client := clngrpc.NewNodeClient(conn)

	ctx, cancel := context.WithCancel(ctx)

	svc := &GreenlightService{
		ctx:            ctx,
		cancel:         cancel,
		conn:           conn,
		client:         client,
		eventPublisher: eventPublisher,
		config:         config,
		workDir:        workDir,
	}

	// Fail fast on a truly unreachable node — but never lock the user out
	// of their hub: a *frozen* node (Blockstream/greenlight#739 — hsmd
	// queue wedge) responds to nothing, so this probe times out even
	// though the node exists. Aborting startup on that would leave the
	// user unable to reach their wallet, backups, or apps exactly when
	// the node is broken. Instead: start degraded and let the health
	// watchdog (health.go) report the node state and keep probing.
	logger.Logger.Info("Testing greenlight gRPC connection")
	connCtx, connCancel := context.WithTimeout(ctx, 15*time.Second)
	info, err := svc.GetInfo(connCtx)
	connCancel()
	if err != nil {
		logger.Logger.WithError(err).Warn("greenlight node unreachable at boot — starting degraded; the health watchdog will keep probing (a frozen node may need a Blockstream restart)")
	} else {
		svc.pubkey = info.Pubkey
		logger.Logger.Info("Successfully connected to greenlight via gRPC")
	}

	// NOTE: no channel-state subscription. The GL node server leaves all six
	// cln.Node Subscribe* stream RPCs unimplemented (they panic the gl-plugin
	// process, killing the node), so nwc_channel_ready/closed are not emitted
	// by this backend. Incoming payments are covered by waitForInvoices (pump)
	// and streamIncoming (keysends) below.
	go svc.waitForInvoices(ctx)
	go svc.streamIncoming(ctx)

	// node health watchdog (health.go): detects a wedged/frozen node even
	// when the node itself stops responding — GetNodeStatus then serves the
	// cached verdict instead of hanging on a live RPC.
	svc.healthCheckInterval = defaultHealthCheckInterval
	svc.healthCheckTimeout = defaultHealthCheckTimeout
	svc.healthFailureThreshold = defaultHealthFailureThreshold
	svc.pumpStallThreshold = defaultPumpStallThreshold
	svc.startHealthWatchdog(ctx)

	return svc, nil
}

func normalizeNodeURI(uri string) string {
	uri = strings.TrimPrefix(uri, "https://")
	return strings.TrimPrefix(uri, "http://")
}

func hostOf(uri string) string {
	host := uri
	if idx := strings.LastIndex(host, "@"); idx != -1 {
		host = host[idx+1:]
	}
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}
	return host
}

func loadTLSCredentials(credsPath string, serverName string) (*tls.Config, error) {
	certPath := filepath.Join(credsPath, "ca.pem")
	clientCertPath := filepath.Join(credsPath, "client.pem")
	clientKeyPath := filepath.Join(credsPath, "client-key.pem")

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
		ServerName:   serverName,
		MinVersion:   tls.VersionTLS12,
	}, nil
}

func (g *GreenlightService) SendPaymentSync(payReq string, amount *uint64) (*lnclient.PayInvoiceResponse, error) {
	logger.Logger.WithFields(logrus.Fields{
		"payReq": payReq,
		"amount": amount,
	}).Debug("Send Payment Sync")

	dec_req := &clngrpc.DecodeRequest{
		String_: payReq,
	}

	dec_resp, err := g.client.Decode(g.ctx, dec_req)
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

	resp, err := g.client.Xpay(g.ctx, req)
	if err != nil {
		logger.Logger.WithError(err).Error("xpay failed")
		return nil, fmt.Errorf("xpay failed: %w", err)
	}

	feePaidMsat := uint64(0)
	if resp.AmountSentMsat != nil {
		if resp.AmountMsat != nil {
			feePaidMsat = resp.AmountSentMsat.Msat - resp.AmountMsat.Msat
		}
	}

	return &lnclient.PayInvoiceResponse{
		Preimage: hex.EncodeToString(resp.PaymentPreimage),
		FeeMsat:  feePaidMsat,
	}, err
}

func (g *GreenlightService) SendKeysend(amount uint64, destination string, customRecords []lnclient.TLVRecord, preimage string) (*lnclient.PayKeysendResponse, error) {
	logger.Logger.WithFields(logrus.Fields{
		"amount":        amount,
		"destination":   destination,
		"customRecords": customRecords,
		"preimage":      preimage,
	}).Debug("Send Keysend")

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

	resp, err := g.client.KeySend(g.ctx, req)
	if err != nil {
		logger.Logger.WithError(err).Error("keysend failed")
		return nil, fmt.Errorf("keysend failed: %w", err)
	}

	if resp == nil {
		return nil, fmt.Errorf("empty keysend response")
	}

	feeMsat := uint64(0)

	if resp.AmountSentMsat != nil && resp.AmountMsat != nil {
		feeMsat = resp.AmountSentMsat.Msat - resp.AmountMsat.Msat
	}
	// Greenlight (CLN-family) derives its own preimage server-side — the
	// caller-supplied preimage cannot be honored via the gl-client, so
	// report the actual preimage and payment hash from the response.
	// Recording anything else would fake the proof-of-payment.
	return &lnclient.PayKeysendResponse{
		FeeMsat:     feeMsat,
		Preimage:    hex.EncodeToString(resp.PaymentPreimage),
		PaymentHash: hex.EncodeToString(resp.PaymentHash),
	}, nil
}

func (g *GreenlightService) GetPubkey() string {
	return g.pubkey
}

func (g *GreenlightService) GetInfo(ctx context.Context) (*lnclient.NodeInfo, error) {
	resp, err := g.client.Getinfo(ctx, &clngrpc.GetinfoRequest{})
	if err != nil {
		return nil, fmt.Errorf("getinfo failed: %w", err)
	}

	return &lnclient.NodeInfo{
		Alias:       resp.GetAlias(),
		Color:       hex.EncodeToString(resp.Color),
		Pubkey:      hex.EncodeToString(resp.Id),
		Network:     resp.Network,
		BlockHeight: resp.Blockheight,
		BlockHash:   "",
	}, nil
}

func (g *GreenlightService) MakeInvoice(ctx context.Context, amountMsat int64, description string, descriptionHash string, expiry int64, throughNodePubkey *string) (*lnclient.Transaction, error) {
	logger.Logger.WithFields(logrus.Fields{
		"amount":              amountMsat,
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
		Value: &clngrpc.AmountOrAny_Amount{Amount: &clngrpc.Amount{Msat: uint64(amountMsat)}}}
	// amount 0 is often used for "any" amount but CLN doesn't support 0 directly
	if amountMsat == 0 {
		Amount = clngrpc.AmountOrAny{
			Value: &clngrpc.AmountOrAny_Any{Any: true}}
	}

	preimage, err := GeneratePreimage()
	if err != nil {
		return nil, err
	}

	req := &clngrpc.InvoiceRequest{
		Description:  description,
		Label:        label,
		Preimage:     preimage,
		Expiry:       &myExpiry,
		Deschashonly: &deschashonly,
		AmountMsat:   &Amount,
	}

	Exposeprivatechannels := []string{}

	if throughNodePubkey != nil {
		throughNodePubkeyBytes, err := hex.DecodeString(*throughNodePubkey)
		if err != nil {
			return nil, err
		}
		lpc, err := g.client.ListPeerChannels(ctx, &clngrpc.ListpeerchannelsRequest{
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

	resp, err := g.client.Invoice(ctx, req)
	if err != nil {
		logger.Logger.WithError(err).Error("invoice failed")
		return nil, fmt.Errorf("invoice failed: %w", err)
	}

	expiresAt := int64(resp.ExpiresAt)

	transaction := &lnclient.Transaction{
		Type:            "incoming",
		Invoice:         resp.Bolt11,
		Description:     description,
		DescriptionHash: descriptionHash,
		Preimage:        hex.EncodeToString(preimage),
		PaymentHash:     hex.EncodeToString(resp.PaymentHash),
		AmountMsat:      amountMsat,
		FeesPaidMsat:    0,
		CreatedAt:       time.Now().Unix(),
		ExpiresAt:       &expiresAt,
		SettledAt:       nil,
		Metadata:        lnclient.Metadata{},
		SettleDeadline:  nil,
	}

	return transaction, nil
}

// Hold invoices are not supported by Greenlight hosted nodes (no plugins).
func (g *GreenlightService) MakeHoldInvoice(ctx context.Context, amountMsat int64, description string, descriptionHash string, expiry int64, paymentHash string, minCltvExpiryDelta *uint64) (*lnclient.Transaction, error) {
	return nil, errors.New("hold invoices are not supported by the greenlight backend")
}

func (g *GreenlightService) SettleHoldInvoice(ctx context.Context, preimage string) error {
	return errors.New("hold invoices are not supported by the greenlight backend")
}

func (g *GreenlightService) CancelHoldInvoice(ctx context.Context, paymentHash string) error {
	return errors.New("hold invoices are not supported by the greenlight backend")
}

func (g *GreenlightService) LookupInvoice(ctx context.Context, paymentHash string) (*lnclient.Transaction, error) {
	logger.Logger.WithFields(logrus.Fields{
		"paymentHash": paymentHash,
	}).Debug("Lookup Invoice")

	paymentHashBytes, err := hex.DecodeString(paymentHash)
	if err != nil {
		logger.Logger.WithError(err).Error("failed to decode payment hash")
		return nil, fmt.Errorf("failed to decode payment hash: %w", err)
	}
	req := &clngrpc.ListinvoicesRequest{PaymentHash: paymentHashBytes}

	resp, err := g.client.ListInvoices(ctx, req)
	if err != nil {
		logger.Logger.WithError(err).Error("listinvoices failed")
		return nil, fmt.Errorf("listinvoices failed: %w", err)
	}
	if len(resp.Invoices) == 0 {
		return nil, fmt.Errorf("invoice not found")
	}

	transaction, err := g.clnInvoiceToTransaction(ctx, resp.Invoices[0])
	if err != nil {
		logger.Logger.WithError(err).Error("failed to convert invoice to transaction")
		return nil, fmt.Errorf("failed to convert invoice to transaction: %w", err)
	}

	return transaction, nil
}

func (g *GreenlightService) ListOnchainTransactions(ctx context.Context) ([]lnclient.OnchainTransaction, error) {
	account := "wallet"
	req := &clngrpc.BkprlistaccounteventsRequest{
		Account: &account,
	}
	bkpr, err := g.client.BkprListAccountEvents(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("bkprlistaccountevents failed: %w", err)
	}

	infoResp, infoErr := g.client.Getinfo(ctx, &clngrpc.GetinfoRequest{})
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

func (g *GreenlightService) Shutdown() error {
	logger.Logger.Info("Cancelling greenlight context")
	g.cancel()

	logger.Logger.Info("Closing gRPC connections")
	if g.conn != nil {
		if err := g.conn.Close(); err != nil {
			logger.Logger.WithError(err).Error("Failed to close greenlight gRPC connection")
		}
	}

	logger.Logger.Info("greenlight backend shutdown complete")
	return nil
}

func (g *GreenlightService) ListChannels(ctx context.Context) ([]lnclient.Channel, error) {
	resp, err := g.client.ListPeerChannels(ctx, &clngrpc.ListpeerchannelsRequest{})
	if err != nil {
		logger.Logger.WithError(err).Error("listpeerchannels failed")
		return nil, err
	}

	infoResp, infoErr := g.client.Getinfo(ctx, &clngrpc.GetinfoRequest{})
	if infoErr != nil {
		logger.Logger.WithError(infoErr).Error("getinfo failed")
		return nil, infoErr
	}

	blockheight := infoResp.Blockheight

	reChanHeight := regexp.MustCompile(`(\d+)x.*`)

	channels := []lnclient.Channel{}

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
			LocalBalanceMsat:                    LocalBalance,
			LocalSpendableBalanceMsat:           msatInt64(channel.SpendableMsat),
			RemoteBalanceMsat:                   RemoteBalance,
			Id:                                  hex.EncodeToString(channel.ChannelId),
			RemotePubkey:                        hex.EncodeToString(channel.PeerId),
			FundingTxId:                         hex.EncodeToString(channel.FundingTxid),
			FundingTxVout:                       channel.GetFundingOutnum(),
			Active:                              isActive,
			Public:                              !channel.GetPrivate(),
			InternalChannel:                     channel,
			Confirmations:                       &Confirmations,
			ConfirmationsRequired:               &ConfirmationsRequired,
			ForwardingFeeBaseMsat:               localFeeBaseMsat(channel),
			ForwardingFeeProportionalMillionths: localFeePPM(channel),
			UnspendablePunishmentReserveSat:     sat(channel.OurReserveMsat),
			CounterpartyUnspendablePunishmentReserveSat: sat(channel.TheirReserveMsat),
			Error:      Error,
			IsOutbound: channel.GetOpener() == clngrpc.ChannelSide_LOCAL,
		})
	}
	return channels, nil
}

func (g *GreenlightService) GetNodeConnectionInfo(ctx context.Context) (*lnclient.NodeConnectionInfo, error) {
	resp, err := g.client.Getinfo(ctx, &clngrpc.GetinfoRequest{})
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

func (g *GreenlightService) GetNodeStatus(ctx context.Context) (*lnclient.NodeStatus, error) {
	// Serve the watchdog's cached state: a live Getinfo here would hang
	// when the node is frozen (the API must keep answering with an
	// "unhealthy" verdict, not block).
	g.healthMtx.RLock()
	cacheWarm := !g.lastHealthCheck.IsZero()
	g.healthMtx.RUnlock()
	if !cacheWarm {
		// cold cache (first seconds after startup): warm it with one
		// bounded probe — never more than the health timeout
		g.runHealthCheck(ctx)
	}

	g.healthMtx.RLock()
	healthy := g.nodeHealthy
	lastErr := g.lastHealthError
	lastCheck := g.lastHealthCheck
	info := g.lastInfo
	g.healthMtx.RUnlock()

	ready := healthy
	if info != nil && (info.WarningBitcoindSync != nil || info.WarningLightningdSync != nil) {
		ready = false
	}

	hs := nodeHealthStatus{
		Healthy:     healthy,
		LastCheckAt: lastCheck.Unix(),
		LastError:   lastErr,
	}
	if g.config.SignerStatusProvider != nil {
		hs.Signer = g.config.SignerStatusProvider()
	}

	return &lnclient.NodeStatus{
		IsReady:            ready,
		InternalNodeStatus: hs,
	}, nil
}

func (g *GreenlightService) ConnectPeer(ctx context.Context, connectPeerRequest *lnclient.ConnectPeerRequest) error {
	logger.Logger.WithFields(logrus.Fields{
		"connectPeerRequest": connectPeerRequest,
	}).Debug("Connecting to Peer")

	// CLN gRPC's ConnectRequest uses a oneof for host/port, so we can't
	// safely set both via the generated Go struct. Send the socket address
	// as the host field instead; CLN will parse host:port itself.
	socket := fmt.Sprintf("%s:%d", connectPeerRequest.Address, connectPeerRequest.Port)
	req := &clngrpc.ConnectRequest{
		Id:   connectPeerRequest.Pubkey,
		Host: &socket,
	}

	_, err := g.client.ConnectPeer(ctx, req)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to connect peer")
		return err
	}

	return nil
}

func (g *GreenlightService) OpenChannel(ctx context.Context, openChannelRequest *lnclient.OpenChannelRequest) (*lnclient.OpenChannelResponse, error) {
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
	resp, err := g.client.FundChannel(ctx, req)
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

func (g *GreenlightService) CloseChannel(ctx context.Context, closeChannelRequest *lnclient.CloseChannelRequest) error {
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

	_, err := g.client.Close(ctx, req)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to close channel")
		return fmt.Errorf("close failed: %w", err)
	}

	return nil
}

func (g *GreenlightService) UpdateChannel(ctx context.Context, updateChannelRequest *lnclient.UpdateChannelRequest) error {
	logger.Logger.WithFields(logrus.Fields{
		"updateChannelRequest": updateChannelRequest,
	}).Debug("Updating Channel")

	req := &clngrpc.SetchannelRequest{
		Id:      updateChannelRequest.ChannelId,
		Feebase: &clngrpc.Amount{Msat: uint64(updateChannelRequest.ForwardingFeeBaseMsat)},
		Feeppm:  &updateChannelRequest.ForwardingFeeProportionalMillionths,
	}

	resp, err := g.client.SetChannel(ctx, req)
	if err != nil {
		logger.Logger.WithError(err).Error("setchannel failed")
		return fmt.Errorf("setchannel failed: %w", err)
	}
	if resp == nil {
		return fmt.Errorf("setchannel result empty")
	}

	return nil
}

func (g *GreenlightService) DisconnectPeer(ctx context.Context, peerId string) error {
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

	_, err = g.client.Disconnect(ctx, req)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to disconnect peer")
		return err
	}

	return nil
}

func (g *GreenlightService) MakeOffer(ctx context.Context, description string) (string, error) {
	logger.Logger.WithFields(logrus.Fields{
		"description": description,
	}).Debug("Make Offer")

	req := &clngrpc.OfferRequest{
		Amount:      "any",
		Description: &description,
	}
	resp, err := g.client.Offer(ctx, req)
	if err != nil {
		logger.Logger.WithError(err).Error("offer failed")
		return "", fmt.Errorf("offer failed: %w", err)
	}
	if resp == nil {
		return "", fmt.Errorf("empty offer response")
	}

	return resp.Bolt12, nil
}

func (g *GreenlightService) GetNewOnchainAddress(ctx context.Context) (string, error) {
	resp, err := g.client.NewAddr(ctx, &clngrpc.NewaddrRequest{})
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

func (g *GreenlightService) ResetRouter(key string) error {
	return errors.New("not implemented")
}

func (g *GreenlightService) GetOnchainBalance(ctx context.Context) (*lnclient.OnchainBalanceResponse, error) {
	lf, err := g.client.ListFunds(ctx, &clngrpc.ListfundsRequest{})
	if err != nil {
		return nil, fmt.Errorf("listfunds failed: %w", err)
	}

	lpc, err := g.client.ListPeerChannels(ctx, &clngrpc.ListpeerchannelsRequest{})
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
		balances.TotalSat += amt

		if utxo.Reserved {
			balances.ReservedSat += amt
			reservedSats += amt
		}

		switch utxo.Status {
		case clngrpc.ListfundsOutputs_CONFIRMED:
			if !utxo.Reserved {
				balances.SpendableSat += amt
			}

		case clngrpc.ListfundsOutputs_UNCONFIRMED:
			balances.PendingSweepBalancesDetails = append(
				balances.PendingSweepBalancesDetails,
				lnclient.PendingBalanceDetails{
					AmountSat:     uint64(amt),
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
		balances.PendingBalancesFromChannelClosuresSat += amt
		chanIdStr := hex.EncodeToString(ch.ChannelId)

		detail := lnclient.PendingBalanceDetails{
			ChannelId: chanIdStr,
			NodeId:    hex.EncodeToString(ch.PeerId),
			AmountSat: amt,
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

func (g *GreenlightService) GetBalances(ctx context.Context, includeInactiveChannels bool) (*lnclient.BalancesResponse, error) {
	logger.Logger.WithFields(logrus.Fields{
		"includeInactiveChannels": includeInactiveChannels,
	}).Debug("Get all Balances")

	onchainBalance, err := g.GetOnchainBalance(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := g.client.ListPeerChannels(ctx, &clngrpc.ListpeerchannelsRequest{})
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
			lightning.TotalSpendableMsat += spendable

			if spendable > lightning.NextMaxSpendableMsat {
				lightning.NextMaxSpendableMsat = spendable
			}
		}

		if ch.ReceivableMsat != nil {
			receivable := int64(ch.ReceivableMsat.Msat)
			lightning.TotalReceivableMsat += receivable

			if receivable > lightning.NextMaxReceivableMsat {
				lightning.NextMaxReceivableMsat = receivable
			}
		}
	}

	lightning.NextMaxSpendableMPPMsat = lightning.TotalSpendableMsat
	lightning.NextMaxReceivableMPPMsat = lightning.TotalReceivableMsat

	return &lnclient.BalancesResponse{
		Onchain:   *onchainBalance,
		Lightning: lightning,
	}, nil
}

func (g *GreenlightService) RedeemOnchainFunds(ctx context.Context, toAddress string, amount uint64, feeRate *uint64, sendAll bool) (string, error) {
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

	resp, err := g.client.Withdraw(ctx, req)
	if err != nil {
		logger.Logger.WithError(err).Error("withdraw failed")
		return "", fmt.Errorf("withdraw failed: %w", err)
	}

	if resp == nil {
		return "", fmt.Errorf("empty withdraw response")
	}

	return hex.EncodeToString(resp.Txid), nil
}

func (g *GreenlightService) ListPeers(ctx context.Context) ([]lnclient.PeerDetails, error) {
	resp, err := g.client.ListPeers(ctx, &clngrpc.ListpeersRequest{})
	if err != nil {
		return nil, fmt.Errorf("listpeers failed: %w", err)
	}

	peers := make([]lnclient.PeerDetails, 0, len(resp.Peers))
	for _, peer := range resp.Peers {
		if peer == nil {
			continue
		}

		req_node := &clngrpc.ListnodesRequest{Id: peer.Id}

		resp_node, err := g.client.ListNodes(ctx, req_node)
		if err != nil {
			return nil, fmt.Errorf("listnodes failed: %w", err)
		}

		if len(resp_node.Nodes) == 0 {
			addr := ""
			peers = append(peers, lnclient.PeerDetails{
				NodeId:      hex.EncodeToString(peer.Id),
				Address:     addr,
				IsPersisted: peer.GetNumChannels() > 0,
				IsConnected: peer.Connected,
			})
			continue
		}

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
			addr := ""
			selected = &clngrpc.ListnodesNodesAddresses{
				Address: &addr,
				Port:    0,
			}
		}

		peers = append(peers, lnclient.PeerDetails{
			NodeId:      hex.EncodeToString(peer.Id),
			Address:     selected.GetAddress(),
			IsPersisted: peer.GetNumChannels() > 0,
			IsConnected: peer.Connected,
		})
	}

	return peers, nil
}

func (g *GreenlightService) GetLogOutput(ctx context.Context, maxLen int) ([]byte, error) {
	return []byte{}, nil
}

func (g *GreenlightService) SignMessage(ctx context.Context, message string) (string, error) {
	// NOT SUPPORTED: empirically (regtest harness + live testnet node,
	// v26.06gl1 with glcli signer) the VLS signer never completes the
	// hsmd SignMessage request, which wedges lightningd's serial hsmd
	// queue and freezes ALL signing operations (invoice creation starts
	// timing out) until the hosted node restarts. The same freeze was
	// documented for the gl-testing python signer; the production Rust
	// VLS signer hangs identically. Return a clean error instead of
	// freezing the node.
	return "", errors.New("sign message is not supported by the greenlight backend (VLS signer hangs on signmessage, freezing the node's hsmd queue)")
}

func (g *GreenlightService) GetStorageDir() (string, error) {
	// Prefer directories that actually exist so Hub .bkp includes device PEMs / seed.
	// SignerDataDir (product path) → CredsPath (connect-only) → workDir.
	candidates := []string{g.config.SignerDataDir, g.config.CredsPath, g.workDir}
	for _, p := range candidates {
		if p == "" {
			continue
		}
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			return p, nil
		}
	}
	if g.config.SignerDataDir != "" {
		return g.config.SignerDataDir, nil
	}
	return g.workDir, nil
}

func (g *GreenlightService) GetNetworkGraph(ctx context.Context, nodeIds []string) (lnclient.NetworkGraphResponse, error) {
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

		listnode, err := g.client.ListNodes(ctx, &clngrpc.ListnodesRequest{Id: nodeIdBytes})
		if err != nil {
			logger.Logger.WithError(err).Error("listnodes failed")
			return nil, err
		}
		listnodes = append(listnodes, listnode.Nodes...)

		listchannel, err := g.client.ListChannels(ctx, &clngrpc.ListchannelsRequest{Source: nodeIdBytes})
		if err != nil {
			logger.Logger.WithError(err).Error("listchannels failed")
			return nil, err
		}
		listchannels = append(listchannels, listchannel.Channels...)

		listchannel, err = g.client.ListChannels(ctx, &clngrpc.ListchannelsRequest{Destination: nodeIdBytes})
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

func (g *GreenlightService) UpdateLastWalletSyncRequest() {
}

func (g *GreenlightService) GetSupportedNIP47Methods() []string {
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
		// sign_message intentionally omitted: the VLS signer hangs on it
		// and freezes the node (see SignMessage) — do not advertise a
		// method that can wedge the hsmd queue.
	}

	return methods
}

// Push tier: the WaitAnyInvoice pump publishes nwc_lnclient_payment_received
// with a persisted pay index, so payment_received notifications are safe.
// Advertising it switches off the hub's reconcile safety net
// (transactions_service.go:703) - the pump covers that gap by replaying
// anything that happened while the hub was down (index-based catch-up).
func (g *GreenlightService) GetSupportedNIP47NotificationTypes() []string {
	return []string{
		"payment_received",
	}
}

func (g *GreenlightService) GetCustomNodeCommandDefinitions() []lnclient.CustomNodeCommandDef {
	return nil
}

func (g *GreenlightService) ExecuteCustomNodeCommand(ctx context.Context, command *lnclient.CustomNodeCommandRequest) (*lnclient.CustomNodeCommandResponse, error) {
	return nil, nil
}

func (g *GreenlightService) clnInvoiceToTransaction(ctx context.Context, invoice *clngrpc.ListinvoicesInvoices) (*lnclient.Transaction, error) {
	var invstring string
	var bolt11Invoice string
	if invoice.Bolt11 != nil {
		invstring = *invoice.Bolt11
		bolt11Invoice = *invoice.Bolt11
	} else if invoice.Bolt12 != nil {
		invstring = *invoice.Bolt12
	} else {
		return nil, fmt.Errorf("bolt11 and bolt12 missing from invoice")
	}

	var amountMsat int64
	if invoice.Status == clngrpc.ListinvoicesInvoices_PAID && invoice.AmountReceivedMsat != nil {
		amountMsat = int64(invoice.AmountReceivedMsat.Msat)
	} else if invoice.AmountMsat != nil {
		amountMsat = int64(invoice.AmountMsat.Msat)
	} else {
		amountMsat = 0
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

	decoded_invoice, err := g.client.Decode(ctx, &clngrpc.DecodeRequest{String_: invstring})
	if err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	var created_at int64
	metadata := map[string]interface{}{}
	var description_hash string
	switch decoded_invoice.ItemType {
	case clngrpc.DecodeResponse_BOLT12_INVOICE:
		if decoded_invoice.InvoiceCreatedAt == nil {
			return nil, fmt.Errorf("invoice_created_at missing from bolt12 invoice")
		}
		created_at = int64(*decoded_invoice.InvoiceCreatedAt)

		offer := map[string]interface{}{}
		offer["id"] = hex.EncodeToString(decoded_invoice.OfferId)
		if invoice.InvreqPayerNote != nil {
			offer["payer_note"] = *invoice.InvreqPayerNote
		}
		metadata["offer"] = offer
	case clngrpc.DecodeResponse_BOLT11_INVOICE:
		if decoded_invoice.CreatedAt == nil {
			return nil, fmt.Errorf("created_at missing from bolt11 invoice")
		}
		created_at = int64(*decoded_invoice.CreatedAt)

		if decoded_invoice.DescriptionHash != nil {
			description_hash = hex.EncodeToString(decoded_invoice.DescriptionHash)
		}

	default:
		return nil, fmt.Errorf("invoice is not a bolt11 or bolt12 invoice")
	}

	var description string
	if invoice.Description != nil {
		description = *invoice.Description
	} else {
		description = ""
	}

	transaction := &lnclient.Transaction{
		Type:            "incoming",
		Invoice:         bolt11Invoice,
		Description:     description,
		DescriptionHash: description_hash,
		Preimage:        hex.EncodeToString(invoice.PaymentPreimage),
		PaymentHash:     hex.EncodeToString(invoice.PaymentHash),
		AmountMsat:      amountMsat,
		FeesPaidMsat:    0,
		CreatedAt:       created_at,
		ExpiresAt:       &expires_at,
		SettledAt:       paid_at,
		Metadata:        metadata,
	}
	return transaction, nil
}

func GeneratePreimage() ([]byte, error) {
	preimage := make([]byte, 32)

	_, err := crand.Read(preimage[:])
	if err != nil {
		return nil, fmt.Errorf("failed to generate preimage: %w", err)
	}

	return preimage, nil
}

// --- helpers ported from the CLN backend (same wire surface) ---

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
	return l.FeeProportionalMillionths
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
