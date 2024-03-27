package main

import (
	"errors"
	"time"

	"github.com/getAlby/nostr-wallet-connect/events"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
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

	npub, err := nip19.EncodePublicKey(svc.cfg.NostrPublicKey)
	if err != nil {
		svc.Logger.Fatalf("Error converting nostr privkey to pubkey: %v", err)
	}

	svc.Logger.Infof("Starting nostr-wallet-connect. npub: %s hex: %s", npub, svc.cfg.NostrPublicKey)
	svc.wg.Add(1)
	go func() {
		//Start infinite loop which will be only broken by canceling ctx (SIGINT)
		//TODO: we can start this loop for multiple relays
		var relay *nostr.Relay

		for i := 0; ; i++ {
			// wait for a delay before retrying except on first iteration
			if i > 0 {
				sleepDuration := 10
				contextCancelled := false
				svc.Logger.Infof("[Iteration %d] Retrying in %d seconds...", i, sleepDuration)

				select {
				case <-svc.ctx.Done(): //context cancelled
					svc.Logger.Info("service context cancelled while waiting for retry")
					contextCancelled = true
				case <-time.After(time.Duration(sleepDuration) * time.Second): //timeout
				}
				if contextCancelled {
					break
				}
			}
			if relay != nil && relay.IsConnected() {
				err := relay.Close()
				if err != nil {
					svc.Logger.WithError(err).Error("Could not close relay connection")
				}
			}

			//connect to the relay
			svc.Logger.Infof("Connecting to the relay: %s", relayUrl)

			relay, err := nostr.RelayConnect(svc.ctx, relayUrl, nostr.WithNoticeHandler(svc.noticeHandler))
			if err != nil {
				svc.Logger.WithError(err).Error("Failed to connect to relay")
				continue
			}

			//publish event with NIP-47 info
			err = svc.PublishNip47Info(svc.ctx, relay)
			if err != nil {
				svc.Logger.WithError(err).Error("Could not publish NIP47 info")
			}

			svc.Logger.Info("Subscribing to events")
			sub, err := relay.Subscribe(svc.ctx, svc.createFilters(svc.cfg.NostrPublicKey))
			if err != nil {
				svc.Logger.WithError(err).Error("Failed to subscribe to events")
				continue
			}
			err = svc.StartSubscription(svc.ctx, sub)
			if err != nil {
				//err being non-nil means that we have an error on the websocket error channel. In this case we just try to reconnect.
				svc.Logger.WithError(err).Error("Got an error from the relay while listening to subscription.")
				continue
			}
			//err being nil means that the context was canceled and we should exit the program.
			break
		}
		svc.Logger.Info("Disconnecting from relay...")
		if relay != nil && relay.IsConnected() {
			err := relay.Close()
			if err != nil {
				svc.Logger.WithError(err).Error("Could not close relay connection")
			}
		}
		svc.Shutdown()
		svc.Logger.Info("Relay subroutine ended")
		svc.wg.Done()
	}()
	return nil
}

func (svc *Service) StartApp(encryptionKey string) error {
	if !svc.cfg.CheckUnlockPassword(encryptionKey) {
		svc.Logger.Errorf("Invalid password")
		return errors.New("invalid password")
	}

	err := svc.launchLNBackend(encryptionKey)
	if err != nil {
		svc.Logger.Errorf("Failed to launch LN backend: %v", err)
		return err
	}

	svc.StartNostr(encryptionKey)
	return nil
}

func (svc *Service) Shutdown() {
	svc.StopLNClient()
	svc.EventLogger.Log(&events.Event{
		Event: "nwc_stopped",
	})
	// wait for any remaining events
	time.Sleep(1 * time.Second)
}
