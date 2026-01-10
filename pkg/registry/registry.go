package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SchemaVersion is the current registry schema version.
const SchemaVersion = 1

// Registry represents the scriptman registry structure.
type Registry struct {
	SchemaVersion int                `json:"schema_version"`
	Scripts       map[string]*Script `json:"scripts"`
}

// Script represents an installed script entry in the registry.
type Script struct {
	Repo        string    `json:"repo"`
	SourcePath  string    `json:"source_path"`
	LocalScript string    `json:"local_script"`
	Interpreter string    `json:"interpreter"`
	WrapperPath string    `json:"wrapper_path"`
	InstalledAt time.Time `json:"installed_at"`
	Version     string    `json:"version,omitempty"`
}

// Load reads the registry from disk. Returns an empty registry if file doesn't exist.
func Load(path string) (*Registry, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Registry{
			SchemaVersion: SchemaVersion,
			Scripts:       make(map[string]*Script),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read registry: %w", err)
	}

	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("failed to parse registry: %w", err)
	}

	// Defensive check: ensure Scripts map is not nil.
	if reg.Scripts == nil {
		reg.Scripts = make(map[string]*Script)
	}

	return &reg, nil
}

// Save writes the registry to disk.
func (r *Registry) Save(path string) error {
	// Ensure parent directory exists.
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create registry directory: %w", err)
	}

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	// Write with proper permissions.
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write registry: %w", err)
	}

	return nil
}

// Add adds or updates a script in the registry.
func (r *Registry) Add(name string, script *Script) {
	if r.Scripts == nil {
		r.Scripts = make(map[string]*Script)
	}
	r.Scripts[name] = script
}

// Remove removes a script from the registry.
func (r *Registry) Remove(name string) error {
	if _, exists := r.Scripts[name]; !exists {
		return fmt.Errorf("script '%s' not found in registry", name)
	}
	delete(r.Scripts, name)
	return nil
}

// Get retrieves a script from the registry.
func (r *Registry) Get(name string) (*Script, error) {
	script, exists := r.Scripts[name]
	if !exists {
		return nil, fmt.Errorf("script '%s' not found in registry", name)
	}
	return script, nil
}

// Exists checks if a script name is registered.
func (r *Registry) Exists(name string) bool {
	_, exists := r.Scripts[name]
	return exists
}
