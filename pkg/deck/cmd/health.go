package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/joeblew999/infra/pkg/deck"
	"github.com/spf13/cobra"
)

// HealthCmd represents the health command
var HealthCmd = &cobra.Command{
	Use:   "health",
	Short: "Run health checks on deck system",
	Long: `Run comprehensive health checks on the deck system to verify:

- All tools are built and functional
- Complete .dsh → XML → SVG pipeline works
- System dependencies are available  
- Fonts and assets are accessible
- Output directories are writable

Examples:
  deck health                    # Run all health checks
  deck health --verbose          # Detailed output
  deck health --json             # JSON report format
  deck health --tool decksh      # Check specific tool only`,
	RunE: runHealthCheck,
}

func runHealthCheck(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	jsonOutput, _ := cmd.Flags().GetBool("json")
	toolFilter, _ := cmd.Flags().GetString("tool")

	checker := deck.NewHealthChecker()
	checker.SetVerbose(verbose)

	if toolFilter != "" {
		// Check specific tool only
		return runSingleToolCheck(checker, toolFilter, jsonOutput)
	}

	// Run full health check
	report := checker.RunFullHealthCheck()

	if jsonOutput {
		return outputJSONReport(report)
	}

	return outputHumanReport(report)
}

func runSingleToolCheck(checker *deck.HealthChecker, toolName string, jsonOutput bool) error {
	fmt.Printf("🔧 Health check: %s\n", toolName)
	
	if err := checker.ValidateTool(toolName); err != nil {
		if jsonOutput {
			result := map[string]any{
				"tool":    toolName,
				"status":  "unhealthy",
				"error":   err.Error(),
			}
			output, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(output))
		} else {
			fmt.Printf("❌ %s: %v\n", toolName, err)
			fmt.Printf("\n💡 Suggestion: Run './infra deck build install' to rebuild %s\n", toolName)
		}
		return err
	}

	if jsonOutput {
		result := map[string]any{
			"tool":   toolName,
			"status": "healthy",
		}
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))
	} else {
		fmt.Printf("✅ %s: Healthy\n", toolName)
	}

	return nil
}

func outputJSONReport(report *deck.HealthReport) error {
	output, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal health report: %w", err)
	}
	
	fmt.Println(string(output))
	return nil
}

func outputHumanReport(report *deck.HealthReport) error {
	// Header
	statusEmoji := getStatusEmoji(report.Overall)
	fmt.Printf("🏥 Deck Health Report %s\n", statusEmoji)
	fmt.Printf("══════════════════════════\n")
	fmt.Printf("Overall Status: %s\n", formatStatus(report.Overall))
	fmt.Printf("Check Duration: %s\n", report.Duration)
	fmt.Printf("Tools: %d/%d healthy\n", report.ToolsOK, report.ToolsTotal)
	
	if report.PipelineOK {
		fmt.Printf("Pipeline: ✅ Working\n")
	} else {
		fmt.Printf("Pipeline: ❌ Issues detected\n")
	}
	
	// Issues summary
	if len(report.Issues) == 0 {
		fmt.Printf("\n🎉 No issues detected! System is fully healthy.\n")
		return nil
	}

	fmt.Printf("\n📋 Issues Found (%d):\n", len(report.Issues))
	fmt.Printf("═══════════════════════\n")

	// Group issues by type
	issuesByType := make(map[string][]deck.HealthIssue)
	for _, issue := range report.Issues {
		issuesByType[issue.Type] = append(issuesByType[issue.Type], issue)
	}

	// Output by type
	typeOrder := []string{"dependency", "tool", "pipeline", "assets"}
	for _, issueType := range typeOrder {
		issues, exists := issuesByType[issueType]
		if !exists {
			continue
		}

		fmt.Printf("\n%s Issues:\n", formatIssueType(issueType))
		for _, issue := range issues {
			severityEmoji := getSeverityEmoji(issue.Severity)
			fmt.Printf("  %s %s", severityEmoji, issue.Message)
			if issue.Tool != "" {
				fmt.Printf(" [%s]", issue.Tool)
			}
			fmt.Printf("\n")
			
			if issue.Suggestion != "" {
				fmt.Printf("    💡 %s\n", issue.Suggestion)
			}
		}
	}

	// Overall recommendation
	fmt.Printf("\n🎯 Recommendations:\n")
	if report.Overall == "unhealthy" {
		fmt.Printf("   System has critical issues that prevent normal operation.\n")
		fmt.Printf("   Address error-level issues first, then rerun health check.\n")
	} else if report.Overall == "degraded" {
		fmt.Printf("   System is functional but has some issues.\n")
		fmt.Printf("   Consider addressing warnings to improve reliability.\n")
	}

	// Exit with error code if unhealthy
	if report.Overall == "unhealthy" {
		os.Exit(1)
	}

	return nil
}

// Helper functions for formatting

func getStatusEmoji(status string) string {
	switch status {
	case "healthy":
		return "✅"
	case "degraded":
		return "⚠️"
	case "unhealthy":
		return "❌"
	default:
		return "❓"
	}
}

func formatStatus(status string) string {
	switch status {
	case "healthy":
		return "✅ HEALTHY"
	case "degraded":
		return "⚠️ DEGRADED"  
	case "unhealthy":
		return "❌ UNHEALTHY"
	default:
		return "❓ UNKNOWN"
	}
}

func getSeverityEmoji(severity string) string {
	switch severity {
	case "error":
		return "❌"
	case "warning":
		return "⚠️"
	case "info":
		return "ℹ️"
	default:
		return "•"
	}
}

func formatIssueType(issueType string) string {
	switch issueType {
	case "dependency":
		return "🔗 System Dependency"
	case "tool":
		return "🔧 Tool"
	case "pipeline":
		return "🔄 Pipeline"
	case "assets":
		return "📁 Assets"
	default:
		return "❓ " + issueType
	}
}