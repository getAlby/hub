package api

import (
	"context"
	"errors"
	"slices"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
)

func getTransactionTimestampMs(tx db.Transaction) int64 {
	if tx.SettledAt != nil {
		return tx.SettledAt.UnixMilli()
	}
	return tx.UpdatedAt.UnixMilli()
}

func getNetSat(tx db.Transaction) int64 {
	amountSat := int64(tx.AmountMsat / 1000)
	feeSat := int64(tx.FeeMsat / 1000)

	switch tx.Type {
	case constants.TRANSACTION_TYPE_INCOMING:
		return amountSat
	case constants.TRANSACTION_TYPE_OUTGOING:
		return -(amountSat + feeSat)
	default:
		return 0
	}
}

func (api *api) GetHomeChartsData(ctx context.Context, from uint64) (*HomeChartsResponse, error) {
	lnClient := api.svc.GetLNClient()
	if lnClient == nil {
		return nil, errors.New("LNClient not started")
	}

	transactions, _, err := api.svc.GetTransactionsService().ListTransactions(
		ctx,
		from,
		0,
		0,
		0,
		false,
		false,
		nil,
		lnClient,
		nil,
		false,
	)
	if err != nil {
		return nil, err
	}

	points := make([]HomeChartsTxPoint, 0, len(transactions))
	hasIncomingDeposit := false
	var netFlowsSat int64

	for _, tx := range transactions {
		netSat := getNetSat(tx)
		if tx.Type == constants.TRANSACTION_TYPE_INCOMING && tx.AmountMsat > 0 {
			hasIncomingDeposit = true
		}
		netFlowsSat += netSat
		points = append(points, HomeChartsTxPoint{
			Timestamp: getTransactionTimestampMs(tx),
			NetSat:    netSat,
		})
	}

	slices.SortFunc(points, func(a, b HomeChartsTxPoint) int {
		if a.Timestamp < b.Timestamp {
			return -1
		}
		if a.Timestamp > b.Timestamp {
			return 1
		}
		return 0
	})

	return &HomeChartsResponse{
		TxPoints:           points,
		HasIncomingDeposit: hasIncomingDeposit,
		NetFlowsSat:        netFlowsSat,
	}, nil
}
