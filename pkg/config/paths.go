package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetDefaultBinDir returns the default directory for wrappers from config.
func GetDefaultBinDir() (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}
	return cfg.BinDir, nil
}

// GetDefaultScriptDir returns the default directory for downloaded scripts from config.
func GetDefaultScriptDir() (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}
	return cfg.ScriptDir, nil
}

// GetDefaultRegistryPath returns the default path for the registry file.
func GetDefaultRegistryPath() string {
	// Registry is always in ~/.config/scriptman/registry.json
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".scriptman", "registry.json")
	}
	return filepath.Join(home, ".config", "scriptman", "registry.json")
}
