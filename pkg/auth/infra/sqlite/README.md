# SQLite Auth Infrastructure

This package contains sqlite-backed infrastructure implementations for the auth domain.

## Purpose

Move auth persistence implementations out of:

```text
internal/store/sqlite3
```

into:

```text
pkg/auth/infra/sqlite
```

This aligns persistence ownership with the auth bounded context.

## Responsibilities

This package should eventually contain:

- auth repositories
- session persistence
- outbox persistence
- migration helpers
- sqlite transaction helpers

## Architectural Role

This package belongs to the infrastructure layer.

Dependency direction:

```text
service -> ports <- infra/sqlite
```

Important:

```text
business logic should never import sqlite directly
```

## Migration Strategy

Migration should occur incrementally:

1. Move concrete sqlite implementations
2. Keep interfaces in `ports`
3. Update services to depend on ports only
4. Remove direct sqlite references from services
5. Eventually retire old internal store paths

## Compile-Time Validation

Repository implementations should validate port compatibility:

```go
var _ ports.UserRepository = (*Store)(nil)
```
