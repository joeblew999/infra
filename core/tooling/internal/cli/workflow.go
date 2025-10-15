package cli

import (
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/joeblew999/infra/core/tooling/pkg/app"
	"github.com/joeblew999/infra/core/tooling/pkg/orchestrator"
	types "github.com/joeblew999/infra/core/tooling/pkg/types"
)

func newWorkflowCommand(profileFlag *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow",
		Short: "Run deployment workflows",
	}
	cmd.AddCommand(newWorkflowDeployCommand(profileFlag))
	return cmd
}

type deployOptions struct {
	profileFlag *string
	appFlag     string
	orgFlag     string
	repoFlag    string
	regionFlag  string
	verbose     bool
	noBrowser   bool
	json        bool
}

func newWorkflowDeployCommand(profileFlag *string) *cobra.Command {
	opts := &deployOptions{profileFlag: profileFlag}

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Build and deploy the core application to Fly.io",
		RunE: func(cmd *cobra.Command, args []string) error {
			svc := app.New()
			var emitter orchestrator.ProgressEmitter
			if opts.json {
				emitter = orchestrator.NewJSONEmitter(cmd.OutOrStdout())
			} else {
				emitter = orchestrator.NewTextEmitter(cmd.OutOrStdout())
			}
			request := types.DeployRequest{
				AppName:   opts.appFlag,
				OrgSlug:   opts.orgFlag,
				Region:    opts.regionFlag,
				Repo:      opts.repoFlag,
				Verbose:   opts.verbose,
				NoBrowser: opts.noBrowser,
				Stdin:     cmd.InOrStdin(),
				Stdout:    cmd.OutOrStdout(),
				Stderr:    cmd.ErrOrStderr(),
			}
			_, err := svc.Deploy(cmd.Context(), orchestrator.DeployOptions{
				ProfileOverride: strings.TrimSpace(*opts.profileFlag),
				Timeout:         30 * time.Minute,
				DeployRequest:   request,
				Emitter:         emitter,
			})
			return err
		},
	}

	cmd.Flags().StringVar(&opts.appFlag, "app", "", "Fly app name override")
	cmd.Flags().StringVar(&opts.orgFlag, "org", "", "Fly organization slug override")
	cmd.Flags().StringVar(&opts.repoFlag, "repo", "", "Container registry override")
	cmd.Flags().StringVar(&opts.regionFlag, "region", "", "Primary region override")
	cmd.Flags().BoolVar(&opts.verbose, "verbose", false, "Enable verbose pipeline logging")
	cmd.Flags().BoolVar(&opts.noBrowser, "no-browser", false, "Do not automatically open authentication URLs")
	cmd.Flags().BoolVar(&opts.json, "json", false, "Stream newline-delimited JSON progress events")

	return cmd
}
