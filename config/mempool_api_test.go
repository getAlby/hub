package config

import "testing"

func TestMempoolApiAndUrl(t *testing.T) {
	const (
		defaultAPI   = "https://mempool.space/api"
		mainnetURL   = "https://mempool.space"
		testnetAPI   = "https://mempool.space/testnet/api"
		testnetURL   = "https://mempool.space/testnet"
		signetAPI    = "https://mempool.space/signet/api"
		signetURL    = "https://mempool.space/signet"
		mutinynetAPI = "https://mutinynet.com/api"
		mutinynetURL = "https://mutinynet.com"
		customAPI    = "https://mempool.example/api"
		customURL    = "https://mempool.example"
	)

	tests := []struct {
		name    string
		env     AppConfig
		wantAPI string
		wantURL string
	}{
		{
			name:    "mainnet default",
			env:     AppConfig{Network: "bitcoin", MempoolApi: defaultAPI},
			wantAPI: defaultAPI,
			wantURL: mainnetURL,
		},
		{
			name:    "testnet default uses testnet API and explorer",
			env:     AppConfig{Network: "testnet", MempoolApi: defaultAPI},
			wantAPI: testnetAPI,
			wantURL: testnetURL,
		},
		{
			name:    "signet default uses standard signet, not Mutinynet",
			env:     AppConfig{Network: "signet", MempoolApi: defaultAPI},
			wantAPI: signetAPI,
			wantURL: signetURL,
		},
		{
			name: "signet with Mutinynet esplora",
			env: AppConfig{
				Network:       "signet",
				MempoolApi:    defaultAPI,
				EsploraServer: "https://mutinynet.com/api",
			},
			wantAPI: mutinynetAPI,
			wantURL: mutinynetURL,
		},
		{
			name: "signet with Mutinynet electrum",
			env: AppConfig{
				Network:        "signet",
				MempoolApi:     defaultAPI,
				ElectrumServer: "electrum.mutinynet.com:50001",
			},
			wantAPI: mutinynetAPI,
			wantURL: mutinynetURL,
		},
		{
			name: "signet with Mutinynet gossip source",
			env: AppConfig{
				Network:         "signet",
				MempoolApi:      defaultAPI,
				LDKGossipSource: "https://rgs.mutinynet.com/snapshot",
			},
			wantAPI: mutinynetAPI,
			wantURL: mutinynetURL,
		},
		{
			name: "signet with deprecated LDK esplora pointing at Mutinynet",
			env: AppConfig{
				Network:          "signet",
				MempoolApi:       defaultAPI,
				LDKEsploraServer: "https://mutinynet.com/api",
			},
			wantAPI: mutinynetAPI,
			wantURL: mutinynetURL,
		},
		{
			name: "signet esplora takes precedence over a leftover Mutinynet electrum",
			env: AppConfig{
				Network:        "signet",
				MempoolApi:     defaultAPI,
				EsploraServer:  "https://mempool.space/signet/api",
				ElectrumServer: "electrum.mutinynet.com:50001",
			},
			wantAPI: signetAPI,
			wantURL: signetURL,
		},
		{
			name: "deprecated LDK esplora takes precedence over a Mutinynet esplora",
			env: AppConfig{
				Network:          "signet",
				MempoolApi:       defaultAPI,
				EsploraServer:    "https://mutinynet.com/api",
				LDKEsploraServer: "https://mempool.space/signet/api",
			},
			wantAPI: signetAPI,
			wantURL: signetURL,
		},
		{
			name: "Mutinynet gossip source counts even with a standard signet esplora",
			env: AppConfig{
				Network:         "signet",
				MempoolApi:      defaultAPI,
				EsploraServer:   "https://mempool.space/signet/api",
				LDKGossipSource: "https://rgs.mutinynet.com/snapshot",
			},
			wantAPI: mutinynetAPI,
			wantURL: mutinynetURL,
		},
		{
			name: "explicit Mutinynet MEMPOOL_API is kept",
			env: AppConfig{
				Network:    "signet",
				MempoolApi: mutinynetAPI,
			},
			wantAPI: mutinynetAPI,
			wantURL: mutinynetURL,
		},
		{
			name: "explicit signet MEMPOOL_API is kept even with a Mutinynet esplora",
			env: AppConfig{
				Network:       "signet",
				MempoolApi:    signetAPI,
				EsploraServer: "https://mutinynet.com/api",
			},
			wantAPI: signetAPI,
			wantURL: signetURL,
		},
		{
			name: "explicit custom override is kept on testnet",
			env: AppConfig{
				Network:    "testnet",
				MempoolApi: customAPI,
			},
			wantAPI: customAPI,
			wantURL: customURL,
		},
		{
			name:    "empty MEMPOOL_API is left unchanged",
			env:     AppConfig{Network: "testnet"},
			wantAPI: "",
			wantURL: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := tt.env
			if got := MempoolApi(&env); got != tt.wantAPI {
				t.Fatalf("MempoolApi() = %q, want %q", got, tt.wantAPI)
			}
			cfg := &config{Env: &env}
			if got := cfg.GetMempoolUrl(); got != tt.wantURL {
				t.Fatalf("GetMempoolUrl() = %q, want %q", got, tt.wantURL)
			}
		})
	}
}
