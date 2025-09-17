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
	entChanged := os.Getenv("ENVIRONMENT")
	defer os.Setenv("ENVIRONMENT", entChanged)

	// default (development) should point at .data-test when running under go test
	os.Unsetenv("ENVIRONMENT")
	actual := GetDataPath()
	expected := filepath.Join(".", TestDataDir)
	if actual != expected {
		t.Errorf("GetDataPath() dev default = %s; want %s", actual, expected)
	}

	// production flag should force .data
	os.Setenv("ENVIRONMENT", "production")
	actual = GetDataPath()
	const prodExpected = "/app/.data"
	if actual != prodExpected {
		t.Errorf("GetDataPath() production = %s; want %s", actual, prodExpected)
	}
}
