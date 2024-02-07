package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"database/sql"
	"errors"
	"os"
	"os/signal"
	"path"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/sirupsen/logrus"

	"github.com/getAlby/nostr-wallet-connect/migrations"
	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/orandin/lumberjackrus"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Service struct {
	// config from .env only. Fetch dynamic config from db
	cfg      *Config
	db       *gorm.DB
	lnClient LNClient
	Logger   *logrus.Logger
	ctx      context.Context
	wg       *sync.WaitGroup
}

// TODO: move to service.go
func NewService(ctx context.Context) (*Service, error) {
	// Load config from environment variables / .env file
	godotenv.Load(".env")
	appConfig := &AppConfig{}
	err := envconfig.Process("", appConfig)
	if err != nil {
		return nil, err
	}

	logger := log.New()
	logger.SetFormatter(&log.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(log.InfoLevel)

	// make sure workdir exists
	os.MkdirAll(appConfig.Workdir, os.ModePerm)

	fileLoggerHook, err := lumberjackrus.NewHook(
		&lumberjackrus.LogFile{
			Filename: path.Join(appConfig.Workdir, "log/nwc-general.log"),
		},
		log.InfoLevel,
		&log.JSONFormatter{},
		&lumberjackrus.LogFileOpts{
			log.ErrorLevel: &lumberjackrus.LogFile{
				Filename:   path.Join(appConfig.Workdir, "log/nwc-error.log"),
				MaxAge:     1,
				MaxBackups: 2,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	logger.AddHook(fileLoggerHook)

	var db *gorm.DB
	var sqlDb *sql.DB
	db, err = gorm.Open(sqlite.Open(appConfig.DatabaseUri), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	// Enable foreign keys for sqlite
	db.Exec("PRAGMA foreign_keys=ON;")
	sqlDb, err = db.DB()
	if err != nil {
		return nil, err
	}
	sqlDb.SetMaxOpenConns(1)

	err = migrations.Migrate(db)
	if err != nil {
		logger.Errorf("Failed to migrate: %v", err)
		return nil, err
	}

	ctx, _ = signal.NotifyContext(ctx, os.Interrupt)

	cfg := &Config{}
	cfg.Init(db, appConfig, logger)

	var wg sync.WaitGroup
	svc := &Service{
		cfg:    cfg,
		db:     db,
		ctx:    ctx,
		wg:     &wg,
		Logger: logger,
	}

	return svc, nil
}

func (svc *Service) launchLNBackend(encryptionKey string) error {
	if svc.lnClient != nil {
		err := svc.lnClient.Shutdown()
		if err != nil {
			return err
		}
		svc.lnClient = nil
	}

	lndBackend, _ := svc.cfg.Get("LNBackendType", "")
	if lndBackend == "" {
		return errors.New("No LNBackendType specified")
	}

	svc.Logger.Infof("Launching LN Backend: %s", lndBackend)
	var lnClient LNClient
	var err error
	switch lndBackend {
	case LNDBackendType:
		LNDAddress, _ := svc.cfg.Get("LNDAddress", encryptionKey)
		LNDCertHex, _ := svc.cfg.Get("LNDCertHex", encryptionKey)
		LNDMacaroonHex, _ := svc.cfg.Get("LNDMacaroonHex", encryptionKey)
		lnClient, err = NewLNDService(svc, LNDAddress, LNDCertHex, LNDMacaroonHex)
	case lndBackend:
		BreezMnemonic, _ := svc.cfg.Get("BreezMnemonic", encryptionKey)
		BreezAPIKey, _ := svc.cfg.Get("BreezAPIKey", encryptionKey)
		GreenlightInviteCode, _ := svc.cfg.Get("GreenlightInviteCode", encryptionKey)
		BreezWorkdir := path.Join(svc.cfg.Env.Workdir, "breez")

		lnClient, err = NewBreezService(BreezMnemonic, BreezAPIKey, GreenlightInviteCode, BreezWorkdir)
	default:
		svc.Logger.Fatalf("Unsupported LNBackendType: %v", lndBackend)
	}
	if err != nil {
		svc.Logger.Errorf("Failed to launch LN backend: %v", err)
		return err
	}
	svc.lnClient = lnClient
	return nil
}

func (svc *Service) createFilters(identityPubkey string) nostr.Filters {
	filter := nostr.Filter{
		Tags:  nostr.TagMap{"p": []string{identityPubkey}},
		Kinds: []int{NIP_47_REQUEST_KIND},
	}
	return []nostr.Filter{filter}
}

func (svc *Service) noticeHandler(notice string) {
	svc.Logger.Infof("Received a notice %s", notice)
}

func (svc *Service) StartSubscription(ctx context.Context, sub *nostr.Subscription) error {
	go func() {
		// block till EOS is received
		<-sub.EndOfStoredEvents
		svc.Logger.Info("Received EOS")

		// loop through incoming events
		for event := range sub.Events {
			go svc.HandleEvent(ctx, sub, event)
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

func (svc *Service) PublishEvent(ctx context.Context, sub *nostr.Subscription, requestEvent *RequestEvent, resp *nostr.Event, app *App, ss []byte) error {
	var appId *uint
	if app != nil {
		appId = &app.ID
	}
	responseEvent := ResponseEvent{NostrId: resp.ID, RequestId: requestEvent.ID, Content: resp.Content, State: "received"}
	err := svc.db.Create(&responseEvent).Error
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":      requestEvent.NostrId,
			"appId":        appId,
			"replyEventId": resp.ID,
		}).Errorf("Failed to save response/reply event: %v", err)
		return err
	}

	status, err := sub.Relay.Publish(ctx, *resp)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":      requestEvent.NostrId,
			"status":       status,
			"appId":        appId,
			"replyEventId": resp.ID,
		}).Errorf("Failed to publish reply: %v", err)
		return err
	}

	if status == nostr.PublishStatusSucceeded {
		responseEvent.State = RESPONSE_EVENT_STATE_PUBLISH_CONFIRMED
		responseEvent.RepliedAt = time.Now()
		svc.Logger.WithFields(logrus.Fields{
			"nostrEventId": requestEvent.ID,
			"eventId":      requestEvent.NostrId,
			"status":       status,
			"appId":        appId,
			"replyEventId": resp.ID,
		}).Info("Published reply")
	} else if status == nostr.PublishStatusFailed {
		responseEvent.State = RESPONSE_EVENT_STATE_PUBLISH_FAILED
		svc.Logger.WithFields(logrus.Fields{
			"nostrEventId": requestEvent.ID,
			"eventId":      requestEvent.NostrId,
			"status":       status,
			"appId":        appId,
			"replyEventId": resp.ID,
		}).Info("Failed to publish reply")
	} else {
		responseEvent.State = RESPONSE_EVENT_STATE_PUBLISH_UNCONFIRMED
		svc.Logger.WithFields(logrus.Fields{
			"nostrEventId": requestEvent.ID,
			"eventId":      requestEvent.NostrId,
			"status":       status,
			"appId":        appId,
			"replyEventId": resp.ID,
		}).Info("Reply sent but no response from relay (timeout)")
	}
	err = svc.db.Save(&responseEvent).Error
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":      requestEvent.NostrId,
			"appId":        appId,
			"replyEventId": resp.ID,
		}).Errorf("Failed to update response/reply event: %v", err)
	}

	return nil
}

func (svc *Service) HandleEvent(ctx context.Context, sub *nostr.Subscription, event *nostr.Event) {
	var nipResponse *Nip47Response
	svc.Logger.WithFields(logrus.Fields{
		"eventId":   event.ID,
		"eventKind": event.Kind,
	}).Info("Processing Event")

	// make sure we don't know the event, yet
	requestEvent := RequestEvent{}
	findEventResult := svc.db.Where("nostr_id = ?", event.ID).Find(&requestEvent)
	if findEventResult.RowsAffected != 0 {
		svc.Logger.WithFields(logrus.Fields{
			"eventId": event.ID,
		}).Warn("Event already processed")
		return
	}

	ss, err := nip04.ComputeSharedSecret(event.PubKey, svc.cfg.NostrSecretKey)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
		}).Errorf("Failed to compute shared secret: %v", err)
		return
	}

	// store request event
	requestEvent = RequestEvent{AppId: nil, NostrId: event.ID, Content: event.Content, State: REQUEST_EVENT_STATE_HANDLER_EXECUTING}
	err = svc.db.Create(&requestEvent).Error
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
		}).Errorf("Failed to save nostr event: %v", err)
		nipResponse = &Nip47Response{
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_INTERNAL,
				Message: fmt.Sprintf("Failed to save nostr event: %s", err.Error()),
			},
		}
		resp, err := svc.createResponse(event, nipResponse, nostr.Tags{}, ss)
		if err != nil {
			svc.Logger.WithFields(logrus.Fields{
				"eventId":   event.ID,
				"eventKind": event.Kind,
			}).Errorf("Failed to process event: %v", err)
		}
		svc.PublishEvent(ctx, sub, &requestEvent, resp, nil, ss)
		return
	}

	app := App{}
	err = svc.db.First(&app, &App{
		NostrPubkey: event.PubKey,
	}).Error
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"nostrPubkey": event.PubKey,
		}).Errorf("Failed to find app for nostr pubkey: %v", err)

		nipResponse = &Nip47Response{
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_UNAUTHORIZED,
				Message: "The public key does not have a wallet connected.",
			},
		}
		resp, err := svc.createResponse(event, nipResponse, nostr.Tags{}, ss)
		if err != nil {
			svc.Logger.WithFields(logrus.Fields{
				"eventId":   event.ID,
				"eventKind": event.Kind,
			}).Errorf("Failed to process event: %v", err)
		}
		svc.PublishEvent(ctx, sub, &requestEvent, resp, &app, ss)

		requestEvent.State = REQUEST_EVENT_STATE_HANDLER_ERROR
		err = svc.db.Save(&requestEvent).Error
		if err != nil {
			svc.Logger.WithFields(logrus.Fields{
				"nostrPubkey": event.PubKey,
			}).Errorf("Failed to save state to nostr event: %v", err)
		}
		return
	}

	requestEvent.AppId = &app.ID
	err = svc.db.Save(&requestEvent).Error
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"nostrPubkey": event.PubKey,
		}).Errorf("Failed to save app to nostr event: %v", err)

		nipResponse = &Nip47Response{
			Error: &Nip47Error{
				Code:    NIP_47_ERROR_UNAUTHORIZED,
				Message: fmt.Sprintf("Failed to save app to nostr event: %s", err.Error()),
			},
		}
		resp, err := svc.createResponse(event, nipResponse, nostr.Tags{}, ss)
		if err != nil {
			svc.Logger.WithFields(logrus.Fields{
				"eventId":   event.ID,
				"eventKind": event.Kind,
			}).Errorf("Failed to process event: %v", err)
		}
		svc.PublishEvent(ctx, sub, &requestEvent, resp, &app, ss)

		requestEvent.State = REQUEST_EVENT_STATE_HANDLER_ERROR
		err = svc.db.Save(&requestEvent).Error
		if err != nil {
			svc.Logger.WithFields(logrus.Fields{
				"nostrPubkey": event.PubKey,
			}).Errorf("Failed to save state to nostr event: %v", err)
		}

		return
	}

	svc.Logger.WithFields(logrus.Fields{
		"eventId":   event.ID,
		"eventKind": event.Kind,
		"appId":     app.ID,
	}).Info("App found for nostr event")

	//to be extra safe, decrypt using the key found from the app
	ss, err = nip04.ComputeSharedSecret(app.NostrPubkey, svc.cfg.NostrSecretKey)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
		}).Errorf("Failed to process event: %v", err)

		requestEvent.State = REQUEST_EVENT_STATE_HANDLER_ERROR
		err = svc.db.Save(&requestEvent).Error
		if err != nil {
			svc.Logger.WithFields(logrus.Fields{
				"nostrPubkey": event.PubKey,
			}).Errorf("Failed to save state to nostr event: %v", err)
		}

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

		requestEvent.State = REQUEST_EVENT_STATE_HANDLER_ERROR
		err = svc.db.Save(&requestEvent).Error
		if err != nil {
			svc.Logger.WithFields(logrus.Fields{
				"nostrPubkey": event.PubKey,
			}).Errorf("Failed to save state to nostr event: %v", err)
		}

		return
	}
	nip47Request := &Nip47Request{}
	err = json.Unmarshal([]byte(payload), nip47Request)
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
		}).Errorf("Failed to process event: %v", err)

		requestEvent.State = REQUEST_EVENT_STATE_HANDLER_ERROR
		err = svc.db.Save(&requestEvent).Error
		if err != nil {
			svc.Logger.WithFields(logrus.Fields{
				"nostrPubkey": event.PubKey,
			}).Errorf("Failed to save state to nostr event: %v", err)
		}

		return
	}

	// TODO: replace with a channel
	// TODO: update all previous occurences of svc.PublishEvent to also use the channel
	publishResponse := func(nipResponse *Nip47Response, tags nostr.Tags) {
		resp, err := svc.createResponse(event, *nipResponse, tags, ss)
		if err != nil {
			svc.Logger.WithFields(logrus.Fields{
				"eventId":   event.ID,
				"eventKind": event.Kind,
			}).Errorf("Failed to create response: %v", err)
		}
		if resp != nil {
			err = svc.PublishEvent(ctx, sub, &requestEvent, resp, &app, ss)

			if err != nil {
				svc.Logger.WithFields(logrus.Fields{
					"eventId":   event.ID,
					"eventKind": event.Kind,
				}).Errorf("Failed to publish event: %v", err)
			}
		}
	}

	// TODO: consider move event publishing to individual methods - multi_* methods are
	// inconsistent with single methods because they publish multiple responses
	switch nip47Request.Method {
	case NIP_47_MULTI_PAY_INVOICE_METHOD:
		err = svc.HandleMultiPayInvoiceEvent(ctx, nip47Request, &requestEvent, &app, publishResponse)
	case NIP_47_MULTI_PAY_KEYSEND_METHOD:
		err = svc.HandleMultiPayKeysendEvent(ctx, nip47Request, &requestEvent, &app, publishResponse)
	case NIP_47_PAY_INVOICE_METHOD:
		nipResponse, err = svc.HandlePayInvoiceEvent(ctx, nip47Request, &requestEvent, &app)
	case NIP_47_PAY_KEYSEND_METHOD:
		nipResponse, err = svc.HandlePayKeysendEvent(ctx, nip47Request, &requestEvent, &app)
	// TODO: should we add a common check for permissions here
	// after payment methods? (as they need decoded amounts)
	case NIP_47_GET_BALANCE_METHOD:
		nipResponse, err = svc.HandleGetBalanceEvent(ctx, nip47Request, &requestEvent, &app)
	case NIP_47_MAKE_INVOICE_METHOD:
		nipResponse, err = svc.HandleMakeInvoiceEvent(ctx, nip47Request, &requestEvent, &app)
	case NIP_47_LOOKUP_INVOICE_METHOD:
		nipResponse, err = svc.HandleLookupInvoiceEvent(ctx, nip47Request, &requestEvent, &app)
	case NIP_47_LIST_TRANSACTIONS_METHOD:
		nipResponse, err = svc.HandleListTransactionsEvent(ctx, nip47Request, &requestEvent, &app)
	case NIP_47_GET_INFO_METHOD:
		nipResponse, err = svc.HandleGetInfoEvent(ctx, nip47Request, &requestEvent, &app)
	default:
		nipResponse, err = svc.handleUnknownMethod(ctx, nip47Request)
	}

	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"eventId":   event.ID,
			"eventKind": event.Kind,
		}).Errorf("Failed to handle event: %v", err)
	}

	if nipResponse != nil {
		publishResponse(nipResponse, nostr.Tags{})
	}

	requestEvent.State = REQUEST_EVENT_STATE_HANDLER_EXECUTED
	err = svc.db.Save(&requestEvent).Error
	if err != nil {
		svc.Logger.WithFields(logrus.Fields{
			"nostrPubkey": event.PubKey,
		}).Errorf("Failed to save state to nostr event: %v", err)
	}
}

func (svc *Service) handleUnknownMethod(ctx context.Context, request *Nip47Request) (result *Nip47Response, err error) {
	return &Nip47Response{
		ResultType: request.Method,
		Error: &Nip47Error{
			Code:    NIP_47_ERROR_NOT_IMPLEMENTED,
			Message: fmt.Sprintf("Unknown method: %s", request.Method),
		}}, nil
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
		PubKey:    svc.cfg.NostrPublicKey,
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
	svc.db.Find(&appPermissions, &AppPermission{
		AppId: app.ID,
	})
	requestMethods := make([]string, 0, len(appPermissions))
	for _, appPermission := range appPermissions {
		requestMethods = append(requestMethods, appPermission.RequestMethod)
	}
	return requestMethods
}

func (svc *Service) hasPermission(app *App, requestMethod string, amount int64) (result bool, code string, message string) {
	appPermission := AppPermission{}
	findPermissionResult := svc.db.Find(&appPermission, &AppPermission{
		AppId:         app.ID,
		RequestMethod: requestMethod,
	})
	if findPermissionResult.RowsAffected == 0 {
		// No permission for this request method
		return false, NIP_47_ERROR_RESTRICTED, fmt.Sprintf("This app does not have permission to request %s", requestMethod)
	}
	expiresAt := appPermission.ExpiresAt
	if !expiresAt.IsZero() && expiresAt.Before(time.Now()) {
		svc.Logger.WithFields(logrus.Fields{
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
	ev.PubKey = svc.cfg.NostrPublicKey
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
