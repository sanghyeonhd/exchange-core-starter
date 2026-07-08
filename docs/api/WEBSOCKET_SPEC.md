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

