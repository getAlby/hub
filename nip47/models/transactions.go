package models

import (
	"github.com/getAlby/nostr-wallet-connect/transactions"
)

func ToNip47Transaction(transaction *transactions.Transaction) *Transaction {
	preimage := ""
	if transaction.Preimage != nil {
		preimage = *transaction.Preimage
	}

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
	if transaction.SettledAt != nil {
		settledAtUnix := transaction.SettledAt.Unix()
		settledAt = &settledAtUnix
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
		// FIXME: re-add
		//Metadata:        metadata,
	}
}
