# Futures Engine Design

## MVP Scope

- Market type: perpetual futures
- Position mode: one-way
- Margin mode: isolated
- Maximum leverage: 10x
- Order types: limit, market, reduce-only
- Price source: mock index and mark price providers

## Position Model

```text
Position
- user_id
- market
- side: LONG | SHORT | NET
- size
- entry_price
- isolated_margin
- leverage
- realized_pnl
- status
```

## Order Flow

1. OMS validates market status and order constraints.
2. Margin service simulates position impact.
3. Margin service locks required initial margin plus fee reserve.
4. Matching engine executes normal price-time priority.
5. Settlement updates position, realized PnL, fees, and ledger entries.

## PnL

Long:

```text
unrealized_pnl = (mark_price - entry_price) * size
```

Short:

```text
unrealized_pnl = (entry_price - mark_price) * size
```

All values use fixed-point integers.

## Funding

- Default funding interval: 8 hours
- Local development interval: 5 minutes
- Funding starts with mock mark/index providers.
- Funding payments are ledger transactions and must be idempotent by position/funding time.

## Liquidation

Liquidation is triggered when the position margin ratio reaches the maintenance threshold.

Initial implementation uses partial liquidation. Insurance fund usage and ADL are explicit events, not hidden balance adjustments.

