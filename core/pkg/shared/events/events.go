package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Envelope is the canonical event wrapper shared between runtime modules and
// services. Payload remains a raw JSON blob so listeners can choose their own
// decoding strategy.
type Envelope struct {
	ID        string            `json:"id"`
	Subject   string            `json:"subject"`
	Type      string            `json:"type"`
	Timestamp time.Time         `json:"ts"`
	Metadata  map[string]string `json:"meta,omitempty"`
	Payload   json.RawMessage   `json:"payload"`
}

// Option applies functional modifiers when constructing a new Envelope.
type Option func(*Envelope)

// WithMetadata attaches metadata key/value pairs to the event.
func WithMetadata(meta map[string]string) Option {
	return func(e *Envelope) {
		if len(meta) == 0 {
			return
		}
		if e.Metadata == nil {
			e.Metadata = make(map[string]string, len(meta))
		}
		for k, v := range meta {
			e.Metadata[k] = v
		}
	}
}

// WithID overrides the generated event identifier.
func WithID(id string) Option {
	return func(e *Envelope) {
		if id != "" {
			e.ID = id
		}
	}
}

// Wrap creates an Envelope for the provided subject/type and arbitrary payload.
func Wrap(subject, typ string, payload any, opts ...Option) (Envelope, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, err
	}
	env := Envelope{
		ID:        uuid.NewString(),
		Subject:   subject,
		Type:      typ,
		Timestamp: time.Now().UTC(),
		Payload:   body,
	}
	for _, opt := range opts {
		opt(&env)
	}
	return env, nil
}

// Decode unmarshals the payload into the provided destination.
func (e Envelope) Decode(dest any) error {
	return json.Unmarshal(e.Payload, dest)
}
