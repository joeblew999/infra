package state

import shared "github.com/joeblew999/infra/core/pkg/shared/state"

// Projection re-exports the shared projection type.
type Projection[T any] = shared.Projection[T]

// Reducer re-exports the shared reducer type.
type Reducer[T any] = shared.Reducer[T]

// NewProjection constructs a projection using the shared implementation.
func NewProjection[T any](initial T, reducer Reducer[T]) *Projection[T] {
	return shared.NewProjection(initial, reducer)
}
