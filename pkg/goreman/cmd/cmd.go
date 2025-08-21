package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/joeblew999/infra/pkg/goreman"
	"github.com/joeblew999/infra/pkg/log"
)

// PsCmd shows running processes
var PsCmd = &cobra.Command{
	Use:     "ps",
	Aliases: []string{"list", "status"},
	Short:   "Show running processes",
	Long:    `Display the status of all supervised processes`,
	Run: func(cmd *cobra.Command, args []string) {
		watch, _ := cmd.Flags().GetBool("watch")
		if watch {
			watchProcesses()
		} else {
			showProcesses()
		}
	},
}

// StartCmd starts a process
var StartCmd = &cobra.Command{
	Use:   "start <process-name>",
	Short: "Start a process",
	Long:  `Start a specific supervised process`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		processName := args[0]
		if err := goreman.Start(processName); err != nil {
			log.Error("Failed to start process", "name", processName, "error", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Started process: %s\n", processName)
	},
}

// StopCmd stops a process
var StopCmd = &cobra.Command{
	Use:   "stop <process-name>",
	Short: "Stop a process",
	Long:  `Stop a specific supervised process`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		processName := args[0]
		if err := goreman.Stop(processName); err != nil {
			log.Error("Failed to stop process", "name", processName, "error", err)
			os.Exit(1)
		}
		fmt.Printf("🛑 Stopped process: %s\n", processName)
	},
}

// RestartCmd restarts a process
var RestartCmd = &cobra.Command{
	Use:   "restart <process-name>",
	Short: "Restart a process",
	Long:  `Restart a specific supervised process`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		processName := args[0]
		if err := goreman.Restart(processName); err != nil {
			log.Error("Failed to restart process", "name", processName, "error", err)
			os.Exit(1)
		}
		fmt.Printf("🔄 Restarted process: %s\n", processName)
	},
}

// RegisterCmd registers and starts a service
var RegisterCmd = &cobra.Command{
	Use:   "register <service-name>",
	Short: "Register and start a service for supervision",
	Long:  `Register and start services that have been registered with the service factory`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceName := args[0]
		if err := goreman.StartService(serviceName); err != nil {
			log.Error("Failed to start service", "service", serviceName, "error", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Started service: %s\n", serviceName)
	},
}

// ServicesCmd lists available services
var ServicesCmd = &cobra.Command{
	Use:     "services",
	Aliases: []string{"list-services"},
	Short:   "List available services",
	Long:    `List services that can be registered and started`,
	Run: func(cmd *cobra.Command, args []string) {
		services := goreman.GetAvailableServices()
		if len(services) == 0 {
			fmt.Println("📭 No services registered")
			fmt.Println("💡 Services register themselves when their packages are imported")
			return
		}
		
		fmt.Println("🔧 Available Services:")
		for _, service := range services {
			fmt.Printf("   • %s\n", service)
		}
		fmt.Printf("\nUse 'infra goreman register <service-name>' to start a service\n")
	},
}

func showProcesses() {
	status := goreman.GetAllStatus()
	
	if len(status) == 0 {
		fmt.Println("📭 No supervised processes found")
		fmt.Println("💡 Use 'infra supervised' to start the demo or register processes")
		return
	}

	fmt.Println("🔍 Supervised Process Status")
	fmt.Println()

	// Create table writer
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	
	// Header
	fmt.Fprintln(w, "NAME\tSTATUS\tINDICATOR")
	fmt.Fprintln(w, "----\t------\t---------")

	// Process rows
	for name, stat := range status {
		indicator := getStatusIndicator(stat)
		fmt.Fprintf(w, "%s\t%s\t%s\n", name, strings.ToUpper(stat), indicator)
	}
	
	w.Flush()
	fmt.Println()
	
	// Summary
	running := 0
	stopped := 0
	for _, stat := range status {
		if stat == "running" {
			running++
		} else {
			stopped++
		}
	}
	
	fmt.Printf("📊 Summary: %d total, %d running, %d stopped\n", len(status), running, stopped)
}

func watchProcesses() {
	fmt.Println("👀 Watching processes (Press Ctrl+C to exit)")
	fmt.Println()
	
	for {
		// Clear screen
		fmt.Print("\033[2J\033[H")
		
		// Show timestamp
		fmt.Printf("🕒 Last updated: %s\n\n", time.Now().Format("15:04:05"))
		
		// Show processes
		showProcesses()
		
		// Wait before refresh
		time.Sleep(2 * time.Second)
	}
}

func getStatusIndicator(status string) string {
	switch status {
	case "running":
		return "🟢"
	case "stopped":
		return "🔴"
	case "starting":
		return "🟡"
	case "stopping":
		return "🟠"
	case "killed":
		return "💀"
	default:
		return "❓"
	}
}

func init() {
	// Add flags
	PsCmd.Flags().BoolP("watch", "w", false, "Watch processes continuously")
}