package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/log"
	serviceruntime "github.com/joeblew999/infra/pkg/service/runtime"
	preflight "github.com/joeblew999/infra/pkg/workflows/preflight"
	natsgo "github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
)

// Build-time variables injected via ldflags
var (
	GitHash   = "dev"
	BuildTime = "unknown"
)

const rootShortDescription = "Infrastructure management without YAML sprawl."

var rootCmd = &cobra.Command{
	Use:     "infra",
	Short:   "Infrastructure management system with goreman supervision",
	Long:    rootShortDescription,
	Version: getVersionString(),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// SetBuildInfo sets build information for display in web pages
func SetBuildInfo(gitHash, buildTime string) {
	GitHash = gitHash
	BuildTime = buildTime
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	ensureBuildInfo()
	registerCommandGroups()

	// Add organized command structure
	RunTools()
	RunWorkflows()
	RunDev()
	assignCommandGroups()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var loggingInitOnce sync.Once
var (
	flagLogLevel  string
	flagLogFormat string
	flagLogOutput []string
)

func initializeLogging() {
	loggingInitOnce.Do(func() {
		level := config.GetLoggingLevel()
		if flagLogLevel != "" {
			level = flagLogLevel
		}

		format := config.GetLoggingFormat()
		if flagLogFormat != "" {
			format = flagLogFormat
		}

		cfg := log.LoadConfig()

		if len(flagLogOutput) > 0 {
			dests, err := destinationsFromFlags(flagLogOutput, level, format)
			if err != nil {
				fallbackLogging(level, format, "Invalid log output flag; using default logging", err)
				return
			}
			cfg.Destinations = dests
		} else if len(cfg.Destinations) == 0 {
			cfg.Destinations = defaultDestinations(config.IsProduction(), level, format)
		}

		normalizeDestinations(&cfg, level, format)

		if len(cfg.Destinations) > 0 {
			nc, natsURL, err := connectLoggingNATS(&cfg)
			if err != nil {
				fallbackLogging(level, format, "Failed to initialize NATS logging; using default logger", err)
				return
			}

			if nc != nil {
				if err := log.InitMultiLoggerWithNATS(cfg, nc); err != nil {
					nc.Close()
					fallbackLogging(level, format, "Failed to initialize multi-destination logging", err)
					return
				}
				log.Info("Initialized multi-destination logging", "destinations", len(cfg.Destinations), "nats_url", natsURL)
				return
			}

			if err := log.InitMultiLogger(cfg); err != nil {
				fallbackLogging(level, format, "Failed to initialize multi-destination logging", err)
				return
			}
			log.Info("Initialized multi-destination logging", "destinations", len(cfg.Destinations))
			return
		}

		log.InitLogger("", level, format == "json")
		log.Info("Initialized default logging", "level", level, "format", format)
	})
}

func destinationsFromFlags(outputs []string, level, format string) ([]log.DestinationConfig, error) {
	var dests []log.DestinationConfig
	for _, output := range outputs {
		switch {
		case output == "stdout":
			dests = append(dests, log.DestinationConfig{Type: "stdout", Level: level, Format: format})
		case output == "stderr":
			dests = append(dests, log.DestinationConfig{Type: "stderr", Level: level, Format: format})
		case strings.HasPrefix(output, "file="):
			path := strings.TrimPrefix(output, "file=")
			dests = append(dests, log.DestinationConfig{Type: "file", Level: level, Format: format, Path: path})
		case strings.HasPrefix(output, "nats"):
			dest, err := parseNATSDestination(output, level)
			if err != nil {
				return nil, err
			}
			dests = append(dests, dest)
		default:
			return nil, fmt.Errorf("unsupported log output '%s'", output)
		}
	}
	return dests, nil
}

func defaultDestinations(isProduction bool, level, format string) []log.DestinationConfig {
	if isProduction {
		return []log.DestinationConfig{
			{Type: "stdout", Level: level, Format: format},
			{Type: "nats", Level: level, Format: "json", URL: config.GetNATSURL(), Subject: config.NATSLogStreamSubject},
		}
	}

	return []log.DestinationConfig{
		{Type: "stdout", Level: level, Format: format},
	}
}

func parseNATSDestination(output, level string) (log.DestinationConfig, error) {
	dest := log.DestinationConfig{Type: "nats", Level: level, Format: "json"}
	spec := strings.TrimSpace(output)

	switch {
	case spec == "nats":
		dest.URL = config.GetNATSURL()
		dest.Subject = config.NATSLogStreamSubject
		return dest, nil

	case strings.HasPrefix(spec, "nats://") || strings.HasPrefix(spec, "nats+tls://"):
		return buildNATSDestinationFromURL(spec, dest)

	case strings.HasPrefix(spec, "nats?subject="):
		query := strings.TrimPrefix(spec, "nats")
		u, err := url.Parse("nats://placeholder" + query)
		if err != nil {
			return dest, fmt.Errorf("invalid NATS log output '%s': %w", output, err)
		}
		dest.URL = config.GetNATSURL()
		dest.Subject = u.Query().Get("subject")
		if dest.Subject == "" {
			dest.Subject = config.NATSLogStreamSubject
		}
		return dest, nil

	case strings.HasPrefix(spec, "nats="):
		value := strings.TrimPrefix(spec, "nats=")
		if strings.HasPrefix(value, "nats://") || strings.HasPrefix(value, "nats+tls://") {
			return buildNATSDestinationFromURL(value, dest)
		}
		dest.URL = config.GetNATSURL()
		if value == "" {
			dest.Subject = config.NATSLogStreamSubject
		} else {
			dest.Subject = value
		}
		return dest, nil

	case strings.HasPrefix(spec, "nats:"):
		subject := strings.TrimPrefix(spec, "nats:")
		dest.URL = config.GetNATSURL()
		if subject == "" {
			dest.Subject = config.NATSLogStreamSubject
		} else {
			dest.Subject = subject
		}
		return dest, nil

	default:
		return dest, fmt.Errorf("unsupported NATS log output '%s'", output)
	}
}

func buildNATSDestinationFromURL(spec string, dest log.DestinationConfig) (log.DestinationConfig, error) {
	u, err := url.Parse(spec)
	if err != nil {
		return dest, fmt.Errorf("invalid NATS URL '%s': %w", spec, err)
	}

	base := *u
	base.Path = ""
	base.RawPath = ""
	base.RawQuery = ""
	base.Fragment = ""
	dest.URL = base.String()
	if dest.URL == "" {
		dest.URL = config.GetNATSURL()
	}

	if subject := u.Query().Get("subject"); subject != "" {
		dest.Subject = subject
	} else {
		trimmedPath := strings.Trim(u.Path, "/")
		if trimmedPath != "" {
			dest.Subject = trimmedPath
		}
	}

	if dest.Subject == "" {
		dest.Subject = config.NATSLogStreamSubject
	}

	return dest, nil
}

func normalizeDestinations(cfg *log.MultiConfig, level, format string) {
	for i := range cfg.Destinations {
		dest := &cfg.Destinations[i]
		if dest.Level == "" {
			dest.Level = level
		}
		if dest.Type == "nats" {
			dest.Format = "json"
			if dest.Subject == "" {
				dest.Subject = config.NATSLogStreamSubject
			}
		} else if dest.Format == "" {
			dest.Format = format
		}
	}
}

func connectLoggingNATS(cfg *log.MultiConfig) (*natsgo.Conn, string, error) {
	var natsURL string
	for i := range cfg.Destinations {
		dest := &cfg.Destinations[i]
		if dest.Type != "nats" {
			continue
		}
		if dest.URL == "" {
			dest.URL = config.GetNATSURL()
		}
		if dest.Subject == "" {
			dest.Subject = config.NATSLogStreamSubject
		}
		if dest.Format != "json" {
			dest.Format = "json"
		}
		if natsURL == "" {
			natsURL = dest.URL
		} else if dest.URL != natsURL {
			return nil, "", fmt.Errorf("multiple NATS URLs specified; only one is supported")
		}
	}

	if natsURL == "" {
		return nil, "", nil
	}

	nc, err := natsgo.Connect(natsURL)
	if err != nil {
		return nil, "", err
	}

	return nc, natsURL, nil
}

func fallbackLogging(level, format, message string, err error) {
	log.InitLogger("", level, format == "json")
	if err != nil {
		log.Warn(message, "error", err)
	} else {
		log.Warn(message)
	}
}

// isRunningInContainer checks if we're running inside a Docker container
// by looking for the .dockerenv file that Docker creates
func isRunningInContainer() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	return false
}

// Removed: getRuntimeGitHash now centralized in pkg/build.GetRuntimeGitHash()

// getVersionString returns version info using build info (DRY)
func getVersionString() string {
	ensureBuildInfo()
	return config.GetFullVersionString()
}

func init() {
	rootCmd.PersistentFlags().String("env", "development", "Environment: production or development")
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug mode")
	rootCmd.PersistentFlags().StringVar(&flagLogLevel, "log-level", "", "Override log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&flagLogFormat, "log-format", "", "Override log format (json, text)")
	rootCmd.PersistentFlags().StringSliceVar(&flagLogOutput, "log-output", nil, "Log destinations e.g. stdout,file=/tmp/app.log")

	cobra.OnInitialize(initializeLogging)

	// Initialize documentation commands
	RunDocs()

	// Set custom help template that shows only our organized structure
	cobra.AddTemplateFuncs(template.FuncMap{
		"commandsForGroup": commandsForGroup,
	})

	rootCmd.SetHelpTemplate(`{{if .Long}}{{trimTrailingWhitespaces .Long}}
{{end}}

Usage:
  {{.UseLine}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}

{{if .HasAvailableSubCommands}}
Commands:
{{- $cmd := .}}
{{- range $group := .Groups}}
{{- $commands := commandsForGroup $cmd $group.ID}}
  {{- if $commands}}
  {{$group.Title}}:
{{- range $commands}}
    {{rpad .Name $cmd.CommandPathPadding}} {{.Short}}
{{- end}}

  {{- end}}
{{- end}}
{{- $ungrouped := commandsForGroup $cmd ""}}
{{- if $ungrouped}}
  Other:
{{- range $ungrouped}}
    {{rpad .Name $cmd.CommandPathPadding}} {{.Short}}
{{- end}}

{{- end}}
{{end}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}
`)
}

func ensureBuildInfo() {
	if GitHash == "dev" {
		if commit := config.GetRuntimeGitHash(); commit != "" {
			GitHash = commit
		}
	}
	if BuildTime == "unknown" {
		BuildTime = time.Now().UTC().Format(time.RFC3339)
	}
	config.SetBuildInfo(GitHash, BuildTime)
}

func defaultServiceOptions(env string) serviceruntime.Options {
	return serviceruntime.Options{
		Mode:      env,
		Preflight: preflight.RunIfNeeded,
	}
}

const (
	runtimeGroupID  = "runtime"
	workflowGroupID = "workflows"
	toolingGroupID  = "tooling"
	devGroupID      = "dev"
	advancedGroupID = "advanced"
)

func registerCommandGroups() {
	rootCmd.AddGroup(&cobra.Group{ID: runtimeGroupID, Title: "Service Runtime"})
	rootCmd.AddGroup(&cobra.Group{ID: workflowGroupID, Title: "Workflows"})
	rootCmd.AddGroup(&cobra.Group{ID: toolingGroupID, Title: "Tooling"})
	rootCmd.AddGroup(&cobra.Group{ID: devGroupID, Title: "Developer"})
	rootCmd.AddGroup(&cobra.Group{ID: advancedGroupID, Title: "Advanced"})
}

func assignCommandGroups() {
	groupMap := map[string]string{
		"runtime":    runtimeGroupID,
		"workflows":  workflowGroupID,
		"tools":      toolingGroupID,
		"dev":        devGroupID,
		"completion": advancedGroupID,
	}

	for _, cmd := range rootCmd.Commands() {
		if groupID, ok := groupMap[cmd.Name()]; ok {
			cmd.GroupID = groupID
		}
	}
}

func commandsForGroup(cmd *cobra.Command, groupID string) []*cobra.Command {
	var out []*cobra.Command
	for _, child := range cmd.Commands() {
		if !child.IsAvailableCommand() || child.Name() == "help" {
			continue
		}
		if child.GroupID == groupID {
			out = append(out, child)
		}
	}
	return out
}
