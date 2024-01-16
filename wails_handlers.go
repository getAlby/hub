package main

import (
	"encoding/json"
	"fmt"

	"github.com/getAlby/nostr-wallet-connect/models/api"
	"github.com/sirupsen/logrus"
)

type WailsRequestRouterResponse struct {
	Body  interface{} `json:"body"`
	Error string      `json:"error"`
}

func (a *WailsApp) WailsRequestRouter(route string, method string, body string) WailsRequestRouterResponse {
	switch route {
	case "/api/apps":
		userApps := []App{}
		a.svc.db.Find(&userApps)
		apps := []api.App{}
		err := a.svc.ListApps(&userApps, &apps)
		if err != nil {
			return WailsRequestRouterResponse{Body: nil, Error: ""}
		}
		return WailsRequestRouterResponse{Body: apps, Error: ""}
	case "/api/info":
		infoResponse := api.InfoResponse{}
		a.svc.GetInfo(&infoResponse)
		res := WailsRequestRouterResponse{Body: infoResponse, Error: ""}
		return res
	case "/api/user/me":
		dummyUser := api.User{
			Email: "",
		}
		return WailsRequestRouterResponse{Body: dummyUser, Error: ""}
	case "/api/csrf":
		return WailsRequestRouterResponse{Body: "dummy", Error: ""}
	case "/api/setup":
		setupRequest := &api.SetupRequest{}
		err := json.Unmarshal([]byte(body), setupRequest)
		if err != nil {
			a.svc.Logger.WithFields(logrus.Fields{
				"route":  route,
				"method": method,
				"body":   body,
			}).Errorf("Failed to decode request to wails router: %v", err)
			return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
		}
		a.svc.Setup(setupRequest)
		return WailsRequestRouterResponse{Body: nil, Error: ""}
	}
	a.svc.Logger.Errorf("Unhandled route: %s", route)
	return WailsRequestRouterResponse{Body: nil, Error: fmt.Sprintf("Unhandled route: %s", route)}
}
