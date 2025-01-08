package cipher

import (
	"fmt"

	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/nbd-wtf/go-nostr/nip44"
)

const (
	SUPPORTED_VERSIONS = "1.0 0.0"
)

type Nip47Cipher struct {
	version         string
	pubkey          string
	privkey         string
	sharedSecret    []byte
	conversationKey [32]byte
}

func NewNip47Cipher(version, pubkey, privkey string) (*Nip47Cipher, error) {
	_, err := isVersionSupported(version)
	if err != nil {
		return nil, err
	}

	var ss []byte
	var ck [32]byte
	if version == "0.0" {
		ss, err = nip04.ComputeSharedSecret(pubkey, privkey)
		if err != nil {
			return nil, err
		}
	} else {
		ck, err = nip44.GenerateConversationKey(pubkey, privkey)
		if err != nil {
			return nil, err
		}
	}

	return &Nip47Cipher{
		version:         version,
		pubkey:          pubkey,
		privkey:         privkey,
		sharedSecret:    ss,
		conversationKey: ck,
	}, nil
}

func (c *Nip47Cipher) Encrypt(message string) (msg string, err error) {
	if c.version == "0.0" {
		msg, err = nip04.Encrypt(message, c.sharedSecret)
		if err != nil {
			return "", err
		}
	} else {
		msg, err = nip44.Encrypt(message, c.conversationKey)
		if err != nil {
			return "", err
		}
	}
	return msg, nil
}

func (c *Nip47Cipher) Decrypt(content string) (payload string, err error) {
	if c.version == "0.0" {
		payload, err = nip04.Decrypt(content, c.sharedSecret)
		if err != nil {
			return "", err
		}
	} else {
		payload, err = nip44.Decrypt(content, c.conversationKey)
		if err != nil {
			return "", err
		}
	}
	return payload, nil
}

func isVersionSupported(version string) (bool, error) {
	if version == "1.0" || version == "0.0" {
		return true, nil
	}

	return false, fmt.Errorf("invalid version: %s", version)
}
