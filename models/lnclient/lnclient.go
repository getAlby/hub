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
	SendPaymentSync(ctx context.Context, payReq string) (preimage string, err error)
	SendKeysend(ctx context.Context, amount int64, destination, preimage string, custom_records []TLVRecord) (preImage string, err error)
	GetBalance(ctx context.Context) (balance int64, err error)
	GetInfo(ctx context.Context) (info *NodeInfo, err error)
	MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64) (transaction *Transaction, err error)
	LookupInvoice(ctx context.Context, paymentHash string) (transaction *Transaction, err error)
	ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaid bool, invoiceType string) (transactions []Transaction, err error)
	Shutdown() error
	ListChannels(ctx context.Context) (channels []Channel, err error)
	GetNodeConnectionInfo(ctx context.Context) (nodeConnectionInfo *NodeConnectionInfo, err error)
	ConnectPeer(ctx context.Context, connectPeerRequest *ConnectPeerRequest) error
	OpenChannel(ctx context.Context, openChannelRequest *OpenChannelRequest) (*OpenChannelResponse, error)
	CloseChannel(ctx context.Context, closeChannelRequest *CloseChannelRequest) (*CloseChannelResponse, error)
	GetNewOnchainAddress(ctx context.Context) (string, error)
	GetOnchainBalance(ctx context.Context) (int64, error)
	SignMessage(ctx context.Context, message []byte) (string, error)
}

type Channel struct {
	LocalBalance  int64  `json:"localBalance"`
	RemoteBalance int64  `json:"remoteBalance"`
	Id            string `json:"id"`
	RemotePubkey  string `json:"remotePubkey"`
	Active        bool   `json:"active"`
	Public        bool   `json:"public"`
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
