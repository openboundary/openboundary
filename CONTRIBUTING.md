# Contributing to OpenBoundary

Thank you for your interest in contributing to OpenBoundary! This document provides guidelines for contributing to the project.

## Contributor License Agreement (CLA)

**Before your contribution can be accepted, you must agree to our [Contributor License Agreement](CLA.md).**

By submitting a pull request, you indicate your agreement to the CLA. This agreement grants OpenBoundary the right to use your contribution under the project's open source license (AGPL-3.0) as well as under commercial licenses.

### Why a CLA?

OpenBoundary uses a dual-licensing model:
- **Community Edition**: AGPL-3.0 (open source)
- **Enterprise Edition**: Commercial license (proprietary)

The CLA ensures we can continue to offer both options. Without it, we would be unable to include your contributions in the commercial product, which funds ongoing development.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally
3. **Create a branch** for your changes
4. **Make your changes** following our guidelines
5. **Test your changes** thoroughly
6. **Submit a pull request**

## Development Setup

```bash
# Clone the repository
git clone https://github.com/openboundary/openboundary.git
cd openboundary

# Build the project
go build -o openboundary ./cmd/openboundary

# Run tests
go test ./...

# Validate example spec
./openboundary validate examples/basic/spec.yaml
```

## Code Guidelines

### Go Code

- Follow standard Go conventions and `gofmt`
- Write tests for new functionality
- Keep functions focused and small
- Document exported types and functions

### Specification Schema

- Update `schemas/openboundary.schema.json` for spec changes
- Copy changes to `internal/validator/openboundary.schema.json`
- Update examples to reflect schema changes
- Add tests for new validation rules

### Commits

- Write clear, concise commit messages
- Reference issues where applicable: `Fix #123: description`
- Keep commits focused on a single change

## Pull Request Process

1. **Ensure tests pass**: `go test ./...`
2. **Update documentation** if needed
3. **Describe your changes** in the PR description
4. **Link related issues** if applicable
5. **Wait for review** - maintainers will provide feedback

## What We're Looking For

### Good First Issues

Look for issues labeled `good first issue` for beginner-friendly tasks.

### Feature Contributions

Before starting significant work:
1. Check existing issues and PRs
2. Open an issue to discuss the feature
3. Wait for maintainer feedback

This prevents wasted effort on features that may not align with the project direction.

### Bug Fixes

- Include a test that reproduces the bug
- Explain the root cause in your PR
- Keep the fix minimal and focused

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Assume good intentions

## Questions?

- Open an issue for technical questions
- Email community@openboundary.org for other inquiries

## License

By contributing, you agree that your contributions will be licensed under the AGPL-3.0 license and that OpenBoundary may also distribute them under commercial licenses as described in the CLA.
