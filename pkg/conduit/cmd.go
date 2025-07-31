package conduit

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

// Cmd represents the conduit command
var Cmd = &cobra.Command{
	Use:   "conduit",
	Short: "Manage Conduit and its connectors",
	Long: `Manage Conduit (https://github.com/ConduitIO/conduit) and its connectors.

This command provides lifecycle management for Conduit and its connectors,
including binary management, process supervision, and health monitoring.`,
}

// conduitStartCmd starts Conduit and connectors
var conduitStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Conduit and connectors",
	Long:  `Start Conduit and all configured connectors as managed processes.`,
	Run: func(cmd *cobra.Command, args []string) {
		service := NewService()
		
		fmt.Println("Starting Conduit service...")
		if err := service.EnsureAndStart(false); err != nil {
			log.Fatal("Failed to start Conduit service:", err)
		}
		
		fmt.Println("✅ Conduit service started successfully")
		
		// Keep running (like a daemon)
		select {}
	},
}

// conduitStopCmd stops Conduit and connectors
var conduitStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop Conduit and connectors",
	Long:  `Stop all running Conduit processes gracefully.`,
	Run: func(cmd *cobra.Command, args []string) {
		service := NewService()
		
		fmt.Println("Stopping Conduit service...")
		if err := service.Initialize(); err != nil {
			log.Fatal("Failed to initialize service:", err)
		}
		
		if err := service.Stop(); err != nil {
			log.Fatal("Failed to stop Conduit service:", err)
		}
		
		fmt.Println("✅ Conduit service stopped successfully")
	},
}

// conduitStatusCmd shows process status
var conduitStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show Conduit process status",
	Long:  `Display the current status of all Conduit processes.`,
	Run: func(cmd *cobra.Command, args []string) {
		service := NewService()
		
		if err := service.Initialize(); err != nil {
			log.Fatal("Failed to initialize service:", err)
		}
		
		status := service.Status()
		if len(status) == 0 {
			fmt.Println("No Conduit processes configured")
			return
		}
		
		fmt.Println("Conduit Process Status:")
		fmt.Println("======================")
		for name, state := range status {
			fmt.Printf("%-30s: %s\n", name, state)
		}
	},
}

// conduitRestartCmd restarts Conduit and connectors
var conduitRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart Conduit and connectors",
	Long:  `Restart all Conduit processes gracefully.`,
	Run: func(cmd *cobra.Command, args []string) {
		service := NewService()
		
		fmt.Println("Restarting Conduit service...")
		if err := service.Initialize(); err != nil {
			log.Fatal("Failed to initialize service:", err)
		}
		
		if err := service.Restart(); err != nil {
			log.Fatal("Failed to restart Conduit service:", err)
		}
		
		fmt.Println("✅ Conduit service restarted successfully")
	},
}

// conduitCoreCmd manages the core conduit process
var conduitCoreCmd = &cobra.Command{
	Use:   "core",
	Short: "Manage the core conduit process",
	Long:  `Manage the core conduit process independently.`,
}

// conduitCoreStartCmd starts only the core conduit
var conduitCoreStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start only the core conduit process",
	Long:  `Start only the core conduit process, without connectors.`,
	Run: func(cmd *cobra.Command, args []string) {
		service := NewService()
		
		if err := service.Initialize(); err != nil {
			log.Fatal("Failed to initialize service:", err)
		}
		
		fmt.Println("Starting core conduit...")
		if err := service.StartCore(); err != nil {
			log.Fatal("Failed to start core conduit:", err)
		}
		
		fmt.Println("✅ Core conduit started successfully")
	},
}

// conduitCoreStopCmd stops only the core conduit
var conduitCoreStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop only the core conduit process",
	Long:  `Stop only the core conduit process, leaving connectors running.`,
	Run: func(cmd *cobra.Command, args []string) {
		service := NewService()
		
		if err := service.Initialize(); err != nil {
			log.Fatal("Failed to initialize service:", err)
		}
		
		fmt.Println("Stopping core conduit...")
		if err := service.StopCore(); err != nil {
			log.Fatal("Failed to stop core conduit:", err)
		}
		
		fmt.Println("✅ Core conduit stopped successfully")
	},
}

// conduitConnectorsCmd manages connectors
var conduitConnectorsCmd = &cobra.Command{
	Use:   "connectors",
	Short: "Manage connector processes",
	Long:  `Manage connector processes independently.`,
}

// conduitConnectorsStartCmd starts all connectors
var conduitConnectorsStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start all connector processes",
	Long:  `Start all configured connector processes.`,
	Run: func(cmd *cobra.Command, args []string) {
		service := NewService()
		
		if err := service.Initialize(); err != nil {
			log.Fatal("Failed to initialize service:", err)
		}
		
		fmt.Println("Starting connectors...")
		if err := service.StartConnectors(); err != nil {
			log.Fatal("Failed to start connectors:", err)
		}
		
		fmt.Println("✅ Connectors started successfully")
	},
}

// conduitConnectorsStopCmd stops all connectors
var conduitConnectorsStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop all connector processes",
	Long:  `Stop all running connector processes.`,
	Run: func(cmd *cobra.Command, args []string) {
		service := NewService()
		
		if err := service.Initialize(); err != nil {
			log.Fatal("Failed to initialize service:", err)
		}
		
		fmt.Println("Stopping connectors...")
		if err := service.StopConnectors(); err != nil {
			log.Fatal("Failed to stop connectors:", err)
		}
		
		fmt.Println("✅ Connectors stopped successfully")
	},
}

func init() {
	// Add debug flag to all commands
	conduitStartCmd.Flags().Bool("debug", false, "Enable debug logging")
	conduitRestartCmd.Flags().Bool("debug", false, "Enable debug logging")
	
	// Add subcommands to root
	Cmd.AddCommand(conduitStartCmd)
	Cmd.AddCommand(conduitStopCmd)
	Cmd.AddCommand(conduitStatusCmd)
	Cmd.AddCommand(conduitRestartCmd)
	
	// Add core subcommands
	conduitCoreCmd.AddCommand(conduitCoreStartCmd)
	conduitCoreCmd.AddCommand(conduitCoreStopCmd)
	Cmd.AddCommand(conduitCoreCmd)
	
	// Add connectors subcommands
	conduitConnectorsCmd.AddCommand(conduitConnectorsStartCmd)
	conduitConnectorsCmd.AddCommand(conduitConnectorsStopCmd)
	Cmd.AddCommand(conduitConnectorsCmd)
}

// RunConduit adds conduit commands to the CLI
func RunConduit() {
	// Commands are added via init()
}