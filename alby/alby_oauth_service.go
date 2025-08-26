package alby

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
	"github.com/tyler-smith/go-bip32"
	"golang.org/x/oauth2"
	"gorm.io/gorm"

	"github.com/getAlby/hub/apps"
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/permissions"
	"github.com/getAlby/hub/service/keys"
	"github.com/getAlby/hub/utils"
	"github.com/getAlby/hub/version"
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
	lightningAddressKey  = "AlbyLightningAddress"
)

const (
	albyOAuthAPIURL    = "https://api.getalby.com"
	albyInternalAPIURL = "https://getalby.com/api"
	albyOAuthAuthUrl   = "https://getalby.com/oauth"
)

const ALBY_ACCOUNT_APP_NAME = "getalby.com"

func NewAlbyOAuthService(db *gorm.DB, cfg config.Config, keys keys.Keys, eventPublisher events.EventPublisher) *albyOAuthService {
	conf := &oauth2.Config{
		ClientID:     cfg.GetEnv().AlbyClientId,
		ClientSecret: cfg.GetEnv().AlbyClientSecret,
		Scopes:       []string{"account:read", "balance:read", "payments:send"},
		Endpoint: oauth2.Endpoint{
			TokenURL:  albyOAuthAPIURL + "/oauth/token",
			AuthURL:   albyOAuthAuthUrl,
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

func (svc *albyOAuthService) RemoveOAuthAccessToken() error {
	err := svc.cfg.SetUpdate(accessTokenKey, "", "")
	if err != nil {
		logger.Logger.WithError(err).Error("failed to remove access token")
	}
	return err
}

func (svc *albyOAuthService) CallbackHandler(ctx context.Context, code string, lnClient lnclient.LNClient) error {
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
		cfgErr := svc.cfg.SetUpdate(accessTokenKey, "", "")
		if cfgErr != nil {
			logger.Logger.WithError(cfgErr).Error("failed to remove existing access token")
		}
		return err
	}

	existingUserIdentifier, err := svc.GetUserIdentifier()
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to get alby user identifier")
		return err
	}

	if existingUserIdentifier == "" {
		// save the user's alby account ID on first time login
		err := svc.cfg.SetUpdate(userIdentifierKey, me.Identifier, "")
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to set user identifier")
			return err
		}
		// notify that this was the first time the user connected their account
		svc.eventPublisher.Publish(&events.Event{
			Event:      "nwc_alby_account_connected",
			Properties: map[string]interface{}{},
		})
	} else if me.Identifier != existingUserIdentifier {
		// remove token so user can retry with correct account
		err := svc.cfg.SetUpdate(accessTokenKey, "", "")
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to set user access token")
		}
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

func (svc *albyOAuthService) GetLightningAddress() (string, error) {
	lightningAddress, err := svc.cfg.Get(lightningAddressKey, "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch lightning address from user configs")
		return "", err
	}
	return lightningAddress, nil
}

func (svc *albyOAuthService) IsConnected(ctx context.Context) bool {
	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to check fetch token")
	}
	return token != nil
}

func (svc *albyOAuthService) saveToken(token *oauth2.Token) {
	err := svc.cfg.SetUpdate(accessTokenExpiryKey, strconv.FormatInt(token.Expiry.Unix(), 10), "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to save access token expiry")
	}
	err = svc.cfg.SetUpdate(accessTokenKey, token.AccessToken, "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to save access token")
	}
	err = svc.cfg.SetUpdate(refreshTokenKey, token.RefreshToken, "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to save refresh token")
	}
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

	// only use the current token if it has at least 60 seconds before expiry
	if currentToken.Expiry.After(time.Now().Add(time.Duration(60) * time.Second)) {
		logger.Logger.Debug("Using existing Alby OAuth token")
		return currentToken, nil
	}

	newToken, err := svc.oauthConf.TokenSource(ctx, currentToken).Token()
	if err != nil {
		logger.Logger.WithError(err).Warn("Failed to refresh existing token")
		return nil, err
	}

	svc.saveToken(newToken)
	return newToken, nil
}

func (svc *albyOAuthService) GetInfo(ctx context.Context) (*AlbyInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/internal/info", albyInternalAPIURL), nil)
	if err != nil {
		logger.Logger.WithError(err).Error("Error creating request to alby info endpoint")
		return nil, err
	}

	setDefaultRequestHeaders(req)

	res, err := client.Do(req)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch /info")
		return nil, err
	}

	type albyInfoHub struct {
		LatestVersion      string `json:"latest_version"`
		LatestReleaseNotes string `json:"latest_release_notes"`
	}

	type albyInfoIncident struct {
		Name    string `json:"name"`
		Started string `json:"started"`
		Status  string `json:"status"`
		Impact  string `json:"impact"`
		Url     string `json:"url"`
	}

	type albyInfo struct {
		Hub              albyInfoHub        `json:"hub"`
		Status           string             `json:"status"`
		Healthy          bool               `json:"healthy"`
		AccountAvailable bool               `json:"account_available"` // false if country is blocked (can still use Alby Hub without an Alby Account)
		Incidents        []albyInfoIncident `json:"incidents"`
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to read response body")
		return nil, errors.New("failed to read response body")
	}

	if res.StatusCode >= 300 {
		logger.Logger.WithFields(logrus.Fields{
			"body":        string(body),
			"status_code": res.StatusCode,
		}).Error("info endpoint returned non-success code")
		return nil, fmt.Errorf("info endpoint returned non-success code: %s", string(body))
	}

	info := &albyInfo{}
	err = json.Unmarshal(body, info)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to decode API response")
		return nil, err
	}

	incidents := []AlbyInfoIncident{}
	for _, incident := range info.Incidents {
		incidents = append(incidents, AlbyInfoIncident{
			Name:    incident.Name,
			Started: incident.Started,
			Status:  incident.Status,
			Impact:  incident.Impact,
			Url:     incident.Url,
		})
	}

	return &AlbyInfo{
		Hub: AlbyInfoHub{
			LatestVersion:      info.Hub.LatestVersion,
			LatestReleaseNotes: info.Hub.LatestReleaseNotes,
		},
		Status:           info.Status,
		Healthy:          info.Healthy,
		AccountAvailable: info.AccountAvailable,
		Incidents:        incidents,
	}, nil
}

func (svc *albyOAuthService) GetVssAuthToken(ctx context.Context, nodeIdentifier string) (string, error) {
	logger.Logger.WithField("node_identifier", nodeIdentifier).Debug("fetching VSS token")
	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch user token")
		return "", err
	}

	client := svc.oauthConf.Client(ctx, token)
	client.Timeout = 10 * time.Second

	type vssAuthTokenRequest struct {
		Identifier string `json:"identifier"`
	}

	body := bytes.NewBuffer([]byte{})
	payload := vssAuthTokenRequest{
		Identifier: nodeIdentifier,
	}
	err = json.NewEncoder(body).Encode(&payload)

	if err != nil {
		logger.Logger.WithError(err).Error("Failed to encode request payload")
		return "", err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/internal/auth_tokens", albyOAuthAPIURL), body)
	if err != nil {
		logger.Logger.WithError(err).Error("Error creating request for vss auth token endpoint")
		return "", err
	}

	setDefaultRequestHeaders(req)

	res, err := client.Do(req)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch vss auth token endpoint")
		return "", err
	}

	if res.StatusCode >= 300 {
		return "", fmt.Errorf("request to /internal/auth_tokens returned non-success status: %d", res.StatusCode)
	}

	type vssTokenResponse struct {
		Token string `json:"token"`
	}

	vssResponse := &vssTokenResponse{}
	err = json.NewDecoder(res.Body).Decode(vssResponse)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to decode API response")
		return "", err
	}

	if vssResponse.Token == "" {
		logger.Logger.WithField("vss_response", vssResponse).WithError(err).Error("No token in API response")
		return "", errors.New("no token in vss response")
	}

	return vssResponse.Token, nil
}

func (svc *albyOAuthService) CreateLightningAddress(ctx context.Context, address string, appId uint) (*CreateLightningAddressResponse, error) {
	logger.Logger.WithFields(logrus.Fields{
		"address": address,
		"app_id":  appId,
	}).Debug("creating lightning address")
	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch user token")
		return nil, err
	}

	client := svc.oauthConf.Client(ctx, token)
	client.Timeout = 10 * time.Second

	type createLightningAddressRequest struct {
		Address string `json:"address"`
		AppId   uint   `json:"app_id"`
	}

	body := bytes.NewBuffer([]byte{})
	payload := createLightningAddressRequest{
		Address: address,
		AppId:   appId,
	}
	err = json.NewEncoder(body).Encode(&payload)

	if err != nil {
		logger.Logger.WithError(err).Error("Failed to encode request payload")
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/internal/lightning_addresses", albyOAuthAPIURL), body)
	if err != nil {
		logger.Logger.WithError(err).Error("Error creating request for vss auth token endpoint")
		return nil, err
	}

	setDefaultRequestHeaders(req)

	res, err := client.Do(req)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch vss auth token endpoint")
		return nil, err
	}

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to read response body")
		return nil, errors.New("failed to read response body")
	}

	if res.StatusCode == 422 {
		type createLightningAddressErrors struct {
			Address []string `json:"address"`
		}
		lightningAddressErrors := &createLightningAddressErrors{}
		err = json.Unmarshal(responseBody, lightningAddressErrors)
		if err != nil {
			logger.Logger.WithError(err).Error("failed to unmarshal errors response")
			return nil, err
		}
		if len(lightningAddressErrors.Address) == 0 {
			return nil, errors.New("unknown error occurred")
		}
		return nil, errors.New(lightningAddressErrors.Address[0])
	}

	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("POST request to /internal/lightning_addresses/%s returned non-success status: %d %s", address, res.StatusCode, string(responseBody))
	}

	createLightningAddressResponse := &CreateLightningAddressResponse{}
	err = json.Unmarshal(responseBody, createLightningAddressResponse)
	if err != nil {
		logger.Logger.WithError(err).Error("failed to unmarshal response")
		return nil, err
	}

	return createLightningAddressResponse, nil
}

func (svc *albyOAuthService) DeleteLightningAddress(ctx context.Context, address string) error {
	logger.Logger.WithFields(logrus.Fields{
		"address": address,
	}).Debug("deleting lightning address")
	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch user token")
		return err
	}

	client := svc.oauthConf.Client(ctx, token)
	client.Timeout = 10 * time.Second

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/internal/lightning_addresses/%s", albyOAuthAPIURL, address), nil)
	if err != nil {
		logger.Logger.WithError(err).Error("Error creating request for delete lightning address endpoint")
		return err
	}

	setDefaultRequestHeaders(req)

	res, err := client.Do(req)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to delete lightning address endpoint")
		return err
	}

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to read response body")
		return errors.New("failed to read response body")
	}

	if res.StatusCode >= 300 {
		return fmt.Errorf("DELETE request to /internal/lightning_addresses/%s returned non-success status: %d %s", address, res.StatusCode, string(responseBody))
	}

	return nil
}

func (svc *albyOAuthService) GetMe(ctx context.Context) (*AlbyMe, error) {
	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch user token")
		return nil, err
	}

	client := svc.oauthConf.Client(ctx, token)
	client.Timeout = 10 * time.Second

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/internal/users", albyOAuthAPIURL), nil)
	if err != nil {
		logger.Logger.WithError(err).Error("Error creating request /me")
		return nil, err
	}

	setDefaultRequestHeaders(req)

	res, err := client.Do(req)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch /me")
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to read response body")
		return nil, errors.New("failed to read response body")
	}

	if res.StatusCode >= 300 {
		logger.Logger.WithFields(logrus.Fields{
			"body":        string(body),
			"status_code": res.StatusCode,
		}).Error("users endpoint returned non-success code")
		return nil, fmt.Errorf("users endpoint returned non-success code: %s", string(body))
	}

	me := &AlbyMe{}
	err = json.Unmarshal(body, me)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to decode API response")
		return nil, err
	}

	err = svc.cfg.SetUpdate(lightningAddressKey, me.LightningAddress, "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to save lightning address")
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
	client.Timeout = 10 * time.Second

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/internal/lndhub/balance", albyOAuthAPIURL), nil)
	if err != nil {
		logger.Logger.WithError(err).Error("Error creating request to balance endpoint")
		return nil, err
	}

	setDefaultRequestHeaders(req)

	res, err := client.Do(req)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch balance endpoint")
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to read response body")
		return nil, errors.New("failed to read response body")
	}

	if res.StatusCode >= 300 {
		logger.Logger.WithFields(logrus.Fields{
			"body":        string(body),
			"status_code": res.StatusCode,
		}).Error("balance endpoint returned non-success code")
		return nil, fmt.Errorf("balance endpoint returned non-success code: %s", string(body))
	}

	balance := &AlbyBalance{}
	err = json.Unmarshal(body, balance)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to decode API response")
		return nil, err
	}

	logger.Logger.WithFields(logrus.Fields{"balance": balance}).Debug("Alby balance response")
	return balance, nil
}

func (svc *albyOAuthService) SendPayment(ctx context.Context, invoice string) error {
	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch user token")
		return err
	}

	client := svc.oauthConf.Client(ctx, token)
	client.Timeout = 10 * time.Second

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

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/internal/lndhub/bolt11", albyOAuthAPIURL), body)
	if err != nil {
		logger.Logger.WithError(err).Error("Error creating request bolt11 endpoint")
		return err
	}

	setDefaultRequestHeaders(req)

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
	}).Info("Alby Payment successful")
	return nil
}

func (svc *albyOAuthService) GetAuthUrl() string {
	return svc.oauthConf.AuthCodeURL("unused")
}

func (svc *albyOAuthService) UnlinkAccount(ctx context.Context) error {
	ldkVssEnabled, err := svc.cfg.Get("LdkVssEnabled", "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch LdkVssEnabled user config")
		return err
	}

	if ldkVssEnabled == "true" {
		return errors.New("alby account cannot be unlinked while VSS is activated")
	}

	destroyAlbyAccountErr := svc.destroyAlbyAccountNWCNode(ctx)
	if destroyAlbyAccountErr != nil {
		// non-critical error - we still want to disconnect
		logger.Logger.WithError(err).Error("Failed to destroy Alby Account NWC node")
	}
	svc.deleteAlbyAccountApps()

	err = svc.cfg.SetUpdate(userIdentifierKey, "", "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to remove user identifier from config")
	}
	err = svc.cfg.SetUpdate(accessTokenKey, "", "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to remove access token from config")
	}
	err = svc.cfg.SetUpdate(accessTokenExpiryKey, "", "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to remove access token expiry from config")
	}
	err = svc.cfg.SetUpdate(refreshTokenKey, "", "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to remove refresh token from config")
	}
	err = svc.cfg.SetUpdate(lightningAddressKey, "", "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to remove lightning address from config")
	}

	return nil
}

func (svc *albyOAuthService) LinkAccount(ctx context.Context, lnClient lnclient.LNClient, budget uint64, renewal string) error {
	svc.deleteAlbyAccountApps()

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

	app, _, err := apps.NewAppsService(svc.db, svc.eventPublisher, svc.keys, svc.cfg).CreateApp(
		ALBY_ACCOUNT_APP_NAME,
		connectionPubkey,
		budget,
		renewal,
		nil,
		scopes,
		false,
		nil,
	)

	if err != nil {
		logger.Logger.WithError(err).Error("Failed to create app connection")
		return err
	}

	logger.Logger.WithFields(logrus.Fields{
		"app": app,
	}).Info("Created alby app connection")

	err = svc.activateAlbyAccountNWCNode(ctx, *app.WalletPubkey)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to activate alby account nwc node")
		return err
	}

	return nil
}

func (svc *albyOAuthService) ConsumeEvent(ctx context.Context, event *events.Event, globalProperties map[string]interface{}) {
	defer func() {
		// ensure the app cannot panic if firing events to Alby API fails
		if r := recover(); r != nil {
			logger.Logger.WithField("event", event).WithField("r", r).Error("Failed to consume event in alby oauth service")
		}
	}()

	accessToken, err := svc.cfg.Get(accessTokenKey, "")
	if err != nil {
		logger.Logger.WithError(err).Error("failed to get access token from config")
		return
	}

	if accessToken == "" {
		logger.Logger.WithFields(logrus.Fields{
			"event": event,
		}).Debug("user has not authed yet, skipping event")
		return
	}

	if !svc.cfg.GetEnv().SendEventsToAlby {
		logger.Logger.WithField("event", event).Debug("Skipped sending to alby events API (alby event logging disabled)")
		return
	}

	// ensure we do not send unintended events to Alby API
	if !slices.Contains(getEventWhitelist(), event.Event) {
		logger.Logger.WithField("event", event).Debug("Skipped sending non-whitelisted event to alby events API")
		return
	}

	if event.Event == "nwc_backup_channels" {
		// if backup fails, try again (max 3 attempts)
		for i := 0; i < 3; i++ {
			if err := svc.backupChannels(ctx, event); err != nil {
				logger.Logger.WithField("attempt", i).WithError(err).Error("Failed to backup channels")
				continue
			}
			break
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
				PaymentHash: event.Properties.(*db.Transaction).PaymentHash,
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
				PaymentHash: event.Properties.(*db.Transaction).PaymentHash,
				Duration:    uint64(event.Properties.(*db.Transaction).SettledAt.Unix() - event.Properties.(*db.Transaction).CreatedAt.Unix()),
			},
		}
	}

	if event.Event == "nwc_payment_failed" {
		transaction, ok := event.Properties.(*db.Transaction)
		if !ok {
			logger.Logger.WithField("event", event).Error("Failed to cast event")
			return
		}

		type paymentFailedEventProperties struct {
			PaymentHash string `json:"payment_hash"`
			Reason      string `json:"reason"`
		}

		// pass a new custom event with less detail
		event = &events.Event{
			Event: event.Event,
			Properties: &paymentFailedEventProperties{
				PaymentHash: transaction.PaymentHash,
				Reason:      transaction.FailureReason,
			},
		}
	}

	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch user token")
		return
	}

	client := svc.oauthConf.Client(ctx, token)
	client.Timeout = 10 * time.Second

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
			logger.Logger.WithField("key", k).Debug("Key already exists in event properties, skipping global property")
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

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/events", albyOAuthAPIURL), body)
	if err != nil {
		logger.Logger.WithError(err).Error("Error creating request /events")
		return
	}

	setDefaultRequestHeaders(req)

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

type channelsBackup struct {
	Description string `json:"description"`
	Data        string `json:"data"`
	NodePubkey  string `json:"node_pubkey"`
}

func (svc *albyOAuthService) createEncryptedChannelBackup(event *events.StaticChannelsBackupEvent) (*channelsBackup, error) {

	eventData := bytes.NewBuffer([]byte{})
	err := json.NewEncoder(eventData).Encode(event)
	if err != nil {
		return nil, fmt.Errorf("failed to encode channels backup data:  %w", err)
	}

	backupKey, err := svc.keys.DeriveKey([]uint32{bip32.FirstHardenedChild})
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to generate channels backup key")
		return nil, err
	}

	encrypted, err := config.AesGcmEncryptWithKey(eventData.String(), backupKey.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt channels backup data: %w", err)
	}

	backup := &channelsBackup{
		Description: "channels_v2",
		Data:        encrypted,
		NodePubkey:  event.NodeID,
	}
	return backup, nil
}

func (svc *albyOAuthService) backupChannels(ctx context.Context, event *events.Event) error {
	bkpEvent, ok := event.Properties.(*events.StaticChannelsBackupEvent)
	if !ok {
		return fmt.Errorf("invalid nwc_backup_channels event properties, could not cast to the expected type: %+v", event.Properties)
	}

	backup, err := svc.createEncryptedChannelBackup(bkpEvent)
	if err != nil {
		return fmt.Errorf("failed to encrypt channel backup: %w", err)
	}

	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch user token: %w", err)
	}

	client := svc.oauthConf.Client(ctx, token)
	client.Timeout = 10 * time.Second

	body := bytes.NewBuffer([]byte{})
	err = json.NewEncoder(body).Encode(backup)
	if err != nil {
		return fmt.Errorf("failed to encode channels backup request payload: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/internal/backups", albyOAuthAPIURL), body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	setDefaultRequestHeaders(req)

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
	client.Timeout = 10 * time.Second

	type createNWCNodeRequest struct {
	}

	createNodeRequest := createNWCNodeRequest{}

	body := bytes.NewBuffer([]byte{})
	err = json.NewEncoder(body).Encode(&createNodeRequest)

	if err != nil {
		logger.Logger.WithError(err).Error("Failed to encode request payload")
		return "", err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/internal/nwcs", albyOAuthAPIURL), body)
	if err != nil {
		logger.Logger.WithError(err).Error("Error creating request /internal/nwcs")
		return "", err
	}

	setDefaultRequestHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to send request to /internal/nwcs")
		return "", err
	}

	if resp.StatusCode >= 300 {
		logger.Logger.WithFields(logrus.Fields{
			"status": resp.StatusCode,
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

func (svc *albyOAuthService) destroyAlbyAccountNWCNode(ctx context.Context) error {
	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch user token")
	}

	client := svc.oauthConf.Client(ctx, token)
	client.Timeout = 10 * time.Second

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/internal/nwcs", albyOAuthAPIURL), nil)
	if err != nil {
		logger.Logger.WithError(err).Error("Error creating request /internal/nwcs")
		return err
	}

	setDefaultRequestHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to send request to /internal/nwcs")
		return err
	}

	if resp.StatusCode >= 300 {
		logger.Logger.WithFields(logrus.Fields{
			"status": resp.StatusCode,
		}).Error("Request to /internal/nwcs returned non-success status")
		return errors.New("request to /internal/nwcs returned non-success status")
	}

	logger.Logger.Info("Removed alby account nwc node successfully")

	return nil
}

func (svc *albyOAuthService) activateAlbyAccountNWCNode(ctx context.Context, walletServicePubkey string) error {
	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch user token")
	}

	client := svc.oauthConf.Client(ctx, token)
	client.Timeout = 10 * time.Second

	type activateNWCNodeRequest struct {
		WalletPubkey string `json:"wallet_pubkey"`
		RelayUrl     string `json:"relay_url"`
	}

	activateNodeRequest := activateNWCNodeRequest{
		WalletPubkey: walletServicePubkey,
		RelayUrl:     svc.cfg.GetRelayUrl(),
	}

	body := bytes.NewBuffer([]byte{})
	err = json.NewEncoder(body).Encode(&activateNodeRequest)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/internal/nwcs/activate", albyOAuthAPIURL), body)
	if err != nil {
		logger.Logger.WithError(err).Error("Error creating request /internal/nwcs/activate")
		return err
	}

	setDefaultRequestHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"activate_node_request": activateNodeRequest,
		}).WithError(err).Error("Failed to send request to /internal/nwcs/activate")
		return err
	}

	if resp.StatusCode >= 300 {
		bodyString := ""
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"activate_node_request": activateNodeRequest,
				"status":                resp.StatusCode,
			}).Error("Failed to read response body from to /internal/nwcs/activate")
		}
		if bodyBytes != nil {
			bodyString = string(bodyBytes)
		}

		logger.Logger.WithFields(logrus.Fields{
			"status":  resp.StatusCode,
			"message": bodyString,
		}).Error("Request to /internal/nwcs/activate returned non-success status")
		return errors.New("request to /internal/nwcs/activate returned non-success status")
	}

	logger.Logger.Info("Activated alby nwc node successfully")

	return nil
}

func (svc *albyOAuthService) GetChannelPeerSuggestions(ctx context.Context) ([]ChannelPeerSuggestion, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/internal/channel_suggestions", albyInternalAPIURL), nil)
	if err != nil {
		logger.Logger.WithError(err).Error("Error creating request to channel_suggestions endpoint")
		return nil, err
	}

	setDefaultRequestHeaders(req)

	res, err := client.Do(req)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch channel_suggestions endpoint")
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to read response body")
		return nil, errors.New("failed to read response body")
	}

	if res.StatusCode >= 300 {
		logger.Logger.WithFields(logrus.Fields{
			"body":        string(body),
			"status_code": res.StatusCode,
		}).Error("channel suggestions endpoint returned non-success code")
		return nil, fmt.Errorf("channel suggestions endpoint returned non-success code: %s", string(body))
	}

	var suggestions []ChannelPeerSuggestion
	err = json.Unmarshal(body, &suggestions)
	if err != nil {
		logger.Logger.WithError(err).Errorf("Failed to decode API response")
		return nil, err
	}

	logger.Logger.WithFields(logrus.Fields{"channel_suggestions": suggestions}).Debug("Alby channel peer suggestions response")
	return suggestions, nil
}

func (svc *albyOAuthService) GetLSPChannelOffer(ctx context.Context) (*LSPChannelOffer, error) {
	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch user token")
		return nil, err
	}

	client := svc.oauthConf.Client(ctx, token)
	client.Timeout = 10 * time.Second

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/internal/lsp", albyOAuthAPIURL), nil)
	if err != nil {
		logger.Logger.WithError(err).Error("Error creating request /me")
		return nil, err
	}

	setDefaultRequestHeaders(req)

	res, err := client.Do(req)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch /me")
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to read response body")
		return nil, errors.New("failed to read response body")
	}

	if res.StatusCode >= 300 {
		logger.Logger.WithFields(logrus.Fields{
			"body":        string(body),
			"status_code": res.StatusCode,
		}).Error("users endpoint returned non-success code")
		return nil, fmt.Errorf("users endpoint returned non-success code: %s", string(body))
	}

	lspChannelOffer := &LSPChannelOffer{}
	err = json.Unmarshal(body, lspChannelOffer)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to decode API response")
		return nil, err
	}

	return lspChannelOffer, nil
}

func (svc *albyOAuthService) GetBitcoinRate(ctx context.Context) (*BitcoinRate, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	currency := svc.cfg.GetCurrency()

	url := fmt.Sprintf("%s/rates/%s", albyInternalAPIURL, currency)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"currency": currency,
			"error":    err,
		}).Error("Error creating request to Bitcoin rate endpoint")
		return nil, err
	}
	setDefaultRequestHeaders(req)

	res, err := client.Do(req)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"currency": currency,
			"error":    err,
		}).Error("Failed to fetch Bitcoin rate from API")
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to read response body")
		return nil, errors.New("failed to read response body")
	}

	if res.StatusCode >= 300 {
		logger.Logger.WithFields(logrus.Fields{
			"currency":    currency,
			"body":        string(body),
			"status_code": res.StatusCode,
		}).Error("Bitcoin rate endpoint returned non-success code")
		return nil, fmt.Errorf("bitcoin rate endpoint returned non-success code: %s", string(body))
	}

	var rate = &BitcoinRate{}
	err = json.Unmarshal(body, rate)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"currency": currency,
			"body":     string(body),
			"error":    err,
		}).Error("Failed to decode Bitcoin rate API response")
		return nil, err
	}

	return rate, nil
}

func (svc *albyOAuthService) RequestAutoChannel(ctx context.Context, lnClient lnclient.LNClient, isPublic bool) (*AutoChannelResponse, error) {
	nodeInfo, err := lnClient.GetInfo(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to request own node info", err)
		return nil, err
	}

	requestUrl := fmt.Sprintf("https://api.getalby.com/internal/lsp/alby/%s", nodeInfo.Network)

	pubkey, address, port, err := svc.getLSPInfo(ctx, requestUrl+"/v1/get_info")

	if err != nil {
		logger.Logger.WithError(err).Error("Failed to request LSP info")
		return nil, err
	}

	err = lnClient.ConnectPeer(ctx, &lnclient.ConnectPeerRequest{
		Pubkey:  pubkey,
		Address: address,
		Port:    port,
	})

	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"pubkey":  pubkey,
			"address": address,
			"port":    port,
		}).WithError(err).Error("Failed to connect to peer")
		return nil, err
	}

	logger.Logger.WithFields(logrus.Fields{
		"pubkey": pubkey,
		"public": isPublic,
	}).Info("Requesting auto channel")

	autoChannelResponse, err := svc.requestAutoChannel(ctx, requestUrl+"/auto_channel", nodeInfo.Pubkey, isPublic)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to request auto channel")
		return nil, err
	}
	return autoChannelResponse, nil
}

func (svc *albyOAuthService) requestAutoChannel(ctx context.Context, url string, pubkey string, isPublic bool) (*AutoChannelResponse, error) {
	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch user token")
	}

	client := svc.oauthConf.Client(ctx, token)
	client.Timeout = 60 * time.Second

	type autoChannelRequest struct {
		PublicKey       string `json:"public_key"`
		AnnounceChannel bool   `json:"announce_channel"`
		NodeType        string `json:"node_type"`
	}

	backendType, err := svc.cfg.Get("LNBackendType", "")
	if err != nil {
		return nil, errors.New("failed to get LN backend type")
	}
	newAutoChannelRequest := autoChannelRequest{
		PublicKey:       pubkey,
		AnnounceChannel: isPublic,
		NodeType:        backendType,
	}

	payloadBytes, err := json.Marshal(newAutoChannelRequest)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest(http.MethodPost, url, bodyReader)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to create auto channel request")
		return nil, err
	}

	setDefaultRequestHeaders(req)

	res, err := client.Do(req)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to request auto channel invoice")
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to read response body")
		return nil, errors.New("failed to read response body")
	}

	if res.StatusCode >= 300 {
		logger.Logger.WithFields(logrus.Fields{
			"request":     newAutoChannelRequest,
			"body":        string(body),
			"status_code": res.StatusCode,
		}).Error("auto channel endpoint returned non-success code")
		return nil, fmt.Errorf("auto channel endpoint returned non-success code: %s", string(body))
	}

	type newLSPS1ChannelPaymentBolt11 struct {
		Invoice     string `json:"invoice"`
		FeeTotalSat string `json:"fee_total_sat"`
	}

	type newLSPS1ChannelPayment struct {
		Bolt11 newLSPS1ChannelPaymentBolt11 `json:"bolt11"`
		// TODO: add onchain
	}
	type autoChannelResponse struct {
		LspBalanceSat string                  `json:"lsp_balance_sat"`
		Payment       *newLSPS1ChannelPayment `json:"payment"`
	}

	var newAutoChannelResponse autoChannelResponse

	err = json.Unmarshal(body, &newAutoChannelResponse)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to deserialize json")
		return nil, fmt.Errorf("failed to deserialize json %s %s", url, string(body))
	}

	var invoice string
	var fee uint64

	if newAutoChannelResponse.Payment != nil {
		invoice = newAutoChannelResponse.Payment.Bolt11.Invoice
		fee, err = strconv.ParseUint(newAutoChannelResponse.Payment.Bolt11.FeeTotalSat, 10, 64)
		if err != nil {
			logger.Logger.WithError(err).WithFields(logrus.Fields{
				"url": url,
			}).Error("Failed to parse fee")
			return nil, fmt.Errorf("failed to parse fee %v", err)
		}

		paymentRequest, err := decodepay.Decodepay(invoice)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to decode bolt11 invoice")
			return nil, err
		}

		if fee != uint64(paymentRequest.MSatoshi/1000) {
			logger.Logger.WithFields(logrus.Fields{
				"invoice_amount": paymentRequest.MSatoshi / 1000,
				"fee":            fee,
			}).WithError(err).Error("Invoice amount does not match LSP fee")
			return nil, errors.New("invoice amount does not match LSP fee")
		}
	}

	channelSize, err := strconv.ParseUint(newAutoChannelResponse.LspBalanceSat, 10, 64)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to parse lsp balance sat")
		return nil, fmt.Errorf("failed to parse lsp balance sat %v", err)
	}

	return &AutoChannelResponse{
		Invoice:     invoice,
		Fee:         fee,
		ChannelSize: channelSize,
	}, nil
}

func (svc *albyOAuthService) getLSPInfo(ctx context.Context, url string) (pubkey string, address string, port uint16, err error) {

	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch user token")
	}

	client := svc.oauthConf.Client(ctx, token)
	client.Timeout = 60 * time.Second

	type lsps1LSPInfo struct {
		URIs []string `json:"uris"`
	}
	var lsps1LspInfo lsps1LSPInfo

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to create lsp info request")
		return "", "", uint16(0), err
	}

	setDefaultRequestHeaders(req)

	res, err := client.Do(req)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to request lsp info")
		return "", "", uint16(0), err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to read response body")
		return "", "", uint16(0), errors.New("failed to read response body")
	}

	err = json.Unmarshal(body, &lsps1LspInfo)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to deserialize json")
		return "", "", uint16(0), fmt.Errorf("failed to deserialize json %s %s", url, string(body))
	}

	httpUris := utils.Filter(lsps1LspInfo.URIs, func(uri string) bool {
		return !strings.Contains(uri, ".onion")
	})
	if len(httpUris) == 0 {
		logger.Logger.WithField("uris", lsps1LspInfo.URIs).WithError(err).Error("Couldn't find HTTP URI")

		return "", "", uint16(0), err
	}
	uri := httpUris[0]

	// make sure it's a valid IPv4 URI
	regex := regexp.MustCompile(`^([0-9a-f]+)@([0-9]+\.[0-9]+\.[0-9]+\.[0-9]+):([0-9]+)$`)
	parts := regex.FindStringSubmatch(uri)
	logger.Logger.WithField("parts", parts).Debug("Split URI")
	if parts == nil || len(parts) != 4 {
		logger.Logger.WithField("parts", parts).Error("Unsupported URI")
		return "", "", uint16(0), errors.New("could not decode LSP URI")
	}

	portValue, err := strconv.Atoi(parts[3])
	if err != nil {
		logger.Logger.WithField("port", parts[3]).WithError(err).Error("Failed to decode port number")

		return "", "", uint16(0), err
	}

	return parts[1], parts[2], uint16(portValue), nil
}

func setDefaultRequestHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "AlbyHub/"+version.Tag)
}

func (svc *albyOAuthService) deleteAlbyAccountApps() {
	// delete any existing getalby.com connections so when re-linking the user only has one
	err := svc.db.Where("name = ?", ALBY_ACCOUNT_APP_NAME).Delete(&db.App{}).Error
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to delete Alby Account apps")
	}
}

// whitelist of events that can be sent to the alby API
// (e.g. to enable encrypted static channel backups and sending email notifications)
func getEventWhitelist() []string {
	return []string{
		"nwc_backup_channels",
		"nwc_payment_received",
		"nwc_payment_sent",
		"nwc_payment_failed",
		"nwc_app_created",
		"nwc_app_updated",
		"nwc_app_deleted",
		"nwc_unlocked",
		"nwc_node_sync_failed",
		"nwc_outgoing_liquidity_required",
		"nwc_incoming_liquidity_required",
		"nwc_budget_warning",
		"nwc_channel_ready",
		"nwc_channel_closed",
		"nwc_permission_denied",
		"nwc_started",
		"nwc_stopped",
		"nwc_node_started",
		"nwc_node_start_failed",
		"nwc_node_stop_failed",
		"nwc_node_stopped",
		"nwc_alby_account_connected",
		"nwc_swap_succeeded",
		"nwc_rebalance_succeeded",

		// client-side events
		"payment_failed_details",
	}
}
