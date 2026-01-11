# List subcommand

Display executables managed by scriptman. 
 - Optionally filters by name (exact match).
 - Optionally add complete details rather than just a summary. 
 - Optionally output in json, full details always given.

Usage:
  scriptman list [executable-name] [flags]

Aliases:
  list, ls

Flags:
  -h, --help   help for list
      --json   Output as JSON
  -l, --long   Show detailed information

Global Flags:
      --version   Print version information

## Output Format Specification

The general concept is that the format is reminiscent of the /usr/bin/ls command. This format specification should be replicated across execman, scriptman, and pathman for consistency.

### Compact Format (default)

The compact format lists only the executable names, one per line, with no headers or additional information.

Example output:
```
decisions
my-script
another-tool
```

**Characteristics:**
- No headers or footers
- No extra information (repo, version, etc.)
- Just executable names, one per line
- Suitable for piping to other commands
- Alphabetically sorted

### Long Format (--long, -l)

The long format shows all available information in a human-friendly labeled format. Each executable is separated by a blank line.

Example output:
```
Name:         decisions
Repository:   sfkleach/scriptman
Source Path:  scripts/decisions.py
Version:      (main branch)
Commit:       a1b2c3d
Interpreter:  /usr/bin/python3
Local Script: /home/user/.local/share/scriptman/scripts/decisions.py
Wrapper Path: /home/user/.local/bin/decisions
Installed At: 2026-01-11 01:01:58

Name:         my-script
Repository:   user/repo
Source Path:  bin/script.sh
Version:      v1.2.3
Commit:       d4e5f6g
Interpreter:  /bin/bash
Local Script: /home/user/.local/share/scriptman/scripts/my-script.sh
Wrapper Path: /home/user/.local/bin/my-script
Installed At: 2026-01-10 15:30:42
```

**Characteristics:**
- Left-aligned labels with colon separator
- All available metadata displayed
- Blank line between entries
- Human-readable timestamps (YYYY-MM-DD HH:MM:SS)
- Version shows "(main branch)" if no release tag
- Commit hash truncated to 7 characters
- Alphabetically sorted

### JSON Format (--json)

The JSON format outputs the raw registry data structure for programmatic consumption.

Example output:
```json
{
  "decisions": {
    "repo": "sfkleach/scriptman",
    "source_path": "scripts/decisions.py",
    "version": "",
    "commit": "a1b2c3d4e5f6g7h8i9j0",
    "interpreter": "/usr/bin/python3",
    "local_script": "/home/user/.local/share/scriptman/scripts/decisions.py",
    "wrapper_path": "/home/user/.local/bin/decisions",
    "installed_at": "2026-01-11T01:01:58Z"
  },
  "my-script": {
    "repo": "user/repo",
    "source_path": "bin/script.sh",
    "version": "v1.2.3",
    "commit": "d4e5f6g7h8i9j0k1l2m3",
    "interpreter": "/bin/bash",
    "local_script": "/home/user/.local/share/scriptman/scripts/my-script.sh",
    "wrapper_path": "/home/user/.local/bin/my-script",
    "installed_at": "2026-01-10T15:30:42Z"
  }
}
```

**Characteristics:**
- Standard JSON formatting
- Complete data structure
- Full commit hashes (not truncated)
- ISO 8601 timestamps
- Empty strings for missing fields (e.g., version)
- Suitable for parsing by other tools

### Filtering Behavior

When an executable name is provided as an argument, the output is filtered to show only exact matches. The format remains the same (compact, long, or JSON) but only includes matching executables. 
