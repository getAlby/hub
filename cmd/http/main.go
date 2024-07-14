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
	"github.com/getAlby/hub/http"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/service"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("NWC Starting in HTTP mode")

	// Create a channel to receive OS signals.
	osSignalChannel := make(chan os.Signal, 1)
	// Notify the channel on os.Interrupt, syscall.SIGTERM, and os.Kill.
	signal.Notify(osSignalChannel, os.Interrupt, syscall.SIGTERM, os.Kill)

	ctx, cancel := context.WithCancel(context.Background())
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

	var signal os.Signal
	go func() {
		// wait for exit signal
		signal = <-osSignalChannel
		logger.Logger.WithField("signal", signal).Info("Received OS signal")
		cancel()
	}()

	//handle graceful shutdown
	<-ctx.Done()
	logger.Logger.WithField("signal", signal).Info("Context Done")
	logger.Logger.Info("Shutting down echo server...")
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	e.Shutdown(ctx)
	logger.Logger.Info("Echo server exited")
	svc.Shutdown()
	logger.Logger.Info("Service exited")
}
