package cipher

import (
	"fmt"
	"testing"

	"github.com/nbd-wtf/go-nostr"

	"github.com/stretchr/testify/assert"
)

func TestCipher(t *testing.T) {
	doTestCipher(t, "nip04")
	doTestCipher(t, "nip44_v2")
}

func doTestCipher(t *testing.T, encryption string) {
	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)

	nip47Cipher, err := NewNip47Cipher(encryption, reqPubkey, reqPrivateKey)
	assert.NoError(t, err)

	payload := "test payload"
	msg, err := nip47Cipher.Encrypt(payload)
	assert.NoError(t, err)

	decrypted, err := nip47Cipher.Decrypt(msg)
	assert.Equal(t, payload, decrypted)
}

func TestCipher_UnsupportedEncrptions(t *testing.T) {
	doTestCipher_UnsupportedEncrptions(t, "nip44")
	doTestCipher_UnsupportedEncrptions(t, "nip44_v0")
	doTestCipher_UnsupportedEncrptions(t, "nip44_v1")
	doTestCipher_UnsupportedEncrptions(t, "nip44v2")
	doTestCipher_UnsupportedEncrptions(t, "nip-44")
}

func doTestCipher_UnsupportedEncrptions(t *testing.T, encryption string) {
	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)

	_, err = NewNip47Cipher(encryption, reqPubkey, reqPrivateKey)
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf("invalid encryption: %s", encryption), err.Error())
}
