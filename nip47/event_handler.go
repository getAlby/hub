package nip47

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/getAlby/nostr-wallet-connect/db"
	"github.com/getAlby/nostr-wallet-connect/events"
	"github.com/getAlby/nostr-wallet-connect/lnclient"
	"github.com/getAlby/nostr-wallet-connect/logger"
	controllers "github.com/getAlby/nostr-wallet-connect/nip47/controllers"
	"github.com/getAlby/nostr-wallet-connect/nip47/models"
	"github.com/getAlby/nostr-wallet-connect/nip47/permissions"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func (svc *nip47Service) HandleEvent(ctx context.Context, sub *nostr.Subscription, event *nostr.Event, lnClient lnclient.LNClient) {
	var nip47Response *models.Response
	logger.Logger.WithFields(logrus.Fields{
		"requestEventNostrId": event.ID,
		"eventKind":           event.Kind,
	}).Info("Processing Event")

	ss, err := nip04.ComputeSharedSecret(event.PubKey, svc.keys.GetNostrSecretKey())
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"requestEventNostrId": event.ID,
			"eventKind":           event.Kind,
		}).WithError(err).Error("Failed to compute shared secret")
		return
	}

	// store request event
	requestEvent := db.RequestEvent{AppId: nil, NostrId: event.ID, State: db.REQUEST_EVENT_STATE_HANDLER_EXECUTING}
	err = svc.db.Create(&requestEvent).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			logger.Logger.WithFields(logrus.Fields{
				"requestEventNostrId": event.ID,
			}).Warn("Event already processed")
			return
		}
		logger.Logger.WithFields(logrus.Fields{
			"requestEventNostrId": event.ID,
			"eventKind":           event.Kind,
		}).WithError(err).Error("Failed to save nostr event")
		nip47Response = &models.Response{
			Error: &models.Error{
				Code:    models.ERROR_INTERNAL,
				Message: fmt.Sprintf("Failed to save nostr event: %s", err.Error()),
			},
		}
		resp, err := svc.CreateResponse(event, nip47Response, nostr.Tags{}, ss)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"requestEventNostrId": event.ID,
				"eventKind":           event.Kind,
			}).WithError(err).Error("Failed to process event")
		}
		svc.publishResponseEvent(ctx, sub, &requestEvent, resp, nil)
		return
	}

	app := db.App{}
	err = svc.db.First(&app, &db.App{
		NostrPubkey: event.PubKey,
	}).Error
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"nostrPubkey": event.PubKey,
		}).WithError(err).Error("Failed to find app for nostr pubkey")

		nip47Response = &models.Response{
			Error: &models.Error{
				Code:    models.ERROR_UNAUTHORIZED,
				Message: "The public key does not have a wallet connected.",
			},
		}
		resp, err := svc.CreateResponse(event, nip47Response, nostr.Tags{}, ss)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"requestEventNostrId": event.ID,
				"eventKind":           event.Kind,
			}).WithError(err).Error("Failed to process event")
		}
		svc.publishResponseEvent(ctx, sub, &requestEvent, resp, &app)

		requestEvent.State = db.REQUEST_EVENT_STATE_HANDLER_ERROR
		err = svc.db.Save(&requestEvent).Error
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"nostrPubkey": event.PubKey,
			}).WithError(err).Error("Failed to save state to nostr event")
		}
		return
	}

	requestEvent.AppId = &app.ID
	err = svc.db.Save(&requestEvent).Error
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"nostrPubkey": event.PubKey,
		}).WithError(err).Error("Failed to save app to nostr event")

		nip47Response = &models.Response{
			Error: &models.Error{
				Code:    models.ERROR_UNAUTHORIZED,
				Message: fmt.Sprintf("Failed to save app to nostr event: %s", err.Error()),
			},
		}
		resp, err := svc.CreateResponse(event, nip47Response, nostr.Tags{}, ss)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"requestEventNostrId": event.ID,
				"eventKind":           event.Kind,
			}).WithError(err).Error("Failed to process event")
		}
		svc.publishResponseEvent(ctx, sub, &requestEvent, resp, &app)

		requestEvent.State = db.REQUEST_EVENT_STATE_HANDLER_ERROR
		err = svc.db.Save(&requestEvent).Error
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"nostrPubkey": event.PubKey,
			}).WithError(err).Error("Failed to save state to nostr event")
		}

		return
	}

	logger.Logger.WithFields(logrus.Fields{
		"requestEventNostrId": event.ID,
		"eventKind":           event.Kind,
		"appId":               app.ID,
	}).Info("App found for nostr event")

	//to be extra safe, decrypt using the key found from the app
	ss, err = nip04.ComputeSharedSecret(app.NostrPubkey, svc.keys.GetNostrSecretKey())
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"requestEventNostrId": event.ID,
			"eventKind":           event.Kind,
		}).WithError(err).Error("Failed to process event")

		requestEvent.State = db.REQUEST_EVENT_STATE_HANDLER_ERROR
		err = svc.db.Save(&requestEvent).Error
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"nostrPubkey": event.PubKey,
			}).WithError(err).Error("Failed to save state to nostr event")
		}

		return
	}
	payload, err := nip04.Decrypt(event.Content, ss)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"requestEventNostrId": event.ID,
			"eventKind":           event.Kind,
			"appId":               app.ID,
		}).WithError(err).Error("Failed to decrypt content")
		logger.Logger.WithFields(logrus.Fields{
			"requestEventNostrId": event.ID,
			"eventKind":           event.Kind,
		}).WithError(err).Error("Failed to process event")

		requestEvent.State = db.REQUEST_EVENT_STATE_HANDLER_ERROR
		err = svc.db.Save(&requestEvent).Error
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"nostrPubkey": event.PubKey,
			}).WithError(err).Error("Failed to save state to nostr event")
		}

		return
	}
	nip47Request := &models.Request{}
	err = json.Unmarshal([]byte(payload), nip47Request)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"requestEventNostrId": event.ID,
			"eventKind":           event.Kind,
		}).WithError(err).Error("Failed to process event")

		requestEvent.State = db.REQUEST_EVENT_STATE_HANDLER_ERROR
		err = svc.db.Save(&requestEvent).Error
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"nostrPubkey": event.PubKey,
			}).WithError(err).Error("Failed to save state to nostr event")
		}

		return
	}

	requestEvent.Method = nip47Request.Method
	requestEvent.ContentData = payload
	svc.db.Save(&requestEvent) // we ignore potential DB errors here as this only saves the method and content data

	// TODO: replace with a channel
	// TODO: update all previous occurences of svc.publishResponseEvent to also use the channel
	publishResponse := func(nip47Response *models.Response, tags nostr.Tags) {
		resp, err := svc.CreateResponse(event, nip47Response, tags, ss)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"requestEventNostrId": event.ID,
				"eventKind":           event.Kind,
				"appId":               app.ID,
			}).WithError(err).Error("Failed to create response")
			requestEvent.State = db.REQUEST_EVENT_STATE_HANDLER_ERROR
		} else {
			err = svc.publishResponseEvent(ctx, sub, &requestEvent, resp, &app)
			if err != nil {
				logger.Logger.WithFields(logrus.Fields{
					"requestEventNostrId": event.ID,
					"eventKind":           event.Kind,
					"appId":               app.ID,
				}).WithError(err).Error("Failed to publish event")
				requestEvent.State = db.REQUEST_EVENT_STATE_HANDLER_ERROR
			} else {
				requestEvent.State = db.REQUEST_EVENT_STATE_HANDLER_EXECUTED
				logger.Logger.WithFields(logrus.Fields{
					"requestEventNostrId": event.ID,
					"eventKind":           event.Kind,
					"appId":               app.ID,
				}).Info("Published response")
			}
		}
		err = svc.db.Save(&requestEvent).Error
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"nostrPubkey": event.PubKey,
			}).WithError(err).Error("Failed to save state to nostr event")
		}
	}

	checkPermission := func(amountMsat uint64) *models.Response {
		scope, err := permissions.RequestMethodToScope(nip47Request.Method)
		if err != nil {
			return &models.Response{
				ResultType: nip47Request.Method,
				Error: &models.Error{
					Code:    models.ERROR_INTERNAL,
					Message: err.Error(),
				},
			}
		}
		hasPermission, code, message := svc.permissionsService.HasPermission(&app, scope, amountMsat)
		if !hasPermission {
			logger.Logger.WithFields(logrus.Fields{
				"request_event_id": requestEvent.ID,
				"app_id":           app.ID,
				"code":             code,
				"message":          message,
			}).Error("App does not have permission")

			svc.eventPublisher.Publish(&events.Event{
				Event: "nwc_permission_denied",
				Properties: map[string]interface{}{
					"request_method": nip47Request.Method,
					"app_name":       app.Name,
					// "app_pubkey":     app.NostrPubkey,
					"code":    code,
					"message": message,
				},
			})

			return &models.Response{
				ResultType: nip47Request.Method,
				Error: &models.Error{
					Code:    code,
					Message: message,
				},
			}
		}
		return nil
	}

	logger.Logger.WithFields(logrus.Fields{
		"requestEventNostrId": event.ID,
		"eventKind":           event.Kind,
		"appId":               app.ID,
		"method":              nip47Request.Method,
		"params":              nip47Request.Params,
	}).Info("Handling NIP-47 request")

	// TODO: controllers should share a common interface
	switch nip47Request.Method {
	case models.MULTI_PAY_INVOICE_METHOD:
		controllers.
			NewMultiPayInvoiceController(lnClient, svc.db, svc.eventPublisher).
			HandleMultiPayInvoiceEvent(ctx, nip47Request, requestEvent.ID, &app, checkPermission, publishResponse)
	case models.MULTI_PAY_KEYSEND_METHOD:
		controllers.
			NewMultiPayKeysendController(lnClient, svc.db, svc.eventPublisher).
			HandleMultiPayKeysendEvent(ctx, nip47Request, requestEvent.ID, &app, checkPermission, publishResponse)
	case models.PAY_INVOICE_METHOD:
		controllers.
			NewPayInvoiceController(lnClient, svc.db, svc.eventPublisher).
			HandlePayInvoiceEvent(ctx, nip47Request, requestEvent.ID, &app, checkPermission, publishResponse, nostr.Tags{})
	case models.PAY_KEYSEND_METHOD:
		controllers.
			NewPayKeysendController(lnClient, svc.db, svc.eventPublisher).
			HandlePayKeysendEvent(ctx, nip47Request, requestEvent.ID, &app, checkPermission, publishResponse, nostr.Tags{})
	case models.GET_BALANCE_METHOD:
		controllers.
			NewGetBalanceController(lnClient).
			HandleGetBalanceEvent(ctx, nip47Request, requestEvent.ID, checkPermission, publishResponse)
	case models.MAKE_INVOICE_METHOD:
		controllers.
			NewMakeInvoiceController(lnClient).
			HandleMakeInvoiceEvent(ctx, nip47Request, requestEvent.ID, checkPermission, publishResponse)
	case models.LOOKUP_INVOICE_METHOD:
		controllers.
			NewLookupInvoiceController(lnClient).
			HandleLookupInvoiceEvent(ctx, nip47Request, requestEvent.ID, checkPermission, publishResponse)
	case models.LIST_TRANSACTIONS_METHOD:
		controllers.
			NewListTransactionsController(lnClient).
			HandleListTransactionsEvent(ctx, nip47Request, requestEvent.ID, checkPermission, publishResponse)
	case models.GET_INFO_METHOD:
		controllers.
			NewGetInfoController(svc.permissionsService, lnClient).
			HandleGetInfoEvent(ctx, nip47Request, requestEvent.ID, &app, checkPermission, publishResponse)
	case models.SIGN_MESSAGE_METHOD:
		controllers.
			NewSignMessageController(lnClient).
			HandleSignMessageEvent(ctx, nip47Request, requestEvent.ID, checkPermission, publishResponse)
	default:
		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error: &models.Error{
				Code:    models.ERROR_NOT_IMPLEMENTED,
				Message: fmt.Sprintf("Unknown method: %s", nip47Request.Method),
			},
		}, nostr.Tags{})
	}
}

func (svc *nip47Service) CreateResponse(initialEvent *nostr.Event, content interface{}, tags nostr.Tags, ss []byte) (result *nostr.Event, err error) {
	payloadBytes, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}
	msg, err := nip04.Encrypt(string(payloadBytes), ss)
	if err != nil {
		return nil, err
	}

	allTags := nostr.Tags{[]string{"p", initialEvent.PubKey}, []string{"e", initialEvent.ID}}
	allTags = append(allTags, tags...)

	resp := &nostr.Event{
		PubKey:    svc.keys.GetNostrPublicKey(),
		CreatedAt: nostr.Now(),
		Kind:      models.RESPONSE_KIND,
		Tags:      allTags,
		Content:   msg,
	}
	err = resp.Sign(svc.keys.GetNostrSecretKey())
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (svc *nip47Service) publishResponseEvent(ctx context.Context, sub *nostr.Subscription, requestEvent *db.RequestEvent, resp *nostr.Event, app *db.App) error {
	var appId *uint
	if app != nil {
		appId = &app.ID
	}
	responseEvent := db.ResponseEvent{NostrId: resp.ID, RequestId: requestEvent.ID, State: "received"}
	err := svc.db.Create(&responseEvent).Error
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"requestEventNostrId": requestEvent.NostrId,
			"appId":               appId,
			"replyEventId":        resp.ID,
		}).WithError(err).Error("Failed to save response/reply event")
		return err
	}

	err = sub.Relay.Publish(ctx, *resp)
	if err != nil {
		responseEvent.State = db.RESPONSE_EVENT_STATE_PUBLISH_FAILED
		logger.Logger.WithFields(logrus.Fields{
			"requestEventId":       requestEvent.ID,
			"requestNostrEventId":  requestEvent.NostrId,
			"appId":                appId,
			"responseEventId":      responseEvent.ID,
			"responseNostrEventId": resp.ID,
		}).WithError(err).Error("Failed to publish reply")
	} else {
		responseEvent.State = db.RESPONSE_EVENT_STATE_PUBLISH_CONFIRMED
		responseEvent.RepliedAt = time.Now()
		logger.Logger.WithFields(logrus.Fields{
			"requestEventId":       requestEvent.ID,
			"requestNostrEventId":  requestEvent.NostrId,
			"appId":                appId,
			"responseEventId":      responseEvent.ID,
			"responseNostrEventId": resp.ID,
		}).Info("Published reply")
	}

	err = svc.db.Save(&responseEvent).Error
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"requestEventId":       requestEvent.ID,
			"requestNostrEventId":  requestEvent.NostrId,
			"appId":                appId,
			"responseEventId":      responseEvent.ID,
			"responseNostrEventId": resp.ID,
		}).WithError(err).Error("Failed to update response/reply event")
		return err
	}

	return nil
}
