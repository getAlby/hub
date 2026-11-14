package cipher

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/getAlby/go-nostr"
	"github.com/getAlby/hub/constants"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCipher(t *testing.T) {
	doTestCipher(t, constants.ENCRYPTION_TYPE_NIP04)
	doTestCipher(t, constants.ENCRYPTION_TYPE_NIP44_V2)
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

func TestCipher_Nip04MalformedContent(t *testing.T) {
	appPrivateKey := nostr.GeneratePrivateKey()
	appPubkey, err := nostr.GetPublicKey(appPrivateKey)
	require.NoError(t, err)
	hubPrivateKey := nostr.GeneratePrivateKey()
	hubPubkey, err := nostr.GetPublicKey(hubPrivateKey)
	require.NoError(t, err)

	appCipher, err := NewNip47Cipher(constants.ENCRYPTION_TYPE_NIP04, hubPubkey, appPrivateKey)
	require.NoError(t, err)
	hubCipher, err := NewNip47Cipher(constants.ENCRYPTION_TYPE_NIP04, appPubkey, hubPrivateKey)
	require.NoError(t, err)

	validContent, err := appCipher.Encrypt(`{"method":"get_info"}`)
	require.NoError(t, err)
	parts := strings.Split(validContent, "?iv=")
	require.Len(t, parts, 2)
	ciphertext, iv := parts[0], parts[1]

	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	require.NoError(t, err)
	truncatedCiphertext := base64.StdEncoding.EncodeToString(ciphertextBytes[:len(ciphertextBytes)-1])

	testCases := []struct {
		name          string
		content       string
		expectedError string
	}{
		{
			name:          "iv too short",
			content:       ciphertext + "?iv=" + base64.StdEncoding.EncodeToString(make([]byte, 8)),
			expectedError: "invalid iv length: 8",
		},
		{
			name:          "ciphertext not a multiple of 16 bytes",
			content:       truncatedCiphertext + "?iv=" + iv,
			expectedError: fmt.Sprintf("invalid ciphertext length: %d", len(ciphertextBytes)-1),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NotPanics(t, func() {
				_, err := hubCipher.Decrypt(tc.content)
				require.EqualError(t, err, tc.expectedError)
			})
		})
	}
}
