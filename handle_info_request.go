package main

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

func (svc *Service) HandleGetInfoEvent(ctx context.Context, request *Nip47Request, requestEvent *NostrEvent, app *App) (result *Nip47Response, err error) {

	hasPermission, code, message := svc.hasPermission(app, request.Method, 0)

	if !hasPermission {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   requestEvent.NostrId,
			"eventKind": requestEvent.Kind,
			"appId":     app.ID,
		}).Errorf("App does not have permission: %s %s", code, message)

		return &Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    code,
				Message: message,
			}}, nil
	}

	svc.Logger.WithFields(logrus.Fields{
		"eventId":   requestEvent.NostrId,
		"eventKind": requestEvent.Kind,
		"appId":     app.ID,
	}).Info("Fetching node info")

	info, err := svc.lnClient.GetInfo(ctx)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"senderPubkey": requestEvent.PubKey,
			"eventId":      requestEvent.NostrId,
			"eventKind":    requestEvent.Kind,
			"appId":        app.ID,
		}).Infof("Failed to fetch node info: %v", err)
		return &Nip47Response{
			ResultType: request.Method,
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: fmt.Sprintf("Something went wrong while fetching node info: %s", err.Error()),
			},
		}, nil
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
	}, nil
}
