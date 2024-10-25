package service

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strconv"
	"time"

	"github.com/getAlby/hub/db"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/sirupsen/logrus"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/lnclient/breez"
	"github.com/getAlby/hub/lnclient/cashu"
	"github.com/getAlby/hub/lnclient/greenlight"
	"github.com/getAlby/hub/lnclient/ldk"
	"github.com/getAlby/hub/lnclient/lnd"
	"github.com/getAlby/hub/lnclient/phoenixd"
	"github.com/getAlby/hub/logger"
)

func (svc *service) startNostr(ctx context.Context, encryptionKey string) error {

	relayUrl := svc.cfg.GetRelayUrl()

	err := svc.keys.Init(svc.cfg, encryptionKey)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to init nostr keys")
		return err
	}

	npub, err := nip19.EncodePublicKey(svc.keys.GetNostrPublicKey())
	if err != nil {
		logger.Logger.WithError(err).Error("Error converting nostr privkey to pubkey")
		return err
	}

	logger.Logger.WithFields(logrus.Fields{
		"npub": npub,
		"hex":  svc.keys.GetNostrPublicKey(),
	}).Info("Starting Alby Hub")
	svc.wg.Add(1)
	go func() {
		// ensure the relay is properly disconnected before exiting
		defer svc.wg.Done()
		//Start infinite loop which will be only broken by canceling ctx (SIGINT)
		var relay *nostr.Relay
		waitToReconnectSeconds := 0

		for i := 0; ; i++ {

			// wait for a delay if any before retrying
			if waitToReconnectSeconds > 0 {
				contextCancelled := false

				select {
				case <-ctx.Done(): //context cancelled
					logger.Logger.Info("service context cancelled while waiting for retry")
					contextCancelled = true
				case <-time.After(time.Duration(waitToReconnectSeconds) * time.Second): //timeout
				}
				if contextCancelled {
					break
				}
			}

			closeRelay(relay)

			//connect to the relay
			logger.Logger.WithFields(logrus.Fields{
				"relay_url": relayUrl,
				"iteration": i,
			}).Info("Connecting to the relay")

			relay, err = nostr.RelayConnect(ctx, relayUrl, nostr.WithNoticeHandler(svc.noticeHandler))
			if err != nil {
				// exponential backoff from 2 - 60 seconds
				waitToReconnectSeconds = max(waitToReconnectSeconds, 1)
				waitToReconnectSeconds *= 2
				waitToReconnectSeconds = min(waitToReconnectSeconds, 60)
				logger.Logger.WithFields(logrus.Fields{
					"iteration":     i,
					"retry_seconds": waitToReconnectSeconds,
				}).WithError(err).Error("Failed to connect to relay")
				continue
			}
			waitToReconnectSeconds = 0

			// start each app wallet subscription which have a child derived wallet key
			svc.startAllExistingAppsWalletSubscriptions(ctx, relay)

			// register a subscriber for events of "app_created" which handles creation of nostr subscription for new app
			svc.eventPublisher.RegisterSubscriber(&createAppSubscriber{svc: svc, relay: relay})

			// legacy single wallet subscription
			err = svc.startAppWalletSubscription(ctx, relay, svc.keys.GetNostrPublicKey(), svc.keys.GetNostrSecretKey())
			if err != nil {
				//err being non-nil means that we have an error on the websocket error channel. In this case we just try to reconnect.
				logger.Logger.WithError(err).Error("Got an error from the relay while listening to subscription.")
				continue
			}
			//err being nil means that the context was canceled and we should exit the program.
			break
		}
		closeRelay(relay)
		logger.Logger.Info("Relay subroutine ended")
	}()
	return nil
}

type createAppSubscriber struct {
	events.EventSubscriber
	svc   *service
	relay *nostr.Relay
}

func (s *createAppSubscriber) ConsumeEvent(ctx context.Context, event *events.Event, globalProperties map[string]interface{}) {
	if event.Event != "app_created" {
		return
	}
	properties, ok := event.Properties.(map[string]interface{})
	if !ok {
		logger.Logger.WithField("event", event).Error("Failed to cast event.Properties to map")
		return
	}
	id, ok := properties["id"].(uint)
	if !ok {
		logger.Logger.WithField("event", event).Error("Failed to get app id")
		return
	}
	walletPrivKey, err := s.svc.keys.GetAppWalletKey(uint32(id))
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to calculate app wallet priv key")
	}
	walletPubKey, err := nostr.GetPublicKey(walletPrivKey)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to calculate app wallet pub key")
	}

	go func() {
		err = s.svc.startAppWalletSubscription(ctx, s.relay, walletPubKey, walletPrivKey)
		if err != nil {
			logger.Logger.WithError(err).WithFields(logrus.Fields{
				"app_id": id}).Error("Failed to subscribe to wallet")
		}
		logger.Logger.WithFields(logrus.Fields{
			"app_id": id}).Info("App Nostr Subscription ended")
	}()
}

type deleteAppSubscriber struct {
	events.EventSubscriber
	walletPubkey      string
	relay             *nostr.Relay
	nostrSubscription *nostr.Subscription
	svc               *service
	infoEventId       string
}

func (s *deleteAppSubscriber) ConsumeEvent(ctx context.Context, event *events.Event, globalProperties map[string]interface{}) {
	if event.Event != "app_deleted" {
		return
	}
	properties, ok := event.Properties.(map[string]interface{})
	if !ok {
		logger.Logger.WithField("event", event).Error("Failed to cast event.Properties to map")
		return
	}
	id, _ := properties["id"].(uint)

	walletPrivKey, err := s.svc.keys.GetAppWalletKey(uint32(id))
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to calculate app wallet priv key")
	}
	walletPubKey, _ := nostr.GetPublicKey(walletPrivKey)
	if s.walletPubkey == walletPubKey {
		s.nostrSubscription.Unsub()
		err := s.svc.nip47Service.PublishNip47InfoDeletion(ctx, s.relay, walletPubKey, walletPrivKey, s.infoEventId)
		if err != nil {
			logger.Logger.WithField("event", event).Error("Failed to publish nip47 info deletion")
		}
	}
}

func (svc *service) startAllExistingAppsWalletSubscriptions(ctx context.Context, relay *nostr.Relay) {
	var apps []db.App
	result := svc.db.Where("wallet_pubkey != ?", "").Find(&apps)
	if result.Error != nil {
		logger.Logger.WithError(result.Error).Error("Failed to fetch App records with non-empty WalletPubkey")
		return
	}

	for _, app := range apps {
		go func(app db.App) {
			err := svc.startAppWalletSubscription(ctx, relay, app.WalletPubkey, "")
			if err != nil {
				logger.Logger.WithError(err).WithFields(logrus.Fields{
					"app_id": app.ID}).Error("Failed to subscribe to wallet")
			}
		}(app)
	}
}

func (svc *service) startAppWalletSubscription(ctx context.Context, relay *nostr.Relay, appWalletPubKey string, appWalletPrivKey string) error {
	var infoEventId string
	if appWalletPrivKey != "" {
		infoEvent, err := svc.GetNip47Service().PublishNip47Info(ctx, relay, appWalletPubKey, appWalletPrivKey, svc.lnClient)
		if err != nil {
			logger.Logger.WithError(err).Error("Could not publish NIP47 info")
		}
		infoEventId = infoEvent.ID
	}

	logger.Logger.Info("Subscribing to events for wallet ", appWalletPubKey)
	sub, err := relay.Subscribe(ctx, svc.createFilters(appWalletPubKey))
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to subscribe to events")
	}

	// register a subscriber for "app_deleted" events, which handles nostr subscription cancel and nip47 info event deletion
	svc.eventPublisher.RegisterSubscriber(&deleteAppSubscriber{nostrSubscription: sub, walletPubkey: appWalletPubKey, svc: svc, relay: relay, infoEventId: infoEventId})

	err = svc.StartSubscription(sub.Context, sub)
	if err != nil {
		logger.Logger.WithError(err).Error("Got an error from the relay while listening to subscription.")
		return err
	}
	return nil
}

func (svc *service) StartApp(encryptionKey string) error {
	if svc.lnClient != nil {
		return errors.New("app already started")
	}
	if !svc.cfg.CheckUnlockPassword(encryptionKey) {
		logger.Logger.Errorf("Invalid password")
		return errors.New("invalid password")
	}

	ctx, cancelFn := context.WithCancel(svc.ctx)

	err := svc.launchLNBackend(ctx, encryptionKey)
	if err != nil {
		logger.Logger.Errorf("Failed to launch LN backend: %v", err)
		svc.eventPublisher.Publish(&events.Event{
			Event: "nwc_node_start_failed",
		})
		cancelFn()
		return err
	}

	err = svc.startNostr(ctx, encryptionKey)
	if err != nil {
		cancelFn()
		return err
	}

	svc.appCancelFn = cancelFn

	return nil
}

func (svc *service) launchLNBackend(ctx context.Context, encryptionKey string) error {
	if svc.lnClient != nil {
		logger.Logger.Error("LNClient already started")
		return errors.New("LNClient already started")
	}

	svc.wg.Add(1)
	go func() {
		// ensure the LNClient is stopped properly before exiting
		<-ctx.Done()
		svc.stopLNClient()
	}()

	lnBackend, _ := svc.cfg.Get("LNBackendType", "")
	if lnBackend == "" {
		return errors.New("no LNBackendType specified")
	}

	logger.Logger.Infof("Launching LN Backend: %s", lnBackend)
	var lnClient lnclient.LNClient
	var err error
	switch lnBackend {
	case config.LNDBackendType:
		LNDAddress, _ := svc.cfg.Get("LNDAddress", encryptionKey)
		LNDCertHex, _ := svc.cfg.Get("LNDCertHex", encryptionKey)
		LNDMacaroonHex, _ := svc.cfg.Get("LNDMacaroonHex", encryptionKey)
		lnClient, err = lnd.NewLNDService(ctx, svc.eventPublisher, LNDAddress, LNDCertHex, LNDMacaroonHex)
	case config.LDKBackendType:
		Mnemonic, _ := svc.cfg.Get("Mnemonic", encryptionKey)
		LDKWorkdir := path.Join(svc.cfg.GetEnv().Workdir, "ldk")

		lnClient, err = ldk.NewLDKService(ctx, svc.cfg, svc.eventPublisher, Mnemonic, LDKWorkdir, svc.cfg.GetEnv().LDKNetwork, nil, false)
	case config.GreenlightBackendType:
		Mnemonic, _ := svc.cfg.Get("Mnemonic", encryptionKey)
		GreenlightInviteCode, _ := svc.cfg.Get("GreenlightInviteCode", encryptionKey)
		GreenlightWorkdir := path.Join(svc.cfg.GetEnv().Workdir, "greenlight")

		lnClient, err = greenlight.NewGreenlightService(svc.cfg, Mnemonic, GreenlightInviteCode, GreenlightWorkdir, encryptionKey)
	case config.BreezBackendType:
		Mnemonic, _ := svc.cfg.Get("Mnemonic", encryptionKey)
		BreezAPIKey, _ := svc.cfg.Get("BreezAPIKey", encryptionKey)
		GreenlightInviteCode, _ := svc.cfg.Get("GreenlightInviteCode", encryptionKey)
		BreezWorkdir := path.Join(svc.cfg.GetEnv().Workdir, "breez")

		lnClient, err = breez.NewBreezService(Mnemonic, BreezAPIKey, GreenlightInviteCode, BreezWorkdir)
	case config.PhoenixBackendType:
		PhoenixdAddress, _ := svc.cfg.Get("PhoenixdAddress", encryptionKey)
		PhoenixdAuthorization, _ := svc.cfg.Get("PhoenixdAuthorization", encryptionKey)

		lnClient, err = phoenixd.NewPhoenixService(PhoenixdAddress, PhoenixdAuthorization)
	case config.CashuBackendType:
		cashuMintUrl, _ := svc.cfg.Get("CashuMintUrl", encryptionKey)
		cashuWorkdir := path.Join(svc.cfg.GetEnv().Workdir, "cashu")

		lnClient, err = cashu.NewCashuService(cashuWorkdir, cashuMintUrl)
	default:
		logger.Logger.WithField("backend_type", lnBackend).Error("Unsupported LNBackendType")
		return fmt.Errorf("unsupported backend type: %s", lnBackend)
	}
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to launch LN backend")
		return err
	}

	// TODO: call a method on the LNClient here to check the LNClient is actually connectable,
	// (e.g. lnClient.CheckConnection()) Rather than it being a side-effect
	// in the LNClient init function

	svc.lnClient = lnClient
	info, err := lnClient.GetInfo(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch node info")
	}
	if info != nil {
		svc.eventPublisher.SetGlobalProperty("node_id", info.Pubkey)
		svc.eventPublisher.SetGlobalProperty("network", info.Network)
	}

	// Mark that the node has successfully started
	// This will ensure the user cannot go through the setup again
	err = svc.cfg.SetUpdate("NodeLastStartTime", strconv.FormatInt(time.Now().Unix(), 10), "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to set last node start time")
	}

	svc.eventPublisher.Publish(&events.Event{
		Event: "nwc_node_started",
		Properties: map[string]interface{}{
			"node_type": lnBackend,
		},
	})

	return nil
}

func closeRelay(relay *nostr.Relay) {
	if relay != nil && relay.IsConnected() {
		logger.Logger.Info("Closing relay connection...")
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Logger.WithField("r", r).Error("Recovered from panic when closing relay")
				}
			}()
			err := relay.Close()
			if err != nil {
				logger.Logger.WithError(err).Error("Could not close relay connection")
			}
		}()
	}
}
