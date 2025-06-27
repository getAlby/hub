package alby

import (
	"context"

	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
)

type AlbyOAuthService interface {
	events.EventSubscriber
	GetInfo(ctx context.Context) (*AlbyInfo, error)
	GetChannelPeerSuggestions(ctx context.Context) ([]ChannelPeerSuggestion, error)
	GetBitcoinRate(ctx context.Context) (*BitcoinRate, error)
	GetAuthUrl() string
	GetUserIdentifier() (string, error)
	GetLightningAddress() (string, error)
	IsConnected(ctx context.Context) bool
	LinkAccount(ctx context.Context, lnClient lnclient.LNClient, budget uint64, renewal string) error
	CallbackHandler(ctx context.Context, code string, lnClient lnclient.LNClient) error
	GetBalance(ctx context.Context) (*AlbyBalance, error)
	GetMe(ctx context.Context) (*AlbyMe, error)
	SendPayment(ctx context.Context, invoice string) error
	UnlinkAccount(ctx context.Context) error
	RequestAutoChannel(ctx context.Context, lnClient lnclient.LNClient, isPublic bool) (*AutoChannelResponse, error)
	GetVssAuthToken(ctx context.Context, nodeIdentifier string) (string, error)
	RemoveOAuthAccessToken() error
}

type AlbyBalanceResponse struct {
	Sats int64 `json:"sats"`
}

type AlbyPayRequest struct {
	Invoice string `json:"invoice"`
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

type AlbyBalance struct {
	Balance  int64  `json:"balance"`
	Unit     string `json:"unit"`
	Currency string `json:"currency"`
}

type ChannelPeerSuggestion struct {
	Network               string `json:"network"`
	PaymentMethod         string `json:"paymentMethod"`
	Pubkey                string `json:"pubkey"`
	Host                  string `json:"host"`
	MinimumChannelSize    uint64 `json:"minimumChannelSize"`
	MaximumChannelSize    uint64 `json:"maximumChannelSize"`
	Name                  string `json:"name"`
	Image                 string `json:"image"`
	BrokenLspUrl          string `json:"lsp_url"`
	BrokenLspType         string `json:"lsp_type"`
	LspUrl                string `json:"lspUrl"`
	LspType               string `json:"lspType"`
	Note                  string `json:"note"`
	PublicChannelsAllowed bool   `json:"publicChannelsAllowed"`
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
