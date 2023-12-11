package main

import (
	"context"
	"encoding/hex"
	"errors"
	"time"

	"github.com/getAlby/nostr-wallet-connect/lnd"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/lightningnetwork/lnd/lnrpc"
)

type LNClient interface {
	SendPaymentSync(ctx context.Context, senderPubkey string, payReq string) (preimage string, err error)
	GetBalance(ctx context.Context, senderPubkey string) (balance int64, err error)
	MakeInvoice(ctx context.Context, senderPubkey string, amount int64, description string, descriptionHash string, expiry int64) (invoice string, paymentHash string, err error)
	LookupInvoice(ctx context.Context, senderPubkey string, paymentHash string) (invoice string, paid bool, err error)
	ListTransactions(ctx context.Context, senderPubkey string, from, until, limit, offset uint64, unpaid bool, invoiceType string) (invoices []Invoice, err error)
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
	resp, err := svc.client.ChannelBalance(ctx, &lnrpc.ChannelBalanceRequest{})
	if err != nil {
		return 0, err
	}
	return int64(resp.LocalBalance.Sat), nil
}

func (svc *LNDService) ListTransactions(ctx context.Context, senderPubkey string, from, until, limit, offset uint64, unpaid bool, invoiceType string) (invoices []Invoice, err error) {
	maxInvoices := uint64(limit)
	if err != nil {
		return nil, err
	}
	indexOffset := uint64(offset)
	if err != nil {
		return nil, err
	}
	resp, err := svc.client.ListInvoices(ctx, &lnrpc.ListInvoiceRequest{NumMaxInvoices: maxInvoices, IndexOffset: indexOffset})
	if err != nil {
		return nil, err
	}

	for _, inv := range resp.Invoices {
		invoice := Invoice{
			Invoice:         inv.PaymentRequest,
			Description:     inv.Memo,
			DescriptionHash: hex.EncodeToString(inv.DescriptionHash),
			Preimage:        hex.EncodeToString(inv.RPreimage),
			PaymentHash:     hex.EncodeToString(inv.RHash),
			Amount:          inv.ValueMsat,
			FeesPaid:        inv.AmtPaidMsat,
			SettledAt:       time.Unix(inv.SettleDate, 0),
		}
		invoices = append(invoices, invoice)
	}
	return invoices, nil
}

func (svc *LNDService) MakeInvoice(ctx context.Context, senderPubkey string, amount int64, description string, descriptionHash string, expiry int64) (invoice string, paymentHash string, err error) {
	var descriptionHashBytes []byte
	
	if descriptionHash != "" {
		descriptionHashBytes, err = hex.DecodeString(descriptionHash)

		if err != nil || len(descriptionHashBytes) != 32 {
			svc.Logger.WithFields(logrus.Fields{
				"senderPubkey":    senderPubkey,
				"amount":          amount,
				"description":     description,
				"descriptionHash": descriptionHash,
				"expiry":          expiry,
			}).Errorf("Invalid description hash")
			return "", "", errors.New("Description hash must be 32 bytes hex")
		}
	}
	
	resp, err := svc.client.AddInvoice(ctx, &lnrpc.Invoice{ValueMsat: amount, Memo: description, DescriptionHash: descriptionHashBytes, Expiry: expiry})
	if err != nil {
		return "", "", err
	}

	return resp.GetPaymentRequest(), hex.EncodeToString(resp.GetRHash()), nil
}

func (svc *LNDService) LookupInvoice(ctx context.Context, senderPubkey string, paymentHash string) (invoice string, paid bool, err error) {
	paymentHashBytes, err := hex.DecodeString(paymentHash)

	if err != nil || len(paymentHashBytes) != 32 {
		svc.Logger.WithFields(logrus.Fields{
			"paymentHash": paymentHash,
		}).Errorf("Invalid payment hash")
		return "", false, errors.New("Payment hash must be 32 bytes hex")
	}

	lndInvoice, err := svc.client.LookupInvoice(ctx, &lnrpc.PaymentHash{RHash: paymentHashBytes})
	if err != nil {
		return "", false, err
	}
	
	return lndInvoice.PaymentRequest, lndInvoice.State == *lnrpc.Invoice_SETTLED.Enum(), nil
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
	}, ctx)
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
