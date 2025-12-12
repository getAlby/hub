package api

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/transactions"
	"github.com/sirupsen/logrus"
)

func (api *api) CreateInvoice(ctx context.Context, amount uint64, description string) (*MakeInvoiceResponse, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}
	transaction, err := api.svc.GetTransactionsService().MakeInvoice(ctx, amount, description, "", 0, nil, api.svc.GetLNClient(), nil, nil, nil)
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

func (api *api) ListTransactions(ctx context.Context, appId *uint, limit uint64, offset uint64) (*ListTransactionsResponse, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}

	forceFilterByAppId := false
	if appId != nil {
		forceFilterByAppId = true
	}

	transactions, totalCount, err := api.svc.GetTransactionsService().ListTransactions(ctx, 0, 0, limit, offset, true, false, nil, api.svc.GetLNClient(), appId, forceFilterByAppId)
	if err != nil {
		return nil, err
	}

	apiTransactions := []Transaction{}
	for _, transaction := range transactions {
		apiTransactions = append(apiTransactions, *toApiTransaction(&transaction))
	}

	return &ListTransactionsResponse{
		Transactions: apiTransactions,
		TotalCount:   totalCount,
	}, nil
}

func (api *api) SendPayment(ctx context.Context, invoice string, amountMsat *uint64, metadata map[string]interface{}) (*SendPaymentResponse, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}
	transaction, err := api.svc.GetTransactionsService().SendPaymentSync(invoice, amountMsat, metadata, api.svc.GetLNClient(), nil, nil)
	if err != nil {
		return nil, err
	}
	return toApiTransaction(transaction), nil
}

func toApiTransaction(transaction *transactions.Transaction) *Transaction {

	updatedAt := transaction.UpdatedAt.Format(time.RFC3339)
	createdAt := transaction.CreatedAt.Format(time.RFC3339)
	var settledAt *string
	var preimage *string
	if transaction.SettledAt != nil {
		settledAtValue := transaction.SettledAt.Format(time.RFC3339)
		settledAt = &settledAtValue
		preimage = transaction.Preimage
	}

	var metadata Metadata
	if transaction.Metadata != nil {
		jsonErr := json.Unmarshal(transaction.Metadata, &metadata)
		if jsonErr != nil {
			logger.Logger.WithError(jsonErr).WithFields(logrus.Fields{
				"payment_hash": transaction.PaymentHash,
				"metadata":     transaction.Metadata,
			}).Error("Failed to deserialize transaction metadata")
		}
	}

	var boostagram *Boostagram
	if transaction.Boostagram != nil {
		var txBoostagram transactions.Boostagram
		jsonErr := json.Unmarshal(transaction.Boostagram, &txBoostagram)
		if jsonErr != nil {
			logger.Logger.WithError(jsonErr).WithFields(logrus.Fields{
				"payment_hash": transaction.PaymentHash,
				"boostagram":   transaction.Boostagram,
			}).Error("Failed to deserialize transaction boostagram info")
		}
		boostagram = toApiBoostagram(&txBoostagram)
	}

	return &Transaction{
		Type:            transaction.Type,
		State:           strings.ToLower(transaction.State),
		Invoice:         transaction.PaymentRequest,
		Description:     transaction.Description,
		DescriptionHash: transaction.DescriptionHash,
		Preimage:        preimage,
		PaymentHash:     transaction.PaymentHash,
		Amount:          transaction.AmountMsat,
		AppId:           transaction.AppId,
		FeesPaid:        transaction.FeeMsat,
		UpdatedAt:       updatedAt,
		CreatedAt:       createdAt,
		SettledAt:       settledAt,
		Metadata:        metadata,
		Boostagram:      boostagram,
		FailureReason:   transaction.FailureReason,
	}
}

func (api *api) Transfer(ctx context.Context, fromAppId *uint, toAppId *uint, amountMsat uint64) error {
	if api.svc.GetLNClient() == nil {
		return errors.New("LNClient not started")
	}

	for _, appId := range []*uint{fromAppId, toAppId} {
		if appId != nil {
			dbApp := api.appsSvc.GetAppById(*appId)
			if dbApp == nil {
				return errors.New("app does not exist")
			}
			if !dbApp.Isolated {
				return errors.New("app is not isolated")
			}
		}
	}

	transaction, err := api.svc.GetTransactionsService().MakeInvoice(ctx, amountMsat, "transfer", "", 0, nil, api.svc.GetLNClient(), toAppId, nil, nil)

	if err != nil {
		return err
	}

	_, err = api.svc.GetTransactionsService().SendPaymentSync(transaction.PaymentRequest, nil, nil, api.svc.GetLNClient(), fromAppId, nil)
	return err
}

func toApiBoostagram(boostagram *transactions.Boostagram) *Boostagram {
	return &Boostagram{
		AppName:        boostagram.AppName,
		Name:           boostagram.Name,
		Podcast:        boostagram.Podcast,
		URL:            boostagram.URL,
		Episode:        boostagram.Episode.String(),
		FeedId:         boostagram.FeedId.String(),
		ItemId:         boostagram.ItemId.String(),
		Timestamp:      boostagram.Timestamp,
		Message:        boostagram.Message,
		SenderId:       boostagram.SenderId.String(),
		SenderName:     boostagram.SenderName,
		Time:           boostagram.Time,
		Action:         boostagram.Action,
		ValueMsatTotal: boostagram.ValueMsatTotal,
	}
}
