package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func newDeployCommand() *cobra.Command {
	var (
		environment string
		appName     string
		region      string
		dryRun      bool
	)

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy core V2 to Fly.io (monolithic)",
		Long: strings.TrimSpace(`
Deploy the entire core stack as a single container to Fly.io.

This command:
1. Builds the core binary with ko
2. Pushes the image to Fly.io registry
3. Deploys to Fly.io with fly.toml configuration

The monolithic deployment includes:
- Process-compose orchestrator
- NATS JetStream
- PocketBase
- Caddy reverse proxy

All services run in a single container with a persistent volume for data.

For microservices deployment (Phase 2), see docs/DEPLOYMENT.md
		`),
		Example: strings.TrimSpace(`
  # Deploy to production
  FLY_API_TOKEN=... go run . deploy

  # Deploy to specific app/region
  go run . deploy --app core-v2-dev --region syd

  # Dry run to see what would happen
  go run . deploy --dry-run
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeploy(cmd, environment, appName, region, dryRun)
		},
	}

	cmd.Flags().StringVarP(&environment, "env", "e", "production", "Deployment environment (production|development|staging)")
	cmd.Flags().StringVarP(&appName, "app", "a", "core-v2", "Fly.io app name")
	cmd.Flags().StringVarP(&region, "region", "r", "syd", "Primary deployment region")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be deployed without actually deploying")

	return cmd
}

func runDeploy(cmd *cobra.Command, environment, appName, region string, dryRun bool) error {
	out := cmd.OutOrStdout()

	fmt.Fprintf(out, "🚀 Deploying core V2 (monolithic) to Fly.io\n")
	fmt.Fprintf(out, "   App: %s\n", appName)
	fmt.Fprintf(out, "   Region: %s\n", region)
	fmt.Fprintf(out, "   Environment: %s\n", environment)
	fmt.Fprintln(out)

	if dryRun {
		fmt.Fprintf(out, "🔍 DRY RUN MODE - No actual deployment will occur\n")
		fmt.Fprintln(out)
	}

	// Check for FLY_API_TOKEN (only required for actual deployment)
	flyToken := os.Getenv("FLY_API_TOKEN")
	if flyToken == "" && !dryRun {
		return fmt.Errorf("FLY_API_TOKEN environment variable is required for deployment")
	}

	if dryRun {
		fmt.Fprintf(out, "🎯 Would execute these steps:\n")
		fmt.Fprintln(out)
		fmt.Fprintf(out, "   1. Check ko build tool\n")
		fmt.Fprintf(out, "   2. Check flyctl\n")
		fmt.Fprintf(out, "   3. Build container: ko build --push=true --bare --platform=linux/amd64,linux/arm64 ./cmd/core\n")
		fmt.Fprintf(out, "      Registry: registry.fly.io/%s\n", appName)
		fmt.Fprintf(out, "   4. Deploy: flyctl deploy --app %s --image registry.fly.io/%s:latest\n", appName, appName)
		fmt.Fprintln(out)
		fmt.Fprintf(out, "✅ Dry run complete\n")
		return nil
	}

	// Step 1: Check if ko is available (check .dep/ first, then PATH)
	fmt.Fprintf(out, "📦 Step 1: Checking ko build tool...\n")
	koPath := ".dep/ko"
	if _, err := os.Stat(koPath); err != nil {
		// Fall back to PATH
		koPath, err = exec.LookPath("ko")
		if err != nil {
			return fmt.Errorf("ko build tool not found in .dep/ or PATH - run: go run . ensure ko")
		}
	}
	fmt.Fprintf(out, "   ✅ Found ko at: %s\n", koPath)
	fmt.Fprintln(out)

	// Step 2: Check if flyctl is available (check .dep/ first, then PATH)
	fmt.Fprintf(out, "✈️  Step 2: Checking flyctl...\n")
	flyctlPath := ".dep/flyctl"
	if _, err := os.Stat(flyctlPath); err != nil {
		// Fall back to PATH
		flyctlPath, err = exec.LookPath("flyctl")
		if err != nil {
			return fmt.Errorf("flyctl not found in .dep/ or PATH - run: go run . ensure flyctl")
		}
	}
	fmt.Fprintf(out, "   ✅ Found flyctl at: %s\n", flyctlPath)
	fmt.Fprintln(out)

	// Step 3: Build and push with ko
	fmt.Fprintf(out, "🔨 Step 3: Building container with ko...\n")

	registry := fmt.Sprintf("registry.fly.io/%s", appName)
	koCmd := exec.Command(koPath, "build",
		"--push=true",
		"--bare",
		"--platform=linux/amd64,linux/arm64",
		"./cmd/core",
	)
	koCmd.Env = append(os.Environ(),
		fmt.Sprintf("KO_DOCKER_REPO=%s", registry),
		fmt.Sprintf("FLY_API_TOKEN=%s", flyToken),
	)
	koCmd.Stdout = out
	koCmd.Stderr = os.Stderr

	if err := koCmd.Run(); err != nil {
		return fmt.Errorf("ko build failed: %w", err)
	}
	fmt.Fprintf(out, "   ✅ Image built and pushed to %s\n", registry)
	fmt.Fprintln(out)

	// Step 4: Deploy to Fly.io
	fmt.Fprintf(out, "🚀 Step 4: Deploying to Fly.io...\n")

	deployCmd := exec.Command(flyctlPath, "deploy",
		"--app", appName,
		"--image", fmt.Sprintf("%s:latest", registry),
		"--config", "fly.toml",
	)
	deployCmd.Env = append(os.Environ(),
		fmt.Sprintf("FLY_API_TOKEN=%s", flyToken),
	)
	deployCmd.Stdout = out
	deployCmd.Stderr = os.Stderr

	if err := deployCmd.Run(); err != nil {
		return fmt.Errorf("fly deploy failed: %w", err)
	}
	fmt.Fprintln(out)

	// Step 5: Verify deployment
	fmt.Fprintf(out, "✅ Deployment complete!\n")
	fmt.Fprintln(out)
	fmt.Fprintf(out, "📋 Next steps:\n")
	fmt.Fprintf(out, "   • Check status: flyctl status --app %s\n", appName)
	fmt.Fprintf(out, "   • View logs: flyctl logs --app %s\n", appName)
	fmt.Fprintf(out, "   • Access app: https://%s.fly.dev\n", appName)
	fmt.Fprintln(out)

	return nil
}
