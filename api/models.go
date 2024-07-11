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
	UpdateApp(userApp *db.App, updateAppRequest *UpdateAppRequest) error
	DeleteApp(userApp *db.App) error
	GetApp(userApp *db.App) *App
	ListApps() ([]App, error)
	ListChannels(ctx context.Context) ([]lnclient.Channel, error)
	GetChannelPeerSuggestions(ctx context.Context) ([]alby.ChannelPeerSuggestion, error)
	ResetRouter(key string) error
	ChangeUnlockPassword(changeUnlockPasswordRequest *ChangeUnlockPasswordRequest) error
	Stop() error
	GetNodeConnectionInfo(ctx context.Context) (*lnclient.NodeConnectionInfo, error)
	GetNodeStatus(ctx context.Context) (*lnclient.NodeStatus, error)
	ListPeers(ctx context.Context) ([]lnclient.PeerDetails, error)
	ConnectPeer(ctx context.Context, connectPeerRequest *ConnectPeerRequest) error
	DisconnectPeer(ctx context.Context, peerId string) error
	OpenChannel(ctx context.Context, openChannelRequest *OpenChannelRequest) (*OpenChannelResponse, error)
	CloseChannel(ctx context.Context, peerId, channelId string, force bool) (*CloseChannelResponse, error)
	UpdateChannel(ctx context.Context, updateChannelRequest *UpdateChannelRequest) error
	GetNewOnchainAddress(ctx context.Context) (string, error)
	GetUnusedOnchainAddress(ctx context.Context) (string, error)
	SignMessage(ctx context.Context, message string) (*SignMessageResponse, error)
	RedeemOnchainFunds(ctx context.Context, toAddress string) (*RedeemOnchainFundsResponse, error)
	GetBalances(ctx context.Context) (*BalancesResponse, error)
	ListTransactions(ctx context.Context, limit uint64, offset uint64) (*ListTransactionsResponse, error)
	SendPayment(ctx context.Context, invoice string) (*SendPaymentResponse, error)
	CreateInvoice(ctx context.Context, amount int64, description string) (*MakeInvoiceResponse, error)
	LookupInvoice(ctx context.Context, paymentHash string) (*LookupInvoiceResponse, error)
	RequestMempoolApi(endpoint string) (interface{}, error)
	GetInfo(ctx context.Context) (*InfoResponse, error)
	GetEncryptedMnemonic() *EncryptedMnemonicResponse
	SetNextBackupReminder(backupReminderRequest *BackupReminderRequest) error
	Start(startRequest *StartRequest) error
	Setup(ctx context.Context, setupRequest *SetupRequest) error
	SendPaymentProbes(ctx context.Context, sendPaymentProbesRequest *SendPaymentProbesRequest) (*SendPaymentProbesResponse, error)
	SendSpontaneousPaymentProbes(ctx context.Context, sendSpontaneousPaymentProbesRequest *SendSpontaneousPaymentProbesRequest) (*SendSpontaneousPaymentProbesResponse, error)
	GetNetworkGraph(nodeIds []string) (NetworkGraphResponse, error)
	SyncWallet() error
	GetLogOutput(ctx context.Context, logType string, getLogRequest *GetLogOutputRequest) (*GetLogOutputResponse, error)
	NewInstantChannelInvoice(ctx context.Context, request *NewInstantChannelInvoiceRequest) (*NewInstantChannelInvoiceResponse, error)
	CreateBackup(unlockPassword string, w io.Writer) error
	RestoreBackup(unlockPassword string, r io.Reader) error
	GetWalletCapabilities(ctx context.Context) (*WalletCapabilitiesResponse, error)
}

type App struct {
	ID            uint       `json:"id"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	NostrPubkey   string     `json:"nostrPubkey"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	LastEventAt   *time.Time `json:"lastEventAt"`
	ExpiresAt     *time.Time `json:"expiresAt"`
	Scopes        []string   `json:"scopes"`
	MaxAmount     uint64     `json:"maxAmount"`
	BudgetUsage   uint64     `json:"budgetUsage"`
	BudgetRenewal string     `json:"budgetRenewal"`
}

type ListAppsResponse struct {
	Apps []App `json:"apps"`
}

type UpdateAppRequest struct {
	MaxAmount     uint64   `json:"maxAmount"`
	BudgetRenewal string   `json:"budgetRenewal"`
	ExpiresAt     string   `json:"expiresAt"`
	Scopes        []string `json:"scopes"`
}

type CreateAppRequest struct {
	Name          string   `json:"name"`
	Pubkey        string   `json:"pubkey"`
	MaxAmount     uint64   `json:"maxAmount"`
	BudgetRenewal string   `json:"budgetRenewal"`
	ExpiresAt     string   `json:"expiresAt"`
	Scopes        []string `json:"scopes"`
	ReturnTo      string   `json:"returnTo"`
}

type StartRequest struct {
	UnlockPassword string `json:"unlockPassword"`
}

type UnlockRequest struct {
	UnlockPassword string `json:"unlockPassword"`
}

type BackupReminderRequest struct {
	NextBackupReminder string `json:"nextBackupReminder"`
}

type SetupRequest struct {
	LNBackendType  string `json:"backendType"`
	UnlockPassword string `json:"unlockPassword"`

	// Breez / Greenlight
	Mnemonic             string `json:"mnemonic"`
	GreenlightInviteCode string `json:"greenlightInviteCode"`
	NextBackupReminder   string `json:"nextBackupReminder"`

	// Breez fields
	BreezAPIKey string `json:"breezApiKey"`

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
	Name          string `json:"name"`
	ReturnTo      string `json:"returnTo"`
}

type User struct {
	Email string `json:"email"`
}

type InfoResponse struct {
	BackendType          string `json:"backendType"`
	SetupCompleted       bool   `json:"setupCompleted"`
	OAuthRedirect        bool   `json:"oauthRedirect"`
	Running              bool   `json:"running"`
	Unlocked             bool   `json:"unlocked"`
	AlbyAuthUrl          string `json:"albyAuthUrl"`
	NextBackupReminder   string `json:"nextBackupReminder"`
	AlbyUserIdentifier   string `json:"albyUserIdentifier"`
	AlbyAccountConnected bool   `json:"albyAccountConnected"`
	Version              string `json:"version"`
	Network              string `json:"network"`
}

type EncryptedMnemonicResponse struct {
	Mnemonic string `json:"mnemonic"`
}

type ChangeUnlockPasswordRequest struct {
	CurrentUnlockPassword string `json:"currentUnlockPassword"`
	NewUnlockPassword     string `json:"newUnlockPassword"`
}

type ConnectPeerRequest = lnclient.ConnectPeerRequest
type OpenChannelRequest = lnclient.OpenChannelRequest
type OpenChannelResponse = lnclient.OpenChannelResponse
type CloseChannelResponse = lnclient.CloseChannelResponse
type UpdateChannelRequest = lnclient.UpdateChannelRequest

type RedeemOnchainFundsRequest struct {
	ToAddress string `json:"toAddress"`
}

type RedeemOnchainFundsResponse struct {
	TxId string `json:"txId"`
}

type OnchainBalanceResponse = lnclient.OnchainBalanceResponse
type BalancesResponse = lnclient.BalancesResponse

type SendPaymentResponse = Transaction
type MakeInvoiceResponse = Transaction
type LookupInvoiceResponse = Transaction
type ListTransactionsResponse = []Transaction

// TODO: camelCase
type Transaction struct {
	Type            string      `json:"type"`
	Invoice         string      `json:"invoice"`
	Description     string      `json:"description"`
	DescriptionHash string      `json:"description_hash"`
	Preimage        *string     `json:"preimage"`
	PaymentHash     string      `json:"payment_hash"`
	Amount          uint64      `json:"amount"`
	FeesPaid        uint64      `json:"fees_paid"`
	CreatedAt       string      `json:"created_at"`
	SettledAt       *string     `json:"settled_at"`
	AppId           *uint       `json:"app_id"`
	Metadata        interface{} `json:"metadata,omitempty"`
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

type MakeInvoiceRequest struct {
	Amount      int64  `json:"amount"`
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

type NewInstantChannelInvoiceRequest struct {
	Amount  uint64 `json:"amount"`
	LSPType string `json:"lspType"`
	LSPUrl  string `json:"lspUrl"`
	Public  bool   `json:"public"`
}

type NewInstantChannelInvoiceResponse struct {
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
