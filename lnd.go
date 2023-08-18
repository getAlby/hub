package main

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/getAlby/lndhub.go/lnd"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/lightningnetwork/lnd/lnrpc"
)

type LNClient interface {
	SendPaymentSync(ctx context.Context, senderPubkey string, payReq string) (preimage string, err error)
	GetBalance(ctx context.Context, senderPubkey string) (balance int64, err error)
	MakeInvoice(ctx context.Context, senderPubkey string, amount int64, description string, descriptionHash string, expiry int64) (invoice string, paymentHash string, err error)
}

// wrap it again :sweat_smile:
// todo: drop dependency on lndhub package
type LNDService struct {
	client *lnd.LNDWrapper
	db     *gorm.DB
	Logger *logrus.Logger
}

func (svc *LNDService) AuthHandler(c echo.Context) error {
	user := &User{}
	err := svc.db.FirstOrInit(user, User{AlbyIdentifier: "lnd"}).Error
	if err != nil {
		return err
	}

	sess, _ := session.Get(CookieName, c)
	sess.Values["user_id"] = user.ID
	sess.Save(c.Request(), c.Response())
	return c.Redirect(302, "/")
}

func (svc *LNDService) GetBalance(ctx context.Context, senderPubkey string) (balance int64, err error) {
	resp, err := svc.client.WalletBalance(ctx, &lnrpc.WalletBalanceRequest{})
	if err != nil {
		return 0, err
	}
	return resp.TotalBalance, nil
}

func (svc *LNDService) MakeInvoice(ctx context.Context, senderPubkey string, amount int64, description string, descriptionHash string, expiry int64) (invoice string, paymentHash string, err error) {
	return "", "", fmt.Errorf("not implemented")
}

func (svc *LNDService) SendPaymentSync(ctx context.Context, senderPubkey, payReq string) (preimage string, err error) {
	resp, err := svc.client.SendPaymentSync(ctx, &lnrpc.SendRequest{PaymentRequest: payReq})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(resp.PaymentPreimage), nil
}

func NewLNDService(ctx context.Context, svc *Service, e *echo.Echo) (result *LNDService, err error) {
	lndClient, err := lnd.NewLNDclient(lnd.LNDoptions{
		Address:      svc.cfg.LNDAddress,
		CertFile:     svc.cfg.LNDCertFile,
		MacaroonFile: svc.cfg.LNDMacaroonFile,
	})
	if err != nil {
		return nil, err
	}
	info, err := lndClient.GetInfo(ctx, &lnrpc.GetInfoRequest{})
	if err != nil {
		return nil, err
	}
	//add default user to db
	user := &User{}
	err = svc.db.FirstOrInit(user, User{AlbyIdentifier: "lnd"}).Error
	if err != nil {
		return nil, err
	}
	err = svc.db.Save(user).Error
	if err != nil {
		return nil, err
	}

	lndService := &LNDService{client: lndClient, Logger: svc.Logger, db: svc.db}

	e.GET("/lnd/auth", lndService.AuthHandler)
	svc.Logger.Infof("Connected to LND - alias %s", info.Alias)

	return lndService, nil
}
