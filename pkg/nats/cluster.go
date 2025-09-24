package nats

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/dep"
	"github.com/joeblew999/infra/pkg/fly"
	"github.com/joeblew999/infra/pkg/goreman"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/joeblew999/infra/pkg/nats/auth"
	"github.com/joeblew999/infra/pkg/service"
)

// ClusterNode represents a single NATS node in the cluster
type ClusterNode struct {
	Name        string `json:"name"`
	Region      string `json:"region"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	ClusterPort int    `json:"cluster_port"`
	HTTPPort    int    `json:"http_port"`
	LeafPort    int    `json:"leaf_port"`
	IsLocal     bool   `json:"is_local"`
	Status      string `json:"status"`
}

// ClusterConfig represents the configuration for a NATS cluster
type ClusterConfig struct {
	Nodes           []ClusterNode `json:"nodes"`
	ClusterName     string        `json:"cluster_name"`
	Environment     string        `json:"environment"`
	EnableWebGUI    bool          `json:"enable_web_gui"`
	EnableJetStream bool          `json:"enable_jetstream"`
}

// Using config functions instead of hardcoded values

// GetLocalClusterConfig returns configuration for local Docker-based NATS cluster
func GetLocalClusterConfig() ClusterConfig {
	nodeCount := config.GetNATSClusterNodeCount()
	nodes := make([]ClusterNode, 0, nodeCount)

	for i := 0; i < nodeCount; i++ {
		client, cluster, http, leaf := config.GetNATSClusterPortsForNode(i)
		nodes = append(nodes, ClusterNode{
			Name:        fmt.Sprintf("nats-%d", i+1),
			Region:      "local",
			Host:        "127.0.0.1",
			Port:        client,
			ClusterPort: cluster,
			HTTPPort:    http,
			LeafPort:    leaf,
			IsLocal:     true,
			Status:      "unknown",
		})
	}

	return ClusterConfig{
		Nodes:           nodes,
		ClusterName:     config.GetNATSClusterName(),
		Environment:     config.EnvDevelopment,
		EnableWebGUI:    true,
		EnableJetStream: true,
	}
}

// GetFlyClusterConfig returns configuration for Fly.io NATS cluster
func GetFlyClusterConfig() ClusterConfig {
	regions := config.GetFlyRegions()
	nodes := make([]ClusterNode, 0, len(regions))

	for i, region := range regions {
		client, cluster, http, leaf := config.GetNATSClusterPortsForNode(i)
		nodes = append(nodes, ClusterNode{
			Name:        fmt.Sprintf("nats-%s", region),
			Region:      region,
			Host:        fmt.Sprintf("nats-%s", region),
			Port:        client,
			ClusterPort: cluster,
			HTTPPort:    http,
			LeafPort:    leaf,
			IsLocal:     false,
			Status:      "unknown",
		})
	}

	return ClusterConfig{
		Nodes:           nodes,
		ClusterName:     config.GetNATSClusterName(),
		Environment:     config.EnvProduction,
		EnableWebGUI:    true,
		EnableJetStream: true,
	}
}

// GetClusterLeafRemotes returns the leaf node remote URLs for the target environment.
func GetClusterLeafRemotes(isLocal bool) []string {
	var clusterConfig ClusterConfig
	if isLocal {
		clusterConfig = GetLocalClusterConfig()
	} else {
		clusterConfig = GetFlyClusterConfig()
	}

	remotes := make([]string, 0, len(clusterConfig.Nodes))
	for _, node := range clusterConfig.Nodes {
		host := node.Host
		if host == "" {
			host = node.Name
		}
		remotes = append(remotes, fmt.Sprintf("nats://%s:%d", host, node.LeafPort))
	}

	return remotes
}

// StartLocalCluster starts (or ensures) a local NATS cluster under goreman supervision
func StartLocalCluster(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	authArtifacts, err := auth.Ensure(ctx)
	if err != nil {
		return err
	}

	return EnsureCluster(ctx, GetLocalClusterConfig(), authArtifacts)
}

// EnsureCluster ensures the provided cluster configuration is running under goreman supervision.
func EnsureCluster(ctx context.Context, clusterConfig ClusterConfig, authArtifacts *auth.Artifacts) error {
	log.Info("Ensuring NATS cluster", "name", clusterConfig.ClusterName, "nodes", len(clusterConfig.Nodes), "environment", clusterConfig.Environment)

	if err := dep.InstallBinary(config.BinaryNatsServer, false); err != nil {
		return fmt.Errorf("failed to ensure nats binary: %w", err)
	}

	clusterDataPath := config.GetNATSClusterDataPath()
	if err := os.MkdirAll(clusterDataPath, 0755); err != nil {
		return fmt.Errorf("failed to create cluster data directory: %w", err)
	}

	processNames := make([]string, 0, len(clusterConfig.Nodes))
	isLocal := clusterConfig.Environment != config.EnvProduction

	for _, node := range clusterConfig.Nodes {
		if err := ensureClusterNode(clusterConfig, node, clusterDataPath, isLocal, authArtifacts); err != nil {
			return err
		}
		processNames = append(processNames, clusterProcessName(node))
	}

	goreman.RegisterGroup("nats-cluster", processNames)

	log.Info("NATS cluster ensured", "nodes", len(processNames))
	return nil
}

func ensureClusterNode(clusterConfig ClusterConfig, node ClusterNode, clusterDataPath string, isLocal bool, authArtifacts *auth.Artifacts) error {
	nodeDataPath := filepath.Join(clusterDataPath, node.Name)
	if err := os.MkdirAll(nodeDataPath, 0755); err != nil {
		return fmt.Errorf("failed to create node data directory: %w", err)
	}

	configPath := filepath.Join(nodeDataPath, "nats.conf")
	processName := clusterProcessName(node)
	processCfg := service.NewConfig(config.Get(config.BinaryNatsServer), []string{"--config", configPath})

	if checkNodeHTTPHealth(node, isLocal) {
		log.Info("NATS node already running", "node", node.Name, "host", node.Host)
		// Register the process configuration for completeness so goreman knows about it in this invocation.
		goreman.Register(processName, processCfg)
		return nil
	}
	if err := writeNodeConfig(clusterConfig, node, configPath, nodeDataPath, authArtifacts); err != nil {
		return err
	}

	if err := service.Start(processName, processCfg); err != nil {
		return fmt.Errorf("failed to start cluster node %s: %w", node.Name, err)
	}

	// Allow node to settle before starting the next one.
	time.Sleep(500 * time.Millisecond)

	return nil
}

func writeNodeConfig(clusterConfig ClusterConfig, node ClusterNode, configPath, dataDir string, authArtifacts *auth.Artifacts) error {
	routes := make([]string, 0, len(clusterConfig.Nodes)-1)
	for _, other := range clusterConfig.Nodes {
		if other.Name == node.Name {
			continue
		}
		host := other.Host
		if host == "" {
			host = other.Name
		}
		routes = append(routes, fmt.Sprintf("\"nats://%s:%d\"", host, other.ClusterPort))
	}

	routesStr := strings.Join(routes, ",")
	jetstreamDir := filepath.Join(dataDir, "jetstream")
	if err := os.MkdirAll(jetstreamDir, 0755); err != nil {
		return fmt.Errorf("failed to create JetStream directory: %w", err)
	}

	natsConfig := fmt.Sprintf(`
# NATS Server Configuration for %s
server_name: %s
host: %s
port: %d
http: %d

operator: %q
system_account: %s

resolver: MEMORY
resolver_preload: {
    %s: %q
    %s: %q
}

cluster {
    name: %s
    port: %d
    routes: [%s]
    no_advertise: false
}

	jetstream: {
	    store_dir: "%s"
	    max_memory_store: 256MB
	    max_file_store: 2GB
	}

leaf {
	listen: %d
}

debug: false
trace: false
logtime: true
`,
		node.Name,
		node.Name,
		node.Host,
		node.Port,
		node.HTTPPort,
		authArtifacts.OperatorJWT,
		authArtifacts.SystemAccountID,
		authArtifacts.SystemAccountID,
		authArtifacts.SystemAccountJWT,
		authArtifacts.ApplicationAccountID,
		authArtifacts.ApplicationAccountJWT,
		clusterConfig.ClusterName,
		node.ClusterPort,
		routesStr,
		jetstreamDir,
		node.LeafPort,
	)

	if err := os.WriteFile(configPath, []byte(natsConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config for node %s: %w", node.Name, err)
	}

	return nil
}

func clusterProcessName(node ClusterNode) string {
	return node.Name
}

// StopLocalCluster stops the goreman-supervised local NATS cluster processes
func StopLocalCluster() error {
	log.Info("Stopping local NATS cluster")
	if err := goreman.StopGroup("nats-cluster"); err != nil {
		log.Debug("goreman group stop", "error", err)
	}
	clusterConfig := GetLocalClusterConfig()
	var stopErr error
	for _, node := range clusterConfig.Nodes {
		configPath := filepath.Join(config.GetNATSClusterDataPath(), node.Name, "nats.conf")
		if err := stopNodeProcessByConfig(configPath); err != nil {
			log.Warn("Failed to stop NATS node", "node", node.Name, "error", err)
			if stopErr == nil {
				stopErr = err
			}
		}
	}
	return stopErr
}

// EnsureSingleFlyNode starts a single NATS node for the specified Fly app under goreman supervision
func EnsureSingleFlyNode(ctx context.Context, flyAppName string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	// Get the Fly cluster config and find our node
	clusterConfig := GetFlyClusterConfig()
	var targetNode *ClusterNode
	for i, node := range clusterConfig.Nodes {
		if node.Name == flyAppName {
			targetNode = &clusterConfig.Nodes[i]
			break
		}
	}

	if targetNode == nil {
		return fmt.Errorf("no NATS node configuration found for Fly app: %s", flyAppName)
	}

	log.Info("Ensuring single NATS node for Fly", "app", flyAppName, "region", targetNode.Region)

	if err := dep.InstallBinary(config.BinaryNatsServer, false); err != nil {
		return fmt.Errorf("failed to ensure nats binary: %w", err)
	}

	authArtifacts, err := auth.Ensure(ctx)
	if err != nil {
		return fmt.Errorf("ensure auth materials: %w", err)
	}

	clusterDataPath := config.GetNATSClusterDataPath()
	if err := os.MkdirAll(clusterDataPath, 0755); err != nil {
		return fmt.Errorf("failed to create cluster data directory: %w", err)
	}

	// Ensure this single node
	if err := ensureClusterNode(clusterConfig, *targetNode, clusterDataPath, false, authArtifacts); err != nil {
		return fmt.Errorf("failed to ensure single node %s: %w", flyAppName, err)
	}

	// Register with goreman
	processName := clusterProcessName(*targetNode)
	goreman.RegisterGroup("nats-cluster", []string{processName})

	log.Info("Single NATS node ensured for Fly", "app", flyAppName, "process", processName)
	return nil
}

// checkNodeHTTPHealth performs HTTP health check on a NATS node's monitoring endpoint
func checkNodeHTTPHealth(node ClusterNode, isLocal bool) bool {
	var url string
	if isLocal {
		// For local nodes, use localhost with the specific HTTP port
		url = fmt.Sprintf("http://127.0.0.1:%d/", node.HTTPPort)
	} else {
		// For Fly nodes, use the fly.dev hostname with standard port 8222
		url = fmt.Sprintf("http://%s.fly.dev:8222/", node.Name)
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		log.Debug("HTTP health check failed", "node", node.Name, "url", url, "error", err)
		return false
	}
	defer resp.Body.Close()

	// Check if we get a successful response (2xx status code)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Debug("HTTP health check passed", "node", node.Name, "url", url, "status", resp.StatusCode)
		return true
	}

	log.Debug("HTTP health check failed with bad status", "node", node.Name, "url", url, "status", resp.StatusCode)
	return false
}

// GetClusterStatus returns the status of cluster nodes with comprehensive health checks
func GetClusterStatus(isLocal bool) (ClusterConfig, error) {
	var clusterConfig ClusterConfig
	if isLocal {
		clusterConfig = GetLocalClusterConfig()
	} else {
		clusterConfig = GetFlyClusterConfig()
	}

	// Check status of each node
	for i, node := range clusterConfig.Nodes {
		if isLocal {
			if httpStatus := checkNodeHTTPHealth(node, isLocal); httpStatus {
				clusterConfig.Nodes[i].Status = "running"
			} else if goreman.IsRunning(clusterProcessName(node)) {
				clusterConfig.Nodes[i].Status = "unhealthy"
			} else {
				clusterConfig.Nodes[i].Status = "stopped"
			}
		} else {
			status, fetchErr := getFlyClusterNodeStatus(node.Name)
			if fetchErr != nil {
				log.Warn("Failed to fetch Fly node status, trying HTTP health check", "node", node.Name, "error", fetchErr)
				// Fallback to HTTP health check if Fly status fails
				if httpStatus := checkNodeHTTPHealth(node, isLocal); httpStatus {
					clusterConfig.Nodes[i].Status = "running"
				} else {
					clusterConfig.Nodes[i].Status = "error"
				}
			} else {
				// Additional HTTP health check for Fly nodes
				if status == "running" || status == "unknown" {
					if httpStatus := checkNodeHTTPHealth(node, isLocal); httpStatus {
						clusterConfig.Nodes[i].Status = "running"
					} else {
						clusterConfig.Nodes[i].Status = "unhealthy"
					}
				} else {
					clusterConfig.Nodes[i].Status = status
				}
			}
		}
	}

	return clusterConfig, nil
}

func getFlyClusterNodeStatus(appName string) (string, error) {
	return fly.GetAppStatus(appName)
}

func stopNodeProcessByConfig(configPath string) error {
	cmd := exec.Command("pgrep", "-f", configPath)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil // nothing to stop
		}
		return fmt.Errorf("pgrep failed for %s: %w", configPath, err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		pidStr := strings.TrimSpace(scanner.Text())
		if pidStr == "" {
			continue
		}
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			log.Warn("Invalid PID from pgrep", "pid", pidStr, "error", err)
			continue
		}
		proc, err := os.FindProcess(pid)
		if err != nil {
			log.Warn("Failed to find process", "pid", pid, "error", err)
			continue
		}
		if err := proc.Signal(os.Interrupt); err != nil {
			if killErr := proc.Kill(); killErr != nil {
				log.Warn("Failed to terminate process", "pid", pid, "error", killErr)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan pgrep output: %w", err)
	}

	return nil
}

// DeployFlyCluster deploys NATS cluster to Fly.io across multiple regions
func DeployFlyCluster(ctx context.Context) error {
	log.Info("Deploying NATS cluster to Fly.io...")

	clusterConfig := GetFlyClusterConfig()

	// Deploy each node to its region
	for _, node := range clusterConfig.Nodes {
		if err := deployFlyClusterNode(ctx, node, clusterConfig); err != nil {
			log.Error("Failed to deploy cluster node", "node", node.Name, "region", node.Region, "error", err)
			return fmt.Errorf("failed to deploy node %s to region %s: %w", node.Name, node.Region, err)
		}
	}

	log.Info("NATS cluster deployed to Fly.io successfully", "nodes", len(clusterConfig.Nodes))
	return nil
}

// deployFlyClusterNode deploys a single NATS node to Fly.io using the unified Fly package
func deployFlyClusterNode(ctx context.Context, node ClusterNode, clusterConfig ClusterConfig) error {
	return fly.DeployNATSCluster(node.Name, node.Region)
}

// UpgradeCluster performs rolling upgrade of NATS cluster using lame duck mode
func UpgradeCluster(ctx context.Context, isLocal bool) error {
	if isLocal {
		return fmt.Errorf("rolling upgrade not supported for local goreman-managed clusters yet")
	}

	log.Info("Starting rolling cluster upgrade with lame duck mode...")

	var clusterConfig ClusterConfig
	if isLocal {
		clusterConfig = GetLocalClusterConfig()
	} else {
		clusterConfig = GetFlyClusterConfig()
	}

	// Upgrade nodes one by one
	for i, node := range clusterConfig.Nodes {
		log.Info("Upgrading cluster node", "node", node.Name, "step", fmt.Sprintf("%d/%d", i+1, len(clusterConfig.Nodes)))

		if err := upgradeClusterNode(ctx, node, isLocal); err != nil {
			log.Error("Failed to upgrade cluster node", "node", node.Name, "error", err)
			return fmt.Errorf("failed to upgrade node %s: %w", node.Name, err)
		}

		// Wait for node to rejoin cluster before proceeding
		if err := waitForNodeReady(ctx, node, isLocal); err != nil {
			log.Error("Node failed to rejoin cluster", "node", node.Name, "error", err)
			return fmt.Errorf("node %s failed to rejoin cluster: %w", node.Name, err)
		}

		log.Info("Node upgraded successfully", "node", node.Name)
	}

	log.Info("Rolling cluster upgrade completed successfully")
	return nil
}

// upgradeClusterNode upgrades a single node using lame duck mode
func upgradeClusterNode(ctx context.Context, node ClusterNode, isLocal bool) error {
	if isLocal {
		return fmt.Errorf("local cluster upgrade is not implemented for goreman-managed clusters")
	}

	// For Fly.io: use Fly's deployment command
	appName := node.Name
	cmd := exec.Command(config.GetFlyctlBinPath(), "deploy", "--app", appName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// waitForNodeReady waits for a node to rejoin the cluster and be ready
func waitForNodeReady(ctx context.Context, node ClusterNode, isLocal bool) error {
	if isLocal {
		return fmt.Errorf("waitForNodeReady not supported for local goreman-managed cluster")
	}

	timeout := 30 * time.Second
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("timeout waiting for node to be ready")
		case <-ticker.C:
			return nil // Placeholder until Fly.io health integration is added
		}
	}
}
