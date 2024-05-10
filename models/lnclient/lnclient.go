// TODO: move to lnclient/models.go
package lnclient

import (
	"context"
)

type TLVRecord struct {
	Type  uint64 `json:"type"`
	Value string `json:"value"`
}

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
	Type            string      `json:"type"`
	Invoice         string      `json:"invoice"`
	Description     string      `json:"description"`
	DescriptionHash string      `json:"description_hash"`
	Preimage        string      `json:"preimage"`
	PaymentHash     string      `json:"payment_hash"`
	Amount          int64       `json:"amount"`
	FeesPaid        int64       `json:"fees_paid"`
	CreatedAt       int64       `json:"created_at"`
	ExpiresAt       *int64      `json:"expires_at"`
	SettledAt       *int64      `json:"settled_at"`
	Metadata        interface{} `json:"metadata,omitempty"`
}

type NodeConnectionInfo struct {
	Pubkey  string `json:"pubkey"`
	Address string `json:"address"`
	Port    int    `json:"port"`
}

type LNClient interface {
	SendPaymentSync(ctx context.Context, payReq string) (*Nip47PayInvoiceResponse, error)
	SendKeysend(ctx context.Context, amount int64, destination, preimage string, customRecords []TLVRecord) (preImage string, err error)
	GetBalance(ctx context.Context) (balance int64, err error)
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
	GetNewOnchainAddress(ctx context.Context) (string, error)
	ResetRouter(ctx context.Context, key string) error
	GetOnchainBalance(ctx context.Context) (*OnchainBalanceResponse, error)
	GetBalances(ctx context.Context) (*BalancesResponse, error)
	RedeemOnchainFunds(ctx context.Context, toAddress string) (txId string, err error)
	SendPaymentProbes(ctx context.Context, invoice string) error
	SendSpontaneousPaymentProbes(ctx context.Context, amountMsat uint64, nodeId string) error
	ListPeers(ctx context.Context) ([]PeerDetails, error)
	GetLogOutput(ctx context.Context, maxLen int) ([]byte, error)
	SignMessage(ctx context.Context, message string) (string, error)
	GetStorageDir() (string, error)
}

type Channel struct {
	LocalBalance          int64       `json:"localBalance"`
	RemoteBalance         int64       `json:"remoteBalance"`
	Id                    string      `json:"id"`
	RemotePubkey          string      `json:"remotePubkey"`
	FundingTxId           string      `json:"fundingTxId"`
	Active                bool        `json:"active"`
	Public                bool        `json:"public"`
	InternalChannel       interface{} `json:"internalChannel"`
	Confirmations         *uint32     `json:"confirmations"`
	ConfirmationsRequired *uint32     `json:"confirmationsRequired"`
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

type Nip47PayInvoiceResponse struct {
	Preimage string  `json:"preimage"`
	Fee      *uint64 `json:"fee"`
}

type BalancesResponse struct {
	Onchain   OnchainBalanceResponse   `json:"onchain"`
	Lightning LightningBalanceResponse `json:"lightning"`
}
