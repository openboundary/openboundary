---
name: ob-validate
description: Validate OpenBoundary specification without generating code
model: inherit
---

# Validate OpenBoundary Specification

Validate the specification file without generating any code. Useful for checking syntax and semantic correctness before compilation.

## Steps

1. **Find the specification file**

   Look for `spec.yaml` in the project root or the path provided by the user.

2. **Run validation**

   ```bash
   openboundary validate spec.yaml
   ```

   Or with a custom path:

   ```bash
   openboundary validate path/to/spec.yaml
   ```

3. **Report results**

   Show the user:
   - Validation status (pass/fail)
   - Any warnings or errors
   - Suggestions for fixes

## What Gets Validated

### Schema Validation

- Required fields present (version, name, components)
- Component IDs follow naming convention
- Valid component kinds (http.server, middleware, postgres, usecase)
- Spec fields match component kind

### Reference Validation

- Middleware references resolve to defined components
- Server references in usecase bindings are valid
- Dependency chains are acyclic

### Semantic Validation

- Usecases bind to existing server routes
- Middleware dependencies are satisfied
- Config file paths exist

## Error Handling

If validation fails:

1. Show all validation errors
2. Point to the specific location in spec.yaml
3. Explain what the error means
4. Suggest the fix

## Common Errors

### Invalid Component Reference

```text
Error: middleware.authn referenced but not defined
```

Fix: Add the missing middleware component to spec.yaml

### Missing Required Field

```text
Error: usecase.create-user missing required field: goal
```

Fix: Add the goal field to the usecase spec

### Invalid Binding Format

```text
Error: invalid binds_to format, expected server-id:METHOD:/path
```

Fix: Use format like `http.server.api:POST:/users`
