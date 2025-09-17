package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/goreman"
	"github.com/joeblew999/infra/pkg/log"
	"github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
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
			watchProcesses(cmd)
		} else {
			showProcesses(cmd)
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
		ctx := cmd.Context()
		resp, err := sendProcessCommand(ctx, "start", processName)
		if err != nil {
			log.Warn("Remote start failed, falling back to local manager", "error", err)
			if err := goreman.Start(processName); err != nil {
				log.Error("Failed to start process", "name", processName, "error", err)
				os.Exit(1)
			}
			fmt.Printf("✅ Started process: %s\n", processName)
			return
		}
		if !resp.Success {
			log.Error("Failed to start process", "name", processName, "error", resp.Message)
			os.Exit(1)
		}
		fmt.Println("✅", resp.Message)
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
		ctx := cmd.Context()
		resp, err := sendProcessCommand(ctx, "stop", processName)
		if err != nil {
			log.Warn("Remote stop failed, falling back to local manager", "error", err)
			if err := goreman.Stop(processName); err != nil {
				log.Error("Failed to stop process", "name", processName, "error", err)
				os.Exit(1)
			}
			fmt.Printf("🛑 Stopped process: %s\n", processName)
			return
		}
		if !resp.Success {
			log.Error("Failed to stop process", "name", processName, "error", resp.Message)
			os.Exit(1)
		}
		fmt.Println("🛑", resp.Message)
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
		ctx := cmd.Context()
		resp, err := sendProcessCommand(ctx, "restart", processName)
		if err != nil {
			log.Warn("Remote restart failed, falling back to local manager", "error", err)
			if err := goreman.Restart(processName); err != nil {
				log.Error("Failed to restart process", "name", processName, "error", err)
				os.Exit(1)
			}
			fmt.Printf("🔄 Restarted process: %s\n", processName)
			return
		}
		if !resp.Success {
			log.Error("Failed to restart process", "name", processName, "error", resp.Message)
			os.Exit(1)
		}
		fmt.Println("🔄", resp.Message)
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

func showProcesses(cmd *cobra.Command) {
	ctx := cmd.Context()
	status, err := fetchRemoteStatuses(ctx)
	if err != nil {
		log.Warn("Failed to fetch remote process status", "error", err)
		status = goreman.GetAllStatus()
	}

	renderProcessTable(status)
}

func watchProcesses(cmd *cobra.Command) {
	fmt.Println("👀 Watching processes (Press Ctrl+C to exit)")
	fmt.Println()

	for {
		// Clear screen
		fmt.Print("\033[2J\033[H")

		// Show timestamp
		fmt.Printf("🕒 Last updated: %s\n\n", time.Now().Format("15:04:05"))

		// Show processes
		status, err := fetchRemoteStatuses(cmd.Context())
		if err != nil {
			log.Warn("Failed to fetch remote process status", "error", err)
			status = goreman.GetAllStatus()
		}
		renderProcessTable(status)

		// Wait before refresh
		time.Sleep(2 * time.Second)
	}
}

func renderProcessTable(status map[string]string) {
	if len(status) == 0 {
		fmt.Println("📭 No supervised processes found")
		fmt.Println("💡 Use 'infra supervised' to start the demo or register processes")
		return
	}

	fmt.Println("🔍 Supervised Process Status")
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tSTATUS\tINDICATOR")
	fmt.Fprintln(w, "----\t------\t---------")

	for name, stat := range status {
		indicator := getStatusIndicator(stat)
		fmt.Fprintf(w, "%s\t%s\t%s\n", name, strings.ToUpper(stat), indicator)
	}

	w.Flush()
	fmt.Println()

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

func fetchRemoteStatuses(ctx context.Context) (map[string]string, error) {
	resp, err := sendProcessCommand(ctx, "status", "")
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Message)
	}
	if resp.Statuses == nil {
		return map[string]string{}, nil
	}
	return resp.Statuses, nil
}

func sendProcessCommand(ctx context.Context, action, name string) (*goreman.ControlResponse, error) {
	nc, err := connectNATS()
	if err != nil {
		return nil, err
	}
	defer nc.Drain()

	cmd := goreman.ControlCommand{Action: action, Name: name}
	return goreman.ExecuteCommand(ctx, nc, cmd)
}

func connectNATS() (*nats.Conn, error) {
	addr := fmt.Sprintf("nats://127.0.0.1:%s", config.GetNATSPort())
	return nats.Connect(addr)
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

// GetGoremanCmd returns the main goreman command for CLI integration
func GetGoremanCmd() *cobra.Command {
	goremanCmd := &cobra.Command{
		Use:   "goreman",
		Short: "Process management and monitoring",
		Long:  `Monitor and manage supervised processes via goreman`,
	}

	// Add subcommands
	goremanCmd.AddCommand(PsCmd)
	goremanCmd.AddCommand(StartCmd)
	goremanCmd.AddCommand(StopCmd)
	goremanCmd.AddCommand(RestartCmd)
	goremanCmd.AddCommand(RegisterCmd)
	goremanCmd.AddCommand(ServicesCmd)

	return goremanCmd
}

func init() {
	// Add flags
	PsCmd.Flags().BoolP("watch", "w", false, "Watch processes continuously")
}
