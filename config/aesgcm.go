package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"strings"

	"golang.org/x/crypto/argon2"
)

func DeriveKey(password string, salt []byte) ([]byte, []byte, error) {
	if salt == nil {
		salt = make([]byte, 32)
		if _, err := rand.Read(salt); err != nil {
			return nil, nil, err
		}
	}

	key := argon2.Key([]byte(password), salt, 3, 32*1024, 1, 32)

	return key, salt, nil
}

func AesGcmEncrypt(plaintext string, password string) (string, error) {
	secretKey, salt, err := DeriveKey(password, nil)
	if err != nil {
		return "", err
	}
	plaintextBytes := []byte(plaintext)

	aes, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(aes)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesgcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nil, nonce, plaintextBytes, nil)

	return hex.EncodeToString(salt) + "-" + hex.EncodeToString(nonce) + "-" + hex.EncodeToString(ciphertext), nil
}

func AesGcmDecrypt(ciphertext string, password string) (string, error) {
	arr := strings.Split(ciphertext, "-")
	salt, _ := hex.DecodeString(arr[0])
	nonce, _ := hex.DecodeString(arr[1])
	data, _ := hex.DecodeString(arr[2])

	secretKey, _, err := DeriveKey(password, salt)
	if err != nil {
		return "", err
	}
	aes, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(aes)
	if err != nil {
		return "", err
	}

	plaintext, err := aesgcm.Open(nil, nonce, data, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
