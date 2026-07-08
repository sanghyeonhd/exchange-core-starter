# ADR 0002: Language and Runtime

## Status

Accepted

## Decision

Core services use Go 1.25.12 or newer.

## Rationale

Go supports small deployable binaries, simple concurrency, and mature NATS/gRPC integration. The project uses Go 1.25.12+ so CI vulnerability checks run on a standard library version that includes current security fixes.

## Consequences

- Core packages avoid floating-point values for financial quantities.
- Interfaces are kept small and testable.
- Service folders can become separate binaries as boundaries harden.
