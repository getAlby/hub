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
	"github.com/getAlby/nostr-wallet-connect/logger"
	"github.com/getAlby/nostr-wallet-connect/service"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("NWC Starting in HTTP mode")
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, os.Kill)
	svc, _ := service.NewService(ctx)

	echologrus.Logger = logger.Logger
	e := echo.New()

	//register shared routes
	httpSvc := http.NewHttpService(svc, svc.GetEventPublisher())
	httpSvc.RegisterSharedRoutes(e)
	//start Echo server
	go func() {
		if err := e.Start(fmt.Sprintf(":%v", svc.GetConfig().GetEnv().Port)); err != nil && err != nethttp.ErrServerClosed {
			logger.Logger.Fatalf("shutting down the server: %v", err)
		}
	}()
	//handle graceful shutdown
	<-ctx.Done()
	logger.Logger.Infof("Shutting down echo server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	e.Shutdown(ctx)
	logger.Logger.Info("Echo server exited")
	svc.WaitShutdown()
	logger.Logger.Info("Service exited")
}
