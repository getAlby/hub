package events

import "context"

type EventSubscriber interface {
	ConsumeEvent(ctx context.Context, event *Event, globalProperties map[string]interface{}) error
}

type EventPublisher interface {
	RegisterSubscriber(eventListener EventSubscriber)
	RemoveSubscriber(eventListener EventSubscriber)
	Publish(event *Event)
	SetGlobalProperty(key string, value interface{})
}

type Event struct {
	Event      string      `json:"event"`
	Properties interface{} `json:"properties,omitempty"`
}

type PaymentReceivedEventProperties struct {
	PaymentHash string `json:"payment_hash"`
}

type PaymentSentEventProperties struct {
	PaymentHash string `json:"payment_hash"`
}

type ChannelBackupEvent struct {
	Channels []ChannelBackupInfo `json:"channels"`
}

type ChannelBackupInfo struct {
	ChannelID     string `json:"channel_id"`
	NodeID        string `json:"node_id"`
	PeerID        string `json:"peer_id"`
	ChannelSize   uint64 `json:"channel_size"`
	FundingTxID   string `json:"funding_tx_id"`
	FundingTxVout uint32 `json:"funding_tx_vout"`
}
