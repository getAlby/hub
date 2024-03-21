//go:build !wails || http
// +build !wails http

// (http tag above is simply to fix go language server issue and is not needed to build the app)

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	echologrus "github.com/davrux/echo-logrus/v4"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

// ignore this warning: we use build tags
// this function will only be executed if no wails tag is set
func main() {
	log.Info("NWC Starting in HTTP mode")
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	svc, _ := NewService(ctx)

	echologrus.Logger = svc.Logger
	e := echo.New()

	//register shared routes
	httpSvc := NewHttpService(svc)
	httpSvc.RegisterSharedRoutes(e)
	//start Echo server
	go func() {
		if err := e.Start(fmt.Sprintf(":%v", svc.cfg.Env.Port)); err != nil && err != http.ErrServerClosed {
			svc.Logger.Fatalf("shutting down the server: %v", err)
		}
	}()
	//handle graceful shutdown
	<-svc.ctx.Done()
	svc.Logger.Infof("Shutting down echo server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	e.Shutdown(ctx)
	svc.Logger.Info("Echo server exited")
	svc.Logger.Info("Waiting for service to exit...")
	svc.wg.Wait()
	svc.Logger.Info("Service exited")
}
