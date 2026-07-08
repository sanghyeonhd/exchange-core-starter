# Service Boundaries

## Boundary Rules

- Gateway owns external protocol concerns only.
- OMS owns order acceptance, idempotency, validation, and reservation requests.
- Matching engine owns price-time priority only.
- Settlement owns conversion from executions to ledger transactions.
- Ledger owns financial truth and integrity verification.
- Wallet gateway owns chain/custody adapter abstraction, not user balance mutation.
- Futures services own position, margin, funding, risk, and liquidation state.

## Synchronous Calls

| Caller | Callee | Reason |
|---|---|---|
| Gateway | Auth | Validate user/API key credentials |
| Gateway | OMS | Submit and cancel orders |
| Gateway | Account | Read balances and ledger history |
| Gateway | Wallet Gateway | Request deposit address or withdrawal |
| OMS | Account | Reserve spot balances |
| OMS | Margin | Reserve futures margin |
| Settlement | Ledger | Create balanced transactions |
| Wallet Gateway | Ledger | Credit confirmed deposits, finalize withdrawals |

## Asynchronous Events

| Producer | Consumer | Event |
|---|---|---|
| OMS | Matching Engine | `OrderCommand` |
| Matching Engine | Settlement, Market Data | `TradeCreated`, `BookDelta` |
| Settlement | Account, Market Data, Gateway | `SettlementCompleted` |
| Ledger | Account, Audit | `LedgerTransactionCreated` |
| Wallet Gateway | Account, Gateway, Audit | `DepositConfirmed`, `WithdrawalStateChanged` |
| Funding | Settlement, Ledger | `FundingPaymentCalculated` |
| Risk | Liquidation, Admin | `RiskRuleTriggered` |
| Liquidation | OMS, Settlement, Insurance | `LiquidationTriggered` |

## Ownership Matrix

| Data | Owner | Read Models |
|---|---|---|
| Users/API keys | Auth | Gateway, Admin |
| Markets | Admin/Market Config | OMS, Matching, Gateway |
| Open orders | OMS/Matching | Gateway, Market Data |
| Trades | Matching output, Settlement finalization | Market Data, Account |
| Balances | Ledger | Account |
| Deposit/withdrawal state | Wallet Gateway | Account, Admin |
| Positions | Position | Gateway, Risk |
| Risk rules | Risk/Admin | OMS, Liquidation |

## Failure Isolation

- A matching engine outage affects one symbol shard where possible.
- Settlement retry must not duplicate ledger entries.
- Market data lag must not block matching.
- Wallet scanners may pause without affecting trading balances, except pending deposits/withdrawals.
- Admin API cannot bypass ledger or signer separation.

