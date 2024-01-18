//go:build !wails || http
// +build !wails http

// (http tag above is simply to fix go language server issue and is not needed to build the app)

package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	echologrus "github.com/davrux/echo-logrus/v4"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

// ignore this warning: we use build tags
// this function will only be executed if no wails tag is set
func main() {
	log.Info("NWC Starting in HTTP mode")
	var wg sync.WaitGroup
	wg.Add(1)

	svc := NewService(&wg)

	if svc.cfg.CookieSecret == "" {
		svc.Logger.Fatalf("required key COOKIE_SECRET missing value")
	}

	echologrus.Logger = svc.Logger
	e := echo.New()

	//register shared routes
	svc.RegisterSharedRoutes(e)
	//start Echo server
	go func() {
		if err := e.Start(fmt.Sprintf(":%v", svc.cfg.Port)); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
		//handle graceful shutdown
		<-svc.ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		e.Shutdown(ctx)
		svc.Logger.Info("Echo server exited")
		wg.Done()
	}()

	wg.Wait()
}
