package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/getAlby/nostr-wallet-connect/lnd"
	decodepay "github.com/nbd-wtf/ln-decodepay"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/lightningnetwork/lnd/lnrpc"
)

type LNClient interface {
	SendPaymentSync(ctx context.Context, senderPubkey string, payReq string) (preimage string, err error)
	SendKeysend(ctx context.Context, senderPubkey string, amount int64, destination, preimage string, custom_records []TLVRecord) (preImage string, err error)
	GetBalance(ctx context.Context, senderPubkey string) (balance int64, err error)
	GetInfo(ctx context.Context, senderPubkey string) (info *NodeInfo, err error)
	MakeInvoice(ctx context.Context, senderPubkey string, amount int64, description string, descriptionHash string, expiry int64) (transaction *Nip47Transaction, err error)
	LookupInvoice(ctx context.Context, senderPubkey string, paymentHash string) (transaction *Nip47Transaction, err error)
	ListTransactions(ctx context.Context, senderPubkey string, from, until, limit, offset uint64, unpaid bool, invoiceType string) (transactions []Nip47Transaction, err error)
	Shutdown() error
}

// wrap it again :sweat_smile:
// todo: drop dependency on lndhub package
type LNDService struct {
	client *lnd.LNDWrapper
	db     *gorm.DB
	Logger *logrus.Logger
}

func (svc *LNDService) GetBalance(ctx context.Context, senderPubkey string) (balance int64, err error) {
	resp, err := svc.client.ChannelBalance(ctx, &lnrpc.ChannelBalanceRequest{})
	if err != nil {
		return 0, err
	}
	return int64(resp.LocalBalance.Sat), nil
}

func (svc *LNDService) ListTransactions(ctx context.Context, senderPubkey string, from, until, limit, offset uint64, unpaid bool, invoiceType string) (transactions []Nip47Transaction, err error) {
	// Fetch invoices
	var invoices []*lnrpc.Invoice
	if invoiceType == "" || invoiceType == "incoming" {
		incomingResp, err := svc.client.ListInvoices(ctx, &lnrpc.ListInvoiceRequest{Reversed: true, NumMaxInvoices: limit, IndexOffset: offset})
		if err != nil {
			return nil, err
		}
		invoices = incomingResp.Invoices
	}
	for _, invoice := range invoices {
		// this will cause retrieved amount to be less than limit if unpaid is false
		if !unpaid && invoice.State != lnrpc.Invoice_SETTLED {
			continue
		}

		transaction := lndInvoiceToTransaction(invoice)
		transactions = append(transactions, *transaction)
	}
	// Fetch payments
	var payments []*lnrpc.Payment
	if invoiceType == "" || invoiceType == "outgoing" {
		// Not just pending but failed payments will also be included because of IncludeIncomplete
		outgoingResp, err := svc.client.ListPayments(ctx, &lnrpc.ListPaymentsRequest{Reversed: true, MaxPayments: limit, IndexOffset: offset, IncludeIncomplete: unpaid})
		if err != nil {
			return nil, err
		}
		payments = outgoingResp.Payments
	}
	for _, payment := range payments {
		if payment.Status == lnrpc.Payment_FAILED {
			// don't return failed payments for now
			// this will cause retrieved amount to be less than limit
			continue
		}
		var paymentRequest decodepay.Bolt11
		var expiresAt *int64
		var description string
		var descriptionHash string
		if payment.PaymentRequest != "" {
			paymentRequest, err = decodepay.Decodepay(strings.ToLower(payment.PaymentRequest))
			if err != nil {
				svc.Logger.WithFields(logrus.Fields{
					"bolt11": payment.PaymentRequest,
				}).Errorf("Failed to decode bolt11 invoice: %v", err)

				return nil, err
			}
			expiresAtUnix := time.UnixMilli(int64(paymentRequest.CreatedAt) * 1000).Add(time.Duration(paymentRequest.Expiry) * time.Second).Unix()
			expiresAt = &expiresAtUnix
			description = paymentRequest.Description
			descriptionHash = paymentRequest.DescriptionHash
		}

		var settledAt *int64
		if payment.Status == lnrpc.Payment_SUCCEEDED {
			// FIXME: how to get the actual settled at time?
			settledAtUnix := time.Unix(0, payment.CreationTimeNs).Unix()
			settledAt = &settledAtUnix
		}

		transaction := Nip47Transaction{
			Type:            "outgoing",
			Invoice:         payment.PaymentRequest,
			Preimage:        payment.PaymentPreimage,
			PaymentHash:     payment.PaymentHash,
			Amount:          payment.ValueMsat,
			FeesPaid:        payment.FeeMsat,
			CreatedAt:       time.Unix(0, payment.CreationTimeNs).Unix(),
			Description:     description,
			DescriptionHash: descriptionHash,
			ExpiresAt:       expiresAt,
			SettledAt:       settledAt,
			//TODO: Metadata:  (e.g. keysend),
		}
		transactions = append(transactions, transaction)
	}

	// sort by created date descending
	sort.SliceStable(transactions, func(i, j int) bool {
		return transactions[i].CreatedAt > transactions[j].CreatedAt
	})

	return transactions, nil
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

func (svc *LNDService) MakeInvoice(ctx context.Context, senderPubkey string, amount int64, description string, descriptionHash string, expiry int64) (transaction *Nip47Transaction, err error) {
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
			return nil, errors.New("Description hash must be 32 bytes hex")
		}
	}

	resp, err := svc.client.AddInvoice(ctx, &lnrpc.Invoice{ValueMsat: amount, Memo: description, DescriptionHash: descriptionHashBytes, Expiry: expiry})
	if err != nil {
		return nil, err
	}

	inv, err := svc.client.LookupInvoice(ctx, &lnrpc.PaymentHash{RHash: resp.RHash})
	if err != nil {
		return nil, err
	}

	transaction = lndInvoiceToTransaction(inv)
	return transaction, nil
}

func (svc *LNDService) LookupInvoice(ctx context.Context, senderPubkey string, paymentHash string) (transaction *Nip47Transaction, err error) {
	paymentHashBytes, err := hex.DecodeString(paymentHash)

	if err != nil || len(paymentHashBytes) != 32 {
		svc.Logger.WithFields(logrus.Fields{
			"paymentHash": paymentHash,
		}).Errorf("Invalid payment hash")
		return nil, errors.New("Payment hash must be 32 bytes hex")
	}

	lndInvoice, err := svc.client.LookupInvoice(ctx, &lnrpc.PaymentHash{RHash: paymentHashBytes})
	if err != nil {
		return nil, err
	}

	transaction = lndInvoiceToTransaction(lndInvoice)
	return transaction, nil
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
		AmtMsat:           amount,
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

func NewLNDService(svc *Service, lndAddress, lndCertHex, lndMacaroonHex string) (result LNClient, err error) {
	if lndAddress == "" || lndCertHex == "" || lndMacaroonHex == "" {
		return nil, errors.New("One or more required LND configuration are missing")
	}

	lndClient, err := lnd.NewLNDclient(lnd.LNDoptions{
		Address:     lndAddress,
		CertHex:     lndCertHex,
		MacaroonHex: lndMacaroonHex,
	}, svc.ctx)
	if err != nil {
		svc.Logger.Errorf("Failed to create new LND client %v", err)
		return nil, err
	}
	info, err := lndClient.GetInfo(svc.ctx, &lnrpc.GetInfoRequest{})
	if err != nil {
		return nil, err
	}

	lndService := &LNDService{client: lndClient, Logger: svc.Logger, db: svc.db}

	svc.Logger.Infof("Connected to LND - alias %s", info.Alias)

	return lndService, nil
}

func (svc *LNDService) Shutdown() error {
	return nil
}

func lndInvoiceToTransaction(invoice *lnrpc.Invoice) *Nip47Transaction {
	var settledAt *int64
	var preimage string
	if invoice.State == lnrpc.Invoice_SETTLED {
		settledAt = &invoice.SettleDate
		// only set preimage if invoice is settled
		preimage = hex.EncodeToString(invoice.RPreimage)
	}
	var expiresAt *int64
	if invoice.Expiry > 0 {
		expiresAtUnix := invoice.CreationDate + invoice.Expiry
		expiresAt = &expiresAtUnix
	}

	return &Nip47Transaction{
		Type:            "incoming",
		Invoice:         invoice.PaymentRequest,
		Description:     invoice.Memo,
		DescriptionHash: hex.EncodeToString(invoice.DescriptionHash),
		Preimage:        preimage,
		PaymentHash:     hex.EncodeToString(invoice.RHash),
		Amount:          invoice.ValueMsat,
		FeesPaid:        invoice.AmtPaidMsat,
		CreatedAt:       invoice.CreationDate,
		SettledAt:       settledAt,
		ExpiresAt:       expiresAt,
		// TODO: Metadata (e.g. keysend)
	}
}
