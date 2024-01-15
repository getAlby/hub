package main

import (
	"context"
	"embed"
	"log"

	"github.com/getAlby/nostr-wallet-connect/models/api"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

type WailsApp struct {
	ctx context.Context
	svc *Service
}

func NewApp(svc *Service) *WailsApp {
	return &WailsApp{svc: svc}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *WailsApp) startup(ctx context.Context) {
	a.ctx = ctx
}

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
		// TODO: move to API
		return api.InfoResponse{
			BackendType: a.svc.cfg.LNBackendType,
		}
	case "/api/user/me":
		// no user in this mode
		return nil
	case "/api/csrf":
		// never used
		return ""
	}
	a.svc.Logger.Fatalf("Unhandled route: %s", route)
	return nil
}

func LaunchWailsApp(app *WailsApp) {
	err := wails.Run(&options.App{
		Title:  "Nostr Wallet Connect",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		//BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatalf("Error %v", err)
	}
}
