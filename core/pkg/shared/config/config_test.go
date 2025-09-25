package config

import (
	"os"
	"testing"
)

func TestEnvironmentDefaults(t *testing.T) {
	t.Setenv(EnvVarEnvironment, "")
	if got := Environment(); got != EnvDevelopment {
		t.Fatalf("Environment() default = %q, want %q", got, EnvDevelopment)
	}
}

func TestShouldEnsureBusCluster(t *testing.T) {
	t.Setenv(EnvVarBusClusterEnabled, "true")
	if !ShouldEnsureBusCluster() {
		t.Fatal("expected bus cluster toggle to be true")
	}
	t.Setenv(EnvVarBusClusterEnabled, "false")
	if ShouldEnsureBusCluster() {
		t.Fatal("expected bus cluster toggle to be false")
	}
}

func TestPathHelpers(t *testing.T) {
	t.Setenv(EnvVarAppRoot, "/tmp/core")
	expect := map[string]string{
		"dep":  "/tmp/core/.dep",
		"bin":  "/tmp/core/.bin",
		"data": "/tmp/core/.data",
		"logs": "/tmp/core/.logs",
	}
	if got := GetDepPath(); got != expect["dep"] {
		t.Fatalf("GetDepPath() = %q, want %q", got, expect["dep"])
	}
	if got := GetBinPath(); got != expect["bin"] {
		t.Fatalf("GetBinPath() = %q, want %q", got, expect["bin"])
	}
	origArgs := os.Args
	os.Args = []string{"core"}
	if got := GetDataPath(); got != expect["data"] {
		os.Args = origArgs
		t.Fatalf("GetDataPath() = %q, want %q", got, expect["data"])
	}
	os.Args = origArgs
	t.Setenv(EnvVarEnvironment, EnvDevelopment)
	t.Setenv("GO_WANT_HELPER_PROCESS", "")
	if got := GetLogsPath(); got != expect["logs"] {
		t.Fatalf("GetLogsPath() = %q, want %q", got, expect["logs"])
	}
}
