package lsp

type LSP struct {
	Pubkey string
}

const (
	LSP_TYPE_LSPS1 = "LSPS1"
)

func OlympusMutinynetLSP() LSP {
	lsp := LSP{
		Pubkey: "032ae843e4d7d177f151d021ac8044b0636ec72b1ce3ffcde5c04748db2517ab03",
	}
	return lsp
}

func OlympusLSP() LSP {
	lsp := LSP{
		Pubkey: "031b301307574bbe9b9ac7b79cbe1700e31e544513eae0b5d7497483083f99e581",
	}
	return lsp
}

func MegalithMutinynetLSP() LSP {
	lsp := LSP{
		Pubkey: "03e30fda71887a916ef5548a4d02b06fe04aaa1a8de9e24134ce7f139cf79d7579",
	}
	return lsp
}

func MegalithLSP() LSP {
	lsp := LSP{
		Pubkey: "038a9e56512ec98da2b5789761f7af8f280baf98a09282360cd6ff1381b5e889bf",
	}
	return lsp
}
