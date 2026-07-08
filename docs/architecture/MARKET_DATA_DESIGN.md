# Market Data Design

## Responsibilities

- Consume matching engine `TradeCreated` and `BookDelta` events.
- Maintain recent trades.
- Build orderbook snapshots and deltas.
- Calculate ticker values.
- Aggregate OHLCV klines.
- Publish public WebSocket streams.

## Streams

| Stream | Source | Persistence |
|---|---|---|
| Trades | `TradeCreated` | PostgreSQL archive, cache for recent reads |
| Orderbook delta | `BookDelta` | Ephemeral plus periodic snapshot |
| Ticker | Trades | Cache/projection |
| Kline | Trades | TimescaleDB/ClickHouse later, PostgreSQL for MVP |

## Snapshot Rule

Clients must load a snapshot before applying deltas. Deltas include `prev_sequence` and `sequence`; a gap requires resubscription.

## Kline Intervals

`1m`, `3m`, `5m`, `15m`, `30m`, `1h`, `4h`, `1d`, `1w`, `1M`.

## Data Integrity

- Market data is derived and can be rebuilt.
- Matching trades are the source of truth for public trade history.
- Kline corrections must be traceable to source trade events.

