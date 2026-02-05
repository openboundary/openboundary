---
name: ob-new-usecase
description: Add a new usecase to the OpenBoundary specification
model: inherit
---

# Add New Usecase to Specification

Guide the user through adding a new usecase component to their OpenBoundary specification.

## Information Needed

Ask the user for:

1. **Usecase ID** - Following pattern `usecase.{action}-{resource}`
   - Examples: `usecase.create-user`, `usecase.get-order`, `usecase.delete-post`

2. **HTTP Binding** - Server, method, and path
   - Server ID (e.g., `http.server.api`)
   - Method: GET, POST, PUT, PATCH, DELETE
   - Path: e.g., `/users`, `/users/{id}`

3. **Security Requirements**
   - Public endpoint: `middleware: []`
   - Authenticated: `middleware: [middleware.authn]`
   - Authorized: `middleware: [middleware.authn, middleware.authz]`

4. **Goal** - What this usecase accomplishes (one sentence)

5. **Actor** - Who performs this action (optional)

6. **Acceptance Criteria** - What defines success (optional but recommended)

## Steps

### 1. Add to spec.yaml

Add the usecase component:

```yaml
- id: usecase.{id}
  kind: usecase
  spec:
    binds_to: {server}:{METHOD}:{path}
    middleware:
      - middleware.authn  # or [] for public
    goal: {goal description}
    actor: {actor}
    acceptance_criteria:
      - {criterion 1}
      - {criterion 2}
```

### 2. Add to openapi.yaml

Add the endpoint definition:

```yaml
paths:
  /path:
    {method}:
      operationId: {operationId}
      summary: {summary}
      requestBody:  # for POST/PUT/PATCH
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/{RequestType}'
      responses:
        '200':  # or 201 for POST
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/{ResponseType}'
```

### 3. Compile

After adding the usecase:

```bash
openboundary compile spec.yaml
```

### 4. Implement

The compiler generates a stub in `src/components/usecases/{usecase-id}.ts`.
Edit only the function body to implement the business logic.

## Example: Add "Update User" Usecase

### User Input

- ID: `usecase.update-user`
- Binding: `http.server.api:PUT:/users/{id}`
- Middleware: `[middleware.authn, middleware.authz]`
- Goal: Update a user's profile information
- Actor: authenticated_user

### spec.yaml Addition

```yaml
- id: usecase.update-user
  kind: usecase
  spec:
    binds_to: http.server.api:PUT:/users/{id}
    middleware:
      - middleware.authn
      - middleware.authz
    goal: Update a user's profile information
    actor: authenticated_user
    preconditions:
      - User is authenticated
      - User has permission to update the profile
    acceptance_criteria:
      - User record is updated with provided fields
      - Returns updated user data
    postconditions:
      - User record reflects new values
```

### openapi.yaml Addition

```yaml
/users/{id}:
  put:
    operationId: updateUser
    summary: Update user
    parameters:
      - name: id
        in: path
        required: true
        schema:
          type: string
    requestBody:
      required: true
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/UpdateUserRequest'
    responses:
      '200':
        description: OK
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UserResponse'
```

## Reminders

- Always add both spec.yaml and openapi.yaml entries
- Run compile after adding the usecase
- Only implement the function body in generated files
- Never edit the generated imports, types, or signatures
