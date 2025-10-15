package spec

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// DesiredState represents the declarative scaling/routing/storage information
// the controller reconciles.
type DesiredState struct {
	Services []Service `yaml:"services" json:"services"`
}

// Service captures desired state for a single logical service.
type Service struct {
	ID          string      `yaml:"id" json:"id"`
	DisplayName string      `yaml:"displayName,omitempty" json:"displayName,omitempty"`
	Description string      `yaml:"description,omitempty" json:"description,omitempty"`
	Scale       ScaleSpec   `yaml:"scale" json:"scale"`
	Storage     StorageSpec `yaml:"storage,omitempty" json:"storage,omitempty"`
	Routing     RoutingSpec `yaml:"routing,omitempty" json:"routing,omitempty"`
}

// ScaleSpec defines desired replica counts and scaling behaviour.
type ScaleSpec struct {
	Strategy  string            `yaml:"strategy" json:"strategy"`                       // local | infra
	Autoscale string            `yaml:"autoscale,omitempty" json:"autoscale,omitempty"` // manual | metrics | disabled
	Cooldown  string            `yaml:"cooldown,omitempty" json:"cooldown,omitempty"`
	BurstTTL  string            `yaml:"burst_ttl,omitempty" json:"burst_ttl,omitempty"`
	Regions   []RegionScaleSpec `yaml:"regions" json:"regions"`
}

// RegionScaleSpec defines desired replica counts within a region.
type RegionScaleSpec struct {
	Name    string `yaml:"name" json:"name"`
	Min     int    `yaml:"min" json:"min"`
	Desired int    `yaml:"desired" json:"desired"`
	Max     int    `yaml:"max" json:"max"`
}

// StorageSpec captures durable state requirements (e.g. Cloudflare R2).
type StorageSpec struct {
	Provider       string            `yaml:"provider" json:"provider"` // cloudflare-r2, s3, gcs
	CredentialsRef string            `yaml:"credentialsRef,omitempty" json:"credentialsRef,omitempty"`
	Buckets        StorageBuckets    `yaml:"buckets" json:"buckets"`
	Options        map[string]string `yaml:"options,omitempty" json:"options,omitempty"`
}

// StorageBuckets enumerates required buckets for a service.
type StorageBuckets struct {
	Litestream string `yaml:"litestream,omitempty" json:"litestream,omitempty"`
	Assets     string `yaml:"assets,omitempty" json:"assets,omitempty"`
}

// RoutingSpec describes edge routing configuration (e.g. Cloudflare).
type RoutingSpec struct {
	Provider      string          `yaml:"provider" json:"provider"` // cloudflare, fly, aws-alb
	Zone          string          `yaml:"zone,omitempty" json:"zone,omitempty"`
	DNSRecords    []DNSRecordSpec `yaml:"dns_records,omitempty" json:"dns_records,omitempty"`
	HealthChecks  []HealthCheck   `yaml:"health_checks,omitempty" json:"health_checks,omitempty"`
	LoadBalancing LoadBalancing   `yaml:"load_balancing,omitempty" json:"load_balancing,omitempty"`
}

// DNSRecordSpec models a DNS record managed by the controller.
type DNSRecordSpec struct {
    Name    string `yaml:"name" json:"name"`
    Type    string `yaml:"type" json:"type"`
    TTL     int    `yaml:"ttl,omitempty" json:"ttl,omitempty"`
    Content string `yaml:"content,omitempty" json:"content,omitempty"`
    Comment string `yaml:"comment,omitempty" json:"comment,omitempty"`
}

// HealthCheck defines routing health checks.
type HealthCheck struct {
	Name     string `yaml:"name" json:"name"`
	Path     string `yaml:"path" json:"path"`
	Interval string `yaml:"interval,omitempty" json:"interval,omitempty"`
	Timeout  string `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// LoadBalancing captures load-balancing strategy configuration.
type LoadBalancing struct {
	Strategy string            `yaml:"strategy,omitempty" json:"strategy,omitempty"` // latency, round_robin, geo
	Options  map[string]string `yaml:"options,omitempty" json:"options,omitempty"`
}

// Load parses desired state from an io.Reader.
func Load(r io.Reader) (DesiredState, error) {
	var state DesiredState
	decoder := yaml.NewDecoder(r)
	decoder.KnownFields(true)
	if err := decoder.Decode(&state); err != nil && err != io.EOF {
		return DesiredState{}, err
	}
	return state, state.Validate()
}

// LoadFile parses desired state from a file path.
func LoadFile(path string) (DesiredState, error) {
	f, err := os.Open(path)
	if err != nil {
		return DesiredState{}, err
	}
	defer f.Close()
	return Load(f)
}

// LoadServiceFile attempts to parse a single service definition from a YAML file.
// The file may contain a service at the root or be wrapped in a desired state.
func LoadServiceFile(path string) (Service, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Service{}, err
	}

	var svc Service
	if err := yaml.Unmarshal(data, &svc); err == nil && svc.ID != "" {
		state := DesiredState{Services: []Service{svc}}
		if err := state.Validate(); err != nil {
			return Service{}, err
		}
		return svc, nil
	}

	var wrapper struct {
		Service  Service   `yaml:"service"`
		Services []Service `yaml:"services"`
	}
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return Service{}, err
	}
	if wrapper.Service.ID != "" {
		state := DesiredState{Services: []Service{wrapper.Service}}
		if err := state.Validate(); err != nil {
			return Service{}, err
		}
		return wrapper.Service, nil
	}
	if len(wrapper.Services) == 1 {
		state := DesiredState{Services: wrapper.Services}
		if err := state.Validate(); err != nil {
			return Service{}, err
		}
		return wrapper.Services[0], nil
	}
	if len(wrapper.Services) > 1 {
		return Service{}, fmt.Errorf("service file %s contains multiple services", path)
	}
	return Service{}, fmt.Errorf("could not parse service definition in %s", path)
}

// Validate ensures the desired state is well-formed.
func (d DesiredState) Validate() error {
	ids := make(map[string]struct{}, len(d.Services))
	for _, svc := range d.Services {
		if svc.ID == "" {
			return fmt.Errorf("service id is required")
		}
		if _, exists := ids[svc.ID]; exists {
			return fmt.Errorf("service id %s defined multiple times", svc.ID)
		}
		ids[svc.ID] = struct{}{}
		if len(svc.Scale.Regions) == 0 {
			return fmt.Errorf("service %s must define at least one region", svc.ID)
		}
		for _, region := range svc.Scale.Regions {
			if region.Name == "" {
				return fmt.Errorf("service %s has a region with empty name", svc.ID)
			}
			if region.Min < 0 || region.Desired < 0 || region.Max < 0 {
				return fmt.Errorf("service %s region %s has negative replica counts", svc.ID, region.Name)
			}
			if region.Min > region.Desired {
				return fmt.Errorf("service %s region %s has min > desired", svc.ID, region.Name)
			}
			if region.Desired > region.Max {
				return fmt.Errorf("service %s region %s has desired > max", svc.ID, region.Name)
			}
		}
	}
	return nil
}
