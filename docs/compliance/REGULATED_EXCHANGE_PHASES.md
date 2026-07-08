# Regulated Exchange Phase Plan

This replaces the earlier MVP-oriented roadmap. The project now targets a compliance-ready exchange platform with separate user frontend, admin frontend, backend services, matching engine, wallet gateway, KYC, AML, whitelist controls, and certification evidence.

## Phase 0: Regulatory Scope and Product Classification

Goal: define what can legally be built and tested before any production launch.

Deliverables:

- Jurisdiction matrix: Korea first, international later.
- Product classification: spot, custody, fiat on/off ramp, futures/perpetuals.
- Explicit disabled-by-default list for futures, leverage, fiat, mainnet withdrawal.
- Legal/compliance owner map.
- Regulatory source register in `docs/compliance`.

Exit criteria:

- Written decision on spot-only first regulated path.
- Written decision that futures/perpetuals require separate review.
- Written decision on KRW market gate and real-name account dependency.

## Phase 1: Regulated Target Architecture

Goal: design the full platform before expanding code.

Workstreams:

- User Web App: markets, order entry, account, wallet, KYC status.
- Admin Web App: user review, KYC review, AML alerts, withdrawal approval, listing control, risk settings.
- Backend Gateway: public/private APIs, auth, HMAC, rate limits.
- Core Trading: OMS, matching engine, settlement, ledger, market data.
- Custody: wallet gateway, signer isolation, cold/hot policies.
- Compliance: KYC, AML, Travel Rule, whitelists, audit evidence.
- Security: ISMS/ISO control mapping, evidence collection, incident response.

Exit criteria:

- Service boundary and API contracts approved.
- Separate admin/frontend boundaries documented.
- Control map linked to ISMS/ISO evidence.

## Phase 2: Product Shell and Access Control

Engineering status: domain RBAC and tamper-evident audit log implemented. Static user and admin product-shell screens exist under `apps/web` and `apps/admin`; backend connectivity is pending.

Goal: build the visible product shell and privileged access model before handling real assets.

Deliverables:

- User frontend skeleton under `apps/web`.
- Admin frontend skeleton under `apps/admin`.
- Backend auth service with RBAC, MFA-ready admin roles, API key scopes.
- Session, HMAC, request id, audit middleware.
- Admin action audit log.

Exit criteria:

- User can view mock markets and account state.
- Admin can view mock users and compliance queues.
- Every admin action creates immutable audit records.

## Phase 3: Spot Trading Core

Goal: production-grade spot trading core in test mode.

Deliverables:

- Market config service.
- OMS idempotency and reservation.
- Matching engine with WAL and snapshots.
- Settlement and ledger persistence.
- Market data WebSocket.
- Account balances and order history.

Exit criteria:

- Alice/Bob spot flow passes against persistent storage.
- Matching replay is deterministic.
- Ledger balances reconcile by asset.
- No negative available or locked balance.

## Phase 4: Custody and Wallet Controls

Engineering status: Mock wallet adapter, mainnet-disabled gate, withdrawal-disabled gate, address whitelist check, admin approval audit, and mock broadcast implemented. Ledger integration for deposits/withdrawals and testnet adapters remain pending.

Goal: support wallet flows without mainnet risk first, then testnet only.

Deliverables:

- MockWalletAdapter.
- Testnet adapter design.
- Deposit scanner lifecycle.
- Withdrawal request, address whitelist, admin approval, signer separation.
- Hot/cold wallet policy documents.
- Travel Rule hook points for withdrawals.

Exit criteria:

- Mock deposit/withdrawal completes with ledger entries.
- Withdrawal default remains disabled.
- Approval/signing/broadcast are separated.
- Reorg and idempotency tests exist before any mainnet work.

## Phase 5: KYC, AML, Travel Rule, and Whitelists

Engineering status: KYC, AML, listing whitelist, and aggregate compliance decision packages implemented for policy gating. Vendor integrations, Travel Rule messaging, and case-management persistence are still pending.

Goal: make customer and transaction controls first-class services.

Deliverables:

- KYC service: identity status, document/vendor integration boundary, review state machine.
- AML service: risk scoring, sanctions/PEP/vendor integration boundary, case management.
- Travel Rule service: originator/beneficiary data, counterparty VASP whitelist, exception workflow.
- Asset/listing whitelist service: asset, network, contract, market listing lifecycle.
- Withdrawal address whitelist and cooldown.
- Suspicious transaction report workflow stub and evidence package.

Exit criteria:

- Trading/withdrawal gates depend on KYC/AML status.
- High-risk actions create AML cases.
- Whitelist state blocks disallowed assets, networks, counterparties, and addresses.

## Phase 6: ISMS/ISO Control Implementation

Engineering status: evidence register package and baseline asset inventory added for tracking control artifacts, owners, classifications, and freshness. Drill automation and formal evidence binders remain pending.

Goal: engineering evidence supports ISMS/ISMS-P and ISO/IEC 27001 readiness.

Deliverables:

- Asset inventory.
- Access control matrix.
- Secure SDLC policy and CI evidence.
- Vulnerability management workflow.
- Logging, monitoring, alerting, retention.
- Incident response runbook.
- Backup and recovery drill.
- Change management evidence.
- Vendor risk register.

Exit criteria:

- Control evidence is generated continuously.
- Security scans run in CI.
- Recovery and incident drills have artifacts.
- Audit log retention and tamper-evidence are implemented.

## Phase 7: VASP Filing and Bank Readiness Package

Engineering status: launch readiness gate added. Actual filing, bank readiness, and legal sign-off remain external dependencies.

Goal: prepare the non-code package needed for regulated operation.

Deliverables:

- FIU filing evidence binder.
- ISMS/preliminary ISMS evidence binder.
- AML policy and procedures.
- Bank risk assessment support package for real-name account discussions.
- User asset protection policy.
- Customer agreement, risk disclosure, privacy policy drafts.

Exit criteria:

- Compliance/legal signs off filing package completeness.
- Engineering can demonstrate controls in a staging environment.

## Phase 8: Closed Pilot in Staging

Goal: run production-like operations without public customer assets.

Deliverables:

- Staging environment with synthetic users and assets.
- Operational runbooks.
- Admin review exercises.
- AML scenario exercises.
- Incident response tabletop.
- Replay/recovery drills.

Exit criteria:

- No critical reconciliation gaps.
- No unaudited privileged operations.
- Support and compliance teams can operate workflows.

## Phase 9: Controlled Production Launch

Engineering status: production launch readiness gate and initial market-surveillance rule package added. Launch remains blocked unless legal, compliance, security, ISMS, VASP filing, operational drill, and surveillance conditions are true.

Goal: launch only the approved product scope.

Deliverables:

- Spot-only market launch checklist.
- Mainnet custody enablement checklist.
- Fiat enablement checklist if approved.
- Market surveillance dashboard.
- Customer support escalation workflow.
- Post-launch risk monitoring.

Exit criteria:

- Regulator/bank/legal prerequisites are satisfied.
- Launch scope is explicitly approved.
- Futures/perpetuals remain disabled unless separately approved.

## Phase 10: Derivatives Review and Expansion

Engineering status: derivatives readiness gate added. Derivatives remain blocked unless separate derivatives legal approval, risk approval, customer suitability, and surveillance controls are satisfied.

Goal: evaluate perpetual futures only after spot compliance maturity.

Deliverables:

- Legal review for derivatives by jurisdiction.
- Margin/risk/liquidation model audit.
- Insurance fund controls.
- Market manipulation surveillance.
- Customer suitability controls.

Exit criteria:

- Written approval exists before enabling any derivative product.
