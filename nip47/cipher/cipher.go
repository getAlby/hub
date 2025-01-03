package cipher

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nbd-wtf/go-nostr/nip04"
	"github.com/nbd-wtf/go-nostr/nip44"
)

const (
	SUPPORTED_VERSIONS = "1.0 0.0"
)

type Nip47Cipher struct {
	Version         string
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
		Version:         version,
		pubkey:          pubkey,
		privkey:         privkey,
		sharedSecret:    ss,
		conversationKey: ck,
	}, nil
}

func (c *Nip47Cipher) Encrypt(message string) (msg string, err error) {
	if c.Version == "0.0" {
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
	if c.Version == "0.0" {
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
	versionParts := strings.Split(version, ".")
	if len(versionParts) != 2 {
		return false, fmt.Errorf("invalid version format: %s", version)
	}

	major, err := strconv.Atoi(versionParts[0])
	if err != nil {
		return false, fmt.Errorf("invalid major version: %s", versionParts[0])
	}
	minor, err := strconv.Atoi(versionParts[1])
	if err != nil {
		return false, fmt.Errorf("invalid minor version: %s", versionParts[1])
	}

	for _, supported := range strings.Split(SUPPORTED_VERSIONS, " ") {
		supportedParts := strings.Split(supported, ".")
		if len(supportedParts) != 2 {
			continue
		}

		supportedMajor, _ := strconv.Atoi(supportedParts[0])
		supportedMinor, _ := strconv.Atoi(supportedParts[1])

		if major == supportedMajor {
			if minor <= supportedMinor {
				return true, nil
			}
			return false, fmt.Errorf("invalid version: %s", version)
		}
	}
	return false, fmt.Errorf("invalid version: %s", version)
}
