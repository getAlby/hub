package lnclient

import (
	"context"
	"errors"
)

// TLVRecord JSON tags are kept because values flow through the freeform
// transaction Metadata blob and are surfaced to NIP-47 clients via
// lookup_invoice / list_transactions.
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
	AmountMsat      int64
	FeesPaidMsat    int64
	CreatedAt       int64
	ExpiresAt       *int64
	SettledAt       *int64
	Metadata        Metadata
	SettleDeadline  *uint32 // block number for accepted hold invoices
}

type OnchainTransaction struct {
	AmountSat        uint64
	CreatedAt        uint64
	State            string
	Type             string
	NumConfirmations uint32
	TxId             string
}

type NodeConnectionInfo struct {
	Pubkey  string
	Address string
	Port    int
}

type LNClient interface {
	SendPaymentSync(payReq string, amountMsat *uint64) (*PayInvoiceResponse, error)
	SendKeysend(amountMsat uint64, destination string, customRecords []TLVRecord, preimage string) (*PayKeysendResponse, error)
	GetPubkey() string
	GetInfo(ctx context.Context) (info *NodeInfo, err error)
	MakeInvoice(ctx context.Context, amountMsat int64, description string, descriptionHash string, expiry int64, throughNodePubkey *string) (transaction *Transaction, err error)
	MakeHoldInvoice(ctx context.Context, amountMsat int64, description string, descriptionHash string, expiry int64, paymentHash string, minCltvExpiryDelta *uint64) (transaction *Transaction, err error)
	SettleHoldInvoice(ctx context.Context, preimage string) (err error)
	CancelHoldInvoice(ctx context.Context, paymentHash string) (err error)
	LookupInvoice(ctx context.Context, paymentHash string) (transaction *Transaction, err error)
	ListOnchainTransactions(ctx context.Context) ([]OnchainTransaction, error)
	Shutdown() error
	ListChannels(ctx context.Context) (channels []Channel, err error)
	GetNodeConnectionInfo(ctx context.Context) (nodeConnectionInfo *NodeConnectionInfo, err error)
	GetNodeStatus(ctx context.Context) (nodeStatus *NodeStatus, err error)
	ConnectPeer(ctx context.Context, connectPeerRequest *ConnectPeerRequest) error
	OpenChannel(ctx context.Context, openChannelRequest *OpenChannelRequest) (*OpenChannelResponse, error)
	CloseChannel(ctx context.Context, closeChannelRequest *CloseChannelRequest) error
	UpdateChannel(ctx context.Context, updateChannelRequest *UpdateChannelRequest) error
	DisconnectPeer(ctx context.Context, peerId string) error
	MakeOffer(ctx context.Context, description string) (string, error)
	GetNewOnchainAddress(ctx context.Context) (string, error)
	ResetRouter(key string) error
	GetOnchainBalance(ctx context.Context) (*OnchainBalanceResponse, error)
	GetBalances(ctx context.Context, includeInactiveChannels bool) (*BalancesResponse, error)
	RedeemOnchainFunds(ctx context.Context, toAddress string, amountSat uint64, feeRate *uint64, sendAll bool) (txId string, err error)
	ListPeers(ctx context.Context) ([]PeerDetails, error)
	GetLogOutput(ctx context.Context, maxLen int) ([]byte, error)
	SignMessage(ctx context.Context, message string) (string, error)
	GetStorageDir() (string, error)
	GetNetworkGraph(ctx context.Context, nodeIds []string) (NetworkGraphResponse, error)
	UpdateLastWalletSyncRequest()
	GetSupportedNIP47Methods() []string
	GetSupportedNIP47NotificationTypes() []string
	GetCustomNodeCommandDefinitions() []CustomNodeCommandDef
	ExecuteCustomNodeCommand(ctx context.Context, command *CustomNodeCommandRequest) (*CustomNodeCommandResponse, error)
}

type Channel struct {
	LocalBalanceMsat                            int64
	LocalSpendableBalanceMsat                   int64
	RemoteBalanceMsat                           int64
	Id                                          string
	RemotePubkey                                string
	FundingTxId                                 string
	FundingTxVout                               uint32
	Active                                      bool
	Public                                      bool
	InternalChannel                             interface{}
	Confirmations                               *uint32
	ConfirmationsRequired                       *uint32
	ForwardingFeeBaseMsat                       uint32
	ForwardingFeeProportionalMillionths         uint32
	UnspendablePunishmentReserveSat             uint64
	CounterpartyUnspendablePunishmentReserveSat uint64
	Error                                       *string
	IsOutbound                                  bool
}

type NodeStatus struct {
	IsReady            bool
	InternalNodeStatus interface{}
}

type ConnectPeerRequest struct {
	Pubkey  string
	Address string
	Port    uint16
}

type OpenChannelRequest struct {
	Pubkey     string
	AmountSats int64
	Public     bool
}

type OpenChannelResponse struct {
	FundingTxId string
}

type CloseChannelRequest struct {
	ChannelId string
	NodeId    string
	Force     bool
}

type UpdateChannelRequest struct {
	ChannelId                                string
	NodeId                                   string
	ForwardingFeeBaseMsat                    uint32
	ForwardingFeeProportionalMillionths      uint32
	MaxDustHtlcExposureFromFeeRateMultiplier uint64
}

type PendingBalanceDetails struct {
	ChannelId     string
	NodeId        string
	AmountSat     uint64
	FundingTxId   string
	FundingTxVout uint32
}

type OnchainBalanceResponse struct {
	SpendableSat                          int64
	TotalSat                              int64
	ReservedSat                           int64
	PendingBalancesFromChannelClosuresSat uint64
	PendingBalancesDetails                []PendingBalanceDetails
	PendingSweepBalancesDetails           []PendingBalanceDetails
	InternalBalances                      interface{}
}

type PeerDetails struct {
	NodeId      string
	Address     string
	IsPersisted bool
	IsConnected bool
}
type LightningBalanceResponse struct {
	TotalSpendableMsat       int64
	TotalReceivableMsat      int64
	NextMaxSpendableMsat     int64
	NextMaxReceivableMsat    int64
	NextMaxSpendableMPPMsat  int64
	NextMaxReceivableMPPMsat int64
}

type PayInvoiceResponse struct {
	Preimage string
	FeeMsat  uint64
}

type PayOfferResponse = struct {
	Preimage    string
	FeeMsat     uint64
	PaymentHash string
}

type PayKeysendResponse struct {
	FeeMsat uint64
}

type BalancesResponse struct {
	Onchain   OnchainBalanceResponse
	Lightning LightningBalanceResponse
}

type NetworkGraphResponse = interface{}

type PaymentFailedEventProperties struct {
	Transaction *Transaction
	Reason      string
}

type PaymentForwardedEventProperties struct {
	TotalFeeEarnedMsat          uint64
	OutboundAmountForwardedMsat uint64
}

type CustomNodeCommandArgDef struct {
	Name        string
	Description string
}

type CustomNodeCommandDef struct {
	Name        string
	Description string
	Args        []CustomNodeCommandArgDef
}

type CustomNodeCommandArg struct {
	Name  string
	Value string
}

type CustomNodeCommandRequest struct {
	Name string
	Args []CustomNodeCommandArg
}

type CustomNodeCommandResponse struct {
	Response interface{}
}

func NewCustomNodeCommandResponseEmpty() *CustomNodeCommandResponse {
	return &CustomNodeCommandResponse{
		Response: struct{}{},
	}
}

var ErrUnknownCustomNodeCommand = errors.New("unknown custom node command")

// default invoice expiry in seconds (1 day)
const DEFAULT_INVOICE_EXPIRY = 86400

type holdInvoiceCanceledError struct {
}

func NewHoldInvoiceCanceledError() error {
	return &holdInvoiceCanceledError{}
}

func (err *holdInvoiceCanceledError) Error() string {
	return "Hold invoice canceled"
}
