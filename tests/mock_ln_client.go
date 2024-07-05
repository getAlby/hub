package tests

import (
	"context"
	"time"

	"github.com/getAlby/hub/lnclient"
)

// for the invoice:
// lnbcrt5u1pjuywzppp5h69dt59cypca2wxu69sw8ga0g39a3yx7dqug5nthrw3rcqgfdu4qdqqcqzzsxqyz5vqsp5gzlpzszyj2k30qmpme7jsfzr24wqlvt9xdmr7ay34lfelz050krs9qyyssq038x07nh8yuv8hdpjh5y8kqp7zcd62ql9na9xh7pla44htjyy02sz23q7qm2tza6ct4ypljk54w9k9qsrsu95usk8ce726ytep6vhhsq9mhf9a
const MockPaymentHash500 = "be8ad5d0b82071d538dcd160e3a3af444bd890de68388a4d771ba23c01096f2a"

const MockInvoice = "lntb1230n1pjypux0pp5xgxzcks5jtx06k784f9dndjh664wc08ucrganpqn52d0ftrh9n8sdqyw3jscqzpgxqyz5vqsp5rkx7cq252p3frx8ytjpzc55rkgyx2mfkzzraa272dqvr2j6leurs9qyyssqhutxa24r5hqxstchz5fxlslawprqjnarjujp5sm3xj7ex73s32sn54fthv2aqlhp76qmvrlvxppx9skd3r5ut5xutgrup8zuc6ay73gqmra29m"
const MockPaymentHash = "320c2c5a1492ccfd5bc7aa4ad9b657d6aaec3cfcc0d1d98413a29af4ac772ccf" // for the above invoice

var MockNodeInfo = lnclient.NodeInfo{
	Alias:       "bob",
	Color:       "#3399FF",
	Pubkey:      "123pubkey",
	Network:     "testnet",
	BlockHeight: 12,
	BlockHash:   "123blockhash",
}

var MockTime = time.Unix(1693876963, 0)
var MockTimeUnix = MockTime.Unix()

var MockTransactions = []lnclient.Transaction{
	{
		Type:            "incoming",
		Invoice:         MockInvoice,
		Description:     "mock invoice 1",
		DescriptionHash: "hash1",
		Preimage:        "preimage1",
		PaymentHash:     "payment_hash_1",
		Amount:          1000,
		FeesPaid:        50,
		SettledAt:       &MockTimeUnix,
		Metadata: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
	},
	{
		Type:            "incoming",
		Invoice:         MockInvoice,
		Description:     "mock invoice 2",
		DescriptionHash: "hash2",
		Preimage:        "preimage2",
		PaymentHash:     "payment_hash_2",
		Amount:          2000,
		FeesPaid:        75,
		SettledAt:       &MockTimeUnix,
	},
}
var MockTransaction = &MockTransactions[0]

type MockLn struct {
}

func NewMockLn() (*MockLn, error) {
	return &MockLn{}, nil
}

func (mln *MockLn) SendPaymentSync(ctx context.Context, payReq string) (*lnclient.PayInvoiceResponse, error) {
	return &lnclient.PayInvoiceResponse{
		Preimage: "123preimage",
	}, nil
}

func (mln *MockLn) SendKeysend(ctx context.Context, amount uint64, destination, preimage string, custom_records []lnclient.TLVRecord) (preImage string, err error) {
	return "12345preimage", nil
}

func (mln *MockLn) GetBalance(ctx context.Context) (balance int64, err error) {
	return 21000, nil
}

func (mln *MockLn) GetInfo(ctx context.Context) (info *lnclient.NodeInfo, err error) {
	return &MockNodeInfo, nil
}

func (mln *MockLn) MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64) (transaction *lnclient.Transaction, err error) {
	return MockTransaction, nil
}

func (mln *MockLn) LookupInvoice(ctx context.Context, paymentHash string) (transaction *lnclient.Transaction, err error) {
	return MockTransaction, nil
}

func (mln *MockLn) ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaid bool, invoiceType string) (invoices []lnclient.Transaction, err error) {
	return MockTransactions, nil
}
func (mln *MockLn) Shutdown() error {
	return nil
}

func (mln *MockLn) ListChannels(ctx context.Context) (channels []lnclient.Channel, err error) {
	return []lnclient.Channel{}, nil
}
func (mln *MockLn) GetNodeConnectionInfo(ctx context.Context) (nodeConnectionInfo *lnclient.NodeConnectionInfo, err error) {
	return nil, nil
}
func (mln *MockLn) ConnectPeer(ctx context.Context, connectPeerRequest *lnclient.ConnectPeerRequest) error {
	return nil
}
func (mln *MockLn) OpenChannel(ctx context.Context, openChannelRequest *lnclient.OpenChannelRequest) (*lnclient.OpenChannelResponse, error) {
	return nil, nil
}
func (mln *MockLn) CloseChannel(ctx context.Context, closeChannelRequest *lnclient.CloseChannelRequest) (*lnclient.CloseChannelResponse, error) {
	return nil, nil
}
func (mln *MockLn) GetNewOnchainAddress(ctx context.Context) (string, error) {
	return "", nil
}
func (mln *MockLn) GetBalances(ctx context.Context) (*lnclient.BalancesResponse, error) {
	return nil, nil
}
func (mln *MockLn) GetOnchainBalance(ctx context.Context) (*lnclient.OnchainBalanceResponse, error) {
	return nil, nil
}
func (mln *MockLn) RedeemOnchainFunds(ctx context.Context, toAddress string) (txId string, err error) {
	return "", nil
}
func (mln *MockLn) ResetRouter(key string) error {
	return nil
}
func (mln *MockLn) SendPaymentProbes(ctx context.Context, invoice string) error {
	return nil
}
func (mln *MockLn) SendSpontaneousPaymentProbes(ctx context.Context, amountMsat uint64, nodeId string) error {
	return nil
}
func (mln *MockLn) ListPeers(ctx context.Context) ([]lnclient.PeerDetails, error) {
	return nil, nil
}
func (mln *MockLn) GetLogOutput(ctx context.Context, maxLen int) ([]byte, error) {
	return []byte{}, nil
}
func (mln *MockLn) SignMessage(ctx context.Context, message string) (string, error) {
	return "", nil
}
func (mln *MockLn) GetStorageDir() (string, error) {
	return "", nil
}
func (mln *MockLn) GetNodeStatus(ctx context.Context) (nodeStatus *lnclient.NodeStatus, err error) {
	return nil, nil
}
func (mln *MockLn) GetNetworkGraph(nodeIds []string) (lnclient.NetworkGraphResponse, error) {
	return nil, nil
}

func (mln *MockLn) UpdateLastWalletSyncRequest() {}

func (mln *MockLn) DisconnectPeer(ctx context.Context, peerId string) error {
	return nil
}

func (mln *MockLn) UpdateChannel(ctx context.Context, updateChannelRequest *lnclient.UpdateChannelRequest) error {
	return nil
}

func (mln *MockLn) GetSupportedNIP47Methods() []string {
	return []string{"pay_invoice", "pay_keysend", "get_balance", "get_info", "make_invoice", "lookup_invoice", "list_transactions", "multi_pay_invoice", "multi_pay_keysend", "sign_message"}
}
func (mln *MockLn) GetSupportedNIP47NotificationTypes() []string {
	return []string{"payment_received", "payment_sent"}
}
