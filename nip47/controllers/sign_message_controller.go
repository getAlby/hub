package controllers

import (
	"context"

	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

type signMessageParams struct {
	Message string `json:"message"`
}

type signMessageResponse struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

type signMessageController struct {
	lnClient lnclient.LNClient
}

func NewSignMessageController(lnClient lnclient.LNClient) *signMessageController {
	return &signMessageController{
		lnClient: lnClient,
	}
}

func (controller *signMessageController) HandleSignMessageEvent(ctx context.Context, nip47Request *models.Request, requestEventId uint, checkPermission checkPermissionFunc, publishResponse publishFunc) {
	// basic permissions check
	resp := checkPermission(0)
	if resp != nil {
		publishResponse(resp, nostr.Tags{})
		return
	}

	signParams := &signMessageParams{}
	resp = decodeRequest(nip47Request, signParams)
	if resp != nil {
		publishResponse(resp, nostr.Tags{})
		return
	}

	logger.Logger.WithFields(logrus.Fields{
		"request_event_id": requestEventId,
	}).Info("Signing message")

	signature, err := controller.lnClient.SignMessage(ctx, signParams.Message)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"request_event_id": requestEventId,
		}).WithError(err).Error("Failed to sign message")
		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error: &models.Error{
				Code:    models.ERROR_INTERNAL,
				Message: err.Error(),
			},
		}, nostr.Tags{})
		return
	}

	responsePayload := signMessageResponse{
		Message:   signParams.Message,
		Signature: signature,
	}

	publishResponse(&models.Response{
		ResultType: nip47Request.Method,
		Result:     responsePayload,
	}, nostr.Tags{})
}
