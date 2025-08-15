package font

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joeblew999/infra/pkg/config"
)

// Registry manages font metadata and lookups
type Registry struct {
	path string
	data map[string]FontInfo
}

// NewRegistry creates a new font registry
func NewRegistry() *Registry {
	registryPath := filepath.Join(config.GetFontPath(), RegistryFilename)
	r := &Registry{
		path: registryPath,
		data: make(map[string]FontInfo),
	}
	r.load()
	return r
}

// GetPath returns the path for a font if it exists
func (r *Registry) GetPath(font Font) (string, bool) {
	key := r.key(font)
	info, exists := r.data[key]
	if !exists {
		return "", false
	}
	
	// Verify file still exists
	if _, err := os.Stat(info.Path); os.IsNotExist(err) {
		delete(r.data, key)
		r.save()
		return "", false
	}
	
	return info.Path, true
}

// Add registers a font in the registry
func (r *Registry) Add(info FontInfo) error {
	key := r.key(info.Font)
	r.data[key] = info
	return r.save()
}

// List returns all registered fonts
func (r *Registry) List() []FontInfo {
	fonts := make([]FontInfo, 0, len(r.data))
	for _, info := range r.data {
		fonts = append(fonts, info)
	}
	return fonts
}

// Remove removes a font from the registry
func (r *Registry) Remove(font Font) error {
	key := r.key(font)
	delete(r.data, key)
	return r.save()
}

// key generates a unique key for a font
func (r *Registry) key(font Font) string {
	return fmt.Sprintf("%s-%d-%s-%s", 
		strings.ToLower(font.Family), 
		font.Weight, 
		font.Style, 
		font.Format)
}

// load reads the registry from disk
func (r *Registry) load() {
	if _, err := os.Stat(r.path); os.IsNotExist(err) {
		return // Registry doesn't exist yet
	}
	
	data, err := os.ReadFile(r.path)
	if err != nil {
		return // Failed to read, start fresh
	}
	
	var registryData map[string]FontInfo
	if err := json.Unmarshal(data, &registryData); err != nil {
		return // Failed to parse, start fresh
	}
	
	r.data = registryData
}

// save writes the registry to disk
func (r *Registry) save() error {
	// Ensure directory exists
	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create font directory: %w", err)
	}
	
	data, err := json.MarshalIndent(r.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}
	
	return os.WriteFile(r.path, data, 0644)
}