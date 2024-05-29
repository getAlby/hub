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
	svc, _ := NewService(ctx)

	app := NewApp(svc)
	LaunchWailsApp(app)
	svc.logger.Info("Wails app exited")

	svc.logger.Info("Cancelling service context...")
	// cancel the service context
	cancel()
	svc.logger.Info("Waiting for service to exit...")
	svc.wg.Wait()
	svc.logger.Info("Service exited")
}
