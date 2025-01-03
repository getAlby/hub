package transactions

import (
	"encoding/json"
	"strings"

	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	"github.com/sirupsen/logrus"
)

func ToNip47Transaction(transaction *Transaction) *models.Transaction {
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

	var metadata map[string]interface{}
	if transaction.Metadata != nil {
		jsonErr := json.Unmarshal(transaction.Metadata, &metadata)
		if jsonErr != nil {
			logger.Logger.WithError(jsonErr).WithFields(logrus.Fields{
				"payment_hash": transaction.PaymentHash,
				"metadata":     transaction.Metadata,
			}).Error("Failed to deserialize transaction metadata")
		}
	}

	return &models.Transaction{
		Type:            transaction.Type,
		State:           strings.ToLower(transaction.State),
		Invoice:         transaction.PaymentRequest,
		Description:     transaction.Description,
		DescriptionHash: transaction.DescriptionHash,
		Preimage:        preimage,
		PaymentHash:     transaction.PaymentHash,
		Amount:          int64(transaction.AmountMsat),
		FeesPaid:        int64(transaction.FeeMsat),
		CreatedAt:       transaction.CreatedAt.Unix(),
		ExpiresAt:       expiresAt,
		SettledAt:       settledAt,
		Metadata:        metadata,
	}
}
