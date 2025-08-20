package litestream

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/joeblew999/infra/pkg/config"
)

// AddCommands adds all litestream commands to the root command
func AddCommands(rootCmd *cobra.Command) {
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

			return RunLitestreamStart(dbPath, backupPath, config, verbose)
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

			return RunLitestreamRestore(dbPath, backupPath, config, timestamp)
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
			return RunLitestreamStatus(config)
		},
	}

	// Add flags
	litestreamStartCmd.Flags().String("db", "./pb_data/data.db", "Database file path")
	litestreamStartCmd.Flags().String("backup", "./backups/data.db", "Backup file path")
	litestreamStartCmd.Flags().String("config", "", "Litestream config file")
	litestreamStartCmd.Flags().Bool("verbose", false, "Verbose output")

	litestreamRestoreCmd.Flags().String("db", "./pb_data/data.db", "Database file path to restore to")
	litestreamRestoreCmd.Flags().String("backup", "", "Backup source path")
	litestreamRestoreCmd.Flags().String("config", "", "Litestream config file")
	litestreamRestoreCmd.Flags().String("timestamp", "", "Restore to specific timestamp (ISO8601)")

	litestreamStatusCmd.Flags().String("config", "", "Litestream config file")

	// Add subcommands
	litestreamCmd.AddCommand(litestreamStartCmd)
	litestreamCmd.AddCommand(litestreamRestoreCmd)
	litestreamCmd.AddCommand(litestreamStatusCmd)

	// Add the main litestream command to root
	rootCmd.AddCommand(litestreamCmd)
}

// RunLitestreamStart starts Litestream replication
func RunLitestreamStart(dbPath, backupPath, configPath string, verbose bool) error {
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
	cmd := exec.Command(getLitestreamBinary(), "replicate", "-config", configPath)
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	
	return cmd.Run()
}

// RunLitestreamRestore restores database from backup
func RunLitestreamRestore(dbPath, backupPath, configPath, timestamp string) error {
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
	cmd := exec.Command(getLitestreamBinary(), cmdArgs...)
	cmd.Dir = filepath.Dir(dbPath)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("restore failed: %w\nOutput: %s", err, string(output))
	}
	
	fmt.Printf("✅ Database restored to: %s\n", dbPath)
	return nil
}

// RunLitestreamStatus shows replication status
func RunLitestreamStatus(configPath string) error {
	fmt.Println("📊 Checking Litestream replication status...")
	
	if configPath == "" {
		configPath = "./pkg/litestream/litestream.yml"
	}
	
	// Check if litestream is running
	cmd := exec.Command(getLitestreamBinary(), "dbs", "-config", configPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("status check failed: %w\nOutput: %s", err, string(output))
	}
	
	fmt.Printf("📋 Status:\n%s", string(output))
	return nil
}

// getLitestreamBinary returns the path to the litestream binary using type-safe constants
func getLitestreamBinary() string {
	return config.Get(config.BinaryLitestream)
}