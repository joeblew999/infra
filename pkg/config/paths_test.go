package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetDepPath(t *testing.T) {
	expected := filepath.Join(".", DepDir)
	actual := GetDepPath()
	if actual != expected {
		t.Errorf("GetDepPath() = %s; want %s", actual, expected)
	}
}

func TestGetBinPath(t *testing.T) {
	expected := filepath.Join(".", BinDir)
	actual := GetBinPath()
	if actual != expected {
		t.Errorf("GetBinPath() = %s; want %s", actual, expected)
	}
}

func TestGetTaskfilesPath(t *testing.T) {
	expected := filepath.Join(".", TaskfilesDir)
	actual := GetTaskfilesPath()
	if actual != expected {
		t.Errorf("GetTaskfilesPath() = %s; want %s", actual, expected)
	}
}

func TestGetDataPath(t *testing.T) {
	origEnv := os.Getenv(EnvVarEnvironment)
	origAppRoot := os.Getenv(EnvVarAppRoot)
	origCluster := os.Getenv(EnvVarNATSCluster)
	origArgs := append([]string{}, os.Args...)
	t.Cleanup(func() {
		os.Setenv(EnvVarEnvironment, origEnv)
		os.Setenv(EnvVarAppRoot, origAppRoot)
		os.Setenv(EnvVarNATSCluster, origCluster)
		os.Args = origArgs
	})

	os.Unsetenv(EnvVarEnvironment)
	os.Unsetenv(EnvVarAppRoot)

	// Running under go test should always use the isolated test data directory.
	if actual := GetDataPath(); actual != filepath.Join(".", TestDataDir) {
		t.Fatalf("GetDataPath() test env = %s; want %s", actual, filepath.Join(".", TestDataDir))
	}

	// Simulate non-test development environment.
	os.Args = []string{"infra"}
	os.Unsetenv(EnvVarEnvironment)
	devExpected := filepath.Join(GetAppRoot(), DataDir)
	if actual := GetDataPath(); actual != devExpected {
		t.Fatalf("GetDataPath() development = %s; want %s", actual, devExpected)
	}

	// Production should use the same resolver but honour container-friendly defaults.
	os.Setenv(EnvVarEnvironment, EnvProduction)
	prodExpected := filepath.Join(GetAppRoot(), DataDir)
	if actual := GetDataPath(); actual != prodExpected {
		t.Fatalf("GetDataPath() production = %s; want %s", actual, prodExpected)
	}

	// APP_ROOT override takes precedence regardless of environment.
	customRoot := filepath.Join(".", "custom-app-root")
	os.Setenv(EnvVarAppRoot, customRoot)
	if actual := GetDataPath(); actual != filepath.Join(filepath.Clean(customRoot), DataDir) {
		t.Fatalf("GetDataPath() with APP_ROOT override = %s; want %s", actual, filepath.Join(filepath.Clean(customRoot), DataDir))
	}
}

func TestShouldEnsureNATSCluster(t *testing.T) {
	origEnv := os.Getenv(EnvVarEnvironment)
	origFlag := os.Getenv(EnvVarNATSCluster)
	t.Cleanup(func() {
		os.Setenv(EnvVarEnvironment, origEnv)
		os.Setenv(EnvVarNATSCluster, origFlag)
	})

	os.Unsetenv(EnvVarEnvironment)
	os.Unsetenv(EnvVarNATSCluster)
	if ensure := ShouldEnsureNATSCluster(); ensure {
		t.Fatalf("expected default to skip cluster")
	}

	os.Setenv(EnvVarEnvironment, EnvProduction)
	os.Unsetenv(EnvVarNATSCluster)
	if ensure := ShouldEnsureNATSCluster(); ensure {
		t.Fatalf("expected production default to skip cluster")
	}

	os.Setenv(EnvVarEnvironment, EnvDevelopment)
	os.Setenv(EnvVarNATSCluster, "true")
	if ensure := ShouldEnsureNATSCluster(); !ensure {
		t.Fatalf("expected explicit true to enable cluster")
	}

	os.Setenv(EnvVarNATSCluster, "0")
	if ensure := ShouldEnsureNATSCluster(); ensure {
		t.Fatalf("expected explicit false to disable cluster")
	}
}
