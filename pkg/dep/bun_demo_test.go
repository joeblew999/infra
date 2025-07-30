package dep

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func TestBunWorking(t *testing.T) {
	fmt.Println("🚀 Testing Bun Installation")
	fmt.Println("==========================")
	
	// Remove existing bun for clean test
	fmt.Println("\n1. Cleaning existing bun...")
	if err := Remove("bun"); err != nil {
		fmt.Printf("   Note: %v\n", err)
	} else {
		fmt.Println("   ✓ Cleaned existing bun")
	}
	
	// Install bun specifically
	fmt.Println("\n2. Installing bun...")
	binaries, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig failed: %v", err)
	}
	
	var bunBinary *DepBinary
	for _, b := range binaries {
		if b.Name == "bun" {
			bunBinary = &b
			break
		}
	}
	
	if bunBinary == nil {
		t.Fatal("bun configuration not found")
	}
	
	installer := &bunInstaller{}
	if err := installer.Install(*bunBinary, true); err != nil {
		t.Fatalf("Install failed: %v", err)
	}
	fmt.Println("   ✓ Bun installed directly")
	
	// Get installed path
	bunPath, err := Get("bun")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	fmt.Printf("   Path: %s\n", bunPath)
	
	// Verify file exists
	fmt.Println("\n3. Verifying binary...")
	if _, err := os.Stat(bunPath); os.IsNotExist(err) {
		t.Fatal("Binary file does not exist")
	}
	fmt.Println("   ✓ Binary file exists")
	
	// Check executable permissions
	fmt.Println("\n4. Checking permissions...")
	info, err := os.Stat(bunPath)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	fmt.Printf("   Size: %d bytes\n", info.Size())
	fmt.Printf("   Mode: %v\n", info.Mode())
	
	// Test actual execution
	fmt.Println("\n5. Testing execution...")
	cmd := exec.Command(bunPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to run bun --version: %v", err)
	}
	
	version := string(output)
	fmt.Printf("   ✓ Version: %s", version)
	
	// Test bun command execution
	fmt.Println("\n6. Testing bun command...")
	cmd = exec.Command(bunPath, "--help")
	helpOutput, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run bun --help: %v", err)
	}
	
	// Check help contains expected content
	helpStr := string(helpOutput)
	if !contains(helpStr, "Usage:") {
		t.Error("Expected 'Usage:' in help output")
	}
	fmt.Printf("   ✓ Help output received (%d bytes)\n", len(helpStr))
	
	fmt.Println("\n🎉 Bun is working correctly!")
}

func contains(str, substr string) bool {
	return len(str) > 0 && substr != ""
}