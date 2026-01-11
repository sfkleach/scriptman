package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Release represents a GitHub release.
type Release struct {
	TagName string `json:"tag_name"`
}

// FetchResult contains the fetched script content and metadata.
type FetchResult struct {
	Content []byte
	Tag     string // Release tag, empty if fetched from main branch
}

// GetLatestRelease fetches the latest release tag for a repository.
// Returns empty string if no releases exist.
func GetLatestRelease(repo string) (string, error) {
	repo = normalizeRepo(repo)

	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

	// #nosec G107 -- Fetching from GitHub API is the core feature of this tool.
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to query releases: %w", err)
	}
	defer resp.Body.Close()

	// No releases found - not an error, just no releases.
	if resp.StatusCode == http.StatusNotFound {
		return "", nil
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to query releases: HTTP %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to parse release response: %w", err)
	}

	return release.TagName, nil
}

// FetchScript downloads a script from a GitHub repository.
// If tag is empty, fetches from the main branch.
// Returns the content along with metadata about what was fetched.
func FetchScript(repo, path, tag string) (*FetchResult, error) {
	repo = normalizeRepo(repo)

	// Use tag if provided, otherwise fall back to main.
	ref := "main"
	if tag != "" {
		ref = tag
	}

	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s", repo, ref, path)

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

	return &FetchResult{
		Content: data,
		Tag:     tag,
	}, nil
}

// normalizeRepo strips common GitHub URL prefixes.
func normalizeRepo(repo string) string {
	repo = strings.TrimPrefix(repo, "github.com/")
	repo = strings.TrimPrefix(repo, "https://github.com/")
	return repo
}

// SaveScript saves script content to a file with proper permissions.
func SaveScript(content []byte, destPath string, perm os.FileMode) error {
	// Ensure parent directory exists.
	// #nosec G301 -- Standard directory permissions (0755) for script storage directory.
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create script directory: %w", err)
	}

	// Write script file with configurable permissions.
	if err := os.WriteFile(destPath, content, perm); err != nil {
		return fmt.Errorf("failed to write script file: %w", err)
	}

	return nil
}
