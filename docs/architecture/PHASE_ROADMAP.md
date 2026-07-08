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

Status: core loop implemented in-memory and exposed over REST.

Goal: seeded internal balances can trade BTC-USDT end to end.

- Market config: static single-market config in the spot exchange composition
- Account projection: in-memory store with reservation, release, and idempotent transaction apply
- Ledger storage: in-memory entries behind balanced-transaction validation (persistence pending)
- OMS validation and reservation calculation: limit-order checks and buy-reservation helper implemented
- Matching engine: in-memory price-time priority implemented, with depth snapshot accessor
- Spot settlement: trade-to-ledger builder implemented
- Composition: `services/oms/spotexchange` wires order placement, reservation, matching, per-trade settlement, excess-reservation release, cancel, open orders, trades, and balances; served by the gateway REST API
- Public WebSocket projections: implemented — `/ws/public` streams trades and orderbook snapshots with per-channel sequences (`services/market-data`, `libs/ws`); private streams pending

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

Status: implemented against the mock adapter with full ledger integration.

- Mock wallet adapter
- Deposit address generation
- Mock deposit confirmation: creates a balanced deposit ledger transaction (idempotent per deposit id)
- Withdrawal request and admin approval: request locks amount+fee, approval and rejection are audited, rejection releases the lock
- Mock broadcast: settles the withdrawal into ledger entries (idempotent per withdrawal id)
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

Status: matching-side recovery implemented; durable account/ledger storage pending.

- Matching WAL: JSON-lines command log with fsync-before-ack (`services/matching-engine/wal`), enabled in the gateway via `MATCHING_WAL_PATH`
- Orderbook snapshots: `engine.Snapshot`/`engine.RestoreOrderBook` with SHA-256 book hash; `spotexchange.Checkpoint()` pairs a snapshot with the WAL sequence
- Event replay: `wal.Replay` with divergence detection; deterministic replay tests in `tests/replay` (full replay, snapshot + tail, repeated recovery)
- Settlement idempotent retry: idempotency keys enforced in account projection (duplicate apply is a no-op)
- Recovery runbook test: replay gates covered by `go test ./tests/replay/`; operational drill pending

## Legacy Phase 6: Performance

- Matching benchmark
- Load test
- p99 latency targets
- Hot symbol sharding review
