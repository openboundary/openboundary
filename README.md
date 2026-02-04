# Stack Bound

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)

**Stop writing boilerplate. Start with a specification.**

stack-bound transforms YAML specifications into production-ready TypeScript backends. Define your architecture once, generate type-safe code with proper wiring.

```yaml
# spec.yaml
components:
  - id: http.server.api
    kind: http.server
    spec:
      framework: hono
      port: 3000
      openapi: ./openapi.yaml
      middleware: [middleware.authn]

  - id: middleware.authn
    kind: middleware
    spec:
      provider: better-auth
      config: ./auth.config.ts
```

```bash
$ stack-bound compile spec.yaml
✓ Parsed specification (3 components)
✓ Validated schemas and references
✓ Generated TypeScript project
→ Output: ./generated/
```

**Result:** A complete Hono application with authentication middleware, type-safe handlers, and proper dependency injection. You write the business logic, stack-bound handles the architecture.

---

## Why stack-bound?

### The Problem

Setting up a new TypeScript backend means:
- 2-3 hours configuring build tools and dependencies
- Copy-pasting server setup from your last project
- Manually wiring middleware chains
- Writing the same validation logic again
- Hoping you didn't forget anything

And then documentation drifts from code within weeks.

### The Solution

**Declare your architecture, generate the scaffolding.**

```yaml
# Write this once
components:
  - id: http.server.api
    kind: http.server
    spec:
      middleware: [middleware.authn, middleware.authz]
      depends_on: [postgres.primary]

  - id: usecase.create-user
    kind: usecase
    spec:
      binds_to: http.server.api:POST:/users
      goal: Register a new user account
      acceptance_criteria:
        - User record is created with provided email
        - Password is hashed before storage
```

```typescript
// Get this generated - with correct types and wiring
import { Hono } from 'hono';
import { authnMiddleware } from './middleware/authn';
import { authzMiddleware } from './middleware/authz';
import { db } from './db/connection';

const app = new Hono();

app.use('*', authnMiddleware);
app.use('*', authzMiddleware);

app.post('/users', async (c) => {
  // Handler stub with correct types
  // You implement the business logic
  throw new Error('Not implemented: Register a new user account');
});

export default app;
```

**Benefits:**
- ✅ Start new projects in minutes, not hours
- ✅ Enforce architectural conventions through generation
- ✅ Specifications stay in sync with code (they ARE the source)
- ✅ Onboard new developers faster (spec documents the architecture)
- ✅ Regenerate when requirements change (handlers preserved)

---

## Quick Start

### Installation

**From source:**
```bash
git clone https://github.com/stack-bound/stack-bound.git
cd stack-bound
go build -o stack-bound ./cmd/stack-bound
```

**Verify installation:**
```bash
./stack-bound --version
stack-bound v0.1.0
```

### Your First Project

**1. Create a specification file:**

```yaml
# my-api.yaml
version: "0.1.0"
name: my-api
description: A simple user management API

components:
  - id: http.server.api
    kind: http.server
    spec:
      framework: hono
      port: 3000
      openapi: ./openapi.yaml
      middleware:
        - middleware.authn
      depends_on:
        - postgres.primary

  - id: middleware.authn
    kind: middleware
    spec:
      provider: better-auth
      config: ./src/auth/auth.config.ts

  - id: postgres.primary
    kind: postgres
    spec:
      provider: drizzle
      schema: ./src/db/schema.ts

  - id: usecase.list-users
    kind: usecase
    spec:
      binds_to: http.server.api:GET:/users
      goal: List all users in the system
      acceptance_criteria:
        - Returns array of user objects
        - Excludes password hashes from response

  - id: usecase.create-user
    kind: usecase
    spec:
      binds_to: http.server.api:POST:/users
      middleware: []  # Public endpoint
      goal: Register a new user account
      acceptance_criteria:
        - User record is created with provided email
        - Password is hashed before storage
        - Confirmation email is queued
```

**2. Create your OpenAPI specification:**

```yaml
# openapi.yaml
openapi: 3.1.0
info:
  title: User Management API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
      responses:
        '200':
          description: List of users
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/User'
    post:
      summary: Create user
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateUserRequest'
      responses:
        '201':
          description: User created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
components:
  schemas:
    User:
      type: object
      properties:
        id: { type: string }
        email: { type: string }
        createdAt: { type: string, format: date-time }
    CreateUserRequest:
      type: object
      required: [email, password]
      properties:
        email: { type: string, format: email }
        password: { type: string, minLength: 8 }
```

**3. Create your Drizzle schema:**

```typescript
// src/db/schema.ts
import { pgTable, uuid, varchar, timestamp } from 'drizzle-orm/pg-core';

export const users = pgTable('users', {
  id: uuid('id').primaryKey().defaultRandom(),
  email: varchar('email', { length: 255 }).notNull().unique(),
  passwordHash: varchar('password_hash', { length: 255 }).notNull(),
  createdAt: timestamp('created_at').defaultNow().notNull(),
});
```

**4. Generate your project:**

```bash
./stack-bound compile my-api.yaml
```

**Output:**
```
✓ Parsed specification (4 components)
✓ Validated schemas and cross-references
✓ Resolved component dependencies
✓ Generated TypeScript project

Generated files:
  ./generated/src/server.ts
  ./generated/src/middleware/authn.ts
  ./generated/src/handlers/users.ts
  ./generated/src/db/connection.ts
  ./generated/src/db/schema.ts
  ./generated/package.json
  ./generated/tsconfig.json

Next steps:
  cd generated
  npm install
  npm run dev
```

**5. Implement your business logic:**

```typescript
// generated/src/handlers/users.ts (edit this file)
export async function listUsers(c: Context) {
  const users = await db.select({
    id: schema.users.id,
    email: schema.users.email,
    createdAt: schema.users.createdAt,
  }).from(schema.users);

  return c.json(users);
}

export async function createUser(c: Context) {
  const body = await c.req.json();
  const passwordHash = await hashPassword(body.password);

  const [user] = await db.insert(schema.users).values({
    email: body.email,
    passwordHash,
  }).returning({
    id: schema.users.id,
    email: schema.users.email,
    createdAt: schema.users.createdAt,
  });

  // TODO: Queue confirmation email

  return c.json(user, 201);
}
```

**6. Run your API:**

```bash
cd generated
npm install
npm run dev

# Server running on http://localhost:3000
```

**That's it.** You have a type-safe API with authentication, database integration, and proper middleware chains.

---

## Core Principles

### 1. Orchestrate, Don't Abstract

stack-bound doesn't reinvent the wheel. It orchestrates existing tools:

- **OpenAPI** defines your API contracts (with full ecosystem: editors, validators, docs generators)
- **Drizzle** defines your database schema (with TypeScript DSL and migration tools)
- **better-auth** handles authentication (with providers and session management)
- **Casbin** handles authorization (with policy files and testing tools)

The spec references these files and composes them into a working system.

**Why this matters:** You use each tool's native format and ecosystem. No leaky abstractions. No "stack-bound-flavored" OpenAPI.

### 2. Specifications as Source of Truth

Traditional flow:
```
Write spec → Write code → Code drifts → Spec becomes stale → Delete spec
```

stack-bound flow:
```
Write spec → Generate code → Change spec → Regenerate code → Spec stays current
```

The specification is always accurate because it's the input to code generation.

### 3. Structure vs Implementation

stack-bound generates **structure**, not **implementation**:

**Generated (structural):**
- Server setup and middleware chains
- Route bindings and dependency injection
- Type definitions and validation
- Database connection and schema imports

**You write (implementation):**
- Business logic inside handlers
- Domain rules and validation
- Error handling and edge cases
- Custom algorithms

This separation lets each be verified appropriately:
- Structure = mechanical correctness (type-checks, routes resolve)
- Implementation = semantic correctness (does it do the right thing?)

---

## Component Kinds

### `http.server`

HTTP server configuration with middleware and routing.

```yaml
id: http.server.api
kind: http.server
spec:
  framework: hono           # Currently: hono (more coming)
  port: 3000
  openapi: ./openapi.yaml   # OpenAPI 3.1 specification
  middleware:               # Middleware execution order
    - middleware.authn
    - middleware.authz
  depends_on:               # Available for dependency injection
    - postgres.primary
```

**Generates:**
- Hono application setup
- Middleware chain configuration
- Route bindings from usecase components
- Dependency injection wiring

### `middleware`

Cross-cutting concerns like authentication and authorization.

**Authentication with better-auth:**
```yaml
id: middleware.authn
kind: middleware
spec:
  provider: better-auth
  config: ./src/auth/auth.config.ts
```

**Authorization with Casbin:**
```yaml
id: middleware.authz
kind: middleware
spec:
  provider: casbin
  depends_on:
    - middleware.authn      # Authz requires authn
  model: ./src/auth/model.conf
  policy: ./src/auth/policy.csv
```

**Generates:**
- Middleware functions with correct signatures
- Integration with your config files
- Proper dependency ordering

### `postgres`

PostgreSQL database connection using Drizzle ORM.

```yaml
id: postgres.primary
kind: postgres
spec:
  provider: drizzle
  schema: ./src/db/schema.ts   # Your Drizzle schema file
```

**Generates:**
- Database connection setup
- Schema imports and type exports
- Connection pooling configuration

**Runtime configuration** via environment variables:
```bash
DATABASE_URL=postgresql://user:pass@localhost:5432/mydb
```

### `usecase`

Business requirements with acceptance criteria. Connects requirements to routes.

```yaml
id: usecase.create-user
kind: usecase
spec:
  binds_to: http.server.api:POST:/users    # Server:Method:Path
  middleware: []                            # Override middleware ([] = public)
  goal: Register a new user account
  actor: anonymous                          # Who performs this action
  preconditions:
    - Email address is not already registered
  acceptance_criteria:
    - User record is created with provided email
    - Password is hashed before storage
    - Confirmation email is queued for delivery
    - Response includes user ID but not password
  postconditions:
    - User exists in database with status pending_verification
```

**Generates:**
- Handler stubs with correct types from OpenAPI
- Route bindings on specified server
- Middleware configuration for this specific route
- Comments with acceptance criteria (implementation guidance)

**The middleware field:**
- Omitted = full middleware chain applies (secure by default)
- `[]` = no middleware (public endpoint)
- `[middleware.authn]` = subset of middleware

---

## Project Structure

```
your-project/
├── spec.yaml                    # Your specification
├── openapi.yaml                 # API contract
├── src/
│   ├── db/
│   │   └── schema.ts           # Drizzle schema
│   └── auth/
│       ├── auth.config.ts      # better-auth config
│       ├── model.conf          # Casbin model (optional)
│       └── policy.csv          # Casbin policies (optional)
└── generated/                   # Generated by stack-bound
    ├── src/
    │   ├── server.ts           # Hono application
    │   ├── middleware/
    │   │   ├── authn.ts
    │   │   └── authz.ts
    │   ├── handlers/
    │   │   └── users.ts        # Handler stubs - implement these
    │   └── db/
    │       └── connection.ts
    ├── package.json
    └── tsconfig.json
```

**Development workflow:**
1. Edit `spec.yaml`, `openapi.yaml`, or `schema.ts`
2. Run `stack-bound compile spec.yaml`
3. Implement business logic in `generated/src/handlers/*.ts`
4. Repeat

**Important:** Currently, regeneration overwrites all files in `generated/`. Save your handler implementations separately or use version control. (Preservation of custom code is coming - see roadmap.)

---

## Examples

### Example 1: Todo API

```yaml
# todo-api.yaml
version: "0.1.0"
name: todo-api
description: Simple todo list API

components:
  - id: http.server.api
    kind: http.server
    spec:
      framework: hono
      port: 3000
      openapi: ./openapi.yaml
      middleware: [middleware.authn]
      depends_on: [postgres.primary]

  - id: middleware.authn
    kind: middleware
    spec:
      provider: better-auth
      config: ./auth.config.ts

  - id: postgres.primary
    kind: postgres
    spec:
      provider: drizzle
      schema: ./db/schema.ts

  - id: usecase.list-todos
    kind: usecase
    spec:
      binds_to: http.server.api:GET:/todos
      goal: List all todos for authenticated user
      acceptance_criteria:
        - Returns only todos owned by current user
        - Sorted by created date descending

  - id: usecase.create-todo
    kind: usecase
    spec:
      binds_to: http.server.api:POST:/todos
      goal: Create a new todo item
      acceptance_criteria:
        - Todo is associated with authenticated user
        - Title is required and non-empty

  - id: usecase.update-todo
    kind: usecase
    spec:
      binds_to: http.server.api:PUT:/todos/:id
      goal: Update an existing todo
      preconditions:
        - Todo exists and belongs to current user
      acceptance_criteria:
        - Can update title and completed status
        - Cannot change owner

  - id: usecase.delete-todo
    kind: usecase
    spec:
      binds_to: http.server.api:DELETE:/todos/:id
      goal: Delete a todo item
      preconditions:
        - Todo exists and belongs to current user
```

### Example 2: Multi-Service Architecture

```yaml
# multi-service.yaml
version: "0.1.0"
name: multi-service-app

components:
  # Public API
  - id: http.server.public-api
    kind: http.server
    spec:
      framework: hono
      port: 3000
      openapi: ./openapi-public.yaml
      middleware: []  # No auth required
      depends_on: [postgres.primary]

  # Admin API
  - id: http.server.admin-api
    kind: http.server
    spec:
      framework: hono
      port: 3001
      openapi: ./openapi-admin.yaml
      middleware: [middleware.admin-auth, middleware.admin-authz]
      depends_on: [postgres.primary]

  - id: middleware.admin-auth
    kind: middleware
    spec:
      provider: better-auth
      config: ./admin-auth.config.ts

  - id: middleware.admin-authz
    kind: middleware
    spec:
      provider: casbin
      depends_on: [middleware.admin-auth]
      model: ./admin-model.conf
      policy: ./admin-policy.csv

  - id: postgres.primary
    kind: postgres
    spec:
      provider: drizzle
      schema: ./db/schema.ts

  # Public endpoints
  - id: usecase.public-health
    kind: usecase
    spec:
      binds_to: http.server.public-api:GET:/health
      middleware: []
      goal: Health check endpoint

  # Admin endpoints
  - id: usecase.admin-list-users
    kind: usecase
    spec:
      binds_to: http.server.admin-api:GET:/users
      goal: List all users (admin only)
      actor: admin
```

---

## Validation

stack-bound validates your specifications at multiple levels:

### Schema Validation

Every component must conform to its JSON Schema:

```bash
$ stack-bound validate spec.yaml

✗ Validation failed:
  - Component 'http.server.api' is missing required field 'port'
  - Component 'middleware.authn' has invalid provider 'custom-auth'
    (supported: better-auth, casbin)
```

### Cross-Reference Validation

All component references must exist:

```bash
$ stack-bound validate spec.yaml

✗ Validation failed:
  - Component 'http.server.api' references middleware 'middleware.authn'
    which is not defined
  - Usecase 'usecase.create-user' binds to server 'http.server.api'
    which is not defined
```

### Semantic Validation

Values must make sense in context:

```bash
$ stack-bound validate spec.yaml

✗ Validation failed:
  - Port 99999 is out of valid range (1-65535)
  - Casbin policy file './policy.csv' does not exist
  - OpenAPI file './openapi.yaml' is not valid OpenAPI 3.x
```

### Dependency Validation

No circular dependencies allowed:

```bash
$ stack-bound validate spec.yaml

✗ Validation failed:
  - Circular dependency detected:
    middleware.authz → middleware.authn → middleware.authz
```

---

## Roadmap

### ✅ Current (v0.1.0)

- YAML specification format
- TypeScript code generation (Hono framework)
- Component kinds: http.server, middleware, postgres, usecase
- OpenAPI 3.1 integration
- Drizzle ORM integration
- better-auth authentication
- Casbin authorization
- Multi-server support
- JSON Schema validation

### 🚧 Coming Soon (v0.2.0 - Q2 2026)

**Safezone Preservation:**
- Mark sections of generated code as "safe zones"
- Preserve your custom code across regeneration
- Regenerate structure without losing business logic

**Enhanced Code Generation:**
- Handler implementations from acceptance criteria
- Basic CRUD operations from schema definitions
- Request validation middleware from OpenAPI schemas

**Additional Component Kinds:**
- `event` - Event definitions for event-driven architecture
- `job` - Background job definitions
- `cron` - Scheduled task definitions

### 📋 Planned (v0.3.0+ - Q3 2026)

**More Frameworks:**
- Go: `chi`, `fiber`, `echo`
- Python: `fastapi`, `flask`
- Rust: `axum`, `actix-web`

**More Databases:**
- MongoDB
- Redis
- MySQL/MariaDB

**Protocol Support:**
- gRPC service definitions
- GraphQL schema generation
- WebSocket endpoints

**Developer Experience:**
- VS Code extension with IntelliSense
- Specification file hot-reload during development
- Interactive CLI for creating specifications

**Testing:**
- Test generation from acceptance criteria
- Mock server generation from OpenAPI
- Integration test scaffolding

### 💭 Exploring (2027+)

- Deployment configuration (Kubernetes, Docker Compose)
- CI/CD pipeline generation
- API documentation generation
- Client SDK generation (TypeScript, Python, Go)
- Specification inference from existing code
- Migration tools for brownfield projects

**Want to influence the roadmap?** Open an issue or discussion on GitHub. We prioritize features based on community feedback.

---

## CLI Reference

### `stack-bound validate`

Validate a specification file without generating code.

```bash
stack-bound validate <spec-file>

Options:
  --strict    Enable strict validation (warnings become errors)
  --json      Output validation results as JSON
```

**Example:**
```bash
$ stack-bound validate my-api.yaml
✓ Specification is valid

$ stack-bound validate my-api.yaml --strict
✗ Validation failed (1 warning in strict mode):
  - Usecase 'create-user' has no postconditions defined
```

### `stack-bound compile`

Compile a specification to code.

```bash
stack-bound compile <spec-file> [options]

Options:
  -o, --output <dir>     Output directory (default: ./generated)
  --dry-run              Show what would be generated without writing files
  --force                Overwrite existing files without confirmation
  --target <platform>    Target platform (default: typescript-hono)
```

**Example:**
```bash
$ stack-bound compile my-api.yaml -o ./backend

✓ Parsed specification (8 components)
✓ Validated schemas and references
✓ Resolved dependencies
✓ Generated TypeScript project
→ Output: ./backend/

Files generated:
  ./backend/src/server.ts
  ./backend/src/middleware/authn.ts
  ./backend/src/handlers/users.ts
  ./backend/package.json
  (8 files total)
```

### `stack-bound init`

Create a new specification from a template.

```bash
stack-bound init [template]

Templates:
  basic      - Basic HTTP API (default)
  crud       - CRUD API with auth
  multiservice - Multiple servers with different auth

Options:
  -o, --output <dir>     Output directory (default: ./)
```

**Example:**
```bash
$ stack-bound init crud -o ./my-project

Created:
  ./my-project/spec.yaml
  ./my-project/openapi.yaml
  ./my-project/src/db/schema.ts
  ./my-project/src/auth/auth.config.ts

Next steps:
  cd my-project
  stack-bound compile spec.yaml
```

### `stack-bound version`

Show version information.

```bash
$ stack-bound version
stack-bound v0.1.0
go version: go1.21.5
platform: darwin/arm64
```

---

## FAQ

### How is this different from code generators like Yeoman or Hygen?

Code generators create files once. stack-bound maintains a living relationship between specification and code. Change the spec, regenerate the project. The specification stays current because it's the source of truth.

### What happens to my custom code when I regenerate?

**Currently (v0.1.0):** Regeneration overwrites all files in the output directory. Save your handler implementations separately or use version control.

**Coming soon (v0.2.0):** Safezone markers let you mark sections that should be preserved across regeneration. The spec generates structure, you add implementation, and regeneration respects both.

### Can I use stack-bound for existing projects?

Not yet. v0.1.0 is designed for greenfield projects. Brownfield migration (generating specs from existing code) is on the roadmap for 2027.

### Does stack-bound replace my framework?

No. stack-bound generates code using your framework (Hono, Drizzle, better-auth, etc.). You still use those frameworks' APIs and ecosystems. stack-bound just wires them together according to your specification.

### What if I need something stack-bound doesn't generate?

Write it yourself! stack-bound generates structural scaffolding. You write business logic, custom middleware, background jobs, webhooks, or anything else your application needs.

The generated code is readable TypeScript. You can edit it, extend it, or ignore parts of it.

### Can I contribute new component kinds?

Yes! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines. Component kinds are modular - you can add support for new frameworks, databases, or patterns.

### Is this production-ready?

v0.1.0 is an early release. Use it for:
- ✅ Greenfield projects and prototypes
- ✅ Internal tools and side projects
- ✅ Learning and experimentation

Be cautious with:
- ⚠️ Production systems with high traffic
- ⚠️ Systems requiring custom middleware logic
- ⚠️ Projects needing safezone preservation (coming in v0.2.0)

The generated code is production-quality, but the tool itself is evolving rapidly.

---

## Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Setting up your development environment
- Adding new component kinds
- Writing tests
- Submitting pull requests

By contributing, you agree to our [Contributor License Agreement](CLA.md).

### Community

- **GitHub Issues:** Bug reports and feature requests
- **GitHub Discussions:** Questions, ideas, and showcases
- **Twitter/X:** [@maxrochefort](https://twitter.com/maxrochefort) for updates

---

## License

stack-bound is licensed under the [GNU Affero General Public License v3.0](LICENSE).

**AGPL-3.0 Summary:**
- ✅ Use commercially
- ✅ Modify and distribute
- ✅ Use privately
- ⚠️ Disclose source if you run it as a service
- ⚠️ Same license for derivatives

Questions? Contact: [impermanent.architect@pm.me](mailto:impermanent.architect@pm.me)

---

## Credits

Built by [Max Rochefort-Shugar](https://maxshugar.com).

Inspired by:
- Peter Naur's "Programming as Theory Building" (1985)
- Infrastructure-as-Code tools (Terraform, Pulumi)
- The generation-gap pattern in model-driven development

Special thanks to the open source communities behind:
- [Hono](https://hono.dev) - Fast, lightweight web framework
- [Drizzle ORM](https://orm.drizzle.team) - TypeScript ORM
- [better-auth](https://www.better-auth.com) - Authentication library
- [Casbin](https://casbin.org) - Authorization framework

---

**Ready to stop writing boilerplate?**

```bash
git clone https://github.com/stack-bound/stack-bound.git
cd stack-bound
go build -o stack-bound ./cmd/stack-bound
./stack-bound init
```

Star the repo if this looks useful: [github.com/stack-bound/stack-bound](https://github.com/stack-bound/stack-bound) ⭐
