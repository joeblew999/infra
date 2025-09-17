package cmd

import (
	"fmt"

	"github.com/joeblew999/infra/pkg/bento"
	caddycmd "github.com/joeblew999/infra/pkg/caddy/cmd"
	"github.com/joeblew999/infra/pkg/conduit"
	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/dep"
	"github.com/joeblew999/infra/pkg/mox"
	"github.com/joeblew999/infra/pkg/nats"
	"github.com/joeblew999/infra/pkg/pocketbase"
	"github.com/spf13/cobra"

	"github.com/joeblew999/infra/pkg/workflows" // New import

	// Import deck package to ensure all deck commands are loaded
	_ "github.com/joeblew999/infra/pkg/deck/cmd"
)

// caddyCmd is now delegated to the caddy package
var caddyCmd = caddycmd.GetCaddyCmd()

var tofuCmd = &cobra.Command{
	Use:                "tofu",
	Short:              "Run tofu commands",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return ExecuteBinary(config.GetTofuBinPath(), args...)
	},
}

var taskCmd = &cobra.Command{
	Use:                "task",
	Short:              "Run task commands",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return ExecuteBinary(config.GetTaskBinPath(), args...)
	},
}

var koCmd = &cobra.Command{
	Use:                "ko",
	Short:              "Run ko commands",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return ExecuteBinary(config.GetKoBinPath(), args...)
	},
}

var flyctlCmd = &cobra.Command{
	Use:                "flyctl",
	Short:              "Run flyctl commands",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return ExecuteBinary(config.GetFlyctlBinPath(), args...)
	},
}

// generateCmd is the parent command for all code generation tools
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Code generation tools",
	Long:  `Tools for generating various types of code within the infra project.`,
}

// gozeroApiGenerateCmd generates go-zero API code
var gozeroApiGenerateCmd = &cobra.Command{
	Use:   "gozero-api",
	Short: "Generate go-zero API code",
	Long:  `Generate go-zero API code from a .api definition file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiFile, _ := cmd.Flags().GetString("api")
		outputDir, _ := cmd.Flags().GetString("output")

		if apiFile == "" {
			return fmt.Errorf("missing --api flag")
		}
		if outputDir == "" {
			return fmt.Errorf("missing --output flag")
		}

		return workflows.GenerateGoZeroCode(cmd.Context(), apiFile, outputDir)
	},
}

func newMoxCmd() *cobra.Command {
	moxCmd := &cobra.Command{
		Use:   "mox",
		Short: "Manage mox mail server",
		Long: `Commands for managing the mox mail server:

  start          Start the mox mail server
  init           Initialize the mox mail server`,
	}

	moxCmd.AddCommand(
		newMoxStartCmd(),
		newMoxInitCmd(),
	)

	return moxCmd
}

func newMoxStartCmd() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the mox mail server",
		RunE: func(cmd *cobra.Command, args []string) error {
			domain, _ := cmd.Flags().GetString("domain")
			adminEmail, _ := cmd.Flags().GetString("admin-email")
			return mox.StartSupervised(domain, adminEmail)
		},
	}

	startCmd.Flags().String("domain", "localhost", "Domain for the mail server")
	startCmd.Flags().String("admin-email", "admin@localhost", "Admin email for the mail server")

	return startCmd
}

func newMoxInitCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the mox mail server",
		RunE: func(cmd *cobra.Command, args []string) error {
			domain, _ := cmd.Flags().GetString("domain")
			adminEmail, _ := cmd.Flags().GetString("admin-email")
			server := mox.NewServer(domain, adminEmail)
			return server.Init()
		},
	}

	initCmd.Flags().String("domain", "localhost", "Domain for the mail server")
	initCmd.Flags().String("admin-email", "admin@localhost", "Admin email for the mail server")

	return initCmd
}

// cliCmd is the parent command for all CLI tool wrappers
var cliCmd = &cobra.Command{
	Use:   "cli",
	Short: "CLI tool wrappers",
	Long: `Access to various CLI tools and utilities:

SCALING & DEPLOYMENT:
  fly              Fly.io operations and scaling
  
DEVELOPMENT TOOLS:
  deck             Deck visualization tools  
  gozero           Go-zero microservices operations
  toki             Translation and i18n workflow
  xtemplate        HTML/template web development server
  
BINARY TOOLS:
  tofu             OpenTofu infrastructure as code
  task             Task runner and build automation
  caddy            Web server with automatic HTTPS
  ko               Container image builder for Go
  flyctl           Direct flyctl commands
  nats             NATS messaging operations
  bento            Stream processing operations
  mox              Mox mail server
  utm              UTM virtual machine management (macOS)

Use "infra cli [tool] --help" for detailed information about each tool.`,
}

// RunCLI adds the CLI parent command and management commands to the root.
func RunCLI() {
	// Add tool wrappers under 'cli' parent command
	cliCmd.AddCommand(tofuCmd)
	cliCmd.AddCommand(taskCmd)
	cliCmd.AddCommand(caddyCmd)
	cliCmd.AddCommand(koCmd)
	cliCmd.AddCommand(flyctlCmd)
	cliCmd.AddCommand(nats.NewNATSCmd())
	cliCmd.AddCommand(bento.NewBentoCmd())
	cliCmd.AddCommand(newMoxCmd())

	// Add development/build workflow tools
	AddWorkflowsToCLI(cliCmd)

	// Add AI commands
	AddAIToCLI(cliCmd)

	// Add debug and translation tools
	AddDebugToCLI(cliCmd)
	AddTokiToCLI(cliCmd)

	// Add deck presentation tools
	AddDeckToCLI(cliCmd)

	// Add font management tools
	AddFontToCLI(cliCmd)

	// Add go-zero microservices tools
	AddGoZeroToCLI(cliCmd)

	// Add xtemplate web development tools
	AddXTemplateToCLI(cliCmd)

	// Add UTM virtual machine management
	AddUTMToCLI(cliCmd)

	// Add goreman process supervision utilities
	AddGoremanToCLI(cliCmd)

	// Add generate command
	cliCmd.AddCommand(generateCmd) // Add this line

	// Add CLI parent command to root
	rootCmd.AddCommand(cliCmd)

	// Keep essential management commands at root level
	rootCmd.AddCommand(dep.Cmd)
	rootCmd.AddCommand(config.Cmd)

	// Move these to cli namespace
	cliCmd.AddCommand(pocketbase.Cmd)
	cliCmd.AddCommand(conduit.Cmd)
}
