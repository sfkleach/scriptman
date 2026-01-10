# Scriptman Design - Script Wrapper Tool, 2026-01-10

## Overview

Scriptman is a companion tool to execman that manages single-file scripts
downloaded from GitHub repositories.

**Platform scope**: Unix-only (Linux, macOS, BSD). No Windows support.

## Value Proposition

Scriptman provides for scripts what execman provides for binaries:

- **Central management** - Track installed scripts and their origins
- **Updates** - Re-fetch latest version from GitHub with one command
- **Clean removal** - Remove script and wrapper together
- **Name mapping** - Install `owner/repo/scripts/long-name.py` as just `tool`
- **Interpreter binding** - Specify interpreter at install time, not in shebang

## Relationship to Execman

| Tool | Purpose |
|------|---------|
| **execman** | Manages standalone executables downloaded from GitHub releases |
| **scriptman** | Manages scripts downloaded from GitHub, wrapped as executables |

Scriptman will use execman as a library, sharing infrastructure such as:
- GitHub API interaction
- Registry management patterns
- Configuration loading

The tools are separate repositories with separate binaries, keeping each
focused on a single responsibility.

## Core Command

```bash
scriptman install REPO PATH [--interpreter COMMAND] [--name NAME] [--into DIR]
```

- `REPO`: GitHub repository (e.g., `owner/repo` or `github.com/owner/repo`)
- `PATH`: Path to the script within the release or source archive
- `--interpreter`: Explicit interpreter command (guessed from extension if omitted)
- `--name`: Name for the wrapper (defaults to script filename without extension)
- `--into`: Target directory for wrapper (defaults to `~/.local/bin`)

### Script Retrieval

Scripts are fetched from GitHub releases:

1. **Release asset**: If PATH matches a release asset, download directly
2. **Source archive**: Otherwise, download source tarball/zipball and extract

Path conventions:
- `scripts/myscript.py` - Look in release assets first, then source archive
- `source:scripts/myscript.py` - Explicitly from source archive
- `asset:myscript.py` - Explicitly from release assets

## Wrapper Strategy

Scriptman creates simple shell script wrappers that directly execute the
downloaded scripts. This approach is simple, portable, and avoids unnecessary
complexity.

### Shell Script Wrapper

For each installed script, scriptman generates an executable shell script with
paths baked in:

```bash
#!/bin/sh
exec /usr/bin/python3 /home/user/.local/share/scriptman/scripts/myscript.py "$@"
```

Saved as: `~/.local/bin/myscript` (marked `chmod +x`)

### Benefits

- **Simple**: No compilation, no hardlinks, just a plain shell script
- **Portable**: Works on any Unix system with `/bin/sh` (POSIX guaranteed)
- **Transparent**: Users can inspect and understand the wrapper
- **No dependencies**: Doesn't require a C compiler or any build tools
- **Direct execution**: The shell script is the wrapper itself

### Example

Installing `myscript.py` creates `~/.local/bin/myscript`:

```bash
#!/bin/sh
exec python3 /home/user/.local/share/scriptman/scripts/myscript.py "$@"
```

Running `myscript arg1 arg2` directly executes the Python script with arguments.



## Registry Structure

The registry tracks installed scripts for management commands (list, check, remove).

Location: `~/.config/scriptman/registry.json`

```json
{
    "schema_version": 1,
    "scripts": {
        "myscript": {
            "repo": "owner/repo",
            "source_path": "scripts/myscript.py",
            "local_script": "/home/user/.local/share/scriptman/scripts/myscript.py",
            "interpreter": "python3",
            "wrapper_path": "/home/user/.local/bin/myscript",
            "installed_at": "2026-01-10T12:00:00Z",
            "version": "v1.2.3"
        }
    }
}
```

## Commands

### install

```bash
scriptman install REPO PATH [--interpreter CMD] [--name NAME] [--into DIR]
```

- Downloads script from GitHub release or source archive
- Stores script in `~/.local/share/scriptman/scripts/`
- Creates shell script wrapper
- Records in registry

### list

```bash
scriptman list [--long] [--json]
```

### remove

```bash
scriptman remove <name> [--yes]
```

Removes wrapper, downloaded script, and registry entry.

### check

```bash
scriptman check
```

Verifies:
- All wrappers exist
- All downloaded scripts exist
- Interpreters are available

### update

```bash
scriptman update [name] [--all]
```

Re-downloads script from GitHub (latest release) and recreates wrapper.

## Interpreter Detection

Priority order:

1. Explicit `--interpreter` flag
2. Shebang line in downloaded script (`#!/usr/bin/env python3`)
3. File extension mapping (`.py` → `python3`, `.rb` → `ruby`, `.sh` → `sh`)
4. Error if none can be determined

## Installation

Scriptman would be installed via execman:

```bash
execman install github.com/sfkleach/scriptman
```

Or via its own bootstrap script similar to execman's `install.sh`.

## Reserved Names

The name `scriptman` is reserved for the management CLI. Attempting to install
a script with this name should error:

```
Error: 'scriptman' is reserved for the management CLI.
       Choose a different name with --name.
```
