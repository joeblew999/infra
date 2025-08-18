package build

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/joeblew999/infra/pkg/deck"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy deck tools release to GitHub",
	Long:  "Create and deploy a new release of deck tools to GitHub",
	Run: func(cmd *cobra.Command, args []string) {
		version, _ := cmd.Flags().GetString("version")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		
		if err := deployRelease(version, dryRun); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	deployCmd.Flags().StringP("version", "v", "", "Release version (e.g., v1.0.0)")
	deployCmd.Flags().BoolP("dry-run", "n", false, "Show what would be deployed without actually deploying")
	deployCmd.MarkFlagRequired("version")
	BuildCmd.AddCommand(deployCmd)
}

func deployRelease(version string, dryRun bool) error {
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	
	if !strings.HasPrefix(version, deck.ReleaseTagPrefix) {
		version = deck.ReleaseTagPrefix + version
	}
	
	fmt.Printf("Deploying deck tools release: %s\n", version)
	
	if dryRun {
		fmt.Println("DRY RUN - No actual deployment will occur")
		fmt.Printf("Would create GitHub release: %s/%s@%s\n", deck.GitHubOwner, deck.GitHubRepo, version)
		return nil
	}
	
	// Build all tools first
	manager := deck.NewManager()
	if err := manager.Install(); err != nil {
		return fmt.Errorf("failed to build tools: %w", err)
	}
	
	// Create release package
	packager := deck.NewPackager()
	releaseFile, err := packager.CreateReleasePackage(version)
	if err != nil {
		return fmt.Errorf("failed to create release package: %w", err)
	}
	
	// Use gh CLI to create release
	cmd := exec.Command("gh", "release", "create", version, releaseFile, 
		"--title", fmt.Sprintf("Deck Tools %s", strings.TrimPrefix(version, deck.ReleaseTagPrefix)),
		"--notes", fmt.Sprintf("Release %s of deck tools", strings.TrimPrefix(version, deck.ReleaseTagPrefix)))
	
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create GitHub release: %w", err)
	}
	
	fmt.Printf("Successfully deployed release: %s\n", version)
	return nil
}