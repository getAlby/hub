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

type SetupRequest struct {
	LNBackendType string `json:"backendType"`

	// Breez / Greenlight
	Mnemonic             string `json:"mnemonic"`
	GreenlightInviteCode string `json:"greenlightInviteCode"`

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
	BackendType    string `json:"backendType"`
	SetupCompleted bool   `json:"setupCompleted"`
	Running        bool   `json:"running"`
	Unlocked       bool   `json:"unlocked"`
}

type ConnectPeerRequest = lnclient.ConnectPeerRequest
type OpenChannelRequest = lnclient.OpenChannelRequest
type OpenChannelResponse = lnclient.OpenChannelResponse

type NewWrappedInvoiceRequest struct {
	Amount uint64 `json:"amount"`
	LSP    string `json:"lsp"`
}

type NewWrappedInvoiceResponse struct {
	WrappedInvoice string `json:"wrappedInvoice"`
	Fee            uint64 `json:"fee"`
}

type NewOnchainAddressResponse struct {
	Address string `json:"address"`
}

type OnchainBalanceResponse struct {
	Sats int64 `json:"sats"`
}
