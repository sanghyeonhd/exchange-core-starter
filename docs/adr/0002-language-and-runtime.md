# ADR 0002: Language and Runtime

## Status

Accepted

## Decision

Core services use Go 1.22 or newer.

## Rationale

Go supports small deployable binaries, simple concurrency, and mature NATS/gRPC integration. The MVP keeps the stack narrow while leaving room for separate admin services later.

## Consequences

- Core packages avoid floating-point values for financial quantities.
- Interfaces are kept small and testable.
- Service folders can become separate binaries as boundaries harden.

