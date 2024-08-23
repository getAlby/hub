package lnclient

import (
	"context"
)

// TODO: remove JSON tags from these models (LNClient models should not be exposed directly)

type TLVRecord struct {
	Type uint64 `json:"type"`
	// hex-encoded value
	Value string `json:"value"`
}

type Metadata = map[string]interface{}

type NodeInfo struct {
	Alias       string
	Color       string
	Pubkey      string
	Network     string
	BlockHeight uint32
	BlockHash   string
}

// TODO: use uint for fields that cannot be negative
type Transaction struct {
	Type            string
	Invoice         string
	Description     string
	DescriptionHash string
	Preimage        string
	PaymentHash     string
	Amount          int64
	FeesPaid        int64
	CreatedAt       int64
	ExpiresAt       *int64
	SettledAt       *int64
	Metadata        Metadata
}

type NodeConnectionInfo struct {
	Pubkey  string `json:"pubkey"`
	Address string `json:"address"`
	Port    int    `json:"port"`
}

type LNClient interface {
	SendPaymentSync(ctx context.Context, payReq string) (*PayInvoiceResponse, error)
	SendKeysend(ctx context.Context, amount uint64, destination string, customRecords []TLVRecord, preimage string) (*PayKeysendResponse, error)
	GetBalance(ctx context.Context) (balance int64, err error)
	GetPubkey() string
	GetInfo(ctx context.Context) (info *NodeInfo, err error)
	MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64) (transaction *Transaction, err error)
	LookupInvoice(ctx context.Context, paymentHash string) (transaction *Transaction, err error)
	ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaid bool, invoiceType string) (transactions []Transaction, err error)
	Shutdown() error
	ListChannels(ctx context.Context) (channels []Channel, err error)
	GetNodeConnectionInfo(ctx context.Context) (nodeConnectionInfo *NodeConnectionInfo, err error)
	GetNodeStatus(ctx context.Context) (nodeStatus *NodeStatus, err error)
	ConnectPeer(ctx context.Context, connectPeerRequest *ConnectPeerRequest) error
	OpenChannel(ctx context.Context, openChannelRequest *OpenChannelRequest) (*OpenChannelResponse, error)
	CloseChannel(ctx context.Context, closeChannelRequest *CloseChannelRequest) (*CloseChannelResponse, error)
	UpdateChannel(ctx context.Context, updateChannelRequest *UpdateChannelRequest) error
	DisconnectPeer(ctx context.Context, peerId string) error
	GetNewOnchainAddress(ctx context.Context) (string, error)
	ResetRouter(key string) error
	GetOnchainBalance(ctx context.Context) (*OnchainBalanceResponse, error)
	GetBalances(ctx context.Context) (*BalancesResponse, error)
	RedeemOnchainFunds(ctx context.Context, toAddress string) (txId string, err error)
	SendPaymentProbes(ctx context.Context, invoice string) error
	SendSpontaneousPaymentProbes(ctx context.Context, amountMsat uint64, nodeId string) error
	ListPeers(ctx context.Context) ([]PeerDetails, error)
	GetLogOutput(ctx context.Context, maxLen int) ([]byte, error)
	SignMessage(ctx context.Context, message string) (string, error)
	GetStorageDir() (string, error)
	GetNetworkGraph(ctx context.Context, nodeIds []string) (NetworkGraphResponse, error)
	UpdateLastWalletSyncRequest()
	GetSupportedNIP47Methods() []string
	GetSupportedNIP47NotificationTypes() []string
}

type Channel struct {
	LocalBalance                             int64
	LocalSpendableBalance                    int64
	RemoteBalance                            int64
	Id                                       string
	RemotePubkey                             string
	FundingTxId                              string
	Active                                   bool
	Public                                   bool
	InternalChannel                          interface{}
	Confirmations                            *uint32
	ConfirmationsRequired                    *uint32
	ForwardingFeeBaseMsat                    uint32
	UnspendablePunishmentReserve             uint64
	CounterpartyUnspendablePunishmentReserve uint64
	Error                                    *string
	IsOutbound                               bool
}

type NodeStatus struct {
	InternalNodeStatus interface{} `json:"internalNodeStatus"`
}

type ConnectPeerRequest struct {
	Pubkey  string `json:"pubkey"`
	Address string `json:"address"`
	Port    uint16 `json:"port"`
}

type OpenChannelRequest struct {
	Pubkey string `json:"pubkey"`
	Amount int64  `json:"amount"`
	Public bool   `json:"public"`
}

type OpenChannelResponse struct {
	FundingTxId string `json:"fundingTxId"`
}

type CloseChannelRequest struct {
	ChannelId string `json:"channelId"`
	NodeId    string `json:"nodeId"`
	Force     bool   `json:"force"`
}

type UpdateChannelRequest struct {
	ChannelId                                string `json:"channelId"`
	NodeId                                   string `json:"nodeId"`
	ForwardingFeeBaseMsat                    uint32 `json:"forwardingFeeBaseMsat"`
	MaxDustHtlcExposureFromFeeRateMultiplier uint64 `json:"maxDustHtlcExposureFromFeeRateMultiplier"`
}

type CloseChannelResponse struct {
}

type OnchainBalanceResponse struct {
	Spendable int64 `json:"spendable"`
	Total     int64 `json:"total"`
	Reserved  int64 `json:"reserved"`
}

type PeerDetails struct {
	NodeId      string `json:"nodeId"`
	Address     string `json:"address"`
	IsPersisted bool   `json:"isPersisted"`
	IsConnected bool   `json:"isConnected"`
}
type LightningBalanceResponse struct {
	TotalSpendable       int64 `json:"totalSpendable"`
	TotalReceivable      int64 `json:"totalReceivable"`
	NextMaxSpendable     int64 `json:"nextMaxSpendable"`
	NextMaxReceivable    int64 `json:"nextMaxReceivable"`
	NextMaxSpendableMPP  int64 `json:"nextMaxSpendableMPP"`
	NextMaxReceivableMPP int64 `json:"nextMaxReceivableMPP"`
}

type PayInvoiceResponse struct {
	Preimage string `json:"preimage"`
	Fee      uint64 `json:"fee"`
}

type PayKeysendResponse struct {
	Fee uint64 `json:"fee"`
}

type BalancesResponse struct {
	Onchain   OnchainBalanceResponse   `json:"onchain"`
	Lightning LightningBalanceResponse `json:"lightning"`
}

type NetworkGraphResponse = interface{}

type PaymentFailedEventProperties struct {
	Transaction *Transaction
	Reason      string
}

// default invoice expiry in seconds (1 day)
const DEFAULT_INVOICE_EXPIRY = 86400

type timeoutError struct {
}

func NewTimeoutError() error {
	return &timeoutError{}
}

func (err *timeoutError) Error() string {
	return "Timeout"
}
