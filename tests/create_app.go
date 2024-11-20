package tests

import (
	"time"

	db "github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	"gorm.io/gorm"
)

func CreateApp(svc *TestService) (app *db.App, ss []byte, err error) {
	return CreateAppWithPrivateKey(svc, "")
}
func CreateAppWithPrivateKey(svc *TestService, senderPrivkey string) (app *db.App, ss []byte, err error) {
	senderPubkey := ""
	if senderPrivkey != "" {
		var err error
		senderPubkey, err = nostr.GetPublicKey(senderPrivkey)
		if err != nil {
			return nil, nil, err
		}
	}

	var expiresAt *time.Time
	app, pairingSecretKey, err := svc.AppsService.CreateApp("test", senderPubkey, 0, "monthly", expiresAt, nil, false, nil)
	if pairingSecretKey == "" {
		pairingSecretKey = senderPrivkey
	}

	ss, err = nip04.ComputeSharedSecret(*app.WalletPubkey, pairingSecretKey)
	if err != nil {
		return nil, nil, err
	}

	return app, ss, nil
}

func CreateLegacyApp(svc *TestService, senderPrivkey string) (app *db.App, ss []byte, err error) {

	pairingPublicKey, _ := nostr.GetPublicKey(senderPrivkey)

	app = &db.App{Name: "test", AppPubkey: pairingPublicKey, Isolated: false}

	err = svc.DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Save(&app).Error
		if err != nil {
			return err
		}

		// commit transaction
		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	svc.EventPublisher.Publish(&events.Event{
		Event: "nwc_app_created",
		Properties: map[string]interface{}{
			"name": "test",
			"id":   app.ID,
		},
	})

	ss, err = nip04.ComputeSharedSecret(svc.Keys.GetNostrPublicKey(), senderPrivkey)
	return app, ss, nil
}
