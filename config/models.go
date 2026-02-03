package config

const (
	LNDBackendType     = "LND"
	LDKBackendType     = "LDK"
	PhoenixBackendType = "PHOENIX"
	CashuBackendType   = "CASHU"
	CLNBackendType     = "CLN"
)

const (
	OnchainAddressKey           = "OnchainAddress"
	AutoSwapBalanceThresholdKey = "AutoSwapBalanceThreshold"
	AutoSwapAmountKey           = "AutoSwapAmount"
	AutoSwapDestinationKey      = "AutoSwapDestination"
	AutoSwapXpubIndexStart      = "AutoSwapXpubIndexStart"
)

type AppConfig struct {
	Relay                              string `envconfig:"RELAY" default:"wss://relay.getalby.com/v1"`
	LNBackendType                      string `envconfig:"LN_BACKEND_TYPE"`
	LNDAddress                         string `envconfig:"LND_ADDRESS"`
	LNDCertFile                        string `envconfig:"LND_CERT_FILE"`
	LNDMacaroonFile                    string `envconfig:"LND_MACAROON_FILE"`
	Workdir                            string `envconfig:"WORK_DIR"`
	Port                               string `envconfig:"PORT" default:"8080"`
	DatabaseUri                        string `envconfig:"DATABASE_URI" default:"nwc.db"`
	LogLevel                           string `envconfig:"LOG_LEVEL" default:"4"`
	LogToFile                          bool   `envconfig:"LOG_TO_FILE" default:"true"`
	Network                            string `envconfig:"NETWORK"`
	LDKNetwork                         string `envconfig:"LDK_NETWORK"`
	LDKEsploraServer                   string `envconfig:"LDK_ESPLORA_SERVER" default:"https://electrs.getalbypro.com"` // TODO: remove LDK prefix
	LDKElectrumServer                  string `envconfig:"LDK_ELECTRUM_SERVER"`
	LDKGossipSource                    string `envconfig:"LDK_GOSSIP_SOURCE"`
	LDKLogLevel                        string `envconfig:"LDK_LOG_LEVEL" default:"3"`
	LDKMaxChannelSaturationPowerOfHalf uint8  `envconfig:"LDK_MAX_CHANNEL_SATURATION" default:"2"`
	LDKMaxPathCount                    uint8  `envconfig:"LDK_MAX_PATH_COUNT" default:"5"`
	LDKChannelMonitorWarningSizeBytes  uint64 `envconfig:"LDK_CHANNEL_MONITOR_WARNING_SIZE_BYTES" default:"5000000"`
	LDKVssUrl                          string `envconfig:"LDK_VSS_URL" default:"https://vss.getalbypro.com/vss"`
	LDKListeningAddresses              string `envconfig:"LDK_LISTENING_ADDRESSES" default:"[::]:9735"`
	LDKAnnouncementAddresses           string `envconfig:"LDK_ANNOUNCEMENT_ADDRESSES"`
	LDKTransientNetworkGraph           bool   `envconfig:"LDK_TRANSIENT_NETWORK_GRAPH" default:"false"`
	RebalanceServiceUrl                string `envconfig:"REBALANCE_SERVICE_URL" default:"https://megalithic.me"`
	LDKBitcoindRpcHost                 string `envconfig:"LDK_BITCOIND_RPC_HOST"`
	LDKBitcoindRpcPort                 string `envconfig:"LDK_BITCOIND_RPC_PORT"`
	LDKBitcoindRpcUser                 string `envconfig:"LDK_BITCOIND_RPC_USER"`
	LDKBitcoindRpcPassword             string `envconfig:"LDK_BITCOIND_RPC_PASSWORD"`
	MempoolApi                         string `envconfig:"MEMPOOL_API" default:"https://mempool.space/api"`
	AlbyClientId                       string `envconfig:"ALBY_OAUTH_CLIENT_ID" default:"J2PbXS1yOf"`
	AlbyClientSecret                   string `envconfig:"ALBY_OAUTH_CLIENT_SECRET" default:"rABK2n16IWjLTZ9M1uKU"`
	BaseUrl                            string `envconfig:"BASE_URL"`
	FrontendUrl                        string `envconfig:"FRONTEND_URL"`
	SendEventsToAlby                   bool   `envconfig:"SEND_EVENTS_TO_ALBY" default:"true"`
	AutoLinkAlbyAccount                bool   `envconfig:"AUTO_LINK_ALBY_ACCOUNT" default:"true"`
	PhoenixdAddress                    string `envconfig:"PHOENIXD_ADDRESS"`
	PhoenixdAuthorization              string `envconfig:"PHOENIXD_AUTHORIZATION"`
	GoProfilerAddr                     string `envconfig:"GO_PROFILER_ADDR"`
	EnableAdvancedSetup                bool   `envconfig:"ENABLE_ADVANCED_SETUP" default:"true"`
	AutoUnlockPassword                 string `envconfig:"AUTO_UNLOCK_PASSWORD"`
	LogDBQueries                       bool   `envconfig:"LOG_DB_QUERIES" default:"false"`
	BoltzApi                           string `envconfig:"BOLTZ_API" default:"https://api.boltz.exchange"`
	HideUpdateBanner                   bool   `envconfig:"HIDE_UPDATE_BANNER" default:"false"`
	CLNAddress                         string `envconfig:"CLN_ADDRESS"`
	CLNLightningDir                    string `envconfig:"CLN_LIGHTNING_DIR"`
	CLNAddressHold                     string `envconfig:"CLN_ADDRESS_HOLD"`
}

func (c *AppConfig) IsDefaultClientId() bool {
	return c.AlbyClientId == "J2PbXS1yOf"
}

func (c *AppConfig) GetBaseFrontendUrl() string {
	url := c.FrontendUrl
	if url == "" {
		url = c.BaseUrl
	}
	return url
}

type Config interface {
	Unlock(encryptionKey string) error
	Get(key string, encryptionKey string) (string, error)
	SetIgnore(key string, value string, encryptionKey string) error
	SetUpdate(key string, value string, encryptionKey string) error
	GetJWTSecret() (string, error)
	GetRelayUrls() []string
	GetNetwork() string
	GetMempoolUrl() string
	GetEnv() *AppConfig
	CheckUnlockPassword(password string) bool
	ChangeUnlockPassword(currentUnlockPassword string, newUnlockPassword string) error
	SetAutoUnlockPassword(unlockPassword string) error
	SaveUnlockPasswordCheck(encryptionKey string) error
	SetupCompleted() (bool, error)
	GetCurrency() string
	SetCurrency(value string) error
	GetBitcoinDisplayFormat() string
	SetBitcoinDisplayFormat(value string) error
}
