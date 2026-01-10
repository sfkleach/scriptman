package wrapper

import (
	"fmt"
	"os"
	"path/filepath"
)

// CreateWrapper creates a shell script wrapper that executes the given script.
func CreateWrapper(interpreterPath, scriptPath, wrapperPath string) error {
	// Generate shell script with baked-in paths.
	shellScript := fmt.Sprintf("#!/bin/sh\nexec %s %s \"$@\"\n", interpreterPath, scriptPath)

	// Ensure parent directory exists.
	if err := os.MkdirAll(filepath.Dir(wrapperPath), 0755); err != nil {
		return fmt.Errorf("failed to create wrapper directory: %w", err)
	}

	// Write shell script wrapper.
	if err := os.WriteFile(wrapperPath, []byte(shellScript), 0755); err != nil {
		return fmt.Errorf("failed to write wrapper: %w", err)
	}

	return nil
}

// Remove removes a wrapper script.
func Remove(wrapperPath string) error {
	if err := os.Remove(wrapperPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove wrapper: %w", err)
	}
	return nil
}
