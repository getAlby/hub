package main

import (
	"context"
	"fmt"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getAlby/hub/http"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/service"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("AlbyHub Starting in HTTP mode")

	// Create a channel to receive OS signals.
	osSignalChannel := make(chan os.Signal, 1)
	// Notify the channel on os.Interrupt, syscall.SIGTERM. os.Kill cannot be caught.
	signal.Notify(osSignalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGPIPE)

	ctx, cancel := context.WithCancel(context.Background())

	var signal os.Signal
	go func() {
		for {
			// wait for exit signal
			signal = <-osSignalChannel
			logger.Logger.WithField("signal", signal).Info("Received OS signal")

			if signal == syscall.SIGPIPE {
				logger.Logger.WithField("signal", signal).Warn("Ignoring SIGPIPE signal")
				continue
			}

			cancel()
			break
		}
	}()

	svc, err := service.NewService(ctx)
	if err != nil {
		log.WithError(err).Fatal("Failed to create service")
		return
	}

	e := echo.New()

	//register shared routes
	httpSvc := http.NewHttpService(svc, svc.GetEventPublisher())
	httpSvc.RegisterSharedRoutes(e)
	//start Echo server
	go func() {
		if err := e.Start(fmt.Sprintf(":%v", svc.GetConfig().GetEnv().Port)); err != nil && err != nethttp.ErrServerClosed {
			logger.Logger.WithError(err).Error("echo server failed to start")
			cancel()
		}
	}()

	//handle graceful shutdown
	<-ctx.Done()
	logger.Logger.WithField("signal", signal).Info("Context Done")
	logger.Logger.Info("Shutting down echo server...")
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = e.Shutdown(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to shutdown echo server")
	}
	logger.Logger.Info("Echo server exited")
	svc.Shutdown()
	logger.Logger.Info("Service exited")
	logger.Logger.Info("Alby Hub needs to stay online to send and receive transactions. Channels may be closed if your hub stays offline for an extended period of time.")
}
