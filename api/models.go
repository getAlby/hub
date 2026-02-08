package api

import (
	"context"
	"io"
	"time"

	"github.com/getAlby/hub/alby"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/swaps"
)

// API defines the interface for all Alby Hub API operations including app management,
// channel operations, payments, and node administration.
type API interface {
	CreateApp(createAppRequest *CreateAppRequest) (*CreateAppResponse, error)
	UpdateApp(app *db.App, updateAppRequest *UpdateAppRequest) error
	Transfer(ctx context.Context, fromAppId *uint, toAppId *uint, amountMsat uint64) error
	DeleteApp(app *db.App) error
	GetApp(app *db.App) *App
	ListApps(limit uint64, offset uint64, filters ListAppsFilters, orderBy string) (*ListAppsResponse, error)
	CreateLightningAddress(ctx context.Context, createLightningAddressRequest *CreateLightningAddressRequest) error
	DeleteLightningAddress(ctx context.Context, appId uint) error
	ListChannels(ctx context.Context) ([]Channel, error)
	GetChannelPeerSuggestions(ctx context.Context) ([]alby.ChannelPeerSuggestion, error)
	GetLSPChannelOffer(ctx context.Context) (*alby.LSPChannelOffer, error)
	ResetRouter(key string) error
	ChangeUnlockPassword(changeUnlockPasswordRequest *ChangeUnlockPasswordRequest) error
	SetAutoUnlockPassword(unlockPassword string) error
	Stop() error
	GetNodeConnectionInfo(ctx context.Context) (*lnclient.NodeConnectionInfo, error)
	GetNodeStatus(ctx context.Context) (*lnclient.NodeStatus, error)
	ListPeers(ctx context.Context) ([]lnclient.PeerDetails, error)
	ConnectPeer(ctx context.Context, connectPeerRequest *ConnectPeerRequest) error
	DisconnectPeer(ctx context.Context, peerId string) error
	OpenChannel(ctx context.Context, openChannelRequest *OpenChannelRequest) (*OpenChannelResponse, error)
	RebalanceChannel(ctx context.Context, rebalanceChannelRequest *RebalanceChannelRequest) (*RebalanceChannelResponse, error)
	CloseChannel(ctx context.Context, peerId, channelId string, force bool) (*CloseChannelResponse, error)
	UpdateChannel(ctx context.Context, updateChannelRequest *UpdateChannelRequest) error
	MakeOffer(ctx context.Context, description string) (string, error)
	GetNewOnchainAddress(ctx context.Context) (string, error)
	GetUnusedOnchainAddress(ctx context.Context) (string, error)
	SignMessage(ctx context.Context, message string) (*SignMessageResponse, error)
	RedeemOnchainFunds(ctx context.Context, toAddress string, amount uint64, feeRate *uint64, sendAll bool) (*RedeemOnchainFundsResponse, error)
	GetBalances(ctx context.Context) (*BalancesResponse, error)
	ListTransactions(ctx context.Context, appId *uint, limit uint64, offset uint64) (*ListTransactionsResponse, error)
	ListOnchainTransactions(ctx context.Context) ([]lnclient.OnchainTransaction, error)
	SendPayment(ctx context.Context, invoice string, amountMsat *uint64, metadata map[string]interface{}) (*SendPaymentResponse, error)
	CreateInvoice(ctx context.Context, amount uint64, description string) (*MakeInvoiceResponse, error)
	LookupInvoice(ctx context.Context, paymentHash string) (*LookupInvoiceResponse, error)
	RequestMempoolApi(ctx context.Context, endpoint string) (interface{}, error)
	GetInfo(ctx context.Context) (*InfoResponse, error)
	GetMnemonic(unlockPassword string) (*MnemonicResponse, error)
	SetNextBackupReminder(backupReminderRequest *BackupReminderRequest) error
	Start(startRequest *StartRequest)
	Setup(ctx context.Context, setupRequest *SetupRequest) error
	SendPaymentProbes(ctx context.Context, sendPaymentProbesRequest *SendPaymentProbesRequest) (*SendPaymentProbesResponse, error)
	SendSpontaneousPaymentProbes(ctx context.Context, sendSpontaneousPaymentProbesRequest *SendSpontaneousPaymentProbesRequest) (*SendSpontaneousPaymentProbesResponse, error)
	GetNetworkGraph(ctx context.Context, nodeIds []string) (NetworkGraphResponse, error)
	SyncWallet() error
	GetLogOutput(ctx context.Context, logType string, getLogRequest *GetLogOutputRequest) (*GetLogOutputResponse, error)
	RequestLSPOrder(ctx context.Context, request *LSPOrderRequest) (*LSPOrderResponse, error)
	CreateBackup(unlockPassword string, w io.Writer) error
	RestoreBackup(unlockPassword string, r io.Reader) error
	MigrateNodeStorage(ctx context.Context, to string) error
	GetWalletCapabilities(ctx context.Context) (*WalletCapabilitiesResponse, error)
	Health(ctx context.Context) (*HealthResponse, error)
	SetCurrency(currency string) error
	SetBitcoinDisplayFormat(format string) error
	UpdateSettings(updateSettingsRequest *UpdateSettingsRequest) error
	LookupSwap(swapId string) (*LookupSwapResponse, error)
	ListSwaps() (*ListSwapsResponse, error)
	GetSwapInInfo() (*SwapInfoResponse, error)
	GetSwapOutInfo() (*SwapInfoResponse, error)
	InitiateSwapIn(ctx context.Context, initiateSwapInRequest *InitiateSwapRequest) (*swaps.SwapResponse, error)
	InitiateSwapOut(ctx context.Context, initiateSwapOutRequest *InitiateSwapRequest) (*swaps.SwapResponse, error)
	RefundSwap(refundSwapRequest *RefundSwapRequest) error
	GetSwapMnemonic() string
	GetAutoSwapConfig() (*GetAutoSwapConfigResponse, error)
	EnableAutoSwapOut(ctx context.Context, autoSwapRequest *EnableAutoSwapRequest) error
	DisableAutoSwap() error
	SetNodeAlias(nodeAlias string) error
	GetCustomNodeCommands() (*CustomNodeCommandsResponse, error)
	ExecuteCustomNodeCommand(ctx context.Context, command string) (interface{}, error)
	SendEvent(event string, properties interface{})
	GetForwards() (*GetForwardsResponse, error)
}

// App represents a connected NWC application with its permissions and usage details.
type App struct {
	ID                 uint       `json:"id"`
	Name               string     `json:"name"`
	Description        string     `json:"description"`
	AppPubkey          string     `json:"appPubkey"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
	LastUsedAt         *time.Time `json:"lastUsedAt"`
	ExpiresAt          *time.Time `json:"expiresAt"`
	Scopes             []string   `json:"scopes"`
	MaxAmountSat       uint64     `json:"maxAmount"`
	BudgetUsage        uint64     `json:"budgetUsage"`
	BudgetRenewal      string     `json:"budgetRenewal"`
	Isolated           bool       `json:"isolated"`
	WalletPubkey       string     `json:"walletPubkey"`
	UniqueWalletPubkey bool       `json:"uniqueWalletPubkey"`
	Balance            int64      `json:"balance"`
	Metadata           Metadata   `json:"metadata,omitempty"`
}

// ListAppsFilters specifies filter criteria for listing applications.
type ListAppsFilters struct {
	Name          string `json:"name"`
	AppStoreAppId string `json:"appStoreAppId"`
	Unused        bool   `json:"unused"`
	SubWallets    *bool  `json:"subWallets"`
}

// ListAppsResponse contains the paginated list of applications.
type ListAppsResponse struct {
	Apps       []App  `json:"apps"`
	TotalCount uint64 `json:"totalCount"`
}

// UpdateAppRequest contains the fields that can be updated for an application.
type UpdateAppRequest struct {
	Name            *string   `json:"name"`
	MaxAmountSat    *uint64   `json:"maxAmount"`
	BudgetRenewal   *string   `json:"budgetRenewal"`
	ExpiresAt       *string   `json:"expiresAt"`
	UpdateExpiresAt bool      `json:"updateExpiresAt"`
	Scopes          []string  `json:"scopes"`
	Metadata        *Metadata `json:"metadata"`
	Isolated        *bool     `json:"isolated"`
}

// TransferRequest specifies a balance transfer between two applications.
type TransferRequest struct {
	AmountSat uint64 `json:"amountSat"`
	FromAppId *uint  `json:"fromAppId"`
	ToAppId   *uint  `json:"toAppId"`
}

// CreateAppRequest contains the parameters for creating a new NWC application connection.
type CreateAppRequest struct {
	Name           string   `json:"name"`
	Pubkey         string   `json:"pubkey"`
	MaxAmountSat   uint64   `json:"maxAmount"`
	BudgetRenewal  string   `json:"budgetRenewal"`
	ExpiresAt      string   `json:"expiresAt"`
	Scopes         []string `json:"scopes"`
	ReturnTo       string   `json:"returnTo"`
	Isolated       bool     `json:"isolated"`
	Metadata       Metadata `json:"metadata,omitempty"`
	UnlockPassword string   `json:"unlockPassword"`
}

// CreateLightningAddressRequest contains the parameters for creating a lightning address for an app.
type CreateLightningAddressRequest struct {
	Address string `json:"address"`
	AppId   uint   `json:"appId"`
}

// InitiateSwapRequest contains the parameters for initiating a swap-in or swap-out.
type InitiateSwapRequest struct {
	SwapAmount  uint64 `json:"swapAmount"`
	Destination string `json:"destination"`
}

// RefundSwapRequest contains the parameters for refunding a swap.
type RefundSwapRequest struct {
	SwapId  string `json:"swapId"`
	Address string `json:"address"`
}

// EnableAutoSwapRequest contains the parameters for enabling automatic swaps.
type EnableAutoSwapRequest struct {
	BalanceThreshold uint64 `json:"balanceThreshold"`
	SwapAmount       uint64 `json:"swapAmount"`
	Destination      string `json:"destination"`
}

// GetAutoSwapConfigResponse contains the current auto-swap configuration.
type GetAutoSwapConfigResponse struct {
	Type             string `json:"type"`
	Enabled          bool   `json:"enabled"`
	BalanceThreshold uint64 `json:"balanceThreshold"`
	SwapAmount       uint64 `json:"swapAmount"`
	Destination      string `json:"destination"`
}

// SwapInfoResponse contains fee and amount limits for swap operations.
type SwapInfoResponse struct {
	AlbyServiceFee  float64 `json:"albyServiceFee"`
	BoltzServiceFee float64 `json:"boltzServiceFee"`
	BoltzNetworkFee uint64  `json:"boltzNetworkFee"`
	MinAmount       uint64  `json:"minAmount"`
	MaxAmount       uint64  `json:"maxAmount"`
}

// ListSwapsResponse contains the list of all swaps.
type ListSwapsResponse struct {
	Swaps []Swap `json:"swaps"`
}

// LookupSwapResponse is an alias for Swap, returned when looking up a single swap.
type LookupSwapResponse = Swap

// Swap represents a submarine swap (swap-in or swap-out) with its current state.
type Swap struct {
	Id                 string `json:"id"`
	Type               string `json:"type"`
	State              string `json:"state"`
	Invoice            string `json:"invoice"`
	SendAmount         uint64 `json:"sendAmount"`
	ReceiveAmount      uint64 `json:"receiveAmount"`
	PaymentHash        string `json:"paymentHash"`
	DestinationAddress string `json:"destinationAddress"`
	RefundAddress      string `json:"refundAddress"`
	LockupAddress      string `json:"lockupAddress"`
	LockupTxId         string `json:"lockupTxId"`
	ClaimTxId          string `json:"claimTxId"`
	AutoSwap           bool   `json:"autoSwap"`
	BoltzPubkey        string `json:"boltzPubkey"`
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
	UsedXpub           bool   `json:"usedXpub"`
}

// StartRequest contains the unlock password needed to start the node.
type StartRequest struct {
	UnlockPassword string `json:"unlockPassword"`
}

// UnlockRequest contains the credentials for unlocking the node.
type UnlockRequest struct {
	UnlockPassword  string  `json:"unlockPassword"`
	TokenExpiryDays *uint64 `json:"tokenExpiryDays"`
	Permission      string  `json:"permission,omitempty"` // "full" or "readonly"
}

// BackupReminderRequest contains the next backup reminder date.
type BackupReminderRequest struct {
	NextBackupReminder string `json:"nextBackupReminder"`
}

// SendEventRequest contains the event name and properties to send for analytics.
type SendEventRequest struct {
	Event      string      `json:"event"`
	Properties interface{} `json:"properties"`
}

// SetupRequest contains all the parameters needed for initial node setup.
type SetupRequest struct {
	LNBackendType  string `json:"backendType"`
	UnlockPassword string `json:"unlockPassword"`

	Mnemonic           string `json:"mnemonic"`
	NextBackupReminder string `json:"nextBackupReminder"`

	// LND fields
	LNDAddress      string `json:"lndAddress"`
	LNDCertFile     string `json:"lndCertFile"`
	LNDMacaroonFile string `json:"lndMacaroonFile"`
	LNDCertHex      string `json:"lndCertHex"`
	LNDMacaroonHex  string `json:"lndMacaroonHex"`

	// Phoenixd fields
	PhoenixdAddress       string `json:"phoenixdAddress"`
	PhoenixdAuthorization string `json:"phoenixdAuthorization"`

	// Cashu fields
	CashuMintUrl string `json:"cashuMintUrl"`
}

// CreateAppResponse contains the pairing details returned after creating a new app connection.
type CreateAppResponse struct {
	PairingUri    string   `json:"pairingUri"`
	PairingSecret string   `json:"pairingSecretKey"`
	Pubkey        string   `json:"pairingPublicKey"`
	RelayUrls     []string `json:"relayUrls"`
	WalletPubkey  string   `json:"walletPubkey"`
	Lud16         string   `json:"lud16"`
	Id            uint     `json:"id"`
	Name          string   `json:"name"`
	ReturnTo      string   `json:"returnTo"`
}

// User represents a user with their email address.
type User struct {
	Email string `json:"email"`
}

// InfoResponseRelay represents a Nostr relay with its URL and online status.
type InfoResponseRelay struct {
	Url    string `json:"url"`
	Online bool   `json:"online"`
}

// InfoResponse contains the overall status and configuration of the Alby Hub instance.
type InfoResponse struct {
	BackendType                 string              `json:"backendType"`
	SetupCompleted              bool                `json:"setupCompleted"`
	OAuthRedirect               bool                `json:"oauthRedirect"`
	Running                     bool                `json:"running"`
	Unlocked                    bool                `json:"unlocked"`
	AlbyAuthUrl                 string              `json:"albyAuthUrl"`
	NextBackupReminder          string              `json:"nextBackupReminder"`
	AlbyUserIdentifier          string              `json:"albyUserIdentifier"`
	AlbyAccountConnected        bool                `json:"albyAccountConnected"`
	Version                     string              `json:"version"`
	Network                     string              `json:"network"`
	EnableAdvancedSetup         bool                `json:"enableAdvancedSetup"`
	LdkVssEnabled               bool                `json:"ldkVssEnabled"`
	VssSupported                bool                `json:"vssSupported"`
	StartupState                string              `json:"startupState"`
	StartupError                string              `json:"startupError"`
	StartupErrorTime            time.Time           `json:"startupErrorTime"`
	AutoUnlockPasswordSupported bool                `json:"autoUnlockPasswordSupported"`
	AutoUnlockPasswordEnabled   bool                `json:"autoUnlockPasswordEnabled"`
	Currency                    string              `json:"currency"`
	BitcoinDisplayFormat        string              `json:"bitcoinDisplayFormat"`
	Relays                      []InfoResponseRelay `json:"relays"`
	NodeAlias                   string              `json:"nodeAlias"`
	MempoolUrl                  string              `json:"mempoolUrl"`
}

// UpdateSettingsRequest contains the settings fields that can be updated.
type UpdateSettingsRequest struct {
	Currency             string `json:"currency"`
	BitcoinDisplayFormat string `json:"bitcoinDisplayFormat"`
}

// SetNodeAliasRequest contains the new node alias to set.
type SetNodeAliasRequest struct {
	NodeAlias string `json:"nodeAlias"`
}

// MnemonicRequest contains the unlock password required to retrieve the mnemonic.
type MnemonicRequest struct {
	UnlockPassword string `json:"unlockPassword"`
}

// MnemonicResponse contains the wallet mnemonic seed phrase.
type MnemonicResponse struct {
	Mnemonic string `json:"mnemonic"`
}

// ChangeUnlockPasswordRequest contains the current and new unlock passwords.
type ChangeUnlockPasswordRequest struct {
	CurrentUnlockPassword string `json:"currentUnlockPassword"`
	NewUnlockPassword     string `json:"newUnlockPassword"`
}
// AutoUnlockRequest contains the password used for automatic unlock on startup.
type AutoUnlockRequest struct {
	UnlockPassword string `json:"unlockPassword"`
}

// ConnectPeerRequest is an alias for lnclient.ConnectPeerRequest.
type ConnectPeerRequest = lnclient.ConnectPeerRequest

// OpenChannelRequest is an alias for lnclient.OpenChannelRequest.
type OpenChannelRequest = lnclient.OpenChannelRequest

// OpenChannelResponse is an alias for lnclient.OpenChannelResponse.
type OpenChannelResponse = lnclient.OpenChannelResponse

// CloseChannelResponse is an alias for lnclient.CloseChannelResponse.
type CloseChannelResponse = lnclient.CloseChannelResponse

// UpdateChannelRequest is an alias for lnclient.UpdateChannelRequest.
type UpdateChannelRequest = lnclient.UpdateChannelRequest

// RebalanceChannelRequest contains the parameters for rebalancing a channel.
type RebalanceChannelRequest struct {
	ReceiveThroughNodePubkey string `json:"receiveThroughNodePubkey"`
	AmountSat                uint64 `json:"amountSat"`
}
// RebalanceChannelResponse contains the fee paid for the channel rebalance.
type RebalanceChannelResponse struct {
	TotalFeeSat uint64 `json:"totalFeeSat"`
}

// RedeemOnchainFundsRequest contains the parameters for sweeping onchain funds.
type RedeemOnchainFundsRequest struct {
	ToAddress string  `json:"toAddress"`
	Amount    uint64  `json:"amount"`
	FeeRate   *uint64 `json:"feeRate"`
	SendAll   bool    `json:"sendAll"`
}

// RedeemOnchainFundsResponse contains the transaction ID of the sweep transaction.
type RedeemOnchainFundsResponse struct {
	TxId string `json:"txId"`
}

// OnchainBalanceResponse is an alias for lnclient.OnchainBalanceResponse.
type OnchainBalanceResponse = lnclient.OnchainBalanceResponse

// BalancesResponse is an alias for lnclient.BalancesResponse.
type BalancesResponse = lnclient.BalancesResponse

// SendPaymentResponse is a Transaction returned after sending a payment.
type SendPaymentResponse = Transaction

// MakeInvoiceResponse is a Transaction returned after creating an invoice.
type MakeInvoiceResponse = Transaction

// LookupInvoiceResponse is a Transaction returned when looking up an invoice.
type LookupInvoiceResponse = Transaction

// ListTransactionsResponse contains the paginated list of transactions.
type ListTransactionsResponse struct {
	TotalCount   uint64        `json:"totalCount"`
	Transactions []Transaction `json:"transactions"`
}

// Transaction represents a Lightning payment or invoice with its current state.
// TODO: camelCase
type Transaction struct {
	Type            string      `json:"type"`
	State           string      `json:"state"`
	Invoice         string      `json:"invoice"`
	Description     string      `json:"description"`
	DescriptionHash string      `json:"descriptionHash"`
	Preimage        *string     `json:"preimage"`
	PaymentHash     string      `json:"paymentHash"`
	Amount          uint64      `json:"amount"`
	FeesPaid        uint64      `json:"feesPaid"`
	UpdatedAt       string      `json:"updatedAt"`
	CreatedAt       string      `json:"createdAt"`
	SettledAt       *string     `json:"settledAt"`
	AppId           *uint       `json:"appId"`
	Metadata        Metadata    `json:"metadata,omitempty"`
	Boostagram      *Boostagram `json:"boostagram,omitempty"`
	FailureReason   string      `json:"failureReason"`
}

// Metadata is a map of arbitrary key-value pairs attached to transactions and apps.
type Metadata = map[string]interface{}

// Boostagram represents a podcast boost payment with its associated metadata.
type Boostagram struct {
	AppName        string `json:"appName"`
	Name           string `json:"name"`
	Podcast        string `json:"podcast"`
	URL            string `json:"url"`
	Episode        string `json:"episode,omitempty"`
	FeedId         string `json:"feedId,omitempty"`
	ItemId         string `json:"itemId,omitempty"`
	Timestamp      int64  `json:"ts,omitempty"`
	Message        string `json:"message,omitempty"`
	SenderId       string `json:"senderId"`
	SenderName     string `json:"senderName"`
	Time           string `json:"time"`
	Action         string `json:"action"`
	ValueMsatTotal int64  `json:"valueMsatTotal"`
}

// SendPaymentProbesRequest contains the invoice for probing a payment route.
type SendPaymentProbesRequest struct {
	Invoice string `json:"invoice"`
}

// SendPaymentProbesResponse contains the result of probing a payment route.
type SendPaymentProbesResponse struct {
	Error string `json:"error"`
}

// SendSpontaneousPaymentProbesRequest contains the parameters for probing a spontaneous payment route.
type SendSpontaneousPaymentProbesRequest struct {
	Amount uint64 `json:"amount"`
	NodeId string `json:"nodeId"`
}

// SendSpontaneousPaymentProbesResponse contains the result of probing a spontaneous payment route.
type SendSpontaneousPaymentProbesResponse struct {
	Error string `json:"error"`
}

const (
	// LogTypeNode is the log type identifier for node logs.
	LogTypeNode = "node"
	// LogTypeApp is the log type identifier for application logs.
	LogTypeApp = "app"
)

// GetLogOutputRequest specifies the maximum length of log output to retrieve.
type GetLogOutputRequest struct {
	MaxLen int `query:"maxLen"`
}

// GetLogOutputResponse contains the log output string.
type GetLogOutputResponse struct {
	Log string `json:"logs"`
}

// SignMessageRequest contains the message to be signed by the node.
type SignMessageRequest struct {
	Message string `json:"message"`
}

// SignMessageResponse contains the signed message and its signature.
type SignMessageResponse struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

// PayInvoiceRequest contains the optional amount and metadata for paying an invoice.
type PayInvoiceRequest struct {
	Amount   *uint64  `json:"amount"`
	Metadata Metadata `json:"metadata"`
}

// MakeOfferRequest contains the description for creating a BOLT12 offer.
type MakeOfferRequest struct {
	Description string `json:"description"`
}

// MakeInvoiceRequest contains the amount and description for creating an invoice.
type MakeInvoiceRequest struct {
	Amount      uint64 `json:"amount"`
	Description string `json:"description"`
}

// ResetRouterRequest contains the key used to authorize a router reset.
type ResetRouterRequest struct {
	Key string `json:"key"`
}

// BasicBackupRequest contains the unlock password needed to create a backup.
type BasicBackupRequest struct {
	UnlockPassword string `json:"unlockPassword"`
}

// BasicRestoreWailsRequest contains the unlock password needed to restore from a backup via Wails.
type BasicRestoreWailsRequest struct {
	UnlockPassword string `json:"unlockPassword"`
}

// NetworkGraphResponse is an alias for lnclient.NetworkGraphResponse.
type NetworkGraphResponse = lnclient.NetworkGraphResponse

// LSPOrderRequest contains the parameters for requesting a channel from an LSP.
type LSPOrderRequest struct {
	Amount        uint64 `json:"amount"`
	LSPType       string `json:"lspType"`
	LSPIdentifier string `json:"lspIdentifier"`
	Public        bool   `json:"public"`
}

// LSPOrderResponse contains the invoice and liquidity details for an LSP channel order.
type LSPOrderResponse struct {
	Invoice           string `json:"invoice"`
	Fee               uint64 `json:"fee"`
	InvoiceAmount     uint64 `json:"invoiceAmount"`
	IncomingLiquidity uint64 `json:"incomingLiquidity"`
	OutgoingLiquidity uint64 `json:"outgoingLiquidity"`
}

// WalletCapabilitiesResponse contains the NIP-47 scopes, methods, and notification types supported by the wallet.
type WalletCapabilitiesResponse struct {
	Scopes            []string `json:"scopes"`
	Methods           []string `json:"methods"`
	NotificationTypes []string `json:"notificationTypes"`
}

// Channel represents a Lightning channel with its balances, fees, and status.
type Channel struct {
	LocalBalance                             int64       `json:"localBalance"`
	LocalSpendableBalance                    int64       `json:"localSpendableBalance"`
	RemoteBalance                            int64       `json:"remoteBalance"`
	Id                                       string      `json:"id"`
	RemotePubkey                             string      `json:"remotePubkey"`
	FundingTxId                              string      `json:"fundingTxId"`
	FundingTxVout                            uint32      `json:"fundingTxVout"`
	Active                                   bool        `json:"active"`
	Public                                   bool        `json:"public"`
	InternalChannel                          interface{} `json:"internalChannel"`
	Confirmations                            *uint32     `json:"confirmations"`
	ConfirmationsRequired                    *uint32     `json:"confirmationsRequired"`
	ForwardingFeeBaseMsat                    uint32      `json:"forwardingFeeBaseMsat"`
	ForwardingFeeProportionalMillionths      uint32      `json:"forwardingFeeProportionalMillionths"`
	UnspendablePunishmentReserve             uint64      `json:"unspendablePunishmentReserve"`
	CounterpartyUnspendablePunishmentReserve uint64      `json:"counterpartyUnspendablePunishmentReserve"`
	Error                                    *string     `json:"error"`
	Status                                   string      `json:"status"`
	IsOutbound                               bool        `json:"isOutbound"`
}

// MigrateNodeStorageRequest contains the target storage backend for migration.
type MigrateNodeStorageRequest struct {
	To string `json:"to"`
}

// HealthAlarmKind represents the type of a health alarm.
type HealthAlarmKind string

const (
	// HealthAlarmKindAlbyService indicates an issue with the Alby service connection.
	HealthAlarmKindAlbyService HealthAlarmKind = "alby_service"
	// HealthAlarmKindNodeNotReady indicates the Lightning node is not ready.
	HealthAlarmKindNodeNotReady HealthAlarmKind = "node_not_ready"
	// HealthAlarmKindChannelsOffline indicates one or more channels are offline.
	HealthAlarmKindChannelsOffline HealthAlarmKind = "channels_offline"
	// HealthAlarmKindNostrRelayOffline indicates a Nostr relay is offline.
	HealthAlarmKindNostrRelayOffline HealthAlarmKind = "nostr_relay_offline"
	// HealthAlarmKindVssNoSubscription indicates VSS has no active subscription.
	HealthAlarmKindVssNoSubscription HealthAlarmKind = "vss_no_subscription"
)

// HealthAlarm represents a single health check alarm with its kind and details.
type HealthAlarm struct {
	Kind       HealthAlarmKind `json:"kind"`
	RawDetails any             `json:"rawDetails,omitempty"`
}

// NewHealthAlarm creates a new HealthAlarm with the given kind and raw details.
func NewHealthAlarm(kind HealthAlarmKind, rawDetails any) HealthAlarm {
	return HealthAlarm{
		Kind:       kind,
		RawDetails: rawDetails,
	}
}

// HealthResponse contains the list of active health alarms.
type HealthResponse struct {
	Alarms []HealthAlarm `json:"alarms,omitempty"`
}

// CustomNodeCommandArgDef defines a single argument for a custom node command.
type CustomNodeCommandArgDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CustomNodeCommandDef defines a custom node command with its name, description, and arguments.
type CustomNodeCommandDef struct {
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Args        []CustomNodeCommandArgDef `json:"args"`
}

// CustomNodeCommandsResponse contains the list of available custom node commands.
type CustomNodeCommandsResponse struct {
	Commands []CustomNodeCommandDef `json:"commands"`
}

// ExecuteCustomNodeCommandRequest contains the command name to execute.
type ExecuteCustomNodeCommandRequest struct {
	Command string `json:"command"`
}

// GetForwardsResponse contains the total forwarding statistics for the node.
type GetForwardsResponse struct {
	OutboundAmountForwardedMsat uint64 `json:"outboundAmountForwardedMsat"`
	TotalFeeEarnedMsat          uint64 `json:"totalFeeEarnedMsat"`
	NumForwards                 uint64 `json:"numForwards"`
}
