package interpreter

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// ExtensionMap maps file extensions to alternative interpreters in priority order.
var ExtensionMap = map[string][]string{
	".py":   {"python3", "python"},
	".rb":   {"ruby"},
	".pl":   {"perl"},
	".sh":   {"sh"},
	".bash": {"bash"},
	".zsh":  {"zsh"},
	".js":   {"node"},
	".lua":  {"lua"},
	".php":  {"php"},
}

// interpreterFamilies maps base interpreter names to their family.
// Used for consistency checking between shebang and extension.
var interpreterFamilies = map[string]string{
	"python":  "python",
	"python2": "python",
	"python3": "python",
	"ruby":    "ruby",
	"ruby2":   "ruby",
	"ruby3":   "ruby",
	"perl":    "perl",
	"perl5":   "perl",
	"sh":      "shell",
	"bash":    "shell",
	"dash":    "shell",
	"zsh":     "shell",
	"ksh":     "shell",
	"node":    "javascript",
	"nodejs":  "javascript",
	"lua":     "lua",
	"php":     "php",
}

// shebangInfo contains parsed shebang information.
type shebangInfo struct {
	interpreter string   // The interpreter name (e.g., "python3")
	arguments   []string // Any arguments passed to the interpreter
	usesEnv     bool     // Whether it uses #!/usr/bin/env form
	fullLine    string   // The complete shebang line for reference
}

// Detect determines the appropriate interpreter for a script.
// Priority:
// 1. Explicit interpreter parameter (if provided)
// 2. Shebang line with consistency checking and user prompting
// 3. File extension mapping (checks which alternative exists on PATH)
// 4. Error if none can be determined
func Detect(scriptPath string, scriptContent []byte, explicitInterpreter string) (string, string, error) {
	// Priority 1: Explicit interpreter.
	if explicitInterpreter != "" {
		path, err := resolveInterpreter(explicitInterpreter)
		return path, "", err
	}

	// Priority 2: Shebang line.
	if shebang := parseShebang(scriptContent); shebang != nil {
		return handleShebang(scriptPath, shebang)
	}

	// Priority 3: File extension.
	ext := filepath.Ext(scriptPath)
	if alternatives, ok := ExtensionMap[ext]; ok {
		path, warning, err := selectBestAlternative(alternatives)
		return path, warning, err
	}

	return "", "", fmt.Errorf("could not determine interpreter for %s (no --interpreter, no shebang, extension %s not recognized)", scriptPath, ext)
}

// handleShebang processes a shebang line with consistency checking.
func handleShebang(scriptPath string, shebang *shebangInfo) (string, string, error) {
	// Check if shebang has arguments and is not using env form.
	if len(shebang.arguments) > 0 && !shebang.usesEnv {
		// Prompt user for approval to use shebang with arguments.
		approved, useShebang := promptShebangWithArguments(shebang)
		if !approved {
			return "", "", fmt.Errorf("installation aborted by user")
		}
		if useShebang {
			// User wants to use the shebang verbatim.
			return shebang.interpreter, "", nil
		}
		// User declined, fall through to extension mapping.
	}

	// Get file extension.
	ext := filepath.Ext(scriptPath)

	// No extension: ask user for permission to copy shebang.
	if ext == "" {
		approved, useShebang := promptNoExtension(shebang)
		if !approved {
			return "", "", fmt.Errorf("installation aborted by user")
		}
		if useShebang {
			return shebang.interpreter, "", nil
		}
		return "", "", fmt.Errorf("no extension and user declined to use shebang")
	}

	// Check extension consistency with shebang.
	alternatives, hasExtMapping := ExtensionMap[ext]
	if !hasExtMapping {
		// Extension not recognized, ask user about using shebang.
		approved, useShebang := promptUnrecognizedExtension(scriptPath, shebang)
		if !approved {
			return "", "", fmt.Errorf("installation aborted by user")
		}
		if useShebang {
			return shebang.interpreter, "", nil
		}
		return "", "", fmt.Errorf("extension %s not recognized and user declined to use shebang", ext)
	}

	// Check if shebang interpreter is consistent with extension alternatives.
	shebangFamily := getInterpreterFamily(shebang.interpreter)
	consistent := false
	for _, alt := range alternatives {
		if getInterpreterFamily(alt) == shebangFamily {
			consistent = true
			break
		}
	}

	if !consistent {
		// Inconsistent: prompt user to choose.
		approved, useShebang := promptInconsistent(scriptPath, shebang, alternatives)
		if !approved {
			return "", "", fmt.Errorf("installation aborted by user")
		}
		if useShebang {
			// Try to resolve the shebang interpreter.
			path, err := resolveInterpreter(shebang.interpreter)
			if err != nil {
				return "", "", fmt.Errorf("shebang interpreter %s: %w", shebang.interpreter, err)
			}
			return path, "", nil
		}
		// Use extension mapping.
	}

	// Consistent or user chose extension: use our configured interpreter.
	path, warning, err := selectBestAlternative(alternatives)
	return path, warning, err
}

// selectBestAlternative finds the first interpreter from alternatives that exists on PATH.
// Returns the resolved path, an optional warning, and an error.
func selectBestAlternative(alternatives []string) (string, string, error) {
	if len(alternatives) == 0 {
		return "", "", fmt.Errorf("no interpreter alternatives provided")
	}

	// Try to find first alternative that exists.
	for _, alt := range alternatives {
		if path, err := exec.LookPath(alt); err == nil {
			return path, "", nil
		}
	}

	// None found, use the first one but generate a warning.
	first := alternatives[0]
	path, err := resolveInterpreterWithoutCheck(first)
	if err != nil {
		return "", "", err
	}

	warning := fmt.Sprintf("Warning: '%s' not found on PATH (tried: %s)", first, strings.Join(alternatives, ", "))
	return path, warning, nil
}

// parseShebang extracts the interpreter and arguments from a shebang line.
// Handles both direct paths and "#!/usr/bin/env interpreter" forms.
func parseShebang(content []byte) *shebangInfo {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	if !scanner.Scan() {
		return nil
	}

	line := strings.TrimSpace(scanner.Text())
	if !strings.HasPrefix(line, "#!") {
		return nil
	}

	// Remove the "#!" prefix.
	fullLine := line
	line = strings.TrimSpace(line[2:])

	info := &shebangInfo{
		fullLine: fullLine,
	}

	// Handle "/usr/bin/env interpreter [args...]" form.
	if strings.Contains(line, "/env") {
		parts := strings.Fields(line)
		for i, part := range parts {
			if strings.HasSuffix(part, "/env") && i+1 < len(parts) {
				info.usesEnv = true
				info.interpreter = parts[i+1]
				if i+2 < len(parts) {
					info.arguments = parts[i+2:]
				}
				return info
			}
		}
	}

	// Handle direct path form: extract basename and arguments.
	parts := strings.Fields(line)
	if len(parts) > 0 {
		info.interpreter = filepath.Base(parts[0])
		if len(parts) > 1 {
			info.arguments = parts[1:]
		}
		return info
	}

	return nil
}

// getInterpreterFamily returns the family name for an interpreter.
// Strips version numbers and normalizes names.
func getInterpreterFamily(interpreter string) string {
	// Strip version numbers (python3.11 -> python3).
	base := regexp.MustCompile(`\d+\.\d+$`).ReplaceAllString(interpreter, "")

	// Look up in family map.
	if family, ok := interpreterFamilies[base]; ok {
		return family
	}

	// Fall back to the interpreter name itself as the family.
	return base
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

// resolveInterpreterWithoutCheck constructs an interpreter path without checking if it exists.
// This is used when no alternatives are found on PATH.
func resolveInterpreterWithoutCheck(name string) (string, error) {
	// If it's already an absolute path, return as-is.
	if filepath.IsAbs(name) {
		return name, nil
	}

	// Construct a likely path based on common locations.
	// For most Unix systems, interpreters are in /usr/bin.
	return filepath.Join("/usr/bin", name), nil
}

// promptShebangWithArguments prompts the user when a shebang has arguments.
// Returns (approved, useShebang) where approved=false means abort.
func promptShebangWithArguments(shebang *shebangInfo) (bool, bool) {
	fmt.Fprintf(os.Stderr, "\nScript has shebang: %s\n", shebang.fullLine)
	fmt.Fprintf(os.Stderr, "This uses interpreter arguments: %s\n", strings.Join(shebang.arguments, " "))
	fmt.Fprintf(os.Stderr, "Options:\n")
	fmt.Fprintf(os.Stderr, "  1. Use our configured interpreter without arguments (recommended)\n")
	fmt.Fprintf(os.Stderr, "  2. Copy shebang verbatim (may be system-specific)\n")
	fmt.Fprintf(os.Stderr, "  3. Abort installation\n")

	choice := promptChoice("[1]", []string{"1", "2", "3"})
	switch choice {
	case "1":
		return true, false // Use our interpreter
	case "2":
		return true, true // Use shebang
	case "3":
		return false, false // Abort
	default:
		return true, false // Default to option 1
	}
}

// promptNoExtension prompts when there's a shebang but no file extension.
func promptNoExtension(shebang *shebangInfo) (bool, bool) {
	fmt.Fprintf(os.Stderr, "\nScript has no file extension.\n")
	fmt.Fprintf(os.Stderr, "Shebang line: %s\n", shebang.fullLine)
	fmt.Fprintf(os.Stderr, "Options:\n")
	fmt.Fprintf(os.Stderr, "  1. Use shebang interpreter (recommended)\n")
	fmt.Fprintf(os.Stderr, "  2. Abort installation\n")

	choice := promptChoice("[1]", []string{"1", "2"})
	switch choice {
	case "1":
		return true, true // Use shebang
	case "2":
		return false, false // Abort
	default:
		return true, true // Default to option 1
	}
}

// promptUnrecognizedExtension prompts when extension is not in our map.
func promptUnrecognizedExtension(scriptPath string, shebang *shebangInfo) (bool, bool) {
	ext := filepath.Ext(scriptPath)
	fmt.Fprintf(os.Stderr, "\nFile extension %s is not recognized.\n", ext)
	fmt.Fprintf(os.Stderr, "Shebang line: %s\n", shebang.fullLine)
	fmt.Fprintf(os.Stderr, "Options:\n")
	fmt.Fprintf(os.Stderr, "  1. Use shebang interpreter (recommended)\n")
	fmt.Fprintf(os.Stderr, "  2. Abort installation\n")

	choice := promptChoice("[1]", []string{"1", "2"})
	switch choice {
	case "1":
		return true, true // Use shebang
	case "2":
		return false, false // Abort
	default:
		return true, true // Default to option 1
	}
}

// promptInconsistent prompts when shebang and extension suggest different interpreters.
func promptInconsistent(scriptPath string, shebang *shebangInfo, alternatives []string) (bool, bool) {
	ext := filepath.Ext(scriptPath)
	fmt.Fprintf(os.Stderr, "\nInterpreter mismatch detected:\n")
	fmt.Fprintf(os.Stderr, "  Shebang: %s\n", shebang.fullLine)
	fmt.Fprintf(os.Stderr, "  Extension %s suggests: %s\n", ext, strings.Join(alternatives, " or "))
	fmt.Fprintf(os.Stderr, "Options:\n")
	fmt.Fprintf(os.Stderr, "  1. Use extension-based interpreter (recommended)\n")
	fmt.Fprintf(os.Stderr, "  2. Use shebang interpreter\n")
	fmt.Fprintf(os.Stderr, "  3. Abort installation\n")

	choice := promptChoice("[1]", []string{"1", "2", "3"})
	switch choice {
	case "1":
		return true, false // Use extension
	case "2":
		return true, true // Use shebang
	case "3":
		return false, false // Abort
	default:
		return true, false // Default to option 1
	}
}

// promptChoice displays a prompt and reads user input.
// defaultChoice is shown in square brackets and returned if user presses Enter.
func promptChoice(defaultPrompt string, validChoices []string) string {
	fmt.Fprintf(os.Stderr, "Choice %s: ", defaultPrompt)

	var input string
	fmt.Fscanln(os.Stdin, &input)
	input = strings.TrimSpace(input)

	// If empty input, extract default from prompt like "[1]".
	if input == "" {
		input = strings.Trim(defaultPrompt, "[]")
	}

	// Validate choice.
	for _, valid := range validChoices {
		if input == valid {
			return input
		}
	}

	// Invalid choice, return first valid option as fallback.
	if len(validChoices) > 0 {
		return validChoices[0]
	}

	return input
}
