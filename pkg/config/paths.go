package config

import (
	"os"
	"path/filepath"
)

// GetDefaultBinDir returns the default directory for wrappers.
func GetDefaultBinDir() string {
	// Use ~/.local/bin as default.
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", "bin")
	}
	return filepath.Join(home, ".local", "bin")
}

// GetDefaultScriptDir returns the default directory for downloaded scripts.
func GetDefaultScriptDir() string {
	// Use ~/.local/share/scriptman/scripts as default.
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".scriptman", "scripts")
	}
	return filepath.Join(home, ".local", "share", "scriptman", "scripts")
}

// GetDefaultRegistryPath returns the default path for the registry file.
func GetDefaultRegistryPath() string {
	// Use ~/.config/scriptman/registry.json as default.
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".scriptman", "registry.json")
	}
	return filepath.Join(home, ".config", "scriptman", "registry.json")
}
