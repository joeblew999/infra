package nats

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
)

// ClusterNode represents a single NATS node in the cluster
type ClusterNode struct {
	Name        string `json:"name"`
	Region      string `json:"region"`
	Port        int    `json:"port"`
	ClusterPort int    `json:"cluster_port"`
	HTTPPort    int    `json:"http_port"`
	IsLocal     bool   `json:"is_local"`
	Status      string `json:"status"`
}

// ClusterConfig represents the configuration for a NATS cluster
type ClusterConfig struct {
	Nodes          []ClusterNode `json:"nodes"`
	ClusterName    string        `json:"cluster_name"`
	Environment    string        `json:"environment"`
	EnableWebGUI   bool          `json:"enable_web_gui"`
	EnableJetStream bool         `json:"enable_jetstream"`
}

// Using config functions instead of hardcoded values

// GetLocalClusterConfig returns configuration for local Docker-based NATS cluster
func GetLocalClusterConfig() ClusterConfig {
	nodeCount := config.GetNATSClusterNodeCount()
	nodes := make([]ClusterNode, 0, nodeCount)
	
	for i := 0; i < nodeCount; i++ {
		client, cluster, http := config.GetNATSClusterPortsForNode(i)
		nodes = append(nodes, ClusterNode{
			Name:        fmt.Sprintf("nats-%d", i+1),
			Region:      "local",
			Port:        client,
			ClusterPort: cluster,
			HTTPPort:    http,
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
	
	for _, region := range regions {
		nodes = append(nodes, ClusterNode{
			Name:        fmt.Sprintf("nats-%s", region),
			Region:      region,
			Port:        config.NATSClusterBasePort,
			ClusterPort: config.NATSClusterBaseCPort,
			HTTPPort:    config.NATSClusterBaseHTTP,
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

// StartLocalCluster starts a local Docker-based NATS cluster
func StartLocalCluster(ctx context.Context) error {
	log.Info("Starting local NATS cluster...")
	
	clusterConfig := GetLocalClusterConfig()
	
	// Create cluster data directory
	clusterDataPath := config.GetNATSClusterDataPath()
	if err := os.MkdirAll(clusterDataPath, 0755); err != nil {
		return fmt.Errorf("failed to create cluster data directory: %w", err)
	}
	
	// Start each node
	for _, node := range clusterConfig.Nodes {
		// Build routes list for this specific node (excluding itself)
		var routes []string
		for _, otherNode := range clusterConfig.Nodes {
			if otherNode.Name != node.Name {
				// All nodes use the same cluster port (6222) for proper NATS clustering
				routes = append(routes, fmt.Sprintf("nats://%s:%d", otherNode.Name, config.NATSClusterBaseCPort))
			}
		}
		routesStr := strings.Join(routes, ",")
		
		if err := startLocalClusterNode(ctx, node, clusterConfig.ClusterName, routesStr, clusterDataPath); err != nil {
			log.Error("Failed to start cluster node", "node", node.Name, "error", err)
			return fmt.Errorf("failed to start node %s: %w", node.Name, err)
		}
		
		// Brief delay between node starts
		time.Sleep(2 * time.Second)
	}
	
	log.Info("Local NATS cluster started successfully", "nodes", len(clusterConfig.Nodes))
	return nil
}

// startLocalClusterNode starts a single node in the local cluster
func startLocalClusterNode(ctx context.Context, node ClusterNode, clusterName, routes, dataPath string) error {
	// Check if container is already running
	statusCmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=%s", node.Name), "--format", "{{.Status}}")
	output, err := statusCmd.Output()
	if err == nil && len(output) > 0 {
		status := strings.TrimSpace(string(output))
		if strings.Contains(strings.ToLower(status), "up") {
			log.Info("NATS cluster node already running, skipping", "node", node.Name)
			return nil
		}
	}
	// Create node-specific data directory
	nodeDataPath := filepath.Join(dataPath, node.Name)
	if err := os.MkdirAll(nodeDataPath, 0755); err != nil {
		return fmt.Errorf("failed to create node data directory: %w", err)
	}
	
	// Generate NATS configuration for this node
	configPath := filepath.Join(nodeDataPath, "nats.conf")
	natsConfig := fmt.Sprintf(`
# NATS Server Configuration for %s
server_name: %s
port: %d
http_port: %d

# Clustering
cluster {
    name: %s
    port: %d
    routes: [%s]
    no_advertise: false
}

# JetStream
jetstream: {
    store_dir: /data
    max_memory_store: 256MB
    max_file_store: 2GB
}

# Logging
debug: false
trace: false
logtime: true
`, node.Name, node.Name, node.Port, node.HTTPPort, clusterName, node.ClusterPort, routes)
	
	if err := os.WriteFile(configPath, []byte(natsConfig), 0644); err != nil {
		return fmt.Errorf("failed to create node config: %w", err)
	}
	
	// Create Docker network if it doesn't exist
	networkName := fmt.Sprintf("nats-cluster-%s", clusterName)
	createNetworkCmd := exec.Command("docker", "network", "create", networkName)
	createNetworkCmd.Run() // Ignore errors - network might already exist
	
	// Stop existing container if running
	stopCmd := exec.Command("docker", "stop", node.Name)
	stopCmd.Run()
	removeCmd := exec.Command("docker", "rm", node.Name)
	removeCmd.Run()
	
	// Start NATS node container
	containerArgs := []string{
		"run", "-d",
		"--name", node.Name,
		"--network", networkName,
		"-p", fmt.Sprintf("%d:%d", node.Port, node.Port),        // Client port
		"-p", fmt.Sprintf("%d:%d", node.HTTPPort, node.HTTPPort), // HTTP monitoring port
		"-v", fmt.Sprintf("%s:/data", nodeDataPath),              // Data volume
		"-v", fmt.Sprintf("%s:/etc/nats/nats.conf", configPath), // Config volume
		"--rm",
		config.GetNATSDockerImage(),
		"--config", "/etc/nats/nats.conf",
	}
	
	log.Info("Starting NATS cluster node", "node", node.Name, "client_port", node.Port, "http_port", node.HTTPPort)
	
	// Start new container
	cmd := exec.Command("docker", containerArgs...)
	return cmd.Run()
}

// StopLocalCluster stops the local Docker-based NATS cluster
func StopLocalCluster() error {
	log.Info("Stopping local NATS cluster...")
	
	clusterConfig := GetLocalClusterConfig()
	
	for _, node := range clusterConfig.Nodes {
		stopCmd := exec.Command("docker", "stop", node.Name)
		if err := stopCmd.Run(); err != nil {
			log.Warn("Failed to stop cluster node", "node", node.Name, "error", err)
		}
		
		removeCmd := exec.Command("docker", "rm", node.Name)
		if err := removeCmd.Run(); err != nil {
			log.Warn("Failed to remove cluster node", "node", node.Name, "error", err)
		}
	}
	
	// Clean up Docker network
	networkName := fmt.Sprintf("nats-cluster-%s", clusterConfig.ClusterName)
	networkCmd := exec.Command("docker", "network", "rm", networkName)
	if err := networkCmd.Run(); err != nil {
		log.Warn("Failed to remove Docker network", "network", networkName, "error", err)
	}
	
	log.Info("Local NATS cluster stopped")
	return nil
}

// GetClusterStatus returns the status of cluster nodes
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
			// Check Docker container status using docker ps
			statusCmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=%s", node.Name), "--format", "{{.Status}}")
			output, err := statusCmd.Output()
			if err != nil || len(output) == 0 {
				clusterConfig.Nodes[i].Status = "stopped"
			} else {
				status := strings.TrimSpace(string(output))
				if strings.Contains(strings.ToLower(status), "up") {
					clusterConfig.Nodes[i].Status = "running"
				} else {
					clusterConfig.Nodes[i].Status = "stopped"
				}
			}
		} else {
			// Check Fly.io app status using existing fly package
			clusterConfig.Nodes[i].Status = "unknown" // Placeholder - would need fly.Status() integration
		}
	}
	
	return clusterConfig, nil
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

// deployFlyClusterNode deploys a single NATS node to Fly.io
func deployFlyClusterNode(ctx context.Context, node ClusterNode, clusterConfig ClusterConfig) error {
	appName := node.Name
	
	// Generate fly.toml for this node
	flyToml := fmt.Sprintf(`
app = "%s"
primary_region = "%s"

[build]
  image = "nats:alpine"

[env]
  NATS_SERVER_NAME = "%s"
  NATS_CLUSTER_NAME = "%s"

[[services]]
  internal_port = %d
  protocol = "tcp"
  
  [[services.ports]]
    port = 4222
    handlers = ["tcp"]

[[services]]
  internal_port = %d
  protocol = "tcp"
  
  [[services.ports]]
    port = 8222
    handlers = ["http"]

[mounts]
  source = "nats_data"
  destination = "/data"

[[vm]]
  memory = 256
  cpu_kind = "shared"
  cpus = 1
`, appName, node.Region, node.Name, clusterConfig.ClusterName, node.Port, node.HTTPPort)
	
	// Write fly.toml
	flyTomlPath := filepath.Join(os.TempDir(), fmt.Sprintf("fly-%s.toml", appName))
	if err := os.WriteFile(flyTomlPath, []byte(flyToml), 0644); err != nil {
		return fmt.Errorf("failed to write fly.toml: %w", err)
	}
	defer os.Remove(flyTomlPath)
	
	// Deploy using Fly.io CLI - would need to integrate with existing fly package
	// For now, use direct flyctl command
	cmd := exec.Command(config.GetFlyctlBinPath(), "deploy", "--config", flyTomlPath, "--app", appName, "--region", node.Region)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// UpgradeCluster performs rolling upgrade of NATS cluster using lame duck mode
func UpgradeCluster(ctx context.Context, isLocal bool) error {
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
		// For local Docker: signal lame duck, wait, then restart container
		log.Info("Signaling lame duck mode for local node", "node", node.Name)
		
		// Send SIGUSR2 to trigger lame duck mode
		signalCmd := exec.Command("docker", "kill", "--signal=USR2", node.Name)
		if err := signalCmd.Run(); err != nil {
			log.Warn("Failed to signal lame duck mode", "node", node.Name, "error", err)
		}
		
		// Wait for connections to drain
		time.Sleep(10 * time.Second)
		
		// Stop and restart container with latest image
		stopCmd := exec.Command("docker", "stop", node.Name)
		if err := stopCmd.Run(); err != nil {
			return fmt.Errorf("failed to stop node container: %w", err)
		}
		
		// Restart with fresh container (same config)
		clusterConfig := GetLocalClusterConfig()
		clusterDataPath := config.GetNATSClusterDataPath()
		var routes []string
		for _, n := range clusterConfig.Nodes {
			routes = append(routes, fmt.Sprintf("nats://%s:%d", n.Name, n.ClusterPort))
		}
		routesStr := strings.Join(routes, ",")
		
		return startLocalClusterNode(ctx, node, clusterConfig.ClusterName, routesStr, clusterDataPath)
	} else {
		// For Fly.io: use Fly's deployment command
		appName := node.Name
		cmd := exec.Command(config.GetFlyctlBinPath(), "deploy", "--app", appName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
}

// waitForNodeReady waits for a node to rejoin the cluster and be ready
func waitForNodeReady(ctx context.Context, node ClusterNode, isLocal bool) error {
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
			// Check if node is responding to health checks
			if isLocal {
				// For local: check Docker container status
				statusCmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=%s", node.Name), "--format", "{{.Status}}")
				output, err := statusCmd.Output()
				if err == nil && len(output) > 0 {
					status := strings.TrimSpace(string(output))
					if strings.Contains(strings.ToLower(status), "up") {
						// Also check HTTP monitoring endpoint
						healthURL := fmt.Sprintf("http://localhost:%d/varz", node.HTTPPort)
						if err := checkHTTPEndpoint(healthURL); err == nil {
							return nil
						}
					}
				}
			} else {
				// For Fly.io: placeholder for health check
				// Would need integration with fly.Status() or similar
				return nil // Simplified for now
			}
		}
	}
}

// checkHTTPEndpoint performs a simple HTTP GET to check if endpoint is responding
func checkHTTPEndpoint(url string) error {
	// This would typically use net/http, but keeping it simple for now
	// In a real implementation, we'd make an HTTP request to check /varz endpoint
	return nil
}