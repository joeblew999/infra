package controller

import (
	"fmt"

	sharedprocess "github.com/joeblew999/infra/core/pkg/shared/process"
	natssvc "github.com/joeblew999/infra/core/services/nats"
	pbsvc "github.com/joeblew999/infra/core/services/pocketbase"
)

// LoadBuiltIn initialises the registry with built-in services. For now we wire up
// PocketBase and NATS as manifest-driven services; additional services will be
// added once the spec loader is refactored into shared packages.
func LoadBuiltIn() (*Registry, error) {
	r := NewRegistry()
	if err := registerPocketBase(r); err != nil {
		return nil, err
	}
	if err := registerNATS(r); err != nil {
		return nil, err
	}
	return r, nil
}

func registerPocketBase(r *Registry) error {
	spec, err := pbsvc.LoadSpec()
	if err != nil {
		return fmt.Errorf("pocketbase: %w", err)
	}
	paths, err := spec.EnsureBinaries()
	if err != nil {
		return fmt.Errorf("pocketbase ensure binaries: %w", err)
	}
	processSpec := sharedprocess.Spec{
		Command:       spec.ResolveCommand(paths),
		Args:          append([]string{}, spec.Process.Args...),
		Env:           spec.ResolveEnv(paths),
		RestartPolicy: sharedprocess.RestartPolicyOnFailure,
	}
	ports := []Port{{Name: "primary", Port: spec.Ports.Primary.Port, Protocol: spec.Ports.Primary.Protocol}}
	metadata := map[string]string{}
	if spec.ScaleStrategy != "" {
		metadata["scale.strategy"] = spec.ScaleStrategy
	}
	if spec.Scalable {
		metadata["scale.local"] = "true"
	} else {
		metadata["scale.local"] = "false"
	}
	svc := ServiceSpec{
		ID:       "pocketbase",
		Process:  processSpec,
		Ports:    ports,
		Metadata: metadata,
	}
	return r.Register(svc)
}

func registerNATS(r *Registry) error {
	spec, err := natssvc.LoadSpec()
	if err != nil {
		return fmt.Errorf("nats: %w", err)
	}
	paths, err := spec.EnsureBinaries()
	if err != nil {
		return fmt.Errorf("nats ensure binaries: %w", err)
	}
	processSpec := sharedprocess.Spec{
		Command:       spec.ResolveCommand(paths),
		Args:          append([]string{}, spec.Process.Args...),
		Env:           spec.ResolveEnv(paths),
		RestartPolicy: sharedprocess.RestartPolicyOnFailure,
	}
	ports := []Port{
		{Name: "client", Port: spec.Ports.Client.Port, Protocol: spec.Ports.Client.Protocol},
		{Name: "cluster", Port: spec.Ports.Cluster.Port, Protocol: spec.Ports.Cluster.Protocol},
		{Name: "http", Port: spec.Ports.HTTP.Port, Protocol: spec.Ports.HTTP.Protocol},
		{Name: "leaf", Port: spec.Ports.Leaf.Port, Protocol: spec.Ports.Leaf.Protocol},
	}
	metadata := map[string]string{}
	if spec.ScaleStrategy != "" {
		metadata["scale.strategy"] = spec.ScaleStrategy
	}
	if spec.Scalable {
		metadata["scale.local"] = "true"
	} else {
		metadata["scale.local"] = "false"
	}
	svc := ServiceSpec{
		ID:       "nats",
		Process:  processSpec,
		Ports:    ports,
		Metadata: metadata,
	}
	return r.Register(svc)
}
