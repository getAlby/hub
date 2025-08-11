package wails

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/getAlby/hub/alby"
	"github.com/getAlby/hub/api"
	"github.com/getAlby/hub/logger"
)

type WailsRequestRouterResponse struct {
	Body  interface{} `json:"body"`
	Error string      `json:"error"`
}

// TODO: make this match echo
func (app *WailsApp) WailsRequestRouter(route string, method string, body string) WailsRequestRouterResponse {
	ctx := app.ctx

	// the grouping is done to avoid other parameters like &unused=true
	albyCallbackRegex := regexp.MustCompile(
		`/api/alby/callback\?code=([^&]+)(&.*)?`,
	)

	authCodeMatch := albyCallbackRegex.FindStringSubmatch(route)

	switch {
	case len(authCodeMatch) > 1:
		code := authCodeMatch[1]

		err := app.svc.GetAlbyOAuthSvc().CallbackHandler(ctx, code, app.svc.GetLNClient())
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: nil, Error: ""}
	}

	appv2Regex := regexp.MustCompile(
		`/api/v2/apps/([0-9a-f]+)`,
	)

	appv2Match := appv2Regex.FindStringSubmatch(route)

	switch {
	case len(appv2Match) > 1:
		appIdStr := appv2Match[1]

		appId, err := strconv.ParseUint(appIdStr, 10, 64)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}

		dbApp := app.appsSvc.GetAppById(uint(appId))
		if dbApp == nil {
			return WailsRequestRouterResponse{Body: nil, Error: "App does not exist"}
		}

		switch method {
		case "GET":
			app := app.api.GetApp(dbApp)
			return WailsRequestRouterResponse{Body: app, Error: ""}
		}
	}

	appRegex := regexp.MustCompile(
		`/api/apps/([0-9a-f]+)`,
	)

	appMatch := appRegex.FindStringSubmatch(route)

	switch {
	case len(appMatch) > 1:
		pubkey := appMatch[1]
		dbApp := app.appsSvc.GetAppByPubkey(pubkey)
		if dbApp == nil {
			return WailsRequestRouterResponse{Body: nil, Error: "App does not exist"}
		}

		switch method {
		case "GET":
			app := app.api.GetApp(dbApp)
			return WailsRequestRouterResponse{Body: app, Error: ""}
		case "PATCH":
			updateAppRequest := &api.UpdateAppRequest{}
			err := json.Unmarshal([]byte(body), updateAppRequest)
			if err != nil {
				logger.Logger.WithFields(logrus.Fields{
					"route":  route,
					"method": method,
					"body":   body,
				}).WithError(err).Error("Failed to decode request to wails router")
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			err = app.api.UpdateApp(dbApp, updateAppRequest)
			if err != nil {
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			return WailsRequestRouterResponse{Body: nil, Error: ""}
		case "DELETE":
			err := app.api.DeleteApp(dbApp)
			if err != nil {
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			return WailsRequestRouterResponse{Body: nil, Error: ""}
		}
	}

	appTopupRegex := regexp.MustCompile(
		`/api/apps/([0-9a-f]+)/topup`,
	)

	appTopupMatch := appTopupRegex.FindStringSubmatch(route)

	switch {
	case len(appTopupMatch) > 1:
		pubkey := appTopupMatch[1]
		dbApp := app.appsSvc.GetAppByPubkey(pubkey)
		if dbApp == nil {
			return WailsRequestRouterResponse{Body: nil, Error: "App does not exist"}
		}

		topupIsolatedAppRequest := &api.TopupIsolatedAppRequest{}
		err := json.Unmarshal([]byte(body), topupIsolatedAppRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		err = app.api.TopupIsolatedApp(ctx, dbApp, topupIsolatedAppRequest.AmountSat*1000)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: nil, Error: ""}
	}

	appDrawDownRegex := regexp.MustCompile(
		`/api/apps/([0-9a-f]+)/drawdown`,
	)

	appDrawDownMatch := appDrawDownRegex.FindStringSubmatch(route)

	switch {
	case len(appDrawDownMatch) > 1:
		pubkey := appDrawDownMatch[1]
		dbApp := app.appsSvc.GetAppByPubkey(pubkey)
		if dbApp == nil {
			return WailsRequestRouterResponse{Body: nil, Error: "App does not exist"}
		}

		drawDownIsolatedAppRequest := &api.DrawDownIsolatedAppRequest{}
		err := json.Unmarshal([]byte(body), drawDownIsolatedAppRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		err = app.api.DrawDownIsolatedApp(ctx, dbApp, drawDownIsolatedAppRequest.AmountSat*1000)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: nil, Error: ""}
	}

	// list apps
	if strings.HasPrefix(route, "/api/apps") && method == "GET" {
		limit := uint64(0)
		offset := uint64(0)
		var filtersJSON string
		var orderBy string

		// Extract limit and offset parameters
		paramRegex := regexp.MustCompile(`[?&](limit|offset|filters|order_by)=([^&]+)`)
		paramMatches := paramRegex.FindAllStringSubmatch(route, -1)
		for _, match := range paramMatches {
			switch match[1] {
			case "limit":
				if parsedLimit, err := strconv.ParseUint(match[2], 10, 64); err == nil {
					limit = parsedLimit
				}
			case "offset":
				if parsedOffset, err := strconv.ParseUint(match[2], 10, 64); err == nil {
					offset = parsedOffset
				}
			case "filters":
				filtersJSON = match[2]
			case "order_by":
				orderBy = match[2]
			}
		}

		var filters api.ListAppsFilters
		if filtersJSON != "" {
			err := json.Unmarshal([]byte(filtersJSON), &filters)
			if err != nil {
				logger.Logger.WithError(err).WithFields(logrus.Fields{
					"filters": filtersJSON,
				}).Error("Failed to deserialize app filters")
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
		}

		apps, err := app.api.ListApps(limit, offset, filters, orderBy)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: apps, Error: ""}
	}

	peerChannelRegex := regexp.MustCompile(
		`/api/peers/([^/]+)/channels/([^/]+)`,
	)

	peerChannelMatch := peerChannelRegex.FindStringSubmatch(route)

	switch {
	case len(peerChannelMatch) == 3:
		peerId := peerChannelMatch[1]
		channelId := peerChannelMatch[2]
		switch method {
		case "PATCH":
			updateChannelRequest := &api.UpdateChannelRequest{}
			err := json.Unmarshal([]byte(body), updateChannelRequest)
			if err != nil {
				logger.Logger.WithFields(logrus.Fields{
					"route":  route,
					"method": method,
					"body":   body,
				}).WithError(err).Error("Failed to decode request to wails router")
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			updateChannelRequest.ChannelId = channelId
			updateChannelRequest.NodeId = peerId

			err = app.api.UpdateChannel(ctx, updateChannelRequest)
			if err != nil {
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			return WailsRequestRouterResponse{Body: nil, Error: ""}
		case "DELETE":
			channelIdParts := strings.SplitN(channelId, "?", 2)

			var force bool
			if len(channelIdParts) == 2 {
				channelId = channelIdParts[0]
				queryParams := channelIdParts[1]
				force = strings.Contains(queryParams, "force=true")
			}

			closeChannelResponse, err := app.api.CloseChannel(ctx, peerId, channelId, force)
			if err != nil {
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			return WailsRequestRouterResponse{Body: closeChannelResponse, Error: ""}
		}
	}

	peerRegex := regexp.MustCompile(
		`/api/peers/([^/]+)`,
	)

	peerMatch := peerRegex.FindStringSubmatch(route)

	switch {
	case len(peerMatch) == 2:
		peerId := peerMatch[1]
		switch method {
		case "DELETE":
			err := app.api.DisconnectPeer(ctx, peerId)
			if err != nil {
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			return WailsRequestRouterResponse{Body: nil, Error: ""}
		}
	}

	networkGraphRegex := regexp.MustCompile(
		`/api/node/network-graph\?nodeIds=(.+)`,
	)

	networkGraphMatch := networkGraphRegex.FindStringSubmatch(route)

	switch {
	case len(networkGraphMatch) == 2:
		nodeIds := networkGraphMatch[1]
		networkGraphResponse, err := app.api.GetNetworkGraph(ctx, strings.Split(nodeIds, ","))
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: networkGraphResponse, Error: ""}
	}

	mempoolApiRegex := regexp.MustCompile(
		`/api/mempool\?endpoint=(.+)`,
	)
	mempoolApiEndpointMatch := mempoolApiRegex.FindStringSubmatch(route)

	switch {
	case len(mempoolApiEndpointMatch) > 1:
		endpoint := mempoolApiEndpointMatch[1]
		node, err := app.api.RequestMempoolApi(endpoint)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}

		return WailsRequestRouterResponse{Body: node, Error: ""}
	}

	transactionRegex := regexp.MustCompile(
		`/api/transactions/([0-9a-fA-F]+)`,
	)
	paymentHashMatch := transactionRegex.FindStringSubmatch(route)

	switch {
	case len(paymentHashMatch) > 1:
		paymentHash := paymentHashMatch[1]
		paymentInfo, err := app.api.LookupInvoice(ctx, paymentHash)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}

		return WailsRequestRouterResponse{Body: paymentInfo, Error: ""}
	}

	listTransactionsRegex := regexp.MustCompile(
		`/api/transactions`,
	)

	switch {
	case listTransactionsRegex.MatchString(route):
		limit := uint64(20)
		offset := uint64(0)
		var appId *uint

		// Extract limit and offset parameters
		paramRegex := regexp.MustCompile(`[?&](limit|offset|appId)=([^&]+)`)
		paramMatches := paramRegex.FindAllStringSubmatch(route, -1)
		for _, match := range paramMatches {
			switch match[1] {
			case "limit":
				if parsedLimit, err := strconv.ParseUint(match[2], 10, 64); err == nil {
					limit = parsedLimit
				}
			case "offset":
				if parsedOffset, err := strconv.ParseUint(match[2], 10, 64); err == nil {
					offset = parsedOffset
				}
			case "appId":
				if parsedAppId, err := strconv.ParseUint(match[2], 10, 64); err == nil {
					var unsignedAppId = uint(parsedAppId)
					appId = &unsignedAppId
				}
			}
		}

		transactions, err := app.api.ListTransactions(ctx, appId, limit, offset)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: transactions, Error: ""}
	}

	paymentRegex := regexp.MustCompile(
		`/api/payments/([0-9a-zA-Z]+)`,
	)
	invoiceMatch := paymentRegex.FindStringSubmatch(route)

	switch {
	case len(invoiceMatch) > 1:
		invoice := invoiceMatch[1]
		payRequest := &api.PayInvoiceRequest{}
		if body != "" {
			err := json.Unmarshal([]byte(body), payRequest)
			if err != nil {
				logger.Logger.WithFields(logrus.Fields{
					"route":  route,
					"method": method,
					"body":   body,
				}).WithError(err).Error("Failed to decode request to wails router")
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
		}
		paymentResponse, err := app.api.SendPayment(ctx, invoice, payRequest.Amount, payRequest.Metadata)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}

		return WailsRequestRouterResponse{Body: paymentResponse, Error: ""}
	}

	switch route {
	case "/api/alby/info":
		info, err := app.svc.GetAlbyOAuthSvc().GetInfo(ctx)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: info, Error: ""}
	case "/api/alby/me":
		me, err := app.svc.GetAlbyOAuthSvc().GetMe(ctx)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: me, Error: ""}
	case "/api/alby/balance":
		balance, err := app.svc.GetAlbyOAuthSvc().GetBalance(ctx)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: &alby.AlbyBalanceResponse{
			Sats: balance.Balance,
		}, Error: ""}
	case "/api/alby/unlink-account":
		err := app.svc.GetAlbyOAuthSvc().UnlinkAccount(ctx)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: nil, Error: ""}
	case "/api/alby/pay":
		payRequest := &alby.AlbyPayRequest{}
		err := json.Unmarshal([]byte(body), payRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		err = app.svc.GetAlbyOAuthSvc().SendPayment(ctx, payRequest.Invoice)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: nil, Error: ""}
	case "/api/alby/rates":
		rate, err := app.svc.GetAlbyOAuthSvc().GetBitcoinRate(ctx)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to get Bitcoin rate")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: rate, Error: ""}
	case "/api/apps":
		switch method {
		case "POST":
			createAppRequest := &api.CreateAppRequest{}
			err := json.Unmarshal([]byte(body), createAppRequest)
			if err != nil {
				logger.Logger.WithFields(logrus.Fields{
					"route":  route,
					"method": method,
					"body":   body,
				}).WithError(err).Error("Failed to decode request to wails router")
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}

			createAppResponse, err := app.api.CreateApp(createAppRequest)
			if err != nil {
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			return WailsRequestRouterResponse{Body: createAppResponse, Error: ""}
		}
	case "/api/reset-router":
		resetRouterRequest := &api.ResetRouterRequest{}
		err := json.Unmarshal([]byte(body), resetRouterRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}

		err = app.api.ResetRouter(resetRouterRequest.Key)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		res := WailsRequestRouterResponse{Body: nil, Error: ""}
		return res
	case "/api/stop":
		err := app.api.Stop()
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		res := WailsRequestRouterResponse{Body: nil, Error: ""}
		return res
	case "/api/channels":
		switch method {
		case "GET":
			channels, err := app.api.ListChannels(ctx)
			if err != nil {
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			res := WailsRequestRouterResponse{Body: channels, Error: ""}
			return res
		case "POST":
			openChannelRequest := &api.OpenChannelRequest{}
			err := json.Unmarshal([]byte(body), openChannelRequest)
			if err != nil {
				logger.Logger.WithFields(logrus.Fields{
					"route":  route,
					"method": method,
					"body":   body,
				}).WithError(err).Error("Failed to decode request to wails router")
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			openChannelResponse, err := app.api.OpenChannel(ctx, openChannelRequest)
			if err != nil {
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			return WailsRequestRouterResponse{Body: openChannelResponse, Error: ""}
		}
	case "/api/channels/suggestions":
		suggestions, err := app.api.GetChannelPeerSuggestions(ctx)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		res := WailsRequestRouterResponse{Body: suggestions, Error: ""}
		return res
	case "/api/channels/rebalance":
		rebalanceChannelRequest := &api.RebalanceChannelRequest{}
		err := json.Unmarshal([]byte(body), rebalanceChannelRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		rebalanceChannelResponse, err := app.api.RebalanceChannel(ctx, rebalanceChannelRequest)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: rebalanceChannelResponse, Error: ""}
	case "/api/balances":
		balancesResponse, err := app.api.GetBalances(ctx)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		res := WailsRequestRouterResponse{Body: *balancesResponse, Error: ""}
		return res
	case "/api/invoices":
		makeInvoiceRequest := &api.MakeInvoiceRequest{}
		err := json.Unmarshal([]byte(body), makeInvoiceRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		invoice, err := app.api.CreateInvoice(ctx, makeInvoiceRequest.Amount, makeInvoiceRequest.Description)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		res := WailsRequestRouterResponse{Body: invoice, Error: ""}
		return res
	case "/api/wallet/sync":
		app.api.SyncWallet()
		return WailsRequestRouterResponse{Body: nil, Error: ""}
	case "/api/wallet/address":
		address, err := app.api.GetUnusedOnchainAddress(ctx)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: address, Error: ""}
	case "/api/wallet/new-address":
		newAddress, err := app.api.GetNewOnchainAddress(ctx)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: newAddress, Error: ""}
	case "/api/wallet/redeem-onchain-funds":
		redeemOnchainFundsRequest := &api.RedeemOnchainFundsRequest{}
		err := json.Unmarshal([]byte(body), redeemOnchainFundsRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}

		redeemOnchainFundsResponse, err := app.api.RedeemOnchainFunds(ctx, redeemOnchainFundsRequest.ToAddress, redeemOnchainFundsRequest.Amount, redeemOnchainFundsRequest.FeeRate, redeemOnchainFundsRequest.SendAll)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: *redeemOnchainFundsResponse, Error: ""}
	case "/api/wallet/sign-message":
		signMessageRequest := &api.SignMessageRequest{}
		err := json.Unmarshal([]byte(body), signMessageRequest)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		signMessageResponse, err := app.api.SignMessage(ctx, signMessageRequest.Message)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: *signMessageResponse, Error: ""}
	case "/api/wallet/capabilities":
		capabilitiesResponse, err := app.api.GetWalletCapabilities(ctx)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: *capabilitiesResponse, Error: ""}
	case "/api/lsp-orders":
		newInstantChannelRequest := &api.LSPOrderRequest{}
		err := json.Unmarshal([]byte(body), newInstantChannelRequest)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		newInstantChannelResponse, err := app.api.RequestLSPOrder(ctx, newInstantChannelRequest)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: *newInstantChannelResponse, Error: ""}
	case "/api/peers":
		switch method {
		case "GET":
			peers, err := app.api.ListPeers(ctx)
			if err != nil {
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			return WailsRequestRouterResponse{Body: peers, Error: ""}
		case "POST":
			connectPeerRequest := &api.ConnectPeerRequest{}
			err := json.Unmarshal([]byte(body), connectPeerRequest)
			if err != nil {
				logger.Logger.WithFields(logrus.Fields{
					"route":  route,
					"method": method,
					"body":   body,
				}).WithError(err).Error("Failed to decode request to wails router")
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			err = app.api.ConnectPeer(ctx, connectPeerRequest)
			if err != nil {
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			return WailsRequestRouterResponse{Body: nil, Error: ""}
		}
	case "/api/node/connection-info":
		nodeConnectionInfo, err := app.api.GetNodeConnectionInfo(ctx)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: *nodeConnectionInfo, Error: ""}
	case "/api/node/status":
		nodeStatus, err := app.api.GetNodeStatus(ctx)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		if nodeStatus == nil {
			return WailsRequestRouterResponse{Body: nil, Error: ""}
		}
		return WailsRequestRouterResponse{Body: *nodeStatus, Error: ""}
	case "/api/node/transactions":
		transactions, err := app.api.ListOnchainTransactions(ctx)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: transactions, Error: ""}
	case "/api/info":
		infoResponse, err := app.api.GetInfo(ctx)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		infoResponse.Unlocked = infoResponse.Running
		res := WailsRequestRouterResponse{Body: *infoResponse, Error: ""}
		return res
	case "/api/node/migrate-storage":
		migrateNodeStorageRequest := &api.MigrateNodeStorageRequest{}
		err := json.Unmarshal([]byte(body), migrateNodeStorageRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		err = app.api.MigrateNodeStorage(ctx, migrateNodeStorageRequest.To)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: nil, Error: ""}
	case "/api/alby/auto-channel":
		newAutoChannelRequest := &alby.AutoChannelRequest{}
		err := json.Unmarshal([]byte(body), newAutoChannelRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		autoChannelResponse, err := app.svc.GetAlbyOAuthSvc().RequestAutoChannel(ctx, app.svc.GetLNClient(), newAutoChannelRequest.IsPublic)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: *autoChannelResponse, Error: ""}

	case "/api/alby/link-account":
		linkAccountRequest := &alby.AlbyLinkAccountRequest{}
		err := json.Unmarshal([]byte(body), linkAccountRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		err = app.svc.GetAlbyOAuthSvc().LinkAccount(ctx, app.svc.GetLNClient(), linkAccountRequest.Budget, linkAccountRequest.Renewal)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		res := WailsRequestRouterResponse{Error: ""}
		return res
	case "/api/mnemonic":
		mnemonicRequest := &api.MnemonicRequest{}
		err := json.Unmarshal([]byte(body), mnemonicRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				// Skip logging the body for this request as we don't want the
				// unlock password to end up in any logs
				// "body": body,
			}).WithError(err).Error("Failed to parse mnemonic request")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		mnemonicResponse, err := app.api.GetMnemonic(mnemonicRequest.UnlockPassword)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				// Skip logging the body for this request as we don't want the
				// unlock password to end up in any logs
				// "body": body,
			}).WithError(err).Error("Failed to get mnemonic")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		res := WailsRequestRouterResponse{Body: *mnemonicResponse, Error: ""}
		return res
	case "/api/backup-reminder":
		backupReminderRequest := &api.BackupReminderRequest{}
		err := json.Unmarshal([]byte(body), backupReminderRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}

		err = app.api.SetNextBackupReminder(backupReminderRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to store backup reminder")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: nil, Error: ""}
	case "/api/unlock-password":
		changeUnlockPasswordRequest := &api.ChangeUnlockPasswordRequest{}
		err := json.Unmarshal([]byte(body), changeUnlockPasswordRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}

		err = app.api.ChangeUnlockPassword(changeUnlockPasswordRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to change unlock password")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: nil, Error: ""}
	case "/api/settings":
		updateSettingsRequest := &api.UpdateSettingsRequest{}
		err := json.Unmarshal([]byte(body), updateSettingsRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}

		err = app.api.SetCurrency(updateSettingsRequest.Currency)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to set Currency")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: nil, Error: ""}
	case "/api/auto-unlock":
		autoUnlockRequest := &api.AutoUnlockRequest{}
		err := json.Unmarshal([]byte(body), autoUnlockRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}

		err = app.api.SetAutoUnlockPassword(autoUnlockRequest.UnlockPassword)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to set auto unlock password")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: nil, Error: ""}
	case "/api/start":
		startRequest := &api.StartRequest{}
		err := json.Unmarshal([]byte(body), startRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}

		go app.api.Start(startRequest)

		return WailsRequestRouterResponse{Body: nil, Error: ""}
	case "/api/setup":
		setupRequest := &api.SetupRequest{}
		err := json.Unmarshal([]byte(body), setupRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		err = app.api.Setup(ctx, setupRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to setup node")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: nil, Error: ""}
	case "/api/send-payment-probes":
		sendPaymentProbesRequest := &api.SendPaymentProbesRequest{}
		err := json.Unmarshal([]byte(body), sendPaymentProbesRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		sendPaymentProbesResponse, err := app.api.SendPaymentProbes(ctx, sendPaymentProbesRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to send payment probes")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: sendPaymentProbesResponse, Error: ""}
	case "/api/send-spontaneous-payment-probes":
		sendSpontaneousPaymentProbesRequest := &api.SendSpontaneousPaymentProbesRequest{}
		err := json.Unmarshal([]byte(body), sendSpontaneousPaymentProbesRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		sendSpontaneousPaymentProbesResponse, err := app.api.SendSpontaneousPaymentProbes(ctx, sendSpontaneousPaymentProbesRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to send spontaneous payment probes")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: sendSpontaneousPaymentProbesResponse, Error: ""}
	case "/api/backup":
		backupRequest := &api.BasicBackupRequest{}
		err := json.Unmarshal([]byte(body), backupRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}

		saveFilePath, err := runtime.SaveFileDialog(ctx, runtime.SaveDialogOptions{
			Title:           "Save Backup File",
			DefaultFilename: "albyhub.bkp",
		})
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to open save file dialog")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}

		backupFile, err := os.Create(saveFilePath)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to create backup file")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}

		defer backupFile.Close()

		err = app.api.CreateBackup(backupRequest.UnlockPassword, backupFile)

		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to create backup")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: nil, Error: ""}
	case "/api/restore":
		restoreRequest := &api.BasicRestoreWailsRequest{}
		err := json.Unmarshal([]byte(body), restoreRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}

		backupFilePath, err := runtime.OpenFileDialog(ctx, runtime.OpenDialogOptions{
			Title:           "Select Backup File",
			DefaultFilename: "albyhub.bkp",
		})
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to open save file dialog")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}

		backupFile, err := os.Open(backupFilePath)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to open backup file")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}

		defer backupFile.Close()

		err = app.api.RestoreBackup(restoreRequest.UnlockPassword, backupFile)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to restore backup")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: nil, Error: ""}
	case "/api/health":
		nodeHealth, err := app.api.Health(ctx)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to check node health")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: *nodeHealth, Error: ""}
	case "/api/offers":
		makeOfferRequest := &api.MakeOfferRequest{}
		err := json.Unmarshal([]byte(body), makeOfferRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		offer, err := app.api.MakeOffer(ctx, makeOfferRequest.Description)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to generate BOLT-12 offer")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: offer, Error: ""}
	case "/api/commands":
		nodeCommandsResponse, err := app.api.GetCustomNodeCommands()
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to get node commands")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: nodeCommandsResponse, Error: ""}
	case "/api/command":
		commandRequest := &api.ExecuteCustomNodeCommandRequest{}
		err := json.Unmarshal([]byte(body), commandRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		commandResponse, err := app.api.ExecuteCustomNodeCommand(ctx, commandRequest.Command)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to execute command")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: commandResponse, Error: ""}
	case "/api/autoswap":
		switch method {
		case "GET":
			autoSwapsConfig, err := app.api.GetAutoSwapConfig()
			if err != nil {
				logger.Logger.WithFields(logrus.Fields{
					"route":  route,
					"method": method,
					"body":   body,
				}).WithError(err).Error("Failed to get auto swap configuration")
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			return WailsRequestRouterResponse{Body: autoSwapsConfig, Error: ""}
		case "POST":
			enableAutoSwapRequest := &api.EnableAutoSwapRequest{}
			err := json.Unmarshal([]byte(body), enableAutoSwapRequest)
			if err != nil {
				logger.Logger.WithFields(logrus.Fields{
					"route":  route,
					"method": method,
					"body":   body,
				}).WithError(err).Error("Failed to decode request to wails router")
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			err = app.api.EnableAutoSwapOut(ctx, enableAutoSwapRequest)
			if err != nil {
				logger.Logger.WithFields(logrus.Fields{
					"route":  route,
					"method": method,
					"body":   body,
				}).WithError(err).Error("Failed to enable auto swap")
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			return WailsRequestRouterResponse{Body: nil, Error: ""}
		case "DELETE":
			err := app.api.DisableAutoSwap()
			if err != nil {
				logger.Logger.WithFields(logrus.Fields{
					"route":  route,
					"method": method,
					"body":   body,
				}).WithError(err).Error("Failed to disable auto swap")
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			return WailsRequestRouterResponse{Body: nil, Error: ""}
		}
	case "/api/swaps/out/fees":
		swapOutFees, err := app.api.GetSwapOutFees()
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to get swap out fees")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: swapOutFees, Error: ""}
	case "/api/swaps/in/fees":
		swapInFees, err := app.api.GetSwapInFees()
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to get swap in fees")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: swapInFees, Error: ""}
	case "/api/swaps/out":
		initiateSwapOutRequest := &api.InitiateSwapRequest{}
		err := json.Unmarshal([]byte(body), initiateSwapOutRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		swapOutResponse, err := app.api.InitiateSwapOut(ctx, initiateSwapOutRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to initiate swap out")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: swapOutResponse, Error: ""}
	case "/api/swaps/in":
		initiateSwapInRequest := &api.InitiateSwapRequest{}
		err := json.Unmarshal([]byte(body), initiateSwapInRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		swapInResponse, err := app.api.InitiateSwapIn(ctx, initiateSwapInRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to initiate swap in")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: swapInResponse, Error: ""}
	case "/api/swaps/refund":
		refundSwapRequest := &api.RefundSwapRequest{}
		err := json.Unmarshal([]byte(body), refundSwapRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		err = app.api.RefundSwap(refundSwapRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to process swap refund")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: nil, Error: ""}
	case "/api/swaps/mnemonic":
		mnemonic := app.api.GetSwapMnemonic()
		return WailsRequestRouterResponse{Body: mnemonic, Error: ""}
	case "/api/node/alias":
		setNodeAliasRequest := &api.SetNodeAliasRequest{}
		err := json.Unmarshal([]byte(body), setNodeAliasRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}

		err = app.api.SetNodeAlias(setNodeAliasRequest.NodeAlias)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to set node alias")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: nil, Error: ""}
	case "/api/event":
		switch method {
		case "POST":
			sendEventRequest := &api.SendEventRequest{}
			err := json.Unmarshal([]byte(body), sendEventRequest)
			if err != nil {
				logger.Logger.WithFields(logrus.Fields{
					"route":  route,
					"method": method,
					"body":   body,
				}).WithError(err).Error("Failed to send event")
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}

			app.api.SendEvent(sendEventRequest.Event)

			return WailsRequestRouterResponse{Body: nil, Error: ""}
		}
	case "/api/lightning-addresses":
		switch method {
		case "POST":
			createLightningAddressRequest := &api.CreateLightningAddressRequest{}
			err := json.Unmarshal([]byte(body), createLightningAddressRequest)
			if err != nil {
				logger.Logger.WithFields(logrus.Fields{
					"route":  route,
					"method": method,
					"body":   body,
				}).WithError(err).Error("Failed to decode request to wails router")
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}

			err = app.api.CreateLightningAddress(ctx, createLightningAddressRequest)
			if err != nil {
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			return WailsRequestRouterResponse{Body: nil, Error: ""}
		}
	}

	lightningAddressRegex := regexp.MustCompile(
		`/api/lightning-addresses/([^/]+)`,
	)
	lightningAddressMatch := lightningAddressRegex.FindStringSubmatch(route)

	switch {
	case len(lightningAddressMatch) == 2:
		appIdStr := lightningAddressMatch[1]
		appId, err := strconv.ParseUint(appIdStr, 10, 64)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: "Invalid app ID"}
		}

		switch method {
		case "DELETE":
			err := app.api.DeleteLightningAddress(ctx, uint(appId))
			if err != nil {
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			return WailsRequestRouterResponse{Body: nil, Error: ""}
		}
	}

	// Swap lookup and listing is shifted to the bottom so it
	// doesn't interfere with other swap endpoints
	swapRegex := regexp.MustCompile(
		`/api/swaps/([0-9a-zA-Z]+)`,
	)
	swapIdMatch := swapRegex.FindStringSubmatch(route)

	switch {
	case len(swapIdMatch) > 1:
		swapId := swapIdMatch[1]
		swapInfo, err := app.api.LookupSwap(swapId)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}

		return WailsRequestRouterResponse{Body: swapInfo, Error: ""}
	}

	listSwapsRegex := regexp.MustCompile(
		`/api/swaps`,
	)

	switch {
	case listSwapsRegex.MatchString(route):
		swaps, err := app.api.ListSwaps()
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: swaps, Error: ""}
	}

	if strings.HasPrefix(route, "/api/log/") {
		logType := strings.TrimPrefix(route, "/api/log/")
		logType = strings.Split(logType, "?")[0]
		if logType != api.LogTypeNode && logType != api.LogTypeApp {
			return WailsRequestRouterResponse{Body: nil, Error: fmt.Sprintf("Invalid log type: '%s'", logType)}
		}
		parsedUrl, err := url.Parse(route)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: "Failed to parse route URL"}
		}
		queryParams := parsedUrl.Query()
		getLogOutputRequest := &api.GetLogOutputRequest{}
		if maxLen := queryParams.Get("maxLen"); maxLen != "" {
			getLogOutputRequest.MaxLen, err = strconv.Atoi(maxLen)
			if err != nil {
				return WailsRequestRouterResponse{Body: nil, Error: "Invalid max length parameter"}
			}
		}
		logOutputResponse, err := app.api.GetLogOutput(ctx, logType, getLogOutputRequest)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to get log output")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: logOutputResponse, Error: ""}
	}

	logger.Logger.WithFields(logrus.Fields{
		"route":  route,
		"method": method,
	}).Error("Unhandled route")
	return WailsRequestRouterResponse{Body: nil, Error: fmt.Sprintf("Unhandled route: %s %s", method, route)}
}
