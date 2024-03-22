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
	Workdir          string `envconfig:"WORK_DIR" default:".data"`
	Port             string `envconfig:"PORT" default:"8080"`
	DatabaseUri      string `envconfig:"DATABASE_URI" default:".data/nwc.db"`
	CookieSecret     string `envconfig:"COOKIE_SECRET"`
	LogLevel         string `envconfig:"LOG_LEVEL"`
	LDKNetwork       string `envconfig:"LDK_NETWORK" default:"bitcoin"`
	LDKEsploraServer string `envconfig:"LDK_ESPLORA_SERVER" default:"https://blockstream.info/api"`
	LDKGossipSource  string `envconfig:"LDK_GOSSIP_SOURCE" default:"https://rapidsync.lightningdevkit.org/snapshot"`
	MempoolApi       string `envconfig:"MEMPOOL_API" default:"https://mempool.space/api"`
	LDKLogLevel      string `envconfig:"LDK_LOG_LEVEL"`
}
