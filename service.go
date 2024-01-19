package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Service struct {
	// config from .env only. Fetch dynamic config from db
	cfg         *Config
	db          *gorm.DB
	lnClient    LNClient
	ReceivedEOS bool
	Logger      *logrus.Logger
	ctx         context.Context
	wg          *sync.WaitGroup
}

func (svc *Service) GetUser(c echo.Context) (user *User, err error) {
	sess, _ := session.Get(CookieName, c)
	userID := sess.Values["user_id"]

	// FIXME: split app into single-user and multi-user
	if svc.cfg.LNBackendType != AlbyBackendType {
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
		<-sub.EndOfStoredEvents
		svc.ReceivedEOS = true
		svc.Logger.Info("Received EOS")
	}()

	go func() {
		for event := range sub.Events {
			go func(event *nostr.Event) {
				resp, err := svc.HandleEvent(ctx, event)
				if err != nil {
					svc.Logger.WithFields(logrus.Fields{
						"eventId":   event.ID,
						"eventKind": event.Kind,
					}).Errorf("Failed to process event: %v", err)
				}
				if resp != nil {
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
			}(event)
		}
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
		svc.Logger.Info("Exiting subscription...")
		return nil
	}
}

func (svc *Service) HandleEvent(ctx context.Context, event *nostr.Event) (result *nostr.Event, err error) {
	//don't process historical events
	if !svc.ReceivedEOS {
		return nil, nil
	}
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
		return nil, nil
	}

	app := App{}
	err = svc.db.Preload("User").First(&app, &App{
		NostrPubkey: event.PubKey,
	}).Error
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"nostrPubkey": event.PubKey,
		}).Errorf("Failed to find app for nostr pubkey: %v", err)

		ss, err := nip04.ComputeSharedSecret(event.PubKey, svc.cfg.NostrSecretKey)
		if err != nil {
			return nil, err
		}
		resp, _ := svc.createResponse(event, Nip47Response{
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_UNAUTHORIZED,
				Message: "The public key does not have a wallet connected.",
			},
		}, ss)
		return resp, err
	}

	svc.Logger.WithFields(logrus.Fields{
		"eventId":   event.ID,
		"eventKind": event.Kind,
		"appId":     app.ID,
	}).Info("App found for nostr event")

	//to be extra safe, decrypt using the key found from the app
	ss, err := nip04.ComputeSharedSecret(app.NostrPubkey, svc.cfg.NostrSecretKey)
	if err != nil {
		return nil, err
	}
	payload, err := nip04.Decrypt(event.Content, ss)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
			"appId":     app.ID,
		}).Errorf("Failed to decrypt content: %v", err)
		return nil, err
	}
	nip47Request := &Nip47Request{}
	err = json.Unmarshal([]byte(payload), nip47Request)
	if err != nil {
		return nil, err
	}
	switch nip47Request.Method {
	case NIP_47_PAY_INVOICE_METHOD:
		return svc.HandlePayInvoiceEvent(ctx, nip47Request, event, app, ss)
	case NIP_47_PAY_KEYSEND_METHOD:
		return svc.HandlePayKeysendEvent(ctx, nip47Request, event, app, ss)
	case NIP_47_GET_BALANCE_METHOD:
		return svc.HandleGetBalanceEvent(ctx, nip47Request, event, app, ss)
	case NIP_47_MAKE_INVOICE_METHOD:
		return svc.HandleMakeInvoiceEvent(ctx, nip47Request, event, app, ss)
	case NIP_47_LOOKUP_INVOICE_METHOD:
		return svc.HandleLookupInvoiceEvent(ctx, nip47Request, event, app, ss)
	case NIP_47_LIST_TRANSACTIONS_METHOD:
		return svc.HandleListTransactionsEvent(ctx, nip47Request, event, app, ss)
	case NIP_47_GET_INFO_METHOD:
		return svc.HandleGetInfoEvent(ctx, nip47Request, event, app, ss)
	default:
		return svc.createResponse(event, Nip47Response{
			ResultType: nip47Request.Method,
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_NOT_IMPLEMENTED,
				Message: fmt.Sprintf("Unknown method: %s", nip47Request.Method),
			}}, ss)
	}
}

func (svc *Service) createResponse(initialEvent *nostr.Event, content interface{}, ss []byte) (result *nostr.Event, err error) {
	payloadBytes, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}
	msg, err := nip04.Encrypt(string(payloadBytes), ss)
	if err != nil {
		return nil, err
	}
	resp := &nostr.Event{
		PubKey:    svc.cfg.IdentityPubkey,
		CreatedAt: nostr.Now(),
		Kind:      NIP_47_RESPONSE_KIND,
		Tags:      nostr.Tags{[]string{"p", initialEvent.PubKey}, []string{"e", initialEvent.ID}},
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
