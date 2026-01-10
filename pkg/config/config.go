package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the scriptman configuration.
type Config struct {
	BinDir    string `json:"bin_dir"`
	ScriptDir string `json:"script_dir"`
}

// Load reads the configuration from disk. Returns default config if file doesn't exist.
func Load() (*Config, error) {
	configPath := GetConfigPath()

	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		// Return default configuration.
		return GetDefaultConfig()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Defensive check: ensure paths are not empty, use defaults if needed.
	if cfg.BinDir == "" {
		binDir, err := getDefaultBinDir()
		if err != nil {
			return nil, err
		}
		cfg.BinDir = binDir
	}
	if cfg.ScriptDir == "" {
		scriptDir, err := getDefaultScriptDir()
		if err != nil {
			return nil, err
		}
		cfg.ScriptDir = scriptDir
	}

	return &cfg, nil
}

// Save writes the configuration to disk.
func (c *Config) Save() error {
	configPath := GetConfigPath()

	// Ensure parent directory exists.
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write with proper permissions.
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// GetDefaultConfig returns a configuration with default values.
func GetDefaultConfig() (*Config, error) {
	binDir, err := getDefaultBinDir()
	if err != nil {
		return nil, err
	}
	scriptDir, err := getDefaultScriptDir()
	if err != nil {
		return nil, err
	}
	return &Config{
		BinDir:    binDir,
		ScriptDir: scriptDir,
	}, nil
}

// GetConfigPath returns the path to the configuration file.
func GetConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".scriptman", "config.json")
	}
	return filepath.Join(home, ".config", "scriptman", "config.json")
}

// getDefaultBinDir returns the default directory for wrappers.
func getDefaultBinDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".local", "bin"), nil
}

// getDefaultScriptDir returns the default directory for downloaded scripts.
func getDefaultScriptDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".local", "share", "scriptman", "scripts"), nil
}
