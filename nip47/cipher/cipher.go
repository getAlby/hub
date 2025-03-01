package cipher

import (
	"fmt"

	"github.com/getAlby/hub/constants"
	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/nbd-wtf/go-nostr/nip44"
)

const (
	SUPPORTED_VERSIONS    = "0.0 1.0"
	SUPPORTED_ENCRYPTIONS = "nip44_v2 nip04"
)

type Nip47Cipher struct {
	encryption      string
	pubkey          string
	privkey         string
	sharedSecret    []byte
	conversationKey [32]byte
}

func NewNip47Cipher(encryption, pubkey, privkey string) (*Nip47Cipher, error) {
	_, err := isEncryptionSupported(encryption)
	if err != nil {
		return nil, err
	}

	var ss []byte
	var ck [32]byte
	if encryption == constants.ENCRYPTION_TYPE_NIP04 {
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
		encryption:      encryption,
		pubkey:          pubkey,
		privkey:         privkey,
		sharedSecret:    ss,
		conversationKey: ck,
	}, nil
}

func (c *Nip47Cipher) Encrypt(message string) (msg string, err error) {
	if c.encryption == constants.ENCRYPTION_TYPE_NIP04 {
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
	if c.encryption == constants.ENCRYPTION_TYPE_NIP04 {
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

func isEncryptionSupported(encryption string) (bool, error) {
	if encryption == constants.ENCRYPTION_TYPE_NIP44_V2 || encryption == constants.ENCRYPTION_TYPE_NIP04 {
		return true, nil
	}

	return false, fmt.Errorf("invalid encryption: %s", encryption)
}
