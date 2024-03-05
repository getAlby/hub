package main

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/getAlby/nostr-wallet-connect/models/api"
	"github.com/sirupsen/logrus"
)

type WailsRequestRouterResponse struct {
	Body  interface{} `json:"body"`
	Error string      `json:"error"`
}

// TODO: make this match echo
func (app *WailsApp) WailsRequestRouter(route string, method string, body string) WailsRequestRouterResponse {
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
				}).Errorf("Failed to decode request to wails router: %v", err)
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			createAppResponse, err := app.api.CreateApp(createAppRequest)
			if err != nil {
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			return WailsRequestRouterResponse{Body: createAppResponse, Error: ""}
		}
	case "/api/channels":
		switch method {
		case "GET":
			channels, err := app.api.ListChannels()
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
				}).Errorf("Failed to decode request to wails router: %v", err)
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			openChannelResponse, err := app.api.OpenChannel(openChannelRequest)
			if err != nil {
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			return WailsRequestRouterResponse{Body: openChannelResponse, Error: ""}
		}
	case "/api/wallet/balance":
		balanceResponse, err := app.api.GetOnchainBalance()
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		res := WailsRequestRouterResponse{Body: *balanceResponse, Error: ""}
		return res
	case "/api/wallet/new-address":
		newAddressResponse, err := app.api.GetNewOnchainAddress()
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: *newAddressResponse, Error: ""}
	case "/api/peers":
		connectPeerRequest := &api.ConnectPeerRequest{}
		err := json.Unmarshal([]byte(body), connectPeerRequest)
		if err != nil {
			app.svc.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).Errorf("Failed to decode request to wails router: %v", err)
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		err = app.api.ConnectPeer(connectPeerRequest)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: nil, Error: ""}
	case "/api/node/connection-info":
		nodeConnectionInfo, err := app.api.GetNodeConnectionInfo()
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: *nodeConnectionInfo, Error: ""}
	case "/api/info":
		infoResponse, err := app.api.GetInfo()
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		infoResponse.Unlocked = infoResponse.Running
		res := WailsRequestRouterResponse{Body: *infoResponse, Error: ""}
		return res
	case "/api/mnemonic":
		switch method {
		case "GET":
			infoResponse := app.api.GetMnemonic()
			res := WailsRequestRouterResponse{Body: *infoResponse, Error: ""}
			return res
		case "PATCH":
			err := app.api.BackupMnemonic()
			if err != nil {
				app.svc.Logger.WithFields(logrus.Fields{
					"route":  route,
					"method": method,
					"body":   body,
				}).Errorf("Failed to backup mnemonic: %v", err)
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			return WailsRequestRouterResponse{Body: nil, Error: ""}
		}
	case "/api/start":
		startRequest := &api.StartRequest{}
		err := json.Unmarshal([]byte(body), startRequest)
		if err != nil {
			app.svc.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).Errorf("Failed to decode request to wails router: %v", err)
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		err = app.api.Start(startRequest)
		if err != nil {
			app.svc.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).Errorf("Failed to setup node: %v", err)
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
			}).Errorf("Failed to decode request to wails router: %v", err)
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		err = app.api.Setup(setupRequest)
		if err != nil {
			app.svc.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).Errorf("Failed to setup node: %v", err)
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		return WailsRequestRouterResponse{Body: nil, Error: ""}
	}
	app.svc.Logger.Errorf("Unhandled route: %s", route)
	return WailsRequestRouterResponse{Body: nil, Error: fmt.Sprintf("Unhandled route: %s %s", method, route)}
}
