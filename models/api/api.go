package api

import (
	"time"

	"github.com/getAlby/nostr-wallet-connect/models/lnclient"
)

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
	MaxAmount      int        `json:"maxAmount"`
	BudgetUsage    int64      `json:"budgetUsage"`
	BudgetRenewal  string     `json:"budgetRenewal"`
}

type ListAppsResponse struct {
	Apps []App `json:"apps"`
}

type UpdateAppRequest struct {
	MaxAmount      int    `json:"maxAmount"`
	BudgetRenewal  string `json:"budgetRenewal"`
	ExpiresAt      string `json:"expiresAt"`
	RequestMethods string `json:"requestMethods"`
}

type CreateAppRequest struct {
	Name           string `json:"name"`
	Pubkey         string `json:"pubkey"`
	MaxAmount      int    `json:"maxAmount"`
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
	LNBackendType string `json:"backendType"`

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
	UnlockPassword  string `json:"unlockPassword"`
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
	OnboardingCompleted  bool   `json:"onboardingCompleted"`
	Running              bool   `json:"running"`
	Unlocked             bool   `json:"unlocked"`
	AlbyAuthUrl          string `json:"albyAuthUrl"`
	ShowBackupReminder   bool   `json:"showBackupReminder"`
	AlbyUserIdentifier   string `json:"albyUserIdentifier"`
	AlbyAccountConnected bool   `json:"albyAccountConnected"`
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
type CloseChannelRequest = lnclient.CloseChannelRequest
type CloseChannelResponse = lnclient.CloseChannelResponse

type NewInstantChannelInvoiceRequest struct {
	Amount uint64 `json:"amount"`
	LSP    string `json:"lsp"`
}

type NewInstantChannelInvoiceResponse struct {
	Invoice string `json:"invoice"`
	Fee     uint64 `json:"fee"`
}

type RedeemOnchainFundsRequest struct {
	ToAddress string `json:"toAddress"`
}

type RedeemOnchainFundsResponse struct {
	TxId string `json:"txId"`
}

type OnchainBalanceResponse = lnclient.OnchainBalanceResponse
type BalancesResponse = lnclient.BalancesResponse

type NewOnchainAddressResponse struct {
	Address string `json:"address"`
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
	NodeID string `json:"nodeID"`
}

type SendSpontaneousPaymentProbesResponse struct {
	Error string `json:"error"`
}

const (
	AppLogOutputSourceGeneral = "general"
	AppLogOutputSourceError   = "error"
)

const (
	LogTypeNode = "node"
	LogTypeApp  = "app"
)

type GetLogOutputRequest struct {
	MaxLen int    `query:"maxLen"`
	Source string `query:"source"`
}

type GetLogOutputResponse struct {
	Log string `json:"logs"`
}

// TODO: move to different file
type AlbyBalanceResponse struct {
	Sats int64 `json:"sats"`
}

type AlbyPayRequest struct {
	Invoice string `json:"invoice"`
}
