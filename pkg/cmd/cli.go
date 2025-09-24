package cmd

import (
	aicmd "github.com/joeblew999/infra/pkg/ai/cmd"
	"github.com/joeblew999/infra/pkg/bento"
	caddycmd "github.com/joeblew999/infra/pkg/caddy/cmd"
	"github.com/joeblew999/infra/pkg/conduit"
	"github.com/joeblew999/infra/pkg/config"
	debugcmd "github.com/joeblew999/infra/pkg/debug/cmd"
	deckcmd "github.com/joeblew999/infra/pkg/deck/cmd"
	depcmd "github.com/joeblew999/infra/pkg/dep/cmd"
	fontcmd "github.com/joeblew999/infra/pkg/font/cmd"
	goremanCmds "github.com/joeblew999/infra/pkg/goreman/cmd"
	gozerocmd "github.com/joeblew999/infra/pkg/gozero/cmd"
	moxcmd "github.com/joeblew999/infra/pkg/mox/cmd"
	natscmd "github.com/joeblew999/infra/pkg/nats/cmd"
	"github.com/joeblew999/infra/pkg/pocketbase"
	tokicmd "github.com/joeblew999/infra/pkg/toki/cmd"
	utmcmd "github.com/joeblew999/infra/pkg/utm/cmd"
	webappcmd "github.com/joeblew999/infra/pkg/webapp/cmd"
	xtemplatecmd "github.com/joeblew999/infra/pkg/xtemplate/cmd"
	"github.com/spf13/cobra"
)

// caddyCmd is now delegated to the caddy package
var caddyCmd = caddycmd.GetCaddyCmd()

// toolsCmd is the parent command for all auxiliary tooling wrappers
var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Developer tooling and binary wrappers",
	Long: `Access to supporting tools and utilities:

SCALING & DEPLOYMENT:
  fly              Fly.io operations and scaling
  
DEVELOPMENT TOOLS:
  deck             Deck visualization tools  
  gozero           Go-zero microservices operations
  toki             Translation and i18n workflow
  xtemplate        HTML/template web development server
  
BINARY TOOLS:
  tofu             OpenTofu infrastructure as code
  task             Task runner and build automation
  caddy            Web server with automatic HTTPS
  ko               Container image builder for Go
  flyctl           Direct flyctl commands
  nats             NATS messaging operations
  bento            Stream processing operations
  mox              Mox mail server
  utm              UTM virtual machine management (macOS)

Use "infra tools [tool] --help" for detailed information about each tool.`,
}

// RunTools mounts the tooling namespace onto the root command.
func RunTools() {
	registerToolCommands(toolsCmd)
	toolsCmd.AddCommand(caddyCmd)
	natscmd.RegisterCLI(toolsCmd)
	toolsCmd.AddCommand(bento.NewBentoCmd())
	moxcmd.Register(toolsCmd)

	aicmd.Register(toolsCmd)
	debugcmd.Register(toolsCmd)
	tokicmd.Register(toolsCmd)
	deckcmd.Register(toolsCmd)
	fontcmd.Register(toolsCmd)
	gozerocmd.Register(toolsCmd)
	utmcmd.Register(toolsCmd)
	goremanCmds.Register(toolsCmd)
	xtemplatecmd.Register(toolsCmd)
	webappcmd.Register(toolsCmd)
	toolsCmd.AddCommand(depcmd.Cmd)
	toolsCmd.AddCommand(config.Cmd)
	toolsCmd.AddCommand(pocketbase.Cmd)
	toolsCmd.AddCommand(conduit.Cmd)

	rootCmd.AddCommand(toolsCmd)
}
