package pocketbase

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase"

	runtimecfg "github.com/joeblew999/infra/core/pkg/runtime/config"
	runtimedep "github.com/joeblew999/infra/core/pkg/runtime/dep"
	composecfg "github.com/joeblew999/infra/core/pkg/runtime/process/composecfg"
)

//go:embed service.json
var manifestFS embed.FS

// Spec models the manifest for PocketBase.
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
	data, err := manifestFS.ReadFile("service.json")
	if err != nil {
		return nil, fmt.Errorf("read pocketbase manifest: %w", err)
	}
	var spec Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("decode pocketbase manifest: %w", err)
	}
	return &spec, nil
}

// EnsureBinaries ensures the PocketBase binary is present.
func (s *Spec) EnsureBinaries() (map[string]string, error) {
	manifest := &runtimedep.Manifest{Binaries: s.Binaries}
	return runtimedep.EnsureManifest(manifest, runtimedep.DefaultInstaller)
}

// ResolveCommand replaces placeholders (e.g. ${dep.*}, ${data}) with actual
// runtime paths.
func (s *Spec) ResolveCommand(paths map[string]string) string {
	cmd := s.Process.Command
	for name, path := range paths {
		placeholder := fmt.Sprintf("${dep.%s}", name)
		cmd = strings.ReplaceAll(cmd, placeholder, path)
	}
	runtime := runtimecfg.Load()
	cmd = strings.ReplaceAll(cmd, "${data}", runtime.Paths.Data)
	return cmd
}

// ResolveEnv performs the same placeholder substitution for environment
// variables as ResolveCommand does for the command string.
func (s *Spec) ResolveEnv(paths map[string]string) map[string]string {
	result := make(map[string]string, len(s.Process.Env))
	runtime := runtimecfg.Load()
	for key, value := range s.Process.Env {
		resolved := value
		// Replace ${dep.*} placeholders with binary paths
		for name, path := range paths {
			placeholder := fmt.Sprintf("${dep.%s}", name)
			resolved = strings.ReplaceAll(resolved, placeholder, path)
		}
		// Replace ${data} placeholder with data directory
		resolved = strings.ReplaceAll(resolved, "${data}", runtime.Paths.Data)
		// Replace ${env.*} placeholders with environment variables
		resolved = replaceEnvPlaceholders(resolved)
		result[key] = resolved
	}
	return result
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

// Run executes an embedded PocketBase instance. Extra args are not currently
// supported because the embedded runner is configured programmatically.
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
		return fmt.Errorf("extra args not supported for embedded PocketBase runner: %v", extraArgs)
	}

	env := spec.ResolveEnv(paths)
	return withEnv(env, func() error {
		return runEmbedded(ctx, spec)
	})
}

func runEmbedded(ctx context.Context, spec *Spec) error {
	cfg := runtimecfg.Load()
	dataDir := filepath.Join(cfg.Paths.Data, "pocketbase")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return fmt.Errorf("prepare data dir: %w", err)
	}

	app := pocketbase.NewWithConfig(pocketbase.Config{
		DefaultDataDir:  dataDir,
		HideStartBanner: true,
	})

	// Bootstrap auth configuration (SMTP, OAuth2, admin user, collection setup)
	if err := BootstrapAuth(app); err != nil {
		return fmt.Errorf("bootstrap auth: %w", err)
	}

	// Register Datastar auth routes
	RegisterDatastarAuth(app, nil)

	port := spec.Ports.Primary.Port
	if port == 0 {
		port = 8090
	}

	app.RootCmd.SetArgs([]string{
		"serve",
		"--dir", dataDir,
		"--http", fmt.Sprintf(":%d", port),
	})

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Execute()
	}()

	if err := waitForTCP(port, 30*time.Second, errCh); err != nil {
		return err
	}
	fmt.Printf("READY: pocketbase http://127.0.0.1:%d\n", port)

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		// trigger graceful shutdown by signalling the process
		_ = signalInterrupt()
		return <-errCh
	}
}

func withEnv(overrides map[string]string, fn func() error) error {
	originals := make(map[string]*string, len(overrides))
	for key, value := range overrides {
		if existing, found := os.LookupEnv(key); found {
			orig := existing
			originals[key] = &orig
		} else {
			originals[key] = nil
		}
		_ = os.Setenv(key, value)
	}

	defer func() {
		for key, val := range originals {
			if val == nil {
				_ = os.Unsetenv(key)
			} else {
				_ = os.Setenv(key, *val)
			}
		}
	}()

	return fn()
}

func signalInterrupt() error {
	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		return err
	}
	return proc.Signal(os.Interrupt)
}

func waitForTCP(port int, timeout time.Duration, errCh <-chan error) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	deadline := time.Now().Add(timeout)
	for {
		select {
		case err := <-errCh:
			return err
		default:
		}

		conn, err := net.DialTimeout("tcp", addr, 250*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("pocketbase port %d not ready: %w", port, err)
		}
		time.Sleep(250 * time.Millisecond)
	}
}
