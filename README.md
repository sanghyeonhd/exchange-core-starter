# exchange-core-starter

Independent CEX and perpetual futures exchange engine starter.

This repository is a clean-room implementation. Public exchange repositories are analyzed for architecture and risk, but production code is not copied without license review.

## Current Scope

- Regulated exchange roadmap: user web, independent admin web, backend, matching engine, wallet gateway, KYC, AML, Travel Rule, whitelists, ISMS/ISO evidence, and Korean VASP readiness.
- Trading core (implemented, in-memory): spot exchange composition (`services/oms/spotexchange`) wiring OMS validation, balance reservation, price-time matching, per-trade double-entry settlement, excess-reservation release, cancel, open orders, trades, and balances.
- Wallet flows (implemented, mock adapter): deposit address, confirmed deposits and broadcasted withdrawals settle as balanced idempotent ledger transactions; withdrawal requests lock amount+fee, audited approval/rejection, rejection releases the lock.
- Matching recovery (implemented): fsynced command WAL (`services/matching-engine/wal`), orderbook snapshot/restore with SHA-256 book hash, atomic checkpoints, and deterministic replay proven by `tests/replay` (full replay, snapshot + tail, repeated recovery).
- Market data WebSocket (implemented): `/ws/public` streams trades and orderbook snapshots with per-channel monotonic sequences (`services/market-data`), served over a dependency-free RFC 6455 implementation (`libs/ws`).
- Gateway REST API (implemented): public market data plus authenticated trading, account, and wallet endpoints following `docs/api/REST_API_SPEC.md`. Authentication is a development placeholder (`X-USER-ID` header); API key + HMAC signing is pending.
- User web shell (`apps/web`): connects to the local gateway with a mock-data fallback. Admin shell (`apps/admin`) remains static.
- Compliance gating: KYC/AML/listing decision packages, readiness gates for spot launch and derivatives, evidence register, RBAC, tamper-evident audit log.
- Overall design contracts: REST/WebSocket specs, proto interfaces, database DDL, event model, wallet/risk/futures design.

Not yet implemented: persistent account/ledger storage, snapshot scheduling and WAL rotation, orderbook deltas and ticker/kline/private streams, testnet wallet adapters, futures/margin/liquidation services, admin API backend.

## Safety Defaults

- Mainnet is disabled.
- Withdrawals are disabled by default.
- Wallet adapter default is mock.
- Prices, quantities, fees, and balances use fixed-point `int64`, never floating point. Decimal strings at API boundaries.
- Balance changes must be explainable by ledger entries; deposits/withdrawals settle against an external omnibus account so every transaction balances per asset.
- Derivatives remain blocked by the readiness gate until separate approval.

## Repository Layout

```
apps/          user and admin web product shells
services/      gateway, oms (+spotexchange), matching-engine, settlement (spot, funding),
               ledger, account, wallet-gateway, kyc, aml, compliance, auth, listing, risk, ...
libs/          decimal, audit, ws (RFC 6455), (others pending)
docs/          architecture, api, compliance, database, security, runbooks, adr
tests/         integration (spot loop, wallet funding, compliance gates)
```

## Local Commands

```bash
make test            # go test ./...
make fmt             # go fmt ./...
make vet             # go vet ./...
go run ./services/gateway/cmd/gateway   # REST API on :8080 (demo users 1 and 2 seeded)

# enable the matching command WAL (fsynced write-ahead log; reopening resumes the sequence)
MATCHING_WAL_PATH=./matching.wal go run ./services/gateway/cmd/gateway

go test ./tests/replay/                 # deterministic recovery gates (snapshot + WAL replay)

# user web shell against the local gateway
python3 -m http.server 8081 --directory apps/web   # open http://localhost:8081
```

Quick API tour (development auth placeholder):

```bash
curl localhost:8080/api/v1/markets
curl -X POST localhost:8080/api/v1/orders -H 'X-USER-ID: 1' \
  -d '{"client_order_id":"demo-1","symbol":"BTC-USDT","side":"BUY","type":"LIMIT","time_in_force":"GTC","price":"50000.00","quantity":"0.01000000"}'
curl localhost:8080/api/v1/orderbook/BTC-USDT
curl localhost:8080/api/v1/account/balances -H 'X-USER-ID: 1'
curl localhost:8080/api/v1/readiness/spot-launch
```

Public market data stream: connect a WebSocket client to `ws://localhost:8080/ws/public`. The first message is an orderbook snapshot (`sequence: 0`); trades and book updates follow with per-channel sequences.

## Design Entry Points

- [Architecture index](docs/architecture/README.md)
- [Phase roadmap (engineering companion)](docs/architecture/PHASE_ROADMAP.md)
- [Regulated exchange phases](docs/compliance/REGULATED_EXCHANGE_PHASES.md)
- [Korea regulated exchange requirements](docs/compliance/KR_REGULATED_EXCHANGE_REQUIREMENTS.md)
- [Control matrix](docs/compliance/CONTROL_MATRIX.md)
- [REST API spec](docs/api/REST_API_SPEC.md)
- [WebSocket spec](docs/api/WEBSOCKET_SPEC.md)
- [PostgreSQL schema](docs/database/POSTGRES_SCHEMA.sql)
- [Wallet security model](docs/security/WALLET_SECURITY_MODEL.md)
- [License and source policy](docs/security/LICENSE_AND_SOURCE_POLICY.md)

## Analysis Sources

Run `./scripts/clone-sources.sh` to clone the reference repositories into `~/exchange-lab/sources`, and `./scripts/audit-sources.sh` to record license and security risk.
