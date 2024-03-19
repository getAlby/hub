package lsp

type LSP struct {
	Pubkey string
	Url    string
}

type LSPConnectionMethod struct {
	Address string `json:"address"`
	Port    uint16 `json:"port"`
	Type    string `json:"type"`
}

type FeeRequest struct {
	AmountMsat uint64 `json:"amount_msat"`
	Pubkey     string `json:"pubkey"`
}
type FeeResponse struct {
	FeeAmountMsat uint64 `json:"fee_amount_msat"`
	Id            string `json:"id"`
}

type ProposalRequest struct {
	Bolt11 string `json:"bolt11"`
	FeeId  string `json:"fee_id"`
}
type ProposalResponse struct {
	Bolt11 string `json:"jit_bolt11"`
}

type LSPInfo struct {
	Pubkey            string                `json:"pubkey"`
	ConnectionMethods []LSPConnectionMethod `json:"connection_methods"`
}

func VoltageLSP() LSP {
	lsp := LSP{
		Pubkey: "03aefa43fbb4009b21a4129d05953974b7dbabbbfb511921410080860fca8ee1f0",
		Url:    "https://lsp.voltageapi.com/api/v1",
	}
	return lsp
}

func OlympusLSP() LSP {
	lsp := LSP{
		Pubkey: "031b301307574bbe9b9ac7b79cbe1700e31e544513eae0b5d7497483083f99e581",
		Url:    "https://0conf.lnolymp.us/api/v1",
	}
	return lsp
}
