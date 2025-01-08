package tests

import (
	"time"

	db "github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/nip47/cipher"
	"github.com/nbd-wtf/go-nostr"
	"gorm.io/gorm"
)

type CreateAppFn func(svc *TestService, senderPrivkey string, nip47Version string) (app *db.App, nip47Cipher *cipher.Nip47Cipher, err error)

func CreateApp(svc *TestService, nip47Version string) (app *db.App, cipher *cipher.Nip47Cipher, err error) {
	return CreateAppWithPrivateKey(svc, "", nip47Version)
}

func CreateAppWithPrivateKey(svc *TestService, senderPrivkey, nip47Version string) (app *db.App, nip47Cipher *cipher.Nip47Cipher, err error) {
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

	nip47Cipher, err = cipher.NewNip47Cipher(nip47Version, *app.WalletPubkey, pairingSecretKey)
	if err != nil {
		return nil, nil, err
	}

	return app, nip47Cipher, nil
}

func CreateAppWithSharedWalletPubkey(svc *TestService, senderPrivkey, nip47Version string) (app *db.App, nip47Cipher *cipher.Nip47Cipher, err error) {

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

	nip47Cipher, err = cipher.NewNip47Cipher(nip47Version, svc.Keys.GetNostrPublicKey(), senderPrivkey)
	return app, nip47Cipher, nil
}
