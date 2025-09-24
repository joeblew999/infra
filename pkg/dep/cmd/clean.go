package cmd

import (
	"fmt"
	"os"

	dep "github.com/joeblew999/infra/pkg/dep"
)

func runClean() {
	fmt.Println("🧹 Cleaning dependency system...")

	if err := dep.CleanSystem(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error cleaning dependency system: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Dependency system cleaned successfully!")
	fmt.Println("\n🔄 Next steps:")
	fmt.Println("  • Run 'go run . tools dep local install' to reinstall binaries")
	fmt.Println("  • Generated code will be recreated on next service startup")
}
