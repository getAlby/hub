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
	reqPrivateKey := nostr.GeneratePrivateKey()
	reqPubkey, err := nostr.GetPublicKey(reqPrivateKey)

	version := "1"
	_, err = NewNip47Cipher(version, reqPubkey, reqPrivateKey)
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf("invalid version format: %s", version), err.Error())

	version = "x.1"
	_, err = NewNip47Cipher(version, reqPubkey, reqPrivateKey)
	assert.Error(t, err)
	assert.Equal(t, "invalid major version: x", err.Error())

	version = "1.x"
	_, err = NewNip47Cipher(version, reqPubkey, reqPrivateKey)
	assert.Error(t, err)
	assert.Equal(t, "invalid minor version: x", err.Error())

	version = "2.0"
	_, err = NewNip47Cipher(version, reqPubkey, reqPrivateKey)
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf("invalid version: %s", version), err.Error())

	version = "0.5"
	_, err = NewNip47Cipher(version, reqPubkey, reqPrivateKey)
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf("invalid version: %s", version), err.Error())
}
