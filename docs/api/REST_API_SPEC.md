# REST API Spec

## Conventions

- Base path: `/api/v1`
- Request id header: `X-REQUEST-ID`
- API key header: `X-API-KEY`
- Signature header: `X-SIGNATURE`
- Timestamp header: `X-TIMESTAMP`
- Signature payload: `timestamp + method + path + body`
- Timestamp drift limit: 5 seconds by default
- Numeric financial values are decimal strings at API boundaries and fixed-point integers internally.

## Error Format

```json
{
  "code": "ORDER_REJECTED_INSUFFICIENT_BALANCE",
  "message": "Insufficient available balance",
  "request_id": "req_...",
  "details": {}
}
```

## Public Endpoints

| Method | Path | Description |
|---|---|---|
| GET | `/health` | Service health |
| GET | `/time` | Server time in milliseconds |
| GET | `/markets` | List markets |
| GET | `/markets/{symbol}` | Market configuration |
| GET | `/orderbook/{symbol}` | Current book snapshot |
| GET | `/trades/{symbol}` | Recent trades |
| GET | `/klines/{symbol}` | OHLCV candles |

## Trading Endpoints

| Method | Path | Description |
|---|---|---|
| POST | `/orders` | Submit order |
| DELETE | `/orders/{order_id}` | Cancel order |
| GET | `/orders/{order_id}` | Read order |
| GET | `/open-orders` | List open orders |
| GET | `/order-history` | List historical orders |

### Submit Order

```json
{
  "client_order_id": "my-order-001",
  "symbol": "BTC-USDT",
  "side": "BUY",
  "type": "LIMIT",
  "time_in_force": "GTC",
  "price": "50000.00",
  "quantity": "0.01000000",
  "post_only": false,
  "reduce_only": false
}
```

## Account Endpoints

| Method | Path | Description |
|---|---|---|
| GET | `/account/balances` | Balance projection |
| GET | `/account/ledger` | Ledger entries |
| POST | `/account/transfer` | Spot/futures internal transfer |

## Wallet Endpoints

| Method | Path | Description |
|---|---|---|
| GET | `/wallet/deposit-address/{asset}` | Get or create deposit address |
| POST | `/wallet/withdraw` | Request withdrawal |
| GET | `/wallet/withdrawals` | Withdrawal history |
| GET | `/wallet/deposits` | Deposit history |

Withdrawals are disabled by default in local development.

## Futures Endpoints

| Method | Path | Description |
|---|---|---|
| GET | `/futures/positions` | Open positions |
| POST | `/futures/leverage` | Set isolated leverage |
| GET | `/futures/funding-rate` | Current funding rate |
| GET | `/futures/funding-history` | Funding payments |

