# Risk and Liquidation Design

## Risk Engine Responsibilities

- Enforce price bands.
- Enforce max order size and open order limits.
- Enforce max leverage by market.
- Provide self-trade prevention policy.
- Monitor account risk and abnormal trading hooks.
- Emit `RiskRuleTriggered` events.

## MVP Rules

| Rule | Default |
|---|---|
| Self-trade prevention | `CANCEL_TAKER` |
| Max leverage | 10x |
| Margin mode | Isolated |
| Circuit breaker | Configured but initially manual |
| Price band | Mark/index percentage band |

## Liquidation Engine Responsibilities

- Detect margin ratio breaches from risk/position data.
- Lock affected position.
- Cancel open orders for the symbol.
- Submit liquidation order.
- Apply partial liquidation before full liquidation.
- Use insurance fund for residual losses.
- Emit ADL trigger only after insurance fund insufficiency.

## Liquidation Flow

```text
Risk breach detected
  -> position lock
  -> open order cancellation
  -> liquidation order creation
  -> execution
  -> position/ledger settlement
  -> insurance fund if needed
  -> ADL if insurance is insufficient
```

## Safety Constraints

- Liquidation events are auditable.
- Liquidation settlement is idempotent by liquidation id.
- Insurance fund is a ledger account.
- No hidden balance mutation is allowed.

