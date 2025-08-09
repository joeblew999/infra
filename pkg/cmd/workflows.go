package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/joeblew999/infra/pkg/fly"
	"github.com/joeblew999/infra/pkg/mcp"
	"github.com/joeblew999/infra/pkg/workflows"
	"github.com/spf13/cobra"
)

// deployCmd provides idempotent deployment workflow
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy application using idempotent workflow",
	Long: `Deploy the application to Fly.io using an idempotent workflow that:
- Ensures all prerequisites are met
- Creates app and volume if needed
- Builds container image with ko
- Deploys and verifies the deployment

This command is safe to run multiple times.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse flags
		appName, _ := cmd.Flags().GetString("app")
		region, _ := cmd.Flags().GetString("region")
		environment, _ := cmd.Flags().GetString("env")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		force, _ := cmd.Flags().GetBool("force")

		// Create and run workflow
		workflow := workflows.NewDeployWorkflow(workflows.DeployOptions{
			AppName:     appName,
			Region:      region,
			Environment: environment,
			DryRun:      dryRun,
			Force:       force,
		})

		return workflow.Execute()
	},
}

// buildCmd provides standardized build workflow
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build application using standardized workflow",
	Long: `Build the application container image using ko with standardized settings:
- Optimized for size and security
- Multi-platform support
- Consistent tagging and metadata`,
	RunE: func(cmd *cobra.Command, args []string) error {
		push, _ := cmd.Flags().GetBool("push")
		platform, _ := cmd.Flags().GetString("platform")
		repo, _ := cmd.Flags().GetString("repo")
		tag, _ := cmd.Flags().GetString("tag")

		// TODO: Implement container build workflow
		workflow := workflows.NewContainerBuildWorkflow(workflows.ContainerBuildOptions{
			Push:     push,
			Platform: platform,
			Repo:     repo,
			Tag:      tag,
			DryRun:   false,
		})
		
		_, err := workflow.Execute()
		return err
	},
}

// statusCmd provides deployment status and health checks
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check deployment status and health",
	Long: `Check the status of deployed applications including:
- Fly.io app status
- Health check results
- Resource usage
- Recent logs`,
	RunE: func(cmd *cobra.Command, args []string) error {
		appName, _ := cmd.Flags().GetString("app")
		verbose, _ := cmd.Flags().GetBool("verbose")
		logs, _ := cmd.Flags().GetInt("logs")

		// TODO: Implement status workflow
		fmt.Printf("Checking status for: app=%s, verbose=%v, logs=%d\n", 
			appName, verbose, logs)
		
		return fmt.Errorf("status workflow not implemented yet")
	},
}

// initCmd initializes a new project with standard configuration
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize new project with standard configuration",
	Long: `Initialize a new project with standard configuration files:
- fly.toml with best practices
- .ko.yaml for optimized builds
- GitHub Actions workflow
- Documentation templates`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		template, _ := cmd.Flags().GetString("template")
		force, _ := cmd.Flags().GetBool("force")

		// TODO: Implement init workflow
		fmt.Printf("Initializing project: name=%s, template=%s, force=%v\n", 
			name, template, force)
		
		return fmt.Errorf("init workflow not implemented yet")
	},
}

// preCommitCmd runs pre-commit checks
var preCommitCmd = &cobra.Command{
	Use:   "pre-commit",
	Short: "Run pre-commit checks",
	Long: `Run pre-commit checks including:
- API compatibility check
- Documentation quality validation
- Go file formatting and linting`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPreCommitChecks()
	},
}

// ciCmd runs CI checks
var ciCmd = &cobra.Command{
	Use:   "ci",
	Short: "Run CI checks",
	Long: `Run CI checks including:
- API compatibility verification
- Documentation quality check
- Full test suite`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCIChecks()
	},
}

// binaryCmd builds cross-platform binaries
var binaryCmd = &cobra.Command{
	Use:   "binary",
	Short: "Build cross-platform binaries",
	Long: `Build cross-platform binaries for the application:
- Uses .bin directory for output (configurable)
- Supports Windows, Darwin, Linux on arm64/amd64
- Follows BINARYNAME_OS_ARCH naming convention
- Static linking for portability`,
	RunE: func(cmd *cobra.Command, args []string) error {
		outputDir, _ := cmd.Flags().GetString("output")
		binaryName, _ := cmd.Flags().GetString("name")
		all, _ := cmd.Flags().GetBool("all")
		localOnly, _ := cmd.Flags().GetBool("local")
		platforms, _ := cmd.Flags().GetStringSlice("platforms")
		architectures, _ := cmd.Flags().GetStringSlice("arch")
		verbose, _ := cmd.Flags().GetBool("verbose")

		workflow := workflows.NewBinaryBuildWorkflow(workflows.BinaryBuildOptions{
			OutputDir:    outputDir,
			BinaryName:   binaryName,
			Platforms:    platforms,
			Environments: architectures,
			BuildAll:     all,
			LocalOnly:    localOnly,
			Verbose:      verbose,
		})

		return workflow.Execute()
	},
}

// devCmd starts development mode
var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Start development mode",
	Long: `Start development mode with:
- File watching
- Automatic rebuilds
- Hot reload capabilities`,
	RunE: func(cmd *cobra.Command, args []string) error {
		watch, _ := cmd.Flags().GetBool("watch")
		return runDevMode(watch)
	},
}

// litestreamCmd provides Litestream database replication commands
var litestreamCmd = &cobra.Command{
	Use:   "litestream",
	Short: "Manage SQLite database replication with Litestream",
	Long: `Manage SQLite database replication using Litestream for continuous backups.

This provides stateless deployment capabilities by automatically backing up
SQLite databases to local filesystem or cloud storage, with point-in-time recovery.`,
}

// litestreamStartCmd starts Litestream replication
var litestreamStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Litestream replication",
	Long: `Start continuous replication of SQLite databases using Litestream.

Uses local filesystem by default (no S3 required). Example:
  go run . litestream start --db ./pb_data/data.db --backup ./backups/data.db`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, _ := cmd.Flags().GetString("db")
		backupPath, _ := cmd.Flags().GetString("backup")
		config, _ := cmd.Flags().GetString("config")
		verbose, _ := cmd.Flags().GetBool("verbose")

		return runLitestreamStart(dbPath, backupPath, config, verbose)
	},
}

// litestreamRestoreCmd restores database from backup
var litestreamRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore database from Litestream backup",
	Long: `Restore SQLite database from Litestream backup.

Example:
  go run . litestream restore --db ./pb_data/data.db --backup ./backups/data.db`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dbPath, _ := cmd.Flags().GetString("db")
		backupPath, _ := cmd.Flags().GetString("backup")
		config, _ := cmd.Flags().GetString("config")
		timestamp, _ := cmd.Flags().GetString("timestamp")

		return runLitestreamRestore(dbPath, backupPath, config, timestamp)
	},
}

// litestreamStatusCmd shows replication status
var litestreamStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show Litestream replication status",
	Long: `Show current replication status and backup information.

Example:
  go run . litestream status --config ./litestream.yml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, _ := cmd.Flags().GetString("config")
		return runLitestreamStatus(config)
	},
}

func init() {
	// Deploy command flags
	deployCmd.Flags().StringP("app", "a", "", "Fly.io app name (default: from env FLY_APP_NAME or 'infra-mgmt')")
	deployCmd.Flags().StringP("region", "r", "", "Fly.io region (default: from env FLY_REGION or 'syd')")
	deployCmd.Flags().StringP("env", "e", "", "Environment (development/production, auto-detected)")
	deployCmd.Flags().Bool("dry-run", false, "Show what would be done without executing")
	deployCmd.Flags().Bool("force", false, "Force deployment even if no changes detected")

	// Build command flags
	buildCmd.Flags().Bool("push", true, "Push image to registry")
	buildCmd.Flags().String("platform", "linux/amd64", "Target platform")
	buildCmd.Flags().String("repo", "", "Container repository (default: auto-detected)")
	buildCmd.Flags().StringP("tag", "t", "", "Image tag (default: auto-generated)")

	// Status command flags
	statusCmd.Flags().StringP("app", "a", "", "Fly.io app name")
	statusCmd.Flags().BoolP("verbose", "v", false, "Show detailed information")
	statusCmd.Flags().Int("logs", 50, "Number of recent log lines to show")

	// Init command flags
	initCmd.Flags().StringP("name", "n", "", "Project name")
	initCmd.Flags().StringP("template", "t", "web", "Project template (web, api, worker)")
	initCmd.Flags().Bool("force", false, "Overwrite existing files")

	// Dev command flags
	devCmd.Flags().Bool("watch", true, "Enable file watching")
	
	// Litestream command flags
	litestreamStartCmd.Flags().String("db", "./pb_data/data.db", "Database file path")
	litestreamStartCmd.Flags().String("backup", "./backups/data.db", "Backup file path")
	litestreamStartCmd.Flags().String("config", "", "Litestream config file")
	litestreamStartCmd.Flags().Bool("verbose", false, "Verbose output")

	litestreamRestoreCmd.Flags().String("db", "./pb_data/data.db", "Database file path to restore to")
	litestreamRestoreCmd.Flags().String("backup", "", "Backup source path")
	litestreamRestoreCmd.Flags().String("config", "", "Litestream config file")
	litestreamRestoreCmd.Flags().String("timestamp", "", "Restore to specific timestamp (ISO8601)")

	litestreamStatusCmd.Flags().String("config", "", "Litestream config file")

	// Add litestream subcommands
	litestreamCmd.AddCommand(litestreamStartCmd)
	litestreamCmd.AddCommand(litestreamRestoreCmd)
	litestreamCmd.AddCommand(litestreamStatusCmd)

	// Binary command flags
	binaryCmd.Flags().StringP("output", "o", "", "Output directory (default: .bin)")
	binaryCmd.Flags().StringP("name", "n", "infra", "Binary name")
	binaryCmd.Flags().BoolP("all", "a", false, "Build for all platforms and architectures")
	binaryCmd.Flags().Bool("local", false, "Build only for local platform/architecture")
	binaryCmd.Flags().StringSlice("platforms", []string{}, "Target platforms (linux, darwin, windows)")
	binaryCmd.Flags().StringSlice("arch", []string{}, "Target architectures (amd64, arm64)")
	binaryCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
}

// RunWorkflows adds all workflow commands to the root command.
func RunWorkflows() {
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(preCommitCmd)
	rootCmd.AddCommand(ciCmd)
	rootCmd.AddCommand(binaryCmd)
	rootCmd.AddCommand(devCmd)
	rootCmd.AddCommand(litestreamCmd)
	mcp.AddCommands(rootCmd)
	
	// Add Fly.io commands
	fly.AddCommands(rootCmd)
}

// runPreCommitChecks implements the pre-commit workflow logic
func runPreCommitChecks() error {
	fmt.Println("🔍 Running pre-commit checks...")
	
	// Check if we have staged Go files
	if hasError := checkStagedGoFiles(); hasError != nil {
		return hasError
	}
	
	// Run API compatibility check
	if err := runAPICompatibilityCheck("HEAD~1", "HEAD"); err != nil {
		return err
	}
	
	// Check documentation quality
	if err := checkDocumentationQuality(); err != nil {
		return err
	}
	
	fmt.Println("✅ All pre-commit checks passed")
	return nil
}

// runCIChecks implements the CI workflow logic
func runCIChecks() error {
	fmt.Println("🔍 Running CI checks...")
	
	// Run API compatibility check
	if err := runAPICompatibilityCheck("", ""); err != nil {
		return err
	}
	
	// Verify documentation quality for all packages
	if err := verifyAllDocumentationQuality(); err != nil {
		return err
	}
	
	fmt.Println("✅ All CI checks passed")
	return nil
}

// runDevMode starts development mode
func runDevMode(watch bool) error {
	fmt.Printf("🚀 Starting development mode (watch: %v)...\n", watch)
	
	if watch {
		fmt.Println("📁 File watching enabled - changes will trigger rebuilds")
	}
	
	// TODO: Implement file watching and hot reload
	return fmt.Errorf("development mode not fully implemented yet")
}

// runLitestreamStart starts Litestream replication
func runLitestreamStart(dbPath, backupPath, configPath string, verbose bool) error {
	fmt.Println("🔄 Starting Litestream replication...")
	
	// Default paths if not provided
	if dbPath == "" {
		dbPath = "./pb_data/data.db"
	}
	if backupPath == "" {
		backupPath = "./backups/data.db"
	}
	
	// Ensure directories exist
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return fmt.Errorf("failed to create db directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(backupPath), 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}
	
	// Create default config if not provided
	if configPath == "" {
		configPath = "./pkg/litestream/litestream.yml"
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			// Create minimal config for filesystem replication
			config := fmt.Sprintf(`
dbs:
  - path: %s
    replicas:
      - type: file
        path: %s
        sync-interval: 1s
        retention: 24h
`, dbPath, backupPath)
			
			if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
				return fmt.Errorf("failed to create config: %w", err)
			}
			fmt.Printf("📄 Created config: %s\n", configPath)
		}
	}
	
	fmt.Printf("📊 Database: %s\n", dbPath)
	fmt.Printf("💾 Backup: %s\n", backupPath)
	fmt.Printf("⚙️  Config: %s\n", configPath)
	
	// Execute litestream
	cmd := exec.Command("litestream", "replicate", "-config", configPath)
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	
	return cmd.Run()
}

// runLitestreamRestore restores database from backup
func runLitestreamRestore(dbPath, backupPath, configPath, timestamp string) error {
	fmt.Println("🔄 Restoring database from Litestream backup...")
	
	if dbPath == "" {
		dbPath = "./pb_data/data.db"
	}
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Build restore command
	cmdArgs := []string{"restore"}
	if configPath != "" {
		cmdArgs = append(cmdArgs, "-config", configPath)
	}
	if timestamp != "" {
		cmdArgs = append(cmdArgs, "-timestamp", timestamp)
	}
	
	// Add backup path as source
	if backupPath != "" {
		cmdArgs = append(cmdArgs, backupPath)
	} else {
		cmdArgs = append(cmdArgs, "-config", "./pkg/litestream/litestream.yml")
	}
	
	// Execute restore
	cmd := exec.Command("litestream", cmdArgs...)
	cmd.Dir = filepath.Dir(dbPath)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("restore failed: %w\nOutput: %s", err, string(output))
	}
	
	fmt.Printf("✅ Database restored to: %s\n", dbPath)
	return nil
}

// runLitestreamStatus shows replication status
func runLitestreamStatus(configPath string) error {
	fmt.Println("📊 Checking Litestream replication status...")
	
	if configPath == "" {
		configPath = "./pkg/litestream/litestream.yml"
	}
	
	// Check if litestream is running
	cmd := exec.Command("litestream", "dbs", "-config", configPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("status check failed: %w\nOutput: %s", err, string(output))
	}
	
	fmt.Printf("📋 Status:\n%s", string(output))
	return nil
}