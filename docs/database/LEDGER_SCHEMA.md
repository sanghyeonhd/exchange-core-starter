# Ledger Schema

The ledger is the source of truth for asset movement.

## Invariants

- Every transaction has at least one debit and one credit side.
- `sum(debit) == sum(credit)` for each asset in the transaction.
- An entry cannot have both debit and credit populated.
- Balance projections may be cached, but balance changes must be explained by entries.
- Idempotency keys are mandatory for externally triggered financial workflows.

## MVP Tables

See `exchange_engine_codex_development_guideline.md` for the initial PostgreSQL schema. The Go package at `services/ledger/ledger` currently implements the transaction invariant checks that database writes must enforce.

The test-only account projection at `services/account/account` applies ledger entries with `Debit` increasing account balance and `Credit` decreasing locked balance first, then available balance. Production storage must preserve the same semantics with database transactions and row-level locking.
