# Auth Infrastructure Layer

This directory contains concrete infrastructure implementations used by `pkg/auth`.

The infra layer should contain:

- sqlite repositories
- JWT signing implementations
- NATS/Kafka publishers
- WebAuthn providers
- cache implementations
- external SDK integrations

## Dependency Direction

Infrastructure depends on ports.

```text
service -> ports <- infra
```

Business logic should never depend directly on:

- sqlite
- gin
- jose/jwt
- NATS
- Redis
- WebAuthn SDKs

## Intended Layout

```text
infra/
├── sqlite/
├── jwt/
├── nats/
├── webauthn/
└── memory/
```

## Example

```go
var _ ports.UserRepository = (*Store)(nil)
```

This provides compile-time validation that an infrastructure implementation satisfies a port interface.
