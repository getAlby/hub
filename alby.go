package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

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

func (svc *AlbyOAuthService) MakeInvoice(ctx context.Context, senderPubkey string, amount int64, description string, descriptionHash string, expiry int64) (transaction *Nip47Transaction, err error) {
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
		return nil, err
	}

	// amount provided in msat, but Alby API currently only supports sats. Will get truncated to a whole sat value
	var amountSat int64 = amount / 1000
	// make sure amount is not converted to 0
	if amount > 0 && amountSat == 0 {
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey":    senderPubkey,
			"amount":          amount,
			"description":     description,
			"descriptionHash": descriptionHash,
			"expiry":          expiry,
		}).Errorf("amount must be 1000 msat or greater")
		return nil, errors.New("amount must be 1000 msat or greater")
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
		return nil, err
	}
	client := svc.oauthConf.Client(ctx, tok)

	body := bytes.NewBuffer([]byte{})
	payload := &MakeInvoiceRequest{
		Amount:          amountSat,
		Description:     description,
		DescriptionHash: descriptionHash,
		// TODO: support expiry
	}
	err = json.NewEncoder(body).Encode(payload)

	// TODO: move to a shared function
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/invoices", svc.cfg.AlbyAPIURL), body)
	if err != nil {
		svc.Logger.WithError(err).Error("Error creating request /invoices")
		return nil, err
	}

	// TODO: move to creation of HTTP client
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
		return nil, err
	}

	if resp.StatusCode < 300 {
		responsePayload := &AlbyInvoice{}
		err = json.NewDecoder(resp.Body).Decode(responsePayload)
		if err != nil {
			return nil, err
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
			"paymentHash":     responsePayload.PaymentHash,
		}).Info("Make invoice successful")

		transaction := albyInvoiceToTransaction(responsePayload)
		return transaction, nil
	}

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
		"APIHttpStatus":   resp.StatusCode,
	}).Errorf("Make invoice failed %s", string(errorPayload.Message))
	return nil, errors.New(errorPayload.Message)
}

func (svc *AlbyOAuthService) LookupInvoice(ctx context.Context, senderPubkey string, paymentHash string) (transaction *Nip47Transaction, err error) {
	// TODO: move to a shared function
	app := App{}
	err = svc.db.Preload("User").First(&app, &App{
		NostrPubkey: senderPubkey,
	}).Error
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey": senderPubkey,
			"paymentHash":  paymentHash,
		}).Errorf("App not found: %v", err)
		return nil, err
	}

	svc.Logger.WithFields(logrus.Fields{
		"senderPubkey": senderPubkey,
		"paymentHash":  paymentHash,
		"appId":        app.ID,
		"userId":       app.User.ID,
	}).Info("Processing lookup invoice request")
	tok, err := svc.FetchUserToken(ctx, app)
	if err != nil {
		return nil, err
	}
	client := svc.oauthConf.Client(ctx, tok)

	body := bytes.NewBuffer([]byte{})

	// TODO: move to a shared function
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/invoices/%s", svc.cfg.AlbyAPIURL, paymentHash), body)
	if err != nil {
		svc.Logger.WithError(err).Errorf("Error creating request /invoices/%s", paymentHash)
		return nil, err
	}

	req.Header.Set("User-Agent", "NWC")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey": senderPubkey,
			"paymentHash":  paymentHash,
			"appId":        app.ID,
			"userId":       app.User.ID,
		}).Errorf("Failed to lookup invoice: %v", err)
		return nil, err
	}

	if resp.StatusCode < 300 {
		responsePayload := &AlbyInvoice{}
		err = json.NewDecoder(resp.Body).Decode(responsePayload)
		if err != nil {
			return nil, err
		}
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey":   senderPubkey,
			"paymentHash":    paymentHash,
			"appId":          app.ID,
			"userId":         app.User.ID,
			"paymentRequest": responsePayload.PaymentRequest,
			"settled":        responsePayload.Settled,
		}).Info("Lookup invoice successful")

		transaction = albyInvoiceToTransaction(responsePayload)
		return transaction, nil
	}

	errorPayload := &ErrorResponse{}
	err = json.NewDecoder(resp.Body).Decode(errorPayload)
	svc.Logger.WithFields(logrus.Fields{
		"senderPubkey":  senderPubkey,
		"paymentHash":   paymentHash,
		"appId":         app.ID,
		"userId":        app.User.ID,
		"APIHttpStatus": resp.StatusCode,
	}).Errorf("Lookup invoice failed %s", string(errorPayload.Message))
	return nil, errors.New(errorPayload.Message)
}

func (svc *AlbyOAuthService) GetInfo(ctx context.Context, senderPubkey string) (info *NodeInfo, err error) {
	app := App{}
	err = svc.db.Preload("User").First(&app, &App{
		NostrPubkey: senderPubkey,
	}).Error
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey": senderPubkey,
		}).Errorf("App not found: %v", err)
		return nil, err
	}

	svc.Logger.WithFields(logrus.Fields{
		"senderPubkey": senderPubkey,
		"appId":        app.ID,
		"userId":       app.User.ID,
	}).Info("Info fetch successful")
	return &NodeInfo{
		Alias:       "getalby.com",
		Color:       "",
		Pubkey:      "",
		Network:     "mainnet",
		BlockHeight: 0,
		BlockHash:   "",
	}, err
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
	}

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

func (svc *AlbyOAuthService) ListTransactions(ctx context.Context, senderPubkey string, from, until, limit, offset uint64, unpaid bool, invoiceType string) (transactions []Nip47Transaction, err error) {
	app := App{}
	err = svc.db.Preload("User").First(&app, &App{
		NostrPubkey: senderPubkey,
	}).Error
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey": senderPubkey,
		}).Errorf("App not found: %v", err)
		return nil, err
	}
	tok, err := svc.FetchUserToken(ctx, app)
	if err != nil {
		return nil, err
	}
	client := svc.oauthConf.Client(ctx, tok)

	urlParams := url.Values{}
	//urlParams.Add("page", "1")

	// TODO: clarify gt/lt vs from to in NWC spec
	if from != 0 {
		urlParams.Add("q[created_at_gt]", strconv.FormatUint(from, 10))
	}
	if until != 0 {
		urlParams.Add("q[created_at_lt]", strconv.FormatUint(until, 10))
	}
	if limit != 0 {
		urlParams.Add("items", strconv.FormatUint(limit, 10))
	}
	// TODO: Add Offset and Unpaid

	endpoint := "/invoices"

	switch invoiceType {
	case "incoming":
		endpoint += "/incoming"
	case "outgoing":
		endpoint += "/outgoing"
	}

	requestUrl := fmt.Sprintf("%s%s?%s", svc.cfg.AlbyAPIURL, endpoint, urlParams.Encode())

	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		svc.Logger.WithError(err).Error("Error creating request /invoices")
		return nil, err
	}

	req.Header.Set("User-Agent", "NWC")

	resp, err := client.Do(req)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey": senderPubkey,
			"appId":        app.ID,
			"userId":       app.User.ID,
			"requestUrl":   requestUrl,
		}).Errorf("Failed to fetch invoices: %v", err)
		return nil, err
	}

	var invoices []AlbyInvoice

	if resp.StatusCode < 300 {
		err = json.NewDecoder(resp.Body).Decode(&invoices)
		if err != nil {
			svc.Logger.WithFields(logrus.Fields{
				"senderPubkey": senderPubkey,
				"appId":        app.ID,
				"userId":       app.User.ID,
				"requestUrl":   requestUrl,
			}).Errorf("Failed to decode invoices: %v", err)
			return nil, err
		}

		transactions = []Nip47Transaction{}
		for _, invoice := range invoices {
			transaction := albyInvoiceToTransaction(&invoice)

			transactions = append(transactions, *transaction)
		}

		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey": senderPubkey,
			"appId":        app.ID,
			"userId":       app.User.ID,
			"requestUrl":   requestUrl,
		}).Info("List transactions successful")
		return transactions, nil
	}

	errorPayload := &ErrorResponse{}
	err = json.NewDecoder(resp.Body).Decode(errorPayload)
	svc.Logger.WithFields(logrus.Fields{
		"senderPubkey":  senderPubkey,
		"appId":         app.ID,
		"userId":        app.User.ID,
		"APIHttpStatus": resp.StatusCode,
		"requestUrl":    requestUrl,
	}).Errorf("List transactions failed %s", string(errorPayload.Message))
	return nil, errors.New(errorPayload.Message)
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
			"paymentHash":  responsePayload.PaymentHash,
		}).Info("Payment successful")
		return responsePayload.Preimage, nil
	}

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

func (svc *AlbyOAuthService) SendKeysend(ctx context.Context, senderPubkey string, amount int64, destination, preimage string, custom_records []TLVRecord) (preImage string, err error) {
	app := App{}
	err = svc.db.Preload("User").First(&app, &App{
		NostrPubkey: senderPubkey,
	}).Error
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey": senderPubkey,
			"payeePubkey":  destination,
		}).Errorf("App not found: %v", err)
		return "", err
	}
	svc.Logger.WithFields(logrus.Fields{
		"senderPubkey": senderPubkey,
		"payeePubkey":  destination,
		"appId":        app.ID,
		"userId":       app.User.ID,
	}).Info("Processing keysend request")
	tok, err := svc.FetchUserToken(ctx, app)
	if err != nil {
		return "", err
	}
	client := svc.oauthConf.Client(ctx, tok)

	customRecordsMap := make(map[string]string)
	for _, record := range custom_records {
		customRecordsMap[strconv.FormatUint(record.Type, 10)] = record.Value
	}

	body := bytes.NewBuffer([]byte{})
	payload := &KeysendRequest{
		Amount:        amount,
		Destination:   destination,
		CustomRecords: customRecordsMap,
	}
	err = json.NewEncoder(body).Encode(payload)

	// here we don't use the preimage from params
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/payments/keysend", svc.cfg.AlbyAPIURL), body)
	if err != nil {
		svc.Logger.WithError(err).Error("Error creating request /payments/keysend")
		return "", err
	}

	req.Header.Set("User-Agent", "NWC")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey": senderPubkey,
			"payeePubkey":  destination,
			"appId":        app.ID,
			"userId":       app.User.ID,
		}).Errorf("Failed to pay keysend: %v", err)
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
			"payeePubkey":  destination,
			"appId":        app.ID,
			"userId":       app.User.ID,
			"preimage":     responsePayload.Preimage,
			"paymentHash":  responsePayload.PaymentHash,
		}).Info("Keysend payment successful")
		return responsePayload.Preimage, nil
	}

	errorPayload := &ErrorResponse{}
	err = json.NewDecoder(resp.Body).Decode(errorPayload)
	svc.Logger.WithFields(logrus.Fields{
		"senderPubkey":  senderPubkey,
		"payeePubkey":   destination,
		"appId":         app.ID,
		"userId":        app.User.ID,
		"APIHttpStatus": resp.StatusCode,
	}).Errorf("Payment failed %s", string(errorPayload.Message))
	return "", errors.New(errorPayload.Message)
}

func (svc *AlbyOAuthService) AuthHandler(c echo.Context) error {
	appName := c.QueryParam("c") // c - for client
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

	url := svc.oauthConf.AuthCodeURL(appName) // pass on the appName as state
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

func albyInvoiceToTransaction(invoice *AlbyInvoice) *Nip47Transaction {
	description := invoice.Comment
	if description == "" {
		description = invoice.Memo
	}
	var preimage string
	if invoice.SettledAt != nil {
		preimage = invoice.Preimage
	}

	return &Nip47Transaction{
		Type:            invoice.Type,
		Invoice:         invoice.PaymentRequest,
		Description:     description,
		DescriptionHash: invoice.DescriptionHash,
		Preimage:        preimage,
		PaymentHash:     invoice.PaymentHash,
		Amount:          invoice.Amount * 1000,
		FeesPaid:        0, // TODO: support fees
		CreatedAt:       invoice.CreatedAt,
		ExpiresAt:       invoice.ExpiresAt,
		SettledAt:       invoice.SettledAt,
		Metadata:        invoice.Metadata,
	}
}
