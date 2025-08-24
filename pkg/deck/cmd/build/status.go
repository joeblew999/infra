package build

import (
	"fmt"

	"github.com/joeblew999/infra/pkg/deck"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show build status of deck tools",
	Long:  "Show which deck tools are built and their locations, with optional health checks",
	Run: func(cmd *cobra.Command, args []string) {
		withHealth, _ := cmd.Flags().GetBool("health")
		
		manager := deck.NewManager()
		status := manager.Status()
		
		if len(status) == 0 {
			fmt.Println("No deck tools found. Run './infra deck build install' to build them.")
			return
		}
		
		fmt.Println("Deck Tools Status:")
		fmt.Println("==================")
		
		for tool, paths := range status {
			fmt.Printf("\n%s:\n", tool)
			fmt.Printf("  Binary: %s\n", paths["binary"])
			fmt.Printf("  WASM:   %s\n", paths["wasm"])
			
			// Add health check if requested
			if withHealth {
				checker := deck.NewHealthChecker()
				if err := checker.ValidateTool(tool); err != nil {
					fmt.Printf("  Health: ❌ %v\n", err)
				} else {
					fmt.Printf("  Health: ✅ OK\n")
				}
			}
		}
		
		// Run pipeline health check if requested
		if withHealth {
			fmt.Printf("\nPipeline Health:\n")
			fmt.Printf("================\n")
			
			checker := deck.NewHealthChecker()
			pipelineOK, issues := checker.TestPipeline()
			
			if pipelineOK {
				fmt.Printf("✅ Complete .dsh → XML → SVG pipeline working\n")
			} else {
				fmt.Printf("❌ Pipeline issues detected:\n")
				for _, issue := range issues {
					fmt.Printf("   • %s\n", issue.Message)
				}
			}
		}
		
		if withHealth {
			fmt.Printf("\n💡 Run './infra deck health' for comprehensive health checks\n")
		}
	},
}

func init() {
	BuildCmd.AddCommand(statusCmd)
	statusCmd.Flags().BoolP("health", "h", false, "Include health checks in status report")
}