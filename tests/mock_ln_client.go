package tests

import (
	"context"
	"errors"
	"time"

	"github.com/getAlby/hub/lnclient"
)

// for the invoice:
// lnbcrt5u1pjuywzppp5h69dt59cypca2wxu69sw8ga0g39a3yx7dqug5nthrw3rcqgfdu4qdqqcqzzsxqyz5vqsp5gzlpzszyj2k30qmpme7jsfzr24wqlvt9xdmr7ay34lfelz050krs9qyyssq038x07nh8yuv8hdpjh5y8kqp7zcd62ql9na9xh7pla44htjyy02sz23q7qm2tza6ct4ypljk54w9k9qsrsu95usk8ce726ytep6vhhsq9mhf9a
const MockPaymentHash500 = "be8ad5d0b82071d538dcd160e3a3af444bd890de68388a4d771ba23c01096f2a"

const MockInvoice = "lntbs1230n1pnkqautdqyw3jsnp4q09a0z84kg4a2m38zjllw43h953fx5zvqe8qxfgw694ymkq26u8zcpp5yvnh6hsnlnj4xnuh2trzlnunx732dv8ta2wjr75pdfxf6p2vlyassp5hyeg97a3ft5u769kjwsn7p0e85h79pzz8kladmnqhpcypz2uawjs9qyysgqcqpcxq8zals8sq9yeg2pa9eywkgj50cyzxd5elatujuc0c0wh6j9nat5mn34pgk8u9ufpgs99tw9ldlfk42cqlkr48au3lmuh09269prg4qkggh4a8cyqpfl0y6j"
const MockPaymentHash = "23277d5e13fce5534f9752c62fcf9337a2a6b0ebea9d21fa816a4c9d054cf93b" // for the above invoice

const Mock0AmountInvoice = "lntbs1pnkjfgudqjd3hkueeqv4u8q6tj0ynp4qws83mqzuqptu5kfvxeles7qmyhsj6u2s6zyuft26mcr4tdmcupuupp533y9nwnsaktr9zlvyxmv97ta23faerygh3t9xvsfwytsr28lgggssp5mku3023z3kdxlpx6vrwtfxvvrxpffrquy6veex4ndk7rxhdtslhq9qyysgqcqpcxqxfvltyqva6y7k89jwtcljx399jl6wsq4lkq29vnm3rj4jxmapc6vcs358sx8mtpgh93rdc6ccqpxwwfga59zrla5m55zwzck2y2rsrxumu852sqkvpcm7"
const Mock0AmountPaymentHash = "8c4859ba70ed96328bec21b6c2f97d5453dc8c88bc56533209711701a8ff4211"

var MockNodeInfo = lnclient.NodeInfo{
	Alias:       "bob",
	Color:       "#3399FF",
	Pubkey:      "123pubkey",
	Network:     "testnet",
	BlockHeight: 12,
	BlockHash:   "123blockhash",
}

var MockLNClientBalances = lnclient.BalancesResponse{
	Lightning: lnclient.LightningBalanceResponse{
		TotalSpendable: 21000,
	},
}

var MockTime = time.Unix(1693876963, 0)
var MockTimeUnix = MockTime.Unix()

var MockLNClientTransactions = []lnclient.Transaction{
	{
		Type:            "incoming",
		Invoice:         MockInvoice,
		Description:     "mock invoice 1",
		DescriptionHash: "hash1",
		Preimage:        "preimage1",
		PaymentHash:     MockPaymentHash,
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
		PaymentHash:     MockPaymentHash,
		Amount:          2000,
		FeesPaid:        75,
		SettledAt:       &MockTimeUnix,
	},
}
var MockLNClientTransaction = &MockLNClientTransactions[0]

var MockLNClientHoldTransaction = &lnclient.Transaction{
	Type:            "incoming",
	Invoice:         "lntb10n1p5zg5p7dqud4hkx6eqdphkcepqd9h8vmmfvdjsnp4qw988hn4lhpu0my4rf0qkraft3wdx5aa0jnjmusgd23z3s0e9qv62pp5yaulxt6x83u4u0x2pck5pyg7fdxhjsd65c9lmu2a9r05qh0cgl6ssp5x57tsnnuc9hr99a9xzg5ylqma5fwvckxa50jqqay5zykqp83h9kq9qyysgqcqpcxqxf92hqqxh0avuskdnuzkk7mxslsdwem3qq3sf79a4ypmx0hax3rupp043yhv97h25vaarj0xrlcg2fdfdhpztsthettskyaylrz6vweztn0twqqv7kxmr",
	Description:     "mock hold invoice",
	DescriptionHash: "",
	Preimage:        "4aa083cad11038359b4f614f3a3d6a8298ae17d5275412bc3eca4f5f4d27f2d4",
	PaymentHash:     "2779f32f463c795e3cca0e2d40911e4b4d7941baa60bfdf15d28df405df847f5",
	Amount:          2000,
}

type MockLn struct {
	PayInvoiceResponses        []*lnclient.PayInvoiceResponse
	PayInvoiceErrors           []error
	Pubkey                     string
	MockTransaction            *lnclient.Transaction
	SupportedNotificationTypes *[]string
}

func NewMockLn() (*MockLn, error) {
	return &MockLn{}, nil
}

func (mln *MockLn) SendPaymentSync(ctx context.Context, payReq string, amount *uint64, timeoutSeconds *int64) (*lnclient.PayInvoiceResponse, error) {
	if len(mln.PayInvoiceResponses) > 0 {
		response := mln.PayInvoiceResponses[0]
		err := mln.PayInvoiceErrors[0]
		mln.PayInvoiceResponses = mln.PayInvoiceResponses[1:]
		mln.PayInvoiceErrors = mln.PayInvoiceErrors[1:]
		return response, err
	}

	return &lnclient.PayInvoiceResponse{
		Preimage: "123preimage",
	}, nil
}

func (mln *MockLn) SendKeysend(ctx context.Context, amount uint64, destination string, custom_records []lnclient.TLVRecord, preimage string) (*lnclient.PayKeysendResponse, error) {
	return &lnclient.PayKeysendResponse{
		Fee: 1,
	}, nil
}

func (mln *MockLn) GetInfo(ctx context.Context) (info *lnclient.NodeInfo, err error) {
	return &MockNodeInfo, nil
}

func (mln *MockLn) MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64) (transaction *lnclient.Transaction, err error) {
	return MockLNClientTransaction, nil
}

func (mln *MockLn) MakeHoldInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64, paymentHash string) (transaction *lnclient.Transaction, err error) {
	return MockLNClientHoldTransaction, nil
}

func (mln *MockLn) SettleHoldInvoice(ctx context.Context, preimage string) (err error) {
	return nil
}

func (mln *MockLn) CancelHoldInvoice(ctx context.Context, paymentHash string) (err error) {
	return nil
}

func (mln *MockLn) LookupInvoice(ctx context.Context, paymentHash string) (transaction *lnclient.Transaction, err error) {
	if mln.MockTransaction != nil {
		return mln.MockTransaction, nil
	}
	return MockLNClientTransaction, nil
}

func (mln *MockLn) ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaid bool, invoiceType string) (invoices []lnclient.Transaction, err error) {
	return MockLNClientTransactions, nil
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
func (mln *MockLn) GetBalances(ctx context.Context, includeInactiveChannels bool) (*lnclient.BalancesResponse, error) {
	return &MockLNClientBalances, nil
}
func (mln *MockLn) GetOnchainBalance(ctx context.Context) (*lnclient.OnchainBalanceResponse, error) {
	return nil, nil
}
func (mln *MockLn) RedeemOnchainFunds(ctx context.Context, toAddress string, amount uint64, feeRate *uint64, sendAll bool) (txId string, err error) {
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
	return &lnclient.NodeStatus{
		IsReady: true,
	}, nil
}
func (mln *MockLn) GetNetworkGraph(ctx context.Context, nodeIds []string) (lnclient.NetworkGraphResponse, error) {
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
	return []string{"pay_invoice", "pay_keysend", "get_balance", "get_budget", "get_info", "make_invoice", "lookup_invoice", "list_transactions", "multi_pay_invoice", "multi_pay_keysend", "sign_message"}
}
func (mln *MockLn) GetSupportedNIP47NotificationTypes() []string {
	if mln.SupportedNotificationTypes != nil {
		return *mln.SupportedNotificationTypes
	}

	return []string{"payment_received", "payment_sent"}
}
func (mln *MockLn) GetPubkey() string {
	if mln.Pubkey != "" {
		return mln.Pubkey
	}

	return "123pubkey"
}

func (mln *MockLn) GetCustomNodeCommandDefinitions() []lnclient.CustomNodeCommandDef {
	return nil
}

func (mln *MockLn) ExecuteCustomNodeCommand(ctx context.Context, command *lnclient.CustomNodeCommandRequest) (*lnclient.CustomNodeCommandResponse, error) {
	return nil, nil
}

func (mln *MockLn) MakeOffer(ctx context.Context, description string) (string, error) {
	return "", errors.New("not supported")
}

func (mln *MockLn) ListOnchainTransactions(ctx context.Context) ([]lnclient.OnchainTransaction, error) {
	return nil, errors.ErrUnsupported
}
