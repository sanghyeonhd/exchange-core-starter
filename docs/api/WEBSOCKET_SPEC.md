# WebSocket Spec

## Connections

- Public endpoint: `/ws/public`
- Private endpoint: `/ws/private`
- Private connections require the same API key/HMAC model as REST during connection setup.

## Envelope

```json
{
  "channel": "trades",
  "symbol": "BTC-USDT",
  "sequence": 12345,
  "timestamp": 1720000000000,
  "data": {}
}
```

## Public Channels

| Channel | Description |
|---|---|
| `ticker` | Last price, volume, high, low |
| `orderbook` | Snapshot and deltas |
| `trades` | Public trade stream |
| `kline` | OHLCV candles |

## Private Channels

| Channel | Description |
|---|---|
| `orders` | Order state changes |
| `fills` | User fills |
| `balances` | Balance projection updates |
| `positions` | Futures position changes |
| `liquidation` | Liquidation alerts |

## Sequencing

- Each channel message includes a monotonically increasing sequence.
- Orderbook deltas are only valid after a fresh snapshot.
- Clients must resubscribe when a sequence gap is detected.

## Example Trade Message

```json
{
  "channel": "trades",
  "symbol": "BTC-USDT",
  "sequence": 12345,
  "timestamp": 1720000000000,
  "data": {
    "price": "50000.00",
    "quantity": "0.01000000",
    "side": "BUY"
  }
}
```

## Example Ticker Message

```json
{
  "channel": "ticker",
  "symbol": "BTC-USDT",
  "sequence": 42,
  "timestamp": 1720000000000,
  "data": {
    "last_price": "50000.00",
    "volume_24h": "1.50000000",
    "high_24h": "51000.00",
    "low_24h": "49000.00",
    "price_change": "1000.00",
    "price_change_pct": "2.04"
  }
}
```

## Example Kline Message

```json
{
  "channel": "kline",
  "symbol": "BTC-USDT",
  "sequence": 7,
  "timestamp": 1720000000000,
  "data": {
    "interval": "1m",
    "open": "50000.00",
    "high": "51000.00",
    "low": "49500.00",
    "close": "50500.00",
    "volume": "0.50000000",
    "open_time": 1720000000000,
    "close_time": 1720000060000
  }
}
```
