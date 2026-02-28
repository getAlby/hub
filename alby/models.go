package alby

import (
	"context"

	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
)

type AlbyService interface {
	GetInfo(ctx context.Context) (*AlbyInfo, error)
	GetBitcoinRate(ctx context.Context) (*BitcoinRate, error)
	GetChannelPeerSuggestions(ctx context.Context) ([]ChannelPeerSuggestion, error)
}

type AlbyOAuthService interface {
	events.EventSubscriber
	GetLSPChannelOffer(ctx context.Context) (*LSPChannelOffer, error)
	GetLSPInfo(ctx context.Context, lsp, network string) (*LSPInfo, error)
	CreateLSPOrder(ctx context.Context, lsp, network string, lspChannelRequest *LSPChannelRequest) (*LSPChannelResponse, error)
	GetAuthUrl() string
	GetUserIdentifier() (string, error)
	GetLightningAddress() (string, error)
	IsConnected(ctx context.Context) bool
	LinkAccount(ctx context.Context, lnClient lnclient.LNClient, budget uint64, renewal string) error
	CallbackHandler(ctx context.Context, code string) error
	GetMe(ctx context.Context) (*AlbyMe, error)
	UnlinkAccount(ctx context.Context) error
	RequestAutoChannel(ctx context.Context, lnClient lnclient.LNClient, isPublic bool) (*AutoChannelResponse, error)
	GetVssAuthToken(ctx context.Context, nodeIdentifier string) (string, error)
	RemoveOAuthAccessToken() error
	CreateLightningAddress(ctx context.Context, address string, appId uint) (*CreateLightningAddressResponse, error)
	DeleteLightningAddress(ctx context.Context, address string) error
}

type CreateLightningAddressResponse struct {
	Address     string `json:"address"`
	FullAddress string `json:"full_address"`
}

type AlbyLinkAccountRequest struct {
	Budget  uint64 `json:"budget"`
	Renewal string `json:"renewal"`
}

type AutoChannelRequest struct {
	IsPublic bool `json:"isPublic"`
}

type AutoChannelResponse struct {
	Invoice     string `json:"invoice"`
	ChannelSize uint64 `json:"channelSize"`
	Fee         uint64 `json:"fee"`
}

type AlbyInfoHub struct {
	LatestVersion      string `json:"latestVersion"`
	LatestReleaseNotes string `json:"latestReleaseNotes"`
}

type AlbyInfoIncident struct {
	Name    string `json:"name"`
	Started string `json:"started"`
	Status  string `json:"status"`
	Impact  string `json:"impact"`
	Url     string `json:"url"`
}

type AlbyInfo struct {
	Hub              AlbyInfoHub        `json:"hub"`
	Status           string             `json:"status"`
	Healthy          bool               `json:"healthy"`
	AccountAvailable bool               `json:"accountAvailable"` // false if country is blocked (can still use Alby Hub without an Alby Account)
	Incidents        []AlbyInfoIncident `json:"incidents"`
}

type AlbyMeHub struct {
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config"`
}

type AlbyMeSubscription struct {
	PlanCode string `json:"plan_code"`
}

type AlbyMe struct {
	Identifier       string             `json:"identifier"`
	NPub             string             `json:"nostr_pubkey"`
	LightningAddress string             `json:"lightning_address"`
	Email            string             `json:"email"`
	Name             string             `json:"name"`
	Avatar           string             `json:"avatar"`
	KeysendPubkey    string             `json:"keysend_pubkey"`
	SharedNode       bool               `json:"shared_node"`
	Hub              AlbyMeHub          `json:"hub"`
	Subscription     AlbyMeSubscription `json:"subscription"`
}

type ChannelPeerSuggestion struct {
	Network                    string  `json:"network"`
	PaymentMethod              string  `json:"paymentMethod"`
	Pubkey                     string  `json:"pubkey"`
	Host                       string  `json:"host"`
	MinimumChannelSize         uint64  `json:"minimumChannelSize"`
	MaximumChannelSize         uint64  `json:"maximumChannelSize"`
	MaximumChannelExpiryBlocks *uint32 `json:"maximumChannelExpiryBlocks"`
	Name                       string  `json:"name"`
	Image                      string  `json:"image"`
	Identifier                 string  `json:"identifier"`
	ContactUrl                 string  `json:"contactUrl"`
	Type                       string  `json:"type"`
	Terms                      string  `json:"terms"`
	Description                string  `json:"description"`
	Note                       string  `json:"note"`
	PublicChannelsAllowed      bool    `json:"publicChannelsAllowed"`
	FeeTotalSat1m              *uint32 `json:"feeTotalSat1m"`
	FeeTotalSat2m              *uint32 `json:"feeTotalSat2m"`
	FeeTotalSat3m              *uint32 `json:"feeTotalSat3m"`
}

type LSPChannelOffer struct {
	LspName              string `json:"lspName"`
	LspContactUrl        string `json:"lspContactUrl"`
	LspBalanceSat        uint64 `json:"lspBalanceSat"`
	FeeTotalSat          uint64 `json:"feeTotalSat"`
	FeeTotalUsd          uint64 `json:"feeTotalUsd"` // in cents
	CurrentPaymentMethod string `json:"currentPaymentMethod"`
	Terms                string `json:"terms"`
	LspDescription       string `json:"lspDescription"`
}

type BitcoinRate struct {
	Code      string  `json:"code"`
	Symbol    string  `json:"symbol"`
	Rate      string  `json:"rate"`
	RateFloat float64 `json:"rate_float"`
	RateCents int64   `json:"rate_cents"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type LSPChannelPaymentBolt11 struct {
	Invoice     string `json:"invoice"`
	FeeTotalSat string `json:"fee_total_sat"`
}

type LSPChannelPayment struct {
	Bolt11 LSPChannelPaymentBolt11 `json:"bolt11"`
	// TODO: add onchain
}

type LSPChannelResponse struct {
	Payment *LSPChannelPayment `json:"payment"`
}

type LSPChannelRequest struct {
	PublicKey                    string `json:"public_key"`
	LSPBalanceSat                string `json:"lsp_balance_sat"`
	ClientBalanceSat             string `json:"client_balance_sat"`
	RequiredChannelConfirmations uint64 `json:"required_channel_confirmations"`
	FundingConfirmsWithinBlocks  uint64 `json:"funding_confirms_within_blocks"`
	ChannelExpiryBlocks          uint64 `json:"channel_expiry_blocks"`
	Token                        string `json:"token"`
	RefundOnchainAddress         string `json:"refund_onchain_address"`
	AnnounceChannel              bool   `json:"announce_channel"`
}

type LSPInfo struct {
	Pubkey                          string
	Address                         string
	Port                            uint16
	MaxChannelExpiryBlocks          uint64
	MinRequiredChannelConfirmations uint64
	MinFundingConfirmsWithinBlocks  uint64
}
