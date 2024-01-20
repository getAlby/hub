package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
)

func AesGcmEncrypt(secretKey string, plaintext string) string {
	plaintextBytes := []byte(plaintext)

	aes, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(aes)
	if err != nil {
		panic(err.Error())
	}

	nonce := make([]byte, aesgcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		panic(err)
	}

	ciphertext := aesgcm.Seal(nonce, nonce, plaintextBytes, nil)

	return string(ciphertext)
}

func AesGcmDecrypt(secretKey string, ciphertext string) string {

	aes, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(aes)
	if err != nil {
		panic(err.Error())
	}

	nonceSize := aesgcm.NonceSize()
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := aesgcm.Open(nil, []byte(nonce), []byte(ciphertext), nil)
	if err != nil {
		panic(err)
	}

	return string(plaintext)
}
