// TODO: move to lsp/models.go
package lsp

type LSP struct {
	Pubkey  string
	Url     string
	LspType string
}

const (
	LSP_TYPE_FLOW_2_0 = "Flow 2.0"
	LSP_TYPE_PMLSP    = "PMLSP"
	LSP_TYPE_LSPS1    = "LSPS1"
)

type NewInstantChannelResponse struct {
	FeeAmountMsat uint64 `json:"fee_amount_msat"`
	Invoice       string `json:"invoice"`
}

func VoltageLSP() LSP {
	lsp := LSP{
		Pubkey:  "03aefa43fbb4009b21a4129d05953974b7dbabbbfb511921410080860fca8ee1f0",
		Url:     "https://lsp.voltageapi.com/api/v1",
		LspType: LSP_TYPE_FLOW_2_0,
	}
	return lsp
}

func OlympusMutinynetFlowLSP() LSP {
	lsp := LSP{
		Pubkey:  "032ae843e4d7d177f151d021ac8044b0636ec72b1ce3ffcde5c04748db2517ab03",
		Url:     "https://mutinynet-flow.lnolymp.us/api/v1",
		LspType: LSP_TYPE_FLOW_2_0,
	}
	return lsp
}

func OlympusMutinynetLSPS1LSP() LSP {
	lsp := LSP{
		Pubkey:  "032ae843e4d7d177f151d021ac8044b0636ec72b1ce3ffcde5c04748db2517ab03",
		Url:     "https://mutinynet-lsps1.lnolymp.us/api/v1",
		LspType: LSP_TYPE_LSPS1,
	}
	return lsp
}

func OlympusLSP() LSP {
	lsp := LSP{
		Pubkey:  "031b301307574bbe9b9ac7b79cbe1700e31e544513eae0b5d7497483083f99e581",
		Url:     "https://0conf.lnolymp.us/api/v1",
		LspType: LSP_TYPE_FLOW_2_0,
	}
	return lsp
}

func AlbyPlebsLSP() LSP {
	lsp := LSP{
		Pubkey:  "029ca15ad2ea3077f5f0524c4c9bc266854c14b9fc81b9cc3d6b48e2460af13f65",
		Url:     "https://lsp.albylabs.com",
		LspType: LSP_TYPE_PMLSP,
	}
	return lsp
}
func AlbyMutinynetPlebsLSP() LSP {
	lsp := LSP{
		Pubkey:  "02f7029c14f3d805843e065d42e9bdc57f5f414249f335906bbe282ff99b2be17a",
		Url:     "https://lsp-mutinynet.albylabs.com",
		LspType: LSP_TYPE_PMLSP,
	}
	return lsp
}
