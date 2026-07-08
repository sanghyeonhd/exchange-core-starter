# ADR 0001: Architecture Principles

## Status

Accepted

## Decision

The engine is built as a clean-room, event-driven exchange core with deterministic matching, double-entry accounting, append-only events, and explicit recovery paths.

## Rationale

Exchange correctness depends on replayable order flow and ledger integrity more than on UI velocity. Source repositories may inform boundaries and tradeoffs, but production logic is independently implemented unless license review permits reuse.

## Consequences

- Matching services do not call databases or remote services on the hot path.
- Balance mutations are only valid through ledger transactions.
- Wallet integrations start with mock/testnet adapters.
- Every financial workflow needs idempotency and replay review.

