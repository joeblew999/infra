package ko

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/google/ko/pkg/commands"

	sharedcfg "github.com/joeblew999/infra/core/pkg/shared/config"
)

// PublishOptions controls how the ko publisher runs.
type PublishOptions struct {
	// ConfigPath points at a ko YAML configuration.
	ConfigPath string
	// Args are additional arguments passed after the subcommand, typically
	// import paths or flags like --preserve-import-paths.
	Args []string
	// Env contains environment variables to set for the duration of the ko run.
	Env map[string]string
	// WorkingDir optionally overrides the working directory while running ko.
	WorkingDir string
	// Repo sets the KO_DOCKER_REPO value. Required for publish operations.
	Repo string
	// Tags overrides the image tags ko should publish (defaults to "latest").
	Tags []string
	// Bare instructs ko to publish images directly to KO_DOCKER_REPO without
	// appending import-path derived suffixes.
	Bare bool
}

// Publish builds and publishes images using ko's in-process CLI. It returns
// the image references emitted by ko (one per import path).

func Publish(ctx context.Context, opts PublishOptions) ([]string, error) {
	env := map[string]string{"KO_CONFIG": opts.ConfigPath}
	repo := strings.TrimSpace(opts.Repo)
	if repo == "" && opts.Env != nil {
		repo = strings.TrimSpace(opts.Env["KO_DOCKER_REPO"])
	}
	if repo == "" {
		if inherited, ok := os.LookupEnv("KO_DOCKER_REPO"); ok {
			repo = strings.TrimSpace(inherited)
		}
	}
	if repo == "" {
		return nil, fmt.Errorf("ko publish: repository not provided (Repo or KO_DOCKER_REPO required)")
	}
	if len(opts.Env) > 0 {
		for k, v := range opts.Env {
			env[k] = v
		}
	}
	env["KO_DOCKER_REPO"] = repo
	restore := setEnv(env)
	defer restore()

	root := commands.New()

	restoreDir := func() {}
	if opts.WorkingDir != "" {
		cwd, err := os.Getwd()
		if err == nil {
			restoreDir = func() { _ = os.Chdir(cwd) }
		}
		_ = os.Chdir(opts.WorkingDir)
	}
	defer restoreDir()

	// Default to the "build" (publish) subcommand so behaviour mirrors `ko build`/`ko publish`.
	args := []string{"build"}
	if opts.Bare {
		args = append(args, "--bare")
	}
	for _, tag := range opts.Tags {
		args = append(args, "--tags", tag)
	}
	args = append(args, opts.Args...)

	root.SetArgs(args)
	var buf bytes.Buffer
	outWriter := io.MultiWriter(&buf, os.Stdout)
	errWriter := io.MultiWriter(os.Stderr, &buf)
	root.SetOut(outWriter)
	root.SetErr(errWriter)
	root.SetIn(os.Stdin)

	if err := root.ExecuteContext(ctx); err != nil {
		return nil, err
	}

	var refs []string
	if repo != "" && len(opts.Tags) > 0 {
		for _, tag := range opts.Tags {
			refs = append(refs, fmt.Sprintf("%s:%s", repo, tag))
		}
	}

	if len(refs) == 0 {
		lines := strings.Split(buf.String(), "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			if strings.Contains(trimmed, "/") {
				refs = append(refs, trimmed)
			}
		}
	}

	if len(refs) == 0 {
		return nil, fmt.Errorf("ko publish produced no image references")
	}
	return refs, nil
}

func setEnv(values map[string]string) func() {
	if len(values) == 0 {
		return func() {}
	}
	previous := make(map[string]*string, len(values))
	for key, value := range values {
		prev, ok := os.LookupEnv(key)
		if ok {
			prevCopy := prev
			previous[key] = &prevCopy
		} else {
			previous[key] = nil
		}
		_ = os.Setenv(key, value)
	}
	return func() {
		for key, prev := range previous {
			if prev == nil {
				_ = os.Unsetenv(key)
			} else {
				_ = os.Setenv(key, *prev)
			}
		}
	}
}

type RunOptions struct {
	Args   []string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// RunDocker executes the docker CLI with the supplied arguments, wiring up the
// provided stdio (defaulting to the current process stdio when unset).
func RunDocker(ctx context.Context, opts RunOptions) error {
	if len(opts.Args) == 0 {
		return fmt.Errorf("docker: no arguments provided")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker binary not found: %w", err)
	}

	cmd := exec.CommandContext(ctx, "docker", opts.Args...)
	if opts.Stdin != nil {
		cmd.Stdin = opts.Stdin
	} else {
		cmd.Stdin = os.Stdin
	}
	if opts.Stdout != nil {
		cmd.Stdout = opts.Stdout
	} else {
		cmd.Stdout = os.Stdout
	}
	if opts.Stderr != nil {
		cmd.Stderr = opts.Stderr
	} else {
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}

// ApplyProfileDefaults overlays configuration defaults from the provided
// tooling profile onto the publish options. Explicit values in opts take
// precedence over profile-derived values.
func ApplyProfileDefaults(profile sharedcfg.ToolingProfile, opts PublishOptions) PublishOptions {
	result := opts
	if strings.TrimSpace(result.Repo) == "" {
		result.Repo = profile.KORepository
	}
	if len(result.Tags) == 0 && strings.TrimSpace(profile.TagTemplate) != "" {
		result.Tags = []string{profile.TagTemplate}
	}
	if !result.Bare && profile.Mode == sharedcfg.ToolingModeFly {
		result.Bare = true
	}
	return result
}
