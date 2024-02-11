package greenlight

type NodeInfo struct {
	ID    string `json:"id"`
	Alias string `json:"alias"`
	// ...other fields
}

type Invoice struct {
	Bolt11      string `json:"bolt11"`
	PaymentHash string `json:"payment_hash"`
	Preimage    string `json:"payment_secret"`
	ExpiresAt   int64  `json:"expires_at"`
	// ...other fields
}

type MsatValue struct {
	Msat int64 `json:"msat"`
}

type PayResponse struct {
	PaymentHash    string    `json:"payment_hash"`
	Preimage       string    `json:"payment_preimage"`
	CreatedAt      float64   `json:"created_at"`
	AmountMsat     MsatValue `json:"amount_msat"`
	AmountSentMsat MsatValue `json:"amount_sent_msat"`
	// ...other fields
}

type ListFundsResponse struct {
	Channels []Channel `json:"channels"`
	Outputs  []Output  `json:"outputs"`
}

type Output struct {
	TxId       string    `json:"txid"`
	AmountMsat MsatValue `json:"amount_msat"`
	Address    string    `json:"address"`
	// ...other fields
}

type Channel struct {
	PeerId        string    `json:"peer_id"`
	OurAmountMsat MsatValue `json:"our_amount_msat"`
	AmountMsat    MsatValue `json:"amount_msat"`
	FundingTxId   string    `json:"funding_txid"`
	Id            string    `json:"channel_id"`
	State         int       `json:"state"`
}

type ScheduleResponse struct {
	NodeId  string `json:"node_id"`
	GrpcUri string `json:"grpc_uri"`
}

type ConnectPeerResponse struct {
	Id string `json:"id"`
	// ...other fields
}

type Outpoint struct {
	//txid
}

type OpenChannelResponse struct {
	Tx        string `json:"tx"`
	TxId      string `json:"txid"`
	ChannelId string `json:"channel_id"`
	// ...other fields
}

type NewAddressResponse struct {
	Bech32 string `json:"bech32"`
}
