package pocketbaseha

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/litesql/pocketbase-ha"
	"github.com/pocketbase/pocketbase/core"

	runtimecfg "github.com/joeblew999/infra/core/pkg/runtime/config"
	runtimedep "github.com/joeblew999/infra/core/pkg/runtime/dep"
	composecfg "github.com/joeblew999/infra/core/pkg/runtime/process/composecfg"
	"github.com/joeblew999/infra/pkg/config"

	// Import the regular pocketbase service to reuse bootstrap and auth handlers
	pbservice "github.com/joeblew999/infra/core/services/pocketbase"
)

//go:embed service.json
//go:embed *.html
var embedFS embed.FS

// Spec models the manifest for PocketBase HA.
type Spec struct {
	Binaries      []runtimedep.BinarySpec `json:"binaries"`
	Process       ProcessSpec             `json:"process"`
	Ports         PortsSpec               `json:"ports"`
	Scalable      bool                    `json:"scalable,omitempty"`
	ScaleStrategy string                  `json:"scale_strategy,omitempty"`
}

// ProcessSpec describes the command invocation.
type ProcessSpec struct {
	Command string             `json:"command"`
	Args    []string           `json:"args,omitempty"`
	Env     map[string]string  `json:"env,omitempty"`
	Compose *composecfg.Config `json:"compose,omitempty"`
}

// PortsSpec holds port mappings.
type PortsSpec struct {
	Primary Port `json:"primary"`
}

// Port defines a single binding.
type Port struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

// LoadSpec decodes the embedded manifest.
func LoadSpec() (*Spec, error) {
	data, err := embedFS.ReadFile("service.json")
	if err != nil {
		return nil, fmt.Errorf("read pocketbase-ha manifest: %w", err)
	}
	var spec Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("decode pocketbase-ha manifest: %w", err)
	}
	return &spec, nil
}

// EnsureBinaries ensures any required binaries are present.
func (s *Spec) EnsureBinaries() (map[string]string, error) {
	if len(s.Binaries) == 0 {
		return make(map[string]string), nil
	}
	manifest := &runtimedep.Manifest{Binaries: s.Binaries}
	return runtimedep.EnsureManifest(manifest, runtimedep.DefaultInstaller)
}

// ResolveCommand replaces placeholders in the command string.
func (s *Spec) ResolveCommand(paths map[string]string) string {
	return replacePlaceholders(s.Process.Command, paths)
}

// ResolveEnv performs placeholder substitution for environment variables.
func (s *Spec) ResolveEnv(paths map[string]string) map[string]string {
	result := make(map[string]string, len(s.Process.Env))
	runtime := runtimecfg.Load()
	for key, value := range s.Process.Env {
		resolved := replacePlaceholders(value, paths)
		resolved = strings.ReplaceAll(resolved, "${data}", runtime.Paths.Data)
		result[key] = resolved
	}
	return result
}

// ResolveArgs applies placeholder replacement to process arguments.
func (s *Spec) ResolveArgs(paths map[string]string) []string {
	if len(s.Process.Args) == 0 {
		return nil
	}
	args := make([]string, len(s.Process.Args))
	runtime := runtimecfg.Load()
	for i, arg := range s.Process.Args {
		resolved := arg
		for name, path := range paths {
			placeholder := fmt.Sprintf("${dep.%s}", name)
			resolved = strings.ReplaceAll(resolved, placeholder, path)
		}
		resolved = strings.ReplaceAll(resolved, "${data}", runtime.Paths.Data)
		args[i] = resolved
	}
	return args
}

// ComposeOverrides returns the optional Process Compose overrides.
func (s *Spec) ComposeOverrides() map[string]any {
	if s.Process.Compose == nil {
		return nil
	}
	return s.Process.Compose.Map()
}

// Run executes an embedded PocketBase-HA instance with Pillow NATS integration.
func Run(ctx context.Context, extraArgs []string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	spec, err := LoadSpec()
	if err != nil {
		return err
	}

	paths, err := spec.EnsureBinaries()
	if err != nil {
		return err
	}

	if len(extraArgs) > 0 {
		return fmt.Errorf("extra args not supported for embedded PocketBase-HA runner: %v", extraArgs)
	}

	env := spec.ResolveEnv(paths)
	return withEnv(env, func() error {
		return runEmbedded(ctx, spec)
	})
}

func runEmbedded(ctx context.Context, spec *Spec) error {
	cfg := runtimecfg.Load()
	dataDir := filepath.Join(cfg.Paths.Data, "pocketbase-ha")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return fmt.Errorf("prepare data dir: %w", err)
	}

	// Configure pocketbase-ha with connection to our Pillow-managed NATS
	app := pocketbaseha.NewWithConfig(pocketbaseha.Config{
		DefaultDataDir:  dataDir,
		HideStartBanner: true,

		// Connect to our Pillow-managed NATS cluster
		// Use NATS URL from environment or default to localhost
		ReplicationURL: getReplicationURL(),
		NodeName:       getNodeName(),
		StreamName:     getStreamName(),
	})

	// Bootstrap auth configuration (reuse from regular pocketbase service)
	if err := pbservice.BootstrapAuth(app.App); err != nil {
		return fmt.Errorf("bootstrap auth: %w", err)
	}

	// Register Datastar auth routes (reuse from regular pocketbase service)
	pbservice.RegisterDatastarAuth(app.App, embedFS)

	port := spec.Ports.Primary.Port
	if port == 0 {
		port = 8090
	}

	// Run the server
	go func() {
		if err := app.Start(fmt.Sprintf("0.0.0.0:%d", port)); err != nil {
			fmt.Fprintf(os.Stderr, "pocketbase-ha: %v\n", err)
		}
	}()

	// Wait for readiness
	fmt.Printf("READY: pocketbase-ha http://127.0.0.1:%d\n", port)

	<-ctx.Done()
	return nil
}

func replacePlaceholders(value string, paths map[string]string) string {
	if value == "" {
		return value
	}
	resolved := value
	runtime := runtimecfg.Load()
	// Replace ${dep.*} placeholders with binary paths
	for name, path := range paths {
		placeholder := fmt.Sprintf("${dep.%s}", name)
		resolved = strings.ReplaceAll(resolved, placeholder, path)
	}
	// Replace runtime path placeholders
	resolved = strings.ReplaceAll(resolved, "${data}", runtime.Paths.Data)
	resolved = strings.ReplaceAll(resolved, "${bin}", runtime.Paths.Bin)
	resolved = strings.ReplaceAll(resolved, "${dep}", runtime.Paths.Dep)
	resolved = strings.ReplaceAll(resolved, "${logs}", runtime.Paths.Logs)
	// Replace ${env.*} placeholders with environment variables
	resolved = replaceEnvPlaceholders(resolved)
	return resolved
}

// replaceEnvPlaceholders substitutes ${env.VARIABLE_NAME} with os.Getenv("VARIABLE_NAME").
func replaceEnvPlaceholders(value string) string {
	// Pattern: ${env.VARIABLE_NAME}
	for {
		start := strings.Index(value, "${env.")
		if start == -1 {
			break
		}
		end := strings.Index(value[start:], "}")
		if end == -1 {
			break
		}
		end += start
		placeholder := value[start : end+1]
		envVar := value[start+len("${env.") : end]
		replacement := os.Getenv(envVar)
		value = strings.ReplaceAll(value, placeholder, replacement)
	}
	return value
}

func withEnv(env map[string]string, fn func() error) error {
	var restore []func()
	for k, v := range env {
		old, hadOld := os.LookupEnv(k)
		if err := os.Setenv(k, v); err != nil {
			for _, r := range restore {
				r()
			}
			return err
		}
		restore = append(restore, func() {
			if hadOld {
				_ = os.Setenv(k, old)
			} else {
				_ = os.Unsetenv(k)
			}
		})
	}
	defer func() {
		for _, r := range restore {
			r()
		}
	}()
	return fn()
}

// getReplicationURL returns the NATS connection URL for replication.
// Uses PB_REPLICATION_URL env var or defaults to Pillow NATS cluster.
func getReplicationURL() string {
	if url := os.Getenv("PB_REPLICATION_URL"); url != "" {
		return url
	}
	// Default to our Pillow-managed NATS on localhost
	return config.GetNATSURL()
}

// getNodeName returns the unique node name for this PocketBase-HA instance.
// Uses PB_NAME env var or generates from hostname.
func getNodeName() string {
	if name := os.Getenv("PB_NAME"); name != "" {
		return name
	}
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "pocketbase-ha-node"
	}
	return hostname
}

// getStreamName returns the NATS stream name for PocketBase replication.
// Uses PB_REPLICATION_STREAM env var or defaults to "pb".
func getStreamName() string {
	if stream := os.Getenv("PB_REPLICATION_STREAM"); stream != "" {
		return stream
	}
	return "pb"
}
