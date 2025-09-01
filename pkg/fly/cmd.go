package fly

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/joeblew999/infra/pkg/config"
)

// AddCommands adds all Fly.io commands to the root command
func AddCommands(rootCmd *cobra.Command) {
	var flyCmd = &cobra.Command{
		Use:   "fly",
		Short: "Fly.io deployment commands",
		Long:  `Manage Fly.io deployments and infrastructure`,
	}

	var deployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploy to Fly.io",
		Long:  `Deploy the application to Fly.io using the current configuration`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Deploy()
		},
	}

	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show Fly.io app status",
		Long:  `Show the current status of the Fly.io application`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Status()
		},
	}

	var logsCmd = &cobra.Command{
		Use:   "logs",
		Short: "Show Fly.io app logs",
		Long:  `Show logs from the Fly.io application`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Logs()
		},
	}

	var sshCmd = &cobra.Command{
		Use:   "ssh",
		Short: "SSH into Fly.io machine",
		Long:  `SSH into the Fly.io machine for debugging`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return SSH()
		},
	}

	var scaleCmd = &cobra.Command{
		Use:   "scale",
		Short: "Scale Fly.io resources",
		Long:  `Scale the Fly.io application resources:
  - Scale machine count: --count 2
  - Scale memory: --memory 1024
  - Scale CPU: --cpu 2
  - Scale to specific machine type: --vm shared-cpu-2x
  
If no flags are provided, shows current scaling configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			count, _ := cmd.Flags().GetInt("count")
			memory, _ := cmd.Flags().GetInt("memory")
			cpu, _ := cmd.Flags().GetInt("cpu")
			vm, _ := cmd.Flags().GetString("vm")
			app, _ := cmd.Flags().GetString("app")
			
			return Scale(count, memory, cpu, vm, app)
		},
	}
	
	// Add flags for scaling options
	scaleCmd.Flags().Int("count", 0, "Number of machines to scale to")
	scaleCmd.Flags().Int("memory", 0, "Memory in MB (e.g., 512, 1024, 2048)")
	scaleCmd.Flags().Int("cpu", 0, "Number of CPUs")
	scaleCmd.Flags().String("vm", "", "VM type (e.g., shared-cpu-2x, performance-2x)")
	scaleCmd.Flags().StringP("app", "a", "", "Fly.io app name (default: from env or fly.toml)")

	flyCmd.AddCommand(deployCmd)
	flyCmd.AddCommand(statusCmd)
	flyCmd.AddCommand(logsCmd)
	flyCmd.AddCommand(sshCmd)
	flyCmd.AddCommand(scaleCmd)

	rootCmd.AddCommand(flyCmd)
}

// Deploy deploys the application to Fly.io
func Deploy() error {
	fmt.Println("🚀 Deploying to Fly.io...")
	
	cmd := exec.Command(config.GetFlyctlBinPath(), "deploy")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

// Status shows Fly.io app status
func Status() error {
	fmt.Println("📊 Checking Fly.io status...")
	
	cmd := exec.Command(config.GetFlyctlBinPath(), "status")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

// Logs shows Fly.io app logs
func Logs() error {
	fmt.Println("📋 Showing Fly.io logs...")
	
	cmd := exec.Command(config.GetFlyctlBinPath(), "logs")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

// SSH connects to Fly.io machine
func SSH() error {
	fmt.Println("🔧 Connecting to Fly.io machine...")
	
	cmd := exec.Command(config.GetFlyctlBinPath(), "ssh", "console")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

// Scale scales Fly.io resources
func Scale(count, memory, cpu int, vm, app string) error {
	// If no scaling flags provided, just show current configuration
	if count == 0 && memory == 0 && cpu == 0 && vm == "" {
		fmt.Println("⚖️  Showing current Fly.io scaling configuration...")
		cmd := exec.Command(config.GetFlyctlBinPath(), "scale", "show")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	
	fmt.Println("⚖️  Scaling Fly.io resources...")
	
	// Build scaling commands based on provided flags
	var commands []*exec.Cmd
	
	// Scale machine count
	if count > 0 {
		fmt.Printf("📊 Scaling machine count to %d...\n", count)
		args := []string{"scale", "count", fmt.Sprintf("%d", count), "--yes"}
		if app != "" {
			args = append(args, "-a", app)
		}
		commands = append(commands, exec.Command(config.GetFlyctlBinPath(), args...))
	}
	
	// Scale memory
	if memory > 0 {
		fmt.Printf("💾 Scaling memory to %dMB...\n", memory)
		args := []string{"scale", "memory", fmt.Sprintf("%d", memory), "--yes"}
		if app != "" {
			args = append(args, "-a", app)
		}
		commands = append(commands, exec.Command(config.GetFlyctlBinPath(), args...))
	}
	
	// Scale CPU
	if cpu > 0 {
		fmt.Printf("🔧 Scaling CPU to %d cores...\n", cpu)
		args := []string{"scale", "cpu", fmt.Sprintf("%d", cpu), "--yes"}
		if app != "" {
			args = append(args, "-a", app)
		}
		commands = append(commands, exec.Command(config.GetFlyctlBinPath(), args...))
	}
	
	// Scale VM type
	if vm != "" {
		fmt.Printf("🖥️  Scaling to VM type: %s...\n", vm)
		args := []string{"scale", "vm", vm, "--yes"}
		if app != "" {
			args = append(args, "-a", app)
		}
		commands = append(commands, exec.Command(config.GetFlyctlBinPath(), args...))
	}
	
	// Execute all scaling commands
	for i, cmd := range commands {
		if i > 0 {
			fmt.Println() // Add spacing between commands
		}
		
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("scaling command failed: %w", err)
		}
	}
	
	fmt.Println("\n✅ Scaling operations completed successfully!")
	return nil
}