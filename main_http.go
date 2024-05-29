//go:build !wails || http
// +build !wails http

// (http tag above is simply to fix go language server issue and is not needed to build the app)

package main

import (
	"context"
	"fmt"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	echologrus "github.com/davrux/echo-logrus/v4"
	"github.com/getAlby/nostr-wallet-connect/http"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

// ignore this warning: we use build tags
// this function will only be executed if no wails tag is set
func main() {
	log.Info("NWC Starting in HTTP mode")
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, os.Kill)
	svc, _ := NewService(ctx)

	echologrus.Logger = svc.logger
	e := echo.New()

	//register shared routes
	httpSvc := http.NewHttpService(svc, svc.logger, svc.db, svc.eventPublisher)
	httpSvc.RegisterSharedRoutes(e)
	//start Echo server
	go func() {
		if err := e.Start(fmt.Sprintf(":%v", svc.cfg.GetEnv().Port)); err != nil && err != nethttp.ErrServerClosed {
			svc.logger.Fatalf("shutting down the server: %v", err)
		}
	}()
	//handle graceful shutdown
	<-ctx.Done()
	svc.logger.Infof("Shutting down echo server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	e.Shutdown(ctx)
	svc.logger.Info("Echo server exited")
	svc.logger.Info("Waiting for service to exit...")
	svc.wg.Wait()
	svc.logger.Info("Service exited")
}
