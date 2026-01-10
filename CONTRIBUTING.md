# Contributing to Scriptman

Thank you for your interest in contributing to Scriptman! This document provides guidelines and instructions for contributing to the project.

## Development Setup

### Prerequisites

- Go 1.24.2 or later
- Git
- [Just](https://github.com/casey/just) command runner (optional but recommended)

### Getting Started

1. Fork the repository on GitHub
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR-USERNAME/scriptman.git
   cd scriptman
   ```
3. Build the project:
   ```bash
   just build
   # or without just:
   go build -o bin/scriptman ./cmd/scriptman
   ```
4. Run tests:
   ```bash
   just test
   # or without just:
   go test ./...
   ```

### Project Structure

```
scriptman/
├── cmd/scriptman/          # Main entry point
├── pkg/                    # Library code
├── docs/                   # Documentation
│   ├── decisions/          # Design decision records
│   └── tasks/              # Task documentation
└── .github/                # GitHub-specific files
```

## Code Style Guidelines

We follow the conventions outlined in `.github/copilot-instructions.md`. Key points:

### General Guidelines

- **Be objective and critical**: Focus on technical correctness over agreeability
- **Challenge assumptions**: If code has clear technical flaws, point them out directly
- **Think through implications**: Consider how users will actually use features in practice

### Code Formatting

- **Line endings**: Use LF (not CRLF) for all text files
- **Line termination**: Files should end with a newline character
- **No trailing whitespace**: Except in Markdown where a single space indicates a line break
- **Maximum line length**: 120 characters (80 for code + 40 for indentation)
- **Indentation**: Use tabs for Go code (as required by Go), 4 spaces for other languages
- **Encoding**: UTF-8 for all text files

### Go-Specific Guidelines

- Follow standard Go conventions (`go fmt`, `go vet`)
- Comments should be complete sentences with proper capitalization and punctuation
- Add defensive checks with explanatory comments when appropriate
- Use meaningful variable and function names

### File Paths

- Use Unix-style paths (forward slashes) in code and documentation, even on Windows
- Use `filepath` package functions for cross-platform compatibility

## Testing Requirements

All contributions must include appropriate tests:

### Running Tests

```bash
just test
```

This command runs:
1. Unit tests (`go test ./...`)
2. Code formatting checks
3. Static analysis (errcheck, staticcheck)
4. Security scanner (gosec)

All checks must pass before submitting a pull request.

### Writing Tests

- Tests should use real filesystem operations via `t.TempDir()` for integration-style testing
- Use function variables for dependency injection in tests
- Each test should be independent and not rely on global state
- Test files should follow the pattern `*_test.go`

Example test structure:
```go
func TestFeature(t *testing.T) {
    // Arrange
    tmpDir := t.TempDir()
    
    // Act
    result, err := Feature(tmpDir)
    
    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

## Making Changes

### Testing Commands

When testing the behavior of scriptman, always use:
```bash
go run ./cmd/scriptman
```

Rather than `./scriptman` directly. This ensures you're testing the latest code rather than an out-of-date compiled binary.

### Artifacts

- Do not create artifacts within the repository folder structure
- EXCEPT in folders starting with an underscore, such as `_build/`

### Design Decisions

Design decisions should be documented in the `docs/decisions/` folder using the established template. However, you don't need to precisely follow the template. In some cases there were not multiple options considered so the pros-and-cons section may be omitted and we simply document the reasoning behind the decision.

### Branch Naming

Use descriptive branch names:
- `feature/add-version-command`
- `fix/broken-wrapper-generation`
- `docs/improve-readme`

### Commit Messages

Follow conventional commit format:
```
type(scope): brief description

Longer explanation if needed.

Fixes #123
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `chore`: Maintenance tasks

### Pull Request Process

1. **Create an issue first** (for significant changes) to discuss the approach
2. **Write tests** for your changes
3. **Update documentation** if you're changing functionality
4. **Run all checks**: `just test` must pass
5. **Keep PRs focused**: One feature or fix per PR
6. **Write clear PR descriptions**:
   - What problem does this solve?
   - How does it solve it?
   - Are there any trade-offs or limitations?

### Code Review

- Be open to feedback and suggestions
- Respond to review comments promptly
- Make requested changes or explain why you disagree
- Keep discussions respectful and technical

## Issue Guidelines

### Reporting Bugs

Include:
- Scriptman version (`scriptman --version` or git commit)
- Operating system and version
- Shell (bash/zsh/fish) and version
- Steps to reproduce
- Expected behavior vs actual behavior
- Relevant error messages or logs

### Requesting Features

Include:
- Use case: Why is this feature needed?
- Proposed solution: How should it work?
- Alternatives considered: What other approaches did you think about?
- Example usage: What would the commands look like?

## Development Tools

### Just Commands

If you have [Just](https://github.com/casey/just) installed:

```bash
just build          # Build the binary
just test           # Run all tests and checks
just fmt            # Format code
just fmt-check      # Check formatting without changing files
just clean          # Remove build artifacts
```

### Without Just

Equivalent commands:

```bash
# Build
mkdir -p bin
go build -o bin/scriptman ./cmd/scriptman

# Test
go test ./...

# Format
go fmt ./...

# Format check
gofmt -l .

# Clean
rm -rf bin/
```

## Security

### Reporting Security Issues

**Do not open public issues for security vulnerabilities.**

Instead, email security concerns to the project maintainer. Include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

### Security Considerations in Code

- Validate all user input
- Use appropriate file permissions (0755 for directories, 0755 for executables, 0644 for files)
- Add `#nosec` comments with justifications for intentional security exceptions
- Never trust data from GitHub repositories without validation
- Consider security implications of executing downloaded scripts

## Getting Help

- Check existing documentation in `docs/` and `README.md`
- Search existing issues and pull requests
- Open a new issue with the `question` label
- Be patient and respectful when asking for help

## License

By contributing to Scriptman, you agree that your contributions will be licensed under the GNU General Public License v3.0 (GPL-3.0).

## Recognition

Contributors will be recognized in the project. Significant contributors may be added to a CONTRIBUTORS file.

Thank you for helping make Scriptman better!
