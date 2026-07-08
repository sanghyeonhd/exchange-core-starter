# Operability Model

## Metrics

| Metric | Owner |
|---|---|
| `orders_received_total` | Gateway/OMS |
| `orders_rejected_total` | OMS |
| `orders_matched_total` | Matching Engine |
| `match_latency_seconds` | Matching Engine |
| `settlement_latency_seconds` | Settlement |
| `ledger_integrity_failures_total` | Ledger |
| `withdrawal_requests_total` | Wallet Gateway |
| `withdrawal_failures_total` | Wallet Gateway |
| `liquidations_total` | Liquidation |
| `funding_payments_total` | Funding |
| `ws_connections` | Gateway |
| `api_rate_limited_total` | Gateway |

## Log Fields

```json
{
  "timestamp": "...",
  "level": "INFO",
  "service": "matching-engine",
  "request_id": "req_...",
  "user_id": "u_...",
  "event_id": "evt_...",
  "message": "..."
}
```

## Alert Rules

- Ledger integrity failure
- Matching engine sequence gap
- Settlement retry spike
- Withdrawal broadcast failure
- Hot wallet low balance
- Abnormal withdrawal spike
- Index price provider failure
- Liquidation spike
- Event replay lag

