package runtime

import "context"

// PreflightFunc allows callers to hook development-time preparation before startup.
type PreflightFunc func(context.Context)

// Options control service startup behaviour.
type Options struct {
	Mode         string
	Preflight    PreflightFunc
	OnlyServices []ServiceID
	SkipServices []ServiceID
}

var (
	activeOptions      Options
	activeServiceSpecs []ServiceSpec
)
