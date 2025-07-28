package conduit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/joeblew999/infra/pkg/dep"
)

// TestLoadCoreConfig tests loading the core configuration
func TestLoadCoreConfig(t *testing.T) {
	// Test with default core config
	config := getDefaultCoreConfig()
	
	// Verify Conduit core is present
	if config.Conduit.Name != "conduit" {
		t.Errorf("Expected conduit name to be 'conduit', got %s", config.Conduit.Name)
	}
	
	if config.Conduit.Version != "v0.12.1" {
		t.Errorf("Expected conduit version to be 'v0.12.1', got %s", config.Conduit.Version)
	}
	
	if config.Conduit.Type != "core" {
		t.Errorf("Expected conduit type to be 'core', got %s", config.Conduit.Type)
	}
}

// TestGetDefaultCoreConfig tests the default core configuration
func TestGetDefaultCoreConfig(t *testing.T) {
	config := getDefaultCoreConfig()
	
	// Test structure
	if config.Conduit.Name == "" {
		t.Error("Conduit name cannot be empty")
	}
	
	// Test JSON serialization
	data, err := json.Marshal(config)
	if err != nil {
		t.Errorf("Failed to marshal core config: %v", err)
	}
	
	var unmarshaledConfig struct {
		Conduit ConduitBinary `json:"conduit"`
	}
	
	if err := json.Unmarshal(data, &unmarshaledConfig); err != nil {
		t.Errorf("Failed to unmarshal core config: %v", err)
	}
	
	if unmarshaledConfig.Conduit.Name != config.Conduit.Name {
		t.Error("Core config serialization/deserialization failed")
	}
}

// TestGetDefaultConnectorsConfig tests the default connectors configuration
func TestGetDefaultConnectorsConfig(t *testing.T) {
	config := getDefaultConnectorsConfig()
	
	// Test structure
	if len(config.Connectors) < 2 {
		t.Errorf("Expected at least 2 connectors, got %d", len(config.Connectors))
	}
	
	// Test JSON serialization
	data, err := json.Marshal(config)
	if err != nil {
		t.Errorf("Failed to marshal connectors config: %v", err)
	}
	
	var unmarshaledConfig struct {
		Connectors []ConduitBinary `json:"connectors"`
	}
	
	if err := json.Unmarshal(data, &unmarshaledConfig); err != nil {
		t.Errorf("Failed to unmarshal connectors config: %v", err)
	}
	
	if len(unmarshaledConfig.Connectors) != len(config.Connectors) {
		t.Error("Connectors config serialization/deserialization failed")
	}
}

// TestGetDefaultProcessorsConfig tests the default processors configuration
func TestGetDefaultProcessorsConfig(t *testing.T) {
	config := getDefaultProcessorsConfig()
	
	// Test structure
	if config.Processors == nil {
		t.Error("Processors slice should not be nil")
	}
	
	// Test JSON serialization
	data, err := json.Marshal(config)
	if err != nil {
		t.Errorf("Failed to marshal processors config: %v", err)
	}
	
	var unmarshaledConfig struct {
		Processors []ConduitBinary `json:"processors"`
	}
	
	if err := json.Unmarshal(data, &unmarshaledConfig); err != nil {
		t.Errorf("Failed to unmarshal processors config: %v", err)
	}
	
	if len(unmarshaledConfig.Processors) != len(config.Processors) {
		t.Error("Processors config serialization/deserialization failed")
	}
}

// TestGet tests the Get function
func TestGet(t *testing.T) {
	path := Get("conduit")
	
	if path == "" {
		t.Error("Get should return a non-empty path")
	}
}

// TestConduitBinaryStructure tests the ConduitBinary structure
func TestConduitBinaryStructure(t *testing.T) {
	binary := ConduitBinary{
		Name:       "test-binary",
		Repo:       "test/repo",
		Version:    "v1.0.0",
		ReleaseURL: "https://github.com/test/repo/releases/tag/v1.0.0",
		Type:       "test",
		Assets: []dep.AssetSelector{
			{OS: "linux", Arch: "amd64", Match: "test-linux-amd64"},
		},
	}
	
	if binary.Name != "test-binary" {
		t.Errorf("Expected name 'test-binary', got %s", binary.Name)
	}
	
	if binary.Type != "test" {
		t.Errorf("Expected type 'test', got %s", binary.Type)
	}
	
	if len(binary.Assets) != 1 {
		t.Errorf("Expected 1 asset, got %d", len(binary.Assets))
	}
}

// TestCoreJSONFile tests the actual core.json file
func TestCoreJSONFile(t *testing.T) {
	configPath := filepath.Join("pkg", "conduit", "config", "core.json")
	
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skip("core.json file does not exist, skipping file test")
	}
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read core.json: %v", err)
	}
	
	var config struct {
		Conduit ConduitBinary `json:"conduit"`
	}
	
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse core.json: %v", err)
	}
	
	// Validate structure
	if config.Conduit.Name == "" {
		t.Error("Conduit name cannot be empty in core.json")
	}
	
	// Validate core binary
	binary := config.Conduit
	if binary.Name == "" {
		t.Error("Binary name cannot be empty")
	}
	if binary.Repo == "" {
		t.Errorf("Repo cannot be empty for %s", binary.Name)
	}
	if binary.Version == "" {
		t.Errorf("Version cannot be empty for %s", binary.Name)
	}
	if binary.Type == "" {
		t.Errorf("Type cannot be empty for %s", binary.Name)
	}
}

// TestConnectorsJSONFile tests the actual connectors.json file
func TestConnectorsJSONFile(t *testing.T) {
	configPath := filepath.Join("pkg", "conduit", "config", "connectors.json")
	
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skip("connectors.json file does not exist, skipping file test")
	}
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read connectors.json: %v", err)
	}
	
	var config struct {
		Connectors []ConduitBinary `json:"connectors"`
	}
	
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse connectors.json: %v", err)
	}
	
	// Validate structure
	if len(config.Connectors) == 0 {
		t.Error("No connectors found in connectors.json")
	}
	
	// Validate all connectors
	for _, binary := range config.Connectors {
		if binary.Name == "" {
			t.Error("Connector name cannot be empty")
		}
		if binary.Repo == "" {
			t.Errorf("Repo cannot be empty for %s", binary.Name)
		}
		if binary.Version == "" {
			t.Errorf("Version cannot be empty for %s", binary.Name)
		}
		if binary.Type == "" {
			t.Errorf("Type cannot be empty for %s", binary.Name)
		}
	}
}

// TestProcessorsJSONFile tests the actual processors.json file
func TestProcessorsJSONFile(t *testing.T) {
	configPath := filepath.Join("pkg", "conduit", "config", "processors.json")
	
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Skip("processors.json file does not exist, skipping file test")
	}
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read processors.json: %v", err)
	}
	
	var config struct {
		Processors []ConduitBinary `json:"processors"`
	}
	
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("Failed to parse processors.json: %v", err)
	}
	
	// Validate all processors
	for _, binary := range config.Processors {
		if binary.Name == "" {
			t.Error("Processor name cannot be empty")
		}
		if binary.Repo == "" {
			t.Errorf("Repo cannot be empty for %s", binary.Name)
		}
		if binary.Version == "" {
			t.Errorf("Version cannot be empty for %s", binary.Name)
		}
		if binary.Type == "" {
			t.Errorf("Type cannot be empty for %s", binary.Name)
		}
	}
}

// TestLoadConnectorsConfig tests loading the connectors configuration
func TestLoadConnectorsConfig(t *testing.T) {
	// Test with default connectors config
	config := getDefaultConnectorsConfig()
	
	// Verify connectors are present
	if len(config.Connectors) == 0 {
		t.Error("Expected at least one connector")
	}
	
	connectorMap := make(map[string]bool)
	for _, connector := range config.Connectors {
		connectorMap[connector.Name] = true
		
		if connector.Type != "connector" {
			t.Errorf("Expected connector %s type to be 'connector', got %s", connector.Name, connector.Type)
		}
		
		if connector.Repo == "" {
			t.Errorf("Connector %s repo cannot be empty", connector.Name)
		}
		
		if connector.Version == "" {
			t.Errorf("Connector %s version cannot be empty", connector.Name)
		}
	}
	
	// Verify expected connectors are present
	expectedConnectors := []string{"conduit-connector-s3", "conduit-connector-postgres"}
	for _, expected := range expectedConnectors {
		if !connectorMap[expected] {
			t.Errorf("Expected connector %s not found", expected)
		}
	}
}

// TestLoadProcessorsConfig tests loading the processors configuration
func TestLoadProcessorsConfig(t *testing.T) {
	// Test with default processors config
	config := getDefaultProcessorsConfig()
	
	// Verify processors slice is initialized (even if empty)
	if config.Processors == nil {
		t.Error("Expected processors slice to be initialized")
	}
	
	// Test that we can iterate over processors without issues
	for _, processor := range config.Processors {
		if processor.Type != "processor" {
			t.Errorf("Expected processor %s type to be 'processor', got %s", processor.Name, processor.Type)
		}
		
		if processor.Repo == "" {
			t.Errorf("Processor %s repo cannot be empty", processor.Name)
		}
		
		if processor.Version == "" {
			t.Errorf("Processor %s version cannot be empty", processor.Name)
		}
	}
}

// TestAssetSelectors tests the asset selector patterns
func TestAssetSelectors(t *testing.T) {
	coreConfig := getDefaultCoreConfig()
	connectorsConfig := getDefaultConnectorsConfig()
	processorsConfig := getDefaultProcessorsConfig()
	
	// Test Conduit assets
	if len(coreConfig.Conduit.Assets) == 0 {
		t.Error("Conduit should have assets defined")
	}
	
	// Test that each asset has required fields
	for _, asset := range coreConfig.Conduit.Assets {
		if asset.OS == "" {
			t.Error("Asset OS cannot be empty")
		}
		if asset.Arch == "" {
			t.Error("Asset Arch cannot be empty")
		}
		if asset.Match == "" {
			t.Error("Asset Match pattern cannot be empty")
		}
	}
	
	// Test connector assets
	for _, connector := range connectorsConfig.Connectors {
		if len(connector.Assets) == 0 {
			t.Errorf("Connector %s should have assets defined", connector.Name)
		}
	}
	
	// Test processor assets
	for _, processor := range processorsConfig.Processors {
		if len(processor.Assets) == 0 {
			t.Errorf("Processor %s should have assets defined", processor.Name)
		}
	}
}

// TestPackageIntegration tests basic package integration
func TestPackageIntegration(t *testing.T) {
	// Test that we can call functions without panicking
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Package integration test panicked: %v", r)
		}
	}()
	
	// Test basic function calls
	_ = Get("test")
	_ = getDefaultCoreConfig()
	_ = getDefaultConnectorsConfig()
	_ = getDefaultProcessorsConfig()
	
	// Test that Ensure doesn't crash (though it may not work fully due to missing implementation)
	// This is just to ensure basic structure works
	_ = Ensure(false)
}

// BenchmarkGet tests performance of Get function
func BenchmarkGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Get("conduit")
	}
}

// BenchmarkLoadCoreConfig tests performance of loading core configuration
func BenchmarkLoadCoreConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = getDefaultCoreConfig()
	}
}

// BenchmarkLoadConnectorsConfig tests performance of loading connectors configuration
func BenchmarkLoadConnectorsConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = getDefaultConnectorsConfig()
	}
}

// BenchmarkLoadProcessorsConfig tests performance of loading processors configuration
func BenchmarkLoadProcessorsConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = getDefaultProcessorsConfig()
	}
}

// TestJSONRoundTrip tests JSON serialization/deserialization
func TestJSONRoundTrip(t *testing.T) {
	// Test core config round-trip
	coreConfig := getDefaultCoreConfig()
	coreData, err := json.Marshal(coreConfig)
	if err != nil {
		t.Fatalf("Failed to marshal core config: %v", err)
	}
	
	var unmarshaledCore struct {
		Conduit ConduitBinary `json:"conduit"`
	}
	
	if err := json.Unmarshal(coreData, &unmarshaledCore); err != nil {
		t.Fatalf("Failed to unmarshal core config: %v", err)
	}
	
	if unmarshaledCore.Conduit.Name != coreConfig.Conduit.Name {
		t.Error("JSON round-trip failed for core config")
	}
	
	// Test connectors config round-trip
	connectorsConfig := getDefaultConnectorsConfig()
	connectorsData, err := json.Marshal(connectorsConfig)
	if err != nil {
		t.Fatalf("Failed to marshal connectors config: %v", err)
	}
	
	var unmarshaledConnectors struct {
		Connectors []ConduitBinary `json:"connectors"`
	}
	
	if err := json.Unmarshal(connectorsData, &unmarshaledConnectors); err != nil {
		t.Fatalf("Failed to unmarshal connectors config: %v", err)
	}
	
	if len(unmarshaledConnectors.Connectors) != len(connectorsConfig.Connectors) {
		t.Error("JSON round-trip failed for connectors config")
	}
	
	// Test processors config round-trip
	processorsConfig := getDefaultProcessorsConfig()
	processorsData, err := json.Marshal(processorsConfig)
	if err != nil {
		t.Fatalf("Failed to marshal processors config: %v", err)
	}
	
	var unmarshaledProcessors struct {
		Processors []ConduitBinary `json:"processors"`
	}
	
	if err := json.Unmarshal(processorsData, &unmarshaledProcessors); err != nil {
		t.Fatalf("Failed to unmarshal processors config: %v", err)
	}
	
	if len(unmarshaledProcessors.Processors) != len(processorsConfig.Processors) {
		t.Error("JSON round-trip failed for processors config")
	}
}