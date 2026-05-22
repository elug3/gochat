# Auth Package Layout Strategy

This document defines the intended package structure for the standalone auth service architecture.

## Goals

- separate business logic from infrastructure
- clarify runtime ownership
- improve testability
- support standalone deployment
- reduce coupling

## Package Structure

```text
pkg/auth/
├── domain/
├── service/
├── ports/
├── infra/
├── http/
└── runtime/
```

## Layer Responsibilities

### domain/

Contains:

- business entities
- domain rules
- auth errors
- domain invariants

Must not depend on infrastructure.

### service/

Contains:

- business workflows
- application services
- orchestration logic

Should depend only on:

- domain
- ports

### ports/

Contains interfaces required by business logic.

Examples:

- repositories
- token signers
- event publishers
- passkey providers

### infra/

Contains concrete implementations.

Examples:

- sqlite
- jwt
- webauthn
- nats
- redis

### http/

Contains transport adapters.

Examples:

- handlers
- middleware
- request/response translation

### runtime/

Contains process lifecycle management.

Examples:

- startup
- shutdown
- worker supervision
- dependency wiring

## Internal Packages

Internal packages remain appropriate for:

```text
internal/bootstrap
internal/config
internal/telemetry
```

These packages contain process-private implementation details.

## Dependency Direction

```text
service -> ports <- infra
```

Important:

```text
domain never imports infra
```
