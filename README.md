# OpenBoundary

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)

**AI agents run wild. Unless you set the boundaries.**

OpenBoundary compiles YAML specifications into type-safe TypeScript backends. Define your architecture once — security requirements, middleware chains, database schemas — and let AI write the business logic while OpenBoundary enforces the rules.

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
$ openboundary compile spec.yaml
✓ Validated security boundaries
✓ Enforced middleware invariants
✓ Generated TypeScript project
```

## Quick Start

```bash
git clone https://github.com/openboundary/openboundary.git
cd openboundary
go build -o openboundary ./cmd/openboundary
./openboundary init basic -o my-project
./openboundary compile my-project/spec.yaml
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
