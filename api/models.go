package api

import (
	"context"
	"io"
	"time"

	"github.com/getAlby/nostr-wallet-connect/alby"
	"github.com/getAlby/nostr-wallet-connect/db"
	"github.com/getAlby/nostr-wallet-connect/lnclient"
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
	GetNewOnchainAddress(ctx context.Context) (string, error)
	GetUnusedOnchainAddress(ctx context.Context) (string, error)
	SignMessage(ctx context.Context, message string) (*SignMessageResponse, error)
	RedeemOnchainFunds(ctx context.Context, toAddress string) (*RedeemOnchainFundsResponse, error)
	GetBalances(ctx context.Context) (*BalancesResponse, error)
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
}

type App struct {
	// ID          uint      `json:"id"` // ID unused - pubkey is used as ID
	Name        string    `json:"name"`
	Description string    `json:"description"`
	NostrPubkey string    `json:"nostrPubkey"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	LastEventAt    *time.Time `json:"lastEventAt"`
	ExpiresAt      *time.Time `json:"expiresAt"`
	RequestMethods []string   `json:"requestMethods"`
	MaxAmount      uint64     `json:"maxAmount"`
	BudgetUsage    uint64     `json:"budgetUsage"`
	BudgetRenewal  string     `json:"budgetRenewal"`
}

type ListAppsResponse struct {
	Apps []App `json:"apps"`
}

type UpdateAppRequest struct {
	MaxAmount      uint64 `json:"maxAmount"`
	BudgetRenewal  string `json:"budgetRenewal"`
	ExpiresAt      string `json:"expiresAt"`
	RequestMethods string `json:"requestMethods"`
}

type CreateAppRequest struct {
	Name           string `json:"name"`
	Pubkey         string `json:"pubkey"`
	MaxAmount      uint64 `json:"maxAmount"`
	BudgetRenewal  string `json:"budgetRenewal"`
	ExpiresAt      string `json:"expiresAt"`
	RequestMethods string `json:"requestMethods"`
	ReturnTo       string `json:"returnTo"`
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

type RedeemOnchainFundsRequest struct {
	ToAddress string `json:"toAddress"`
}

type RedeemOnchainFundsResponse struct {
	TxId string `json:"txId"`
}

type OnchainBalanceResponse = lnclient.OnchainBalanceResponse
type BalancesResponse = lnclient.BalancesResponse

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
	Amount uint64 `json:"amount"`
	LSP    string `json:"lsp"`
	Public bool   `json:"public"`
}

type NewInstantChannelInvoiceResponse struct {
	Invoice string `json:"invoice"`
	Fee     uint64 `json:"fee"`
}
