package github

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FetchScript downloads a script from a GitHub repository.
// For now, this is a simplified implementation that fetches from the main branch.
// TODO: Support release assets and version tags.
func FetchScript(repo, path string) ([]byte, error) {
	// Normalize repo format (strip github.com/ prefix if present).
	repo = strings.TrimPrefix(repo, "github.com/")
	repo = strings.TrimPrefix(repo, "https://github.com/")

	// For now, fetch from main branch (simplified version).
	// TODO: Implement proper release asset checking and source archive extraction.
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/main/%s", repo, path)

	// #nosec G107 -- Fetching from user-specified GitHub URLs is the core feature of this tool.
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch script: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch script: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read script content: %w", err)
	}

	return data, nil
}

// SaveScript saves script content to a file with proper permissions.
func SaveScript(content []byte, destPath string, perm os.FileMode) error {
	// Ensure parent directory exists.
	// #nosec G301 -- Standard directory permissions (0755) for script storage directory.
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create script directory: %w", err)
	}

	// Write script file with read/write permissions for owner, read for others.
	if err := os.WriteFile(destPath, content, perm); err != nil {
		return fmt.Errorf("failed to write script file: %w", err)
	}

	return nil
}
