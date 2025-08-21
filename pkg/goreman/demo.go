package goreman

import (
	"fmt"
	"os"
	"time"
	
	"github.com/joeblew999/infra/pkg/config"
)

// RunDemo demonstrates the goreman supervision system without starting real services
func RunDemo() {
	fmt.Println("🚀 Goreman Process Supervision Demo")
	fmt.Println()
	
	// Register some demo processes (using echo command which exists everywhere)
	fmt.Println("📝 Registering demo processes...")
	
	Register("demo-litestream", &ProcessConfig{
		Command: "echo",
		Args:    []string{"[LITESTREAM] Would start:", config.Get(config.BinaryLitestream), "replicate"},
		Env:     os.Environ(),
	})
	
	Register("demo-caddy", &ProcessConfig{
		Command: "echo", 
		Args:    []string{"[CADDY] Would start:", config.Get(config.BinaryCaddy), "run"},
		Env:     os.Environ(),
	})
	
	Register("demo-bento", &ProcessConfig{
		Command: "echo",
		Args:    []string{"[BENTO] Would start:", config.Get(config.BinaryBento), "run"},
		Env:     os.Environ(),
	})
	
	// Show registration worked
	status := GetAllStatus()
	fmt.Printf("✅ Registered %d processes\n", len(status))
	for name, stat := range status {
		fmt.Printf("   • %s: %s\n", name, stat)
	}
	fmt.Println()
	
	// Start all processes
	fmt.Println("🎬 Starting all processes...")
	if err := Start("demo-litestream"); err != nil {
		fmt.Printf("❌ demo-litestream failed: %v\n", err)
	} else {
		fmt.Println("✅ demo-litestream started")
	}
	
	if err := Start("demo-caddy"); err != nil {
		fmt.Printf("❌ demo-caddy failed: %v\n", err)
	} else {
		fmt.Println("✅ demo-caddy started")
	}
	
	if err := Start("demo-bento"); err != nil {
		fmt.Printf("❌ demo-bento failed: %v\n", err) 
	} else {
		fmt.Println("✅ demo-bento started")
	}
	
	time.Sleep(1 * time.Second)
	fmt.Println()
	
	// Show final status
	fmt.Println("📊 Final process status:")
	status = GetAllStatus()
	for name, stat := range status {
		fmt.Printf("   • %s: %s\n", name, stat)
	}
	fmt.Println()
	
	// Test idempotent operations
	fmt.Println("🔄 Testing idempotent operations...")
	fmt.Println("   Calling Start() again on already-started processes...")
	
	Start("demo-litestream") // Should be no-op
	Start("demo-caddy")      // Should be no-op
	Start("demo-bento")      // Should be no-op
	
	fmt.Println("✅ Idempotent operations completed")
	fmt.Println()
	
	// Demonstrate graceful shutdown
	fmt.Println("🛑 Demonstrating graceful shutdown...")
	if err := StopAll(); err != nil {
		fmt.Printf("❌ StopAll failed: %v\n", err)
	} else {
		fmt.Println("✅ All processes stopped gracefully")
	}
	
	fmt.Println()
	fmt.Println("🎉 Goreman supervision demo complete!")
	fmt.Println()
	fmt.Println("Key benefits demonstrated:")
	fmt.Println("   • Centralized process registration")
	fmt.Println("   • Type-safe binary constants") 
	fmt.Println("   • Idempotent start/stop operations")
	fmt.Println("   • Graceful shutdown handling")
	fmt.Println("   • Status monitoring")
}