package auth

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/joeblew999/infra/pkg/config"
	"github.com/joeblew999/infra/pkg/dep"
	"github.com/joeblew999/infra/pkg/log"
)

// Artifacts captures the generated authentication material for the NATS cluster.
type Artifacts struct {
	OperatorJWT           string
	SystemAccountID       string
	SystemAccountJWT      string
	ApplicationAccountID  string
	ApplicationAccountJWT string
	ApplicationCredsPath  string
	SystemCredsPath       string
	StoreDir              string
}

var ensureMu sync.Mutex

// Ensure generates (if needed) and returns authentication artifacts for the NATS cluster.
func Ensure(ctx context.Context) (*Artifacts, error) {
	ensureMu.Lock()
	defer ensureMu.Unlock()

	if err := dep.InstallBinary(config.BinaryNsc, false); err != nil {
		return nil, fmt.Errorf("install nsc binary: %w", err)
	}

	storeDir := config.GetNATSAuthStorePath()
	credsDir := config.GetNATSAuthCredsPath()

	paths := requiredPaths()

	if !artifactsReady(paths) {
		if err := os.MkdirAll(storeDir, 0o755); err != nil {
			return nil, fmt.Errorf("create nsc store dir: %w", err)
		}
		if err := os.MkdirAll(credsDir, 0o755); err != nil {
			return nil, fmt.Errorf("create creds dir: %w", err)
		}

		if err := initializeStore(ctx, storeDir); err != nil {
			return nil, err
		}

		if err := generateCreds(ctx, storeDir, config.GetNATSApplicationCredsPath(), config.NATSApplicationAccount, config.NATSApplicationUserName); err != nil {
			return nil, err
		}
		if err := generateCreds(ctx, storeDir, config.GetNATSSystemCredsPath(), config.NATSSystemAccountName, config.NATSSystemUserName); err != nil {
			return nil, err
		}
	}

	return loadArtifacts(paths, storeDir)
}

func requiredPaths() map[string]string {
	return map[string]string{
		"operator":           config.GetNATSOperatorJWTPath(),
		"systemAccount":      config.GetNATSAccountJWTPath(config.NATSSystemAccountName),
		"applicationAccount": config.GetNATSAccountJWTPath(config.NATSApplicationAccount),
		"systemCreds":        config.GetNATSSystemCredsPath(),
		"applicationCreds":   config.GetNATSApplicationCredsPath(),
	}
}

func artifactsReady(paths map[string]string) bool {
	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			return false
		}
	}
	return true
}

func initializeStore(ctx context.Context, storeDir string) error {
	log.Info("Bootstrapping NATS identity store", "store", storeDir)

	if err := runNSC(ctx, storeDir, "init", "--name", config.NATSOperatorName); err != nil {
		return fmt.Errorf("nsc init: %w", err)
	}

	return nil
}

func generateCreds(ctx context.Context, storeDir, outputPath, account, user string) error {
	if err := os.Remove(outputPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cleanup creds %s: %w", outputPath, err)
	}

	args := []string{
		"generate", "creds",
		"--account", account,
		"--name", user,
		"--output-file", outputPath,
	}

	if err := runNSC(ctx, storeDir, args...); err != nil {
		return fmt.Errorf("generate creds for %s/%s: %w", account, user, err)
	}

	return nil
}

func runNSC(ctx context.Context, storeDir string, args ...string) error {
	binary, err := dep.Get(config.BinaryNsc)
	if err != nil {
		return fmt.Errorf("resolve nsc binary: %w", err)
	}

	abs, err := filepath.Abs(binary)
	if err != nil {
		return fmt.Errorf("abs nsc binary: %w", err)
	}

	cmdArgs := append([]string{}, args...)
	cmdArgs = append(cmdArgs, "--all-dirs", storeDir)
	if len(args) > 0 && args[0] == "init" {
		cmdArgs = append(cmdArgs, "--dir", storeDir)
	}

	cmd := exec.CommandContext(ctx, abs, cmdArgs...)
	cmd.Env = append(os.Environ(),
		"NSC_STORE_DIR="+storeDir,
		"NSC_NO_GITHUB_UPDATES=1",
	)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("nsc %s failed: %w\n%s", strings.Join(args, " "), err, out.String())
	}

	return nil
}

func loadArtifacts(paths map[string]string, storeDir string) (*Artifacts, error) {
	read := func(key string) (string, error) {
		data, err := os.ReadFile(paths[key])
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(data)), nil
	}

	operatorJWT, err := read("operator")
	if err != nil {
		return nil, fmt.Errorf("read operator jwt: %w", err)
	}

	systemJWT, err := read("systemAccount")
	if err != nil {
		return nil, fmt.Errorf("read system account jwt: %w", err)
	}

	appJWT, err := read("applicationAccount")
	if err != nil {
		return nil, fmt.Errorf("read application account jwt: %w", err)
	}

	systemSubject, err := accountSubject(systemJWT)
	if err != nil {
		return nil, fmt.Errorf("parse system account claims: %w", err)
	}

	appSubject, err := accountSubject(appJWT)
	if err != nil {
		return nil, fmt.Errorf("parse application account claims: %w", err)
	}

	return &Artifacts{
		OperatorJWT:           operatorJWT,
		SystemAccountID:       systemSubject,
		SystemAccountJWT:      systemJWT,
		ApplicationAccountID:  appSubject,
		ApplicationAccountJWT: appJWT,
		ApplicationCredsPath:  paths["applicationCreds"],
		SystemCredsPath:       paths["systemCreds"],
		StoreDir:              storeDir,
	}, nil
}

func accountSubject(jwtString string) (string, error) {
	parts := strings.Split(jwtString, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid jwt structure")
	}
	payloadRaw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("decode jwt payload: %w", err)
	}
	var payload struct {
		Sub string `json:"sub"`
	}
	if err := json.Unmarshal(payloadRaw, &payload); err != nil {
		return "", fmt.Errorf("unmarshal jwt payload: %w", err)
	}
	if payload.Sub == "" {
		return "", fmt.Errorf("jwt payload missing subject")
	}
	return payload.Sub, nil
}
