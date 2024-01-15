package main

import (
	"context"
	"embed"
	"log"

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
