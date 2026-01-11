# 0000 - Missing release info, 2026-01-11

## Issue

When installing a script from a GitHub repository, we need to track what version was installed so that `check` and `update` commands can determine if a newer version is available.

The complication is that not all repositories use GitHub releases. Some projects only commit to their main branch without ever creating releases. In these cases, how do we detect that changes have occurred?

## Factors

- GitHub releases provide clear version tags (e.g., `v1.2.3`) that are easy to compare
- Repositories without releases still receive updates via commits to main
- Users should be informed when updates are available, regardless of release strategy
- The `check` command needs a way to detect changes
- The `update` command needs to know what to fetch

## Decision

Record both the release tag AND the commit hash when installing a script:

1. **When a release exists:** Store the tag (e.g., `v1.2.3`) and the commit hash at time of install
2. **When no release exists:** Store empty tag and the commit hash from main branch

This allows the `check` command to:
- Compare release tags when available (semantic comparison)
- Fall back to commit hash comparison when no releases exist
- Detect the case where a repo that previously had no releases now has one

## Consequences

- Registry schema needs a `Commit` field in addition to `Version` (tag)
- `FetchScript` needs to capture the commit hash from the response or make an additional API call
- `check` command will have more complex logic to handle both versioned and unversioned repos
- Users will see "(main branch)" in listings for unversioned installs, with commit info available in detailed view

## Implementation Notes

The GitHub raw content API doesn't return commit info in headers, so we may need to:
1. Query the commits API to get the latest commit hash for the ref (tag or main)
2. Or query the repository API for the default branch's HEAD commit

The `check` command should handle these cases:
- Tag installed, newer tag available → "Update available: v1.0.0 → v1.1.0"
- Tag installed, no newer tag → "Up to date (v1.0.0)"
- No tag installed, commit changed → "Changes detected on main branch"
- No tag installed, commit same → "Up to date (main branch)"
- No tag installed, release now exists → "Release available: v1.0.0 (currently tracking main)"
