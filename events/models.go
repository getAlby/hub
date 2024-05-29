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
	Amount      uint64 `json:"amount"`
	NodeType    string `json:"node_type"`
}
