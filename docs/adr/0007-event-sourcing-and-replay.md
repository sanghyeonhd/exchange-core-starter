# ADR 0007: Event Sourcing and Replay

## Status

Accepted

## Decision

Order commands, matching output, settlements, wallet state changes, and risk events use append-only event envelopes with schema versions.

## Rationale

Recovery and audit require reconstructing state from durable input and output logs. Deterministic replay is a release gate for matching and settlement.

## Consequences

- Events are immutable.
- Consumers must be idempotent.
- Snapshot plus WAL replay must rebuild the same orderbook for the same command sequence.

