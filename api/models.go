package api

import (
	"context"
	"io"
	"time"

	"github.com/getAlby/hub/alby"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/lnclient"
)

type API interface {
	CreateApp(createAppRequest *CreateAppRequest) (*CreateAppResponse, error)
	UpdateApp(app *db.App, updateAppRequest *UpdateAppRequest) error
	TopupIsolatedApp(ctx context.Context, app *db.App, amountMsat uint64) error
	DeleteApp(app *db.App) error
	GetApp(app *db.App) *App
	ListApps() ([]App, error)
	ListChannels(ctx context.Context) ([]Channel, error)
	GetChannelPeerSuggestions(ctx context.Context) ([]alby.ChannelPeerSuggestion, error)
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
	RequestMempoolApi(endpoint string) (interface{}, error)
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
	GetAutoSwapsConfig() (*GetAutoSwapsConfigResponse, error)
	DisableAutoSwaps() error
	EnableAutoSwaps(ctx context.Context, autoSwapsRequest *EnableAutoSwapsRequest) error
	GetCustomNodeCommands() (*CustomNodeCommandsResponse, error)
	ExecuteCustomNodeCommand(ctx context.Context, command string) (interface{}, error)
	SendEvent(event string)
}

type App struct {
	ID                 uint       `json:"id"`
	Name               string     `json:"name"`
	Description        string     `json:"description"`
	AppPubkey          string     `json:"appPubkey"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
	LastEventAt        *time.Time `json:"lastEventAt"`
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

type ListAppsResponse struct {
	Apps []App `json:"apps"`
}

type UpdateAppRequest struct {
	Name          string   `json:"name"`
	MaxAmountSat  uint64   `json:"maxAmount"`
	BudgetRenewal string   `json:"budgetRenewal"`
	ExpiresAt     string   `json:"expiresAt"`
	Scopes        []string `json:"scopes"`
	Metadata      Metadata `json:"metadata,omitempty"`
	Isolated      bool     `json:"isolated"`
}

type TopupIsolatedAppRequest struct {
	AmountSat uint64 `json:"amountSat"`
}

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

type EnableAutoSwapsRequest struct {
	BalanceThreshold uint64 `json:"balanceThreshold"`
	SwapAmount       uint64 `json:"swapAmount"`
	Destination      string `json:"destination"`
}

type GetAutoSwapsConfigResponse struct {
	Enabled          bool    `json:"enabled"`
	BalanceThreshold uint64  `json:"balanceThreshold"`
	SwapAmount       uint64  `json:"swapAmount"`
	Destination      string  `json:"destination"`
	AlbyServiceFee   float64 `json:"albyServiceFee"`
	BoltzServiceFee  float64 `json:"boltzServiceFee"`
	BoltzNetworkFee  uint64  `json:"boltzNetworkFee"`
}

type StartRequest struct {
	UnlockPassword string `json:"unlockPassword"`
}

type UnlockRequest struct {
	UnlockPassword  string  `json:"unlockPassword"`
	TokenExpiryDays *uint64 `json:"tokenExpiryDays"`
}

type BackupReminderRequest struct {
	NextBackupReminder string `json:"nextBackupReminder"`
}

type SendEventRequest struct {
	Event string `json:"event"`
}

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

type CreateAppResponse struct {
	PairingUri    string `json:"pairingUri"`
	PairingSecret string `json:"pairingSecretKey"`
	Pubkey        string `json:"pairingPublicKey"`
	RelayUrl      string `json:"relayUrl"`
	WalletPubkey  string `json:"walletPubkey"`
	Lud16         string `json:"lud16"`
	Id            uint   `json:"id"`
	Name          string `json:"name"`
	ReturnTo      string `json:"returnTo"`
}

type User struct {
	Email string `json:"email"`
}

type InfoResponse struct {
	BackendType                 string    `json:"backendType"`
	SetupCompleted              bool      `json:"setupCompleted"`
	OAuthRedirect               bool      `json:"oauthRedirect"`
	Running                     bool      `json:"running"`
	Unlocked                    bool      `json:"unlocked"`
	AlbyAuthUrl                 string    `json:"albyAuthUrl"`
	NextBackupReminder          string    `json:"nextBackupReminder"`
	AlbyUserIdentifier          string    `json:"albyUserIdentifier"`
	AlbyAccountConnected        bool      `json:"albyAccountConnected"`
	Version                     string    `json:"version"`
	Network                     string    `json:"network"`
	EnableAdvancedSetup         bool      `json:"enableAdvancedSetup"`
	LdkVssEnabled               bool      `json:"ldkVssEnabled"`
	VssSupported                bool      `json:"vssSupported"`
	StartupState                string    `json:"startupState"`
	StartupError                string    `json:"startupError"`
	StartupErrorTime            time.Time `json:"startupErrorTime"`
	AutoUnlockPasswordSupported bool      `json:"autoUnlockPasswordSupported"`
	AutoUnlockPasswordEnabled   bool      `json:"autoUnlockPasswordEnabled"`
	Currency                    string    `json:"currency"`
	Relay                       string    `json:"relay"`
}

type UpdateSettingsRequest struct {
	Currency string `json:"currency"`
}

type MnemonicRequest struct {
	UnlockPassword string `json:"unlockPassword"`
}

type MnemonicResponse struct {
	Mnemonic string `json:"mnemonic"`
}

type ChangeUnlockPasswordRequest struct {
	CurrentUnlockPassword string `json:"currentUnlockPassword"`
	NewUnlockPassword     string `json:"newUnlockPassword"`
}
type AutoUnlockRequest struct {
	UnlockPassword string `json:"unlockPassword"`
}

type ConnectPeerRequest = lnclient.ConnectPeerRequest
type OpenChannelRequest = lnclient.OpenChannelRequest
type OpenChannelResponse = lnclient.OpenChannelResponse
type CloseChannelResponse = lnclient.CloseChannelResponse
type UpdateChannelRequest = lnclient.UpdateChannelRequest

type RedeemOnchainFundsRequest struct {
	ToAddress string  `json:"toAddress"`
	Amount    uint64  `json:"amount"`
	FeeRate   *uint64 `json:"feeRate"`
	SendAll   bool    `json:"sendAll"`
}

type RedeemOnchainFundsResponse struct {
	TxId string `json:"txId"`
}

type OnchainBalanceResponse = lnclient.OnchainBalanceResponse
type BalancesResponse = lnclient.BalancesResponse

type SendPaymentResponse = Transaction
type MakeInvoiceResponse = Transaction
type LookupInvoiceResponse = Transaction

type ListTransactionsResponse struct {
	TotalCount   uint64        `json:"totalCount"`
	Transactions []Transaction `json:"transactions"`
}

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

type Metadata = map[string]interface{}

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

// debug api
type SendPaymentProbesRequest struct {
	Invoice string `json:"invoice"`
}

type SendPaymentProbesResponse struct {
	Error string `json:"error"`
}

type SendSpontaneousPaymentProbesRequest struct {
	Amount uint64 `json:"amount"`
	NodeId string `json:"nodeId"`
}

type SendSpontaneousPaymentProbesResponse struct {
	Error string `json:"error"`
}

const (
	LogTypeNode = "node"
	LogTypeApp  = "app"
)

type GetLogOutputRequest struct {
	MaxLen int `query:"maxLen"`
}

type GetLogOutputResponse struct {
	Log string `json:"logs"`
}

type SignMessageRequest struct {
	Message string `json:"message"`
}

type SignMessageResponse struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

type PayInvoiceRequest struct {
	Amount   *uint64  `json:"amount"`
	Metadata Metadata `json:"metadata"`
}

type MakeOfferRequest struct {
	Description string `json:"description"`
}

type MakeInvoiceRequest struct {
	Amount      uint64 `json:"amount"`
	Description string `json:"description"`
}

type ResetRouterRequest struct {
	Key string `json:"key"`
}

type BasicBackupRequest struct {
	UnlockPassword string `json:"unlockPassword"`
}

type BasicRestoreWailsRequest struct {
	UnlockPassword string `json:"unlockPassword"`
}

type NetworkGraphResponse = lnclient.NetworkGraphResponse

type LSPOrderRequest struct {
	Amount  uint64 `json:"amount"`
	LSPType string `json:"lspType"`
	LSPUrl  string `json:"lspUrl"`
	Public  bool   `json:"public"`
}

type LSPOrderResponse struct {
	Invoice           string `json:"invoice"`
	Fee               uint64 `json:"fee"`
	InvoiceAmount     uint64 `json:"invoiceAmount"`
	IncomingLiquidity uint64 `json:"incomingLiquidity"`
	OutgoingLiquidity uint64 `json:"outgoingLiquidity"`
}

type WalletCapabilitiesResponse struct {
	Scopes            []string `json:"scopes"`
	Methods           []string `json:"methods"`
	NotificationTypes []string `json:"notificationTypes"`
}

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
	UnspendablePunishmentReserve             uint64      `json:"unspendablePunishmentReserve"`
	CounterpartyUnspendablePunishmentReserve uint64      `json:"counterpartyUnspendablePunishmentReserve"`
	Error                                    *string     `json:"error"`
	Status                                   string      `json:"status"`
	IsOutbound                               bool        `json:"isOutbound"`
}

type MigrateNodeStorageRequest struct {
	To string `json:"to"`
}

type HealthAlarmKind string

const (
	HealthAlarmKindAlbyService       HealthAlarmKind = "alby_service"
	HealthAlarmKindNodeNotReady      HealthAlarmKind = "node_not_ready"
	HealthAlarmKindChannelsOffline   HealthAlarmKind = "channels_offline"
	HealthAlarmKindNostrRelayOffline HealthAlarmKind = "nostr_relay_offline"
	HealthAlarmKindVssNoSubscription HealthAlarmKind = "vss_no_subscription"
)

type HealthAlarm struct {
	Kind       HealthAlarmKind `json:"kind"`
	RawDetails any             `json:"rawDetails,omitempty"`
}

func NewHealthAlarm(kind HealthAlarmKind, rawDetails any) HealthAlarm {
	return HealthAlarm{
		Kind:       kind,
		RawDetails: rawDetails,
	}
}

type HealthResponse struct {
	Alarms []HealthAlarm `json:"alarms,omitempty"`
}

type CustomNodeCommandArgDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CustomNodeCommandDef struct {
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Args        []CustomNodeCommandArgDef `json:"args"`
}

type CustomNodeCommandsResponse struct {
	Commands []CustomNodeCommandDef `json:"commands"`
}

type ExecuteCustomNodeCommandRequest struct {
	Command string `json:"command"`
}
