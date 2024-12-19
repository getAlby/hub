package controllers

import (
	"context"
	"slices"
	"time"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/nip47/permissions"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

type createConnectionBudgetParams struct {
	Budget        uint64 `json:"budget"`
	RenewalPeriod string `json:"renewal_period"`
}

type createConnectionParams struct {
	Pubkey    string                       `json:"pubkey"` // pubkey of the app connection
	Name      string                       `json:"name"`
	Methods   []string                     `json:"methods"`
	Budget    createConnectionBudgetParams `json:"budget"`
	ExpiresAt *uint64                      `json:"expires_at"` // unix timestamp
	Isolated  bool                         `json:"isolated"`
	Metadata  map[string]interface{}       `json:"metadata,omitempty"`
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

	// TODO: verify the LNClient supports the methods
	supportedMethods := controller.lnClient.GetSupportedNIP47Methods()
	if slices.ContainsFunc(params.Methods, func(method string) bool {
		return !slices.Contains(supportedMethods, method)
	}) {
		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error: &models.Error{
				Code:    constants.ERROR_INTERNAL,
				Message: "One or more methods are not supported by the current LNClient",
			},
		}, nostr.Tags{})
		return
	}
	scopes, err := permissions.RequestMethodsToScopes(params.Methods)

	app, _, err := controller.appsService.CreateApp(params.Name, params.Pubkey, params.Budget.Budget, params.Budget.RenewalPeriod, expiresAt, scopes, params.Isolated, params.Metadata)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"request_event_id": requestEventId,
		}).WithError(err).Error("Failed to create app")
		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error: &models.Error{
				Code:    constants.ERROR_INTERNAL,
				Message: err.Error(),
			},
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
