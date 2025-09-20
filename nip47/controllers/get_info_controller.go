package controllers

import (
	"context"
	"encoding/json"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

type getInfoResponse struct {
	Alias            string      `json:"alias,omitempty"`
	Color            string      `json:"color,omitempty"`
	Pubkey           string      `json:"pubkey,omitempty"`
	Network          string      `json:"network,omitempty"`
	BlockHeight      uint32      `json:"block_height,omitempty"`
	BlockHash        string      `json:"block_hash,omitempty"`
	Methods          []string    `json:"methods"`
	Notifications    []string    `json:"notifications"`
	Metadata         interface{} `json:"metadata,omitempty"`
	LightningAddress string      `json:"lud16,omitempty"`
}

func (controller *nip47Controller) HandleGetInfoEvent(ctx context.Context, nip47Request *models.Request, requestEventId uint, app *db.App, publishResponse publishFunc) {
	supportedNotifications := []string{}
	if controller.permissionsService.PermitsNotifications(app) {
		supportedNotifications = controller.lnClient.GetSupportedNIP47NotificationTypes()
	}

	responsePayload := &getInfoResponse{
		Methods:       controller.permissionsService.GetPermittedMethods(app, controller.lnClient),
		Notifications: supportedNotifications,
	}

	// basic permissions check
	// this is inconsistent with other methods. Ideally we move fetching node info to a separate method,
	// so that get_info does not require its own scope. This would require a change in the NIP-47 spec.
	hasPermission, _, _ := controller.permissionsService.HasPermission(app, constants.GET_INFO_SCOPE)
	if hasPermission {
		logger.Logger.WithFields(logrus.Fields{
			"request_event_id": requestEventId,
		}).Debug("Getting info")

		info, err := controller.lnClient.GetInfo(ctx)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"request_event_id": requestEventId,
			}).Infof("Failed to fetch node info: %v", err)

			publishResponse(&models.Response{
				ResultType: nip47Request.Method,
				Error:      mapNip47Error(err),
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

		if app != nil {
			metadata := map[string]interface{}{}
			if app.Metadata != nil {
				jsonErr := json.Unmarshal(app.Metadata, &metadata)
				if jsonErr != nil {
					logger.Logger.WithError(jsonErr).WithFields(logrus.Fields{
						"id":       app.ID,
						"metadata": app.Metadata,
					}).Error("Failed to deserialize app metadata")
				}
			}
			if metadata["id"] == nil {
				metadata["id"] = app.ID
			}
			if metadata["name"] == nil {
				metadata["name"] = app.Name
			}
			if !app.Isolated {
				lightningAddress, _ := controller.albyOAuthService.GetLightningAddress()
				responsePayload.LightningAddress = lightningAddress
			}

			responsePayload.Metadata = metadata
		}
	}

	publishResponse(&models.Response{
		ResultType: nip47Request.Method,
		Result:     responsePayload,
	}, nostr.Tags{})
}
