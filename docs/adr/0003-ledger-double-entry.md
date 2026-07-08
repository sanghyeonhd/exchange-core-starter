# ADR 0003: Ledger Double Entry

## Status

Accepted

## Decision

All asset movements are represented as balanced double-entry ledger transactions.

## Rationale

Manual balance mutation is not auditable enough for an exchange. Every debit and credit must share one transaction id and satisfy `sum(debit) == sum(credit)`.

## Consequences

- Balance rows are projections, not the source of truth.
- Trade settlement, deposits, withdrawals, fees, funding, liquidation, and insurance fund changes use the same ledger model.
- Ledger APIs must be idempotent by transaction key.

