package process

import shared "github.com/joeblew999/infra/core/pkg/shared/process"

// Spec mirrors the shared process specification.
type Spec = shared.Spec

// EnsureStep mirrors the shared ensure step.
type EnsureStep = shared.EnsureStep

// RestartPolicy mirrors shared restart policy.
type RestartPolicy = shared.RestartPolicy

// Backoff mirrors shared backoff configuration.
type Backoff = shared.Backoff

const (
	RestartPolicyNever     = shared.RestartPolicyNever
	RestartPolicyOnFailure = shared.RestartPolicyOnFailure
	RestartPolicyAlways    = shared.RestartPolicyAlways
)
