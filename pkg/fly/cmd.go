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
		Long:  `Scale the Fly.io application resources`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Scale()
		},
	}

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
func Scale() error {
	fmt.Println("⚖️  Scaling Fly.io resources...")
	
	cmd := exec.Command(config.GetFlyctlBinPath(), "scale", "show")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}