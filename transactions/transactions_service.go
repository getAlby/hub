package transactions

import (
	"context"
	"encoding/json"

	"github.com/getAlby/nostr-wallet-connect/db"
	"github.com/getAlby/nostr-wallet-connect/lnclient"
	"github.com/getAlby/nostr-wallet-connect/logger"
	"gorm.io/gorm"
)

type transactionsService struct {
	db *gorm.DB
}

const (
	TRANSACTION_STATE_CREATED = "CREATED"
	TRANSACTION_STATE_PENDING = "PENDING"
	TRANSACTION_STATE_PAID    = "PAID"
	TRANSACTION_STATE_FAILED  = "FAILED"
)

func NewTransactionsService(db *gorm.DB) *transactionsService {
	return &transactionsService{
		db: db,
	}
}

func (svc *transactionsService) MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64, lnClient lnclient.LNClient, appId *uint, requestEventId *uint) (*db.Transaction, error) {
	transaction, err := lnClient.MakeInvoice(ctx, amount, description, descriptionHash, expiry)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to create transaction")
		return nil, err
	}

	var preimage *string
	if transaction.Preimage != "" {
		preimage = &transaction.Preimage
	}

	var metadata string
	if transaction.Metadata != nil {
		metadataBytes, err := json.Marshal(transaction.Metadata)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to serialize transaction metadata")
			return nil, err
		}
		metadata = string(metadataBytes)
	}

	dbTransaction := &db.Transaction{
		AppId:          appId,
		RequestEventId: requestEventId,
		Type:           transaction.Type,
		State:          TRANSACTION_STATE_CREATED,
		Amount:         uint(transaction.Amount),
		Fee:            0,
		PaymentRequest: transaction.Invoice,
		PaymentHash:    transaction.PaymentHash,
		Preimage:       preimage,
		Metadata:       metadata,
	}
	err = svc.db.Save(dbTransaction).Error
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to create DB invoice")
		return nil, err
	}
	return dbTransaction, nil
}
