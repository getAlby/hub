package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/getAlby/nostr-wallet-connect/lnd"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/lightningnetwork/lnd/lnrpc"
)

type LNClient interface {
	SendPaymentSync(ctx context.Context, senderPubkey string, payReq string) (preimage string, err error)
	SendKeysend(ctx context.Context, senderPubkey string, amount int64, destination, preimage string, custom_records []TLVRecord) (preImage string, err error)
	GetBalance(ctx context.Context, senderPubkey string) (balance int64, err error)
	GetInfo(ctx context.Context, senderPubkey string) (info *NodeInfo, err error)
	MakeInvoice(ctx context.Context, senderPubkey string, amount int64, description string, descriptionHash string, expiry int64) (invoice string, paymentHash string, err error)
	LookupInvoice(ctx context.Context, senderPubkey string, paymentHash string) (invoice string, paid bool, err error)
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

func (svc *LNDService) GetInfo(ctx context.Context, senderPubkey string) (info *NodeInfo, err error) {
	resp, err := svc.client.GetInfo(ctx, &lnrpc.GetInfoRequest{})
	if err != nil {
		return nil, err
	}
	return &NodeInfo{
		Alias:       resp.Alias,
		Color:       resp.Color,
		Pubkey:      resp.IdentityPubkey,
		Network:     resp.Chains[0].Network,
		BlockHeight: resp.BlockHeight,
		BlockHash:   resp.BlockHash,
	}, nil
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

func (svc *LNDService) SendKeysend(ctx context.Context, senderPubkey string, amount int64, destination, preimage string, custom_records []TLVRecord) (respPreimage string, err error) {
	destBytes, err := hex.DecodeString(destination)
	if err != nil {
		return "", err
	}
	var preImageBytes []byte

	if preimage == "" {
		preImageBytes, err = makePreimageHex()
		preimage = hex.EncodeToString(preImageBytes)
	} else {
		preImageBytes, err = hex.DecodeString(preimage)
	}
	if err != nil || len(preImageBytes) != 32 {
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey":  senderPubkey,
			"amount":        amount,
			"destination":   destination,
			"preimage":      preimage,
			"customRecords": custom_records,
			"error":         err,
		}).Errorf("Invalid preimage")
		return "", err
	}

	paymentHash := sha256.New()
	paymentHash.Write(preImageBytes)
	paymentHashBytes := paymentHash.Sum(nil)
	paymentHashHex := hex.EncodeToString(paymentHashBytes)

	destCustomRecords := map[uint64][]byte{}
	for _, record := range custom_records {
		destCustomRecords[record.Type] = []byte(record.Value)
	}
	const KEYSEND_CUSTOM_RECORD = 5482373484
	destCustomRecords[KEYSEND_CUSTOM_RECORD] = preImageBytes
	sendPaymentRequest := &lnrpc.SendRequest{
		Dest:              destBytes,
		Amt:               amount,
		PaymentHash:       paymentHashBytes,
		DestFeatures:      []lnrpc.FeatureBit{lnrpc.FeatureBit_TLV_ONION_REQ},
		DestCustomRecords: destCustomRecords,
	}

	resp, err := svc.client.SendPaymentSync(ctx, sendPaymentRequest)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey":  senderPubkey,
			"amount":        amount,
			"payeePubkey":   destination,
			"paymentHash":   paymentHashHex,
			"preimage":      preimage,
			"customRecords": custom_records,
			"error":         err,
		}).Errorf("Failed to send keysend payment")
		return "", err
	}
	if resp.PaymentError != "" {
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey":  senderPubkey,
			"amount":        amount,
			"payeePubkey":   destination,
			"paymentHash":   paymentHashHex,
			"preimage":      preimage,
			"customRecords": custom_records,
			"paymentError":  resp.PaymentError,
		}).Errorf("Keysend payment has payment error")
		return "", errors.New(resp.PaymentError)
	}
	respPreimage = hex.EncodeToString(resp.PaymentPreimage)
	if respPreimage == "" {
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey":  senderPubkey,
			"amount":        amount,
			"payeePubkey":   destination,
			"paymentHash":   paymentHashHex,
			"preimage":      preimage,
			"customRecords": custom_records,
			"paymentError":  resp.PaymentError,
		}).Errorf("No preimage in keysend response")
		return "", errors.New("No preimage in keysend response")
	}
	svc.Logger.WithFields(logrus.Fields{
		"senderPubkey":  senderPubkey,
		"amount":        amount,
		"payeePubkey":   destination,
		"paymentHash":   paymentHashHex,
		"preimage":      preimage,
		"customRecords": custom_records,
		"respPreimage":  respPreimage,
	}).Info("Keysend payment successful")

	return respPreimage, nil
}

func makePreimageHex() ([]byte, error) {
	bytes := make([]byte, 32) // 32 bytes * 8 bits/byte = 256 bits
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
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
