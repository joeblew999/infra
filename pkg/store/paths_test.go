package store

import (
	"path/filepath"
	"testing"
)

func TestGetDepPath(t *testing.T) {
	expected := filepath.Join(".", ".dep")
	actual := GetDepPath()
	if actual != expected {
		t.Errorf("GetDepPath() = %s; want %s", actual, expected)
	}
}

func TestGetBinPath(t *testing.T) {
	expected := filepath.Join(".", ".bin")
	actual := GetBinPath()
	if actual != expected {
		t.Errorf("GetBinPath() = %s; want %s", actual, expected)
	}
}

func TestGetTaskfilesPath(t *testing.T) {
	expected := filepath.Join(".", "taskfiles")
	actual := GetTaskfilesPath()
	if actual != expected {
		t.Errorf("GetTaskfilesPath() = %s; want %s", actual, expected)
	}
}

func TestGetDataPath(t *testing.T) {
	expected := filepath.Join(".", ".data")
	actual := GetDataPath()
	if actual != expected {
		t.Errorf("GetDataPath() = %s; want %s", actual, expected)
	}
}
