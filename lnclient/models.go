package lnclient

import (
	"context"
	"errors"
)

// TODO: remove JSON tags from these models (LNClient models should not be exposed directly)

// TLVRecord represents a Type-Length-Value record used in Lightning keysend payments.
type TLVRecord struct {
	Type uint64 `json:"type"`
	// hex-encoded value
	Value string `json:"value"`
}

// Metadata is a map of arbitrary key-value pairs associated with transactions.
type Metadata = map[string]interface{}

// NodeInfo contains basic information about the Lightning node.
type NodeInfo struct {
	Alias       string
	Color       string
	Pubkey      string
	Network     string
	BlockHeight uint32
	BlockHash   string
}

// Transaction represents a Lightning transaction (payment or invoice).
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
	SettleDeadline  *uint32 // block number for accepted hold invoices
}

// OnchainTransaction represents a Bitcoin onchain transaction.
type OnchainTransaction struct {
	AmountSat        uint64 `json:"amountSat"`
	CreatedAt        uint64 `json:"createdAt"`
	State            string `json:"state"`
	Type             string `json:"type"`
	NumConfirmations uint32 `json:"numConfirmations"`
	TxId             string `json:"txId"`
}

// NodeConnectionInfo contains the connection details for a Lightning node, including optional Tor address.
type NodeConnectionInfo struct {
	Pubkey  string `json:"pubkey"`
	Address string `json:"address"`
	Port    int    `json:"port"`
}

// LNClient defines the interface that all Lightning node backend implementations must satisfy.
type LNClient interface {
	SendPaymentSync(payReq string, amount *uint64) (*PayInvoiceResponse, error)
	SendKeysend(amount uint64, destination string, customRecords []TLVRecord, preimage string) (*PayKeysendResponse, error)
	GetPubkey() string
	GetInfo(ctx context.Context) (info *NodeInfo, err error)
	MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64, throughNodePubkey *string) (transaction *Transaction, err error)
	MakeHoldInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64, paymentHash string) (transaction *Transaction, err error)
	SettleHoldInvoice(ctx context.Context, preimage string) (err error)
	CancelHoldInvoice(ctx context.Context, paymentHash string) (err error)
	LookupInvoice(ctx context.Context, paymentHash string) (transaction *Transaction, err error)
	ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaid bool, invoiceType string) (transactions []Transaction, err error)
	ListOnchainTransactions(ctx context.Context) ([]OnchainTransaction, error)
	Shutdown() error
	ListChannels(ctx context.Context) (channels []Channel, err error)
	GetNodeConnectionInfo(ctx context.Context) (nodeConnectionInfo *NodeConnectionInfo, err error)
	GetNodeStatus(ctx context.Context) (nodeStatus *NodeStatus, err error)
	ConnectPeer(ctx context.Context, connectPeerRequest *ConnectPeerRequest) error
	OpenChannel(ctx context.Context, openChannelRequest *OpenChannelRequest) (*OpenChannelResponse, error)
	CloseChannel(ctx context.Context, closeChannelRequest *CloseChannelRequest) (*CloseChannelResponse, error)
	UpdateChannel(ctx context.Context, updateChannelRequest *UpdateChannelRequest) error
	DisconnectPeer(ctx context.Context, peerId string) error
	MakeOffer(ctx context.Context, description string) (string, error)
	GetNewOnchainAddress(ctx context.Context) (string, error)
	ResetRouter(key string) error
	GetOnchainBalance(ctx context.Context) (*OnchainBalanceResponse, error)
	GetBalances(ctx context.Context, includeInactiveChannels bool) (*BalancesResponse, error)
	RedeemOnchainFunds(ctx context.Context, toAddress string, amount uint64, feeRate *uint64, sendAll bool) (txId string, err error)
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
	GetCustomNodeCommandDefinitions() []CustomNodeCommandDef
	ExecuteCustomNodeCommand(ctx context.Context, command *CustomNodeCommandRequest) (*CustomNodeCommandResponse, error)
}

// Channel represents a Lightning payment channel with its balances and configuration.
type Channel struct {
	LocalBalance                             int64
	LocalSpendableBalance                    int64
	RemoteBalance                            int64
	Id                                       string
	RemotePubkey                             string
	FundingTxId                              string
	FundingTxVout                            uint32
	Active                                   bool
	Public                                   bool
	InternalChannel                          interface{}
	Confirmations                            *uint32
	ConfirmationsRequired                    *uint32
	ForwardingFeeBaseMsat                    uint32
	ForwardingFeeProportionalMillionths      uint32
	UnspendablePunishmentReserve             uint64
	CounterpartyUnspendablePunishmentReserve uint64
	Error                                    *string
	IsOutbound                               bool
}

// NodeStatus indicates whether the Lightning node is ready and provides internal status details.
type NodeStatus struct {
	IsReady            bool        `json:"isReady"`
	InternalNodeStatus interface{} `json:"internalNodeStatus"`
}

// ConnectPeerRequest contains the pubkey and network address of a peer to connect to.
type ConnectPeerRequest struct {
	Pubkey  string `json:"pubkey"`
	Address string `json:"address"`
	Port    uint16 `json:"port"`
}

// OpenChannelRequest contains the parameters for opening a new Lightning channel.
type OpenChannelRequest struct {
	Pubkey     string `json:"pubkey"`
	AmountSats int64  `json:"amountSats"`
	Public     bool   `json:"public"`
}

// OpenChannelResponse contains the funding transaction ID of the newly opened channel.
type OpenChannelResponse struct {
	FundingTxId string `json:"fundingTxId"`
}

// CloseChannelRequest contains the parameters for closing a Lightning channel.
type CloseChannelRequest struct {
	ChannelId string `json:"channelId"`
	NodeId    string `json:"nodeId"`
	Force     bool   `json:"force"`
}

// UpdateChannelRequest contains the parameters for updating a channel's forwarding fees.
type UpdateChannelRequest struct {
	ChannelId                                string `json:"channelId"`
	NodeId                                   string `json:"nodeId"`
	ForwardingFeeBaseMsat                    uint32 `json:"forwardingFeeBaseMsat"`
	ForwardingFeeProportionalMillionths      uint32 `json:"forwardingFeeProportionalMillionths"`
	MaxDustHtlcExposureFromFeeRateMultiplier uint64 `json:"maxDustHtlcExposureFromFeeRateMultiplier"`
}

// CloseChannelResponse is returned after successfully closing a channel.
type CloseChannelResponse struct {
}

// PendingBalanceDetails contains details about funds pending from a channel closure.
type PendingBalanceDetails struct {
	ChannelId     string `json:"channelId"`
	NodeId        string `json:"nodeId"`
	Amount        uint64 `json:"amount"`
	FundingTxId   string `json:"fundingTxId"`
	FundingTxVout uint32 `json:"fundingTxVout"`
}

// OnchainBalanceResponse contains the onchain wallet balance breakdown.
type OnchainBalanceResponse struct {
	Spendable                          int64                   `json:"spendable"`
	Total                              int64                   `json:"total"`
	Reserved                           int64                   `json:"reserved"`
	PendingBalancesFromChannelClosures uint64                  `json:"pendingBalancesFromChannelClosures"`
	PendingBalancesDetails             []PendingBalanceDetails `json:"pendingBalancesDetails"`
	PendingSweepBalancesDetails        []PendingBalanceDetails `json:"pendingSweepBalancesDetails"`
	InternalBalances                   interface{}             `json:"internalBalances"`
}

// PeerDetails contains information about a connected Lightning peer.
type PeerDetails struct {
	NodeId      string `json:"nodeId"`
	Address     string `json:"address"`
	IsPersisted bool   `json:"isPersisted"`
	IsConnected bool   `json:"isConnected"`
}
// LightningBalanceResponse contains the Lightning channel balance breakdown.
type LightningBalanceResponse struct {
	TotalSpendable       int64 `json:"totalSpendable"`
	TotalReceivable      int64 `json:"totalReceivable"`
	NextMaxSpendable     int64 `json:"nextMaxSpendable"`
	NextMaxReceivable    int64 `json:"nextMaxReceivable"`
	NextMaxSpendableMPP  int64 `json:"nextMaxSpendableMPP"`
	NextMaxReceivableMPP int64 `json:"nextMaxReceivableMPP"`
}

// PayInvoiceResponse contains the preimage and fee for a paid invoice.
type PayInvoiceResponse struct {
	Preimage string `json:"preimage"`
	Fee      uint64 `json:"fee"`
}

// PayOfferResponse contains the preimage, fee, and payment hash for a paid BOLT12 offer.
type PayOfferResponse = struct {
	Preimage    string `json:"preimage"`
	Fee         uint64 `json:"fee"`
	PaymentHash string `json:"payment_hash"`
}

// PayKeysendResponse contains the fee for a keysend payment.
type PayKeysendResponse struct {
	Fee uint64 `json:"fee"`
}

// BalancesResponse contains both the onchain and Lightning channel balances.
type BalancesResponse struct {
	Onchain   OnchainBalanceResponse   `json:"onchain"`
	Lightning LightningBalanceResponse `json:"lightning"`
}

// NetworkGraphResponse is an alias for the network graph data returned by the node.
type NetworkGraphResponse = interface{}

// PaymentFailedEventProperties contains the transaction and reason for a failed payment event.
type PaymentFailedEventProperties struct {
	Transaction *Transaction
	Reason      string
}

// PaymentForwardedEventProperties contains the forwarding statistics for a forwarded payment event.
type PaymentForwardedEventProperties struct {
	TotalFeeEarnedMsat          uint64
	OutboundAmountForwardedMsat uint64
}

// CustomNodeCommandArgDef defines a single argument for a custom node command.
type CustomNodeCommandArgDef struct {
	Name        string
	Description string
}

// CustomNodeCommandDef defines a custom node command with its name, description, and arguments.
type CustomNodeCommandDef struct {
	Name        string
	Description string
	Args        []CustomNodeCommandArgDef
}

// CustomNodeCommandArg represents a name-value pair argument for a custom node command.
type CustomNodeCommandArg struct {
	Name  string
	Value string
}

// CustomNodeCommandRequest contains the command name and arguments to execute.
type CustomNodeCommandRequest struct {
	Name string
	Args []CustomNodeCommandArg
}

// CustomNodeCommandResponse contains the response from executing a custom node command.
type CustomNodeCommandResponse struct {
	Response interface{}
}

// NewCustomNodeCommandResponseEmpty creates a CustomNodeCommandResponse with an empty struct response.
func NewCustomNodeCommandResponseEmpty() *CustomNodeCommandResponse {
	return &CustomNodeCommandResponse{
		Response: struct{}{},
	}
}

// ErrUnknownCustomNodeCommand is returned when an unrecognized custom node command is requested.
var ErrUnknownCustomNodeCommand = errors.New("unknown custom node command")

// DEFAULT_INVOICE_EXPIRY is the default invoice expiry in seconds (1 day).
const DEFAULT_INVOICE_EXPIRY = 86400

type holdInvoiceCanceledError struct {
}

// NewHoldInvoiceCanceledError returns an error indicating that a hold invoice was canceled.
func NewHoldInvoiceCanceledError() error {
	return &holdInvoiceCanceledError{}
}

func (err *holdInvoiceCanceledError) Error() string {
	return "Hold invoice canceled"
}
