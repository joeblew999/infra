package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/nats"
	"github.com/spf13/cobra"
)

// NewClusterCmd returns the cluster management commands.
func NewClusterCmd() *cobra.Command {
	var clusterCmd = &cobra.Command{
		Use:   "cluster",
		Short: "NATS cluster management commands",
		Long:  `Manage NATS clusters for local development and Fly.io production deployment`,
	}

	// Local cluster commands
	var localCmd = &cobra.Command{
		Use:   "local",
		Short: "Local NATS cluster management",
		Long:  `Manage the goreman-supervised local NATS cluster used in development`,
	}

	var localStartCmd = &cobra.Command{
		Use:   "start",
		Short: "Start local NATS cluster",
		Long:  `Start or ensure the local 6-node NATS cluster using goreman-managed processes with JetStream enabled`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			return nats.StartLocalCluster(ctx)
		},
	}

	var localStopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stop local NATS cluster",
		Long:  `Stop all local NATS cluster processes supervised by goreman`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nats.StopLocalCluster()
		},
	}

	var localStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show local NATS cluster status",
		Long:  `Display the status of all local NATS cluster nodes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showClusterStatus(true)
		},
	}

	var localUpgradeCmd = &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade local NATS cluster",
		Long:  `Placeholder for rolling upgrades of the goreman-managed local NATS cluster (not yet implemented)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			return nats.UpgradeCluster(ctx, true)
		},
	}

	// Production cluster commands
	var deployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploy NATS cluster to Fly.io",
		Long:  `Deploy a 6-node NATS cluster across multiple Fly.io regions (iad, lhr, nrt, syd, fra, sjc)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			return nats.DeployFlyCluster(ctx)
		},
	}

	var prodStatusCmd = &cobra.Command{
		Use:   "prod-status",
		Short: "Show production NATS cluster status",
		Long:  `Display the status of all production NATS cluster nodes on Fly.io`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return showClusterStatus(false)
		},
	}

	var prodUpgradeCmd = &cobra.Command{
		Use:   "prod-upgrade",
		Short: "Upgrade production NATS cluster with lame duck mode",
		Long:  `Perform rolling upgrade of production NATS cluster using lame duck mode for zero downtime`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			return nats.UpgradeCluster(ctx, false)
		},
	}

	// Global status command
	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show cluster status",
		Long:  `Display the status of NATS clusters (local in development, production on Fly.io)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Show appropriate cluster based on environment
			return showClusterStatus(config.IsDevelopment())
		},
	}

	// Bootstrap command for initial setup
	var bootstrapCmd = &cobra.Command{
		Use:   "bootstrap",
		Short: "Bootstrap NATS cluster infrastructure",
		Long: `Bootstrap complete NATS cluster infrastructure:
- Local: Ensure goreman-managed 6-node cluster for development
- Production: Deploy 6-node cluster across Fly.io regions

This command is idempotent and safe to run multiple times.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if config.IsProduction() {
				fmt.Println("🚀 Bootstrapping production NATS cluster on Fly.io...")
				return nats.DeployFlyCluster(ctx)
			} else {
				fmt.Println("🐳 Bootstrapping local NATS cluster with goreman...")
				return nats.StartLocalCluster(ctx)
			}
		},
	}

	// Add subcommands to local
	localCmd.AddCommand(localStartCmd)
	localCmd.AddCommand(localStopCmd)
	localCmd.AddCommand(localStatusCmd)
	localCmd.AddCommand(localUpgradeCmd)

	// Add all commands to cluster
	clusterCmd.AddCommand(localCmd)
	clusterCmd.AddCommand(deployCmd)
	clusterCmd.AddCommand(prodStatusCmd)
	clusterCmd.AddCommand(prodUpgradeCmd)
	clusterCmd.AddCommand(statusCmd)
	clusterCmd.AddCommand(bootstrapCmd)

	return clusterCmd
}

// showClusterStatus displays cluster status in a user-friendly format
func showClusterStatus(isLocal bool) error {
	clusterConfig, err := nats.GetClusterStatus(isLocal)
	if err != nil {
		return fmt.Errorf("failed to get cluster status: %w", err)
	}

	// Display cluster overview
	environment := "Production (Fly.io)"
	if isLocal {
		environment = "Local (goreman)"
	}

	fmt.Printf("📊 NATS Cluster Status - %s\n", environment)
	fmt.Printf("Cluster Name: %s\n", clusterConfig.ClusterName)
	fmt.Printf("Environment: %s\n", clusterConfig.Environment)
	fmt.Printf("Nodes: %d\n", len(clusterConfig.Nodes))
	fmt.Printf("JetStream: %v\n", clusterConfig.EnableJetStream)
	fmt.Printf("Web GUI: %v\n\n", clusterConfig.EnableWebGUI)

	// Display node details
	fmt.Println("Node Details:")
	fmt.Printf("%-12s %-8s %-12s %-14s %-12s %-10s\n", "NAME", "REGION", "CLIENT_PORT", "CLUSTER_PORT", "HTTP_PORT", "STATUS")
	fmt.Println(strings.Repeat("-", 80))

	for _, node := range clusterConfig.Nodes {
		status := node.Status
		statusEmoji := "❓"
		switch status {
		case "running":
			statusEmoji = "✅"
		case "stopped":
			statusEmoji = "⛔"
		case "error":
			statusEmoji = "❌"
		}

		fmt.Printf("%-12s %-8s %-12d %-14d %-12d %s %s\n",
			node.Name, node.Region, node.Port, node.ClusterPort, node.HTTPPort, statusEmoji, status)
	}

	// Display web GUI URLs for local cluster
	if isLocal {
		fmt.Println("\n🌐 Web GUI URLs (HTTP Monitoring):")
		for _, node := range clusterConfig.Nodes {
			if node.Status == "running" {
				fmt.Printf("  %s: %s/\n", node.Name, config.FormatLocalHTTP(fmt.Sprintf("%d", node.HTTPPort)))
			}
		}
	}

	// JSON output option can be added later if needed
	// For now, basic status display is sufficient

	return nil
}
