package lnclient

import (
	"context"
	"errors"
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
	AmountMsat      int64
	FeesPaidMsat    int64
	CreatedAt       int64
	ExpiresAt       *int64
	SettledAt       *int64
	Metadata        Metadata
	SettleDeadline  *uint32 // block number for accepted hold invoices
}

type OnchainTransaction struct {
	AmountSat        uint64 `json:"amountSat"`
	CreatedAt        uint64 `json:"createdAt"`
	State            string `json:"state"`
	Type             string `json:"type"`
	NumConfirmations uint32 `json:"numConfirmations"`
	TxId             string `json:"txId"`
}

type NodeConnectionInfo struct {
	Pubkey  string `json:"pubkey"`
	Address string `json:"address"`
	Port    int    `json:"port"`
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
	CloseChannel(ctx context.Context, closeChannelRequest *CloseChannelRequest) (*CloseChannelResponse, error)
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
	IsReady            bool        `json:"isReady"`
	InternalNodeStatus interface{} `json:"internalNodeStatus"`
}

type ConnectPeerRequest struct {
	Pubkey  string `json:"pubkey"`
	Address string `json:"address"`
	Port    uint16 `json:"port"`
}

type OpenChannelRequest struct {
	Pubkey     string `json:"pubkey"`
	AmountSats int64  `json:"amountSats"`
	Public     bool   `json:"public"`
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
	ForwardingFeeProportionalMillionths      uint32 `json:"forwardingFeeProportionalMillionths"`
	MaxDustHtlcExposureFromFeeRateMultiplier uint64 `json:"maxDustHtlcExposureFromFeeRateMultiplier"`
}

type CloseChannelResponse struct {
}

type PendingBalanceDetails struct {
	ChannelId     string `json:"channelId"`
	NodeId        string `json:"nodeId"`
	Amount        uint64 `json:"amount"` // deprecated
	AmountSat     uint64 `json:"amountSat"`
	FundingTxId   string `json:"fundingTxId"`
	FundingTxVout uint32 `json:"fundingTxVout"`
}

type OnchainBalanceResponse struct {
	Spendable                             int64                   `json:"spendable"` // deprecated
	SpendableSat                          int64                   `json:"spendableSat"`
	Total                                 int64                   `json:"total"` // deprecated
	TotalSat                              int64                   `json:"totalSat"`
	Reserved                              int64                   `json:"reserved"` // deprecated
	ReservedSat                           int64                   `json:"reservedSat"`
	PendingBalancesFromChannelClosures    uint64                  `json:"pendingBalancesFromChannelClosures"` // deprecated
	PendingBalancesFromChannelClosuresSat uint64                  `json:"pendingBalancesFromChannelClosuresSat"`
	PendingBalancesDetails                []PendingBalanceDetails `json:"pendingBalancesDetails"`
	PendingSweepBalancesDetails           []PendingBalanceDetails `json:"pendingSweepBalancesDetails"`
	InternalBalances                      interface{}             `json:"internalBalances"`
}

type PeerDetails struct {
	NodeId      string `json:"nodeId"`
	Address     string `json:"address"`
	IsPersisted bool   `json:"isPersisted"`
	IsConnected bool   `json:"isConnected"`
}
type LightningBalanceResponse struct {
	TotalSpendable           int64 `json:"totalSpendable"` // deprecated
	TotalSpendableSat        int64 `json:"totalSpendableSat"`
	TotalSpendableMsat       int64 `json:"totalSpendableMsat"`
	TotalReceivable          int64 `json:"totalReceivable"` // deprecated
	TotalReceivableSat       int64 `json:"totalReceivableSat"`
	TotalReceivableMsat      int64 `json:"totalReceivableMsat"`
	NextMaxSpendable         int64 `json:"nextMaxSpendable"` // deprecated
	NextMaxSpendableSat      int64 `json:"nextMaxSpendableSat"`
	NextMaxSpendableMsat     int64 `json:"nextMaxSpendableMsat"`
	NextMaxReceivable        int64 `json:"nextMaxReceivable"` // deprecated
	NextMaxReceivableSat     int64 `json:"nextMaxReceivableSat"`
	NextMaxReceivableMsat    int64 `json:"nextMaxReceivableMsat"`
	NextMaxSpendableMPP      int64 `json:"nextMaxSpendableMPP"` // deprecated
	NextMaxSpendableMPPSat   int64 `json:"nextMaxSpendableMPPSat"`
	NextMaxSpendableMPPMsat  int64 `json:"nextMaxSpendableMPPMsat"`
	NextMaxReceivableMPP     int64 `json:"nextMaxReceivableMPP"` // deprecated
	NextMaxReceivableMPPSat  int64 `json:"nextMaxReceivableMPPSat"`
	NextMaxReceivableMPPMsat int64 `json:"nextMaxReceivableMPPMsat"`
}

type PayInvoiceResponse struct {
	Preimage string `json:"preimage"`
	FeeMsat  uint64 `json:"feeMsat"`
}

type PayOfferResponse = struct {
	Preimage    string `json:"preimage"`
	FeeMsat     uint64 `json:"feeMsat"`
	PaymentHash string `json:"paymentHash"`
}

type PayKeysendResponse struct {
	FeeMsat uint64 `json:"feeMsat"`
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
