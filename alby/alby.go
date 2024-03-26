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
	"github.com/getAlby/nostr-wallet-connect/models/config"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type AlbyOAuthService struct {
	appConfig *config.AppConfig
	kvStore   config.ConfigKVStore
	oauthConf *oauth2.Config
	logger    *logrus.Logger
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
}

type AlbyBalance struct {
	Balance  int64  `json:"balance"`
	Unit     string `json:"unit"`
	Currency string `json:"currency"`
}

const (
	ACCESS_TOKEN_KEY        = "AlbyOAuthAccessToken"
	ACCESS_TOKEN_EXPIRY_KEY = "AlbyOAuthAccessTokenExpiry"
	REFRESH_TOKEN_KEY       = "AlbyOAuthRefreshToken"
)

func NewAlbyOauthService(logger *logrus.Logger, kvStore config.ConfigKVStore, appConfig *config.AppConfig) *AlbyOAuthService {
	conf := &oauth2.Config{
		ClientID:     appConfig.AlbyClientId,
		ClientSecret: appConfig.AlbyClientSecret,
		Scopes:       []string{"account:read", "balance:read", "payments:send"},
		Endpoint: oauth2.Endpoint{
			TokenURL:  appConfig.AlbyAPIURL + "/oauth/token",
			AuthURL:   appConfig.AlbyOAuthAuthUrl,
			AuthStyle: 2, // use HTTP Basic Authorization https://pkg.go.dev/golang.org/x/oauth2#AuthStyle
		},
		RedirectURL: appConfig.BaseUrl + "/api/alby/callback",
	}

	albyOAuthSvc := &AlbyOAuthService{
		appConfig: appConfig,
		oauthConf: conf,
		kvStore:   kvStore,
		logger:    logger,
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

	return nil
}

func (svc *AlbyOAuthService) saveToken(token *oauth2.Token) {
	svc.kvStore.SetUpdate(ACCESS_TOKEN_EXPIRY_KEY, strconv.FormatInt(token.Expiry.Unix(), 10), "")
	svc.kvStore.SetUpdate(ACCESS_TOKEN_KEY, token.AccessToken, "")
	svc.kvStore.SetUpdate(REFRESH_TOKEN_KEY, token.RefreshToken, "")
}

var tokenMutex sync.Mutex

func (svc *AlbyOAuthService) fetchUserToken(ctx context.Context) (*oauth2.Token, error) {
	tokenMutex.Lock()
	defer tokenMutex.Unlock()
	accessToken, err := svc.kvStore.Get("AccessToken", "")
	if err != nil {
		return nil, err
	}
	expiry, err := svc.kvStore.Get(ACCESS_TOKEN_EXPIRY_KEY, "")
	if err != nil {
		return nil, err
	}
	expiry64, err := strconv.ParseInt(expiry, 10, 64)
	if err != nil {
		return nil, err
	}
	refreshToken, err := svc.kvStore.Get(REFRESH_TOKEN_KEY, "")
	if err != nil {
		return nil, err
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

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/user/me", svc.appConfig.AlbyAPIURL), nil)
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

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/balance", svc.appConfig.AlbyAPIURL), nil)
	if err != nil {
		svc.logger.WithError(err).Error("Error creating request /balance")
		return nil, err
	}

	req.Header.Set("User-Agent", "NWC-next")

	res, err := client.Do(req)
	if err != nil {
		svc.logger.WithError(err).Error("Failed to fetch /balance")
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

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/payments/bolt11", svc.appConfig.AlbyAPIURL), body)
	if err != nil {
		svc.logger.WithError(err).Error("Error creating request /payments/bolt11")
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
	return svc.oauthConf.AuthCodeURL("unused")
}

func (svc *AlbyOAuthService) Log(ctx context.Context, event *events.Event) error {
	token, err := svc.fetchUserToken(ctx)
	if err != nil {
		svc.logger.WithError(err).Error("Failed to fetch user token")
	}

	client := svc.oauthConf.Client(ctx, token)

	body := bytes.NewBuffer([]byte{})
	err = json.NewEncoder(body).Encode(event)

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
			"event": event,
		}).WithError(err).Error("Failed to send request to /events")
		return err
	}

	if resp.StatusCode >= 300 {
		svc.logger.WithFields(logrus.Fields{
			"event":  event,
			"status": resp.StatusCode,
		}).Error("Request to /events returned non-success status")
		return errors.New("request to /events returned non-success status")
	}

	return nil
}
