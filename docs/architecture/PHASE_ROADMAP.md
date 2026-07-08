# Phase Roadmap

## Phase 0: Research and Clean-Room Design

Status: in progress.

- Clone reference repositories.
- Record license and security risk.
- Produce architecture, API, database, event, and runbook documents.
- Avoid direct code reuse before license review.

## Phase 1: Minimal Spot Loop

Goal: seeded internal balances can trade BTC-USDT end to end.

- Market config
- Account projection
- Ledger storage
- OMS validation and idempotency
- Matching engine
- Spot settlement
- Public/private WebSocket projections

Exit test:

```text
Alice USDT + Bob BTC -> buy/sell match -> settlement -> balanced ledger -> updated balances
```

## Phase 2: Wallet MVP

- Mock wallet adapter
- Deposit address generation
- Mock deposit confirmation
- Withdrawal request and admin approval
- Mock broadcast
- Audit log

## Phase 3: Perpetual Futures MVP

- PERP markets
- Isolated one-way positions
- Leverage max 10x
- Mark price mock
- Funding payment mock
- Liquidation condition detection

## Phase 4: Risk and Liquidation

- Price bands
- Circuit breaker
- Partial liquidation
- Insurance fund ledger account
- Duplicate liquidation prevention

## Phase 5: Recovery

- Matching WAL
- Orderbook snapshots
- Event replay
- Settlement idempotent retry
- Recovery runbook test

## Phase 6: Performance

- Matching benchmark
- Load test
- p99 latency targets
- Hot symbol sharding review

