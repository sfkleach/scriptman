package interpreter

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// ExtensionMap maps file extensions to default interpreters.
var ExtensionMap = map[string]string{
	".py":   "python3",
	".rb":   "ruby",
	".pl":   "perl",
	".sh":   "sh",
	".bash": "bash",
	".zsh":  "zsh",
	".js":   "node",
	".lua":  "lua",
	".php":  "php",
}

// Detect determines the appropriate interpreter for a script.
// Priority:
// 1. Explicit interpreter parameter (if provided)
// 2. Shebang line in script content
// 3. File extension mapping
// 4. Error if none can be determined
func Detect(scriptPath string, scriptContent []byte, explicitInterpreter string) (string, error) {
	// Priority 1: Explicit interpreter.
	if explicitInterpreter != "" {
		return resolveInterpreter(explicitInterpreter)
	}

	// Priority 2: Shebang line.
	if interp := parseShebang(scriptContent); interp != "" {
		return resolveInterpreter(interp)
	}

	// Priority 3: File extension.
	ext := filepath.Ext(scriptPath)
	if interp, ok := ExtensionMap[ext]; ok {
		return resolveInterpreter(interp)
	}

	return "", fmt.Errorf("could not determine interpreter for %s (no --interpreter, no shebang, extension %s not recognized)", scriptPath, ext)
}

// parseShebang extracts the interpreter from a shebang line.
// Handles both direct paths and "#!/usr/bin/env interpreter" forms.
func parseShebang(content []byte) string {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	if !scanner.Scan() {
		return ""
	}

	line := strings.TrimSpace(scanner.Text())
	if !strings.HasPrefix(line, "#!") {
		return ""
	}

	// Remove the "#!" prefix.
	line = strings.TrimSpace(line[2:])

	// Handle "/usr/bin/env interpreter" form.
	if strings.Contains(line, "/env ") {
		parts := strings.Fields(line)
		for i, part := range parts {
			if strings.HasSuffix(part, "/env") && i+1 < len(parts) {
				return parts[i+1]
			}
		}
	}

	// Handle direct path form: extract basename.
	parts := strings.Fields(line)
	if len(parts) > 0 {
		return filepath.Base(parts[0])
	}

	return ""
}

// resolveInterpreter resolves an interpreter name to its full path.
func resolveInterpreter(name string) (string, error) {
	// If it's already an absolute path, verify it exists.
	if filepath.IsAbs(name) {
		return name, nil
	}

	// Look up in PATH.
	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("interpreter '%s' not found in PATH", name)
	}

	return path, nil
}
