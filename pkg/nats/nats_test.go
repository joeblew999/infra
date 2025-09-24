package nats

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/joeblew999/infra/pkg/config"
)

func TestNATSDataDirectory(t *testing.T) {
	// Test NATS data directory creation and isolation
	natsPath := config.GetNATSClusterDataPath()

	// Create NATS cluster directory
	err := os.MkdirAll(natsPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create NATS cluster data directory: %v", err)
	}

	t.Logf("✅ NATS cluster data directory created: %s", natsPath)

	// Create some test artifacts to simulate NATS storage
	testFiles := []string{
		"jetstream.db",
		"raft.log",
		"server.cfg",
		"cluster-state.json",
	}

	for _, filename := range testFiles {
		testFile := filepath.Join(natsPath, filename)
		err := os.WriteFile(testFile, []byte("test nats data"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
		t.Logf("✅ Test NATS file: %s", testFile)
	}

	t.Logf("📁 Test artifacts in: %s", natsPath)

	// Test individual node directories
	for i := 0; i < config.GetNATSClusterNodeCount(); i++ {
		nodeDir := filepath.Join(natsPath, "node-"+string(rune('0'+i)))
		err := os.MkdirAll(nodeDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create node directory: %v", err)
		}

		// Create node-specific files
		nodeFiles := []string{
			"nats.log",
			"jetstream.db",
		}

		for _, filename := range nodeFiles {
			nodeFile := filepath.Join(nodeDir, filename)
			err := os.WriteFile(nodeFile, []byte("node data"), 0644)
			if err != nil {
				t.Fatalf("Failed to create node file %s: %v", filename, err)
			}
		}

		t.Logf("✅ Node directory created: %s", nodeDir)
	}
}

func TestNATSClusterConfiguration(t *testing.T) {
	// Test NATS cluster configuration with test isolation
	clusterName := config.GetNATSClusterName()
	nodeCount := config.GetNATSClusterNodeCount()

	if clusterName == "" {
		t.Fatal("Cluster name is empty")
	}

	if nodeCount <= 0 {
		t.Fatal("Node count must be positive")
	}

	t.Logf("✅ Cluster name: %s", clusterName)
	t.Logf("✅ Node count: %d", nodeCount)

	// Test port configuration for each node
	for i := 0; i < nodeCount; i++ {
		client, cluster, http, leaf := config.GetNATSClusterPortsForNode(i)

		if client <= 0 || cluster <= 0 || http <= 0 || leaf <= 0 {
			t.Errorf("Invalid ports for node %d: client=%d, cluster=%d, http=%d, leaf=%d",
				i, client, cluster, http, leaf)
		}

		t.Logf("✅ Node %d ports - Client: %d, Cluster: %d, HTTP: %d, Leaf: %d",
			i, client, cluster, http, leaf)
	}

	// Test Fly.io regions
	regions := config.GetFlyRegions()
	if len(regions) == 0 {
		t.Fatal("No Fly.io regions configured")
	}

	for _, region := range regions {
		if region == "" {
			t.Error("Empty region in configuration")
		}
	}

	t.Logf("✅ Fly.io regions: %v", regions)

	// Save cluster config to test directory for inspection
	natsPath := config.GetNATSClusterDataPath()
	err := os.MkdirAll(natsPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create NATS directory: %v", err)
	}

	configFile := filepath.Join(natsPath, "test-cluster-config.txt")
	configData := "Test NATS Cluster Configuration:\n"
	configData += "Cluster Name: " + clusterName + "\n"
	configData += "Node Count: " + string(rune('0'+nodeCount)) + "\n"
	configData += "Docker Image: " + config.GetNATSDockerImage() + "\n"
	configData += "Data Path: " + natsPath + "\n"

	for i, region := range regions {
		configData += "Region " + string(rune('0'+i)) + ": " + region + "\n"
	}

	err = os.WriteFile(configFile, []byte(configData), 0644)
	if err != nil {
		t.Fatalf("Failed to save cluster config: %v", err)
	}

	t.Logf("✅ Cluster config saved: %s", configFile)
}

func TestEmbeddedNATSSetup(t *testing.T) {
	// Test embedded NATS setup with test isolation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test data path creation (without actually starting server)
	natsDataPath := filepath.Join(config.GetDataPath(), "nats")
	err := os.MkdirAll(natsDataPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create NATS data path: %v", err)
	}

	t.Logf("✅ NATS data path created: %s", natsDataPath)

	// Create mock server files to simulate embedded NATS
	serverFiles := []string{
		"nats-server.pid",
		"nats-server.log",
		"jetstream.db",
	}

	for _, filename := range serverFiles {
		serverFile := filepath.Join(natsDataPath, filename)
		err := os.WriteFile(serverFile, []byte("mock server data"), 0644)
		if err != nil {
			t.Fatalf("Failed to create server file %s: %v", filename, err)
		}
		t.Logf("✅ Mock server file: %s", serverFile)
	}

	// Test configuration values
	port := config.GetNATSPort()
	if port == "" {
		t.Fatal("NATS port is empty")
	}

	s3Port := config.GetNatsS3Port()
	if s3Port == "" {
		t.Fatal("NATS S3 port is empty")
	}

	t.Logf("✅ NATS port: %s", port)
	t.Logf("✅ NATS S3 port: %s", s3Port)

	// Save embedded NATS config to test directory
	configFile := filepath.Join(natsDataPath, "test-embedded-config.txt")
	configData := "Test Embedded NATS Configuration:\n"
	configData += "Data Path: " + natsDataPath + "\n"
	configData += "NATS Port: " + port + "\n"
	configData += "S3 Gateway Port: " + s3Port + "\n"
	if ctx.Err() != nil {
		configData += "Context Timeout: " + ctx.Err().Error() + "\n"
	} else {
		configData += "Context: Active\n"
	}

	err = os.WriteFile(configFile, []byte(configData), 0644)
	if err != nil {
		t.Fatalf("Failed to save embedded config: %v", err)
	}

	t.Logf("✅ Embedded config saved: %s", configFile)
}

func TestNATSEnvironmentIsolation(t *testing.T) {
	// Test that NATS uses test-isolated paths
	clusterDataPath := config.GetNATSClusterDataPath()

	// Verify test isolation
	if !filepath.HasPrefix(clusterDataPath, ".data-test") {
		t.Errorf("NATS cluster data path not test-isolated: %s", clusterDataPath)
	}

	generalDataPath := config.GetDataPath()
	if !filepath.HasPrefix(generalDataPath, ".data-test") {
		t.Errorf("General data path not test-isolated: %s", generalDataPath)
	}

	t.Logf("✅ Test-isolated cluster path: %s", clusterDataPath)
	t.Logf("✅ Test-isolated data path: %s", generalDataPath)

	// Create isolation test artifact
	isolationFile := filepath.Join(clusterDataPath, "isolation-test.txt")
	err := os.MkdirAll(filepath.Dir(isolationFile), 0755)
	if err != nil {
		t.Fatalf("Failed to create isolation directory: %v", err)
	}

	isolationData := "NATS Test Isolation Verification:\n"
	isolationData += "Environment: Test\n"
	isolationData += "Cluster Data Path: " + clusterDataPath + "\n"
	isolationData += "General Data Path: " + generalDataPath + "\n"
	isolationData += "Isolation: " + string(rune(map[bool]int{true: 1, false: 0}[filepath.HasPrefix(clusterDataPath, ".data-test")])) + "\n"

	err = os.WriteFile(isolationFile, []byte(isolationData), 0644)
	if err != nil {
		t.Fatalf("Failed to save isolation test: %v", err)
	}

	t.Logf("✅ Isolation test saved: %s", isolationFile)
}
