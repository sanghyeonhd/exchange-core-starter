# ADR 0006: Futures Margin Model

## Status

Accepted

## Decision

The first perpetual futures MVP uses one-way positions, isolated margin, and a conservative maximum leverage of 10x.

## Rationale

Cross margin and hedge mode increase accounting and liquidation complexity. Isolated one-way positions are enough to validate the position, funding, margin, and liquidation loop.

## Consequences

- Cross margin is out of scope for the first futures MVP.
- `reduce_only` support is required before public futures trading.
- Funding and mark price start from mock providers.

