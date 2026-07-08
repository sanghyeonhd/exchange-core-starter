# OMS Design

## Responsibilities

- Accept order requests from Gateway.
- Enforce idempotency using `user_id + market + client_order_id`.
- Validate market, user, price, quantity, notional, tick, and lot constraints.
- Request balance or margin reservation.
- Route commands to the correct symbol shard.
- Track order state transitions.

## Validation Order

1. Request schema is valid.
2. User is active and allowed to trade.
3. Market exists and status is `TRADING`.
4. Price and quantity match tick/lot configuration.
5. Notional meets minimum rules.
6. Duplicate client order id returns the existing order.
7. Spot balance or futures margin can be reserved.
8. Post-only/reduce-only constraints are checked before matching when possible.

## State Machine

```text
NEW
  -> VALIDATED
  -> REJECTED
  -> ACCEPTED
  -> PARTIALLY_FILLED
  -> FILLED
  -> CANCELED
  -> EXPIRED
```

## Reservation Model

Spot buy:

- Reserve quote notional plus taker fee upper bound.

Spot sell:

- Reserve base quantity.

Perpetual futures:

- Reserve initial margin plus fee reserve.
- Simulate position impact before accepting the command.

## Idempotency

The OMS stores an acceptance record before publishing to matching. If a client retries the same client order id, the existing order status is returned.

## Out of Scope for Phase 1

- Stop orders
- Trailing stops
- Cross margin
- Advanced self-trade prevention modes beyond the default policy hook

