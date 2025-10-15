package configinit

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
	flysettings "github.com/joeblew999/infra/core/tooling/pkg/fly"
)

type Options struct {
	Profile     sharedcfg.ToolingProfile
	ProfileName string
	RepoRoot    string
	CoreDir     string
	AppName     string
	OrgSlug     string
	Region      string
	Repository  string
	Force       bool
	SkipPrompt  bool
	KoOutput    string
	FlyOutput   string
	Stdout      io.Writer
	Stderr      io.Writer
	Stdin       io.Reader
}

type Result struct {
	Files []FileResult
}

// Plan captures the resolved template targets prior to rendering.
type Plan struct {
	Targets []renderTarget
}

type FileResult struct {
	Path   string
	Action Action
}

type Action string

const (
	ActionCreated   Action = "created"
	ActionUpdated   Action = "updated"
	ActionUnchanged Action = "unchanged"
	ActionSkipped   Action = "skipped"
)

func Run(ctx context.Context, opts Options) (Result, error) {
	if opts.Stdout == nil {
		opts.Stdout = io.Discard
	}
	if opts.Stderr == nil {
		opts.Stderr = io.Discard
	}
	if opts.Stdin == nil {
		opts.Stdin = os.Stdin
	}

	plan, err := Prepare(ctx, opts)
	if err != nil {
		return Result{}, err
	}

	var result Result
	for _, target := range plan.Targets {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		rendered, err := renderTemplate(target.TemplatePath, target.Data)
		if err != nil {
			return result, err
		}
		fileResult, err := writeRendered(target.OutputPath, rendered, opts)
		if err != nil {
			return result, err
		}
		result.Files = append(result.Files, fileResult)
	}

	return result, nil
}

// Prepare resolves configuration data into a plan without writing files.
func Prepare(ctx context.Context, opts Options) (Plan, error) {
	var plan Plan

	repoRoot := strings.TrimSpace(opts.RepoRoot)
	if repoRoot == "" {
		return plan, errors.New("config init: repository root not provided")
	}

	coreDir := strings.TrimSpace(opts.CoreDir)
	if coreDir == "" {
		coreDir = filepath.Join(repoRoot, "core")
	}

	modulePath, err := detectModulePath(coreDir)
	if err != nil {
		return plan, err
	}

	importPath := firstNonEmpty(opts.Profile.ImportPath, "./cmd/core")

	repository := strings.TrimSpace(opts.Repository)
	if repository == "" {
		repository = strings.TrimSpace(opts.Profile.KORepository)
	}
	if repository == "" {
		repository = defaultRepository(importPath)
	}

	appName := firstNonEmpty(opts.AppName, opts.Profile.FlyApp)
	appName = strings.TrimSpace(appName)
	if appName == "" {
		return plan, errors.New("config init: Fly app name missing; provide --app or configure the profile")
	}

	settings, err := flysettings.LoadSettings()
	if err != nil {
		return plan, fmt.Errorf("load fly settings: %w", err)
	}

	region := firstNonEmpty(opts.Region, settings.RegionCode, opts.Profile.FlyRegion)
	region = strings.TrimSpace(region)
	if region == "" {
		return plan, errors.New("config init: Fly region missing; provide --region or configure the profile")
	}

	orgSlug := strings.TrimSpace(firstNonEmpty(opts.OrgSlug, settings.OrgSlug, opts.Profile.FlyOrg))

	koTemplatePath, err := resolveTemplatePath(coreDir, opts.Profile.KoTemplate, "config/templates/ko.yaml.tmpl")
	if err != nil {
		return plan, err
	}
	flyTemplatePath, err := resolveTemplatePath(coreDir, opts.Profile.FlyTemplate, "config/templates/fly.toml.tmpl")
	if err != nil {
		return plan, err
	}

	koOutput := strings.TrimSpace(opts.KoOutput)
	if koOutput == "" {
		koOutput = filepath.Join(coreDir, firstNonEmpty(opts.Profile.KoConfig, ".ko.yaml"))
	} else if !filepath.IsAbs(koOutput) {
		koOutput = filepath.Join(coreDir, koOutput)
	}
	flyOutput := strings.TrimSpace(opts.FlyOutput)
	if flyOutput == "" {
		flyOutput = filepath.Join(repoRoot, firstNonEmpty(opts.Profile.FlyConfig, "fly.toml"))
	} else if !filepath.IsAbs(flyOutput) {
		flyOutput = filepath.Join(repoRoot, flyOutput)
	}

	koData := koTemplateData{
		BaseImage:  "alpine:latest",
		Module:     modulePath,
		ImportPath: importPath,
		BuildID:    "core-app",
		Platforms:  []string{"linux/amd64", "linux/arm64"},
		BuildEnv:   []string{"CGO_ENABLED=0", "GOOS=linux"},
		BuildFlags: []string{"-trimpath"},
		LDFlags:    []string{"-s", "-w"},
		DefaultEnv: []string{"CGO_ENABLED=0", "GOOS=linux"},
		Repository: repository,
	}

	flyData := flyTemplateData{
		AppName:       appName,
		PrimaryRegion: region,
		KillSignal:    "SIGINT",
		KillTimeout:   "5s",
		Environment: []envVar{
			{Key: "ENVIRONMENT", Value: "production"},
			{Key: "PORT", Value: "1337"},
		},
		Volume: volumeConfig{
			Source:      "core_data",
			Destination: "/app/.data",
		},
		InternalPort:   1337,
		Processes:      []string{"app"},
		HTTPChecksPath: "/status",
		MetricsPort:    9091,
		MetricsPath:    "/metrics",
		OrgSlug:        orgSlug,
	}

	sort.Slice(flyData.Environment, func(i, j int) bool {
		return flyData.Environment[i].Key < flyData.Environment[j].Key
	})
	for i := range flyData.Environment {
		flyData.Environment[i].Value = escapeSingleQuotes(flyData.Environment[i].Value)
	}

	plan.Targets = []renderTarget{
		{TemplatePath: koTemplatePath, OutputPath: koOutput, Data: koData},
		{TemplatePath: flyTemplatePath, OutputPath: flyOutput, Data: flyData},
	}

	return plan, nil
}

// Render renders the prepared plan to memory.
func Render(plan Plan) (map[string][]byte, error) {
	outputs := make(map[string][]byte, len(plan.Targets))
	for _, target := range plan.Targets {
		rendered, err := renderTemplate(target.TemplatePath, target.Data)
		if err != nil {
			return nil, err
		}
		outputs[target.OutputPath] = rendered
	}
	return outputs, nil
}

type renderTarget struct {
	TemplatePath string
	OutputPath   string
	Data         any
}

type koTemplateData struct {
	BaseImage  string
	Module     string
	ImportPath string
	BuildID    string
	Platforms  []string
	BuildEnv   []string
	BuildFlags []string
	LDFlags    []string
	DefaultEnv []string
	Repository string
}

type flyTemplateData struct {
	AppName        string
	PrimaryRegion  string
	KillSignal     string
	KillTimeout    string
	Environment    []envVar
	Volume         volumeConfig
	InternalPort   int
	Processes      []string
	HTTPChecksPath string
	MetricsPort    int
	MetricsPath    string
	OrgSlug        string
}

type envVar struct {
	Key   string
	Value string
}

type volumeConfig struct {
	Source      string
	Destination string
}

func renderTemplate(tplPath string, data any) ([]byte, error) {
	tpl, err := template.New(filepath.Base(tplPath)).ParseFiles(tplPath)
	if err != nil {
		return nil, fmt.Errorf("load template %q: %w", tplPath, err)
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render template %q: %w", tplPath, err)
	}
	return buf.Bytes(), nil
}

func writeRendered(outputPath string, rendered []byte, opts Options) (FileResult, error) {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return FileResult{}, fmt.Errorf("create parent directory for %q: %w", outputPath, err)
	}

	existing, readErr := os.ReadFile(outputPath)
	fileExists := readErr == nil
	if fileExists {
		if bytes.Equal(existing, rendered) {
			return FileResult{Path: outputPath, Action: ActionUnchanged}, nil
		}
		if !opts.Force {
			if opts.SkipPrompt {
				return FileResult{Path: outputPath, Action: ActionSkipped}, nil
			}
			proceed, promptErr := promptOverwrite(opts.Stdout, opts.Stdin, outputPath)
			if promptErr != nil {
				return FileResult{}, promptErr
			}
			if !proceed {
				return FileResult{Path: outputPath, Action: ActionSkipped}, nil
			}
		}
	} else if !errors.Is(readErr, os.ErrNotExist) {
		return FileResult{}, fmt.Errorf("read existing %q: %w", outputPath, readErr)
	}

	action := ActionCreated
	if fileExists {
		action = ActionUpdated
	}
	if err := os.WriteFile(outputPath, rendered, 0o644); err != nil {
		return FileResult{}, fmt.Errorf("write %q: %w", outputPath, err)
	}
	return FileResult{Path: outputPath, Action: action}, nil
}

func promptOverwrite(out io.Writer, in io.Reader, path string) (bool, error) {
	fmt.Fprintf(out, "Overwrite %s? [y/N]: ", path)
	reader := bufio.NewReader(in)
	response, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("read overwrite confirmation: %w", err)
	}
	response = strings.TrimSpace(strings.ToLower(response))
	switch response {
	case "y", "yes":
		return true, nil
	case "", "n", "no":
		return false, nil
	default:
		return false, fmt.Errorf("config init: unrecognised response %q", response)
	}
}

func resolveTemplatePath(coreDir, candidate, fallback string) (string, error) {
	value := firstNonEmpty(candidate, fallback)
	if strings.TrimSpace(value) == "" {
		return "", errors.New("template path missing")
	}
	if filepath.IsAbs(value) {
		if _, err := os.Stat(value); err != nil {
			return "", fmt.Errorf("template %q not found: %w", value, err)
		}
		return value, nil
	}
	path := filepath.Join(coreDir, value)
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("template %q not found: %w", path, err)
	}
	return path, nil
}

func detectModulePath(coreDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(coreDir, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("read go.mod: %w", err)
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			module := strings.TrimSpace(strings.TrimPrefix(line, "module "))
			if module != "" {
				return module, nil
			}
		}
	}
	return "", errors.New("config init: unable to detect module path from go.mod")
}

func escapeSingleQuotes(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func defaultRepository(importPath string) string {
	cleaned := strings.TrimSpace(importPath)
	if cleaned == "" {
		cleaned = "core"
	}
	for strings.HasPrefix(cleaned, "./") || strings.HasPrefix(cleaned, "../") {
		cleaned = strings.TrimPrefix(cleaned, "./")
		cleaned = strings.TrimPrefix(cleaned, "../")
	}
	cleaned = strings.TrimSuffix(cleaned, "/")
	base := filepath.Base(cleaned)
	if base == "" || base == "." || base == "/" {
		base = "core"
	}
	slug := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r + 32
		case r >= '0' && r <= '9':
			return r
		case r == '-' || r == '_':
			return r
		default:
			return '-'
		}
	}, base)
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "core"
	}
	return fmt.Sprintf("ko.local/%s", slug)
}
