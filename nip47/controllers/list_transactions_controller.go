package controllers

import (
	"context"

	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

type listTransactionsParams struct {
	From           uint64 `json:"from,omitempty"`
	Until          uint64 `json:"until,omitempty"`
	Limit          uint64 `json:"limit,omitempty"`
	Offset         uint64 `json:"offset,omitempty"`
	Unpaid         bool   `json:"unpaid,omitempty"`
	UnpaidOutgoing bool   `json:"unpaid_outgoing,omitempty"`
	UnpaidIncoming bool   `json:"unpaid_incoming,omitempty"`
	Type           string `json:"type,omitempty"`
}

type listTransactionsResponse struct {
	Transactions []models.Transaction `json:"transactions"`
	TotalCount   uint64               `json:"total_count"`
}

func (controller *nip47Controller) HandleListTransactionsEvent(ctx context.Context, nip47Request *models.Request, requestEventId uint, appId uint, publishResponse publishFunc) {

	listParams := &listTransactionsParams{}
	resp := decodeRequest(nip47Request, listParams)
	if resp != nil {
		publishResponse(resp, nostr.Tags{})
		return
	}

	logger.Logger.WithFields(logrus.Fields{
		"params":           listParams,
		"request_event_id": requestEventId,
	}).Debug("Fetching transactions")

	limit := listParams.Limit
	maxLimit := uint64(50)
	if limit == 0 || limit > maxLimit {
		// make sure a sensible limit is passed
		limit = maxLimit
	}
	var transactionType *string
	if listParams.Type != "" {
		transactionType = &listParams.Type
	}

	dbTransactions, totalCount, err := controller.transactionsService.ListTransactions(ctx, listParams.From, listParams.Until, limit, listParams.Offset, listParams.Unpaid || listParams.UnpaidOutgoing, listParams.Unpaid || listParams.UnpaidIncoming, transactionType, controller.lnClient, &appId, false)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"params":           listParams,
			"request_event_id": requestEventId,
		}).WithError(err).Error("Failed to fetch transactions")

		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error:      mapNip47Error(err),
		}, nostr.Tags{})
		return
	}

	transactions := []models.Transaction{}
	for _, dbTransaction := range dbTransactions {
		transactions = append(transactions, *models.ToNip47Transaction(&dbTransaction))
	}

	responsePayload := &listTransactionsResponse{
		Transactions: transactions,
		TotalCount:   totalCount,
	}

	publishResponse(&models.Response{
		ResultType: nip47Request.Method,
		Result:     responsePayload,
	}, nostr.Tags{})
}
