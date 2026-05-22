# Auth Service Decomposition

This document describes the intended service split for `pkg/auth`.

## Problem

`AuthService` currently owns too many unrelated responsibilities.

Examples:

- user registration
- password login
- session lifecycle
- JWT generation
- WebAuthn/passkeys
- token exchange
- websocket token handling
- event publishing

This creates a god-service.

## Goal

Split auth functionality into cohesive business services.

## Target Services

### AuthService

Responsibilities:

- registration
- password authentication
- identity validation

### SessionService

Responsibilities:

- session lifecycle
- token exchange
- refresh/revoke
- session validation

### PasskeyService

Responsibilities:

- WebAuthn registration
- WebAuthn login
- credential management

### TokenRuntimeService

Responsibilities:

- token signing
- token validation
- JWKS exposure
- JWT lifecycle

## Architectural Goal

Each service should own:

- one business capability
- one cohesive responsibility

Services should not own unrelated infrastructure concerns.

## Important Rule

Business services should depend on ports/interfaces only.

Example:

```text
service -> ports <- infra
```

NOT:

```text
service -> sqlite
service -> jose/jwt
service -> gin
```

## Migration Strategy

1. Create new service boundaries
2. Move methods incrementally
3. Update handlers
4. Remove unused dependencies from AuthService
5. Move infrastructure behind ports

This minimizes refactor risk while keeping the system functional.
