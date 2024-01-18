package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	echologrus "github.com/davrux/echo-logrus/v4"
	"github.com/getAlby/nostr-wallet-connect/migrations"
	"github.com/getAlby/nostr-wallet-connect/models/db"
	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo/v4"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	log "github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/jackc/pgx/v5/stdlib"
	sqltrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"
	gormtrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gorm.io/gorm.v1"
)

func main() {

	// Load config from environment variables / .env file
	godotenv.Load(".env")
	cfg := &Config{}
	err := envconfig.Process("", cfg)
	if err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	var db *gorm.DB
	var sqlDb *sql.DB
	if strings.HasPrefix(cfg.DatabaseUri, "postgres://") || strings.HasPrefix(cfg.DatabaseUri, "postgresql://") || strings.HasPrefix(cfg.DatabaseUri, "unix://") {
		if os.Getenv("DATADOG_AGENT_URL") != "" {
			sqltrace.Register("pgx", &stdlib.Driver{}, sqltrace.WithServiceName("nostr-wallet-connect"))
			sqlDb, err = sqltrace.Open("pgx", cfg.DatabaseUri)
			if err != nil {
				log.Fatalf("Failed to open DB %v", err)
			}
			db, err = gormtrace.Open(postgres.New(postgres.Config{Conn: sqlDb}), &gorm.Config{})
			if err != nil {
				log.Fatalf("Failed to open DB %v", err)
			}
		} else {
			db, err = gorm.Open(postgres.Open(cfg.DatabaseUri), &gorm.Config{})
			if err != nil {
				log.Fatalf("Failed to open DB %v", err)
			}
			sqlDb, err = db.DB()
			if err != nil {
				log.Fatalf("Failed to set DB config: %v", err)
			}
		}
	} else {
		db, err = gorm.Open(sqlite.Open(cfg.DatabaseUri), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to open DB %v", err)
		}
		// Override SQLite config to max one connection
		cfg.DatabaseMaxConns = 1
		// Enable foreign keys for sqlite
		db.Exec("PRAGMA foreign_keys=ON;")
		sqlDb, err = db.DB()
		if err != nil {
			log.Fatalf("Failed to set DB config: %v", err)
		}
	}
	sqlDb.SetMaxOpenConns(cfg.DatabaseMaxConns)
	sqlDb.SetMaxIdleConns(cfg.DatabaseMaxIdleConns)
	sqlDb.SetConnMaxLifetime(time.Duration(cfg.DatabaseConnMaxLifetime) * time.Second)

	err = migrations.Migrate(db)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Any pending migrations ran successfully")

	if cfg.NostrSecretKey == "" {
		if cfg.LNBackendType == AlbyBackendType {
			//not allowed
			log.Fatal("Nostr private key is required with this backend type.")
		}
		//first look up if we already have the private key in the database
		//else, generate and store private key
		identity := &Identity{}
		err = db.FirstOrInit(identity).Error
		if err != nil {
			log.WithError(err).Fatal("Error retrieving private key from database")
		}
		if identity.Privkey == "" {
			log.Info("No private key found in database, generating & saving.")
			identity.Privkey = nostr.GeneratePrivateKey()
			err = db.Save(identity).Error
			if err != nil {
				log.WithError(err).Fatal("Error saving private key to database")
			}
		}
		cfg.NostrSecretKey = identity.Privkey
	}

	identityPubkey, err := nostr.GetPublicKey(cfg.NostrSecretKey)
	if err != nil {
		log.Fatalf("Error converting nostr privkey to pubkey: %v", err)
	}
	cfg.IdentityPubkey = identityPubkey
	npub, err := nip19.EncodePublicKey(identityPubkey)
	if err != nil {
		log.Fatalf("Error converting nostr privkey to pubkey: %v", err)
	}

	log.Infof("Starting nostr-wallet-connect. npub: %s hex: %s", npub, identityPubkey)
	ctx := context.Background()
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt)

	svc := &Service{
		cfg: cfg,
		db:  db,
		ctx: ctx,
	}

	err = svc.setupDbConfig()
	if err != nil {
		log.Fatalf("Failed to setup DB config: %v", err)
	}

	// TODO: remove Datadog etc.
	if os.Getenv("DATADOG_AGENT_URL") != "" {
		tracer.Start(tracer.WithService("nostr-wallet-connect"))
		defer tracer.Stop()
	}

	logger := log.New()
	logger.SetFormatter(&log.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(log.InfoLevel)
	svc.Logger = logger

	var wg sync.WaitGroup
	wg.Add(1)
	// TODO: can we move these into two separate entrypoint files?
	switch cfg.AppType {
	case WailsAppType:
		err := svc.launchLNBackend()
		if err != nil {
			// LN backend not needed immediately, just log errors
			svc.Logger.Warnf("Failed to launch LN backend: %v", err)
		}
		go func() {
			app := NewApp(svc)
			LaunchWailsApp(app)
			wg.Done()
			svc.Logger.Info("Wails app exited")
		}()
	case HttpAppType:
		// using echo
		echologrus.Logger = logger
		e := echo.New()

		// Alby backend only supported in HTTP app type
		if svc.cfg.LNBackendType == AlbyBackendType {
			oauthService, err := NewAlbyOauthService(svc, e)
			if err != nil {
				svc.Logger.Fatal(err)
			}
			svc.lnClient = oauthService
		} else {
			err := svc.launchLNBackend()
			if err != nil {
				// LN backend not needed immediately, just log errors
				svc.Logger.Warnf("Failed to launch LN backend: %v", err)
			}
		}

		//register shared routes
		svc.RegisterSharedRoutes(e)
		//start Echo server
		go func() {
			if err := e.Start(fmt.Sprintf(":%v", svc.cfg.Port)); err != nil && err != http.ErrServerClosed {
				e.Logger.Fatal("shutting down the server")
			}
			//handle graceful shutdown
			<-ctx.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			e.Shutdown(ctx)
			svc.Logger.Info("Echo server exited")
			wg.Done()
		}()
	default:
		svc.Logger.Fatalf("Unknown app type: %s", cfg.AppType)
	}

	go func() {
		//connect to the relay
		svc.Logger.Infof("Connecting to the relay: %s", cfg.Relay)

		relay, err := nostr.RelayConnect(ctx, cfg.Relay, nostr.WithNoticeHandler(svc.noticeHandler))
		if err != nil {
			svc.Logger.Fatal(err)
		}

		//publish event with NIP-47 info
		err = svc.PublishNip47Info(ctx, relay)
		if err != nil {
			svc.Logger.WithError(err).Error("Could not publish NIP47 info")
		}

		//Start infinite loop which will be only broken by canceling ctx (SIGINT)
		//TODO: we can start this loop for multiple relays
		for {
			svc.Logger.Info("Subscribing to events")
			sub, err := relay.Subscribe(ctx, svc.createFilters())
			if err != nil {
				svc.Logger.Fatal(err)
			}
			err = svc.StartSubscription(ctx, sub)
			if err != nil {
				//err being non-nil means that we have an error on the websocket error channel. In this case we just try to reconnect.
				svc.Logger.WithError(err).Error("Got an error from the relay while listening to subscription. Reconnecting...")
				relay, err = nostr.RelayConnect(ctx, cfg.Relay)
				if err != nil {
					svc.Logger.Fatal(err)
				}
				continue
			}
			//err being nil means that the context was canceled and we should exit the program.
			break
		}
		err = relay.Close()
		if err != nil {
			svc.Logger.Error(err)
		}
		svc.Logger.Info("Graceful shutdown completed. Goodbye.")
		wg.Done()
	}()
	wg.Wait()
}

func (svc *Service) setupDbConfig() error {
	// setup database config from the env on first run
	// after first run, changes to ENV will have no effect for these fields
	// because the database values always take precedence!

	var existing db.Config
	res := svc.db.Limit(1).Find(&existing)

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected > 0 {
		// do not overwrite the existing entry
		return nil
	}

	newDbConfig := db.Config{
		ID:                   1,
		LNBackendType:        svc.cfg.LNBackendType,
		LNDAddress:           svc.cfg.LNDAddress,
		LNDCertFile:          svc.cfg.LNDCertFile,
		LNDCertHex:           svc.cfg.LNDCertHex,
		LNDMacaroonFile:      svc.cfg.LNDMacaroonFile,
		LNDMacaroonHex:       svc.cfg.LNDMacaroonHex,
		BreezMnemonic:        svc.cfg.BreezMnemonic,
		GreenlightInviteCode: svc.cfg.GreenlightInviteCode,
	}
	err := svc.db.Save(&newDbConfig).Error
	if err != nil {
		return err
	}

	return nil
}

func (svc *Service) launchLNBackend() error {
	if svc.lnClient != nil {
		err := svc.lnClient.Shutdown()
		if err != nil {
			svc.Logger.Fatalf("Failed to disconnect from current node: %v", err)
		}
		svc.lnClient = nil
	}

	dbConfig := db.Config{}
	svc.db.First(&dbConfig)

	if dbConfig.LNBackendType == "" {
		return errors.New("No LNBackendType specified")
	}

	svc.Logger.Infof("Launching LN Backend: %s", dbConfig.LNBackendType)
	var err error
	var lnClient LNClient
	switch dbConfig.LNBackendType {
	case LNDBackendType:
		lnClient, err = NewLNDService(svc, dbConfig.LNDAddress, dbConfig.LNDCertFile, dbConfig.LNDCertHex, dbConfig.LNDMacaroonFile, dbConfig.LNDMacaroonHex)
	case BreezBackendType:
		lnClient, err = NewBreezService(svc, dbConfig.BreezMnemonic, svc.cfg.BreezAPIKey, dbConfig.GreenlightInviteCode, svc.cfg.BreezWorkdir)
	default:
		svc.Logger.Fatalf("Unsupported LNBackendType: %v", dbConfig.LNBackendType)
	}
	if err != nil {
		return err
	}
	svc.lnClient = lnClient
	return nil
}

func (svc *Service) createFilters() nostr.Filters {
	filter := nostr.Filter{
		Tags:  nostr.TagMap{"p": []string{svc.cfg.IdentityPubkey}},
		Kinds: []int{NIP_47_REQUEST_KIND},
	}
	if svc.cfg.ClientPubkey != "" {
		filter.Authors = []string{svc.cfg.ClientPubkey}
	}
	return []nostr.Filter{filter}
}

func (svc *Service) noticeHandler(notice string) {
	svc.Logger.Infof("Received a notice %s", notice)
}
