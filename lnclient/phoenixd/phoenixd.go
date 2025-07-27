package phoenixd

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	decodepay "github.com/nbd-wtf/ln-decodepay"

	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"

	"github.com/sirupsen/logrus"
)

type InvoiceResponse struct {
	PaymentHash string `json:"paymentHash"`
	Preimage    string `json:"preimage"`
	ExternalId  string `json:"externalId"`
	Description string `json:"description"`
	Invoice     string `json:"invoice"`
	IsPaid      bool   `json:"isPaid"`
	ReceivedSat int64  `json:"receivedSat"`
	Fees        int64  `json:"fees"`
	CompletedAt int64  `json:"completedAt"`
	CreatedAt   int64  `json:"createdAt"`
}

type OutgoingPaymentResponse struct {
	PaymentHash string `json:"paymentHash"`
	Preimage    string `json:"preimage"`
	Invoice     string `json:"invoice"`
	IsPaid      bool   `json:"isPaid"`
	Sent        int64  `json:"sent"`
	Fees        int64  `json:"fees"`
	CompletedAt int64  `json:"completedAt"`
	CreatedAt   int64  `json:"createdAt"`
}

type PayResponse struct {
	PaymentHash     string `json:"paymentHash"`
	PaymentId       string `json:"paymentId"`
	PaymentPreimage string `json:"paymentPreimage"`
	RoutingFeeSat   int64  `json:"routingFeeSat"`
}

type MakeInvoiceResponse struct {
	AmountSat   int64  `json:"amountSat"`
	PaymentHash string `json:"paymentHash"`
	Serialized  string `json:"serialized"`
}

type InfoResponse struct {
	NodeId string `json:"nodeId"`
}

type BalanceResponse struct {
	BalanceSat   int64 `json:"balanceSat"`
	FeeCreditSat int64 `json:"feeCreditSat"`
}

type PhoenixService struct {
	Address       string
	Authorization string
	pubkey        string
	nodeInfo      *lnclient.NodeInfo
}

func NewPhoenixService(address string, authorization string) (result lnclient.LNClient, err error) {
	authorizationBase64 := b64.StdEncoding.EncodeToString([]byte(":" + authorization))
	// some environments (e.g. in a cloud environment like render.com) can only get the address and the port but not the protocol
	// in those cases we default to http for local requests
	if !strings.HasPrefix(address, "http") {
		address = "http://" + address
	}
	phoenixService := &PhoenixService{Address: address, Authorization: authorizationBase64}

	info, err := fetchNodeInfo(phoenixService)
	if err != nil {
		return nil, err
	}
	phoenixService.nodeInfo = info
	phoenixService.pubkey = info.Pubkey

	return phoenixService, nil
}

func (svc *PhoenixService) GetBalances(ctx context.Context, includeInactiveChannels bool) (*lnclient.BalancesResponse, error) {
	req, err := http.NewRequest(http.MethodGet, svc.Address+"/getbalance", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Basic "+svc.Authorization)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var balanceRes BalanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&balanceRes); err != nil {
		return nil, err
	}

	balance := balanceRes.BalanceSat * 1000

	return &lnclient.BalancesResponse{
		Onchain: lnclient.OnchainBalanceResponse{
			Spendable: 0,
			Total:     0,
		},
		Lightning: lnclient.LightningBalanceResponse{
			TotalSpendable:       balance,
			TotalReceivable:      0,
			NextMaxSpendable:     balance,
			NextMaxReceivable:    0,
			NextMaxSpendableMPP:  balance,
			NextMaxReceivableMPP: 0,
		},
	}, nil
}

func (svc *PhoenixService) ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaid bool, invoiceType string) (transactions []lnclient.Transaction, err error) {
	incomingQuery := url.Values{}
	if from != 0 {
		incomingQuery.Add("from", strconv.FormatUint(from*1000, 10))
	}
	if until != 0 {
		incomingQuery.Add("to", strconv.FormatUint(until*1000, 10))
	}
	if limit != 0 {
		incomingQuery.Add("limit", strconv.FormatUint(limit, 10))
	}
	if offset != 0 {
		incomingQuery.Add("offset", strconv.FormatUint(offset, 10))
	}
	incomingQuery.Add("all", strconv.FormatBool(unpaid))

	incomingUrl := svc.Address + "/payments/incoming?" + incomingQuery.Encode()

	logger.Logger.WithFields(logrus.Fields{
		"url": incomingUrl,
	}).Infof("Fetching incoming transactions: %s", incomingUrl)
	incomingReq, err := http.NewRequest(http.MethodGet, incomingUrl, nil)
	if err != nil {
		return nil, err
	}
	incomingReq.Header.Add("Authorization", "Basic "+svc.Authorization)
	client := &http.Client{Timeout: 5 * time.Second}

	incomingResp, err := client.Do(incomingReq)
	if err != nil {
		return nil, err
	}
	defer incomingResp.Body.Close()

	var incomingPayments []InvoiceResponse
	if err := json.NewDecoder(incomingResp.Body).Decode(&incomingPayments); err != nil {
		return nil, err
	}
	transactions = []lnclient.Transaction{}
	for _, invoice := range incomingPayments {
		transaction, err := phoenixInvoiceToTransaction(&invoice)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, *transaction)
	}

	// get outgoing payments
	outgoingQuery := url.Values{}
	if from != 0 {
		outgoingQuery.Add("from", strconv.FormatUint(from*1000, 10))
	}
	if until != 0 {
		outgoingQuery.Add("to", strconv.FormatUint(until*1000, 10))
	}
	if limit != 0 {
		outgoingQuery.Add("limit", strconv.FormatUint(limit, 10))
	}
	if offset != 0 {
		outgoingQuery.Add("offset", strconv.FormatUint(offset, 10))
	}
	outgoingQuery.Add("all", strconv.FormatBool(unpaid))

	outgoingUrl := svc.Address + "/payments/outgoing?" + outgoingQuery.Encode()

	logger.Logger.WithFields(logrus.Fields{
		"url": outgoingUrl,
	}).Infof("Fetching outgoing transactions: %s", outgoingUrl)
	outgoingReq, err := http.NewRequest(http.MethodGet, outgoingUrl, nil)
	if err != nil {
		return nil, err
	}
	outgoingReq.Header.Add("Authorization", "Basic "+svc.Authorization)
	outgoingResp, err := client.Do(outgoingReq)
	if err != nil {
		return nil, err
	}
	defer outgoingResp.Body.Close()

	var outgoingPayments []OutgoingPaymentResponse
	if err := json.NewDecoder(outgoingResp.Body).Decode(&outgoingPayments); err != nil {
		return nil, err
	}
	for _, invoice := range outgoingPayments {
		var settledAt *int64
		if invoice.CompletedAt != 0 {
			settledAtUnix := time.UnixMilli(invoice.CompletedAt).Unix()
			settledAt = &settledAtUnix
		}
		transaction := lnclient.Transaction{
			Type:        "outgoing",
			Invoice:     invoice.Invoice,
			Preimage:    invoice.Preimage,
			PaymentHash: invoice.PaymentHash,
			Amount:      invoice.Sent * 1000,
			FeesPaid:    invoice.Fees * 1000,
			CreatedAt:   time.UnixMilli(invoice.CreatedAt).Unix(),
			SettledAt:   settledAt,
		}
		transactions = append(transactions, transaction)
	}

	// sort by created date descending
	sort.SliceStable(transactions, func(i, j int) bool {
		return transactions[i].CreatedAt > transactions[j].CreatedAt
	})

	return transactions, nil
}

func (svc *PhoenixService) GetInfo(ctx context.Context) (info *lnclient.NodeInfo, err error) {
	return svc.nodeInfo, nil
}

func fetchNodeInfo(svc *PhoenixService) (info *lnclient.NodeInfo, err error) {
	req, err := http.NewRequest(http.MethodGet, svc.Address+"/getinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Basic "+svc.Authorization)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var infoRes InfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&infoRes); err != nil {
		return nil, err
	}
	return &lnclient.NodeInfo{
		Alias:       "Phoenix",
		Color:       "",
		Pubkey:      infoRes.NodeId,
		Network:     "bitcoin",
		BlockHeight: 0,
		BlockHash:   "",
	}, nil
}

func (svc *PhoenixService) ListChannels(ctx context.Context) ([]lnclient.Channel, error) {
	channels := []lnclient.Channel{}
	return channels, nil
}

func (svc *PhoenixService) MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64) (transaction *lnclient.Transaction, err error) {
	// TODO: support expiry
	if expiry == 0 {
		expiry = lnclient.DEFAULT_INVOICE_EXPIRY
	}
	form := url.Values{}
	amountSat := strconv.FormatInt(amount/1000, 10)
	form.Add("amountSat", amountSat)
	if description != "" {
		form.Add("description", description)
	} else if descriptionHash != "" {
		form.Add("descriptionHash", descriptionHash)
	} else {
		form.Add("description", "invoice")
	}

	today := time.Now().UTC().Format("2006-02-01") // querying is too slow so we limit the invoices we query with the date - see list transactions
	form.Add("externalId", today)                  // for some resone phoenixd requires an external id to query a list of invoices. thus we set this to nwc
	logger.Logger.WithFields(logrus.Fields{
		"externalId": today,
		"amountSat":  amountSat,
	}).Infof("Requesting phoenix invoice")
	req, err := http.NewRequest(http.MethodPost, svc.Address+"/createinvoice", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Basic "+svc.Authorization)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var invoiceRes MakeInvoiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&invoiceRes); err != nil {
		return nil, err
	}

	tx, err := svc.LookupInvoice(ctx, invoiceRes.PaymentHash)
	if err != nil {
		logger.Logger.WithError(err).Error("failed to lookup newly created invoice")
		return nil, err
	}

	return tx, nil
}

func (svc *PhoenixService) MakeHoldInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64, paymentHash string) (transaction *lnclient.Transaction, err error) {
	return nil, errors.New("not implemented")
}

func (svc *PhoenixService) SettleHoldInvoice(ctx context.Context, preimage string) (err error) {
	return errors.New("not implemented")
}

func (svc *PhoenixService) CancelHoldInvoice(ctx context.Context, paymentHash string) (err error) {
	return errors.New("not implemented")
}

func (svc *PhoenixService) LookupInvoice(ctx context.Context, paymentHash string) (transaction *lnclient.Transaction, err error) {
	req, err := http.NewRequest(http.MethodGet, svc.Address+"/payments/incoming/"+paymentHash, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Basic "+svc.Authorization)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var invoiceRes InvoiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&invoiceRes); err != nil {
		return nil, err
	}

	transaction, err = phoenixInvoiceToTransaction(&invoiceRes)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (svc *PhoenixService) SendPaymentSync(ctx context.Context, payReq string, amount *uint64, timeoutSeconds *int64) (*lnclient.PayInvoiceResponse, error) {
	// TODO: support 0-amount invoices
	if amount != nil {
		return nil, errors.New("0-amount invoices not supported")
	}
	form := url.Values{}
	form.Add("invoice", payReq)
	req, err := http.NewRequest(http.MethodPost, svc.Address+"/payinvoice", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Basic "+svc.Authorization)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payRes PayResponse
	if err := json.NewDecoder(resp.Body).Decode(&payRes); err != nil {
		return nil, err
	}

	return &lnclient.PayInvoiceResponse{
		Preimage: payRes.PaymentPreimage,
		Fee:      uint64(payRes.RoutingFeeSat) * 1000,
	}, nil
}

func (svc *PhoenixService) SendKeysend(ctx context.Context, amount uint64, destination string, custom_records []lnclient.TLVRecord, preimage string) (*lnclient.PayKeysendResponse, error) {
	return nil, errors.New("not implemented")
}

func (svc *PhoenixService) RedeemOnchainFunds(ctx context.Context, toAddress string, amount uint64, feeRate *uint64, sendAll bool) (txId string, err error) {
	return "", errors.New("not implemented")
}

func (svc *PhoenixService) ResetRouter(key string) error {
	return nil
}

func (svc *PhoenixService) Shutdown() error {
	// No specific shutdown actions needed for Phoenixd client via HTTP
	return nil
}

func (svc *PhoenixService) GetNodeConnectionInfo(ctx context.Context) (nodeConnectionInfo *lnclient.NodeConnectionInfo, err error) {
	req, err := http.NewRequest(http.MethodGet, svc.Address+"/getinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Basic "+svc.Authorization)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var infoRes InfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&infoRes); err != nil {
		return nil, err
	}
	return &lnclient.NodeConnectionInfo{
		Pubkey: infoRes.NodeId,
	}, nil
}

func (svc *PhoenixService) ConnectPeer(ctx context.Context, connectPeerRequest *lnclient.ConnectPeerRequest) error {
	return nil
}
func (svc *PhoenixService) OpenChannel(ctx context.Context, openChannelRequest *lnclient.OpenChannelRequest) (*lnclient.OpenChannelResponse, error) {
	return nil, nil
}

func (svc *PhoenixService) CloseChannel(ctx context.Context, closeChannelRequest *lnclient.CloseChannelRequest) (*lnclient.CloseChannelResponse, error) {
	return nil, nil
}

func (svc *PhoenixService) GetNewOnchainAddress(ctx context.Context) (string, error) {
	return "", nil
}

func (svc *PhoenixService) GetOnchainBalance(ctx context.Context) (*lnclient.OnchainBalanceResponse, error) {
	return nil, errors.New("not implemented")
}

func (svc *PhoenixService) SignMessage(ctx context.Context, message string) (string, error) {
	return "", errors.New("not implemented")
}

func (svc *PhoenixService) SendPaymentProbes(ctx context.Context, invoice string) error {
	return nil
}

func (svc *PhoenixService) SendSpontaneousPaymentProbes(ctx context.Context, amountMsat uint64, nodeId string) error {
	return nil
}

func (svc *PhoenixService) ListPeers(ctx context.Context) ([]lnclient.PeerDetails, error) {
	return nil, nil
}

func (svc *PhoenixService) GetLogOutput(ctx context.Context, maxLen int) ([]byte, error) {
	return []byte{}, nil
}

func (svc *PhoenixService) GetNodeStatus(ctx context.Context) (nodeStatus *lnclient.NodeStatus, err error) {
	_, err = fetchNodeInfo(svc)
	if err != nil {
		return nil, err
	}

	return &lnclient.NodeStatus{
		IsReady: true,
	}, nil
}

func (svc *PhoenixService) GetStorageDir() (string, error) {
	return "", nil
}

func (svc *PhoenixService) GetNetworkGraph(ctx context.Context, nodeIds []string) (lnclient.NetworkGraphResponse, error) {
	return nil, nil
}

func (svc *PhoenixService) UpdateLastWalletSyncRequest() {}

func (svc *PhoenixService) DisconnectPeer(ctx context.Context, peerId string) error {
	return nil
}

func (svc *PhoenixService) UpdateChannel(ctx context.Context, updateChannelRequest *lnclient.UpdateChannelRequest) error {
	return nil
}

func (svc *PhoenixService) GetSupportedNIP47Methods() []string {
	return []string{"pay_invoice", "get_balance", "get_budget", "get_info", "make_invoice", "lookup_invoice", "list_transactions", "multi_pay_invoice"}
}

func (svc *PhoenixService) GetSupportedNIP47NotificationTypes() []string {
	return []string{}
}

func (svc *PhoenixService) GetPubkey() string {
	return svc.pubkey
}

func phoenixInvoiceToTransaction(invoiceRes *InvoiceResponse) (*lnclient.Transaction, error) {
	var settledAt *int64
	if invoiceRes.CompletedAt != 0 {
		settledAtUnix := time.UnixMilli(invoiceRes.CompletedAt).Unix()
		settledAt = &settledAtUnix
	}

	paymentRequest, err := decodepay.Decodepay(invoiceRes.Invoice)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"bolt11": invoiceRes.Invoice,
		}).Errorf("Failed to decode bolt11 invoice: %v", err)

		return nil, err
	}

	expiresAt := time.UnixMilli(int64(paymentRequest.CreatedAt) * 1000).Add(time.Duration(paymentRequest.Expiry) * time.Second).Unix()

	return &lnclient.Transaction{
		Type:            "incoming",
		Invoice:         invoiceRes.Invoice,
		Preimage:        invoiceRes.Preimage,
		PaymentHash:     invoiceRes.PaymentHash,
		Amount:          paymentRequest.MSatoshi,
		FeesPaid:        invoiceRes.Fees * 1000,
		CreatedAt:       time.UnixMilli(invoiceRes.CreatedAt).Unix(),
		Description:     invoiceRes.Description,
		SettledAt:       settledAt,
		ExpiresAt:       &expiresAt,
		DescriptionHash: paymentRequest.DescriptionHash,
	}, nil
}

func (svc *PhoenixService) GetCustomNodeCommandDefinitions() []lnclient.CustomNodeCommandDef {
	return nil
}

func (svc *PhoenixService) ExecuteCustomNodeCommand(ctx context.Context, command *lnclient.CustomNodeCommandRequest) (*lnclient.CustomNodeCommandResponse, error) {
	return nil, nil
}

func (svc *PhoenixService) MakeOffer(ctx context.Context, description string) (string, error) {
	return "", errors.New("not supported")
}

func (svc *PhoenixService) ListOnchainTransactions(ctx context.Context) ([]lnclient.OnchainTransaction, error) {
	return nil, errors.ErrUnsupported
}
