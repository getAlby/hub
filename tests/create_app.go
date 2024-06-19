package tests

import (
	"github.com/getAlby/nostr-wallet-connect/db"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
)

func CreateApp(svc *TestService) (app *db.App, ss []byte, err error) {
	senderPrivkey := nostr.GeneratePrivateKey()
	senderPubkey, err := nostr.GetPublicKey(senderPrivkey)
	if err != nil {
		return nil, nil, err
	}

	ss, err = nip04.ComputeSharedSecret(svc.Keys.GetNostrPublicKey(), senderPrivkey)
	if err != nil {
		return nil, nil, err
	}

	app = &db.App{Name: "test", NostrPubkey: senderPubkey}
	err = svc.DB.Create(app).Error
	if err != nil {
		return nil, nil, err
	}

	return app, ss, nil
}
