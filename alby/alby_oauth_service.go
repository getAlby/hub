package alby

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"gorm.io/gorm"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/permissions"
	"github.com/getAlby/hub/service/keys"
	"github.com/getAlby/hub/transactions"
)

type albyOAuthService struct {
	cfg            config.Config
	oauthConf      *oauth2.Config
	db             *gorm.DB
	keys           keys.Keys
	eventPublisher events.EventPublisher
}

const (
	accessTokenKey       = "AlbyOAuthAccessToken"
	accessTokenExpiryKey = "AlbyOAuthAccessTokenExpiry"
	refreshTokenKey      = "AlbyOAuthRefreshToken"
	userIdentifierKey    = "AlbyUserIdentifier"
)

func NewAlbyOAuthService(db *gorm.DB, cfg config.Config, keys keys.Keys, eventPublisher events.EventPublisher) *albyOAuthService {
	conf := &oauth2.Config{
		ClientID:     cfg.GetEnv().AlbyClientId,
		ClientSecret: cfg.GetEnv().AlbyClientSecret,
		Scopes:       []string{"account:read", "balance:read", "payments:send"},
		Endpoint: oauth2.Endpoint{
			TokenURL:  cfg.GetEnv().AlbyAPIURL + "/oauth/token",
			AuthURL:   cfg.GetEnv().AlbyOAuthAuthUrl,
			AuthStyle: 2, // use HTTP Basic Authorization https://pkg.go.dev/golang.org/x/oauth2#AuthStyle
		},
	}

	if cfg.GetEnv().IsDefaultClientId() {
		conf.RedirectURL = "https://getalby.com/hub/callback"
	} else {
		conf.RedirectURL = cfg.GetEnv().BaseUrl + "/api/alby/callback"
	}

	albyOAuthSvc := &albyOAuthService{
		oauthConf:      conf,
		cfg:            cfg,
		db:             db,
		keys:           keys,
		eventPublisher: eventPublisher,
	}
	return albyOAuthSvc
}

func (svc *albyOAuthService) CallbackHandler(ctx context.Context, code string) error {
	token, err := svc.oauthConf.Exchange(ctx, code)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to exchange token")
		return err
	}
	svc.saveToken(token)

	me, err := svc.GetMe(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch user me")
		// remove token so user can retry
		svc.cfg.SetUpdate(accessTokenKey, "", "")
		return err
	}

	existingUserIdentifier, err := svc.GetUserIdentifier()
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to get alby user identifier")
		return err
	}

	// save the user's alby account ID on first time login
	if existingUserIdentifier == "" {
		svc.cfg.SetUpdate(userIdentifierKey, me.Identifier, "")
	} else if me.Identifier != existingUserIdentifier {
		// remove token so user can retry with correct account
		svc.cfg.SetUpdate(accessTokenKey, "", "")
		return errors.New("Alby Hub is connected to a different alby account. Please log out of your Alby Account at getalby.com and try again.")
	}

	return nil
}

func (svc *albyOAuthService) GetUserIdentifier() (string, error) {
	userIdentifier, err := svc.cfg.Get(userIdentifierKey, "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch user identifier from user configs")
		return "", err
	}
	return userIdentifier, nil
}

func (svc *albyOAuthService) IsConnected(ctx context.Context) bool {
	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to check fetch token")
	}
	return token != nil
}

func (svc *albyOAuthService) saveToken(token *oauth2.Token) {
	svc.cfg.SetUpdate(accessTokenExpiryKey, strconv.FormatInt(token.Expiry.Unix(), 10), "")
	svc.cfg.SetUpdate(accessTokenKey, token.AccessToken, "")
	svc.cfg.SetUpdate(refreshTokenKey, token.RefreshToken, "")
}

var tokenMutex sync.Mutex

func (svc *albyOAuthService) fetchUserToken(ctx context.Context) (*oauth2.Token, error) {
	tokenMutex.Lock()
	defer tokenMutex.Unlock()
	accessToken, err := svc.cfg.Get(accessTokenKey, "")
	if err != nil {
		return nil, err
	}

	if accessToken == "" {
		return nil, nil
	}

	expiry, err := svc.cfg.Get(accessTokenExpiryKey, "")
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
	refreshToken, err := svc.cfg.Get(refreshTokenKey, "")
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

	// only use the current token if it has at least 20 seconds before expiry
	if currentToken.Expiry.After(time.Now().Add(time.Duration(20) * time.Second)) {
		logger.Logger.Info("Using existing Alby OAuth token")
		return currentToken, nil
	}

	newToken, err := svc.oauthConf.TokenSource(ctx, currentToken).Token()
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to refresh existing token")
		return nil, err
	}

	svc.saveToken(newToken)
	return newToken, nil
}

func (svc *albyOAuthService) GetMe(ctx context.Context) (*AlbyMe, error) {
	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch user token")
		return nil, err
	}

	client := svc.oauthConf.Client(ctx, token)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/internal/users", svc.cfg.GetEnv().AlbyAPIURL), nil)
	if err != nil {
		logger.Logger.WithError(err).Error("Error creating request /me")
		return nil, err
	}

	req.Header.Set("User-Agent", "NWC-next")

	res, err := client.Do(req)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch /me")
		return nil, err
	}

	me := &AlbyMe{}
	err = json.NewDecoder(res.Body).Decode(me)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to decode API response")
		return nil, err
	}

	logger.Logger.WithFields(logrus.Fields{"me": me}).Info("Alby me response")
	return me, nil
}

func (svc *albyOAuthService) GetBalance(ctx context.Context) (*AlbyBalance, error) {

	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch user token")
		return nil, err
	}

	client := svc.oauthConf.Client(ctx, token)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/internal/lndhub/balance", svc.cfg.GetEnv().AlbyAPIURL), nil)
	if err != nil {
		logger.Logger.WithError(err).Error("Error creating request to balance endpoint")
		return nil, err
	}

	req.Header.Set("User-Agent", "NWC-next")

	res, err := client.Do(req)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch balance endpoint")
		return nil, err
	}
	balance := &AlbyBalance{}
	err = json.NewDecoder(res.Body).Decode(balance)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to decode API response")
		return nil, err
	}

	logger.Logger.WithFields(logrus.Fields{"balance": balance}).Info("Alby balance response")
	return balance, nil
}

func (svc *albyOAuthService) DrainSharedWallet(ctx context.Context, lnClient lnclient.LNClient) error {
	balance, err := svc.GetBalance(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch shared balance")
		return err
	}

	balanceSat := float64(balance.Balance)

	amountSat := int64(math.Floor(
		balanceSat- // Alby shared node balance in sats
			(balanceSat*(8.0/1000.0))- // Alby service fee (0.8%)
			(balanceSat*0.01))) - // Maximum potential routing fees (1%)
		10 // Alby fee reserve (10 sats)

	if amountSat < 1 {
		return errors.New("Not enough balance remaining")
	}
	amount := amountSat * 1000

	logger.Logger.WithField("amount", amount).WithError(err).Error("Draining Alby shared wallet funds")

	transaction, err := transactions.NewTransactionsService(svc.db).MakeInvoice(ctx, amount, "Send shared wallet funds to Alby Hub", "", 120, lnClient, nil, nil)
	if err != nil {
		logger.Logger.WithField("amount", amount).WithError(err).Error("Failed to make invoice")
		return err
	}

	err = svc.SendPayment(ctx, transaction.PaymentRequest)
	if err != nil {
		logger.Logger.WithField("amount", amount).WithError(err).Error("Failed to pay invoice from shared node")
		return err
	}
	return nil
}

func (svc *albyOAuthService) SendPayment(ctx context.Context, invoice string) error {
	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch user token")
		return err
	}

	client := svc.oauthConf.Client(ctx, token)

	type payRequest struct {
		Invoice string `json:"invoice"`
	}

	body := bytes.NewBuffer([]byte{})
	payload := payRequest{
		Invoice: invoice,
	}
	err = json.NewEncoder(body).Encode(&payload)

	if err != nil {
		logger.Logger.WithError(err).Error("Failed to encode request payload")
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/internal/lndhub/bolt11", svc.cfg.GetEnv().AlbyAPIURL), body)
	if err != nil {
		logger.Logger.WithError(err).Error("Error creating request bolt11 endpoint")
		return err
	}

	req.Header.Set("User-Agent", "NWC-next")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
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
			logger.Logger.WithFields(logrus.Fields{
				"status": resp.StatusCode,
			}).WithError(err).Error("Failed to decode payment error response payload")
			return err
		}

		logger.Logger.WithFields(logrus.Fields{
			"invoice": invoice,
			"status":  resp.StatusCode,
			"message": errorPayload.Message,
		}).Error("Payment failed")
		return errors.New(errorPayload.Message)
	}

	responsePayload := &PayResponse{}
	err = json.NewDecoder(resp.Body).Decode(responsePayload)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to decode response payload")
		return err
	}
	logger.Logger.WithFields(logrus.Fields{
		"invoice":     invoice,
		"paymentHash": responsePayload.PaymentHash,
		"preimage":    responsePayload.Preimage,
	}).Info("Payment successful")
	return nil
}

func (svc *albyOAuthService) GetAuthUrl() string {
	if svc.cfg.GetEnv().AlbyClientId == "" || svc.cfg.GetEnv().AlbyClientSecret == "" {
		logger.Logger.Fatalf("No ALBY_OAUTH_CLIENT_ID or ALBY_OAUTH_CLIENT_SECRET set")
	}
	return svc.oauthConf.AuthCodeURL("unused")
}

func (svc *albyOAuthService) LinkAccount(ctx context.Context, lnClient lnclient.LNClient, budget uint64, renewal string) error {
	connectionPubkey, err := svc.createAlbyAccountNWCNode(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to create alby account nwc node")
		return err
	}

	scopes, err := permissions.RequestMethodsToScopes(lnClient.GetSupportedNIP47Methods())
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to get scopes from LNClient request methods")
		return err
	}
	notificationTypes := lnClient.GetSupportedNIP47NotificationTypes()
	if len(notificationTypes) > 0 {
		scopes = append(scopes, constants.NOTIFICATIONS_SCOPE)
	}

	app, _, err := db.NewDBService(svc.db, svc.eventPublisher).CreateApp(
		"getalby.com",
		connectionPubkey,
		budget,
		renewal,
		nil,
		scopes,
		false,
	)

	if err != nil {
		logger.Logger.WithError(err).Error("Failed to create app connection")
		return err
	}

	logger.Logger.WithFields(logrus.Fields{
		"app": app,
	}).Info("Created alby app connection")

	err = svc.activateAlbyAccountNWCNode(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to activate alby account nwc node")
		return err
	}

	return nil
}

func (svc *albyOAuthService) ConsumeEvent(ctx context.Context, event *events.Event, globalProperties map[string]interface{}) {
	// run non-blocking
	go svc.consumeEvent(ctx, event, globalProperties)
}

func (svc *albyOAuthService) consumeEvent(ctx context.Context, event *events.Event, globalProperties map[string]interface{}) {
	// TODO: rename this config option to be specific to the alby API
	if !svc.cfg.GetEnv().LogEvents {
		logger.Logger.WithField("event", event).Debug("Skipped sending to alby events API")
		return
	}

	if event.Event == "nwc_backup_channels" {
		if err := svc.backupChannels(ctx, event); err != nil {
			logger.Logger.WithError(err).Error("Failed to backup channels")
		}
		return
	}

	if event.Event == "nwc_payment_received" {
		type paymentReceivedEventProperties struct {
			PaymentHash string `json:"payment_hash"`
		}
		// pass a new custom event with less detail
		event = &events.Event{
			Event: event.Event,
			Properties: &paymentReceivedEventProperties{
				PaymentHash: event.Properties.(*lnclient.Transaction).PaymentHash,
			},
		}
	}

	if event.Event == "nwc_payment_sent" {
		type paymentSentEventProperties struct {
			PaymentHash string `json:"payment_hash"`
			Duration    uint64 `json:"duration"`
		}

		// pass a new custom event with less detail
		event = &events.Event{
			Event: event.Event,
			Properties: &paymentSentEventProperties{
				PaymentHash: event.Properties.(*lnclient.Transaction).PaymentHash,
				Duration:    uint64(*event.Properties.(*lnclient.Transaction).SettledAt - event.Properties.(*lnclient.Transaction).CreatedAt),
			},
		}
	}

	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch user token")
		return
	}

	client := svc.oauthConf.Client(ctx, token)

	// encode event without global properties
	originalEventBuffer := bytes.NewBuffer([]byte{})
	err = json.NewEncoder(originalEventBuffer).Encode(event)

	if err != nil {
		logger.Logger.WithError(err).Error("Failed to encode request payload")
		return
	}

	type eventWithPropertiesMap struct {
		Event      string                 `json:"event"`
		Properties map[string]interface{} `json:"properties"`
	}

	var eventWithGlobalProperties eventWithPropertiesMap
	err = json.Unmarshal(originalEventBuffer.Bytes(), &eventWithGlobalProperties)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to decode request payload")
		return
	}
	if eventWithGlobalProperties.Properties == nil {
		eventWithGlobalProperties.Properties = map[string]interface{}{}
	}

	// add global properties to each published event
	for k, v := range globalProperties {
		_, exists := eventWithGlobalProperties.Properties[k]
		if exists {
			logger.Logger.WithField("key", k).Error("Key already exists in event properties, skipping global property")
			continue
		}
		eventWithGlobalProperties.Properties[k] = v
	}

	body := bytes.NewBuffer([]byte{})
	err = json.NewEncoder(body).Encode(&eventWithGlobalProperties)

	if err != nil {
		logger.Logger.WithError(err).Error("Failed to encode request payload")
		return
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/events", svc.cfg.GetEnv().AlbyAPIURL), body)
	if err != nil {
		logger.Logger.WithError(err).Error("Error creating request /events")
		return
	}

	req.Header.Set("User-Agent", "NWC-next")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"event": eventWithGlobalProperties,
		}).WithError(err).Error("Failed to send request to /events")
		return
	}

	if resp.StatusCode >= 300 {
		logger.Logger.WithFields(logrus.Fields{
			"event":  eventWithGlobalProperties,
			"status": resp.StatusCode,
		}).Error("Request to /events returned non-success status")
		return
	}
}

func (svc *albyOAuthService) backupChannels(ctx context.Context, event *events.Event) error {
	bkpEvent, ok := event.Properties.(*events.ChannelBackupEvent)
	if !ok {
		return fmt.Errorf("invalid nwc_backup_channels event properties, could not cast to the expected type: %+v", event.Properties)
	}

	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch user token: %w", err)
	}

	client := svc.oauthConf.Client(ctx, token)

	type channelsBackup struct {
		Description string `json:"description"`
		Data        string `json:"data"`
	}

	channelsData := bytes.NewBuffer([]byte{})
	err = json.NewEncoder(channelsData).Encode(bkpEvent.Channels)
	if err != nil {
		return fmt.Errorf("failed to encode channels backup data:  %w", err)
	}

	// use the encrypted mnemonic as the password to encrypt the backup data
	encryptedMnemonic, err := svc.cfg.Get("Mnemonic", "")
	if err != nil {
		return fmt.Errorf("failed to fetch encryption key: %w", err)
	}

	encrypted, err := config.AesGcmEncrypt(channelsData.String(), encryptedMnemonic)
	if err != nil {
		return fmt.Errorf("failed to encrypt channels backup data: %w", err)
	}

	body := bytes.NewBuffer([]byte{})
	err = json.NewEncoder(body).Encode(&channelsBackup{
		Description: "channels",
		Data:        encrypted,
	})
	if err != nil {
		return fmt.Errorf("failed to encode channels backup request payload: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/internal/backups", svc.cfg.GetEnv().AlbyAPIURL), body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "NWC-next")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to /internal/backups: %w", err)
	}

	if resp.StatusCode >= 300 {
		return fmt.Errorf("request to /internal/backups returned non-success status: %d", resp.StatusCode)
	}

	return nil
}

func (svc *albyOAuthService) createAlbyAccountNWCNode(ctx context.Context) (string, error) {
	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch user token")
	}

	client := svc.oauthConf.Client(ctx, token)

	type createNWCNodeRequest struct {
		WalletPubkey string `json:"wallet_pubkey"`
	}

	createNodeRequest := createNWCNodeRequest{
		WalletPubkey: svc.keys.GetNostrPublicKey(),
	}

	body := bytes.NewBuffer([]byte{})
	err = json.NewEncoder(body).Encode(&createNodeRequest)

	if err != nil {
		logger.Logger.WithError(err).Error("Failed to encode request payload")
		return "", err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/internal/nwcs", svc.cfg.GetEnv().AlbyAPIURL), body)
	if err != nil {
		logger.Logger.WithError(err).Error("Error creating request /internal/nwcs")
		return "", err
	}

	req.Header.Set("User-Agent", "NWC-next")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"createNodeRequest": createNodeRequest,
		}).WithError(err).Error("Failed to send request to /internal/nwcs")
		return "", err
	}

	if resp.StatusCode >= 300 {
		logger.Logger.WithFields(logrus.Fields{
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
		logger.Logger.WithError(err).Error("Failed to decode response payload")
		return "", err
	}

	logger.Logger.WithFields(logrus.Fields{
		"pubkey": responsePayload.Pubkey,
	}).Info("Created alby nwc node successfully")

	return responsePayload.Pubkey, nil
}

func (svc *albyOAuthService) activateAlbyAccountNWCNode(ctx context.Context) error {
	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch user token")
	}

	client := svc.oauthConf.Client(ctx, token)

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/internal/nwcs/activate", svc.cfg.GetEnv().AlbyAPIURL), nil)
	if err != nil {
		logger.Logger.WithError(err).Error("Error creating request /internal/nwcs/activate")
		return err
	}

	req.Header.Set("User-Agent", "NWC-next")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to send request to /internal/nwcs/activate")
		return err
	}

	if resp.StatusCode >= 300 {
		logger.Logger.WithFields(logrus.Fields{
			"status": resp.StatusCode,
		}).Error("Request to /internal/nwcs/activate returned non-success status")
		return errors.New("request to /internal/nwcs/activate returned non-success status")
	}

	logger.Logger.Info("Activated alby nwc node successfully")

	return nil
}

func (svc *albyOAuthService) GetChannelPeerSuggestions(ctx context.Context) ([]ChannelPeerSuggestion, error) {

	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch user token")
		return nil, err
	}

	client := svc.oauthConf.Client(ctx, token)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/internal/channel_suggestions", svc.cfg.GetEnv().AlbyAPIURL), nil)
	if err != nil {
		logger.Logger.WithError(err).Error("Error creating request to channel_suggestions endpoint")
		return nil, err
	}

	req.Header.Set("User-Agent", "NWC-next")

	res, err := client.Do(req)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch channel_suggestions endpoint")
		return nil, err
	}
	var suggestions []ChannelPeerSuggestion
	err = json.NewDecoder(res.Body).Decode(&suggestions)
	if err != nil {
		logger.Logger.WithError(err).Errorf("Failed to decode API response")
		return nil, err
	}

	// TODO: remove once alby API is updated
	for i, suggestion := range suggestions {
		if suggestion.BrokenLspType != "" {
			suggestions[i].LspType = suggestion.BrokenLspType
		}
		if suggestion.BrokenLspUrl != "" {
			suggestions[i].LspUrl = suggestion.BrokenLspUrl
		}
	}

	logger.Logger.WithFields(logrus.Fields{"channel_suggestions": suggestions}).Info("Alby channel peer suggestions response")
	return suggestions, nil
}
