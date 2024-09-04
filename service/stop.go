package service

import (
	"fmt"

	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/logger"
)

func (svc *service) StopApp() {
	if svc.appCancelFn != nil {
		logger.Logger.Info("Stopping app...")
		svc.appCancelFn()
		svc.wg.Wait()
		logger.Logger.Info("app stopped")
	}
}

func (svc *service) stopLNClient() {
	defer svc.wg.Done()
	if svc.lnClient == nil {
		return
	}
	lnClient := svc.lnClient
	svc.lnClient = nil

	logger.Logger.Info("Shutting down LN client")
	err := lnClient.Shutdown()
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to stop LN client")
		svc.eventPublisher.Publish(&events.Event{
			Event: "nwc_node_stop_failed",
			Properties: map[string]interface{}{
				"error": fmt.Sprintf("%v", err),
			},
		})
		return
	}
	logger.Logger.Info("Publishing node shutdown event")
	svc.eventPublisher.Publish(&events.Event{
		Event: "nwc_node_stopped",
	})
	logger.Logger.Info("LNClient stopped successfully")
}
