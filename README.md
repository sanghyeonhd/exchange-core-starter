# exchange-core-starter

Independent CEX and perpetual futures exchange engine starter.

This repository is a clean-room implementation. Public exchange repositories are analyzed for architecture and risk, but production code is not copied without license review.

## Current Scope

- Phase 0: source clone/audit scripts and architecture decisions.
- Phase 1: deterministic spot orderbook matching with fixed-point integer values.

## Safety Defaults

- Mainnet is disabled.
- Wallet adapter default is mock.
- Prices, quantities, fees, and balances use fixed-point `int64`, never floating point.
- Balance changes must be explainable by ledger entries.

## Local Commands

```bash
make test
make clone-sources
make audit-sources
```

## Analysis Sources

Run `./scripts/clone-sources.sh` to clone the reference repositories into `~/exchange-lab/sources`.

