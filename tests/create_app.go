package tests

import (
	"github.com/getAlby/hub/db"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
)

func CreateApp(svc *TestService) (app *db.App, ss []byte, err error) {
	senderPrivkey := nostr.GeneratePrivateKey()
	return CreateAppWithPrivateKey(svc, senderPrivkey)
}
func CreateAppWithPrivateKey(svc *TestService, senderPrivkey string) (app *db.App, ss []byte, err error) {

	senderPubkey, err := nostr.GetPublicKey(senderPrivkey)
	if err != nil {
		return nil, nil, err
	}

	ss, err = nip04.ComputeSharedSecret(svc.Keys.GetNostrPublicKey(), senderPrivkey)
	if err != nil {
		return nil, nil, err
	}

	app = &db.App{Name: "test", AppPubkey: senderPubkey}
	err = svc.DB.Create(app).Error
	if err != nil {
		return nil, nil, err
	}

	return app, ss, nil
}
