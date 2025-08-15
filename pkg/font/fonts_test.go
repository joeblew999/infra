package font

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFontManager(t *testing.T) {
	// Use temp directory for testing
	tempDir := t.TempDir()
	fontDir := filepath.Join(tempDir, "font")
	
	// Create manager with temp directory
	manager := &Manager{
		cacheDir: fontDir,
		registry: NewRegistryWithPath(filepath.Join(fontDir, RegistryFilename)),
	}

	t.Run("ListAvailableFonts", func(t *testing.T) {
		fonts := manager.List()
		assert.Empty(t, fonts, "Should start with no fonts")
	})

	t.Run("CheckFontAvailability", func(t *testing.T) {
		available := manager.Available("Roboto", 400)
		assert.False(t, available, "Roboto 400 should not be available initially")
	})

	t.Run("FontPathsCorrect", func(t *testing.T) {
		expected := filepath.Join(tempDir, "font")
		assert.Equal(t, expected, manager.cacheDir)
	})

	t.Run("RegistryCreation", func(t *testing.T) {
		registryPath := filepath.Join(tempDir, "test-registry.json")
		registry := NewRegistryWithPath(registryPath)
		assert.NotNil(t, registry)
		// Registry file is created on first write, not creation
	assert.NoFileExists(t, registry.path)
	})

	t.Run("RegistryOperations", func(t *testing.T) {
		registryPath := filepath.Join(tempDir, "test-registry.json")
		registry := NewRegistryWithPath(registryPath)
		
		font := Font{
			Family: "Roboto",
			Weight: 400,
			Style:  "normal",
			Format: "woff2",
		}

		// Add mock font
		mockPath := filepath.Join(tempDir, "mock.woff2")
		require.NoError(t, os.WriteFile(mockPath, []byte("mock font data"), 0644))
		
		info := FontInfo{
			Font:    font,
			Path:    mockPath,
			Version: "1.0",
			Source:  "test",
		}

		// Test registration
		err := registry.Add(info)
		require.NoError(t, err)

		// Test retrieval
		path, exists := registry.GetPath(font)
		assert.True(t, exists)
		assert.Equal(t, mockPath, path)

		// Test listing
		fonts := registry.List()
		assert.Len(t, fonts, 1)
		assert.Equal(t, "Roboto", fonts[0].Family)

		// Test removal
		err = registry.Remove(font)
		require.NoError(t, err)
		
		_, exists = registry.GetPath(font)
		assert.False(t, exists)
	})
}

func TestRegistryKeyGeneration(t *testing.T) {
	registry := NewRegistryWithPath("/tmp/test-registry.json")
	
	font := Font{
		Family: "Roboto",
		Weight: 400,
		Style:  "normal",
		Format: "woff2",
	}
	
	key := registry.key(font)
	assert.Equal(t, "roboto-400-normal-woff2", key)
}

func TestFontFormats(t *testing.T) {
	font := Font{
		Family: "Test",
		Weight: 400,
		Style:  "italic",
		Format: "ttf",
	}

	assert.Equal(t, "Test", font.Family)
	assert.Equal(t, 400, font.Weight)
	assert.Equal(t, "italic", font.Style)
	assert.Equal(t, "ttf", font.Format)
}

func TestFileSystemOperations(t *testing.T) {
	tempDir := t.TempDir()
	fontPath := filepath.Join(tempDir, "font")
	
	// Test directory creation
	assert.NoError(t, os.MkdirAll(fontPath, 0755))
	assert.DirExists(t, fontPath)

	// Test family directory creation
	familyPath := filepath.Join(fontPath, "TestFamily")
	assert.NoError(t, os.MkdirAll(familyPath, 0755))
	assert.DirExists(t, familyPath)
}

// NewRegistryWithPath creates a registry with a custom path for testing
func NewRegistryWithPath(path string) *Registry {
	return &Registry{
		path: path,
		data: make(map[string]FontInfo),
	}
}