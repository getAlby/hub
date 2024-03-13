package api

type SendPaymentProbesRequest struct {
	Invoice string `json:"invoice"`
}

type SendPaymentProbesResponse struct {
	Error string `json:"error"`
}

type SendSpontaneousPaymentProbesRequest struct {
	Amount uint64 `json:"amount"`
	NodeID string `json:"nodeID"`
}

type SendSpontaneousPaymentProbesResponse struct {
	Error string `json:"error"`
}
