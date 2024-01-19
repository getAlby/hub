//go:build wails
// +build wails

package main

import (
	"context"

	log "github.com/sirupsen/logrus"
)

// ignore this warning: we use build tags
// this function will only be executed if the wails tag is set
func main() {
	log.Info("NWC Starting in WAILS mode")
	ctx, cancel := context.WithCancel(context.Background())
	svc := NewService(ctx)

	app := NewApp(svc)
	LaunchWailsApp(app)
	svc.Logger.Info("Wails app exited")

	svc.Logger.Info("Cancelling service context...")
	// cancel the service context
	cancel()
	svc.Logger.Info("Waiting for service to exit...")
	svc.wg.Wait()
	svc.Logger.Info("Service exited")
}
