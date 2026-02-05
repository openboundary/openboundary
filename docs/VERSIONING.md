# Versioning

Copyright 2026 The OpenBoundary Authors
SPDX-License-Identifier: Apache-2.0

---

OpenBoundary follows [Semantic Versioning 2.0.0](https://semver.org/) (SemVer).

## Version Format

```
MAJOR.MINOR.PATCH[-PRERELEASE]
```

Examples: `1.0.0`, `2.1.3`, `1.0.0-alpha.1`, `2.0.0-rc.1`

## Version Bumps

### MAJOR (Breaking Changes)

Increment MAJOR when making incompatible changes:

- **Spec format changes** that break existing YAML specifications
- **CLI interface changes** that break scripts or automation
- **Removing features** or component types
- **Changing default behavior** in ways that affect existing users
- **Generated code structure changes** that require migration

Examples:
- Renaming a required field in component specs
- Changing the `bound compile` output directory structure
- Removing support for a component kind

### MINOR (New Features)

Increment MINOR for backwards-compatible additions:

- **New features** and capabilities
- **New component types** (kinds)
- **New CLI commands** or flags
- **New spec fields** that are optional or have defaults
- **New middleware types** or integrations

Examples:
- Adding a new `redis.cache` component kind
- Adding `--dry-run` flag to `bound compile`
- Supporting a new database type in `database.postgres`

### PATCH (Bug Fixes)

Increment PATCH for backwards-compatible fixes:

- **Bug fixes** that don't change the API
- **Documentation updates**
- **Performance improvements**
- **Security fixes** (that don't require API changes)
- **Dependency updates** (that don't affect the public API)

Examples:
- Fixing a code generation bug
- Improving compile time performance
- Fixing a typo in error messages

## Pre-release Versions

Pre-release versions indicate unstable releases:

| Tag | Purpose | Example |
|-----|---------|---------|
| `alpha` | Early development, expect breaking changes | `v1.0.0-alpha.1` |
| `beta` | Feature complete, may have bugs | `v1.0.0-beta.1` |
| `rc` | Release candidate, final testing | `v1.0.0-rc.1` |

Pre-releases:
- Are not recommended for production use
- May have incomplete documentation
- May change without notice between pre-release versions

## Version 0.x.x

While the major version is `0` (e.g., `0.1.0`, `0.2.0`):

- The API is considered **unstable**
- MINOR bumps may include breaking changes
- PATCH bumps are for bug fixes only

This allows rapid iteration during initial development. Once the API stabilizes, version `1.0.0` will be released with full SemVer guarantees.

## Compatibility Promise

Starting from version `1.0.0`:

- **Spec files** written for `1.x` will work with all `1.x` releases
- **CLI scripts** written for `1.x` will work with all `1.x` releases
- **Deprecations** will be announced at least one MINOR version before removal
- **Breaking changes** will only occur in MAJOR version bumps
