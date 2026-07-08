# Target Architecture

## Goal

Build a clean-room CEX and perpetual futures engine with deterministic matching, double-entry settlement, mock-first wallet integrations, and replayable operational state.

## Service Boundaries

| Service | Responsibility |
|---|---|
| Gateway | REST, WebSocket, API key auth, HMAC verification, rate limits |
| Auth | Users, password hashing, API keys, permission scopes |
| Account | Account projection, available/locked balances, transfers |
| Ledger | Double-entry transactions and integrity verification |
| OMS | Order validation, idempotency, balance/margin reservation |
| Matching Engine | Symbol-sharded price-time priority matching |
| Settlement | Trade settlement, fee calculation, ledger transaction creation |
| Market Data | Orderbook, trade, ticker, and kline streams |
| Wallet Gateway | Mock/testnet/full-node/explorer/custody wallet adapters |
| Position | Perpetual futures positions and PnL |
| Margin | Initial/maintenance margin, leverage limits |
| Risk | Price bands, self-trade prevention, circuit breakers |
| Liquidation | Partial liquidation, insurance fund usage, ADL hooks |
| Funding | Index price, mark price, funding rate/payment |
| Admin API | Market, wallet approval, user lock, audit log operations |

## Critical Data Flow

1. Gateway authenticates and normalizes a request.
2. OMS validates market/user/account constraints and applies idempotency.
3. Matching engine accepts sequenced commands per symbol and emits trades.
4. Settlement consumes trades exactly once and creates ledger transactions.
5. Ledger verifies balanced debit/credit entries and updates projections.
6. Market data publishes public streams; private streams publish order/fill/balance events.

## Non-Negotiable Constraints

- No `float32` or `float64` for money, price, quantity, or fee.
- Matching hot path performs no database or network calls.
- Mainnet wallets are disabled until a separate production security review.
- Every externally triggered financial action needs an idempotency key.
- Event replay must reconstruct matching and settlement state deterministically.

