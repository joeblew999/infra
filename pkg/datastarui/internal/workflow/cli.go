package workflow

import (
	"flag"
	"log"
	"strings"
	"time"
)

// CLIOptions captures the command-line overrides supported by the helper binaries.
type CLIOptions struct {
	TailwindInput   string
	TailwindOutput  string
	TailwindContent string
	Binary          string
	ServerCommand   string
	BaseURL         string
	Workflow        string
	Src             string
	Headed          bool
	Timeout         time.Duration
	ShowVersion     bool
}

// RegisterFlags populates the flag set with shared options and returns a struct
// that will receive the parsed values.
func RegisterFlags(fs *flag.FlagSet, cfg Config, defaultTimeout time.Duration) *CLIOptions {
    opts := &CLIOptions{}

	fs.StringVar(&opts.TailwindInput, "tailwind-input", cfg.TailwindInput, "path to Tailwind input file")
	fs.StringVar(&opts.TailwindOutput, "tailwind-output", cfg.TailwindOutput, "path to Tailwind output file")
	fs.StringVar(&opts.TailwindContent, "tailwind-content", strings.Join(cfg.TailwindContent, ","), "comma-separated Tailwind content globs")
	fs.StringVar(&opts.Binary, "binary", cfg.Binary, "compiled server binary name (relative to source root)")

	defaultServer := ""
	if len(cfg.ServerCommand) > 0 {
		defaultServer = strings.Join(cfg.ServerCommand, " ")
	}
	fs.StringVar(&opts.ServerCommand, "server-command", defaultServer, "custom server command (quoted string, e.g. './bin/app --flag')")

	fs.StringVar(&opts.BaseURL, "base-url", cfg.BaseURL, "base URL used for readiness checks and Playwright suite")
	fs.StringVar(&opts.Workflow, "workflow", string(cfg.Workflow), "automation workflow to use (bun or node)")
	fs.StringVar(&opts.Src, "src", "", "path to the DatastarUI project (defaults to the packaged sample app)")
	fs.BoolVar(&opts.Headed, "headed", false, "run Playwright in headed mode")

	fs.DurationVar(&opts.Timeout, "timeout", defaultTimeout, "overall timeout for the run")
	fs.BoolVar(&opts.ShowVersion, "version", false, "print command version and exit")

	return opts
}

// Apply merges the parsed options into the provided config.
func (o *CLIOptions) Apply(cfg *Config) {
	cfg.TailwindInput = o.TailwindInput
	cfg.TailwindOutput = o.TailwindOutput
	cfg.BaseURL = o.BaseURL
	cfg.Binary = o.Binary

	if strings.TrimSpace(o.TailwindContent) != "" {
		cfg.TailwindContent = splitCSV(o.TailwindContent)
	}
	if strings.TrimSpace(o.ServerCommand) != "" {
		cfg.ServerCommand = strings.Fields(o.ServerCommand)
	} else {
		cfg.ServerCommand = nil
	}

	if strings.TrimSpace(o.Workflow) != "" {
		cfg.Workflow = WorkflowMode(strings.ToLower(strings.TrimSpace(o.Workflow)))
	}
	cfg.Headed = o.Headed
}

// Report logs the resolved configuration to aid debugging and reproducibility.
func (o *CLIOptions) Report(runner, version, src string, cfg Config) {
	log.Printf("%s %s", runner, version)
	log.Printf("config: src=%s tailwind-in=%s tailwind-out=%s content=%v binary=%s base-url=%s server=%v workflow=%s headed=%t", src, cfg.TailwindInput, cfg.TailwindOutput, cfg.TailwindContent, cfg.Binary, cfg.BaseURL, cfg.ServerCommand, cfg.Workflow, cfg.Headed)
}

func splitCSV(input string) []string {
	parts := strings.Split(input, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
