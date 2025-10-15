package events

import shared "github.com/joeblew999/infra/core/pkg/shared/events"

// Envelope mirrors the shared event envelope type.
type Envelope = shared.Envelope

// Option mirrors shared event options type.
type Option = shared.Option

var (
	Wrap         = shared.Wrap
	WithMetadata = shared.WithMetadata
	WithID       = shared.WithID
)
