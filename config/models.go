package config

const (
	LNDBackendType        = "LND"
	GreenlightBackendType = "GREENLIGHT"
	LDKBackendType        = "LDK"
	BreezBackendType      = "BREEZ"
	PhoenixBackendType    = "PHOENIX"
	CashuBackendType      = "CASHU"
)

type AppConfig struct {
	Relay                 string `envconfig:"RELAY" default:"wss://relay.getalby.com/v1"`
	LNBackendType         string `envconfig:"LN_BACKEND_TYPE"`
	LNDAddress            string `envconfig:"LND_ADDRESS"`
	LNDCertFile           string `envconfig:"LND_CERT_FILE"`
	LNDMacaroonFile       string `envconfig:"LND_MACAROON_FILE"`
	Workdir               string `envconfig:"WORK_DIR"`
	Port                  string `envconfig:"PORT" default:"8080"`
	DatabaseUri           string `envconfig:"DATABASE_URI" default:"nwc.db"`
	CookieSecret          string `envconfig:"COOKIE_SECRET"`
	LogLevel              string `envconfig:"LOG_LEVEL"`
	LDKNetwork            string `envconfig:"LDK_NETWORK" default:"bitcoin"`
	LDKEsploraServer      string `envconfig:"LDK_ESPLORA_SERVER" default:"https://electrs.albylabs.com"`
	LDKGossipSource       string `envconfig:"LDK_GOSSIP_SOURCE" default:"https://rapidsync.lightningdevkit.org/snapshot"`
	LDKLogLevel           string `envconfig:"LDK_LOG_LEVEL"`
	MempoolApi            string `envconfig:"MEMPOOL_API" default:"https://mempool.space/api"`
	AlbyAPIURL            string `envconfig:"ALBY_API_URL" default:"https://api.getalby.com"`
	AlbyClientId          string `envconfig:"ALBY_OAUTH_CLIENT_ID" default:"J2PbXS1yOf"`
	AlbyClientSecret      string `envconfig:"ALBY_OAUTH_CLIENT_SECRET" default:"rABK2n16IWjLTZ9M1uKU"`
	AlbyOAuthAuthUrl      string `envconfig:"ALBY_OAUTH_AUTH_URL" default:"https://getalby.com/oauth"`
	BaseUrl               string `envconfig:"BASE_URL" default:"http://localhost:8080"`
	FrontendUrl           string `envconfig:"FRONTEND_URL"`
	LogEvents             bool   `envconfig:"LOG_EVENTS" default:"false"`
	PhoenixdAddress       string `envconfig:"PHOENIXD_ADDRESS" default:"http://127.0.0.1:9740"`
	PhoenixdAuthorization string `envconfig:"PHOENIXD_AUTHORIZATION"`
	GoProfilerAddr        string `envconfig:"GO_PROFILER_ADDR"`
	DdProfilerEnabled     bool   `envconfig:"DD_PROFILER_ENABLED" default:"false"`
}

func (c *AppConfig) IsDefaultClientId() bool {
	return c.AlbyClientId == "J2PbXS1yOf"
}

type Config interface {
	Get(key string, encryptionKey string) (string, error)
	SetIgnore(key string, value string, encryptionKey string)
	SetUpdate(key string, value string, encryptionKey string)
	GetNostrPublicKey() string
	GetNostrSecretKey() string
	GetCookieSecret() string
	GetRelayUrl() string
	GetEnv() *AppConfig
	CheckUnlockPassword(password string) bool
	ChangeUnlockPassword(currentUnlockPassword string, newUnlockPassword string) error
	Setup(encryptionKey string)
	Start(encryptionKey string) error
}
