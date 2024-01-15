package main

import "github.com/getAlby/nostr-wallet-connect/models/api"

func (a *WailsApp) WailsRequestRouter(route string) interface{} {
	switch route {
	case "/api/apps":

		userApps := []App{}
		a.svc.db.Find(&userApps)
		apps := []api.App{}
		a.svc.ListApps(&userApps, &apps)
		a.svc.Logger.Infof("END WailsRequestRouter %v", len(apps))
		return apps
	case "/api/info":
		infoResponse := api.InfoResponse{}
		a.svc.GetInfo(&infoResponse)
		return infoResponse
	case "/api/user/me":
		return nil
	case "/api/csrf":
		return "dummy"
	}
	a.svc.Logger.Fatalf("Unhandled route: %s", route)
	return nil
}
