# Event Model

## Envelope

```json
{
  "event_id": "evt_...",
  "event_type": "TradeCreated",
  "aggregate_type": "Order",
  "aggregate_id": "ord_...",
  "sequence": 12345,
  "timestamp": 1720000000000,
  "idempotency_key": "spot-trade:trd_...",
  "schema_version": 1,
  "payload": {}
}
```

## Event Rules

- Events are append-only.
- Consumers must tolerate duplicate delivery.
- Every event schema has a version.
- Financial events include an idempotency key.
- Event order is guaranteed only within the aggregate stream or symbol stream explicitly named by the producer.

## Core Events

| Event | Producer | Idempotency |
|---|---|---|
| `UserCreated` | Auth | user id |
| `ApiKeyCreated` | Auth | API key id |
| `MarketCreated` | Admin | symbol |
| `OrderSubmitted` | Gateway/OMS | user + market + client order id |
| `OrderAccepted` | OMS | order id |
| `OrderRejected` | OMS | order id |
| `OrderCanceled` | Matching Engine | order id + cancel command id |
| `TradeCreated` | Matching Engine | market + match sequence |
| `SettlementCompleted` | Settlement | trade id |
| `LedgerTransactionCreated` | Ledger | ledger transaction idempotency key |
| `DepositDetected` | Wallet Gateway | network + txid + address + memo |
| `DepositConfirmed` | Wallet Gateway | deposit id |
| `WithdrawalRequested` | Wallet Gateway | withdrawal id |
| `WithdrawalApproved` | Admin/Wallet Gateway | withdrawal id + approval id |
| `WithdrawalBroadcasted` | Wallet Gateway | withdrawal id |
| `PositionUpdated` | Position | settlement id |
| `FundingPaymentSettled` | Funding/Settlement | position id + funding time |
| `LiquidationTriggered` | Liquidation | position id + liquidation sequence |
| `InsuranceFundUsed` | Insurance | liquidation id |
| `RiskRuleTriggered` | Risk | rule id + aggregate id + sequence |

## Retention

- Command WAL: retained until snapshots and archived events are verified.
- Financial event archive: retained indefinitely unless legal retention policy says otherwise.
- Market data derived events may have shorter retention after OHLCV aggregation.

## Replay Gates

- Matching replay compares final book state, trade list, and sequence.
- Settlement replay compares ledger idempotency keys and per-asset totals.
- Wallet replay never rebroadcasts withdrawals without chain/database reconciliation.

