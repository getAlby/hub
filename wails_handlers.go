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
func (a *WailsApp) WailsRequestRouter(route string, method string, body string) WailsRequestRouterResponse {
	appRegex := regexp.MustCompile(
		`/api/apps/([0-9a-f]+)`,
	)

	appMatch := appRegex.FindStringSubmatch(route)

	switch {
	case appMatch != nil && len(appMatch) > 1:
		pubkey := appMatch[1]
		userApp := App{}
		findResult := a.svc.db.Where("nostr_pubkey = ?", pubkey).First(&userApp)

		if findResult.RowsAffected == 0 {
			return WailsRequestRouterResponse{Body: nil, Error: "App does not exist"}
		}

		switch method {
		case "GET":
			app := a.svc.GetApp(&userApp)
			return WailsRequestRouterResponse{Body: app, Error: ""}
		case "DELETE":
			err := a.svc.DeleteApp(&userApp)
			if err != nil {
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			return WailsRequestRouterResponse{Body: nil, Error: ""}
		}
	}

	//e.POST("/api/apps", svc.AppsCreateHandler, authMiddleware)

	switch route {
	case "/api/apps":
		switch method {
		case "GET":
			userApps := []App{}
			a.svc.db.Find(&userApps)
			apps, err := a.svc.ListApps(&userApps)
			if err != nil {
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			return WailsRequestRouterResponse{Body: apps, Error: ""}
		case "POST":
			createAppRequest := &api.CreateAppRequest{}
			err := json.Unmarshal([]byte(body), createAppRequest)
			if err != nil {
				a.svc.Logger.WithFields(logrus.Fields{
					"route":  route,
					"method": method,
					"body":   body,
				}).Errorf("Failed to decode request to wails router: %v", err)
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			user := User{}
			a.svc.db.First(&user)
			createAppResponse, err := a.svc.CreateApp(&user, createAppRequest)
			if err != nil {
				return WailsRequestRouterResponse{Body: nil, Error: err.Error()}
			}
			return WailsRequestRouterResponse{Body: createAppResponse, Error: ""}
		}
	case "/api/info":
		infoResponse := a.svc.GetInfo()
		res := WailsRequestRouterResponse{Body: *infoResponse, Error: ""}
		return res
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
	return WailsRequestRouterResponse{Body: nil, Error: fmt.Sprintf("Unhandled route: %s %s", method, route)}
}
