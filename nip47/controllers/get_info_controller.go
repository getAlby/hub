package controllers

import (
	"context"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

type getInfoResponse struct {
	Alias         string   `json:"alias"`
	Color         string   `json:"color"`
	Pubkey        string   `json:"pubkey"`
	Network       string   `json:"network"`
	BlockHeight   uint32   `json:"block_height"`
	BlockHash     string   `json:"block_hash"`
	Methods       []string `json:"methods"`
	Notifications []string `json:"notifications"`
}

func (controller *nip47Controller) HandleGetInfoEvent(ctx context.Context, nip47Request *models.Request, requestEventId uint, app *db.App, checkPermission checkPermissionFunc, publishResponse publishFunc) {
	supportedNotifications := []string{}
	if controller.permissionsService.PermitsNotifications(app) {
		supportedNotifications = controller.lnClient.GetSupportedNIP47NotificationTypes()
	}

	responsePayload := &getInfoResponse{
		Methods:       controller.permissionsService.GetPermittedMethods(app, controller.lnClient),
		Notifications: supportedNotifications,
	}

	// basic permissions check
	hasPermission, _, _ := controller.permissionsService.HasPermission(app, constants.GET_INFO_SCOPE, 0)
	if hasPermission {
		logger.Logger.WithFields(logrus.Fields{
			"request_event_id": requestEventId,
		}).Info("Getting info")

		info, err := controller.lnClient.GetInfo(ctx)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"request_event_id": requestEventId,
			}).Infof("Failed to fetch node info: %v", err)

			publishResponse(&models.Response{
				ResultType: nip47Request.Method,
				Error: &models.Error{
					Code:    models.ERROR_INTERNAL,
					Message: err.Error(),
				},
			}, nostr.Tags{})
			return
		}

		network := info.Network
		// Some implementations return "bitcoin" while NIP47 expects "mainnet"
		if network == "bitcoin" {
			network = "mainnet"
		}

		responsePayload.Alias = info.Alias
		responsePayload.Color = info.Color
		responsePayload.Pubkey = info.Pubkey
		responsePayload.Network = network
		responsePayload.BlockHeight = info.BlockHeight
		responsePayload.BlockHash = info.BlockHash
	}

	publishResponse(&models.Response{
		ResultType: nip47Request.Method,
		Result:     responsePayload,
	}, nostr.Tags{})
}
