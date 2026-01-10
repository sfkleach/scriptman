# Shebang Line Handling in Interpreter Detection, 2026-01-10

## Issue

When installing scripts from GitHub, we need to determine which interpreter to use. Scripts often contain shebang lines (`#!/usr/bin/env python3`) that suggest an interpreter, but we must decide when to trust these shebangs versus using our own configured interpreters.

## Factors

**Security and reliability concerns:**
- Shebang lines might reference hard-coded locations that don't exist on the target system (e.g., `/usr/local/bin/python3` when only `/usr/bin/python3` exists)
- Shebang lines may be intended for editor hints rather than actual execution (VSCode famously misidentifies file types without shebangs)
- Shebang lines might contain options specific to different Unix variants that could be incompatible
- Arbitrary shebang execution could pose security risks

**User experience:**
- Users expect scripts to "just work" when possible
- Users should be warned when we deviate from script's declared interpreter
- Interactive prompting should be minimal but present for ambiguous cases

**Technical considerations:**
- File extensions provide reliable interpreter hints (`.py` → Python, `.rb` → Ruby)
- `#!/usr/bin/env` form is well-understood and portable across systems
- Shebang arguments beyond the interpreter require careful handling

## Decision

Implement a nuanced interpreter detection strategy with consistency checking and selective user prompting:

**Priority 1: Explicit `--interpreter` flag**
- Always honored, no questions asked

**Priority 2: Shebang line exists**
- Parse shebang to extract interpreter and any arguments
- Check for shebang arguments:
  - If arguments present AND not `#!/usr/bin/env` form → **prompt user for approval**
  - `#!/usr/bin/env interpreter` is considered safe (well-understood, portable)
- Check file extension consistency:
  - Extension maps to **consistent** interpreter (e.g., `.py` with `python`/`python3`) → **use our configured interpreter** (from alternatives)
  - Extension maps to **inconsistent** interpreter (e.g., `.py` with `ruby`) → **prompt user** to choose between shebang vs our extension mapping
  - No extension → **prompt user** for permission to copy shebang verbatim

**Priority 3: Extension mapping only** (no shebang)
- Use extension alternatives with PATH checking
- Warn if no alternative found on PATH

**Priority 4: Error** if nothing can be determined

## Consequences

**Positive:**
- Safer execution: don't blindly trust arbitrary shebang lines
- Better portability: prefer PATH-based interpreter resolution over hard-coded locations
- Flexibility: allow users to make informed decisions in ambiguous cases
- Consistency: `.py` files consistently use Python even if shebang says `python` vs `python3`

**Negative:**
- More complex implementation with multiple code paths
- User prompting adds friction to installation process
- Need to define what "consistent" means (e.g., `python` vs `python3` for `.py`)

**Neutral:**
- Requires defining interpreter families/consistency rules
- Need good error messages to explain why prompting

## Consistency Rules

Define interpreter families where variations are considered consistent:
- Python family: `python`, `python2`, `python3`, `python3.11`, etc.
- Ruby family: `ruby`, `ruby2`, `ruby3`, etc.
- Perl family: `perl`, `perl5`, etc.
- Shell family: `sh`, `bash`, `dash`, `zsh`, etc.

For consistency checking:
- Extract base interpreter name from shebang (strip version numbers, paths)
- Compare family of shebang interpreter vs extension-mapped interpreter
- Same family → consistent, use our configured version
- Different family → inconsistent, prompt user

## Prompting Design

When prompting users:
1. Show what we detected (shebang, extension, arguments)
2. Explain the concern (hard-coded path, inconsistency, arguments present)
3. Offer clear options with recommendations
4. Allow `--assume-yes` or `--force` flags to skip prompts in automation

Example prompts:
```
Script has shebang: #!/usr/local/bin/python3 -u
This uses a hard-coded path and arguments.
Options:
  1. Use our configured Python interpreter (recommended)
  2. Copy shebang verbatim (may fail if path doesn't exist)
  3. Abort installation
Choice [1]:
```

```
Script has shebang: #!/usr/bin/env ruby
But file extension .py suggests Python interpreter.
Options:
  1. Use Python (from extension)
  2. Use Ruby (from shebang)
  3. Abort installation
Choice [1]:
```

## Additional Notes

This design prioritizes safety and consistency over blind trust of downloaded scripts. The wrapper mechanism (explicitly calling interpreter) means we can safely ignore problematic shebangs and substitute safer alternatives.

Future enhancement: Add `--trust-shebang` flag to skip consistency checks for advanced users who know what they're doing.
