package api

import (
	"context"
	"errors"
	"time"

	"github.com/getAlby/hub/transactions"
)

func (api *api) CreateInvoice(ctx context.Context, amount int64, description string) (*MakeInvoiceResponse, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}
	transaction, err := api.svc.GetTransactionsService().MakeInvoice(ctx, amount, description, "", 0, api.svc.GetLNClient(), nil, nil)
	if err != nil {
		return nil, err
	}
	return toApiTransaction(transaction), nil
}

func (api *api) LookupInvoice(ctx context.Context, paymentHash string) (*LookupInvoiceResponse, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}
	transaction, err := api.svc.GetTransactionsService().LookupTransaction(ctx, paymentHash, api.svc.GetLNClient(), nil)
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
	transactions, err := api.svc.GetTransactionsService().ListTransactions(ctx, 0, 0, limit, offset, false, "", api.svc.GetLNClient())
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
	fee := uint64(0)
	if transaction.Fee != nil {
		fee = *transaction.Fee
	}

	createdAt := transaction.CreatedAt.Format(time.RFC3339)
	var settledAt *string
	if transaction.SettledAt != nil {
		settledAtValue := transaction.SettledAt.Format(time.RFC3339)
		settledAt = &settledAtValue
	}

	return &Transaction{
		Type:            transaction.Type,
		Invoice:         transaction.PaymentRequest,
		Description:     transaction.Description,
		DescriptionHash: transaction.DescriptionHash,
		Preimage:        transaction.Preimage,
		PaymentHash:     transaction.PaymentHash,
		Amount:          transaction.Amount,
		AppId:           transaction.AppId,
		FeesPaid:        fee,
		CreatedAt:       createdAt,
		SettledAt:       settledAt,
		Metadata:        transaction.Metadata,
	}
}
