package cmd

import (
	workflowscmd "github.com/joeblew999/infra/pkg/workflows/cmd"
	"github.com/spf13/cobra"
)

// RunWorkflows adds all workflow commands to the root command.
func RunWorkflows() {
	// Get root-level workflow commands from the workflows package
	rootCmds := workflowscmd.GetRootWorkflowCmds()
	for _, cmd := range rootCmds {
		rootCmd.AddCommand(cmd)
	}
	
	// Add NATS cluster management commands
	addClusterCommands(rootCmd)
}

// AddWorkflowsToCLI adds development/build tools to the CLI namespace
func AddWorkflowsToCLI(cliParent *cobra.Command) {
	// Get CLI workflow commands from the workflows package
	cliCmds := workflowscmd.GetWorkflowCmds()
	for _, cmd := range cliCmds {
		cliParent.AddCommand(cmd)
	}
	
	// Add Fly.io commands under CLI namespace
	workflowscmd.AddFlyCommands(cliParent)
}