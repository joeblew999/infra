package cmd

import (
	"sort"

	natscmd "github.com/joeblew999/infra/pkg/nats/cmd"
	workflowscmd "github.com/joeblew999/infra/pkg/workflows/cmd"
	"github.com/spf13/cobra"
)

// RunWorkflows mounts the workflow namespace onto the root command.
func RunWorkflows() {
	workflowsCmd := &cobra.Command{
		Use:   "workflows",
		Short: "Application build, deployment, and maintenance workflows",
		Long:  "High-level workflows for building, testing, and deploying the infrastructure.",
	}

	seen := make(map[string]struct{})
	add := func(cmd *cobra.Command) {
		if cmd == nil {
			return
		}
		name := cmd.Name()
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		workflowsCmd.AddCommand(cmd)
	}

	for _, cmd := range workflowscmd.GetRootWorkflowCmds() {
		add(cmd)
	}
	for _, cmd := range workflowscmd.GetWorkflowCmds() {
		add(cmd)
	}

	workflowscmd.AddFlyCommands(workflowsCmd)

	natscmd.RegisterWorkflows(workflowsCmd)

	sort.SliceStable(workflowsCmd.Commands(), func(i, j int) bool {
		return workflowsCmd.Commands()[i].Name() < workflowsCmd.Commands()[j].Name()
	})

	rootCmd.AddCommand(workflowsCmd)
}
