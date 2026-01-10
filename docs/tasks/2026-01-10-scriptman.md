# Initial design for the scriptman utility

## Summary

Design a companion tool to execman that manages single-file scripts downloaded
from GitHub repositories. The tool should wrap scripts with executables,
providing central management, updates, and clean removal.

## Status

- [x] Create design document

## Background

Execman manages standalone executables from GitHub releases. However, many
useful tools are distributed as single-file scripts (Python, Ruby, Bash, etc.)
that don't have pre-built binaries. Scriptman fills this gap.

## Requirements

### Functional Requirements

1. Install scripts from GitHub repositories (release assets or source archives)
2. Wrap scripts with executables that invoke the correct interpreter
3. Support name mapping (install `long-script-name.py` as `tool`)
4. Track installed scripts in a registry for management
5. Provide update, remove, list, and check commands
6. Detect interpreter from shebang or file extension

### Non-Functional Requirements

1. Unix-only (Linux, macOS, BSD) - no Windows support
2. No runtime registry lookups - paths baked into wrappers
3. Work without requiring a compiler (fallback to shell scripts)
4. Use execman as a library for shared infrastructure

## Design Decisions

### Hybrid Wrapper Strategy

Two strategies depending on environment:

| Strategy | When Used | How It Works |
|----------|-----------|--------------|
| Compiled C | `cc` available | Generate and compile tiny C binary with paths baked in |
| Shell + Hardlink | No compiler | Generate shell script + hardlink to scriptman binary |

The compiled C approach has zero runtime overhead. The shell script approach
uses POSIX-guaranteed `/bin/sh` as fallback.

### Registry for Management Only

The registry (`~/.config/scriptman/registry.json`) is used only for management
commands (list, check, remove, update). Runtime dispatch never reads the
registry - paths are baked into the wrapper (compiled binary or shell script).

### Command Structure

Primary command:

```bash
scriptman install REPO PATH [--interpreter CMD] [--name NAME] [--into DIR]
```

Other commands: `list`, `remove`, `check`, `update`

### Script Storage

Downloaded scripts stored in `~/.local/share/scriptman/scripts/` rather than
directly in PATH. Wrappers in PATH point to these stored scripts.

## Implementation Plan

The implementation will be in a separate repository (`sfkleach/scriptman`)
using execman as a Go module dependency.

### Phase 1: Core Infrastructure

1. Project setup (Go module, Cobra CLI)
2. Registry management (adapted from execman)
3. Configuration loading
4. GitHub API integration (via execman library)

### Phase 2: Install Command

1. Script retrieval from GitHub releases/source
2. Interpreter detection (shebang, extension mapping)
3. C code generation and compilation
4. Shell script + hardlink fallback
5. Registry update

### Phase 3: Dual-Role Binary

1. Detection logic (`os.Args[0]` check)
2. Runner mode (find and exec companion shell script)
3. Management mode (Cobra CLI)

### Phase 4: Management Commands

1. `list` - Display installed scripts
2. `remove` - Remove wrapper, script, and registry entry
3. `check` - Verify all scripts and interpreters exist
4. `update` - Re-fetch from GitHub and recreate wrapper

### Phase 5: Polish

1. Error handling and user feedback
2. Cross-filesystem detection (copy instead of hardlink)
3. Documentation and README
4. Release via GitHub Actions

## References

- Design document: [2026-01-10-scriptman-design.md](../decisions/2026-01-10-scriptman-design.md)
