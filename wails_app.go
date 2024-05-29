package main

import (
	"context"
	"embed"
	"log"

	"github.com/getAlby/nostr-wallet-connect/api"
	"github.com/sirupsen/logrus"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

type WailsApp struct {
	ctx context.Context
	svc *Service
	api api.API
}

func NewApp(svc *Service) *WailsApp {
	return &WailsApp{
		svc: svc,
		api: api.NewAPI(svc, svc.logger, svc.db),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (app *WailsApp) startup(ctx context.Context) {
	app.ctx = ctx
}

func LaunchWailsApp(app *WailsApp) {
	logger := NewWailsLogger(app.svc.logger)

	err := wails.Run(&options.App{
		Title:  "Nostr Wallet Connect",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		// HideWindowOnClose: true, // with this on, there is no way to close the app - wait for v3
		Logger: logger,

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

func NewWailsLogger(appLogger *logrus.Logger) WailsLogger {
	return WailsLogger{
		AppLogger: appLogger,
	}
}

type WailsLogger struct {
	AppLogger *logrus.Logger
}

func (logger WailsLogger) Print(message string) {
	logger.AppLogger.Print(message)
}

func (logger WailsLogger) Trace(message string) {
	logger.AppLogger.Trace(message)
}

func (logger WailsLogger) Debug(message string) {
	logger.AppLogger.Debug(message)
}

func (logger WailsLogger) Info(message string) {
	logger.AppLogger.Info(message)
}

func (logger WailsLogger) Warning(message string) {
	logger.AppLogger.Warning(message)
}

func (logger WailsLogger) Error(message string) {
	logger.AppLogger.Error(message)
}

func (logger WailsLogger) Fatal(message string) {
	logger.AppLogger.Fatal(message)
}
