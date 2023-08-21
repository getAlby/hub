package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type AlbyOAuthService struct {
	cfg       *Config
	oauthConf *oauth2.Config
	db        *gorm.DB
	Logger    *logrus.Logger
}

func NewAlbyOauthService(svc *Service, e *echo.Echo) (result *AlbyOAuthService, err error) {
	conf := &oauth2.Config{
		ClientID:     svc.cfg.AlbyClientId,
		ClientSecret: svc.cfg.AlbyClientSecret,
		//Todo: do we really need all these permissions?
		Scopes: []string{"account:read", "payments:send", "invoices:read", "transactions:read", "invoices:create", "balance:read"},
		Endpoint: oauth2.Endpoint{
			TokenURL:  svc.cfg.OAuthTokenUrl,
			AuthURL:   svc.cfg.OAuthAuthUrl,
			AuthStyle: 2, // use HTTP Basic Authorization https://pkg.go.dev/golang.org/x/oauth2#AuthStyle
		},
		RedirectURL: svc.cfg.OAuthRedirectUrl,
	}

	albySvc := &AlbyOAuthService{
		cfg:       svc.cfg,
		oauthConf: conf,
		db:        svc.db,
		Logger:    svc.Logger,
	}

	e.GET("/alby/auth", albySvc.AuthHandler)
	e.GET("/alby/callback", albySvc.CallbackHandler)

	return albySvc, err
}

func (svc *AlbyOAuthService) FetchUserToken(ctx context.Context, app App) (token *oauth2.Token, err error) {
	user := app.User
	tok, err := svc.oauthConf.TokenSource(ctx, &oauth2.Token{
		AccessToken:  user.AccessToken,
		RefreshToken: user.RefreshToken,
		Expiry:       user.Expiry,
	}).Token()
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey": app.NostrPubkey,
			"appId":        app.ID,
			"userId":       app.User.ID,
		}).Errorf("Token error: %v", err)
		return nil, err
	}
	// we always update the user's token for future use
	// the oauth library handles the token refreshing
	user.AccessToken = tok.AccessToken
	user.RefreshToken = tok.RefreshToken
	user.Expiry = tok.Expiry // TODO; probably needs some calculation
	err = svc.db.Save(&user).Error
	if err != nil {
		svc.Logger.WithError(err).Error("Error saving user")
		return nil, err
	}
	return tok, nil
}

func (svc *AlbyOAuthService) MakeInvoice(ctx context.Context, senderPubkey string, amount int64, description string, descriptionHash string, expiry int64) (invoice string, paymentHash string, err error) {
	// TODO: move to a shared function
	app := App{}
	err = svc.db.Preload("User").First(&app, &App{
		NostrPubkey: senderPubkey,
	}).Error
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey":    senderPubkey,
			"amount":          amount,
			"description":     description,
			"descriptionHash": descriptionHash,
			"expiry":          expiry,
		}).Errorf("App not found: %v", err)
		return "", "", err
	}

	// amount provided in msat, but Alby API currently only supports sats. Will get truncated to a whole sat value
	var amountSat int64 = amount / 1000
	// make sure amount is not converted to 0
	if (amount > 0 && amountSat == 0) {
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey":    senderPubkey,
			"amount":          amount,
			"description":     description,
			"descriptionHash": descriptionHash,
			"expiry":          expiry,
		}).Errorf("Value must be 1 sat or greater");
		return "", "", errors.New("Value must be 1 sat or greater")
	}

	svc.Logger.WithFields(logrus.Fields{
		"senderPubkey":    senderPubkey,
		"amount":          amount,
		"description":     description,
		"descriptionHash": descriptionHash,
		"expiry":          expiry,
		"appId":           app.ID,
		"userId":          app.User.ID,
	}).Info("Processing make invoice request")
	tok, err := svc.FetchUserToken(ctx, app)
	if err != nil {
		return "", "", err
	}
	client := svc.oauthConf.Client(ctx, tok)

	body := bytes.NewBuffer([]byte{})
	payload := &MakeInvoiceRequest{
		Amount: amountSat,
		Description: description,
		DescriptionHash: descriptionHash,
		// TODO: support expiry
	}
	err = json.NewEncoder(body).Encode(payload)

	// TODO: move to a shared function
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/invoices", svc.cfg.AlbyAPIURL), body)
	if err != nil {
		svc.Logger.WithError(err).Error("Error creating request /invoices")
		return "", "", err
	}

	req.Header.Set("User-Agent", "NWC")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey":    senderPubkey,
			"amount":          amount,
			"description":     description,
			"descriptionHash": descriptionHash,
			"expiry":          expiry,
			"appId":           app.ID,
			"userId":          app.User.ID,
		}).Errorf("Failed to make invoice: %v", err)
		return "", "", err
	}

	if resp.StatusCode < 300 {
		responsePayload := &MakeInvoiceResponse{}
		err = json.NewDecoder(resp.Body).Decode(responsePayload)
		if err != nil {
			return "", "", err
		}
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey":    senderPubkey,
			"amount":          amount,
			"description":     description,
			"descriptionHash": descriptionHash,
			"expiry":          expiry,
			"appId":           app.ID,
			"userId":          app.User.ID,
			"paymentRequest":  responsePayload.PaymentRequest,
			"paymentHash":  responsePayload.PaymentHash,
		}).Info("Make invoice successful")
		return responsePayload.PaymentRequest, responsePayload.PaymentHash, nil
	} else {
		errorPayload := &ErrorResponse{}
		err = json.NewDecoder(resp.Body).Decode(errorPayload)
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey":    senderPubkey,
			"amount":          amount,
			"description":     description,
			"descriptionHash": descriptionHash,
			"expiry":          expiry,
			"appId":           app.ID,
			"userId":          app.User.ID,
			"APIHttpStatus": resp.StatusCode,
		}).Errorf("Make invoice failed %s", string(errorPayload.Message))
		return "", "", errors.New(errorPayload.Message)
	}
}

func (svc *AlbyOAuthService) GetBalance(ctx context.Context, senderPubkey string) (balance int64, err error) {
	app := App{}
	err = svc.db.Preload("User").First(&app, &App{
		NostrPubkey: senderPubkey,
	}).Error
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey": senderPubkey,
		}).Errorf("App not found: %v", err)
		return 0, err
	}
	tok, err := svc.FetchUserToken(ctx, app)
	if err != nil {
		return 0, err
	}
	client := svc.oauthConf.Client(ctx, tok)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/balance", svc.cfg.AlbyAPIURL), nil)
	if err != nil {
		svc.Logger.WithError(err).Error("Error creating request /balance")
		return 0, err
	}

	req.Header.Set("User-Agent", "NWC")

	resp, err := client.Do(req)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey": senderPubkey,
			"appId":        app.ID,
			"userId":       app.User.ID,
		}).Errorf("Failed to fetch balance: %v", err)
		return 0, err
	}

	if resp.StatusCode < 300 {
		responsePayload := &BalanceResponse{}
		err = json.NewDecoder(resp.Body).Decode(responsePayload)
		if err != nil {
			return 0, err
		}
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey": senderPubkey,
			"appId":        app.ID,
			"userId":       app.User.ID,
		}).Info("Balance fetch successful")
		return int64(responsePayload.Balance), nil
	} else {
		errorPayload := &ErrorResponse{}
		err = json.NewDecoder(resp.Body).Decode(errorPayload)
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey":  senderPubkey,
			"appId":         app.ID,
			"userId":        app.User.ID,
			"APIHttpStatus": resp.StatusCode,
		}).Errorf("Balance fetch failed %s", string(errorPayload.Message))
		return 0, errors.New(errorPayload.Message)
	}
}

func (svc *AlbyOAuthService) SendPaymentSync(ctx context.Context, senderPubkey, payReq string) (preimage string, err error) {
	app := App{}
	err = svc.db.Preload("User").First(&app, &App{
		NostrPubkey: senderPubkey,
	}).Error
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey": senderPubkey,
			"bolt11":       payReq,
		}).Errorf("App not found: %v", err)
		return "", err
	}
	svc.Logger.WithFields(logrus.Fields{
		"senderPubkey": senderPubkey,
		"bolt11":       payReq,
		"appId":        app.ID,
		"userId":       app.User.ID,
	}).Info("Processing payment request")
	tok, err := svc.FetchUserToken(ctx, app)
	if err != nil {
		return "", err
	}
	client := svc.oauthConf.Client(ctx, tok)

	body := bytes.NewBuffer([]byte{})
	payload := &PayRequest{
		Invoice: payReq,
	}
	err = json.NewEncoder(body).Encode(payload)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/payments/bolt11", svc.cfg.AlbyAPIURL), body)
	if err != nil {
		svc.Logger.WithError(err).Error("Error creating request /payments/bolt11")
		return "", err
	}

	req.Header.Set("User-Agent", "NWC")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey": senderPubkey,
			"bolt11":       payReq,
			"appId":        app.ID,
			"userId":       app.User.ID,
		}).Errorf("Failed to pay invoice: %v", err)
		return "", err
	}

	if resp.StatusCode < 300 {
		responsePayload := &PayResponse{}
		err = json.NewDecoder(resp.Body).Decode(responsePayload)
		if err != nil {
			return "", err
		}
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey": senderPubkey,
			"bolt11":       payReq,
			"appId":        app.ID,
			"userId":       app.User.ID,
		}).Info("Payment successful")
		return responsePayload.Preimage, nil
	} else {
		errorPayload := &ErrorResponse{}
		err = json.NewDecoder(resp.Body).Decode(errorPayload)
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey":  senderPubkey,
			"bolt11":        payReq,
			"appId":         app.ID,
			"userId":        app.User.ID,
			"APIHttpStatus": resp.StatusCode,
		}).Errorf("Payment failed %s", string(errorPayload.Message))
		return "", errors.New(errorPayload.Message)
	}
}

func (svc *AlbyOAuthService) AuthHandler(c echo.Context) error {
	// clear current session
	sess, _ := session.Get(CookieName, c)
	if sess.Values["user_id"] != nil {
		delete(sess.Values, "user_id")
		sess.Options.MaxAge = 0
		sess.Options.SameSite = http.SameSiteLaxMode
		if svc.cfg.CookieDomain != "" {
			sess.Options.Domain = svc.cfg.CookieDomain
		}
		sess.Save(c.Request(), c.Response())
	}

	url := svc.oauthConf.AuthCodeURL("")
	return c.Redirect(302, url)
}

func (svc *AlbyOAuthService) CallbackHandler(c echo.Context) error {
	code := c.QueryParam("code")
	tok, err := svc.oauthConf.Exchange(c.Request().Context(), code)
	if err != nil {
		svc.Logger.WithError(err).Error("Failed to exchange token")
		return err
	}
	client := svc.oauthConf.Client(c.Request().Context(), tok)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/user/me", svc.cfg.AlbyAPIURL), nil)
	if err != nil {
		svc.Logger.WithError(err).Error("Error creating request /me")
		return err
	}

	req.Header.Set("User-Agent", "NWC")

	res, err := client.Do(req)
	if err != nil {
		svc.Logger.WithError(err).Error("Failed to fetch /me")
		return err
	}
	me := AlbyMe{}
	err = json.NewDecoder(res.Body).Decode(&me)
	if err != nil {
		svc.Logger.WithError(err).Error("Failed to decode API response")
		return err
	}

	user := User{}
	svc.db.FirstOrInit(&user, User{AlbyIdentifier: me.Identifier})
	user.AccessToken = tok.AccessToken
	user.RefreshToken = tok.RefreshToken
	user.Expiry = tok.Expiry // TODO; probably needs some calculation
	user.Email = me.Email
	user.LightningAddress = me.LightningAddress
	svc.db.Save(&user)

	sess, _ := session.Get(CookieName, c)
	sess.Options.MaxAge = 0
	sess.Options.SameSite = http.SameSiteLaxMode
	if svc.cfg.CookieDomain != "" {
		sess.Options.Domain = svc.cfg.CookieDomain
	}
	sess.Values["user_id"] = user.ID
	sess.Save(c.Request(), c.Response())
	return c.Redirect(302, "/")
}
