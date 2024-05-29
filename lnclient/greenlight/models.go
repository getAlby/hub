package greenlight

import "github.com/getAlby/nostr-wallet-connect/lnclient"

type NodeInfo struct {
	ID          string `json:"id"`
	Alias       string `json:"alias"`
	Color       string `json:"color"`
	Network     string `json:"network"`
	BlockHeight uint32 `json:"blockheight"`
	// ...other fields
}

type Invoice struct {
	Bolt11             string     `json:"bolt11"`
	PaymentHash        string     `json:"payment_hash"`
	Preimage           string     `json:"payment_secret"`
	ExpiresAt          int64      `json:"expires_at"`
	PaidAt             *int64     `json:"paid_at"`
	Label              string     `json:"label"`
	Description        string     `json:"description"`
	AmountMsat         MsatValue  `json:"amount_msat"`
	AmountReceivedMsat *MsatValue `json:"amount_received_msat"`
	Status             int        `json:"status"`
	// ...other fields
}

type Payment struct {
	PaymentHash    string    `json:"payment_hash"`
	Status         int       `json:"status"`
	Destination    string    `json:"destination"`
	CreatedAt      int64     `json:"created_at"`
	CompletedAt    int64     `json:"completed_at"`
	Bolt11         string    `json:"bolt11"`
	AmountMsat     MsatValue `json:"amount_msat"`
	AmountSentMsat MsatValue `json:"amount_sent_msat"`
	Preimage       string    `json:"preimage"`
	// ...other fields
}

type MsatValue struct {
	Msat int64 `json:"msat"`
}

type PayResponse struct {
	PaymentHash    string    `json:"payment_hash"`
	Preimage       string    `json:"payment_preimage"`
	CreatedAt      float64   `json:"created_at"`
	AmountMsat     MsatValue `json:"amount_msat"`
	AmountSentMsat MsatValue `json:"amount_sent_msat"`
	// ...other fields
}

type ListFundsResponse struct {
	Channels []Channel `json:"channels"`
	Outputs  []Output  `json:"outputs"`
}

type Output struct {
	TxId       string    `json:"txid"`
	AmountMsat MsatValue `json:"amount_msat"`
	Address    string    `json:"address"`
	// ...other fields
}

type Channel struct {
	PeerId        string    `json:"peer_id"`
	OurAmountMsat MsatValue `json:"our_amount_msat"`
	AmountMsat    MsatValue `json:"amount_msat"`
	FundingTxId   string    `json:"funding_txid"`
	Id            string    `json:"channel_id"`
	State         int       `json:"state"`
}

type ConnectPeerResponse struct {
	Id string `json:"id"`
	// ...other fields
}

type Outpoint struct {
	//txid
}

type OpenChannelResponse struct {
	Tx        string `json:"tx"`
	TxId      string `json:"txid"`
	ChannelId string `json:"channel_id"`
	// ...other fields
}

type NewAddressResponse struct {
	Bech32 string `json:"bech32"`
}

type ListInvoicesResponse struct {
	Invoices []Invoice `json:"invoices"`
}

type ListPaymentsResponse struct {
	Payments []Payment `json:"pays"`
}

type GreenlightNodeCredentials struct {
	DeviceCert []byte `json:"deviceCert"`
	DeviceKey  []byte `json:"deviceKey"`
	Seed       []byte `json:"seed"`
}

// EXAMPLE ONLY: ideally the rust bindings generate this interface
type GreenlightRustInterface interface {
	NewGreenlightService() *GreenlightRustService
}

type GreenlightRustService interface {
	/**
	Attempts to recover or register a new greenlight node and returns the credentials of the node.
	The invite code is only needed for registration.
	*/
	CreateCredentials(mnemonic string, inviteCode string) (*GreenlightNodeCredentials, error)
	/**
	Starts the Greenlight node hsmd process
	*/
	Start(*GreenlightNodeCredentials) error

	/**
	Stops the Greenlight node hsmd process
	*/
	Shutdown() error

	/* Methods required for Greenlight LNClient - not a 1:1 mapping */
	// TODO: remove usages of lnclient.* (not a 1:1 mapping there)
	SendPaymentSync(payReq string) (preimage string, err error)
	ListChannels() ([]Channel, error)
	ListFunds() (*ListFundsResponse, error)
	MakeInvoice(amount int64, description string, descriptionHash string, expiry int64) (transaction *lnclient.Transaction, err error)
	GetInfo() (info *NodeInfo, err error)
	GetNodeConnectionInfo() (nodeConnectionInfo *lnclient.NodeConnectionInfo, err error)
	ConnectPeer(connectPeerRequest *lnclient.ConnectPeerRequest) error
	OpenChannel(openChannelRequest *lnclient.OpenChannelRequest) (*OpenChannelResponse, error)
	GetNewOnchainAddress() (string, error)
	ListInvoices() ([]Invoice, error) // TODO: add paging params
	ListPayments() ([]Payment, error) // TODO: add paging params
	SendKeysend(amount int64, destination, preimage string, custom_records []lnclient.TLVRecord) (preImage string, err error)
	// TODO:
	// LookupInvoice(paymentHash string) (transaction *lnclient.Transaction, err error)
	// GetRoute
	// ListClosedChannels
	// ListHTLCs
	// ListPeers
	// SignMessage
	// ?ListOnchainTransactions
}
