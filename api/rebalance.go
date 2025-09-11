package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/version"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
)

func (api *api) RebalanceChannel(ctx context.Context, rebalanceChannelRequest *RebalanceChannelRequest) (*RebalanceChannelResponse, error) {
	if api.svc.GetLNClient() == nil {
		return nil, errors.New("LNClient not started")
	}

	receiveMetadata := map[string]interface{}{
		"receive_through": rebalanceChannelRequest.ReceiveThroughNodePubkey,
	}

	receiveInvoice, err := api.svc.GetTransactionsService().MakeInvoice(ctx, rebalanceChannelRequest.AmountSat*1000, "Alby Hub Rebalance through "+rebalanceChannelRequest.ReceiveThroughNodePubkey, "", 0, receiveMetadata, api.svc.GetLNClient(), nil, nil)
	if err != nil {
		logger.Logger.WithError(err).Error("failed to generate rebalance receive invoice")
		return nil, err
	}

	type rspCreateOrderRequest struct {
		Token                   string `json:"token"`
		PayRequest              string `json:"pay_request"`
		PayThroughThisPublicKey string `json:"pay_through_this_public_key"`
	}

	newRspCreateOrderRequest := rspCreateOrderRequest{
		Token:                   "alby-hub",
		PayRequest:              receiveInvoice.PaymentRequest,
		PayThroughThisPublicKey: rebalanceChannelRequest.ReceiveThroughNodePubkey,
	}

	payloadBytes, err := json.Marshal(newRspCreateOrderRequest)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(payloadBytes)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, api.cfg.GetEnv().RebalanceServiceUrl+"/api/rebalance/v1/create_order", bodyReader)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"request": newRspCreateOrderRequest,
		}).Error("Failed to create new rebalance request")
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "AlbyHub/"+version.Tag)

	client := http.Client{
		Timeout: time.Second * 60,
	}

	res, err := client.Do(req)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"request": newRspCreateOrderRequest,
		}).Error("Failed to request new rebalance order")
		return nil, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"request": newRspCreateOrderRequest,
		}).Error("Failed to read response body")
		return nil, errors.New("failed to read response body")
	}

	if res.StatusCode >= 300 {
		logger.Logger.WithFields(logrus.Fields{
			"request":    newRspCreateOrderRequest,
			"body":       string(body),
			"statusCode": res.StatusCode,
		}).Error("rebalance create_order endpoint returned non-success code")
		return nil, fmt.Errorf("rebalance create_order endpoint returned non-success code: %s", string(body))
	}

	type rspRebalanceCreateOrderResponse struct {
		OrderId    string `json:"order_id"`
		PayRequest string `json:"pay_request"`
	}

	var rebalanceCreateOrderResponse rspRebalanceCreateOrderResponse

	err = json.Unmarshal(body, &rebalanceCreateOrderResponse)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"request": newRspCreateOrderRequest,
		}).Error("Failed to deserialize json")
		return nil, fmt.Errorf("failed to deserialize json from rebalance create order response: %s", string(body))
	}

	logger.Logger.WithField("response", rebalanceCreateOrderResponse).Info("New rebalance order created")

	paymentRequest, err := decodepay.Decodepay(rebalanceCreateOrderResponse.PayRequest)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to decode bolt11 invoice")
		return nil, err
	}

	if paymentRequest.MSatoshi > int64(float64(rebalanceChannelRequest.AmountSat)*float64(1000)*float64(1.003)+1 /*0.3% fees*/) {
		return nil, errors.New("rebalance payment is more expensive than expected")
	}

	payMetadata := map[string]interface{}{
		"receive_through": rebalanceChannelRequest.ReceiveThroughNodePubkey,
		"amount_sat":      rebalanceChannelRequest.AmountSat,
		"order_id":        rebalanceCreateOrderResponse.OrderId,
	}

	payRebalanceInvoiceResponse, err := api.svc.GetTransactionsService().SendPaymentSync(ctx, rebalanceCreateOrderResponse.PayRequest, nil, payMetadata, api.svc.GetLNClient(), nil, nil)

	if err != nil {
		logger.Logger.WithError(err).Error("failed to pay rebalance invoice")
		return nil, err
	}

	api.eventPublisher.Publish(&events.Event{
		Event:      "nwc_rebalance_succeeded",
		Properties: map[string]interface{}{},
	})

	return &RebalanceChannelResponse{
		TotalFeeSat: uint64(paymentRequest.MSatoshi)/1000 + payRebalanceInvoiceResponse.FeeMsat/1000 - rebalanceChannelRequest.AmountSat,
	}, nil
}
