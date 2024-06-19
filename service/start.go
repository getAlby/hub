package service

import (
	"context"
	"errors"
	"path"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/sirupsen/logrus"

	"github.com/getAlby/nostr-wallet-connect/config"
	"github.com/getAlby/nostr-wallet-connect/events"
	"github.com/getAlby/nostr-wallet-connect/lnclient"
	"github.com/getAlby/nostr-wallet-connect/lnclient/breez"
	"github.com/getAlby/nostr-wallet-connect/lnclient/cashu"
	"github.com/getAlby/nostr-wallet-connect/lnclient/greenlight"
	"github.com/getAlby/nostr-wallet-connect/lnclient/ldk"
	"github.com/getAlby/nostr-wallet-connect/lnclient/lnd"
	"github.com/getAlby/nostr-wallet-connect/lnclient/phoenixd"
	"github.com/getAlby/nostr-wallet-connect/logger"
)

func (svc *service) StartNostr(ctx context.Context, encryptionKey string) error {

	relayUrl := svc.cfg.GetRelayUrl()

	err := svc.keys.Init(svc.cfg, encryptionKey)
	if err != nil {
		logger.Logger.WithError(err).Fatal("Failed to init nostr keys")
	}

	npub, err := nip19.EncodePublicKey(svc.keys.GetNostrPublicKey())
	if err != nil {
		logger.Logger.WithError(err).Fatal("Error converting nostr privkey to pubkey")
	}

	logger.Logger.WithFields(logrus.Fields{
		"npub": npub,
		"hex":  svc.keys.GetNostrPublicKey(),
	}).Info("Starting nostr-wallet-connect")
	svc.wg.Add(1)
	go func() {
		//Start infinite loop which will be only broken by canceling ctx (SIGINT)
		var relay *nostr.Relay

		for i := 0; ; i++ {
			// wait for a delay before retrying except on first iteration
			if i > 0 {
				sleepDuration := 10
				contextCancelled := false
				logger.Logger.Infof("[Iteration %d] Retrying in %d seconds...", i, sleepDuration)

				select {
				case <-ctx.Done(): //context cancelled
					logger.Logger.Info("service context cancelled while waiting for retry")
					contextCancelled = true
				case <-time.After(time.Duration(sleepDuration) * time.Second): //timeout
				}
				if contextCancelled {
					break
				}
			}
			closeRelay(relay)

			//connect to the relay
			logger.Logger.Infof("Connecting to the relay: %s", relayUrl)

			relay, err = nostr.RelayConnect(ctx, relayUrl, nostr.WithNoticeHandler(svc.noticeHandler))
			if err != nil {
				logger.Logger.WithError(err).Error("Failed to connect to relay")
				continue
			}

			//publish event with NIP-47 info
			err = svc.nip47Service.PublishNip47Info(ctx, relay)
			if err != nil {
				logger.Logger.WithError(err).Error("Could not publish NIP47 info")
			}

			logger.Logger.Info("Subscribing to events")
			sub, err := relay.Subscribe(ctx, svc.createFilters(svc.keys.GetNostrPublicKey()))
			if err != nil {
				logger.Logger.WithError(err).Error("Failed to subscribe to events")
				continue
			}
			err = svc.StartSubscription(sub.Context, sub)
			if err != nil {
				//err being non-nil means that we have an error on the websocket error channel. In this case we just try to reconnect.
				logger.Logger.WithError(err).Error("Got an error from the relay while listening to subscription.")
				continue
			}
			//err being nil means that the context was canceled and we should exit the program.
			break
		}
		closeRelay(relay)
		svc.Shutdown()
		logger.Logger.Info("Relay subroutine ended")
		svc.wg.Done()
	}()
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

	svc.StartNostr(ctx, encryptionKey)
	svc.appCancelFn = cancelFn
	return nil
}

func (svc *service) launchLNBackend(ctx context.Context, encryptionKey string) error {
	err := svc.StopLNClient()
	if err != nil {
		return err
	}

	lnBackend, _ := svc.cfg.Get("LNBackendType", "")
	if lnBackend == "" {
		return errors.New("no LNBackendType specified")
	}

	logger.Logger.Infof("Launching LN Backend: %s", lnBackend)
	var lnClient lnclient.LNClient
	switch lnBackend {
	case config.LNDBackendType:
		LNDAddress, _ := svc.cfg.Get("LNDAddress", encryptionKey)
		LNDCertHex, _ := svc.cfg.Get("LNDCertHex", encryptionKey)
		LNDMacaroonHex, _ := svc.cfg.Get("LNDMacaroonHex", encryptionKey)
		lnClient, err = lnd.NewLNDService(ctx, LNDAddress, LNDCertHex, LNDMacaroonHex)
	case config.LDKBackendType:
		Mnemonic, _ := svc.cfg.Get("Mnemonic", encryptionKey)
		LDKWorkdir := path.Join(svc.cfg.GetEnv().Workdir, "ldk")

		lnClient, err = ldk.NewLDKService(ctx, svc.cfg, svc.eventPublisher, Mnemonic, LDKWorkdir, svc.cfg.GetEnv().LDKNetwork, svc.cfg.GetEnv().LDKEsploraServer, svc.cfg.GetEnv().LDKGossipSource)
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
		logger.Logger.Fatalf("Unsupported LNBackendType: %v", lnBackend)
	}
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to launch LN backend")
		return err
	}

	info, err := lnClient.GetInfo(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch node info")
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
