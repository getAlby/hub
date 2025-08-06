package service

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strconv"
	"time"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/swaps"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/sirupsen/logrus"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/lnclient/cashu"
	"github.com/getAlby/hub/lnclient/ldk"
	"github.com/getAlby/hub/lnclient/lnd"
	"github.com/getAlby/hub/lnclient/phoenixd"
	"github.com/getAlby/hub/logger"
)

func (svc *service) startNostr(ctx context.Context) error {
	relayUrl := svc.cfg.GetRelayUrl()

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
		// Start infinite loop which will be only broken by canceling ctx (SIGINT)
		var relay *nostr.Relay
		waitToReconnectSeconds := 0
		var createAppEventListener events.EventSubscriber
		var updateAppEventListener events.EventSubscriber
		for i := 0; ; i++ {
			// wait for a delay if any before retrying
			contextCancelled := false

			svc.setRelayReady(false)

			select {
			case <-ctx.Done(): // application service context cancelled
				logger.Logger.Info("service context cancelled")
				contextCancelled = true
			case <-time.After(time.Duration(waitToReconnectSeconds) * time.Second): // timeout
			}
			if contextCancelled {
				break
			}

			closeRelay(relay)

			// connect to the relay
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
			logger.Logger.WithFields(logrus.Fields{
				"relay_url": relayUrl,
			}).Info("Connected to the relay")
			waitToReconnectSeconds = 0

			svc.nip47Service.StartNotifier(relay)
			svc.nip47Service.StartNip47InfoPublisher(relay, svc.lnClient)

			// register a subscriber for events of "nwc_app_created" which handles creation of nostr subscription for new app
			if createAppEventListener != nil {
				svc.eventPublisher.RemoveSubscriber(createAppEventListener)
			}
			createAppEventListener = &createAppConsumer{svc: svc, relay: relay}
			svc.eventPublisher.RegisterSubscriber(createAppEventListener)

			// register a subscriber for events of "nwc_app_updated" which handles re-publishing of nip47 event info
			if updateAppEventListener != nil {
				svc.eventPublisher.RemoveSubscriber(updateAppEventListener)
			}
			updateAppEventListener = &updateAppConsumer{svc: svc, relay: relay}
			svc.eventPublisher.RegisterSubscriber(updateAppEventListener)

			// start each app wallet subscription which have a child derived wallet key
			svc.startAllExistingAppsWalletSubscriptions(ctx, relay)

			// check if there are still legacy apps in DB
			var legacyAppCount int64
			result := svc.db.Model(&db.App{}).Where("wallet_pubkey IS NULL").Count(&legacyAppCount)
			if result.Error != nil {
				logger.Logger.WithError(result.Error).Error("Failed to count Legacy Apps")
				return
			}
			if legacyAppCount > 0 {
				go func() {
					logger.Logger.WithField("legacy_app_count", legacyAppCount).Info("Starting legacy app subscription")
					// legacy single wallet subscription - only subscribe once for all legacy apps
					// to ensure we do not get duplicate events
					err = svc.startAppWalletSubscription(ctx, relay, svc.keys.GetNostrPublicKey())
					if err != nil && !errors.Is(err, context.Canceled) {
						// err being non-nil means that we have an error on the websocket error channel. In this case we just try to reconnect.
						logger.Logger.WithError(err).Error("Got an error from the relay while listening to legacy subscription.")
					}
				}()
			}

			svc.setRelayReady(true)

			select {
			case <-ctx.Done():
				logger.Logger.Info("Main context cancelled, exiting...")
			case <-relay.Context().Done():
				// err being non-nil means that we have an error on the websocket error channel. In this case we just try to reconnect.
				if relay.ConnectionError != nil {
					logger.Logger.WithError(relay.ConnectionError).Error("Got an error from the relay, trying to reconnect")
				} else {
					logger.Logger.Error("Relay context cancelled, but no connection error...trying to reconnect")
				}
			}
		}
		closeRelay(relay)
		logger.Logger.Info("Relay subroutine ended")
	}()
	return nil
}

// In case the relay somehow loses events or the hub updates with
// new capabilities, we re-publish info events for all apps on startup
// to ensure that they are retrievable for all connections
func (svc *service) publishAllAppInfoEvents() {
	func() {
		var legacyAppCount int64
		result := svc.db.Model(&db.App{}).Where("wallet_pubkey IS NULL").Count(&legacyAppCount)
		if result.Error != nil {
			logger.Logger.WithError(result.Error).Error("Failed to fetch App records with empty WalletPubkey")
			return
		}
		if legacyAppCount > 0 {
			logger.Logger.WithField("legacy_app_count", legacyAppCount).Debug("Enqueuing publish of legacy info event")
			svc.nip47Service.EnqueueNip47InfoPublishRequest(svc.keys.GetNostrPublicKey(), svc.keys.GetNostrSecretKey())
		}
	}()

	var apps []db.App
	result := svc.db.Where("wallet_pubkey IS NOT NULL").Find(&apps)
	if result.Error != nil {
		logger.Logger.WithError(result.Error).Error("Failed to fetch App records with non-empty WalletPubkey")
		return
	}

	for _, app := range apps {
		func(app db.App) {
			// queue info event publish request for all existing apps
			walletPrivKey, err := svc.keys.GetAppWalletKey(app.ID)
			if err != nil {
				logger.Logger.WithError(err).WithFields(logrus.Fields{
					"app_id": app.ID}).Error("Could not get app wallet key")
				return
			}
			logger.Logger.WithField("app_id", app.ID).Debug("Enqueuing publish of app info event")
			svc.nip47Service.EnqueueNip47InfoPublishRequest(*app.WalletPubkey, walletPrivKey)
		}(app)
	}
}

func (svc *service) startAllExistingAppsWalletSubscriptions(ctx context.Context, relay *nostr.Relay) {
	var apps []db.App
	result := svc.db.Where("wallet_pubkey IS NOT NULL").Find(&apps)
	if result.Error != nil {
		logger.Logger.WithError(result.Error).Error("Failed to fetch App records with non-empty WalletPubkey")
		return
	}

	for _, app := range apps {
		go func(app db.App) {
			err := svc.startAppWalletSubscription(ctx, relay, *app.WalletPubkey)
			if err != nil && !errors.Is(err, context.Canceled) {
				logger.Logger.WithError(err).WithFields(logrus.Fields{
					"app_id": app.ID}).Error("Subscription error")
				return
			}
		}(app)
	}
}

func (svc *service) startAppWalletSubscription(ctx context.Context, relay *nostr.Relay, appWalletPubKey string) error {

	logger.Logger.Info("Subscribing to events for wallet ", appWalletPubKey)
	sub, err := relay.Subscribe(ctx, svc.createFilters(appWalletPubKey))
	if err != nil {
		return fmt.Errorf("failed to subscribe to events: %w", err)
	}

	// register a subscriber for "nwc_app_deleted" events, which handles nostr subscription cancel and nip47 info event deletion
	deleteEventSubscriber := deleteAppConsumer{nostrSubscription: sub, walletPubkey: appWalletPubKey, svc: svc, relay: relay}
	svc.eventPublisher.RegisterSubscriber(&deleteEventSubscriber)

	err = svc.StartSubscription(sub.Context, sub)
	svc.eventPublisher.RemoveSubscriber(&deleteEventSubscriber)
	if err != nil {
		return fmt.Errorf("got an error from the relay while listening to subscription: %w", err)
	}
	return nil
}

func (svc *service) StartSubscription(ctx context.Context, sub *nostr.Subscription) error {
	go func() {
		// loop through incoming events
		for event := range sub.Events {
			go svc.nip47Service.HandleEvent(ctx, sub.Relay, event, svc.lnClient)
		}
		logger.Logger.Debug("Relay subscription events channel ended")
	}()

	<-ctx.Done()

	if err := sub.Relay.ConnectionError; err != nil {
		return fmt.Errorf("relay connection error: %w", err)
	}
	logger.Logger.Info("Exiting subscription...")
	return nil
}

func (svc *service) StartApp(encryptionKey string) error {
	defer func() {
		svc.startupState = ""
	}()

	svc.startupState = "Initializing"
	albyIdentifier, err := svc.albyOAuthSvc.GetUserIdentifier()
	if err != nil {
		return err
	}
	if albyIdentifier != "" && !svc.albyOAuthSvc.IsConnected(svc.ctx) {
		return errors.New("alby account is not authenticated")
	}

	if svc.lnClient != nil {
		return errors.New("app already started")
	}
	if !svc.cfg.CheckUnlockPassword(encryptionKey) {
		logger.Logger.Errorf("Invalid password")
		return errors.New("invalid password")
	}

	ctx, cancelFn := context.WithCancel(svc.ctx)

	err = svc.keys.Init(svc.cfg, encryptionKey)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to init nostr keys")
		cancelFn()
		return err
	}

	svc.startupState = "Launching Node"
	err = svc.launchLNBackend(ctx, encryptionKey)
	if err != nil {
		logger.Logger.Errorf("Failed to launch LN backend: %v", err)
		svc.eventPublisher.Publish(&events.Event{
			Event: "nwc_node_start_failed",
		})
		cancelFn()
		return err
	}

	svc.swapsService = swaps.NewSwapsService(ctx, svc.db, svc.cfg, svc.keys, svc.eventPublisher, svc.lnClient, svc.transactionsService)

	svc.publishAllAppInfoEvents()

	svc.startupState = "Connecting To Relay"
	err = svc.startNostr(ctx)
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
	vssEnabled := false
	switch lnBackend {
	case config.LNDBackendType:
		LNDAddress, _ := svc.cfg.Get("LNDAddress", encryptionKey)
		LNDCertHex, _ := svc.cfg.Get("LNDCertHex", encryptionKey)
		LNDMacaroonHex, _ := svc.cfg.Get("LNDMacaroonHex", encryptionKey)
		lnClient, err = lnd.NewLNDService(ctx, svc.eventPublisher, LNDAddress, LNDCertHex, LNDMacaroonHex)
	case config.LDKBackendType:
		mnemonic, _ := svc.cfg.Get("Mnemonic", encryptionKey)
		ldkWorkdir := path.Join(svc.cfg.GetEnv().Workdir, "ldk")
		var vssToken string
		vssToken, err = svc.requestVssToken(ctx)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to request VSS token")
			return err
		}
		vssEnabled = vssToken != ""

		svc.startupState = "Launching Node"
		setStartupState := func(startupState string) {
			svc.startupState = startupState
		}
		lnClient, err = ldk.NewLDKService(ctx, svc.cfg, svc.eventPublisher, mnemonic, ldkWorkdir, svc.cfg.GetNetwork(), vssToken, setStartupState)
	case config.PhoenixBackendType:
		PhoenixdAddress, _ := svc.cfg.Get("PhoenixdAddress", encryptionKey)
		PhoenixdAuthorization, _ := svc.cfg.Get("PhoenixdAuthorization", encryptionKey)

		lnClient, err = phoenixd.NewPhoenixService(PhoenixdAddress, PhoenixdAuthorization)
	case config.CashuBackendType:
		mnemonic, _ := svc.cfg.Get("Mnemonic", encryptionKey)
		cashuMintUrl, _ := svc.cfg.Get("CashuMintUrl", encryptionKey)
		cashuWorkdir := path.Join(svc.cfg.GetEnv().Workdir, "cashu")

		lnClient, err = cashu.NewCashuService(svc.cfg, cashuWorkdir, mnemonic, cashuMintUrl)
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
			"node_type":   lnBackend,
			"vss_enabled": vssEnabled,
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

func (svc *service) requestVssToken(ctx context.Context) (string, error) {
	nodeLastStartTime, _ := svc.cfg.Get("NodeLastStartTime", "")

	// for brand new nodes, consider enabling VSS
	if nodeLastStartTime == "" && svc.cfg.GetEnv().LDKVssUrl != "" {
		svc.startupState = "Checking Subscription"
		albyUserIdentifier, err := svc.albyOAuthSvc.GetUserIdentifier()
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to fetch alby user identifier")
			return "", err
		}
		if albyUserIdentifier != "" {
			me, err := svc.albyOAuthSvc.GetMe(ctx)
			if err != nil {
				logger.Logger.WithError(err).Error("Failed to fetch alby user")
				return "", err
			}
			// only activate VSS for Alby paid subscribers
			if me.Subscription.PlanCode != "" {
				svc.cfg.SetUpdate("LdkVssEnabled", "true", "")
			}
		}
	}

	vssToken := ""
	vssEnabled, _ := svc.cfg.Get("LdkVssEnabled", "")
	if vssEnabled == "true" {
		svc.startupState = "Fetching VSS token"
		vssNodeIdentifier, err := ldk.GetVssNodeIdentifier(svc.keys)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to get VSS node identifier")
			return "", err
		}
		vssToken, err = svc.albyOAuthSvc.GetVssAuthToken(ctx, vssNodeIdentifier)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to fetch VSS JWT token")

			existingVssToken, _ := svc.cfg.Get("VssToken", "")
			if existingVssToken != "" {
				logger.Logger.Warn("Using stored VSS JWT token")
				return existingVssToken, nil
			}

			return "", err
		}
		err = svc.cfg.SetUpdate("VssToken", vssToken, "")
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to save VSS JWT token to user config")
		}
	}
	return vssToken, nil
}
