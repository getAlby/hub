package main

import (
	"context"

	"github.com/sirupsen/logrus"
)

func (svc *Service) HandleGetInfoEvent(ctx context.Context, request *Nip47Request, requestEvent *RequestEvent, app *App) *Nip47Response {

	resp := svc.checkPermission(request, requestEvent, app, 0)
	if resp != nil {
		return resp
	}

	svc.Logger.WithFields(logrus.Fields{
		"eventId": requestEvent.NostrId,
		"appId":   app.ID,
	}).Info("Fetching node info")

	info, err := svc.lnClient.GetInfo(ctx)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId": requestEvent.NostrId,
			"appId":   app.ID,
		}).Infof("Failed to fetch node info: %v", err)
		return &Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: err.Error(),
			},
		}
	}

	responsePayload := &Nip47GetInfoResponse{
		Alias:       info.Alias,
		Color:       info.Color,
		Pubkey:      info.Pubkey,
		Network:     info.Network,
		BlockHeight: info.BlockHeight,
		BlockHash:   info.BlockHash,
		Methods:     svc.GetMethods(app),
	}
	return &Nip47Response{
		ResultType: request.Method,
		Result:     responsePayload,
	}
}
