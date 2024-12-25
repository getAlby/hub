package nip47

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/controllers"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/nip47/permissions"
	nostrmodels "github.com/getAlby/hub/nostr/models"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/nbd-wtf/go-nostr/nip44"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func (svc *nip47Service) HandleEvent(ctx context.Context, relay nostrmodels.Relay, event *nostr.Event, lnClient lnclient.LNClient) {
	var nip47Response *models.Response
	logger.Logger.WithFields(logrus.Fields{
		"requestEventNostrId": event.ID,
		"eventKind":           event.Kind,
	}).Info("Processing Event")

	// go-nostr already checks this, but just to be sure:
	validEventSignature, err := event.CheckSignature()
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"requestEventNostrId": event.ID,
			"eventKind":           event.Kind,
		}).WithError(err).Error("invalid event signature")
		return
	}
	if !validEventSignature {
		logger.Logger.WithFields(logrus.Fields{
			"requestEventNostrId": event.ID,
			"eventKind":           event.Kind,
		}).Error("invalid event signature")
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
		return
	}
	app := db.App{}
	err = svc.db.First(&app, &db.App{
		AppPubkey: event.PubKey,
	}).Error
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"appPubkey": event.PubKey,
		}).WithError(err).Error("Failed to find app for nostr pubkey")
		return
	}

	logger.Logger.WithFields(logrus.Fields{
		"requestEventNostrId": event.ID,
		"eventKind":           event.Kind,
		"appId":               app.ID,
	}).Debug("App found for nostr event")

	appWalletPrivKey := svc.keys.GetNostrSecretKey()

	if app.WalletPubkey != nil {
		// This is a new child key derived from master using app ID as index
		appWalletPrivKey, err = svc.keys.GetAppWalletKey(app.ID)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"appId": app.ID,
			}).WithError(err).Error("error deriving child key")
			return
		}
	}

	var ss []byte
	var ck [32]byte
	isNip04Encrypted := strings.Contains(event.Content, "?iv=")
	if isNip04Encrypted {
		ss, err = nip04.ComputeSharedSecret(app.AppPubkey, appWalletPrivKey)
	} else {
		ck, err = nip44.GenerateConversationKey(app.AppPubkey, appWalletPrivKey)
	}

	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"requestEventNostrId": event.ID,
			"eventKind":           event.Kind,
			"isNip04Encrypted":    isNip04Encrypted,
		}).WithError(err).Error("Failed to compute shared secret or conversation key")

		requestEvent.State = db.REQUEST_EVENT_STATE_HANDLER_ERROR
		err = svc.db.Save(&requestEvent).Error
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"appPubkey": event.PubKey,
			}).WithError(err).Error("Failed to save state to nostr event")
		}
		return
	}

	requestEvent.AppId = &app.ID
	err = svc.db.Save(&requestEvent).Error
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"appPubkey": event.PubKey,
		}).WithError(err).Error("Failed to save app to nostr event")

		nip47Response = &models.Response{
			Error: &models.Error{
				Code:    constants.ERROR_INTERNAL,
				Message: fmt.Sprintf("Failed to save app to nostr event: %s", err.Error()),
			},
		}
		resp, err := svc.CreateResponse(event, nip47Response, nostr.Tags{}, ss, ck, appWalletPrivKey)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"requestEventNostrId": event.ID,
				"eventKind":           event.Kind,
			}).WithError(err).Error("Failed to process event")
		}
		svc.publishResponseEvent(ctx, relay, &requestEvent, resp, &app)

		requestEvent.State = db.REQUEST_EVENT_STATE_HANDLER_ERROR
		err = svc.db.Save(&requestEvent).Error
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"appPubkey": event.PubKey,
			}).WithError(err).Error("Failed to save state to nostr event")
		}

		return
	}

	var payload string
	if isNip04Encrypted {
		payload, err = nip04.Decrypt(event.Content, ss)
	} else {
		payload, err = nip44.Decrypt(event.Content, ck)
	}

	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"requestEventNostrId": event.ID,
			"eventKind":           event.Kind,
			"appId":               app.ID,
		}).WithError(err).Error("Failed to decrypt content")
		logger.Logger.WithFields(logrus.Fields{
			"requestEventNostrId": event.ID,
			"eventKind":           event.Kind,
		}).WithError(err).Error("Failed to decrypt request event")

		requestEvent.State = db.REQUEST_EVENT_STATE_HANDLER_ERROR
		err = svc.db.Save(&requestEvent).Error
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"appPubkey": event.PubKey,
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
				"appPubkey": event.PubKey,
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
		resp, err := svc.CreateResponse(event, nip47Response, tags, ss, ck, appWalletPrivKey)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"requestEventNostrId": event.ID,
				"eventKind":           event.Kind,
				"appId":               app.ID,
			}).WithError(err).Error("Failed to create response")
			requestEvent.State = db.REQUEST_EVENT_STATE_HANDLER_ERROR
		} else {
			err = svc.publishResponseEvent(ctx, relay, &requestEvent, resp, &app)
			if err != nil {
				logger.Logger.WithFields(logrus.Fields{
					"requestEventNostrId":  event.ID,
					"responseEventNostrId": resp.ID,
					"eventKind":            event.Kind,
					"appId":                app.ID,
				}).WithError(err).Error("Failed to publish event")
				requestEvent.State = db.REQUEST_EVENT_STATE_HANDLER_ERROR
			} else {
				requestEvent.State = db.REQUEST_EVENT_STATE_HANDLER_EXECUTED
				logger.Logger.WithFields(logrus.Fields{
					"requestEventNostrId":  event.ID,
					"responseEventNostrId": resp.ID,
					"eventKind":            event.Kind,
					"appId":                app.ID,
				}).Debug("Published response")
			}
		}
		err = svc.db.Save(&requestEvent).Error
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"appPubkey": event.PubKey,
			}).WithError(err).Error("Failed to save state to nostr event")
		}
	}

	logger.Logger.WithFields(logrus.Fields{
		"requestEventNostrId": event.ID,
		"eventKind":           event.Kind,
		"appId":               app.ID,
		"method":              nip47Request.Method,
		"params":              nip47Request.Params,
	}).Debug("Handling NIP-47 request")

	version := "0.0"
	vTag := event.Tags.GetFirst([]string{"v"})

	if vTag != nil && vTag.Value() != "" {
		version = vTag.Value()
	}

	isVersionSupported, err := svc.isVersionSupported(version, isNip04Encrypted)
	if !isVersionSupported {
		logger.Logger.WithFields(logrus.Fields{
			"request_event_id": requestEvent.ID,
			"app_id":           app.ID,
			"version":          version,
		}).Error(err.Error())

		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error: &models.Error{
				Code:    constants.ERROR_UNSUPPORTED_VERSION,
				Message: err.Error(),
			},
		}, nostr.Tags{})
		return
	}

	if !slices.Contains(permissions.GetAlwaysGrantedMethods(), nip47Request.Method) {
		scope, err := permissions.RequestMethodToScope(nip47Request.Method)
		if err != nil {
			publishResponse(&models.Response{
				ResultType: nip47Request.Method,
				Error: &models.Error{
					Code:    constants.ERROR_INTERNAL,
					Message: err.Error(),
				},
			}, nostr.Tags{})
			return
		}
		hasPermission, code, message := svc.permissionsService.HasPermission(&app, scope)
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
					// "app_pubkey":     app.AppPubkey,
					"code":    code,
					"message": message,
				},
			})

			publishResponse(&models.Response{
				ResultType: nip47Request.Method,
				Error: &models.Error{
					Code:    code,
					Message: message,
				},
			}, nostr.Tags{})
			return
		}
	}

	controller := controllers.NewNip47Controller(lnClient, svc.db, svc.eventPublisher, svc.permissionsService, svc.transactionsService)

	switch nip47Request.Method {
	case models.MULTI_PAY_INVOICE_METHOD:
		controller.
			HandleMultiPayInvoiceEvent(ctx, nip47Request, requestEvent.ID, &app, publishResponse)
	case models.MULTI_PAY_KEYSEND_METHOD:
		controller.
			HandleMultiPayKeysendEvent(ctx, nip47Request, requestEvent.ID, &app, publishResponse)
	case models.PAY_INVOICE_METHOD:
		controller.
			HandlePayInvoiceEvent(ctx, nip47Request, requestEvent.ID, &app, publishResponse, nostr.Tags{})
	case models.PAY_KEYSEND_METHOD:
		controller.
			HandlePayKeysendEvent(ctx, nip47Request, requestEvent.ID, &app, publishResponse, nostr.Tags{})
	case models.GET_BALANCE_METHOD:
		controller.
			HandleGetBalanceEvent(ctx, nip47Request, requestEvent.ID, &app, publishResponse)
	case models.GET_BUDGET_METHOD:
		controller.
			HandleGetBudgetEvent(ctx, nip47Request, requestEvent.ID, &app, publishResponse)
	case models.MAKE_INVOICE_METHOD:
		controller.
			HandleMakeInvoiceEvent(ctx, nip47Request, requestEvent.ID, app.ID, publishResponse)
	case models.LOOKUP_INVOICE_METHOD:
		controller.
			HandleLookupInvoiceEvent(ctx, nip47Request, requestEvent.ID, app.ID, publishResponse)
	case models.LIST_TRANSACTIONS_METHOD:
		controller.
			HandleListTransactionsEvent(ctx, nip47Request, requestEvent.ID, app.ID, publishResponse)
	case models.GET_INFO_METHOD:
		controller.
			HandleGetInfoEvent(ctx, nip47Request, requestEvent.ID, &app, publishResponse)
	case models.SIGN_MESSAGE_METHOD:
		controller.
			HandleSignMessageEvent(ctx, nip47Request, requestEvent.ID, publishResponse)
	default:
		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error: &models.Error{
				Code:    constants.ERROR_NOT_IMPLEMENTED,
				Message: fmt.Sprintf("Unknown method: %s", nip47Request.Method),
			},
		}, nostr.Tags{})
	}
}

func (svc *nip47Service) CreateResponse(initialEvent *nostr.Event, content interface{}, tags nostr.Tags, ss []byte, ck [32]byte, appWalletPrivKey string) (result *nostr.Event, err error) {
	payloadBytes, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}

	var msg string
	if ss != nil {
		msg, err = nip04.Encrypt(string(payloadBytes), ss)
		if err != nil {
			return nil, err
		}
	} else {
		msg, err = nip44.Encrypt(string(payloadBytes), ck)
		if err != nil {
			return nil, err
		}
	}

	allTags := nostr.Tags{[]string{"p", initialEvent.PubKey}, []string{"e", initialEvent.ID}}
	allTags = append(allTags, tags...)

	appWalletPubKey, err := nostr.GetPublicKey(appWalletPrivKey)
	if err != nil {
		logger.Logger.WithError(err).Error("Error converting nostr privkey to pubkey")
		return
	}

	resp := &nostr.Event{
		PubKey:    appWalletPubKey,
		CreatedAt: nostr.Now(),
		Kind:      models.RESPONSE_KIND,
		Tags:      allTags,
		Content:   msg,
	}
	err = resp.Sign(appWalletPrivKey)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (svc *nip47Service) publishResponseEvent(ctx context.Context, relay nostrmodels.Relay, requestEvent *db.RequestEvent, resp *nostr.Event, app *db.App) error {
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

	err = relay.Publish(ctx, *resp)
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

func (svc *nip47Service) isVersionSupported(version string, isNip04Encrypted bool) (bool, error) {
	versionParts := strings.Split(version, ".")
	if len(versionParts) != 2 {
		return false, fmt.Errorf("invalid version format: %s", version)
	}

	major, err := strconv.Atoi(versionParts[0])
	if err != nil {
		return false, fmt.Errorf("invalid major version: %s", versionParts[0])
	}
	minor, err := strconv.Atoi(versionParts[1])
	if err != nil {
		return false, fmt.Errorf("invalid minor version: %s", versionParts[1])
	}

	for _, supported := range strings.Split(models.SUPPORTED_VERSIONS, " ") {
		supportedParts := strings.Split(supported, ".")
		if len(supportedParts) != 2 {
			continue
		}

		supportedMajor, _ := strconv.Atoi(supportedParts[0])
		supportedMinor, _ := strconv.Atoi(supportedParts[1])

		if major == supportedMajor {
			if minor <= supportedMinor {
				if major > 0 && isNip04Encrypted {
					return false, fmt.Errorf("used nip-04 encryption with version: %s", version)
				}
				return true, nil
			}
			return false, fmt.Errorf("invalid version: %s", version)
		}
	}
	return false, fmt.Errorf("invalid version: %s", version)
}
