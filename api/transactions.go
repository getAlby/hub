package api

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/transactions"
	"github.com/sirupsen/logrus"
)

func (api *api) CreateInvoice(ctx context.Context, amount int64, description string) (*MakeInvoiceResponse, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}
	transaction, err := api.svc.GetTransactionsService().MakeInvoice(ctx, amount, description, "", 0, nil, api.svc.GetLNClient(), nil, nil)
	if err != nil {
		return nil, err
	}
	return toApiTransaction(transaction), nil
}

func (api *api) LookupInvoice(ctx context.Context, paymentHash string) (*LookupInvoiceResponse, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}
	transaction, err := api.svc.GetTransactionsService().LookupTransaction(ctx, paymentHash, nil, api.svc.GetLNClient(), nil)
	if err != nil {
		return nil, err
	}
	return toApiTransaction(transaction), nil
}

// TODO: accept offset, limit params for pagination
func (api *api) ListTransactions(ctx context.Context, limit uint64, offset uint64) (*ListTransactionsResponse, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}
	transactions, err := api.svc.GetTransactionsService().ListTransactions(ctx, 0, 0, limit, offset, false, nil, api.svc.GetLNClient(), nil)
	if err != nil {
		return nil, err
	}

	apiTransactions := []Transaction{}
	for _, transaction := range transactions {
		apiTransactions = append(apiTransactions, *toApiTransaction(&transaction))
	}

	return &apiTransactions, nil
}

func (api *api) SendPayment(ctx context.Context, invoice string) (*SendPaymentResponse, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}
	transaction, err := api.svc.GetTransactionsService().SendPaymentSync(ctx, invoice, api.svc.GetLNClient(), nil, nil)
	if err != nil {
		return nil, err
	}
	return toApiTransaction(transaction), nil
}

func toApiTransaction(transaction *transactions.Transaction) *Transaction {

	createdAt := transaction.CreatedAt.Format(time.RFC3339)
	var settledAt *string
	var preimage *string
	if transaction.SettledAt != nil {
		settledAtValue := transaction.SettledAt.Format(time.RFC3339)
		settledAt = &settledAtValue
		preimage = transaction.Preimage
	}

	var metadata *Metadata
	if transaction.Metadata != nil {
		jsonErr := json.Unmarshal(transaction.Metadata, &metadata)
		if jsonErr != nil {
			logger.Logger.WithError(jsonErr).WithFields(logrus.Fields{
				"id":       transaction.ID,
				"metadata": transaction.Metadata,
			}).Error("Failed to deserialize transaction metadata")
		}
	}

	var boostagram *Boostagram
	if transaction.Boostagram != nil {
		var txBoostagram *transactions.Boostagram
		jsonErr := json.Unmarshal(transaction.Boostagram, &txBoostagram)
		if jsonErr != nil {
			logger.Logger.WithError(jsonErr).WithFields(logrus.Fields{
				"id":         transaction.ID,
				"boostagram": transaction.Boostagram,
			}).Error("Failed to deserialize transaction boostagram info")
		}
		boostagram = toApiBoostagram(txBoostagram)
	}

	return &Transaction{
		Type:            transaction.Type,
		Invoice:         transaction.PaymentRequest,
		Description:     transaction.Description,
		DescriptionHash: transaction.DescriptionHash,
		Preimage:        preimage,
		PaymentHash:     transaction.PaymentHash,
		Amount:          transaction.AmountMsat,
		AppId:           transaction.AppId,
		FeesPaid:        transaction.FeeMsat,
		CreatedAt:       createdAt,
		SettledAt:       settledAt,
		Metadata:        metadata,
		Boostagram:      boostagram,
	}
}

func toApiBoostagram(boostagram *transactions.Boostagram) *Boostagram {
	return &Boostagram{
		AppName:        boostagram.AppName,
		Name:           boostagram.Name,
		Podcast:        boostagram.Podcast,
		URL:            boostagram.URL,
		Episode:        boostagram.Episode,
		FeedId:         boostagram.FeedId,
		ItemId:         boostagram.ItemId,
		Timestamp:      boostagram.Timestamp,
		Message:        boostagram.Message,
		SenderId:       boostagram.SenderId,
		SenderName:     boostagram.SenderName,
		Time:           boostagram.Time,
		Action:         boostagram.Action,
		ValueMsatTotal: boostagram.ValueMsatTotal,
	}
}
