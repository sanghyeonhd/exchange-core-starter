# ADR 0005: Matching Engine Sharding

## Status

Accepted

## Decision

Matching is sharded by symbol. Each symbol has one ordered command stream and one single-writer orderbook.

## Rationale

Price-time priority must be deterministic. A single writer per symbol avoids locks in the core matching path and gives straightforward replay semantics.

## Consequences

- Commands are sequenced before entering a symbol engine.
- The matching engine emits events and never settles balances itself.
- WAL and snapshots are mandatory before production use.

