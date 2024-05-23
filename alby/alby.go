package alby

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/getAlby/nostr-wallet-connect/events"
	"github.com/getAlby/nostr-wallet-connect/models/api"
	"github.com/getAlby/nostr-wallet-connect/models/config"
	"github.com/getAlby/nostr-wallet-connect/nip47"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type AlbyOAuthService struct {
	appConfig *config.AppConfig
	config    config.Config
	oauthConf *oauth2.Config
	logger    *logrus.Logger
	api       api.API
}

// TODO: move to models/alby
type AlbyMe struct {
	Identifier       string `json:"identifier"`
	NPub             string `json:"nostr_pubkey"`
	LightningAddress string `json:"lightning_address"`
	Email            string `json:"email"`
	Name             string `json:"name"`
	Avatar           string `json:"avatar"`
	KeysendPubkey    string `json:"keysend_pubkey"`
	SharedNode       bool   `json:"shared_node"`
}

type AlbyBalance struct {
	Balance  int64  `json:"balance"`
	Unit     string `json:"unit"`
	Currency string `json:"currency"`
}

type ChannelPeerSuggestion struct {
	Network            string `json:"network"`
	PaymentMethod      string `json:"paymentMethod"`
	Pubkey             string `json:"pubkey"`
	Host               string `json:"host"`
	MinimumChannelSize uint64 `json:"minimumChannelSize"`
	Name               string `json:"name"`
	Image              string `json:"image"`
	Lsp                string `json:"lsp"`
}

const (
	accessTokenKey       = "AlbyOAuthAccessToken"
	accessTokenExpiryKey = "AlbyOAuthAccessTokenExpiry"
	refreshTokenKey      = "AlbyOAuthRefreshToken"
	userIdentifierKey    = "AlbyUserIdentifier"
)

func NewAlbyOAuthService(logger *logrus.Logger, kvStore config.Config, appConfig *config.AppConfig, api api.API) *AlbyOAuthService {
	conf := &oauth2.Config{
		ClientID:     appConfig.AlbyClientId,
		ClientSecret: appConfig.AlbyClientSecret,
		Scopes:       []string{"account:read", "balance:read", "payments:send"},
		Endpoint: oauth2.Endpoint{
			TokenURL:  appConfig.AlbyAPIURL + "/oauth/token",
			AuthURL:   appConfig.AlbyOAuthAuthUrl,
			AuthStyle: 2, // use HTTP Basic Authorization https://pkg.go.dev/golang.org/x/oauth2#AuthStyle
		},
	}

	if appConfig.IsDefaultClientId() {
		conf.RedirectURL = "https://getalby.com/hub/callback"
	} else {
		conf.RedirectURL = appConfig.BaseUrl + "/api/alby/callback"
	}

	albyOAuthSvc := &AlbyOAuthService{
		appConfig: appConfig,
		oauthConf: conf,
		config:    kvStore,
		logger:    logger,
		api:       api,
	}
	return albyOAuthSvc
}

func (svc *AlbyOAuthService) CallbackHandler(ctx context.Context, code string) error {
	token, err := svc.oauthConf.Exchange(ctx, code)
	if err != nil {
		svc.logger.WithError(err).Error("Failed to exchange token")
		return err
	}
	svc.saveToken(token)

	existingUserIdentifier, err := svc.GetUserIdentifier()
	if err != nil {
		svc.logger.WithError(err).Error("Failed to get alby user identifier")
		return err
	}

	// setup Alby account on first time login
	if existingUserIdentifier == "" {
		// fetch and save the user's alby account ID. This cannot be changed.
		me, err := svc.GetMe(ctx)
		if err != nil {
			svc.logger.WithError(err).Error("Failed to fetch user me")
			// remove token so user can retry
			svc.config.SetUpdate(accessTokenKey, me.Identifier, "")
			return err
		}

		svc.config.SetUpdate(userIdentifierKey, me.Identifier, "")
	}

	return nil
}

func (svc *AlbyOAuthService) GetUserIdentifier() (string, error) {
	userIdentifier, err := svc.config.Get(userIdentifierKey, "")
	if err != nil {
		svc.logger.WithError(err).Error("Failed to fetch user identifier from user configs")
		return "", err
	}
	return userIdentifier, nil
}

func (svc *AlbyOAuthService) IsConnected(ctx context.Context) bool {
	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		svc.logger.WithError(err).Error("Failed to check fetch token")
	}
	return token != nil
}

func (svc *AlbyOAuthService) saveToken(token *oauth2.Token) {
	svc.config.SetUpdate(accessTokenExpiryKey, strconv.FormatInt(token.Expiry.Unix(), 10), "")
	svc.config.SetUpdate(accessTokenKey, token.AccessToken, "")
	svc.config.SetUpdate(refreshTokenKey, token.RefreshToken, "")
}

var tokenMutex sync.Mutex

func (svc *AlbyOAuthService) fetchUserToken(ctx context.Context) (*oauth2.Token, error) {
	tokenMutex.Lock()
	defer tokenMutex.Unlock()
	accessToken, err := svc.config.Get(accessTokenKey, "")
	if err != nil {
		return nil, err
	}

	if accessToken == "" {
		return nil, nil
	}

	expiry, err := svc.config.Get(accessTokenExpiryKey, "")
	if err != nil {
		return nil, err
	}

	if expiry == "" {
		return nil, nil
	}

	expiry64, err := strconv.ParseInt(expiry, 10, 64)
	if err != nil {
		return nil, err
	}
	refreshToken, err := svc.config.Get(refreshTokenKey, "")
	if err != nil {
		return nil, err
	}

	if refreshToken == "" {
		return nil, nil
	}

	currentToken := &oauth2.Token{
		AccessToken:  accessToken,
		Expiry:       time.Unix(expiry64, 0),
		RefreshToken: refreshToken,
	}

	if currentToken.Expiry.After(time.Now().Add(time.Duration(1) * time.Second)) {
		svc.logger.Info("Using existing Alby OAuth token")
		return currentToken, nil
	}

	newToken, err := svc.oauthConf.TokenSource(ctx, currentToken).Token()
	if err != nil {
		svc.logger.WithError(err).Error("Failed to refresh existing token")
		return nil, err
	}

	svc.saveToken(newToken)
	return newToken, nil
}

func (svc *AlbyOAuthService) GetMe(ctx context.Context) (*AlbyMe, error) {

	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		svc.logger.WithError(err).Error("Failed to fetch user token")
		return nil, err
	}

	client := svc.oauthConf.Client(ctx, token)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/internal/users", svc.appConfig.AlbyAPIURL), nil)
	if err != nil {
		svc.logger.WithError(err).Error("Error creating request /me")
		return nil, err
	}

	req.Header.Set("User-Agent", "NWC-next")

	res, err := client.Do(req)
	if err != nil {
		svc.logger.WithError(err).Error("Failed to fetch /me")
		return nil, err
	}

	me := &AlbyMe{}
	err = json.NewDecoder(res.Body).Decode(me)
	if err != nil {
		svc.logger.WithError(err).Error("Failed to decode API response")
		return nil, err
	}

	svc.logger.WithFields(logrus.Fields{"me": me}).Info("Alby me response")
	return me, nil
}

func (svc *AlbyOAuthService) GetBalance(ctx context.Context) (*AlbyBalance, error) {

	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		svc.logger.WithError(err).Error("Failed to fetch user token")
		return nil, err
	}

	client := svc.oauthConf.Client(ctx, token)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/internal/lndhub/balance", svc.appConfig.AlbyAPIURL), nil)
	if err != nil {
		svc.logger.WithError(err).Error("Error creating request to balance endpoint")
		return nil, err
	}

	req.Header.Set("User-Agent", "NWC-next")

	res, err := client.Do(req)
	if err != nil {
		svc.logger.WithError(err).Error("Failed to fetch balance endpoint")
		return nil, err
	}
	balance := &AlbyBalance{}
	err = json.NewDecoder(res.Body).Decode(balance)
	if err != nil {
		svc.logger.WithError(err).Error("Failed to decode API response")
		return nil, err
	}

	svc.logger.WithFields(logrus.Fields{"balance": balance}).Info("Alby balance response")
	return balance, nil
}

func (svc *AlbyOAuthService) SendPayment(ctx context.Context, invoice string) error {
	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		svc.logger.WithError(err).Error("Failed to fetch user token")
		return err
	}

	client := svc.oauthConf.Client(ctx, token)

	type PayRequest struct {
		Invoice string `json:"invoice"`
	}

	body := bytes.NewBuffer([]byte{})
	payload := &PayRequest{
		Invoice: invoice,
	}
	err = json.NewEncoder(body).Encode(payload)

	if err != nil {
		svc.logger.WithError(err).Error("Failed to encode request payload")
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/internal/lndhub/bolt11", svc.appConfig.AlbyAPIURL), body)
	if err != nil {
		svc.logger.WithError(err).Error("Error creating request bolt11 endpoint")
		return err
	}

	req.Header.Set("User-Agent", "NWC-next")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		svc.logger.WithFields(logrus.Fields{
			"invoice": invoice,
		}).WithError(err).Error("Failed to pay invoice")
		return err
	}

	type PayResponse struct {
		Preimage    string `json:"payment_preimage"`
		PaymentHash string `json:"payment_hash"`
	}

	if resp.StatusCode >= 300 {

		type ErrorResponse struct {
			Error   bool   `json:"error"`
			Code    int    `json:"code"`
			Message string `json:"message"`
		}

		errorPayload := &ErrorResponse{}
		err = json.NewDecoder(resp.Body).Decode(errorPayload)
		if err != nil {
			svc.logger.WithFields(logrus.Fields{
				"status": resp.StatusCode,
			}).WithError(err).Error("Failed to decode payment error response payload")
			return err
		}

		svc.logger.WithFields(logrus.Fields{
			"invoice": invoice,
			"status":  resp.StatusCode,
			"message": errorPayload.Message,
		}).Error("Payment failed")
		return errors.New(errorPayload.Message)
	}

	responsePayload := &PayResponse{}
	err = json.NewDecoder(resp.Body).Decode(responsePayload)
	if err != nil {
		svc.logger.WithError(err).Error("Failed to decode response payload")
		return err
	}
	svc.logger.WithFields(logrus.Fields{
		"invoice":     invoice,
		"paymentHash": responsePayload.PaymentHash,
		"preimage":    responsePayload.Preimage,
	}).Info("Payment successful")
	return nil
}

func (svc *AlbyOAuthService) GetAuthUrl() string {
	if svc.appConfig.AlbyClientId == "" || svc.appConfig.AlbyClientSecret == "" {
		svc.logger.Fatalf("No ALBY_OAUTH_CLIENT_ID or ALBY_OAUTH_CLIENT_SECRET set")
	}
	return svc.oauthConf.AuthCodeURL("unused")
}

func (svc *AlbyOAuthService) LinkAccount(ctx context.Context) error {
	connectionPubkey, err := svc.createAlbyAccountNWCNode(ctx)
	if err != nil {
		svc.logger.WithError(err).Error("Failed to create alby account nwc node")
		return err
	}

	app, err := svc.api.CreateApp(&api.CreateAppRequest{
		Name:           "getalby.com",
		Pubkey:         connectionPubkey,
		MaxAmount:      1_000_000,
		BudgetRenewal:  nip47.BUDGET_RENEWAL_MONTHLY,
		RequestMethods: nip47.CAPABILITIES,
	})

	if err != nil {
		svc.logger.WithError(err).Error("Failed to create app connection")
		return err
	}

	svc.logger.WithFields(logrus.Fields{
		"app": app,
	}).Info("Created alby app connection")

	err = svc.activateAlbyAccountNWCNode(ctx)
	if err != nil {
		svc.logger.WithError(err).Error("Failed to activate alby account nwc node")
		return err
	}

	return nil
}

func (svc *AlbyOAuthService) ConsumeEvent(ctx context.Context, event *events.Event, globalProperties map[string]interface{}) error {
	// TODO: rename this config option to be specific to the alby API
	if !svc.appConfig.LogEvents {
		svc.logger.WithField("event", event).Debug("Skipped sending to alby events API")
		return nil
	}

	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		svc.logger.WithError(err).Error("Failed to fetch user token")
		return err
	}

	client := svc.oauthConf.Client(ctx, token)

	// encode event without global properties
	originalEventBuffer := bytes.NewBuffer([]byte{})
	err = json.NewEncoder(originalEventBuffer).Encode(event)

	if err != nil {
		svc.logger.WithError(err).Error("Failed to encode request payload")
		return err
	}

	type EventWithPropertiesMap struct {
		Event      string                 `json:"event"`
		Properties map[string]interface{} `json:"properties"`
	}

	var eventWithGlobalProperties EventWithPropertiesMap
	err = json.Unmarshal(originalEventBuffer.Bytes(), &eventWithGlobalProperties)
	if err != nil {
		svc.logger.WithError(err).Error("Failed to decode request payload")
		return err
	}
	if eventWithGlobalProperties.Properties == nil {
		eventWithGlobalProperties.Properties = map[string]interface{}{}
	}

	// add global properties to each published event
	for k, v := range globalProperties {
		_, exists := eventWithGlobalProperties.Properties[k]
		if exists {
			svc.logger.WithField("key", k).Error("Key already exists in event properties, skipping global property")
			continue
		}
		eventWithGlobalProperties.Properties[k] = v
	}

	body := bytes.NewBuffer([]byte{})
	err = json.NewEncoder(body).Encode(eventWithGlobalProperties)

	if err != nil {
		svc.logger.WithError(err).Error("Failed to encode request payload")
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/events", svc.appConfig.AlbyAPIURL), body)
	if err != nil {
		svc.logger.WithError(err).Error("Error creating request /events")
		return err
	}

	req.Header.Set("User-Agent", "NWC-next")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		svc.logger.WithFields(logrus.Fields{
			"event": eventWithGlobalProperties,
		}).WithError(err).Error("Failed to send request to /events")
		return err
	}

	if resp.StatusCode >= 300 {
		svc.logger.WithFields(logrus.Fields{
			"event":  eventWithGlobalProperties,
			"status": resp.StatusCode,
		}).Error("Request to /events returned non-success status")
		return errors.New("request to /events returned non-success status")
	}

	return nil
}

func (svc *AlbyOAuthService) createAlbyAccountNWCNode(ctx context.Context) (string, error) {
	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		svc.logger.WithError(err).Error("Failed to fetch user token")
	}

	client := svc.oauthConf.Client(ctx, token)

	type CreateNWCNodeRequest struct {
		WalletPubkey string `json:"wallet_pubkey"`
	}

	createNodeRequest := CreateNWCNodeRequest{
		WalletPubkey: svc.config.GetNostrPublicKey(),
	}

	body := bytes.NewBuffer([]byte{})
	err = json.NewEncoder(body).Encode(createNodeRequest)

	if err != nil {
		svc.logger.WithError(err).Error("Failed to encode request payload")
		return "", err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/internal/nwcs", svc.appConfig.AlbyAPIURL), body)
	if err != nil {
		svc.logger.WithError(err).Error("Error creating request /internal/nwcs")
		return "", err
	}

	req.Header.Set("User-Agent", "NWC-next")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		svc.logger.WithFields(logrus.Fields{
			"createNodeRequest": createNodeRequest,
		}).WithError(err).Error("Failed to send request to /internal/nwcs")
		return "", err
	}

	if resp.StatusCode >= 300 {
		svc.logger.WithFields(logrus.Fields{
			"createNodeRequest": createNodeRequest,
			"status":            resp.StatusCode,
		}).Error("Request to /internal/nwcs returned non-success status")
		return "", errors.New("request to /internal/nwcs returned non-success status")
	}

	type CreateNWCNodeResponse struct {
		Pubkey string `json:"pubkey"`
	}

	responsePayload := &CreateNWCNodeResponse{}
	err = json.NewDecoder(resp.Body).Decode(responsePayload)
	if err != nil {
		svc.logger.WithError(err).Error("Failed to decode response payload")
		return "", err
	}

	svc.logger.WithFields(logrus.Fields{
		"pubkey": responsePayload.Pubkey,
	}).Info("Created alby nwc node successfully")

	return responsePayload.Pubkey, nil
}

func (svc *AlbyOAuthService) activateAlbyAccountNWCNode(ctx context.Context) error {
	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		svc.logger.WithError(err).Error("Failed to fetch user token")
	}

	client := svc.oauthConf.Client(ctx, token)

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/internal/nwcs/activate", svc.appConfig.AlbyAPIURL), nil)
	if err != nil {
		svc.logger.WithError(err).Error("Error creating request /internal/nwcs/activate")
		return err
	}

	req.Header.Set("User-Agent", "NWC-next")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		svc.logger.WithError(err).Error("Failed to send request to /internal/nwcs/activate")
		return err
	}

	if resp.StatusCode >= 300 {
		svc.logger.WithFields(logrus.Fields{
			"status": resp.StatusCode,
		}).Error("Request to /internal/nwcs/activate returned non-success status")
		return errors.New("request to /internal/nwcs/activate returned non-success status")
	}

	svc.logger.Info("Activated alby nwc node successfully")

	return nil
}

func (svc *AlbyOAuthService) GetChannelPeerSuggestions(ctx context.Context) ([]ChannelPeerSuggestion, error) {

	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		svc.logger.WithError(err).Error("Failed to fetch user token")
		return nil, err
	}

	client := svc.oauthConf.Client(ctx, token)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/internal/channel_suggestions", svc.appConfig.AlbyAPIURL), nil)
	if err != nil {
		svc.logger.WithError(err).Error("Error creating request to channel_suggestions endpoint")
		return nil, err
	}

	req.Header.Set("User-Agent", "NWC-next")

	res, err := client.Do(req)
	if err != nil {
		svc.logger.WithError(err).Error("Failed to fetch channel_suggestions endpoint")
		return nil, err
	}
	var suggestions []ChannelPeerSuggestion
	err = json.NewDecoder(res.Body).Decode(&suggestions)
	if err != nil {
		svc.logger.WithError(err).Errorf("Failed to decode API response")
		return nil, err
	}

	svc.logger.WithFields(logrus.Fields{"channel_suggestions": suggestions}).Info("Alby channel peer suggestions response")
	return suggestions, nil
}
