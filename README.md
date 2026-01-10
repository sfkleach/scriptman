# Scriptman

Scriptman is a command-line tool that manages single-file scripts downloaded
from GitHub repositories. It wraps scripts with native executables, providing
central management, updates, and clean removal.

**Platform support**: Unix-only (Linux, macOS, BSD). No Windows support.

## Features

- **Central management** - Track installed scripts and their origins
- **Updates** - Re-fetch latest version from GitHub with one command
- **Clean removal** - Remove script and wrapper together
- **Name mapping** - Install `owner/repo/scripts/long-name.py` as just `tool`
- **Interpreter binding** - Specify interpreter at install time, not in shebang

## Installation

Install via [execman](https://github.com/sfkleach/execman):

```bash
execman install github.com/sfkleach/scriptman
```

Or download from [releases](https://github.com/sfkleach/scriptman/releases).

## Usage

### Install a script

```bash
scriptman install owner/repo scripts/tool.py
```

With options:

```bash
scriptman install owner/repo scripts/tool.py --name mytool --interpreter python3
```

### List installed scripts

```bash
scriptman list
```

### Update a script

```bash
scriptman update mytool
```

### Remove a script

```bash
scriptman remove mytool
```

### Check all scripts

```bash
scriptman check
```

## How It Works

Scriptman uses a hybrid wrapper strategy:

1. **With C compiler**: Generates and compiles a tiny C binary with paths baked in
2. **Without compiler**: Creates a shell script + hardlink to scriptman binary

Both approaches avoid runtime registry lookups - paths are baked into the
wrapper at install time.

## Relationship to Execman

| Tool | Purpose |
|------|---------|
| **execman** | Manages standalone executables from GitHub releases |
| **scriptman** | Manages scripts from GitHub, wrapped as executables |

Scriptman uses execman as a library for shared infrastructure (GitHub API,
registry patterns, configuration).

## License

MIT License - see [LICENSE](LICENSE) for details.
