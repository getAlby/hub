package nip47

import (
	"context"
	"testing"

	"github.com/getAlby/go-nostr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/getAlby/hub/alby"
	"github.com/getAlby/hub/tests"
)

// When an app connection has been deleted, publishing its NIP47 info must
// surface gorm.ErrRecordNotFound so the publish queue can drop the item
// instead of retrying forever.
func TestPublishNip47Info_AppNotFound(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	albyOAuthSvc := alby.NewAlbyOAuthService(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher)
	nip47svc := NewNip47Service(svc.DB, svc.Cfg, svc.Keys, svc.EventPublisher, albyOAuthSvc)

	walletPrivKey := nostr.GeneratePrivateKey()
	walletPubKey, err := nostr.GetPublicKey(walletPrivKey)
	require.NoError(t, err)

	// app id 9999999 does not exist; the DB lookup fails before the relay pool
	// is ever used, so a nil pool/lnClient is fine here.
	_, err = nip47svc.PublishNip47Info(context.Background(), nil, 9999999, walletPubKey, walletPrivKey, "wss://relay.example.com", nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
