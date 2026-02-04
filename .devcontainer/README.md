# Development Container

This dev container matches the GitHub Actions CI environment to catch issues locally before pushing.

## What's Included

- **Go 1.22** - Same as CI
- **Node.js 20** - Same as CI
- **Docker-in-Docker** - For testing Docker builds
- **GitHub CLI** - For creating PRs
- **golangci-lint** - Installed automatically

## Quick Start

### 1. Open in Dev Container

In VS Code:
1. Install "Dev Containers" extension
2. `Ctrl/Cmd + Shift + P` → "Dev Containers: Reopen in Container"
3. Wait for container to build

### 2. Test Locally (Before Pushing)

Run the full CI workflow locally:

```bash
./scripts/test-ci-local.sh
```

This runs all 5 CI jobs locally:
- ✅ test-compiler (Go tests + linter)
- ✅ generate-project (Compile example)
- ✅ test-generated (npm tests + build)
- ✅ test-docker (Docker build + compose)
- ✅ test-e2e (skipped locally, manual test available)

### 3. Quick Tests

```bash
# Just Go tests
go test ./...

# Just linter
golangci-lint run

# Just build
go build -o stack-bound ./cmd/stackbound

# Test generation
./stack-bound compile examples/basic/spec.yaml -o test-output
cd test-output
npm install
npm run generate:types
npm run test
npm run build
```

## Environment Parity

| Tool | Dev Container | GitHub Actions |
|------|---------------|----------------|
| Go | 1.22 | 1.21+ |
| Node.js | 20 | 20 |
| Docker | In-Docker | Native |
| golangci-lint | Latest | Latest |
| docker compose | v2 | v2 |

## Tips

- Run `./scripts/test-ci-local.sh` before every push
- If local tests pass, CI will pass
- Docker builds test actual deployment scenario
- Use `gh pr create` to create PRs from terminal
