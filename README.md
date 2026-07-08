# exchange-core-starter

Independent CEX and perpetual futures exchange engine starter.

This repository is a clean-room implementation. Public exchange repositories are analyzed for architecture and risk, but production code is not copied without license review.

## Current Scope

- Phase 0: source clone/audit scripts and architecture decisions.
- Phase 1: deterministic spot orderbook matching with fixed-point integer values.
- Overall design contracts: REST/WebSocket specs, proto interfaces, database DDL, event model, wallet/risk/futures design.

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

## Design Entry Points

- [Architecture index](docs/architecture/README.md)
- [REST API spec](docs/api/REST_API_SPEC.md)
- [WebSocket spec](docs/api/WEBSOCKET_SPEC.md)
- [PostgreSQL schema](docs/database/POSTGRES_SCHEMA.sql)
- [Wallet security model](docs/security/WALLET_SECURITY_MODEL.md)
- [License and source policy](docs/security/LICENSE_AND_SOURCE_POLICY.md)

## Analysis Sources

Run `./scripts/clone-sources.sh` to clone the reference repositories into `~/exchange-lab/sources`.
