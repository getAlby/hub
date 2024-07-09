package models

import (
	"encoding/json"

	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/transactions"
	"github.com/sirupsen/logrus"
)

func ToNip47Transaction(transaction *transactions.Transaction) *Transaction {
	fees := int64(0)
	if transaction.Fee != nil {
		fees = int64(*transaction.Fee)
	}

	var expiresAt *int64
	if transaction.ExpiresAt != nil {
		expiresAtUnix := transaction.ExpiresAt.Unix()
		expiresAt = &expiresAtUnix
	}

	var settledAt *int64
	preimage := ""
	if transaction.SettledAt != nil {
		settledAtUnix := transaction.SettledAt.Unix()
		settledAt = &settledAtUnix
		preimage = *transaction.Preimage
	}

	var metadata interface{}
	if transaction.Metadata != "" {
		jsonErr := json.Unmarshal([]byte(transaction.Metadata), &metadata)
		if jsonErr != nil {
			logger.Logger.WithError(jsonErr).WithFields(logrus.Fields{
				"id":       transaction.ID,
				"metadata": transaction.Metadata,
			}).Error("Failed to deserialize transaction metadata")
		}
	}

	return &Transaction{
		Type:            transaction.Type,
		Invoice:         transaction.PaymentRequest,
		Description:     transaction.Description,
		DescriptionHash: transaction.DescriptionHash,
		Preimage:        preimage,
		PaymentHash:     transaction.PaymentHash,
		Amount:          int64(transaction.Amount),
		FeesPaid:        fees,
		CreatedAt:       transaction.CreatedAt.Unix(),
		ExpiresAt:       expiresAt,
		SettledAt:       settledAt,
		Metadata:        metadata,
	}
}
