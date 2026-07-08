# Architecture Index

Read these documents in order when onboarding to the exchange-core-starter design.

1. `TARGET_ARCHITECTURE.md` - overall system target and critical data flow.
2. `SERVICE_BOUNDARIES.md` - service ownership, sync calls, async events, and failure isolation.
3. `EVENT_MODEL.md` - event envelope, idempotency, retention, and replay rules.
4. `MATCHING_ENGINE_DESIGN.md` - spot matching rules and recovery direction.
5. `OMS_DESIGN.md` - order validation, state machine, reservation, and idempotency.
6. `SPOT_SETTLEMENT_DESIGN.md` - trade-to-ledger transaction design.
7. `WALLET_GATEWAY_DESIGN.md` - adapter model, deposit/withdrawal lifecycle, and mainnet gate.
8. `FUTURES_ENGINE_DESIGN.md` - isolated perpetual futures MVP.
9. `RISK_LIQUIDATION_DESIGN.md` - risk rules, liquidation flow, and insurance safety.
10. `MARKET_DATA_DESIGN.md` - public stream projections and snapshot rules.
11. `OPERABILITY_MODEL.md` - metrics, logs, and alerts.
12. `PHASE_ROADMAP.md` - legacy trading-core implementation companion.

The ADRs in `docs/adr` record why the highest-impact architecture decisions were made.

## Regulated Exchange Readiness

Use these as the primary roadmap and control references:

1. `../compliance/REGULATED_EXCHANGE_PHASES.md` - regulated exchange phase plan.
2. `../compliance/KR_REGULATED_EXCHANGE_REQUIREMENTS.md` - Korea-focused regulatory baseline.
3. `../compliance/CONTROL_MATRIX.md` - ISMS/ISO/compliance evidence mapping.
4. `../compliance/LAUNCH_READINESS_CHECKLIST.md` - production launch, fiat, custody, and derivatives gates.

## Application Boundaries

- `apps/web` - user-facing frontend.
- `apps/admin` - independent admin frontend.
- `services/kyc` - identity verification state and review workflow.
- `services/aml` - customer/transaction risk and case management.
- `services/compliance` - consolidated policy decision and evidence service.
- `services/listing` - asset, network, market, counterparty, and withdrawal whitelist controls.
