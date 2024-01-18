package db

type Config struct {
	ID                   int    // primary key, always 1
	LNBackendType        string `envconfig:"LN_BACKEND_TYPE"`
	LNDAddress           string `envconfig:"LND_ADDRESS"`
	LNDCertFile          string `envconfig:"LND_CERT_FILE"`
	LNDCertHex           string `envconfig:"LND_CERT_HEX"`
	LNDMacaroonFile      string `envconfig:"LND_MACAROON_FILE"`
	LNDMacaroonHex       string `envconfig:"LND_MACAROON_HEX"`
	BreezMnemonic        string `envconfig:"BREEZ_MNEMONIC"`
	GreenlightInviteCode string `envconfig:"GREENLIGHT_INVITE_CODE"`
}
