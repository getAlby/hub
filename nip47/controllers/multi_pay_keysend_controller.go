package controllers

import (
	"context"
	"sync"

	"github.com/getAlby/nostr-wallet-connect/db"
	"github.com/getAlby/nostr-wallet-connect/nip47/models"
	"github.com/nbd-wtf/go-nostr"
)

type multiPayKeysendParams struct {
	Keysends []multiPayKeysendElement `json:"keysends"`
}

type multiPayKeysendElement struct {
	payKeysendParams
	Id string `json:"id"`
}

func (controller *nip47Controller) HandleMultiPayKeysendEvent(ctx context.Context, nip47Request *models.Request, requestEventId uint, app *db.App, checkPermission checkPermissionFunc, publishResponse publishFunc) {
	multiPayParams := &multiPayKeysendParams{}
	resp := decodeRequest(nip47Request, multiPayParams)
	if resp != nil {
		publishResponse(resp, nostr.Tags{})
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(multiPayParams.Keysends))
	for _, keysendInfo := range multiPayParams.Keysends {
		go func(keysendInfo multiPayKeysendElement) {
			defer wg.Done()

			keysendDTagValue := keysendInfo.Id
			if keysendDTagValue == "" {
				keysendDTagValue = keysendInfo.Pubkey
			}
			dTag := []string{"d", keysendDTagValue}

			controller.
				payKeysend(ctx, &keysendInfo.payKeysendParams, nip47Request, requestEventId, app, checkPermission, publishResponse, nostr.Tags{dTag})
		}(keysendInfo)
	}

	wg.Wait()
}
