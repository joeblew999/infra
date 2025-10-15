package nats

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

	"github.com/Nintron27/pillow"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/jwt/v2"

	runtimecfg "github.com/joeblew999/infra/core/pkg/runtime/config"
	runtimedep "github.com/joeblew999/infra/core/pkg/runtime/dep"
	composecfg "github.com/joeblew999/infra/core/pkg/runtime/process/composecfg"
	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/nats/auth"
)

//go:embed service.json
var manifestFS embed.FS

// Spec models the manifest for the NATS message bus.
type Spec struct {
	Binaries      []runtimedep.BinarySpec `json:"binaries"`
	Process       ProcessSpec             `json:"process"`
	Ports         PortsSpec               `json:"ports"`
	Config        ConfigSpec              `json:"config"`
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

// PortsSpec holds port mappings for NATS.
type PortsSpec struct {
	Client  Port `json:"client"`
	Cluster Port `json:"cluster"`
	HTTP    Port `json:"http"`
	Leaf    Port `json:"leaf"`
}

// Port defines a single binding.
type Port struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

// ConfigSpec defines NATS-specific configuration.
type ConfigSpec struct {
	Backend    string         `json:"backend"`    // "legacy" or "pillow"
	Topology   string         `json:"topology"`   // "mesh" or "hub-spoke"
	AutoScale  bool           `json:"auto_scale"` // Let Pillow handle node scaling
	JetStream  bool           `json:"jetstream"`  // Enable JetStream
	Deployment DeploymentSpec `json:"deployment"` // Environment-specific deployment
}

// DeploymentSpec defines deployment configuration for different environments.
type DeploymentSpec struct {
	Local      LocalDeployment      `json:"local"`
	Production ProductionDeployment `json:"production"`
}

// LocalDeployment defines local development deployment.
type LocalDeployment struct {
	Nodes int    `json:"nodes"` // Number of local nodes
	Mode  string `json:"mode"`  // "embedded" or "standalone"
}

// ProductionDeployment defines production deployment on Fly.io.
type ProductionDeployment struct {
	HubRegion          string   `json:"hub_region"`
	LeafRegions        []string `json:"leaf_regions"`
	MinHubNodes        int      `json:"min_hub_nodes"`
	LeafNodesPerRegion int      `json:"leaf_nodes_per_region"`
}

// LoadSpec decodes the embedded manifest.
func LoadSpec() (*Spec, error) {
	data, err := manifestFS.ReadFile("service.json")
	if err != nil {
		return nil, fmt.Errorf("read nats manifest: %w", err)
	}
	var spec Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("decode nats manifest: %w", err)
	}
	return &spec, nil
}

// EnsureBinaries ensures the NATS server binary is present.
func (s *Spec) EnsureBinaries() (map[string]string, error) {
	manifest := &runtimedep.Manifest{Binaries: s.Binaries}
	return runtimedep.EnsureManifest(manifest, runtimedep.DefaultInstaller)
}

// ResolveCommand replaces placeholders (e.g. ${dep.*}, ${data}) with actual
// runtime paths.
func (s *Spec) ResolveCommand(paths map[string]string) string {
	return replacePlaceholders(s.Process.Command, paths)
}

// ResolveEnv performs the same placeholder substitution for environment
// variables as ResolveCommand does for the command string.
func (s *Spec) ResolveEnv(paths map[string]string) map[string]string {
	result := make(map[string]string, len(s.Process.Env))
	for key, value := range s.Process.Env {
		result[key] = replacePlaceholders(value, paths)
	}
	return result
}

// ResolveArgs applies placeholder replacement to the process arguments.
func (s *Spec) ResolveArgs(paths map[string]string) []string {
	if len(s.Process.Args) == 0 {
		return nil
	}
	args := make([]string, len(s.Process.Args))
	for i, arg := range s.Process.Args {
		args[i] = replacePlaceholders(arg, paths)
	}
	return args
}

// ComposeOverrides returns the optional Process Compose overrides defined in the manifest.
func (s *Spec) ComposeOverrides() map[string]any {
	if s.Process.Compose == nil {
		return nil
	}
	return s.Process.Compose.Map()
}

// Run executes the NATS service using Pillow. Additional arguments are appended
// to the manifest-defined arguments when invoking Pillow.
func Run(ctx context.Context, extraArgs []string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	spec, err := LoadSpec()
	if err != nil {
		return err
	}

	if _, err := spec.EnsureBinaries(); err != nil {
		return err
	}

	if len(extraArgs) > 0 {
		return fmt.Errorf("extra args not supported for embedded Pillow runner: %v", extraArgs)
	}

	return runEmbedded(ctx, spec)
}

// GetNodeCount returns the expected number of NATS nodes based on environment and configuration.
func (s *Spec) GetNodeCount(environment string) int {
	if environment == "development" || environment == "local" {
		return s.Config.Deployment.Local.Nodes
	}

	// Production: hub nodes + leaf nodes
	hubNodes := s.Config.Deployment.Production.MinHubNodes
	leafNodes := len(s.Config.Deployment.Production.LeafRegions) * s.Config.Deployment.Production.LeafNodesPerRegion
	return hubNodes + leafNodes
}

// GetDeploymentStrategy returns a description of how NATS will be deployed.
func (s *Spec) GetDeploymentStrategy(environment string) string {
	if environment == "development" || environment == "local" {
		return fmt.Sprintf("Local embedded: %d node(s) in %s mode",
			s.Config.Deployment.Local.Nodes,
			s.Config.Deployment.Local.Mode)
	}

	prod := s.Config.Deployment.Production
	return fmt.Sprintf("Hub-spoke: %d nodes in %s (hub), %d nodes across %d leaf regions",
		prod.MinHubNodes,
		prod.HubRegion,
		len(prod.LeafRegions)*prod.LeafNodesPerRegion,
		len(prod.LeafRegions))
}

// IsPillowManaged returns true if this service uses Pillow for clustering.
func (s *Spec) IsPillowManaged() bool {
	return s.Config.Backend == "pillow" && s.Config.AutoScale
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

func runEmbedded(ctx context.Context, spec *Spec) error {
	cfg := runtimecfg.Load()
	storeDir := filepath.Join(cfg.Paths.Data, "nats", "jetstream")
	if err := os.MkdirAll(storeDir, 0o755); err != nil {
		return fmt.Errorf("prepare jetstream dir: %w", err)
	}

	// Ensure NSC auth artifacts are available
	authArtifacts, err := auth.Ensure(ctx)
	if err != nil {
		return fmt.Errorf("ensure auth artifacts: %w", err)
	}

	// Parse operator JWT for trusted operators
	operatorClaims, err := jwt.DecodeOperatorClaims(authArtifacts.OperatorJWT)
	if err != nil {
		return fmt.Errorf("decode operator JWT: %w", err)
	}

	// Create memory account resolver
	memResolver := &server.MemAccResolver{}

	// Store system account JWT
	if err := memResolver.Store(authArtifacts.SystemAccountID, authArtifacts.SystemAccountJWT); err != nil {
		return fmt.Errorf("store system account JWT: %w", err)
	}

	// Store application account JWT
	if err := memResolver.Store(authArtifacts.ApplicationAccountID, authArtifacts.ApplicationAccountJWT); err != nil {
		return fmt.Errorf("store application account JWT: %w", err)
	}

	natsOpts := &server.Options{
		Host:                   "0.0.0.0",
		Port:                   spec.Ports.Client.Port,
		HTTPHost:               "0.0.0.0",
		HTTPPort:               spec.Ports.HTTP.Port,
		LeafNode:               server.LeafNodeOpts{Host: "0.0.0.0", Port: spec.Ports.Leaf.Port},
		JetStream:              spec.Config.JetStream,
		StoreDir:               storeDir,
		NoSigs:                 true,
		DisableJetStreamBanner: true,

		// NSC Authentication
		TrustedOperators: []*jwt.OperatorClaims{operatorClaims},
		AccountResolver:  memResolver,
		SystemAccount:    authArtifacts.SystemAccountID,
	}

	if spec.Config.Deployment.Local.Nodes > 1 {
		natsOpts.Cluster = server.ClusterOpts{Host: "0.0.0.0", Port: spec.Ports.Cluster.Port, Name: config.GetNATSClusterName()}
	} else {
		// Single-node local mode: disable clustering/leaf and let HTTP monitor choose a free port
		natsOpts.Cluster = server.ClusterOpts{}
		natsOpts.LeafNode = server.LeafNodeOpts{}
		natsOpts.HTTPPort = 0
	}

	natsOpts.ServerName = fmt.Sprintf("core-nats-%s", cfg.Environment)

	opts := []pillow.Option{
		pillow.WithNATSServerOptions(natsOpts),
		pillow.WithLogging(true),
		pillow.WithTimeout(10 * time.Second),
	}

	server, err := pillow.Run(opts...)
	if err != nil {
		return fmt.Errorf("pillow run: %w", err)
	}

	if err := waitForTCP(spec.Ports.Client.Port, 10*time.Second, nil); err != nil {
		server.NATSServer.Shutdown()
		server.NATSServer.WaitForShutdown()
		return err
	}
	fmt.Printf("READY: nats tcp://127.0.0.1:%d\n", spec.Ports.Client.Port)

	<-ctx.Done()
	server.NATSServer.Shutdown()
	server.NATSServer.WaitForShutdown()
	return nil
}

func waitForTCP(port int, timeout time.Duration, errCh <-chan error) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	deadline := time.Now().Add(timeout)
	for {
		if errCh != nil {
			select {
			case err := <-errCh:
				return err
			default:
			}
		}
		conn, err := net.DialTimeout("tcp", addr, 250*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("port %d not ready: %w", port, err)
		}
		time.Sleep(250 * time.Millisecond)
	}
}
