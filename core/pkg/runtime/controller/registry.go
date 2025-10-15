package controller

import sharedcontroller "github.com/joeblew999/infra/core/pkg/shared/controller"

// Re-export shared controller types so runtime packages stay within the
// runtime namespace while delegating behaviour to the shared implementation.
type (
	Registry    = sharedcontroller.Registry
	ServiceSpec = sharedcontroller.ServiceSpec
	Port        = sharedcontroller.Port
)

var (
	NewRegistry = sharedcontroller.NewRegistry
)
