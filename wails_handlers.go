package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/getAlby/nostr-wallet-connect/models/api"
)

type WailsRequestRouterResponse struct {
	Body  interface{} `json:"body"`
	Error string      `json:"error"`
}

// TODO: make this match echo
func (app *WailsApp) WailsRequestRouter(route string, method string, body string) WailsRequestRouterResponse {
	ctx := app.ctx

	appRegex := regexp.MustCompile(
		`/api/apps/([0-9a-f]+)`,
	)

	appMatch := appRegex.FindStringSubmatch(route)

	switch {
	case len(appMatch) > 1:
		pubkey := appMatch[1]

		userApp := App{}
		findResult := app.svc.db.Where("nostr_pubkey = ?", pubkey).First(&userApp)

		if findResult.RowsAffected == 0 {
			return WailsRequestRouterResponse{Body: nil, Error: "App does not exist"}
		}

		switch method {
		case "GET":
			app := app.api.GetApp(&userApp)
			return WailsRequestRouterResponse{Body: app, Error: ""}
		case "PATCH":
			updateAppRequest := &api.UpdateAppRequest{}
			err := json.Unmarshal([]byte(body), updateAppRequest)
			if err != nil {
				app.svc.Logger.WithFields(logrus.Fields{
					"route":  route,
					"method": method,
					"body":   body,
				}).WithError(err).Error("Failed to decode request to wails router")
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			err = app.api.UpdateApp(&userApp, updateAppRequest)
			if err != nil {
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			return WailsRequestRouterResponse{Body: nil, Error: ""}
		case "DELETE":
			err := app.api.DeleteApp(&userApp)
			if err != nil {
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			return WailsRequestRouterResponse{Body: nil, Error: ""}
		}
	}

	mempoolLightningNodePubkeyRegex := regexp.MustCompile(
		`/api/mempool/lightning/nodes/([0-9a-f]+)`,
	)
	mempoolLightningNodePubkeyMatch := mempoolLightningNodePubkeyRegex.FindStringSubmatch(route)

	switch {
	case len(mempoolLightningNodePubkeyMatch) > 1:
		pubkey := mempoolLightningNodePubkeyMatch[1]
		node, err := app.api.GetMempoolLightningNode(pubkey)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}

		return WailsRequestRouterResponse{Body: node, Error: ""}
	}

	switch route {
	case "/api/apps":
		switch method {
		case "GET":
			apps, err := app.api.ListApps()
			if err != nil {
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			return WailsRequestRouterResponse{Body: apps, Error: ""}
		case "POST":
			createAppRequest := &api.CreateAppRequest{}
			err := json.Unmarshal([]byte(body), createAppRequest)
			if err != nil {
				app.svc.Logger.WithFields(logrus.Fields{
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
	// TODO: should this be DELETE /api/channels/:id?
	case "/api/channels/close":
		closeChannelRequest := &api.CloseChannelRequest{}
		err := json.Unmarshal([]byte(body), closeChannelRequest)
		if err != nil {
			app.svc.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		closeChannelResponse, err := app.api.CloseChannel(ctx, closeChannelRequest)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: closeChannelResponse, Error: ""}
	case "/api/reset-router":
		err := app.api.ResetRouter(ctx)
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
				app.svc.Logger.WithFields(logrus.Fields{
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
	case "/api/balances":
		balancesResponse, err := app.api.GetBalances(ctx)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		res := WailsRequestRouterResponse{Body: *balancesResponse, Error: ""}
		return res

	case "/api/wallet/new-address":
		newAddressResponse, err := app.api.GetNewOnchainAddress(ctx)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: *newAddressResponse, Error: ""}
	case "/api/wallet/redeem-onchain-funds":

		redeemOnchainFundsRequest := &api.RedeemOnchainFundsRequest{}
		err := json.Unmarshal([]byte(body), redeemOnchainFundsRequest)
		if err != nil {
			app.svc.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}

		redeemOnchainFundsResponse, err := app.api.RedeemOnchainFunds(ctx, redeemOnchainFundsRequest.ToAddress)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: *redeemOnchainFundsResponse, Error: ""}
	case "/api/peers":
		connectPeerRequest := &api.ConnectPeerRequest{}
		err := json.Unmarshal([]byte(body), connectPeerRequest)
		if err != nil {
			app.svc.Logger.WithFields(logrus.Fields{
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
	case "/api/node/connection-info":
		nodeConnectionInfo, err := app.api.GetNodeConnectionInfo(ctx)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: *nodeConnectionInfo, Error: ""}
	case "/api/info":
		infoResponse, err := app.api.GetInfo(ctx)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		infoResponse.Unlocked = infoResponse.Running
		res := WailsRequestRouterResponse{Body: *infoResponse, Error: ""}
		return res
	case "/api/encrypted-mnemonic":
		infoResponse := app.api.GetEncryptedMnemonic()
		res := WailsRequestRouterResponse{Body: *infoResponse, Error: ""}
		return res
	case "/api/backup-reminder":
		backupReminderRequest := &api.BackupReminderRequest{}
		err := json.Unmarshal([]byte(body), backupReminderRequest)
		if err != nil {
			app.svc.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}

		err = app.api.SetNextBackupReminder(backupReminderRequest)
		if err != nil {
			app.svc.Logger.WithFields(logrus.Fields{
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
			app.svc.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}

		err = app.api.ChangeUnlockPassword(changeUnlockPasswordRequest)
		if err != nil {
			app.svc.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to change unlock password")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: nil, Error: ""}
	case "/api/start":
		startRequest := &api.StartRequest{}
		err := json.Unmarshal([]byte(body), startRequest)
		if err != nil {
			app.svc.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		err = app.api.Start(startRequest)
		if err != nil {
			app.svc.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to setup node")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: nil, Error: ""}
	case "/api/csrf":
		return WailsRequestRouterResponse{Body: "dummy", Error: ""}
	case "/api/setup":
		setupRequest := &api.SetupRequest{}
		err := json.Unmarshal([]byte(body), setupRequest)
		if err != nil {
			app.svc.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		err = app.api.Setup(ctx, setupRequest)
		if err != nil {
			app.svc.Logger.WithFields(logrus.Fields{
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
			app.svc.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		sendPaymentProbesResponse, err := app.api.SendPaymentProbes(ctx, sendPaymentProbesRequest)
		if err != nil {
			app.svc.Logger.WithFields(logrus.Fields{
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
			app.svc.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		sendSpontaneousPaymentProbesResponse, err := app.api.SendSpontaneousPaymentProbes(ctx, sendSpontaneousPaymentProbesRequest)
		if err != nil {
			app.svc.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to send spontaneous payment probes")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: sendSpontaneousPaymentProbesResponse, Error: ""}
	}

	if strings.HasPrefix(route, "/api/log/") {
		logType := strings.TrimPrefix(route, "/api/log/")
		if logType != api.LogTypeNode && logType != api.LogTypeApp {
			return WailsRequestRouterResponse{Body: nil, Error: fmt.Sprintf("Invalid log type: '%s'", logType)}
		}
		getLogOutputRequest := &api.GetLogOutputRequest{}
		err := json.Unmarshal([]byte(body), getLogOutputRequest)
		if err != nil {
			app.svc.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to decode request to wails router")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		logOutputResponse, err := app.api.GetLogOutput(ctx, logType, getLogOutputRequest)
		if err != nil {
			app.svc.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).WithError(err).Error("Failed to get log output")
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: logOutputResponse, Error: ""}
	}

	app.svc.Logger.WithFields(logrus.Fields{
		"route":  route,
		"method": method,
	}).Error("Unhandled route")
	return WailsRequestRouterResponse{Body: nil, Error: fmt.Sprintf("Unhandled route: %s %s", method, route)}
}
