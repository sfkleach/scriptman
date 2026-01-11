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

// InterpreterChoice represents a possible interpreter choice with reasoning.
type InterpreterChoice struct {
	Source         string   // "explicit", "shebang", "extension", "extension-alternatives"
	Interpreter    string   // The interpreter name or path
	Alternatives   []string // Alternative interpreters (for extension-based)
	UseShebang     bool     // If true, use shebang line verbatim
	Reason         string   // Human-readable reason for this choice
	RequiresPrompt bool     // Whether this choice requires user confirmation
}

// DecisionResult contains the interpreter choices determined for a script.
type DecisionResult struct {
	Choices []InterpreterChoice // Possible choices (0=error, 1=automatic, 2=needs prompt)
	Error   error               // Error if no valid choices available
}

// Detect determines the appropriate interpreter for a script.
// Priority:
// 1. Explicit interpreter parameter (if provided)
// 2. Shebang line with consistency checking and user prompting (unless trustShebang is true)
// 3. File extension mapping (checks which alternative exists on PATH)
// 4. Error if none can be determined
func Detect(scriptPath string, scriptContent []byte, explicitInterpreter string, trustShebang bool) (string, string, error) {
	// Get decision result.
	decision := DetermineInterpreterChoices(scriptPath, scriptContent, explicitInterpreter, trustShebang)

	// Handle error case.
	if decision.Error != nil {
		return "", "", decision.Error
	}

	// Handle no choices (shouldn't happen but defensive).
	if len(decision.Choices) == 0 {
		return "", "", fmt.Errorf("internal error: no choices determined for %s", scriptPath)
	}

	// Single choice - automatic decision.
	if len(decision.Choices) == 1 {
		choice := decision.Choices[0]

		// If prompt required (edge case), prompt user.
		if choice.RequiresPrompt {
			approved := promptSingleChoice(choice)
			if !approved {
				return "", "", fmt.Errorf("installation aborted by user")
			}
		}

		return resolveChoice(choice)
	}

	// Multiple choices - need user input.
	selectedChoice := promptMultipleChoices(decision.Choices)
	if selectedChoice == nil {
		return "", "", fmt.Errorf("installation aborted by user")
	}

	return resolveChoice(*selectedChoice)
}

// resolveChoice converts an InterpreterChoice into an actual interpreter path.
func resolveChoice(choice InterpreterChoice) (string, string, error) {
	// Handle explicit interpreter.
	if choice.Source == "explicit" {
		path, err := resolveInterpreter(choice.Interpreter)
		return path, "", err
	}

	// Handle shebang (use verbatim).
	if choice.UseShebang || choice.Source == "shebang" {
		path, err := resolveInterpreter(choice.Interpreter)
		return path, "", err
	}

	// Handle extension-based alternatives.
	if choice.Source == "extension-alternatives" {
		return selectBestAlternative(choice.Alternatives)
	}

	return "", "", fmt.Errorf("internal error: unknown choice source %s", choice.Source)
}

// DetermineInterpreterChoices analyzes a script and returns possible interpreter choices.
// Returns a DecisionResult with:
//   - 0 choices + error: Cannot determine interpreter (error case)
//   - 1 choice: Automatic decision (no prompt needed)
//   - 2 choices: Ambiguous, requires user input
//
// If trustShebang is true, shebang lines are used without consistency checks or prompts.
func DetermineInterpreterChoices(scriptPath string, scriptContent []byte, explicitInterpreter string, trustShebang bool) DecisionResult {
	shebang := parseShebang(scriptContent)
	ext := filepath.Ext(scriptPath)

	// Priority 1: Explicit interpreter always wins (automatic, single choice).
	if explicitInterpreter != "" {
		return DecisionResult{
			Choices: []InterpreterChoice{{
				Source:      "explicit",
				Interpreter: explicitInterpreter,
				Reason:      "Explicitly specified via --interpreter flag",
			}},
		}
	}

	// Priority 2: Shebang exists - complex logic (or trust it directly).
	if shebang != nil {
		if trustShebang {
			// Trust shebang without any checks.
			return DecisionResult{
				Choices: []InterpreterChoice{{
					Source:      "shebang",
					Interpreter: shebang.interpreter,
					UseShebang:  false,
					Reason:      "Trusting shebang via --trust-shebang flag",
				}},
			}
		}
		return determineWithShebang(scriptPath, ext, shebang)
	}

	// Priority 3: Extension mapping only (no shebang).
	if alternatives, ok := ExtensionMap[ext]; ok {
		return DecisionResult{
			Choices: []InterpreterChoice{{
				Source:       "extension-alternatives",
				Alternatives: alternatives,
				Reason:       fmt.Sprintf("Based on file extension %s", ext),
			}},
		}
	}

	// Priority 4: No information available.
	return DecisionResult{
		Error: fmt.Errorf("could not determine interpreter for %s (no --interpreter, no shebang, extension %s not recognized)", scriptPath, ext),
	}
}

// determineWithShebang handles the complex shebang scenarios.
func determineWithShebang(scriptPath string, ext string, shebang *shebangInfo) DecisionResult {
	// Case 1: Shebang has arguments and is NOT using env form.
	// This is potentially dangerous, so offer both options.
	if len(shebang.arguments) > 0 && !shebang.usesEnv {
		extAlternatives, hasExt := ExtensionMap[ext]
		if hasExt {
			// Offer extension-based or shebang with args.
			return DecisionResult{
				Choices: []InterpreterChoice{
					{
						Source:         "extension-alternatives",
						Alternatives:   extAlternatives,
						Reason:         fmt.Sprintf("Use extension-based interpreter without arguments (recommended)"),
						RequiresPrompt: true,
					},
					{
						Source:         "shebang",
						Interpreter:    shebang.interpreter,
						UseShebang:     true,
						Reason:         fmt.Sprintf("Use shebang verbatim: %s (may be system-specific)", shebang.fullLine),
						RequiresPrompt: true,
					},
				},
			}
		}
		// No extension, only shebang available (but requires prompt).
		return DecisionResult{
			Choices: []InterpreterChoice{{
				Source:         "shebang",
				Interpreter:    shebang.interpreter,
				UseShebang:     true,
				Reason:         fmt.Sprintf("Shebang with arguments: %s (requires confirmation)", shebang.fullLine),
				RequiresPrompt: true,
			}},
		}
	}

	// Case 2: No file extension.
	// Use shebang (automatic if env form, else requires prompt).
	if ext == "" {
		return DecisionResult{
			Choices: []InterpreterChoice{{
				Source:         "shebang",
				Interpreter:    shebang.interpreter,
				UseShebang:     false,
				Reason:         fmt.Sprintf("No file extension, using shebang: %s", shebang.fullLine),
				RequiresPrompt: !shebang.usesEnv,
			}},
		}
	}

	// Case 3: Extension not recognized.
	// Use shebang (requires prompt).
	alternatives, hasExtMapping := ExtensionMap[ext]
	if !hasExtMapping {
		return DecisionResult{
			Choices: []InterpreterChoice{{
				Source:         "shebang",
				Interpreter:    shebang.interpreter,
				UseShebang:     false,
				Reason:         fmt.Sprintf("Extension %s not recognized, using shebang: %s", ext, shebang.fullLine),
				RequiresPrompt: true,
			}},
		}
	}

	// Case 4: Check consistency between shebang and extension.
	shebangFamily := getInterpreterFamily(shebang.interpreter)
	consistent := false
	for _, alt := range alternatives {
		if getInterpreterFamily(alt) == shebangFamily {
			consistent = true
			break
		}
	}

	if consistent {
		// Consistent: use extension-based (automatic, single choice).
		return DecisionResult{
			Choices: []InterpreterChoice{{
				Source:       "extension-alternatives",
				Alternatives: alternatives,
				Reason:       fmt.Sprintf("Shebang (%s) consistent with extension %s", shebang.interpreter, ext),
			}},
		}
	}

	// Inconsistent: offer both options.
	return DecisionResult{
		Choices: []InterpreterChoice{
			{
				Source:         "extension-alternatives",
				Alternatives:   alternatives,
				Reason:         fmt.Sprintf("Use extension-based interpreter (recommended)"),
				RequiresPrompt: true,
			},
			{
				Source:         "shebang",
				Interpreter:    shebang.interpreter,
				UseShebang:     false,
				Reason:         fmt.Sprintf("Use shebang interpreter: %s", shebang.fullLine),
				RequiresPrompt: true,
			},
		},
	}
}

// handleShebang processes a shebang line with consistency checking.
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

// promptSingleChoice prompts the user for a single choice that requires confirmation.
func promptSingleChoice(choice InterpreterChoice) bool {
	fmt.Fprintf(os.Stderr, "\n%s\n", choice.Reason)
	fmt.Fprintf(os.Stderr, "Options:\n")
	fmt.Fprintf(os.Stderr, "  1. Proceed\n")
	fmt.Fprintf(os.Stderr, "  2. Abort installation\n")

	selected := promptChoice("[1]", []string{"1", "2"})
	return selected == "1"
}

// promptMultipleChoices prompts the user to select from multiple choices.
// Returns the selected choice or nil if aborted.
func promptMultipleChoices(choices []InterpreterChoice) *InterpreterChoice {
	fmt.Fprintf(os.Stderr, "\nMultiple interpreter options available:\n")
	for i, choice := range choices {
		fmt.Fprintf(os.Stderr, "  %d. %s\n", i+1, choice.Reason)
	}
	fmt.Fprintf(os.Stderr, "  %d. Abort installation\n", len(choices)+1)

	validChoices := make([]string, len(choices)+1)
	for i := 0; i < len(choices); i++ {
		validChoices[i] = fmt.Sprintf("%d", i+1)
	}
	validChoices[len(choices)] = fmt.Sprintf("%d", len(choices)+1)

	selected := promptChoice("[1]", validChoices)
	idx := 0
	fmt.Sscanf(selected, "%d", &idx)

	if idx < 1 || idx > len(choices) {
		return nil // Abort
	}

	return &choices[idx-1]
}

// promptShebangWithArguments is kept for backward compatibility but simplified.
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

// promptNoExtension is kept for backward compatibility but simplified.
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

// promptUnrecognizedExtension is kept for backward compatibility but simplified.
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

// promptInconsistent is kept for backward compatibility but simplified.
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
