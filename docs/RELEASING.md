# Releasing

Copyright 2026 The OpenBoundary Authors
SPDX-License-Identifier: Apache-2.0

---

## Automated Releases

Releases are fully automated via [GoReleaser](https://goreleaser.com/) and GitHub Actions. When a version tag is pushed, the release workflow builds binaries for all supported platforms and creates a GitHub release with checksums.

### Trigger

Push a tag matching the pattern `v*` (e.g., `v1.0.0`, `v0.2.0-beta.1`) to trigger the release workflow.

### What Gets Built

GoReleaser builds the `bound` CLI binary and creates archives:

| OS      | Architecture | Archive Format |
|---------|--------------|----------------|
| Linux   | amd64        | tar.gz         |
| Linux   | arm64        | tar.gz         |
| macOS   | amd64        | tar.gz         |
| macOS   | arm64        | tar.gz         |
| Windows | amd64        | zip            |
| Windows | arm64        | zip            |

Each release also includes a `checksums.txt` file for verification.

## Creating a Release

### Standard Release

1. **Ensure `main` is ready**
   - All CI checks pass
   - CHANGELOG is updated (if applicable)
   - Version bump commits are merged

2. **Create and push the tag**
   ```bash
   git checkout main
   git pull origin main
   git tag v1.0.0
   git push origin v1.0.0
   ```

3. **Wait for the workflow**
   - Monitor the [Actions tab](../../actions) for the release workflow
   - GoReleaser will build all binaries and create the GitHub release

4. **Verify the release**
   - Check the [Releases page](../../releases) for the new release
   - Verify all expected assets are present
   - Test the install script works: `curl -fsSL https://raw.githubusercontent.com/openboundary/openboundary/main/website/static/install.sh | sh`

### Pre-release

For alpha, beta, or release candidates:

```bash
git tag v1.0.0-alpha.1
git push origin v1.0.0-alpha.1
```

GitHub will automatically mark these as pre-releases based on the tag format.

## Hotfix Release

For urgent fixes that need to bypass the normal release cycle:

1. **Create a hotfix branch from the release tag**
   ```bash
   git checkout -b hotfix/v1.0.1 v1.0.0
   ```

2. **Apply the fix**
   - Make the minimal necessary changes
   - Ensure tests pass

3. **Merge to main** (if applicable)
   ```bash
   git checkout main
   git merge hotfix/v1.0.1
   git push origin main
   ```

4. **Tag and release**
   ```bash
   git checkout hotfix/v1.0.1
   git tag v1.0.1
   git push origin v1.0.1
   ```

5. **Clean up**
   ```bash
   git branch -d hotfix/v1.0.1
   ```

## Changelog

The release workflow automatically generates a changelog from commit messages. Commits are filtered:

- **Included**: Feature commits, bug fixes, breaking changes
- **Excluded**: Commits prefixed with `docs:`, `test:`, or `chore:`

## Troubleshooting

### Release workflow failed

1. Check the Actions tab for error details
2. Common issues:
   - GoReleaser configuration errors
   - Missing `GITHUB_TOKEN` permissions
   - Tag already exists

### Need to re-release

If a release needs to be redone:

```bash
# Delete the tag locally and remotely
git tag -d v1.0.0
git push origin :refs/tags/v1.0.0

# Delete the GitHub release manually via the web UI

# Re-tag and push
git tag v1.0.0
git push origin v1.0.0
```
