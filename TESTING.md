# Local CI Testing Guide

How to ensure GitHub Actions CI will pass before pushing.

## Quick Answer

```bash
# Option 1: Fast local tests (2-3 minutes)
./scripts/test-ci-local.sh

# Option 2: Exact CI simulation (10-15 minutes)
./scripts/run-ci-locally.sh
```

## Three Levels of Testing

### Level 1: Unit Tests (30 seconds)

Run before every commit:

```bash
# Go tests with race detection
go test -v -race ./...

# Linter
golangci-lint run

# Build
go build -o stack-bound ./cmd/stackbound
```

**Guarantees:** Code compiles, tests pass, linter is happy

### Level 2: Integration Tests (2-3 minutes)

Run before pushing to develop:

```bash
./scripts/test-ci-local.sh
```

This mimics all 5 CI jobs:
1. **test-compiler** - Go tests + linter
2. **generate-project** - Compile example spec
3. **test-generated** - npm install, lint, test, build
4. **test-docker** - Docker build + compose
5. **test-e2e** - Manual (run separately)

**Guarantees:** Generated code builds, Docker works, npm tests pass

### Level 3: Exact CI Simulation (10-15 minutes)

Run before important PRs:

```bash
./scripts/run-ci-locally.sh
```

Uses **Act** to run the actual `.github/workflows/ci.yml` in Docker containers.

**Guarantees:** Exact same environment as GitHub Actions

## Handling Common Issues

### Docker Hub Errors (502 Bad Gateway)

**Problem:** Docker Hub is temporarily unavailable

**Solution:**
- Wait 5-10 minutes and retry
- Use cached images: `docker pull node:20-alpine` beforehand
- Not a code issue - just infrastructure

### E2E Tests Locally

E2E tests need a running server. Test manually:

```bash
cd generated
npm install
npm run generate:types
npm run build

# Terminal 1: Start server
npm run dev

# Terminal 2: Run E2E tests
npm run test:e2e
```

### Docker Build Locally

Test Docker build without CI:

```bash
cd generated
docker build -t test-app .
docker compose up -d
docker compose ps  # Check health
docker compose logs
docker compose down -v
```

## Dev Container

The dev container matches CI exactly:

**Setup:**
1. Install "Dev Containers" VS Code extension
2. Reopen in container: `Ctrl/Cmd+Shift+P` → "Reopen in Container"
3. Auto-installs: Go, Node, Docker, golangci-lint, Act

**Pre-installed:**
- Go 1.22
- Node.js 20
- Docker-in-Docker
- golangci-lint
- Act (GitHub Actions runner)
- GitHub CLI

## CI Workflow Jobs

### Job 1: test-compiler
```bash
go test -v -race -coverprofile=coverage.out ./...
golangci-lint run
go build -o stack-bound ./cmd/stackbound
```

### Job 2: generate-project
```bash
./stack-bound validate examples/basic/spec.yaml
./stack-bound compile examples/basic/spec.yaml -o generated
npm install --package-lock-only  # Create lock file
```

### Job 3: test-generated
```bash
cd generated
npm ci
npm run generate:types  # orval generates schemas
npm run lint
npm run test
npm run build
```

### Job 4: test-docker
```bash
cd generated
docker build -t test-app:latest .
docker compose up -d
docker compose ps
curl http://localhost:3000/health
docker compose down -v
```

### Job 5: test-e2e
```bash
cd generated
npm ci
npx playwright install --with-deps chromium
npm run generate:types
npm run build
npm run start &
npm run test:e2e
```

## Best Practices

**Before Every Commit:**
```bash
go test ./...
golangci-lint run
```

**Before Every Push:**
```bash
./scripts/test-ci-local.sh
```

**Before Important PRs:**
```bash
./scripts/run-ci-locally.sh  # Full Act simulation
```

**If CI Fails:**
1. Check the error message
2. Run the failing job locally
3. Fix the issue
4. Test locally again
5. Push the fix

## Debugging Failed CI

### Find Which Job Failed

Check: https://github.com/stack-bound/stack-bound/actions

### Run That Job Locally

```bash
# Example: If test-generated failed
go build -o stack-bound ./cmd/stackbound
./stack-bound compile examples/basic/spec.yaml -o test-debug
cd test-debug
npm install
npm run generate:types
npm run lint
npm run test
npm run build
```

### Compare Local vs CI

If it works locally but fails in CI:
- Check Go/Node versions match
- Check for race conditions (run with `-race`)
- Check for absolute vs relative paths
- Check environment variables

## Quick Reference

| Command | Time | Confidence |
|---------|------|------------|
| `go test ./...` | 30s | 60% |
| `./scripts/test-ci-local.sh` | 3min | 90% |
| `./scripts/run-ci-locally.sh` | 15min | 99% |

**Use test-ci-local.sh before every push for best results!**
