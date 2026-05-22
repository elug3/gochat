# Auth Server Lifecycle Ownership

This document describes the intended runtime ownership model for `pkg/auth`.

## Goal

Run the auth package as a fully standalone service.

The auth server should:

- own all long-lived resources
- supervise all goroutines
- coordinate graceful shutdown
- clean up dependencies in reverse order
- avoid hidden background loops

## Current Problems

### Resource ownership is split

`NewServerDeps()` currently creates:

- database connections
- event publisher
- outbox worker
- HTTP server
- auth service

But `Server` does not explicitly own or close all of them.

This makes graceful shutdown difficult.

### Worker lifecycle is implicit

The outbox worker is attached to `AuthService`, but its runtime ownership is unclear.

Workers should instead be supervised directly by `Server`.

### Constructors partially manage runtime behavior

Constructors should only:

- validate
- allocate
- wire dependencies

Constructors should not:

- start goroutines
- manage signals
- self-register shutdown hooks

## Recommended Architecture

```text
main
 └── Server
      ├── HTTP server
      ├── Auth service
      ├── database
      ├── event publisher
      ├── outbox worker
      ├── goroutine supervisor
      └── shutdown coordinator
```

## Recommended Ownership Rules

### Server owns runtime lifecycle

`Server` should:

- start goroutines
- stop goroutines
- coordinate shutdown
- close resources
- aggregate runtime errors

### Services remain passive

`AuthService` should provide business logic only.

It should not:

- spawn background goroutines
- manage process lifetime
- own signal handling

## Suggested Shutdown Order

Resources should shut down in reverse order.

Example:

```text
HTTP
  -> worker
      -> publisher
          -> database
```

Shutdown order:

```text
HTTP
worker
publisher
database
```

## Suggested Next Refactor

1. Move worker ownership from `AuthService` into `Server`
2. Add explicit `startWorker()` and `shutdown()` methods
3. Move cleanup handling into `Server`
4. Replace concrete store types with interfaces
5. Add health/readiness endpoints
6. Add structured config loading from env vars

## Desired Runtime Shape

The final service should behave similarly to:

- Kubernetes-ready Go services
- systemd-managed daemons
- containerized microservices

with explicit lifecycle management and graceful shutdown behavior.
