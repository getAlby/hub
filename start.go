package main

import (
	"github.com/nbd-wtf/go-nostr"
)

func (svc *Service) StartNostr(encryptionKey string) error {
	relayUrl, _ := svc.cfg.Get("Relay", encryptionKey)
	nostrSecretKey, _ := svc.cfg.Get("NostrSecretKey", encryptionKey)
	if nostrSecretKey == "" {
		nostrSecretKey = nostr.GeneratePrivateKey()
		svc.cfg.SetUpdate("NostrSecretKey", nostrSecretKey, encryptionKey)
	}
	nostrPublicKey, err := nostr.GetPublicKey(nostrSecretKey)
	if err != nil {
		svc.Logger.Errorf("Error converting nostr privkey to pubkey: %v", err)
		return err
	}
	svc.cfg.NostrSecretKey = nostrSecretKey
	svc.cfg.NostrPublicKey = nostrPublicKey

	svc.Logger.Infof("Starting nostr-wallet-connect. npub: %s", svc.cfg.NostrPublicKey)
	svc.wg.Add(1)
	go func() {
		//connect to the relay
		svc.Logger.Infof("Connecting to the relay: %s", relayUrl)

		relay, err := nostr.RelayConnect(svc.ctx, relayUrl, nostr.WithNoticeHandler(svc.noticeHandler))
		if err != nil {
			svc.Logger.Errorf("Failed to connect to relay: %v", err)
			svc.wg.Done()
			return
		}

		//publish event with NIP-47 info
		err = svc.PublishNip47Info(svc.ctx, relay)
		if err != nil {
			svc.Logger.WithError(err).Error("Could not publish NIP47 info")
		}

		//Start infinite loop which will be only broken by canceling ctx (SIGINT)
		//TODO: we can start this loop for multiple relays
		for {
			svc.Logger.Info("Subscribing to events")
			sub, err := relay.Subscribe(svc.ctx, svc.createFilters(svc.cfg.NostrPublicKey))
			if err != nil {
				svc.Logger.Fatal(err)
			}
			err = svc.StartSubscription(svc.ctx, sub)
			if err != nil {
				//err being non-nil means that we have an error on the websocket error channel. In this case we just try to reconnect.
				svc.Logger.WithError(err).Error("Got an error from the relay while listening to subscription. Reconnecting...")
				relay, err = nostr.RelayConnect(svc.ctx, relayUrl)
				if err != nil {
					svc.Logger.Fatal(err)
				}
				continue
			}
			//err being nil means that the context was canceled and we should exit the program.
			break
		}
		svc.Logger.Info("Disconnecting from relay...")
		err = relay.Close()
		if err != nil {
			svc.Logger.Error(err)
		}
		svc.Shutdown()
		svc.Logger.Info("Relay subroutine ended")
		svc.wg.Done()
	}()
	return nil
}

func (svc *Service) StartApp(encryptionKey string) error {
	// svc.Logger.Infof("Starting nostr-wallet-connect. npub: %s hex: %s", npub, identityPubkey)
	err := svc.launchLNBackend(encryptionKey)
	if err != nil {
		svc.Logger.Warnf("Failed to launch LN backend: %v", err)
		return err
	}

	svc.StartNostr(encryptionKey)
	return nil
}

func (svc *Service) Shutdown() {
	if svc.lnClient != nil {
		svc.Logger.Info("Shutting down LN backend...")
		err := svc.lnClient.Shutdown()
		if err != nil {
			svc.Logger.Error(err)
		}
	}
}
