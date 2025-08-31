package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/gozero"
	"github.com/spf13/cobra"
)

// goZeroCmd represents the gozero command
var goZeroCmd = &cobra.Command{
	Use:   "gozero",
	Short: "Go-zero microservices framework operations",
	Long: `Generate and manage go-zero microservices with infra patterns.

Examples:
  infra gozero api create myservice --mcp --output ./api/myservice
  infra gozero api generate myservice.api --swagger --output .
  infra gozero quickstart mono --output ./my-project`,
}

var goZeroApiCmd = &cobra.Command{
	Use:   "api",
	Short: "Go-zero API operations",
	Long:  "Generate, validate, and manage go-zero API services.",
}

var goZeroApiCreateCmd = &cobra.Command{
	Use:   "create [service-name]",
	Short: "Create a new go-zero API service",
	Long: `Create a new go-zero API service following infra patterns.

Use --mcp flag to create MCP-compatible service.
Use --output to specify output directory.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceName := args[0]
		
		mcp, _ := cmd.Flags().GetBool("mcp")
		output, _ := cmd.Flags().GetString("output")
		debug, _ := cmd.Flags().GetBool("debug")
		description, _ := cmd.Flags().GetString("description")
		
		if output == "" {
			output = filepath.Join("api", serviceName)
		}
		
		// Create output directory
		if err := os.MkdirAll(output, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
			os.Exit(1)
		}
		
		service := gozero.NewService(debug)
		ctx := context.Background()
		
		var err error
		if mcp {
			if description == "" {
				description = fmt.Sprintf("MCP integration service for %s", serviceName)
			}
			err = service.CreateMCPAPI(ctx, serviceName, description, output)
		} else {
			// For now, create a basic API with health endpoint
			endpoints := []gozero.Endpoint{
				{
					Name:         "Health",
					Method:       "get",
					Path:         "/health",
					ResponseType: "HealthResponse",
					Handler:      "HealthHandler",
				},
			}
			err = service.CreateStandardAPI(ctx, serviceName, endpoints, output)
		}
		
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating API service: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("✓ Created go-zero API service: %s\n", serviceName)
		fmt.Printf("📁 Output directory: %s\n", output)
		fmt.Printf("🚀 Start server: cd %s && go run %s.go\n", output, serviceName)
	},
}

var goZeroApiGenerateCmd = &cobra.Command{
	Use:   "generate [api-file]",
	Short: "Generate go-zero service from API file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		apiFile := args[0]
		
		output, _ := cmd.Flags().GetString("output")
		swagger, _ := cmd.Flags().GetBool("swagger")
		debug, _ := cmd.Flags().GetBool("debug")
		
		if output == "" {
			output = "."
		}
		
		runner := gozero.NewGoZeroRunner(debug)
		runner.SetWorkDir(output)
		
		// Generate service
		if err := runner.ApiGenerate(apiFile, output); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating service: %v\n", err)
			os.Exit(1)
		}
		
		// Generate swagger if requested
		if swagger {
			if err := runner.ApiSwagger(apiFile, output); err != nil {
				fmt.Printf("Warning: Failed to generate Swagger docs: %v\n", err)
			}
		}
		
		fmt.Printf("✓ Generated go-zero service from %s\n", apiFile)
	},
}

var goZeroQuickstartCmd = &cobra.Command{
	Use:   "quickstart [type]",
	Short: "Create quickstart project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceType := args[0] // mono or micro
		
		output, _ := cmd.Flags().GetString("output")
		debug, _ := cmd.Flags().GetBool("debug")
		
		if output == "" {
			output = fmt.Sprintf("./go-zero-%s-project", serviceType)
		}
		
		runner := gozero.NewGoZeroRunner(debug)
		
		if err := runner.QuickStart(serviceType, output); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating quickstart project: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("✓ Created go-zero %s project\n", serviceType)
		fmt.Printf("📁 Output directory: %s\n", output)
	},
}

func init() {
	// Add to root command
	rootCmd.AddCommand(goZeroCmd)
	
	// Add subcommands
	goZeroCmd.AddCommand(goZeroApiCmd)
	goZeroCmd.AddCommand(goZeroQuickstartCmd)
	
	// Add API subcommands
	goZeroApiCmd.AddCommand(goZeroApiCreateCmd)
	goZeroApiCmd.AddCommand(goZeroApiGenerateCmd)
	
	// API create flags
	goZeroApiCreateCmd.Flags().Bool("mcp", false, "Create MCP-compatible service")
	goZeroApiCreateCmd.Flags().String("output", "", "Output directory")
	goZeroApiCreateCmd.Flags().String("description", "", "Service description")
	
	// API generate flags
	goZeroApiGenerateCmd.Flags().String("output", ".", "Output directory")
	goZeroApiGenerateCmd.Flags().Bool("swagger", false, "Generate Swagger documentation")
	
	// Quickstart flags
	goZeroQuickstartCmd.Flags().String("output", "", "Output directory")
}