package db

import "time"

type Config struct {
	ID                   int    // primary key, always 1
	LNBackendType        string `envconfig:"LN_BACKEND_TYPE"`
	LNDAddress           string `envconfig:"LND_ADDRESS"`
	LNDCertHex           string `envconfig:"LND_CERT_HEX"`
	LNDMacaroonHex       string `envconfig:"LND_MACAROON_HEX"`
	BreezMnemonic        string `envconfig:"BREEZ_MNEMONIC"`
	BreezAPIKey          string `envconfig:"BREEZ_API_KEY"`
	GreenlightInviteCode string `envconfig:"GREENLIGHT_INVITE_CODE"`
	NostrSecretKey       string `envconfig:"NOSTR_PRIVKEY"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
