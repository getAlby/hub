package config

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

func TestDecryptExistingCiphertextWithPassword(t *testing.T) {
	value, err := AesGcmDecryptWithPassword("323f41394d3175b72454ccae9c0081f94df5fb4c2fb0b9283a87e5aafba81839-c335b9eeea75c28a6f823354-5055b90dadbdd01c52fbdbb7efb80609e4410357481651e89ceb1501c8e1dea1f33a8e3322a1cef4f641773667423bca5154dfeccac390cfcd719b36965adc3e6ae56fd5d6c82819596e9ef4ff07193ae345eb291fa412a1ce6066864b", "123")
	assert.NoError(t, err)
	assert.Equal(t, "connect maximum march lava ignore resist visa kind kiwi kidney develop animal", value)
}

func TestEncryptDecryptWithPassword(t *testing.T) {
	mnemonic := "connect maximum march lava ignore resist visa kind kiwi kidney develop animal"
	encrypted, err := AesGcmEncryptWithPassword(mnemonic, "123")
	assert.NoError(t, err)
	value, err := AesGcmDecryptWithPassword(encrypted, "123")
	assert.NoError(t, err)
	assert.Equal(t, mnemonic, value)
}

func TestDecryptExistingCiphertextWithKey(t *testing.T) {
	mnemonic := "connect maximum march lava ignore resist visa kind kiwi kidney develop animal"
	masterKey, err := bip32.NewMasterKey(bip39.NewSeed(mnemonic, ""))
	assert.NoError(t, err)
	value, err := AesGcmDecryptWithKey("22ad485dea4f49696594c7c4-afe35ce65fc5a45249bf1b9078472fb28395fc88c30a79c76c7d8d37cf", masterKey.Key)
	assert.NoError(t, err)
	assert.Equal(t, "Hello, world!", value)
}

func TestEncryptDecryptWithKey(t *testing.T) {
	plaintext := "Hello, world!"
	mnemonic := "connect maximum march lava ignore resist visa kind kiwi kidney develop animal"
	masterKey, err := bip32.NewMasterKey(bip39.NewSeed(mnemonic, ""))
	assert.NoError(t, err)

	assert.Equal(t, "409e902eafba273b21dff921f0eb4bec6cbb0b657fdce8d245ca78d2920f8b73", hex.EncodeToString(masterKey.Key))

	encrypted, err := AesGcmEncryptWithKey(plaintext, masterKey.Key)
	assert.NoError(t, err)
	value, err := AesGcmDecryptWithKey(encrypted, masterKey.Key)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, value)
}
