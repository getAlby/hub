package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/swaps"
	"github.com/getAlby/hub/version"

	"github.com/getAlby/go-nostr"
	"github.com/getAlby/go-nostr/nip19"
	"github.com/sirupsen/logrus"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/lnclient/lnd"
	"github.com/getAlby/hub/logger"
)

func (svc *service) startNostr(ctx context.Context) error {
	relayUrls := svc.cfg.GetRelayUrls()
	if len(relayUrls) == 0 {
		return errors.New("No relay URLs found")
	}

	npub, err := nip19.EncodePublicKey(svc.keys.GetNostrPublicKey())
	if err != nil {
		logger.Logger.WithError(err).Error("Error converting nostr privkey to pubkey")
		return err
	}

	logger.Logger.WithFields(logrus.Fields{
		"npub":       npub,
		"hex":        svc.keys.GetNostrPublicKey(),
		"version":    version.Tag,
		"relay_urls": relayUrls,
	}).Info("Starting Alby Hub")

	// To debug go-nostr, run with -tags "debug dev" (dev tag so LND build doesn't break with debug tag set)
	// go run -tags "debug dev" -ldflags="-X 'github.com/getAlby/hub/version.Tag=v1.20.0'" cmd/http/main.go
	if logger.Logger.GetLevel() >= logrus.DebugLevel {
		nostr.InfoLogger.SetOutput(logger.Logger.Out)
		nostr.DebugLogger.SetOutput(logger.Logger.Out)
	}

	// Start infinite loop which will be only broken by canceling ctx (SIGINT)
	pool := nostr.NewSimplePool(ctx, nostr.WithRelayOptions(
		nostr.WithNoticeHandler(svc.noticeHandler),
		nostr.WithRequestHeader(http.Header{
			"User-Agent": {"AlbyHub/" + version.Tag},
		}),
	))

	// initially try connect to relays (if hub has no apps, pool won't connect to relays by default)
	for _, relayUrl := range svc.cfg.GetRelayUrls() {
		_, err := pool.EnsureRelay(relayUrl)
		if err != nil {
			logger.Logger.WithError(err).WithField("relay_url", relayUrl).Error("failed to initially connect to relay")
		}
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				svc.relayStatuses = nil
				for _, relayUrl := range svc.cfg.GetRelayUrls() {
					relay, ok := pool.Relays.Load(relayUrl)
					svc.relayStatuses = append(svc.relayStatuses, RelayStatus{
						Url:    relayUrl,
						Online: ok && relay != nil && relay.IsConnected(),
					})
				}
				time.Sleep(10 * time.Second)
			}
		}
	}()

	svc.nip47Service.StartNotifier(ctx, pool)
	svc.nip47Service.StartNip47InfoPublisher(ctx, pool, svc.GetLNClient())

	// register a subscriber for events of "nwc_app_created" which handles creation of nostr subscription for new app
	createAppEventListener := &createAppConsumer{svc: svc, pool: pool}
	svc.eventPublisher.RegisterSubscriber(createAppEventListener)

	// register a subscriber for events of "nwc_app_updated" which handles re-publishing of nip47 event info
	updateAppEventListener := &updateAppConsumer{svc: svc}
	svc.eventPublisher.RegisterSubscriber(updateAppEventListener)

	// start each app wallet subscription which have a child derived wallet key
	svc.startAllExistingAppsWalletSubscriptions(ctx, pool)

	// check if there are still legacy apps in DB
	var legacyAppCount int64
	result := svc.db.Model(&db.App{}).Where("wallet_pubkey IS NULL").Count(&legacyAppCount)
	if result.Error != nil {
		logger.Logger.WithError(result.Error).Error("Failed to count Legacy Apps")
	}
	if legacyAppCount > 0 {
		go func() {
			logger.Logger.WithField("legacy_app_count", legacyAppCount).Info("Starting legacy app subscription")
			// legacy single wallet subscription - only subscribe once for all legacy apps
			// to ensure we do not get duplicate events
			svc.startAppWalletSubscription(ctx, pool, svc.keys.GetNostrPublicKey())
		}()
	}

	go func() {
		<-ctx.Done()
		logger.Logger.Info("Main context cancelled, exiting...")

		pool.Close("exiting")
		logger.Logger.Info("Relay subroutine ended")

		svc.eventPublisher.RemoveSubscriber(createAppEventListener)
		svc.eventPublisher.RemoveSubscriber(updateAppEventListener)
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
			for _, relayUrl := range svc.cfg.GetRelayUrls() {
				svc.nip47Service.EnqueueNip47InfoPublishRequest(0 /* unused */, svc.keys.GetNostrPublicKey(), svc.keys.GetNostrSecretKey(), relayUrl)
			}
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
			for _, relayUrl := range svc.cfg.GetRelayUrls() {
				svc.nip47Service.EnqueueNip47InfoPublishRequest(app.ID, *app.WalletPubkey, walletPrivKey, relayUrl)
			}
		}(app)
	}
}

func (svc *service) startAllExistingAppsWalletSubscriptions(ctx context.Context, pool *nostr.SimplePool) {
	var apps []db.App
	result := svc.db.Where("wallet_pubkey IS NOT NULL").Find(&apps)
	if result.Error != nil {
		logger.Logger.WithError(result.Error).Error("Failed to fetch App records with non-empty WalletPubkey")
		return
	}

	for _, app := range apps {
		go func(app db.App) {
			svc.startAppWalletSubscription(ctx, pool, *app.WalletPubkey)
		}(app)
	}
}

func (svc *service) startAppWalletSubscription(ctx context.Context, pool *nostr.SimplePool, appWalletPubKey string) error {

	logger.Logger.Info("Subscribing to events for wallet ", appWalletPubKey)

	filter := nostr.Filter{
		Tags:  nostr.TagMap{"p": []string{appWalletPubKey}},
		Kinds: []int{models.REQUEST_KIND},
	}

	for {
		subCtx, cancelSubscription := context.WithCancel(ctx)
		eventsChannel := pool.SubscribeMany(subCtx, svc.cfg.GetRelayUrls(), filter)

		// register a subscriber for "nwc_app_deleted" events, which handles
		// cancelling the nostr subscription and nip47 info event deletion
		deleteAppSubscriber := deleteAppConsumer{
			cancelSubscription: cancelSubscription,
			walletPubkey:       appWalletPubKey,
			svc:                svc,
			pool:               pool,
		}

		svc.eventPublisher.RegisterSubscriber(&deleteAppSubscriber)

		err := svc.watchSubscription(subCtx, pool, eventsChannel)

		svc.eventPublisher.RemoveSubscriber(&deleteAppSubscriber)
		if err != nil {
			logger.Logger.WithError(err).Error("got an error from the relay while listening to subscription, resubscribing")
			time.Sleep(3 * time.Second)
			continue
		}
		break
	}
	return nil
}

func (svc *service) watchSubscription(ctx context.Context, pool *nostr.SimplePool, eventsChannel chan nostr.RelayEvent) error {
	eventsChannelClosed := make(chan struct{})
	go func() {
		// loop through incoming events
		for event := range eventsChannel {
			select {
			case <-ctx.Done():
				return
			default:
				go svc.nip47Service.HandleEvent(ctx, pool, event.Event, svc.GetLNClient())
			}
		}
		logger.Logger.Debug("Relay subscription events channel ended")
		eventsChannelClosed <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		logger.Logger.Info("Exiting subscription due to context exit...")
		return nil
	case <-eventsChannelClosed:
		// in go-nostr pool, currently if the relay sends a close that is not "auth-required:"
		// this will trigger closing the subscription channel. We return an error to trigger a resubscribe.
		logger.Logger.Info("Subscription was exited abnormally")
		return errors.New("subscription exited abnormally")
	}
}

// startAppWithRetries runs StartApp with capped exponential backoff until it
// succeeds or fails permanently. It is used for auto-unlock on startup: a hub
// that boots before its internet connection is restored (e.g. after a power
// cut) must not silently stay locked because the first start attempt failed.
func (svc *service) startAppWithRetries(encryptionKey string) {
	backoff := 10 * time.Second
	const maxBackoff = 5 * time.Minute

	for {
		err := svc.StartApp(encryptionKey)
		if err == nil {
			return
		}
		if errors.Is(err, ErrAlreadyStarted) {
			logger.Logger.Info("App was started manually, stopping auto-unlock retries")
			return
		}
		if errors.Is(err, ErrInvalidPassword) || errors.Is(err, ErrIncompleteWalletData) {
			logger.Logger.WithError(err).Error("Auto-unlock failed permanently, not retrying")
			return
		}

		logger.Logger.WithError(err).WithField("retry_in", backoff.String()).Error("Auto-unlock failed, retrying")
		select {
		case <-svc.ctx.Done():
			return
		case <-time.After(backoff):
		}
		backoff = min(backoff*2, maxBackoff)
	}
}

// WithStartLock runs fn while holding the start/stop lock, returning
// ErrAppBusy if a start, stop or setup operation is already in progress.
func (svc *service) WithStartLock(fn func() error) error {
	if !svc.startMutex.TryLock() {
		return ErrAppBusy
	}
	defer svc.startMutex.Unlock()
	return fn()
}

func (svc *service) StartApp(encryptionKey string) error {
	// do not allow to start twice in case this is somehow called twice
	return svc.WithStartLock(func() error {
		return svc.startAppInternal(encryptionKey)
	})
}

func (svc *service) startAppInternal(encryptionKey string) error {
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
		return ErrAlreadyStarted
	}
	unlockPasswordCheckSet, err := svc.cfg.IsUnlockPasswordCheckSet()
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to check unlock password check")
		return fmt.Errorf("check unlock password check: %w", err)
	}
	if !unlockPasswordCheckSet {
		logger.Logger.Error("Unlock password check is missing from the database")
		return ErrIncompleteWalletData
	}
	if !svc.cfg.CheckUnlockPassword(encryptionKey) {
		logger.Logger.Errorf("Invalid password")
		return ErrInvalidPassword
	}

	err = svc.cfg.LoadJWTSecret(encryptionKey)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to load JWT secret")
		return err
	}

	err = svc.keys.Init(svc.cfg, encryptionKey)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to init nostr keys")
		return err
	}

	ctx, cancelFn := context.WithCancel(svc.ctx)

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

	svc.swapsService = swaps.NewSwapsService(ctx, svc.db, svc.cfg, svc.keys, svc.eventPublisher, svc.GetLNClient(), svc.transactionsService, svc.albySvc, encryptionKey)

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
	switch lnBackend {
	case config.LNDBackendType:
		LNDAddress, _ := svc.cfg.Get("LNDAddress", encryptionKey)
		LNDCertHex, _ := svc.cfg.Get("LNDCertHex", encryptionKey)
		LNDMacaroonHex, _ := svc.cfg.Get("LNDMacaroonHex", encryptionKey)
		lnClient, err = lnd.NewLNDService(ctx, svc.eventPublisher, LNDAddress, LNDCertHex, LNDMacaroonHex)
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
