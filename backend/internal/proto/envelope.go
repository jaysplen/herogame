package proto

import (
	"encoding/json"
	"time"
)

// Envelope is the wire format for every WebSocket message (architecture.md §7.1).
type Envelope struct {
	Type       string          `json:"type"`
	Payload    json.RawMessage `json:"payload"`
	Seq        int64           `json:"seq"`
	ServerTime int64           `json:"serverTime"`
}

// NewEnvelope builds an outbound envelope with the current server time.
func NewEnvelope(msgType string, payload any, seq int64) (Envelope, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, err
	}
	return Envelope{
		Type:       msgType,
		Payload:    raw,
		Seq:        seq,
		ServerTime: time.Now().UnixMilli(),
	}, nil
}

// MustEnvelope panics on marshal error (tests only).
func MustEnvelope(msgType string, payload any, seq int64) Envelope {
	env, err := NewEnvelope(msgType, payload, seq)
	if err != nil {
		panic(err)
	}
	return env
}

// ParseInbound decodes raw JSON into an Envelope.
func ParseInbound(data []byte) (Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return Envelope{}, err
	}
	return env, nil
}

// DecodePayload unmarshals the payload into dest.
func (e Envelope) DecodePayload(dest any) error {
	return json.Unmarshal(e.Payload, dest)
}
