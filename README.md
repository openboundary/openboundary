<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="website/public/images/logo-dark.svg">
    <source media="(prefers-color-scheme: light)" srcset="website/public/images/logo-light.svg">
    <img alt="OpenBoundary" src="website/public/images/logo-dark.svg" width="120" height="120">
  </picture>
</p>

<h1 align="center">OpenBoundary</h1>

<p align="center">
  <a href="https://www.gnu.org/licenses/agpl-3.0"><img src="https://img.shields.io/badge/License-AGPL_v3-blue.svg" alt="License: AGPL v3"></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go" alt="Go Version"></a>
</p>

**Architectural guardrails for AI-assisted development.**

OpenBoundary compiles YAML specifications into type-safe TypeScript backends. Define your architecture once—routes, middleware chains, database schemas, security policies—and generate code that enforces these constraints. Business logic stays flexible; architectural invariants stay fixed.

```yaml
components:
  - id: http.server.api
    kind: http.server
    spec:
      framework: hono
      port: 3000
      middleware: [middleware.authn, middleware.authz]

  - id: usecase.create-user
    kind: usecase
    spec:
      binds_to: http.server.api:POST:/users
      middleware: []  # Explicit: public endpoint
      goal: Register a new user
```

```bash
$ bound compile spec.yaml
✓ Validated security boundaries
✓ Enforced middleware invariants
✓ Generated TypeScript project
```

## Quick Start

```bash
# Install the CLI
curl -fsSL https://openboundary.org/install.sh | sh

# Or build from source
git clone https://github.com/openboundary/openboundary.git
cd openboundary
go build -o bound ./cmd/bound

# Create and compile a project
bound init basic -o my-project
bound compile my-project/spec.yaml
```

**[Read the full documentation →](https://openboundary.org/docs/)**

## What It Does

- **Separates boundaries from logic** — Infrastructure (auth, routing, middleware) rarely changes. Domain logic changes constantly. Keep them separate.
- **Enforces security invariants** — Authentication and authorization requirements are declared once. Generated code enforces them. AI can't accidentally create an unprotected endpoint.
- **Orchestrates open standards** — OpenAPI for contracts, Drizzle for schemas, Casbin for authorization. No lock-in.

## Supported

| Component | Providers |
|-----------|-----------|
| HTTP Server | Hono |
| Database | Drizzle (PostgreSQL) |
| Authentication | better-auth |
| Authorization | Casbin |

More coming. See [roadmap](https://openboundary.org/docs/#roadmap).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). By contributing, you agree to our [CLA](CLA.md).

## License

[AGPL-3.0](LICENSE) — Use commercially, modify, distribute. Disclose source if running as a service.

---

**[Documentation](https://openboundary.org/docs/)** · **[Roadmap](https://openboundary.org/docs/#roadmap)** · **[Releases](https://github.com/openboundary/openboundary/releases)**
