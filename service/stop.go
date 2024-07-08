package service

import (
	"fmt"

	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/logger"
)

// TODO: this should happen on ctx.Done() rather than having to call manually
// see svc.appCancelFn and how svc.StartNostr works
func (svc *service) StopLNClient() error {
	if svc.lnClient != nil {
		logger.Logger.Info("Shutting down LN client")
		err := svc.lnClient.Shutdown()
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to stop LN client")
			svc.eventPublisher.Publish(&events.Event{
				Event: "nwc_node_stop_failed",
				Properties: map[string]interface{}{
					"error": fmt.Sprintf("%v", err),
				},
			})
			return err
		}
		logger.Logger.Info("Publishing node shutdown event")
		svc.lnClient = nil
		svc.eventPublisher.Publish(&events.Event{
			Event: "nwc_node_stopped",
		})
	}
	logger.Logger.Info("LNClient stopped successfully")
	return nil
}
