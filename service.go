package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/adrg/xdg"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"

	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/orandin/lumberjackrus"
	"gorm.io/gorm"

	alby "github.com/getAlby/nostr-wallet-connect/alby"
	"github.com/getAlby/nostr-wallet-connect/events"
	"github.com/getAlby/nostr-wallet-connect/utils"

	"github.com/getAlby/nostr-wallet-connect/config"
	"github.com/getAlby/nostr-wallet-connect/db"
	"github.com/getAlby/nostr-wallet-connect/lnclient"
	"github.com/getAlby/nostr-wallet-connect/lnclient/breez"
	"github.com/getAlby/nostr-wallet-connect/lnclient/cashu"
	"github.com/getAlby/nostr-wallet-connect/lnclient/greenlight"
	"github.com/getAlby/nostr-wallet-connect/lnclient/ldk"
	"github.com/getAlby/nostr-wallet-connect/lnclient/lnd"
	"github.com/getAlby/nostr-wallet-connect/lnclient/phoenixd"
	"github.com/getAlby/nostr-wallet-connect/migrations"
	"github.com/getAlby/nostr-wallet-connect/nip47"
)

const (
	logDir      = "log"
	logFilename = "nwc.log"
)

// TODO: move to service/
// TODO: do not expose service struct
type Service struct {
	// config from .GetEnv() only. Fetch dynamic config from db
	cfg                    config.Config
	db                     *gorm.DB
	lnClient               lnclient.LNClient
	logger                 *logrus.Logger
	albyOAuthSvc           alby.AlbyOAuthService
	eventPublisher         events.EventPublisher
	ctx                    context.Context
	wg                     *sync.WaitGroup
	nip47NotificationQueue nip47.Nip47NotificationQueue
	appCancelFn            context.CancelFunc
}

// TODO: move to service.go
func NewService(ctx context.Context) (*Service, error) {
	// Load config from environment variables / .GetEnv() file
	godotenv.Load(".env")
	appConfig := &config.AppConfig{}
	err := envconfig.Process("", appConfig)
	if err != nil {
		return nil, err
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logLevel, err := strconv.Atoi(appConfig.LogLevel)
	if err != nil {
		logLevel = int(logrus.InfoLevel)
	}
	logger.SetLevel(logrus.Level(logLevel))

	if appConfig.Workdir == "" {
		appConfig.Workdir = filepath.Join(xdg.DataHome, "/alby-nwc")
		logger.WithField("workdir", appConfig.Workdir).Info("No workdir specified, using default")
	}
	// make sure workdir exists
	os.MkdirAll(appConfig.Workdir, os.ModePerm)

	fileLoggerHook, err := lumberjackrus.NewHook(
		&lumberjackrus.LogFile{
			Filename:   filepath.Join(appConfig.Workdir, logDir, logFilename),
			MaxAge:     3,
			MaxBackups: 3,
		},
		logrus.InfoLevel,
		&logrus.JSONFormatter{},
		nil,
	)
	if err != nil {
		return nil, err
	}
	logger.AddHook(fileLoggerHook)

	finishRestoreNode(logger, appConfig.Workdir)

	// If DATABASE_URI is a URI or a path, leave it unchanged.
	// If it only contains a filename, prepend the workdir.
	if !strings.HasPrefix(appConfig.DatabaseUri, "file:") {
		databasePath, _ := filepath.Split(appConfig.DatabaseUri)
		if databasePath == "" {
			appConfig.DatabaseUri = filepath.Join(appConfig.Workdir, appConfig.DatabaseUri)
		}
	}

	var gormDB *gorm.DB
	var sqlDb *sql.DB
	gormDB, err = gorm.Open(sqlite.Open(appConfig.DatabaseUri), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	// Enable foreign keys for sqlite
	gormDB.Exec("PRAGMA foreign_keys=ON;")
	sqlDb, err = gormDB.DB()
	if err != nil {
		return nil, err
	}
	sqlDb.SetMaxOpenConns(1)

	err = migrations.Migrate(gormDB, appConfig, logger)
	if err != nil {
		logger.WithError(err).Error("Failed to migrate")
		return nil, err
	}

	cfg := config.NewConfig(gormDB, appConfig, logger)

	eventPublisher := events.NewEventPublisher(logger)

	if err != nil {
		logger.WithError(err).Error("Failed to create Alby OAuth service")
		return nil, err
	}

	nip47NotificationQueue := nip47.NewNip47NotificationQueue(logger)
	eventPublisher.RegisterSubscriber(nip47NotificationQueue)

	var wg sync.WaitGroup
	svc := &Service{
		cfg:                    cfg,
		db:                     gormDB,
		ctx:                    ctx,
		wg:                     &wg,
		logger:                 logger,
		eventPublisher:         eventPublisher,
		nip47NotificationQueue: nip47NotificationQueue,
		albyOAuthSvc:           alby.NewAlbyOAuthService(logger, cfg, cfg.GetEnv(), db.NewDBService(gormDB, logger)),
	}

	eventPublisher.RegisterSubscriber(svc.albyOAuthSvc)

	eventPublisher.Publish(&events.Event{
		Event: "nwc_started",
	})

	if appConfig.GoProfilerAddr != "" {
		startProfiler(ctx, appConfig.GoProfilerAddr)
	}

	if appConfig.DdProfilerEnabled {
		startDataDogProfiler(ctx)
	}

	return svc, nil
}

func (svc *Service) StopLNClient() error {
	if svc.lnClient != nil {
		svc.logger.Info("Shutting down LDK client")
		err := svc.lnClient.Shutdown()
		if err != nil {
			svc.logger.WithError(err).Error("Failed to stop LN backend")
			svc.eventPublisher.Publish(&events.Event{
				Event: "nwc_node_stop_failed",
				Properties: map[string]interface{}{
					"error": fmt.Sprintf("%v", err),
				},
			})
			return err
		}
		svc.logger.Info("Publishing node shutdown event")
		svc.lnClient = nil
		svc.eventPublisher.Publish(&events.Event{
			Event: "nwc_node_stopped",
		})
	}
	svc.logger.Info("LNClient stopped successfully")
	return nil
}

func (svc *Service) launchLNBackend(ctx context.Context, encryptionKey string) error {
	err := svc.StopLNClient()
	if err != nil {
		return err
	}

	lnBackend, _ := svc.cfg.Get("LNBackendType", "")
	if lnBackend == "" {
		return errors.New("no LNBackendType specified")
	}

	svc.logger.Infof("Launching LN Backend: %s", lnBackend)
	var lnClient lnclient.LNClient
	switch lnBackend {
	case config.LNDBackendType:
		LNDAddress, _ := svc.cfg.Get("LNDAddress", encryptionKey)
		LNDCertHex, _ := svc.cfg.Get("LNDCertHex", encryptionKey)
		LNDMacaroonHex, _ := svc.cfg.Get("LNDMacaroonHex", encryptionKey)
		lnClient, err = lnd.NewLNDService(ctx, svc.logger, LNDAddress, LNDCertHex, LNDMacaroonHex)
	case config.LDKBackendType:
		Mnemonic, _ := svc.cfg.Get("Mnemonic", encryptionKey)
		LDKWorkdir := path.Join(svc.cfg.GetEnv().Workdir, "ldk")

		lnClient, err = ldk.NewLDKService(ctx, svc.logger, svc.cfg, svc.eventPublisher, Mnemonic, LDKWorkdir, svc.cfg.GetEnv().LDKNetwork, svc.cfg.GetEnv().LDKEsploraServer, svc.cfg.GetEnv().LDKGossipSource)
	case config.GreenlightBackendType:
		Mnemonic, _ := svc.cfg.Get("Mnemonic", encryptionKey)
		GreenlightInviteCode, _ := svc.cfg.Get("GreenlightInviteCode", encryptionKey)
		GreenlightWorkdir := path.Join(svc.cfg.GetEnv().Workdir, "greenlight")

		lnClient, err = greenlight.NewGreenlightService(svc.cfg, svc.logger, Mnemonic, GreenlightInviteCode, GreenlightWorkdir, encryptionKey)
	case config.BreezBackendType:
		Mnemonic, _ := svc.cfg.Get("Mnemonic", encryptionKey)
		BreezAPIKey, _ := svc.cfg.Get("BreezAPIKey", encryptionKey)
		GreenlightInviteCode, _ := svc.cfg.Get("GreenlightInviteCode", encryptionKey)
		BreezWorkdir := path.Join(svc.cfg.GetEnv().Workdir, "breez")

		lnClient, err = breez.NewBreezService(svc.logger, Mnemonic, BreezAPIKey, GreenlightInviteCode, BreezWorkdir)
	case config.PhoenixBackendType:
		PhoenixdAddress, _ := svc.cfg.Get("PhoenixdAddress", encryptionKey)
		PhoenixdAuthorization, _ := svc.cfg.Get("PhoenixdAuthorization", encryptionKey)

		lnClient, err = phoenixd.NewPhoenixService(svc.logger, PhoenixdAddress, PhoenixdAuthorization)
	case config.CashuBackendType:
		cashuMintUrl, _ := svc.cfg.Get("CashuMintUrl", encryptionKey)
		cashuWorkdir := path.Join(svc.cfg.GetEnv().Workdir, "cashu")

		lnClient, err = cashu.NewCashuService(svc.logger, cashuWorkdir, cashuMintUrl)
	default:
		svc.logger.Fatalf("Unsupported LNBackendType: %v", lnBackend)
	}
	if err != nil {
		svc.logger.WithError(err).Error("Failed to launch LN backend")
		return err
	}

	info, err := lnClient.GetInfo(ctx)
	if err != nil {
		svc.logger.WithError(err).Error("Failed to fetch node info")
	}
	if info != nil && info.Pubkey != "" {
		svc.eventPublisher.SetGlobalProperty("node_id", info.Pubkey)
	}

	svc.eventPublisher.Publish(&events.Event{
		Event: "nwc_node_started",
		Properties: map[string]interface{}{
			"node_type": lnBackend,
		},
	})
	svc.lnClient = lnClient
	return nil
}

func (svc *Service) createFilters(identityPubkey string) nostr.Filters {
	filter := nostr.Filter{
		Tags:  nostr.TagMap{"p": []string{identityPubkey}},
		Kinds: []int{nip47.REQUEST_KIND},
	}
	return []nostr.Filter{filter}
}

func (svc *Service) noticeHandler(notice string) {
	svc.logger.Infof("Received a notice %s", notice)
}

func (svc *Service) GetLNClient() lnclient.LNClient {
	return svc.lnClient
}

func (svc *Service) StartSubscription(ctx context.Context, sub *nostr.Subscription) error {
	nip47Notifier := NewNip47Notifier(svc, sub.Relay)
	go func() {
		for {
			select {
			case <-ctx.Done():
				// subscription ended
				return
			case event := <-svc.nip47NotificationQueue.Channel():
				nip47Notifier.ConsumeEvent(ctx, event)
			}
		}
	}()

	go func() {
		// block till EOS is received
		<-sub.EndOfStoredEvents
		svc.logger.Info("Received EOS")

		// loop through incoming events
		for event := range sub.Events {
			go svc.HandleEvent(ctx, sub, event)
		}
		svc.logger.Info("Relay subscription events channel ended")
	}()

	<-ctx.Done()

	if sub.Relay.ConnectionError != nil {
		svc.logger.WithField("connectionError", sub.Relay.ConnectionError).Error("Relay error")
		return sub.Relay.ConnectionError
	}
	svc.logger.Info("Exiting subscription...")
	return nil
}

func (svc *Service) PublishEvent(ctx context.Context, sub *nostr.Subscription, requestEvent *db.RequestEvent, resp *nostr.Event, app *db.App) error {
	var appId *uint
	if app != nil {
		appId = &app.ID
	}
	responseEvent := db.ResponseEvent{NostrId: resp.ID, RequestId: requestEvent.ID, State: "received"}
	err := svc.db.Create(&responseEvent).Error
	if err != nil {
		svc.logger.WithFields(logrus.Fields{
			"requestEventNostrId": requestEvent.NostrId,
			"appId":               appId,
			"replyEventId":        resp.ID,
		}).Errorf("Failed to save response/reply event: %v", err)
		return err
	}

	err = sub.Relay.Publish(ctx, *resp)
	if err != nil {
		responseEvent.State = db.RESPONSE_EVENT_STATE_PUBLISH_FAILED
		svc.logger.WithFields(logrus.Fields{
			"requestEventId":       requestEvent.ID,
			"requestNostrEventId":  requestEvent.NostrId,
			"appId":                appId,
			"responseEventId":      responseEvent.ID,
			"responseNostrEventId": resp.ID,
		}).Errorf("Failed to publish reply: %v", err)
	} else {
		responseEvent.State = db.RESPONSE_EVENT_STATE_PUBLISH_CONFIRMED
		responseEvent.RepliedAt = time.Now()
		svc.logger.WithFields(logrus.Fields{
			"requestEventId":       requestEvent.ID,
			"requestNostrEventId":  requestEvent.NostrId,
			"appId":                appId,
			"responseEventId":      responseEvent.ID,
			"responseNostrEventId": resp.ID,
		}).Info("Published reply")
	}

	err = svc.db.Save(&responseEvent).Error
	if err != nil {
		svc.logger.WithFields(logrus.Fields{
			"requestEventId":       requestEvent.ID,
			"requestNostrEventId":  requestEvent.NostrId,
			"appId":                appId,
			"responseEventId":      responseEvent.ID,
			"responseNostrEventId": resp.ID,
		}).Errorf("Failed to update response/reply event: %v", err)
		return err
	}

	return nil
}

func (svc *Service) HandleEvent(ctx context.Context, sub *nostr.Subscription, event *nostr.Event) {
	var nip47Response *nip47.Response
	svc.logger.WithFields(logrus.Fields{
		"requestEventNostrId": event.ID,
		"eventKind":           event.Kind,
	}).Info("Processing Event")

	ss, err := nip04.ComputeSharedSecret(event.PubKey, svc.cfg.GetNostrSecretKey())
	if err != nil {
		svc.logger.WithFields(logrus.Fields{
			"requestEventNostrId": event.ID,
			"eventKind":           event.Kind,
		}).Errorf("Failed to compute shared secret: %v", err)
		return
	}

	// store request event
	requestEvent := db.RequestEvent{AppId: nil, NostrId: event.ID, State: db.REQUEST_EVENT_STATE_HANDLER_EXECUTING}
	err = svc.db.Create(&requestEvent).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			svc.logger.WithFields(logrus.Fields{
				"requestEventNostrId": event.ID,
			}).Warn("Event already processed")
			return
		}
		svc.logger.WithFields(logrus.Fields{
			"requestEventNostrId": event.ID,
			"eventKind":           event.Kind,
		}).Errorf("Failed to save nostr event: %v", err)
		nip47Response = &nip47.Response{
			Error: &nip47.Error{
				Code:    nip47.ERROR_INTERNAL,
				Message: fmt.Sprintf("Failed to save nostr event: %s", err.Error()),
			},
		}
		resp, err := svc.createResponse(event, nip47Response, nostr.Tags{}, ss)
		if err != nil {
			svc.logger.WithFields(logrus.Fields{
				"requestEventNostrId": event.ID,
				"eventKind":           event.Kind,
			}).Errorf("Failed to process event: %v", err)
		}
		svc.PublishEvent(ctx, sub, &requestEvent, resp, nil)
		return
	}

	app := db.App{}
	err = svc.db.First(&app, &db.App{
		NostrPubkey: event.PubKey,
	}).Error
	if err != nil {
		svc.logger.WithFields(logrus.Fields{
			"nostrPubkey": event.PubKey,
		}).Errorf("Failed to find app for nostr pubkey: %v", err)

		nip47Response = &nip47.Response{
			Error: &nip47.Error{
				Code:    nip47.ERROR_UNAUTHORIZED,
				Message: "The public key does not have a wallet connected.",
			},
		}
		resp, err := svc.createResponse(event, nip47Response, nostr.Tags{}, ss)
		if err != nil {
			svc.logger.WithFields(logrus.Fields{
				"requestEventNostrId": event.ID,
				"eventKind":           event.Kind,
			}).Errorf("Failed to process event: %v", err)
		}
		svc.PublishEvent(ctx, sub, &requestEvent, resp, &app)

		requestEvent.State = db.REQUEST_EVENT_STATE_HANDLER_ERROR
		err = svc.db.Save(&requestEvent).Error
		if err != nil {
			svc.logger.WithFields(logrus.Fields{
				"nostrPubkey": event.PubKey,
			}).Errorf("Failed to save state to nostr event: %v", err)
		}
		return
	}

	requestEvent.AppId = &app.ID
	err = svc.db.Save(&requestEvent).Error
	if err != nil {
		svc.logger.WithFields(logrus.Fields{
			"nostrPubkey": event.PubKey,
		}).Errorf("Failed to save app to nostr event: %v", err)

		nip47Response = &nip47.Response{
			Error: &nip47.Error{
				Code:    nip47.ERROR_UNAUTHORIZED,
				Message: fmt.Sprintf("Failed to save app to nostr event: %s", err.Error()),
			},
		}
		resp, err := svc.createResponse(event, nip47Response, nostr.Tags{}, ss)
		if err != nil {
			svc.logger.WithFields(logrus.Fields{
				"requestEventNostrId": event.ID,
				"eventKind":           event.Kind,
			}).Errorf("Failed to process event: %v", err)
		}
		svc.PublishEvent(ctx, sub, &requestEvent, resp, &app)

		requestEvent.State = db.REQUEST_EVENT_STATE_HANDLER_ERROR
		err = svc.db.Save(&requestEvent).Error
		if err != nil {
			svc.logger.WithFields(logrus.Fields{
				"nostrPubkey": event.PubKey,
			}).Errorf("Failed to save state to nostr event: %v", err)
		}

		return
	}

	svc.logger.WithFields(logrus.Fields{
		"requestEventNostrId": event.ID,
		"eventKind":           event.Kind,
		"appId":               app.ID,
	}).Info("App found for nostr event")

	//to be extra safe, decrypt using the key found from the app
	ss, err = nip04.ComputeSharedSecret(app.NostrPubkey, svc.cfg.GetNostrSecretKey())
	if err != nil {
		svc.logger.WithFields(logrus.Fields{
			"requestEventNostrId": event.ID,
			"eventKind":           event.Kind,
		}).Errorf("Failed to process event: %v", err)

		requestEvent.State = db.REQUEST_EVENT_STATE_HANDLER_ERROR
		err = svc.db.Save(&requestEvent).Error
		if err != nil {
			svc.logger.WithFields(logrus.Fields{
				"nostrPubkey": event.PubKey,
			}).Errorf("Failed to save state to nostr event: %v", err)
		}

		return
	}
	payload, err := nip04.Decrypt(event.Content, ss)
	if err != nil {
		svc.logger.WithFields(logrus.Fields{
			"requestEventNostrId": event.ID,
			"eventKind":           event.Kind,
			"appId":               app.ID,
		}).Errorf("Failed to decrypt content: %v", err)
		svc.logger.WithFields(logrus.Fields{
			"requestEventNostrId": event.ID,
			"eventKind":           event.Kind,
		}).Errorf("Failed to process event: %v", err)

		requestEvent.State = db.REQUEST_EVENT_STATE_HANDLER_ERROR
		err = svc.db.Save(&requestEvent).Error
		if err != nil {
			svc.logger.WithFields(logrus.Fields{
				"nostrPubkey": event.PubKey,
			}).Errorf("Failed to save state to nostr event: %v", err)
		}

		return
	}
	nip47Request := &nip47.Request{}
	err = json.Unmarshal([]byte(payload), nip47Request)
	if err != nil {
		svc.logger.WithFields(logrus.Fields{
			"requestEventNostrId": event.ID,
			"eventKind":           event.Kind,
		}).Errorf("Failed to process event: %v", err)

		requestEvent.State = db.REQUEST_EVENT_STATE_HANDLER_ERROR
		err = svc.db.Save(&requestEvent).Error
		if err != nil {
			svc.logger.WithFields(logrus.Fields{
				"nostrPubkey": event.PubKey,
			}).Errorf("Failed to save state to nostr event: %v", err)
		}

		return
	}

	requestEvent.Method = nip47Request.Method
	requestEvent.ContentData = payload
	svc.db.Save(&requestEvent) // we ignore potential DB errors here as this only saves the method and content data

	// TODO: replace with a channel
	// TODO: update all previous occurences of svc.PublishEvent to also use the channel
	publishResponse := func(nip47Response *nip47.Response, tags nostr.Tags) {
		resp, err := svc.createResponse(event, nip47Response, tags, ss)
		if err != nil {
			svc.logger.WithFields(logrus.Fields{
				"requestEventNostrId": event.ID,
				"eventKind":           event.Kind,
			}).Errorf("Failed to create response: %v", err)
			requestEvent.State = db.REQUEST_EVENT_STATE_HANDLER_ERROR
		} else {
			err = svc.PublishEvent(ctx, sub, &requestEvent, resp, &app)
			if err != nil {
				svc.logger.WithFields(logrus.Fields{
					"requestEventNostrId": event.ID,
					"eventKind":           event.Kind,
				}).Errorf("Failed to publish event: %v", err)
				requestEvent.State = db.REQUEST_EVENT_STATE_HANDLER_ERROR
			} else {
				requestEvent.State = db.REQUEST_EVENT_STATE_HANDLER_EXECUTED
			}
		}
		err = svc.db.Save(&requestEvent).Error
		if err != nil {
			svc.logger.WithFields(logrus.Fields{
				"nostrPubkey": event.PubKey,
			}).Errorf("Failed to save state to nostr event: %v", err)
		}
	}

	switch nip47Request.Method {
	case nip47.MULTI_PAY_INVOICE_METHOD:
		svc.HandleMultiPayInvoiceEvent(ctx, nip47Request, &requestEvent, &app, publishResponse)
	case nip47.MULTI_PAY_KEYSEND_METHOD:
		svc.HandleMultiPayKeysendEvent(ctx, nip47Request, &requestEvent, &app, publishResponse)
	case nip47.PAY_INVOICE_METHOD:
		svc.HandlePayInvoiceEvent(ctx, nip47Request, &requestEvent, &app, publishResponse)
	case nip47.PAY_KEYSEND_METHOD:
		svc.HandlePayKeysendEvent(ctx, nip47Request, &requestEvent, &app, publishResponse)
	case nip47.GET_BALANCE_METHOD:
		svc.HandleGetBalanceEvent(ctx, nip47Request, &requestEvent, &app, publishResponse)
	case nip47.MAKE_INVOICE_METHOD:
		svc.HandleMakeInvoiceEvent(ctx, nip47Request, &requestEvent, &app, publishResponse)
	case nip47.LOOKUP_INVOICE_METHOD:
		svc.HandleLookupInvoiceEvent(ctx, nip47Request, &requestEvent, &app, publishResponse)
	case nip47.LIST_TRANSACTIONS_METHOD:
		svc.HandleListTransactionsEvent(ctx, nip47Request, &requestEvent, &app, publishResponse)
	case nip47.GET_INFO_METHOD:
		svc.HandleGetInfoEvent(ctx, nip47Request, &requestEvent, &app, publishResponse)
	case nip47.SIGN_MESSAGE_METHOD:
		svc.HandleSignMessageEvent(ctx, nip47Request, &requestEvent, &app, publishResponse)
	default:
		svc.handleUnknownMethod(ctx, nip47Request, publishResponse)
	}
}

func (svc *Service) handleUnknownMethod(ctx context.Context, nip47Request *nip47.Request, publishResponse func(*nip47.Response, nostr.Tags)) {
	publishResponse(&nip47.Response{
		ResultType: nip47Request.Method,
		Error: &nip47.Error{
			Code:    nip47.ERROR_NOT_IMPLEMENTED,
			Message: fmt.Sprintf("Unknown method: %s", nip47Request.Method),
		},
	}, nostr.Tags{})
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
		PubKey:    svc.cfg.GetNostrPublicKey(),
		CreatedAt: nostr.Now(),
		Kind:      nip47.RESPONSE_KIND,
		Tags:      allTags,
		Content:   msg,
	}
	err = resp.Sign(svc.cfg.GetNostrSecretKey())
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (svc *Service) GetMethods(app *db.App) []string {
	appPermissions := []db.AppPermission{}
	svc.db.Find(&appPermissions, &db.AppPermission{
		AppId: app.ID,
	})
	requestMethods := make([]string, 0, len(appPermissions))
	for _, appPermission := range appPermissions {
		requestMethods = append(requestMethods, appPermission.RequestMethod)
	}
	if slices.Contains(requestMethods, nip47.PAY_INVOICE_METHOD) {
		// all payment methods are tied to the pay_invoice permission
		requestMethods = append(requestMethods, nip47.PAY_KEYSEND_METHOD, nip47.MULTI_PAY_INVOICE_METHOD, nip47.MULTI_PAY_KEYSEND_METHOD)
	}

	return requestMethods
}

func (svc *Service) decodeNip47Request(nip47Request *nip47.Request, requestEvent *db.RequestEvent, app *db.App, methodParams interface{}) *nip47.Response {
	err := json.Unmarshal(nip47Request.Params, methodParams)
	if err != nil {
		svc.logger.WithFields(logrus.Fields{
			"requestEventNostrId": requestEvent.NostrId,
			"appId":               app.ID,
		}).Errorf("Failed to decode nostr event: %v", err)
		return &nip47.Response{
			ResultType: nip47Request.Method,
			Error: &nip47.Error{
				Code:    nip47.ERROR_BAD_REQUEST,
				Message: err.Error(),
			}}
	}
	return nil
}

func (svc *Service) checkPermission(nip47Request *nip47.Request, requestNostrEventId string, app *db.App, amount int64) *nip47.Response {
	hasPermission, code, message := svc.hasPermission(app, nip47Request.Method, amount)
	if !hasPermission {
		svc.logger.WithFields(logrus.Fields{
			"requestEventNostrId": requestNostrEventId,
			"appId":               app.ID,
			"code":                code,
			"message":             message,
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

		return &nip47.Response{
			ResultType: nip47Request.Method,
			Error: &nip47.Error{
				Code:    code,
				Message: message,
			},
		}
	}
	return nil
}

func (svc *Service) hasPermission(app *db.App, requestMethod string, amount int64) (result bool, code string, message string) {
	switch requestMethod {
	case nip47.PAY_INVOICE_METHOD, nip47.PAY_KEYSEND_METHOD, nip47.MULTI_PAY_INVOICE_METHOD, nip47.MULTI_PAY_KEYSEND_METHOD:
		requestMethod = nip47.PAY_INVOICE_METHOD
	}

	appPermission := db.AppPermission{}
	findPermissionResult := svc.db.Find(&appPermission, &db.AppPermission{
		AppId:         app.ID,
		RequestMethod: requestMethod,
	})
	if findPermissionResult.RowsAffected == 0 {
		// No permission for this request method
		return false, nip47.ERROR_RESTRICTED, fmt.Sprintf("This app does not have permission to request %s", requestMethod)
	}
	expiresAt := appPermission.ExpiresAt
	if expiresAt != nil && expiresAt.Before(time.Now()) {
		svc.logger.WithFields(logrus.Fields{
			"requestMethod": requestMethod,
			"expiresAt":     expiresAt.Unix(),
			"appId":         app.ID,
			"pubkey":        app.NostrPubkey,
		}).Info("This pubkey is expired")

		return false, nip47.ERROR_EXPIRED, "This app has expired"
	}

	if requestMethod == nip47.PAY_INVOICE_METHOD {
		maxAmount := appPermission.MaxAmount
		if maxAmount != 0 {
			budgetUsage := svc.GetBudgetUsage(&appPermission)

			if budgetUsage+amount/1000 > int64(maxAmount) {
				return false, nip47.ERROR_QUOTA_EXCEEDED, "Insufficient budget remaining to make payment"
			}
		}
	}
	return true, "", ""
}

// TODO: move somewhere else
func (svc *Service) GetBudgetUsage(appPermission *db.AppPermission) int64 {
	var result struct {
		Sum uint
	}
	// TODO: discard failed payments from this check instead of checking payments that have a preimage
	svc.db.Table("payments").Select("SUM(amount) as sum").Where("app_id = ? AND preimage IS NOT NULL AND created_at > ?", appPermission.AppId, utils.GetStartOfBudget(appPermission.BudgetRenewal, appPermission.App.CreatedAt)).Scan(&result)
	return int64(result.Sum)
}

func (svc *Service) PublishNip47Info(ctx context.Context, relay *nostr.Relay) error {
	ev := &nostr.Event{}
	ev.Kind = nip47.INFO_EVENT_KIND
	ev.Content = nip47.CAPABILITIES
	ev.CreatedAt = nostr.Now()
	ev.PubKey = svc.cfg.GetNostrPublicKey()
	ev.Tags = nostr.Tags{[]string{"notifications", nip47.NOTIFICATION_TYPES}}
	err := ev.Sign(svc.cfg.GetNostrSecretKey())
	if err != nil {
		return err
	}
	err = relay.Publish(ctx, *ev)
	if err != nil {
		return fmt.Errorf("nostr publish not successful: %s", err)
	}
	return nil
}

func (svc *Service) GetLogFilePath() string {
	return filepath.Join(svc.cfg.GetEnv().Workdir, logDir, logFilename)
}

func finishRestoreNode(logger *logrus.Logger, workDir string) {
	restoreDir := filepath.Join(workDir, "restore")
	if restoreDirStat, err := os.Stat(restoreDir); err == nil && restoreDirStat.IsDir() {
		logger.WithField("restoreDir", restoreDir).Infof("Restore directory found. Finishing Node restore")

		existingFiles, err := os.ReadDir(restoreDir)
		if err != nil {
			logger.WithError(err).Fatal("Failed to read WORK_DIR")
		}

		for _, file := range existingFiles {
			if file.Name() != "restore" {
				err = os.RemoveAll(filepath.Join(workDir, file.Name()))
				if err != nil {
					logger.WithField("filename", file.Name()).WithError(err).Fatal("Failed to remove file")
				}
				logger.WithField("filename", file.Name()).Info("removed file")
			}
		}

		files, err := os.ReadDir(restoreDir)
		if err != nil {
			logger.WithError(err).Fatal("Failed to read restore directory")
		}
		for _, file := range files {
			err = os.Rename(filepath.Join(restoreDir, file.Name()), filepath.Join(workDir, file.Name()))
			if err != nil {
				logger.WithField("filename", file.Name()).WithError(err).Fatal("Failed to move file")
			}
			logger.WithField("filename", file.Name()).Info("copied file from restore directory")
		}
		err = os.RemoveAll(restoreDir)
		if err != nil {
			logger.WithError(err).Fatal("Failed to remove restore directory")
		}
		logger.WithField("restoreDir", restoreDir).Info("removed restore directory")
	}
}

func startProfiler(ctx context.Context, addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		err := server.Shutdown(context.Background())
		if err != nil {
			panic("pprof server shutdown failed: " + err.Error())
		}
	}()

	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic("pprof server failed: " + err.Error())
		}
	}()
}

func startDataDogProfiler(ctx context.Context) {
	opts := make([]profiler.Option, 0)

	opts = append(opts, profiler.WithProfileTypes(
		profiler.CPUProfile,
		profiler.HeapProfile,
		// higher overhead
		profiler.BlockProfile,
		profiler.MutexProfile,
		profiler.GoroutineProfile,
	))

	err := profiler.Start(opts...)
	if err != nil {
		panic("failed to start DataDog profiler: " + err.Error())
	}

	go func() {
		<-ctx.Done()
		profiler.Stop()
	}()
}

func (svc *Service) StopDb() error {
	db, err := svc.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	err = db.Close()
	if err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	return nil
}

func (svc *Service) GetConfig() config.Config {
	return svc.cfg
}

func (svc *Service) GetAlbyOAuthSvc() alby.AlbyOAuthService {
	return svc.albyOAuthSvc
}
