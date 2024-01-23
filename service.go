package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Service struct {
	cfg         *Config
	db          *gorm.DB
	lnClient    LNClient
	ReceivedEOS bool
	Logger      *logrus.Logger
}

/*var supportedMethods = map[string]bool{
	NIP_47_PAY_INVOICE_METHOD:       true,
	NIP_47_GET_BALANCE_METHOD:       true,
	NIP_47_GET_INFO_METHOD:          true,
	NIP_47_MAKE_INVOICE_METHOD:      true,
	NIP_47_LOOKUP_INVOICE_METHOD:    true,
	NIP_47_LIST_TRANSACTIONS_METHOD: true,
}*/

func (svc *Service) GetUser(c echo.Context) (user *User, err error) {
	sess, _ := session.Get(CookieName, c)
	userID := sess.Values["user_id"]
	if svc.cfg.LNBackendType == LNDBackendType {
		//if we self-host, there is always only one user
		userID = 1
	}
	if userID == nil {
		return nil, nil
	}
	user = &User{}
	err = svc.db.Preload("Apps").First(&user, userID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return
}

func (svc *Service) StartSubscription(ctx context.Context, sub *nostr.Subscription) error {
	go func() {
		// block till EOS is reached
		<-sub.EndOfStoredEvents
		svc.ReceivedEOS = true
		svc.Logger.Info("Received EOS")

		for event := range sub.Events {
			go svc.handleAndPublishEvent(ctx, sub, event)
		}
		svc.Logger.Info("Subscription ended")
	}()

	select {
	case <-sub.Relay.Context().Done():
		svc.Logger.Errorf("Relay error %v", sub.Relay.ConnectionError)
		return sub.Relay.ConnectionError
	case <-ctx.Done():
		if ctx.Err() != context.Canceled {
			svc.Logger.Errorf("Subscription error %v", ctx.Err())
			return ctx.Err()
		}
		svc.Logger.Info("Exiting subscription.")
		return nil
	}
}

func (svc *Service) PublishEvent(ctx context.Context, sub *nostr.Subscription, event *nostr.Event, resp *nostr.Event) {
	status, err := sub.Relay.Publish(ctx, *resp)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":      event.ID,
			"status":       status,
			"replyEventId": resp.ID,
		}).Errorf("Failed to publish reply: %v", err)
		return
	}

	nostrEvent := NostrEvent{}
	result := svc.db.Where("nostr_id = ?", event.ID).First(&nostrEvent)
	if result.Error != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":      event.ID,
			"status":       status,
			"replyEventId": resp.ID,
		}).Error(result.Error)
		return
	}
	nostrEvent.ReplyId = resp.ID

	if status == nostr.PublishStatusSucceeded {
		nostrEvent.State = NOSTR_EVENT_STATE_PUBLISH_CONFIRMED
		nostrEvent.RepliedAt = time.Now()
		svc.db.Save(&nostrEvent)
		svc.Logger.WithFields(logrus.Fields{
			"nostrEventId": nostrEvent.ID,
			"eventId":      event.ID,
			"status":       status,
			"replyEventId": resp.ID,
			"appId":        nostrEvent.AppId,
		}).Info("Published reply")
	} else if status == nostr.PublishStatusFailed {
		nostrEvent.State = NOSTR_EVENT_STATE_PUBLISH_FAILED
		svc.db.Save(&nostrEvent)
		svc.Logger.WithFields(logrus.Fields{
			"nostrEventId": nostrEvent.ID,
			"eventId":      event.ID,
			"status":       status,
			"replyEventId": resp.ID,
			"appId":        nostrEvent.AppId,
		}).Info("Failed to publish reply")
	} else {
		nostrEvent.State = NOSTR_EVENT_STATE_PUBLISH_UNCONFIRMED
		svc.db.Save(&nostrEvent)
		svc.Logger.WithFields(logrus.Fields{
			"nostrEventId": nostrEvent.ID,
			"eventId":      event.ID,
			"status":       status,
			"replyEventId": resp.ID,
			"appId":        nostrEvent.AppId,
		}).Info("Reply sent but no response from relay (timeout)")
	}
}

func (svc *Service) handleAndPublishEvent(ctx context.Context, sub *nostr.Subscription, event *nostr.Event) {
	var resp *nostr.Event
	svc.Logger.WithFields(logrus.Fields{
		"eventId":   event.ID,
		"eventKind": event.Kind,
	}).Info("Processing Event")

	// make sure we don't know the event, yet
	nostrEvent := NostrEvent{}
	findEventResult := svc.db.Where("nostr_id = ?", event.ID).Find(&nostrEvent)
	if findEventResult.RowsAffected != 0 {
		svc.Logger.WithFields(logrus.Fields{
			"eventId": event.ID,
		}).Warn("Event already processed")
		return
	}

	app := App{}
	err := svc.db.Preload("User").First(&app, &App{
		NostrPubkey: event.PubKey,
	}).Error
	if err != nil {
		ss, err := nip04.ComputeSharedSecret(event.PubKey, svc.cfg.NostrSecretKey)
		if err != nil {
			svc.Logger.WithFields(logrus.Fields{
				"eventId":   event.ID,
				"eventKind": event.Kind,
			}).Errorf("Failed to process event: %v", err)
			return
		}
		resp, _ = svc.createResponse(event, Nip47Response{
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_UNAUTHORIZED,
				Message: "The public key does not have a wallet connected.",
			},
		}, nostr.Tags{}, ss)
		svc.PublishEvent(ctx, sub, event, resp)
		return
	}

	svc.Logger.WithFields(logrus.Fields{
		"eventId":   event.ID,
		"eventKind": event.Kind,
		"appId":     app.ID,
	}).Info("App found for nostr event")

	//to be extra safe, decrypt using the key found from the app
	ss, err := nip04.ComputeSharedSecret(app.NostrPubkey, svc.cfg.NostrSecretKey)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
		}).Errorf("Failed to process event: %v", err)
		return
	}
	payload, err := nip04.Decrypt(event.Content, ss)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
			"appId":     app.ID,
		}).Errorf("Failed to decrypt content: %v", err)
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
		}).Errorf("Failed to process event: %v", err)
		return
	}
	nip47Request := &Nip47Request{}
	err = json.Unmarshal([]byte(payload), nip47Request)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
		}).Errorf("Failed to process event: %v", err)
		return
	}

	// TODO: consider move event publishing to individual methods - multi_* methods are
	// inconsistent with single methods because they publish multiple responses
	switch nip47Request.Method {
	case NIP_47_MULTI_PAY_INVOICE_METHOD:
		svc.HandleMultiPayInvoiceEvent(ctx, sub, nip47Request, event, app, ss)
		return
	case NIP_47_PAY_INVOICE_METHOD:
		resp, err = svc.HandlePayInvoiceEvent(ctx, nip47Request, event, app, ss)
	case NIP_47_PAY_KEYSEND_METHOD:
		resp, err = svc.HandlePayKeysendEvent(ctx, nip47Request, event, app, ss)
	case NIP_47_GET_BALANCE_METHOD:
		resp, err = svc.HandleGetBalanceEvent(ctx, nip47Request, event, app, ss)
	case NIP_47_MAKE_INVOICE_METHOD:
		resp, err = svc.HandleMakeInvoiceEvent(ctx, nip47Request, event, app, ss)
	case NIP_47_LOOKUP_INVOICE_METHOD:
		resp, err = svc.HandleLookupInvoiceEvent(ctx, nip47Request, event, app, ss)
	case NIP_47_LIST_TRANSACTIONS_METHOD:
		resp, err = svc.HandleListTransactionsEvent(ctx, nip47Request, event, app, ss)
	case NIP_47_GET_INFO_METHOD:
		resp, err = svc.HandleGetInfoEvent(ctx, nip47Request, event, app, ss)
	default:
		resp, err = svc.handleUnknownMethod(ctx, nip47Request, event, app, ss)
	}

	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
		}).Errorf("Failed to process event: %v", err)
	}
	if resp != nil {
		svc.PublishEvent(ctx, sub, event, resp)
	}
}

func (svc *Service) handleUnknownMethod(ctx context.Context, request *Nip47Request, event *nostr.Event, app App, ss []byte) (result *nostr.Event, err error) {
	return svc.createResponse(event, Nip47Response{
		ResultType: request.Method,
		Error: &Nip47Error{
			Code:    NIP_47_ERROR_NOT_IMPLEMENTED,
			Message: fmt.Sprintf("Unknown method: %s", request.Method),
		}}, nostr.Tags{}, ss)
}

func (svc *Service) createResponse(initialEvent *nostr.Event, content interface{}, tags nostr.Tags, ss []byte) (result *nostr.Event, err error) {
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
		PubKey:    svc.cfg.IdentityPubkey,
		CreatedAt: nostr.Now(),
		Kind:      NIP_47_RESPONSE_KIND,
		Tags:      allTags,
		Content:   msg,
	}
	err = resp.Sign(svc.cfg.NostrSecretKey)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (svc *Service) GetMethods(app *App) []string {
	appPermissions := []AppPermission{}
	findPermissionsResult := svc.db.Find(&appPermissions, &AppPermission{
		AppId: app.ID,
	})
	if findPermissionsResult.RowsAffected == 0 {
		// No permissions created for this app. It can do anything
		return strings.Split(NIP_47_CAPABILITIES, ",")
	}
	requestMethods := make([]string, 0, len(appPermissions))
	for _, appPermission := range appPermissions {
		requestMethods = append(requestMethods, appPermission.RequestMethod)
	}
	return requestMethods
}

func (svc *Service) hasPermission(app *App, event *nostr.Event, requestMethod string, amount int64) (result bool, code string, message string) {
	// find all permissions for the app
	appPermissions := []AppPermission{}
	findPermissionsResult := svc.db.Find(&appPermissions, &AppPermission{
		AppId: app.ID,
	})
	if findPermissionsResult.RowsAffected == 0 {
		// No permissions created for this app. It can do anything
		svc.Logger.WithFields(logrus.Fields{
			"eventId":       event.ID,
			"requestMethod": requestMethod,
			"appId":         app.ID,
			"pubkey":        app.NostrPubkey,
		}).Info("No permissions found for app")
		return true, "", ""
	}

	appPermission := AppPermission{}
	findPermissionResult := findPermissionsResult.Limit(1).Find(&appPermission, &AppPermission{
		RequestMethod: requestMethod,
	})
	if findPermissionResult.RowsAffected == 0 {
		// No permission for this request method
		return false, NIP_47_ERROR_RESTRICTED, fmt.Sprintf("This app does not have permission to request %s", requestMethod)
	}
	expiresAt := appPermission.ExpiresAt
	if !expiresAt.IsZero() && expiresAt.Before(time.Now()) {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":       event.ID,
			"requestMethod": requestMethod,
			"expiresAt":     expiresAt.Unix(),
			"appId":         app.ID,
			"pubkey":        app.NostrPubkey,
		}).Info("This pubkey is expired")
		return false, NIP_47_ERROR_EXPIRED, "This app has expired"
	}

	if requestMethod == NIP_47_PAY_INVOICE_METHOD {
		maxAmount := appPermission.MaxAmount
		if maxAmount != 0 {
			budgetUsage := svc.GetBudgetUsage(&appPermission)

			if budgetUsage+amount/1000 > int64(maxAmount) {
				return false, NIP_47_ERROR_QUOTA_EXCEEDED, "Insufficient budget remaining to make payment"
			}
		}
	}
	return true, "", ""
}

func (svc *Service) GetBudgetUsage(appPermission *AppPermission) int64 {
	var result struct {
		Sum uint
	}
	svc.db.Table("payments").Select("SUM(amount) as sum").Where("app_id = ? AND preimage IS NOT NULL AND created_at > ?", appPermission.AppId, GetStartOfBudget(appPermission.BudgetRenewal, appPermission.App.CreatedAt)).Scan(&result)
	return int64(result.Sum)
}

func (svc *Service) PublishNip47Info(ctx context.Context, relay *nostr.Relay) error {
	ev := &nostr.Event{}
	ev.Kind = NIP_47_INFO_EVENT_KIND
	ev.Content = NIP_47_CAPABILITIES
	ev.CreatedAt = nostr.Now()
	ev.PubKey = svc.cfg.IdentityPubkey
	err := ev.Sign(svc.cfg.NostrSecretKey)
	if err != nil {
		return err
	}
	status, err := relay.Publish(ctx, *ev)
	if err != nil || status != nostr.PublishStatusSucceeded {
		return fmt.Errorf("Nostr publish not successful: %s error: %s", status, err)
	}
	return nil
}
