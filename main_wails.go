//go:build wails
// +build wails

package main

import (
	"context"
	"embed"

	"github.com/getAlby/nostr-wallet-connect/logger"
	"github.com/getAlby/nostr-wallet-connect/service"
	"github.com/getAlby/nostr-wallet-connect/wails"
	log "github.com/sirupsen/logrus"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed appicon.png
var appIcon []byte

func main() {
	log.Info("NWC Starting in WAILS mode")
	ctx, cancel := context.WithCancel(context.Background())
	svc, _ := service.NewService(ctx)

	app := wails.NewApp(svc)
	wails.LaunchWailsApp(app, assets, appIcon)
	logger.Logger.Info("Wails app exited")

	logger.Logger.Info("Cancelling service context...")
	// cancel the service context
	cancel()
	svc.WaitShutdown()
	logger.Logger.Info("Service exited")
}
