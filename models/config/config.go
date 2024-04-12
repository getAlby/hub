// TODO: move to config/models.go
package config

const (
	LNDBackendType        = "LND"
	GreenlightBackendType = "GREENLIGHT"
	LDKBackendType        = "LDK"
	BreezBackendType      = "BREEZ"
)

type AppConfig struct {
	Relay            string `envconfig:"RELAY" default:"wss://relay.getalby.com/v1"`
	LNBackendType    string `envconfig:"LN_BACKEND_TYPE"`
	LNDAddress       string `envconfig:"LND_ADDRESS"`
	LNDCertFile      string `envconfig:"LND_CERT_FILE"`
	LNDMacaroonFile  string `envconfig:"LND_MACAROON_FILE"`
	Workdir          string `envconfig:"WORK_DIR"`
	Port             string `envconfig:"PORT" default:"8080"`
	DatabaseUri      string `envconfig:"DATABASE_URI" default:"nwc.db"`
	CookieSecret     string `envconfig:"COOKIE_SECRET"`
	LogLevel         string `envconfig:"LOG_LEVEL"`
	LDKNetwork       string `envconfig:"LDK_NETWORK" default:"bitcoin"`
	LDKEsploraServer string `envconfig:"LDK_ESPLORA_SERVER" default:"https://blockstream.info/api"`
	LDKGossipSource  string `envconfig:"LDK_GOSSIP_SOURCE" default:"https://rapidsync.lightningdevkit.org/snapshot"`
	LDKLogLevel      string `envconfig:"LDK_LOG_LEVEL"`
	MempoolApi       string `envconfig:"MEMPOOL_API" default:"https://mempool.space/api"`
	AlbyAPIURL       string `envconfig:"ALBY_API_URL" default:"https://api.getalby.com"`
	AlbyClientId     string `envconfig:"ALBY_OAUTH_CLIENT_ID"`
	AlbyClientSecret string `envconfig:"ALBY_OAUTH_CLIENT_SECRET"`
	AlbyOAuthAuthUrl string `envconfig:"ALBY_OAUTH_AUTH_URL" default:"https://getalby.com/oauth"`
	BaseUrl          string `envconfig:"BASE_URL" default:"http://localhost:8080"`
	FrontendUrl      string `envconfig:"FRONTEND_URL"`
	LogEvents        bool   `envconfig:"LOG_EVENTS" default:"false"`
}

type Config interface {
	Get(key string, encryptionKey string) (string, error)
	SetIgnore(key string, value string, encryptionKey string)
	SetUpdate(key string, value string, encryptionKey string)
	GetNostrPublicKey() string
}
