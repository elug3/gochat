# SQLite Infrastructure Migration Plan

## Goal

Move auth-specific sqlite implementations from:

```text
internal/store/sqlite3
```

into:

```text
pkg/auth/infra/sqlite
```

This aligns storage ownership with the auth bounded context.

## Why

The current structure mixes:

- infrastructure ownership
- runtime concerns
- business domain boundaries

inside generic internal packages.

As auth evolves into a standalone service, storage should belong to the auth architecture itself.

## Target Structure

```text
pkg/auth/
├── domain/
├── service/
├── ports/
├── infra/
│   └── sqlite/
├── http/
└── runtime/
```

## Desired Dependency Direction

```text
service -> ports <- infra/sqlite
```

Important:

```text
services should not depend directly on sqlite
```

## Migration Steps

### Step 1

Create package skeletons.

### Step 2

Move auth-specific sqlite repositories.

### Step 3

Implement repository ports.

### Step 4

Inject interfaces into services.

### Step 5

Remove direct sqlite dependencies from services.

### Step 6

Retire old internal store paths.

## Important Rule

Infrastructure implementations belong to the bounded context they support.

Auth persistence belongs to:

```text
pkg/auth/infra/sqlite
```

not generic shared storage locations.
