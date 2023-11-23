package main

import (
	"context"
	"fmt"

	"github.com/nbd-wtf/go-nostr"
	"github.com/sirupsen/logrus"
)

func (svc *Service) HandleGetInfoEvent(ctx context.Context, request *Nip47Request, event *nostr.Event, app App, ss []byte) (result *nostr.Event, err error) {

	nostrEvent := NostrEvent{App: app, NostrId: event.ID, Content: event.Content, State: "received"}
	err = svc.db.Create(&nostrEvent).Error
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
			"appId":     app.ID,
		}).Errorf("Failed to save nostr event: %v", err)
		return nil, err
	}

	hasPermission, code, message := svc.hasPermission(&app, event, request.Method, nil)

	if !hasPermission {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
			"appId":     app.ID,
		}).Errorf("App does not have permission: %s %s", code, message)

		return svc.createResponse(event, Nip47Response{
			ResultType: NIP_47_GET_INFO_METHOD,
			Error: &Nip47Error{
			Code:    code,
			Message: message,
		}}, ss)
	}

	svc.Logger.WithFields(logrus.Fields{
		"eventId":   event.ID,
		"eventKind": event.Kind,
		"appId":     app.ID,
	}).Info("Fetching node info")

	alias, color, pubkey, network, block_height, block_hash, err := svc.lnClient.GetInfo(ctx, event.PubKey)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
			"appId":     app.ID,
		}).Infof("Failed to fetch node info: %v", err)
		nostrEvent.State = NOSTR_EVENT_STATE_HANDLER_ERROR
		svc.db.Save(&nostrEvent)
		return svc.createResponse(event, Nip47Response{
			ResultType: NIP_47_GET_BALANCE_METHOD,
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: fmt.Sprintf("Something went wrong while fetching node info: %s", err.Error()),
			},
		}, ss)
	}

	responsePayload := &Nip47GetInfoResponse{
		Alias:       alias,
		Color:       color,
		Pubkey:      pubkey,
		Network:     network,
		BlockHeight: block_height,
		BlockHash:   block_hash,
	}

	nostrEvent.State = NOSTR_EVENT_STATE_HANDLER_EXECUTED
	svc.db.Save(&nostrEvent)
	return svc.createResponse(event, Nip47Response{
		ResultType: NIP_47_GET_BALANCE_METHOD,
		Result:     responsePayload,
	},
		ss)
}
