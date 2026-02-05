---
name: ob-rollback
description: Rollback unauthorized changes to protected OpenBoundary files
model: inherit
---

# Rollback Boundary Violations

Undo changes to protected files that should not have been modified.

## When to Use

- After a boundary violation is detected
- When generated files were accidentally edited
- To restore protected infrastructure code

## Steps

### Step 1: Detect Violations

Check for modified protected files:

```bash
git diff --name-only | grep -E "(src/components/servers/|src/middleware/|\.generated\.ts$|src/components/usecases/(index|schemas|types)\.ts)"
```

### Step 2: Show What Will Be Rolled Back

List the files that will be restored:

```text
The following files will be restored to their last committed state:

  - src/components/servers/http-server-api.ts
  - src/middleware/authn.ts

This will DISCARD any changes made to these files.
```

### Step 3: Confirm with User

Ask for confirmation before proceeding:

```text
Proceed with rollback? This cannot be undone.
```

### Step 4: Execute Rollback

For each protected file:

```bash
git checkout -- <filename>
```

### Step 5: Verify Rollback

Confirm files are restored:

```bash
git diff --name-only | grep -E "(src/components/servers/|src/middleware/)" || echo "Rollback complete. No protected files modified."
```

## Full Rollback Option

If user wants to rollback ALL changes:

```bash
git checkout -- .
```

**WARNING:** This discards ALL uncommitted changes, including safe zone files.

## Selective Rollback

To rollback only specific patterns:

```bash
# Rollback all server files
git checkout -- src/components/servers/

# Rollback all middleware
git checkout -- src/middleware/

# Rollback all generated files
git checkout -- '*.generated.ts'
```

## Preserving Work

If the user wants to keep their changes for reference:

```bash
# Create a backup branch
git stash push -m "boundary-violation-backup"

# Or create a patch file
git diff > ~/boundary-violation-backup.patch
```

Then proceed with rollback.

## After Rollback

Remind the user:

```text
Rollback complete. Protected files restored.

To make infrastructure changes:
1. Edit spec.yaml with your requirements
2. Run: /ob-compile
3. Implement business logic in usecase bodies only

Your changes to safe zone files (spec.yaml, auth config, schema) are preserved.
```

## Error Handling

If not a git repository:

```text
ERROR: Not a git repository. Cannot rollback.

To manually restore files, you'll need to:
1. Delete the modified generated files
2. Run: openboundary compile spec.yaml
```

If git checkout fails:

```text
ERROR: Could not restore {filename}

The file may have been deleted or renamed. Try:
1. git status to see file state
2. git checkout HEAD -- {filename} for specific commit
3. openboundary compile spec.yaml to regenerate
```
