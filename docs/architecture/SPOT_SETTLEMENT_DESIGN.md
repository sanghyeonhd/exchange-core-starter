# Spot Settlement Design

## Scope

The first settlement implementation converts one spot trade into one double-entry ledger transaction.

The request identifies buyer and seller base/quote accounts separately. This avoids relying on one account id for multiple assets and keeps settlement independent from account lookup.

## Direction Convention

Within exchange wallet accounts:

- `Debit` represents asset entering an account.
- `Credit` represents asset leaving an account.
- A transaction must balance per asset.

## Taker Buy

For a BTC-USDT trade where the buyer is the taker:

- Buyer debits BTC quantity.
- Seller credits BTC quantity.
- Seller debits quote notional minus maker fee.
- Fee revenue debits maker fee plus taker fee.
- Buyer credits quote notional plus taker fee.

If the buyer was reserved using a taker-fee upper bound but settles as maker, OMS/account projection releases the unused quote lock after settlement.

## Taker Sell

When the seller is the taker:

- Buyer fee uses maker fee rate.
- Seller fee uses taker fee rate.
- Quote asset entries remain balanced by charging buyer quote notional plus buyer fee and crediting seller proceeds plus fee revenue.

## Idempotency

Settlement idempotency key is `spot-trade:{trade_id}`. Database storage must enforce uniqueness on this key before retry-based workers are added.
