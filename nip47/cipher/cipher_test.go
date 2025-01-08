package cipher

import (
	"fmt"
	"testing"

	"github.com/nbd-wtf/go-nostr"

	"github.com/stretchr/testify/assert"
)

func TestCipher(t *testing.T) {
	doTestCipher(t, "0.0")
	doTestCipher(t, "1.0")
}

func doTestCipher(t *testing.T, version string) {
	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)

	nip47Cipher, err := NewNip47Cipher(version, reqPubkey, reqPrivateKey)
	assert.NoError(t, err)

	payload := "test payload"
	msg, err := nip47Cipher.Encrypt(payload)
	assert.NoError(t, err)

	decrypted, err := nip47Cipher.Decrypt(msg)
	assert.Equal(t, payload, decrypted)
}

func TestCipher_UnsupportedVersions(t *testing.T) {
	doTestCipher_UnsupportedVersions(t, "1")
	doTestCipher_UnsupportedVersions(t, "x.1")
	doTestCipher_UnsupportedVersions(t, "1.x")
	doTestCipher_UnsupportedVersions(t, "2.0")
	doTestCipher_UnsupportedVersions(t, "0.5")
}

func doTestCipher_UnsupportedVersions(t *testing.T, version string) {
	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)

	_, err = NewNip47Cipher(version, reqPubkey, reqPrivateKey)
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf("invalid version: %s", version), err.Error())
}
