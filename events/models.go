package events

import (
	"context"
)

type EventSubscriber interface {
	ConsumeEvent(ctx context.Context, event *Event, globalProperties map[string]interface{})
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

type StaticChannelsBackupEvent struct {
	NodeID   string                        `json:"node_id"`
	Channels []ChannelBackup               `json:"channels"`
	Monitors []EncodedChannelMonitorBackup `json:"monitors"`
}

type EncodedChannelMonitorBackup struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ChannelBackup struct {
	ChannelID         string `json:"channel_id"`
	PeerID            string `json:"peer_id"`
	PeerSocketAddress string `json:"peer_socket_address"`
	ChannelSize       uint64 `json:"channel_size"`
	FundingTxID       string `json:"funding_tx_id"`
	FundingTxVout     uint32 `json:"funding_tx_vout"`
}
