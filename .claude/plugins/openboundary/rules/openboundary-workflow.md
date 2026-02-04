---
globs:
  - "spec.yaml"
  - "openapi.yaml"
  - "src/components/**/*.ts"
  - "src/auth/**/*"
  - "src/db/**/*"
---

# OpenBoundary Development Workflow

## What is OpenBoundary?

OpenBoundary compiles YAML specifications into type-safe TypeScript backends. It separates infrastructure boundaries from business logic, ensuring AI agents cannot accidentally break security or routing code.

## The Separation Principle

```text
+------------------+     +-----------------+
|    spec.yaml     |     |   Generated     |
|   (boundaries)   | --> |  Infrastructure |
+------------------+     +-----------------+
        |                       |
        v                       v
+------------------+     +-----------------+
|   Config Files   |     | Usecase Bodies  |
|  (human-managed) |     | (business logic)|
+------------------+     +-----------------+
```

**Boundaries** = What the system allows (routes, security, middleware)
**Logic** = How the system behaves (implementation details)

## Standard Workflow

### 1. Define Architecture in spec.yaml

Add components:

```yaml
components:
  - id: http.server.api
    kind: http.server
    spec:
      framework: hono
      port: 3000
      middleware: [middleware.authn]
      depends_on: [postgres.primary]

  - id: usecase.create-order
    kind: usecase
    spec:
      binds_to: http.server.api:POST:/orders
      middleware: [middleware.authn]
      goal: Create a new order for the customer
```

### 2. Define API Contract in openapi.yaml

```yaml
paths:
  /orders:
    post:
      operationId: createOrder
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateOrderRequest'
      responses:
        '201':
          description: Created
```

### 3. Configure Infrastructure

Edit config files as needed:

- `src/auth/auth.config.ts` - Authentication settings
- `src/auth/policy.csv` - Authorization rules
- `src/db/schema.ts` - Database tables

### 4. Compile

```bash
openboundary compile spec.yaml
```

This generates:

- Server setup with routing
- Middleware chains
- Usecase stubs with type-safe signatures
- Test scaffolding

### 5. Implement Business Logic

In generated usecase files, implement the function body:

```typescript
// src/components/usecases/usecase-create-order.ts

export async function createOrder(
  input: CreateOrderRequest,
  ctx: ContextWith<'db' | 'user'>
): Promise<CreateOrderResponse> {
  // Only edit between these braces
  const order = await ctx.db.insert(orders).values({
    userId: ctx.user.id,
    items: input.items,
  }).returning();

  return { orderId: order[0].id };
}
```

### 6. Iterate

When requirements change:

1. Update spec.yaml with new boundaries
2. Recompile
3. Generated infrastructure updates
4. Implementation bodies are preserved

## Key Commands

| Command | Purpose |
|---------|---------|
| `openboundary compile spec.yaml` | Generate TypeScript from spec |
| `openboundary validate spec.yaml` | Check spec without generating |
| `openboundary init <template>` | Start new project from template |

## File Ownership

| File | Owner | Editable By AI |
|------|-------|----------------|
| spec.yaml | Human/AI | Yes |
| openapi.yaml | Human/AI | Yes |
| src/auth/*.ts | Human/AI | Yes |
| src/auth/*.conf | Human/AI | Yes |
| src/auth/*.csv | Human/AI | Yes |
| src/db/schema.ts | Human/AI | Yes |
| src/components/usecases/*.ts (body) | Human/AI | Yes (body only) |
| src/components/servers/*.ts | OpenBoundary | Never |
| src/middleware/*.ts | OpenBoundary | Never |
| tests/*.ts | OpenBoundary | Never |
