# Phase Roadmap

The project roadmap has moved from MVP delivery to regulated exchange readiness. Use `docs/compliance/REGULATED_EXCHANGE_PHASES.md` as the primary roadmap.

This file is kept as an engineering implementation companion for the trading core.

## Legacy Phase 0: Research and Clean-Room Design

Status: in progress.

- Clone reference repositories.
- Record license and security risk.
- Produce architecture, API, database, event, and runbook documents.
- Avoid direct code reuse before license review.

## Legacy Phase 1: Minimal Spot Loop

Status: in progress.

Goal: seeded internal balances can trade BTC-USDT end to end.

- Market config
- Account projection: in-memory MVP implemented for tests
- Ledger storage
- OMS validation and reservation calculation: initial limit-order checks implemented
- Matching engine: in-memory price-time priority implemented
- Spot settlement: trade-to-ledger builder implemented
- Public/private WebSocket projections

Exit test:

```text
Alice USDT + Bob BTC -> buy/sell match -> settlement -> balanced ledger -> updated balances
```

Current integration coverage:

```text
Alice USDT 10000 + Bob BTC 1
  -> Alice limit buy reserves quote + max fee
  -> Bob limit sell reserves base
  -> Matching creates one trade
  -> Settlement builds balanced ledger transaction
  -> Account projection applies balances and releases excess reservation
  -> Duplicate settlement apply is idempotent
```

## Legacy Phase 2: Wallet MVP

- Mock wallet adapter
- Deposit address generation
- Mock deposit confirmation
- Withdrawal request and admin approval
- Mock broadcast
- Audit log

## Legacy Phase 3: Perpetual Futures MVP

- PERP markets
- Isolated one-way positions
- Leverage max 10x
- Mark price mock
- Funding payment mock
- Liquidation condition detection

## Legacy Phase 4: Risk and Liquidation

- Price bands
- Circuit breaker
- Partial liquidation
- Insurance fund ledger account
- Duplicate liquidation prevention

## Legacy Phase 5: Recovery

- Matching WAL
- Orderbook snapshots
- Event replay
- Settlement idempotent retry
- Recovery runbook test

## Legacy Phase 6: Performance

- Matching benchmark
- Load test
- p99 latency targets
- Hot symbol sharding review
