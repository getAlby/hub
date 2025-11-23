package controllers

import (
	"context"
	"slices"
	"time"

	"github.com/getAlby/hub/alby"
	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/nip47/permissions"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

type createConnectionParams struct {
	Pubkey            string                 `json:"pubkey"` // pubkey of the app connection
	Name              string                 `json:"name"`
	RequestMethods    []string               `json:"request_methods"`
	NotificationTypes []string               `json:"notification_types"`
	MaxAmount         uint64                 `json:"max_amount"`
	BudgetRenewal     string                 `json:"budget_renewal"`
	ExpiresAt         *uint64                `json:"expires_at"` // unix timestamp
	Isolated          bool                   `json:"isolated"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

type createConnectionResponse struct {
	// pubkey is given, user requesting already knows relay.
	WalletPubkey string `json:"wallet_pubkey"`
}

func (controller *nip47Controller) HandleCreateConnectionEvent(ctx context.Context, nip47Request *models.Request, requestEventId uint, publishResponse publishFunc) {
	params := &createConnectionParams{}
	resp := decodeRequest(nip47Request, params)
	if resp != nil {
		publishResponse(resp, nostr.Tags{})
		return
	}

	logger.Logger.WithFields(logrus.Fields{
		"request_event_id": requestEventId,
		"params":           params,
	}).Info("creating app")

	var expiresAt *time.Time
	if params.ExpiresAt != nil {
		expiresAtUnsigned := *params.ExpiresAt
		expiresAtValue := time.Unix(int64(expiresAtUnsigned), 0)
		expiresAt = &expiresAtValue
	}

	maxAmountSat := params.MaxAmount / 1000

	if params.Name == alby.ALBY_ACCOUNT_APP_NAME {
		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error: &models.Error{
				Code:    constants.ERROR_BAD_REQUEST,
				Message: "cannot create a new app that has reserved name: " + alby.ALBY_ACCOUNT_APP_NAME,
			},
		}, nostr.Tags{})
		return
	}

	// explicitly do not allow creating an app with create_connection permission
	if slices.Contains(params.RequestMethods, models.CREATE_CONNECTION_METHOD) {
		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error: &models.Error{
				Code:    constants.ERROR_BAD_REQUEST,
				Message: "cannot create a new app that has create_connection permission via NWC",
			},
		}, nostr.Tags{})
		return
	}

	// ensure there is at least one request method
	if len(params.RequestMethods) == 0 {
		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error: &models.Error{
				Code:    constants.ERROR_BAD_REQUEST,
				Message: "No request methods provided",
			},
		}, nostr.Tags{})
		return
	}

	supportedMethods := controller.lnClient.GetSupportedNIP47Methods()
	if slices.ContainsFunc(params.RequestMethods, func(method string) bool {
		return !slices.Contains(supportedMethods, method)
	}) {
		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error: &models.Error{
				Code:    constants.ERROR_BAD_REQUEST,
				Message: "One or more methods are not supported by the current LNClient",
			},
		}, nostr.Tags{})
		return
	}

	scopes, err := permissions.RequestMethodsToScopes(params.RequestMethods)

	supportedNotificationTypes := controller.lnClient.GetSupportedNIP47NotificationTypes()
	if len(params.NotificationTypes) > 0 {
		if slices.ContainsFunc(params.NotificationTypes, func(method string) bool {
			return !slices.Contains(supportedNotificationTypes, method)
		}) {
			publishResponse(&models.Response{
				ResultType: nip47Request.Method,
				Error: &models.Error{
					Code:    constants.ERROR_BAD_REQUEST,
					Message: "One or more notification types are not supported by the current LNClient",
				},
			}, nostr.Tags{})
			return
		}
		scopes = append(scopes, constants.NOTIFICATIONS_SCOPE)
	}

	app, _, err := controller.appsService.CreateApp(params.Name, params.Pubkey, maxAmountSat, params.BudgetRenewal, expiresAt, scopes, params.Isolated, params.Metadata)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"request_event_id": requestEventId,
		}).WithError(err).Error("Failed to create app")
		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error:      mapNip47Error(err),
		}, nostr.Tags{})
		return
	}

	responsePayload := createConnectionResponse{
		WalletPubkey: *app.WalletPubkey,
	}

	publishResponse(&models.Response{
		ResultType: nip47Request.Method,
		Result:     responsePayload,
	}, nostr.Tags{})
}
