package nip47

import (
	"encoding/json"
	"testing"

	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/tests"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/stretchr/testify/assert"
)

// TODO: test HandleEvent
// TODO: test a request cannot be processed twice
// TODO: test if an app doesn't exist it returns the right error code

func TestCreateResponse(t *testing.T) {
	defer tests.RemoveTestService()
	svc, err := tests.CreateTestService()
	assert.NoError(t, err)

	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)
	assert.NoError(t, err)

	reqEvent := &nostr.Event{
		Kind:    models.REQUEST_KIND,
		PubKey:  reqPubkey,
		Content: "1",
	}

	reqEvent.ID = "12345"

	ss, err := nip04.ComputeSharedSecret(reqPubkey, svc.Keys.GetNostrSecretKey())
	assert.NoError(t, err)

	type dummyResponse struct {
		Foo int
	}

	nip47Response := &models.Response{
		ResultType: "dummy_method",
		Result: dummyResponse{
			Foo: 1000,
		},
	}

	nip47svc := NewNip47Service(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)

	res, err := nip47svc.CreateResponse(reqEvent, nip47Response, nostr.Tags{}, ss)
	assert.NoError(t, err)
	assert.Equal(t, reqPubkey, res.Tags.GetFirst([]string{"p"}).Value())
	assert.Equal(t, reqEvent.ID, res.Tags.GetFirst([]string{"e"}).Value())
	assert.Equal(t, svc.Keys.GetNostrPublicKey(), res.PubKey)

	decrypted, err := nip04.Decrypt(res.Content, ss)
	assert.NoError(t, err)
	unmarshalledResponse := models.Response{
		Result: &dummyResponse{},
	}

	err = json.Unmarshal([]byte(decrypted), &unmarshalledResponse)
	assert.NoError(t, err)
	assert.Equal(t, nip47Response.ResultType, unmarshalledResponse.ResultType)
	assert.Equal(t, nip47Response.Result, *unmarshalledResponse.Result.(*dummyResponse))
}
