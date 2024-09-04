package config

const (
	LNDBackendType        = "LND"
	GreenlightBackendType = "GREENLIGHT"
	LDKBackendType        = "LDK"
	BreezBackendType      = "BREEZ"
	PhoenixBackendType    = "PHOENIX"
	CashuBackendType      = "CASHU"
)

const (
	OnchainAddressKey = "OnchainAddress"
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
	JWTSecret             string `envconfig:"JWT_SECRET"`
	LogLevel              string `envconfig:"LOG_LEVEL" default:"4"`
	LDKNetwork            string `envconfig:"LDK_NETWORK" default:"bitcoin"`
	LDKEsploraServer      string `envconfig:"LDK_ESPLORA_SERVER" default:"https://electrs.getalbypro.com"` // TODO: remove LDK prefix
	LDKGossipSource       string `envconfig:"LDK_GOSSIP_SOURCE"`
	LDKLogLevel           string `envconfig:"LDK_LOG_LEVEL" default:"3"`
	MempoolApi            string `envconfig:"MEMPOOL_API" default:"https://mempool.space/api"`
	AlbyAPIURL            string `envconfig:"ALBY_API_URL" default:"https://api.getalby.com"`
	AlbyClientId          string `envconfig:"ALBY_OAUTH_CLIENT_ID" default:"J2PbXS1yOf"`
	AlbyClientSecret      string `envconfig:"ALBY_OAUTH_CLIENT_SECRET" default:"rABK2n16IWjLTZ9M1uKU"`
	AlbyOAuthAuthUrl      string `envconfig:"ALBY_OAUTH_AUTH_URL" default:"https://getalby.com/oauth"`
	BaseUrl               string `envconfig:"BASE_URL"`
	FrontendUrl           string `envconfig:"FRONTEND_URL"`
	LogEvents             bool   `envconfig:"LOG_EVENTS" default:"true"`
	AutoLinkAlbyAccount   bool   `envconfig:"AUTO_LINK_ALBY_ACCOUNT" default:"true"`
	PhoenixdAddress       string `envconfig:"PHOENIXD_ADDRESS"`
	PhoenixdAuthorization string `envconfig:"PHOENIXD_AUTHORIZATION"`
	GoProfilerAddr        string `envconfig:"GO_PROFILER_ADDR"`
	DdProfilerEnabled     bool   `envconfig:"DD_PROFILER_ENABLED" default:"false"`
	EnableAdvancedSetup   bool   `envconfig:"ENABLE_ADVANCED_SETUP" default:"true"`
	AutoUnlockPassword    string `envconfig:"AUTO_UNLOCK_PASSWORD"`
}

func (c *AppConfig) IsDefaultClientId() bool {
	return c.AlbyClientId == "J2PbXS1yOf"
}

type Config interface {
	Get(key string, encryptionKey string) (string, error)
	SetIgnore(key string, value string, encryptionKey string)
	SetUpdate(key string, value string, encryptionKey string)
	GetJWTSecret() string
	GetRelayUrl() string
	GetEnv() *AppConfig
	CheckUnlockPassword(password string) bool
	ChangeUnlockPassword(currentUnlockPassword string, newUnlockPassword string) error
	Setup(encryptionKey string)
	SetupCompleted() bool
}
