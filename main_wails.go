//go:build wails
// +build wails

package main

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

// ignore this warning: we use build tags
// this function will only be executed if the wails tag is set
func main() {
	log.Info("NWC Starting in WAILS mode")
	var wg sync.WaitGroup
	wg.Add(1)

	svc := CreateService(&wg)

	go func() {
		app := NewApp(svc)
		LaunchWailsApp(app)
		wg.Done()
		svc.Logger.Info("Wails app exited")
	}()

	wg.Wait()
}
