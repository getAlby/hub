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
	Msat uint `json:"msat"`
}

type PayResponse struct {
	PaymentHash    string    `json:"payment_hash"`
	Preimage       string    `json:"payment_preimage"`
	CreatedAt      float64   `json:"created_at"`
	AmountMsat     MsatValue `json:"amount_msat"`
	AmountSentMsat MsatValue `json:"amount_sent_msat"`
	// ...other fields
}
