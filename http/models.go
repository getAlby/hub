package http

import "encoding/json"

type ErrorResponse struct {
	Message string `json:"message"`
}

type Lud6Response struct {
	Tag            string                 `json:"tag" default:"payRequest" validate:"required"`
	Callback       string                 `json:"callback" validate:"required"`
	MinSendable    uint64                 `json:"minSendable" default:"1000" validate:"required"`
	MaxSendable    uint64                 `json:"maxSendable" default:"10000000000" validate:"required"`
	CommentAllowed uint                   `json:"commentAllowed,omitempty"`
	Metadata       string                 `json:"metadata,omitempty"`
	PayerData      map[string]interface{} `json:"payerData,omitempty"`
	NostrPubkey    string                 `json:"nostrPubkey,omitempty"`
	AllowsNostr    bool                   `json:"allowsNostr,omitempty"`
}

type LNURLPErrorResponse struct {
	Status string `json:"status" default:"ERROR" validate:"required"`
	Reason string `json:"reason" validate:"required"`
}

func (r LNURLPErrorResponse) MarshalJSON() ([]byte, error) {
	type Alias LNURLPErrorResponse
	if r.Status == "" {
		r.Status = "ERROR"
	}
	return json.Marshal((Alias)(r))
}

type LNURLPCallbackResponse struct {
	Pr     string   `json:"pr" validate:"required"`
	Routes []string `json:"routes" default:"[]"`
	Verify string   `json:"verify,omitempty"`
}

type LNURLPVerifyResponse struct {
	Status   string  `json:"status" validate:"required"`
	Settled  bool    `json:"settled" validate:"required"`
	Preimage *string `json:"preimage"`
	Pr       string  `json:"pr" validate:"required"`
}
