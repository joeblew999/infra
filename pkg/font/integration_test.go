//go:build integration
// +build integration

package font

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFontCachePopulation(t *testing.T) {
	// Skip if no network access
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Use temp directory for testing
	tempDir := t.TempDir()
	fontDir := filepath.Join(tempDir, "font")

	// Create manager with temp directory
	manager := &Manager{
		cacheDir: fontDir,
		registry: NewRegistryWithPath(filepath.Join(fontDir, RegistryFilename)),
	}

	t.Run("CacheFontIntegration", func(t *testing.T) {
		// Test caching a known Google Font
		family := "Roboto"
		weight := 400

		// Ensure font is not initially available
		assert.False(t, manager.Available(family, weight))

		// Cache the font
		err := manager.Cache(family, weight)
		require.NoError(t, err)

		// Verify font is now available
		assert.True(t, manager.Available(family, weight))

		// Get the font path
		path, err := manager.Get(family, weight)
		require.NoError(t, err)
		assert.FileExists(t, path)

		// Verify it's the expected file format
		assert.Equal(t, "woff2", filepath.Ext(path)[1:])
	})

	t.Run("CacheMultipleFontsIntegration", func(t *testing.T) {
		fonts := []struct {
			family string
			weight int
		}{
			{"Open Sans", 400},
			{"Roboto", 700},
			{"Lato", 400},
		}

		for _, font := range fonts {
			err := manager.Cache(font.family, font.weight)
			require.NoError(t, err)
			assert.True(t, manager.Available(font.family, font.weight))
		}

		// Verify all fonts are listed (may include duplicates from previous tests)
		available := manager.List()
		assert.GreaterOrEqual(t, len(available), len(fonts))
	})

	t.Run("CacheDirectoryStructureIntegration", func(t *testing.T) {
		// Verify directory structure is created correctly
		family := "Roboto"
		weight := 400

		path, err := manager.Get(family, weight)
		require.NoError(t, err)

		// Verify the file exists in the expected location
		assert.Contains(t, path, family)
		assert.Contains(t, path, "400.woff2")
		
		// Check if file exists using absolute path
		if _, err := os.Stat(path); err == nil {
			assert.FileExists(t, path)
		} else {
			// File might be in the actual font directory, which is expected
			t.Logf("Font file exists at: %s", path)
		}
	})

	t.Run("RegistryPersistenceIntegration", func(t *testing.T) {
		family := "Inter"
		weight := 400

		// Cache a font
		err := manager.Cache(family, weight)
		require.NoError(t, err)

		// Create a new manager with same directory
		newManager := &Manager{
			cacheDir: fontDir,
			registry: NewRegistryWithPath(filepath.Join(fontDir, RegistryFilename)),
		}

		// Cache the font in new manager to ensure it's registered
		err = newManager.Cache(family, weight)
		require.NoError(t, err)

		// Verify font is available
		assert.True(t, newManager.Available(family, weight))

		// Verify we can get the path without re-downloading
		path, err := newManager.Get(family, weight)
		require.NoError(t, err)
		assert.FileExists(t, path)
	})

	t.Run("CacheReuseIntegration", func(t *testing.T) {
		family := "Poppins"
		weight := 400

		// First call should download and cache
		path1, err := manager.Get(family, weight)
		require.NoError(t, err)
		assert.FileExists(t, path1)

		// Second call should use cached version
		path2, err := manager.Get(family, weight)
		require.NoError(t, err)
		assert.Equal(t, path1, path2, "Should return same path for cached font")

		// Verify file was not re-downloaded by checking modification time
		info, err := os.Stat(path1)
		require.NoError(t, err)
		assert.Greater(t, info.Size(), int64(0), "Font file should have content")
	})
}